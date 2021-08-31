package sched

import (
	"runtime"
	"sync"
	"time"

	"github.com/Jeffail/tunny"

	"github.com/cima-go/sched/pqueue"
)

type Manager interface {
	Dispatcher
	Scheduler
	Start() error
	Stop() error
}

type Scheduler interface {
	Once(date time.Time, job *Job) error
	Every(period time.Duration, job *Job) error
	Cancel(job *Job) error
}

type handler func(c Context) error

func New(db Storage, opts ...Option) Manager {
	s := &sched{
		store:    db,
		ticker:   time.Second,
		logger:   &noopLogger{},
		qItems:   pqueue.New(256),
		mItems:   make(map[string]*Task),
		regMaps:  make(map[string]handler),
		closeCh:  make(chan struct{}),
		closedCh: make(chan struct{}),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

type sched struct {
	store Storage
	// option
	ticker time.Duration
	logger Logger
	// task queue
	qLock  sync.Mutex
	qItems pqueue.PriorityQueue
	// task maps
	mLock  sync.RWMutex
	mItems map[string]*Task
	// register
	regLock sync.RWMutex
	regMaps map[string]handler
	// signal
	closeCh  chan struct{}
	closedCh chan struct{}
	// misc
	workers *tunny.Pool
}

func (s *sched) Start() error {
	s.workers = tunny.NewFunc(runtime.NumCPU(), func(i interface{}) interface{} {
		return s.invoke(i.(*Task))
	})

	go func() {
		ticker := time.NewTicker(s.ticker)
		defer func() {
			ticker.Stop()
			close(s.closedCh)
		}()

		for {
			select {
			case <-ticker.C:
				s.dispatch(time.Now().UnixNano())
			case <-s.closeCh:
				return
			}
		}
	}()

	// restore from database
	items, err := s.store.Query()
	if err != nil {
		return err
	}

	for _, item := range items {
		task := &Task{}
		if err := task.Decode(item.Data); err != nil {
			return err
		}
		if err := s.queued(task); err != nil {
			return err
		}
	}

	return nil
}

func (s *sched) Stop() error {
	close(s.closeCh)
	<-s.closedCh
	s.workers.Close()
	return nil
}

func (s *sched) Once(date time.Time, job *Job) error {
	task := &Task{Job: job, Next: date}
	if err := s.storeUpsert(task); err != nil {
		return err
	}
	return s.queued(task)
}

func (s *sched) Every(period time.Duration, job *Job) error {
	task := &Task{Job: job, Next: time.Now().Add(period), Period: period}
	if err := s.storeUpsert(task); err != nil {
		return err
	}
	return s.queued(task)
}

func (s *sched) Cancel(job *Job) error {
	if err := s.storeRemove(job); err != nil {
		return err
	}
	return s.forget(job.Id)
}

func (s *sched) storeRemove(job *Job) error {
	return s.store.Delete(&StoreItem{Id: job.Id})
}

func (s *sched) storeUpsert(task *Task) error {
	if data, err := task.Encode(); err != nil {
		return err
	} else {
		return s.store.Upsert(&StoreItem{Id: task.Id, Data: data})
	}
}
