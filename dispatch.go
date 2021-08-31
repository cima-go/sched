package sched

import (
	"container/heap"
	"reflect"

	"github.com/cima-go/sched/pqueue"
)

type Dispatcher interface {
	Register(typ string, handler interface{}) error
}

func (s *sched) queued(task *Task) error {
	item := &pqueue.Item{
		Value:    task.Id,
		Priority: task.Next.UnixNano(),
	}

	// add to queue
	s.qLock.Lock()
	heap.Push(&s.qItems, item)
	s.qLock.Unlock()

	// add to map
	s.mLock.Lock()
	s.mItems[task.Id] = task
	s.mLock.Unlock()

	return nil
}

func (s *sched) forget(id string) error {
	// skip queue

	// del from map
	s.mLock.Lock()
	delete(s.mItems, id)
	s.mLock.Unlock()

	return nil
}

func (s *sched) dispatch(before int64) {
	for {
		s.qLock.Lock()
		item, _ := s.qItems.PeekAndShift(before)
		s.qLock.Unlock()

		if item == nil {
			return
		}

		tid := item.Value.(string)
		s.mLock.Lock()
		task, found := s.mItems[tid]
		if !found {
			s.mLock.Unlock()
			s.logger.Debugf("missing job in maps, maybe cancelled")
			continue
		}
		delete(s.mItems, tid)
		s.mLock.Unlock()

		resp := s.workers.Process(task)
		if err, is := resp.(error); is {
			s.logger.Infof("worker process got error: %s", err.Error())
		}
	}
}

func (s *sched) invoke(task *Task) error {
	s.regLock.RLock()
	defer s.regLock.RUnlock()

	handler, has := s.regMaps[task.Typ]
	if !has {
		return ErrTypeNotRegistered
	}

	ctx := &ctx{s: s, t: task}
	if err := handler(ctx); err != nil {
		return err
	}

	if ctx.cancelled {
		return nil
	}

	if task.Period == 0 {
		return s.storeRemove(task.Job)
	} else {
		return s.Every(task.Period, task.Job)
	}
}

// Register ("test", func(ctx Context, data struct{}) {})
func (s *sched) Register(typ string, handler interface{}) error {
	s.regLock.Lock()
	defer s.regLock.Unlock()

	if _, has := s.regMaps[typ]; has {
		return ErrTypeBeenRegistered
	}

	x := reflect.TypeOf(handler)
	if x.Kind() != reflect.Func {
		panic("handler must be function")
	}

	if x.NumIn() != 2 {
		panic("handler must have two params")
	}

	i0 := x.In(0)
	if !i0.Implements(reflect.TypeOf((*Context)(nil)).Elem()) {
		panic("handler params[1] must be sched.Context")
	}

	i1 := x.In(1)
	if i1.Kind() != reflect.Struct {
		panic("handler params[2] must be struct")
	}

	v := reflect.ValueOf(handler)

	s.regMaps[typ] = func(c Context) error {
		i1r := reflect.New(i1)
		if err := c.Task().Assign(i1r.Interface()); err != nil {
			return err
		}

		vs := v.Call([]reflect.Value{reflect.ValueOf(c), i1r.Elem()})

		if len(vs) == 0 {
			return nil
		}

		if vs[0].IsNil() || !vs[0].Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			return nil
		}

		return vs[0].Interface().(error)
	}

	return nil
}
