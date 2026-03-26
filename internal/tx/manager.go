package tx

import (
	"github.com/DotNetAge/gograph/internal/storage"
)

type Manager struct {
	store *storage.DB
}

func NewManager(store *storage.DB) *Manager {
	return &Manager{store: store}
}

type Transaction struct {
	store    *storage.DB
	batch    *storage.Batch
	readOnly bool
	closed   bool
}

func (m *Manager) Begin(readOnly bool) (*Transaction, error) {
	var batch *storage.Batch
	if !readOnly {
		batch = m.store.NewBatch()
	}
	return &Transaction{
		store:    m.store,
		batch:    batch,
		readOnly: readOnly,
	}, nil
}

func (t *Transaction) Get(key []byte) ([]byte, error) {
	if t.closed {
		return nil, ErrTransactionClosed
	}
	return t.store.Get(key)
}

func (t *Transaction) Put(key, value []byte) error {
	if t.closed {
		return ErrTransactionClosed
	}
	if t.readOnly {
		return ErrReadOnlyTransaction
	}
	return t.batch.Put(key, value)
}

func (t *Transaction) Delete(key []byte) error {
	if t.closed {
		return ErrTransactionClosed
	}
	if t.readOnly {
		return ErrReadOnlyTransaction
	}
	return t.batch.Delete(key)
}

func (t *Transaction) Commit() error {
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

func (t *Transaction) Rollback() error {
	if t.closed {
		return nil
	}
	if t.batch != nil {
		t.batch.Close()
	}
	t.closed = true
	return nil
}

func (t *Transaction) IsReadOnly() bool {
	return t.readOnly
}

func (t *Transaction) IsClosed() bool {
	return t.closed
}
