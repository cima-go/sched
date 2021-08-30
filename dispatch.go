package sched

import (
	"reflect"
	"time"

	"go.uber.org/zap"
)

type Dispatcher interface {
	Register(typ string, handler interface{}) error
}

func (s *sched) dispatch(before int64) {
	for {
		s.pLock.Lock()
		item, next := s.pQueue.PeekAndShift(before)
		s.pLock.Unlock()

		if item == nil {
			s.logger.With(zap.Duration("wait", time.Duration(next))).Debug("nearest task")
			return
		}

		s.pool.Process(item.Value)
	}
}

func (s *sched) invoke(job *Job) error {
	s.regLock.RLock()
	defer s.regLock.RUnlock()

	handler, has := s.regMaps[job.Typ]
	if !has {
		return ErrTypeNotRegister
	}

	return handler(job)
}

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

	i0 := x.In(0)
	if !i0.Implements(reflect.TypeOf((*Context)(nil)).Elem()) {
		panic("handler params[1] must be sched.Context")
	}

	i1 := x.In(1)
	if i1.Kind() == reflect.Struct {
		panic("handler params[2] must be struct")
	}

	v := reflect.ValueOf(handler)

	s.regMaps[typ] = func(e *Job) error {
		i1r := reflect.New(i1)
		if err := e.Assign(i1r.Interface()); err != nil {
			return err
		}

		vs := v.Call([]reflect.Value{reflect.ValueOf(&ctx{s: s, j: e}), i1r.Elem()})

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
