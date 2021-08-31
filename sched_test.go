package sched_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/cima-go/sched"
)

type testDB struct {
	items sync.Map
}

func (t *testDB) Query() ([]*sched.StoreItem, error) {
	var items []*sched.StoreItem
	t.items.Range(func(key, value interface{}) bool {
		items = append(items, value.(*sched.StoreItem))
		return true
	})
	return items, nil
}

func (t *testDB) Upsert(item *sched.StoreItem) error {
	t.items.Store(item.Id, item)
	return nil
}

func (t *testDB) Delete(item *sched.StoreItem) error {
	t.items.Delete(item.Id)
	return nil
}

func TestSchedRun(t *testing.T) {
	as := assert.New(t)

	type ts1 struct {
		in   int
		max  int
		over int
	}

	gi := 0

	db := &testDB{}
	sm := sched.New(db, sched.WithTicker(10*time.Millisecond), sched.WithLogger(&stdoutLogger{}))
	as.NoError(sm.Register("test", func(ctx sched.Context, data ts1) {
		if data.max == 0 || gi < data.max {
			gi += data.in
		}
		if data.over > 0 && gi >= data.over {
			as.NoError(ctx.Cancel())
		}
	}))

	if as.NoError(sm.Start()) {
		// test run Once
		if as.NoError(sm.Once(time.Now().Add(10*time.Millisecond), sched.MakeJob("test", ts1{in: 100}))) {
			time.Sleep(15 * time.Millisecond)
			as.Equal(100, gi) // should run
			time.Sleep(15 * time.Millisecond)
			as.Equal(100, gi) // should not run
		}

		// test run Period
		gi = 0
		j1 := sched.MakeJob("test", ts1{in: 10, max: 50})
		if as.NoError(sm.Every(10*time.Millisecond, j1)) {
			time.Sleep(100 * time.Millisecond)
			as.Equal(50, gi)
			as.NoError(sm.Cancel(j1))
		}

		// test ctx cancel
		gi = 0
		j2 := sched.MakeJob("test", ts1{in: 10, over: 50})
		if as.NoError(sm.Every(10*time.Millisecond, j2)) {
			time.Sleep(100 * time.Millisecond)
			as.Equal(50, gi)
			as.NoError(sm.Cancel(j2))
		}

		if as.NoError(sm.Stop()) {
			return
		}
	}
}
