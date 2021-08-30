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
	At   time.Time
	Typ  string
	Once bool
	Data interface{} `json:"-"`
	Raw  []byte
}

func (j *Job) Encode() ([]byte, error) {
	if len(j.Raw) == 0 {
		if dat, err := json.Marshal(j.Data); err != nil {
			return nil, err
		} else {
			j.Raw = dat
		}
	}

	return json.Marshal(j)
}

func (j *Job) Decode(data []byte) error {
	return json.Unmarshal(data, j)
}

func (j *Job) Assign(to interface{}) error {
	if j.Data != nil {
		reflect.ValueOf(to).Elem().Set(reflect.ValueOf(j.Data))
		return nil
	}

	if len(j.Raw) == 0 {
		return errors.New("no raw data")
	}

	if err := json.Unmarshal(j.Raw, to); err != nil {
		return err
	}

	j.Data = reflect.ValueOf(to).Elem().Interface()
	j.Raw = nil

	return nil
}
