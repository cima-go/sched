package sched

import (
	"encoding/json"
	"errors"
	"reflect"
	"time"

	"github.com/rs/xid"
)

func MakeJob(typ string, data interface{}) *Job {
	return &Job{
		Id:   xid.New().String(),
		Typ:  typ,
		Data: data,
	}
}

type Job struct {
	Id   string
	Typ  string
	Data interface{} `json:"-"`
	Raw  []byte
}

type Task struct {
	*Job
	Next   time.Time
	Period time.Duration
}

func (t *Task) Encode() ([]byte, error) {
	if len(t.Raw) == 0 {
		if dat, err := json.Marshal(t.Data); err != nil {
			return nil, err
		} else {
			t.Raw = dat
		}
	}

	return json.Marshal(t)
}

func (t *Task) Decode(data []byte) error {
	return json.Unmarshal(data, t)
}

func (t *Task) Assign(to interface{}) error {
	if t.Data != nil {
		reflect.ValueOf(to).Elem().Set(reflect.ValueOf(t.Data))
		return nil
	}

	if len(t.Raw) == 0 {
		return errors.New("no raw data")
	}

	if err := json.Unmarshal(t.Raw, to); err != nil {
		return err
	}

	t.Data = reflect.ValueOf(to).Elem().Interface()
	t.Raw = nil

	return nil
}
