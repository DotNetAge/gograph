package tx

import (
	"errors"

	"github.com/DotNetAge/gograph/pkg/storage"
)

var ErrTxClosed = errors.New("transaction is closed")

type MemoryManager struct {
	store *storage.MemoryDB
}

func NewMemoryManager(store *storage.MemoryDB) *MemoryManager {
	return &MemoryManager{store: store}
}

type MemoryTransaction struct {
	store    *storage.MemoryDB
	batch    *storage.MemoryBatch
	readOnly bool
	closed   bool
}

func (m *MemoryManager) Begin(readOnly bool) (*MemoryTransaction, error) {
	if readOnly {
		return &MemoryTransaction{
			store:    m.store,
			readOnly: true,
		}, nil
	}
	return &MemoryTransaction{
		store:    m.store,
		batch:    m.store.NewMemoryBatch(),
		readOnly: false,
	}, nil
}

func (t *MemoryTransaction) Get(key []byte) ([]byte, error) {
	if t.closed {
		return nil, ErrTransactionClosed
	}
	val, err := t.store.Get(key)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, errors.New("key not found")
	}
	return val, nil
}

func (t *MemoryTransaction) Put(key, value []byte) error {
	if t.closed {
		return ErrTransactionClosed
	}
	if t.readOnly {
		return ErrReadOnlyTransaction
	}
	return t.batch.Put(key, value)
}

func (t *MemoryTransaction) Delete(key []byte) error {
	if t.closed {
		return ErrTransactionClosed
	}
	if t.readOnly {
		return ErrReadOnlyTransaction
	}
	return t.batch.Delete(key)
}

func (t *MemoryTransaction) Commit() error {
	if t.closed {
		return ErrTransactionClosed
	}
	if t.readOnly {
		t.closed = true
		return nil
	}
	if err := t.batch.Commit(); err != nil {
		return err
	}
	t.closed = true
	return nil
}

func (t *MemoryTransaction) Rollback() error {
	if t.closed {
		return nil
	}
	if t.batch != nil {
		t.batch.Close()
	}
	t.closed = true
	return nil
}

func (t *MemoryTransaction) IsReadOnly() bool {
	return t.readOnly
}

func (t *MemoryTransaction) IsClosed() bool {
	return t.closed
}
