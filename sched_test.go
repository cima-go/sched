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

type TestStruct2 struct {
	In   int
	Max  int
	Over int
}

func TestSchedRun(t *testing.T) {
	as := assert.New(t)

	gi := 0
	fn := func(ctx sched.Context, data TestStruct2) {
		if data.Max == 0 || gi < data.Max {
			gi += data.In
		}
		if data.Over > 0 && gi >= data.Over {
			as.NoError(ctx.Cancel())
		}
	}

	db := &testDB{}
	log := &stdoutLogger{}
	sm1 := sched.New(db, sched.WithTicker(10*time.Millisecond), sched.WithLogger(log))
	as.NoError(sm1.Register("test", fn))

	// sched test 1 --- common
	if as.NoError(sm1.Start()) {
		// test run once
		if as.NoError(sm1.Once(time.Now().Add(10*time.Millisecond), sched.MakeJob("test", TestStruct2{In: 100}))) {
			time.Sleep(15 * time.Millisecond)
			as.Equal(100, gi) // should run
			time.Sleep(15 * time.Millisecond)
			as.Equal(100, gi) // should not run
		}

		// test run period
		gi = 0
		j1 := sched.MakeJob("test", TestStruct2{In: 10, Max: 50})
		if as.NoError(sm1.Every(10*time.Millisecond, j1)) {
			time.Sleep(100 * time.Millisecond)
			as.Equal(50, gi)
			as.NoError(sm1.Cancel(j1))
		}

		// test ctx cancel
		gi = 0
		j2 := sched.MakeJob("test", TestStruct2{In: 10, Over: 50})
		if as.NoError(sm1.Every(10*time.Millisecond, j2)) {
			time.Sleep(100 * time.Millisecond)
			as.Equal(50, gi)
			as.NoError(sm1.Cancel(j2))
		}

		// test restore - prepare
		gi = 0
		as.NoError(sm1.Once(time.Now().Add(10*time.Millisecond), sched.MakeJob("test", TestStruct2{In: 100})))

		// shutdown
		as.NoError(sm1.Stop())
	}

	// ensure db records
	if jobs, err := db.Query(); as.NoError(err) {
		as.Equal(1, len(jobs))
	}

	// sched test 2 --- test restore
	sm2 := sched.New(db, sched.WithTicker(10*time.Millisecond), sched.WithLogger(log))
	as.NoError(sm2.Register("test", fn))
	if as.NoError(sm2.Start()) {
		time.Sleep(20 * time.Millisecond)
		as.Equal(100, gi)

		// shutdown
		as.NoError(sm2.Stop())
	}
}
