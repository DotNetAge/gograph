package tx

type Tx interface {
	Get(key []byte) ([]byte, error)
	Put(key, value []byte) error
	Delete(key []byte) error
	Commit() error
	Rollback() error
	IsReadOnly() bool
	IsClosed() bool
}

var _ Tx = (*Transaction)(nil)
var _ Tx = (*MemoryTransaction)(nil)
