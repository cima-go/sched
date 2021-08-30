package sched

import (
	"runtime"
	"sync"
	"time"

	"github.com/Jeffail/tunny"
	"go.uber.org/zap"

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

func New(db Storage, opts ...Option) Manager {
	s := &sched{
		store:    db,
		pQueue:   pqueue.New(256),
		closeCh:  make(chan struct{}),
		closedCh: make(chan struct{}),
		logger:   zap.NewNop(),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

type sched struct {
	store Storage
	// queue
	pLock  sync.Mutex
	pQueue pqueue.PriorityQueue
	// register
	regLock sync.RWMutex
	regMaps map[string]func(e *Job) error
	// signal
	closeCh  chan struct{}
	closedCh chan struct{}
	// misc
	pool   *tunny.Pool
	logger *zap.Logger
}

func (s *sched) Start() error {
	s.pool = tunny.NewFunc(runtime.NumCPU(), func(i interface{}) interface{} {
		return s.invoke(i.(*Job))
	})

	go func() {
		ticker := time.NewTicker(time.Second)
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

	return nil
}

func (s *sched) Stop() error {
	close(s.closeCh)
	<-s.closedCh
	s.pool.Close()
	return nil
}

func (s *sched) Once(date time.Time, job *Job) error {
	return nil
}

func (s *sched) Every(period time.Duration, job *Job) error {
	return nil
}

func (s *sched) Cancel(job *Job) error {
	return nil
}
