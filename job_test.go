package sched_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cima-go/sched"
)

type TestStruct1 struct {
	Data1 string
	Data2 int
	data3 int
}

func TestJobCodec(t *testing.T) {
	as := assert.New(t)

	t1 := &TestStruct1{Data1: "111", Data2: 222, data3: 333}
	j1 := &sched.Task{Job: sched.MakeJob("test", t1)}
	bs1, err := j1.Encode()
	if as.NoError(err) {
		as.NotEmpty(bs1)
	}

	j2 := &sched.Task{}
	if as.NoError(j2.Decode(bs1)) {
		as.Equal("test", j2.Typ)
		var t2 *TestStruct1
		if as.NoError(j2.Assign(&t2)) {
			as.Equal(t1.Data1, t2.Data1)
			as.Equal(t1.Data2, t2.Data2)
			as.Equal(0, t2.data3)
			as.Equal(t2, j2.Data)
		}

		var t3 *TestStruct1
		if as.NoError(j2.Assign(&t3)) {
			as.Equal(t1.Data1, t3.Data1)
			as.Equal(t1.Data2, t3.Data2)
			as.Equal(j2.Data, t3)
		}
	}
}
