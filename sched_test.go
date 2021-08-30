package sched_test

import (
	"sync"
	"testing"

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

	db := &testDB{}
	sm := sched.New(db)

	if as.NoError(sm.Start()) {
		if as.NoError(sm.Stop()) {
			return
		}
	}
}
