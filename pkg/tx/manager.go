// Package tx provides transaction management for gograph's storage layer.
package tx

import (
	"github.com/DotNetAge/gograph/pkg/storage"
)

// Manager creates and manages transactions for the storage layer.
type Manager struct {
	store *storage.DB
}

// NewManager creates a new transaction manager for the given storage.
func NewManager(store *storage.DB) *Manager {
	return &Manager{store: store}
}

// Transaction represents a database transaction. It provides methods for
// reading and writing data within a transaction.
type Transaction struct {
	store    *storage.DB
	batch    *storage.Batch
	readOnly bool
	closed   bool
}

// Begin starts a new transaction. If readOnly is true, the transaction
// cannot modify data but may have better performance.
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

// Get retrieves the value for the given key within the transaction.
func (t *Transaction) Get(key []byte) ([]byte, error) {
	if t.closed {
		return nil, ErrTransactionClosed
	}
	return t.store.Get(key)
}

// Put stores a key-value pair within the transaction.
func (t *Transaction) Put(key, value []byte) error {
	if t.closed {
		return ErrTransactionClosed
	}
	if t.readOnly {
		return ErrReadOnlyTransaction
	}
	return t.batch.Put(key, value)
}

// Delete removes the value for the given key within the transaction.
func (t *Transaction) Delete(key []byte) error {
	if t.closed {
		return ErrTransactionClosed
	}
	if t.readOnly {
		return ErrReadOnlyTransaction
	}
	return t.batch.Delete(key)
}

// Commit commits the transaction, making all changes permanent.
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

// Rollback aborts the transaction and discards all changes.
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

// IsReadOnly returns true if the transaction is read-only.
func (t *Transaction) IsReadOnly() bool {
	return t.readOnly
}

// IsClosed returns true if the transaction has been closed.
func (t *Transaction) IsClosed() bool {
	return t.closed
}
