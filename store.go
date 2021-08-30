package sched

type StoreItem struct {
	Id   string
	Data []byte
}

type Storage interface {
	Query() ([]*StoreItem, error)
	Upsert(item *StoreItem) error
	Delete(item *StoreItem) error
}
