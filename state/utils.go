package state

// DatabaseReader wraps the Get method of a backing data store.
type DatabaseReader interface {
	Get(key []byte) (value []byte, err error)
}

// DatabaseDeleter wraps the Delete method of a backing data store.
type DatabaseDeleter interface {
	Delete(key []byte) error
}

// DatabasePutter wraps the Put method of a backing data store.
type DatabasePutter interface {
	Put(key []byte, value []byte) error
}
