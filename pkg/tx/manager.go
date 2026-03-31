// Package tx provides transaction management for gograph's storage layer.
// It implements ACID transactions using PebbleDB's batch operations.
//
// Transactions ensure that multiple operations either all succeed or all fail,
// maintaining data consistency. The package supports both read-only and
// read-write transactions.
//
// Basic Usage:
//
//	manager := tx.NewManager(store)
//
//	// Begin a read-write transaction
//	tx, err := manager.Begin(false)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Perform operations
//	err = tx.Put([]byte("key"), []byte("value"))
//	if err != nil {
//	    tx.Rollback()
//	    log.Fatal(err)
//	}
//
//	// Commit the transaction
//	err = tx.Commit()
//
// Thread Safety:
//
// Manager is safe for concurrent use. Each transaction is independent
// and should only be used by a single goroutine.
package tx

import (
	"errors"

	"github.com/DotNetAge/gograph/pkg/storage"
)

// ErrTransactionClosed is returned when attempting to use a closed transaction.
var ErrTransactionClosed = errors.New("transaction is closed")

// ErrReadOnlyTransaction is returned when attempting to write to a read-only transaction.
var ErrReadOnlyTransaction = errors.New("cannot write to read-only transaction")

// Manager creates and manages transactions for the storage layer.
// It provides methods to begin new transactions with different isolation levels.
type Manager struct {
	store *storage.DB
}

// NewManager creates a new transaction manager for the given storage.
//
// Parameters:
//   - store: The storage database to manage transactions for
//
// Returns a new Manager instance.
//
// Example:
//
//	store, _ := storage.Open("/path/to/db")
//	manager := tx.NewManager(store)
func NewManager(store *storage.DB) *Manager {
	return &Manager{store: store}
}

// Transaction represents a database transaction. It provides methods for
// reading and writing data within a transaction context.
//
// Transactions support:
//   - Read operations (Get)
//   - Write operations (Put, Delete) - only in read-write transactions
//   - Atomic commit or rollback
//
// Example:
//
//	tx, _ := manager.Begin(false)
//	defer func() {
//	    if err != nil {
//	        tx.Rollback()
//	    }
//	}()
//
//	// Perform operations
//	value, _ := tx.Get([]byte("key"))
//	tx.Put([]byte("key"), []byte("new value"))
//
//	// Commit
//	err = tx.Commit()
type Transaction struct {
	store    *storage.DB
	batch    *storage.Batch
	readOnly bool
	closed   bool
}

// Begin starts a new transaction.
//
// Parameters:
//   - readOnly: If true, the transaction cannot modify data but may have
//     better performance. If false, the transaction supports both reads and writes.
//
// Returns a new Transaction or an error if the transaction cannot be started.
//
// Example:
//
//	// Read-only transaction for queries
//	tx, _ := manager.Begin(true)
//	value, _ := tx.Get([]byte("key"))
//	tx.Commit()
//
//	// Read-write transaction for updates
//	tx, _ := manager.Begin(false)
//	tx.Put([]byte("key"), []byte("value"))
//	tx.Commit()
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
//
// Parameters:
//   - key: The key to look up
//
// Returns the value as a byte slice, or an error if the key is not found
// or if the transaction is closed.
//
// Example:
//
//	value, err := tx.Get([]byte("mykey"))
//	if err != nil {
//	    log.Printf("Key not found: %v", err)
//	}
func (t *Transaction) Get(key []byte) ([]byte, error) {
	if t.closed {
		return nil, ErrTransactionClosed
	}
	return t.store.Get(key)
}

// Put stores a key-value pair within the transaction.
// The operation is not persisted until Commit is called.
//
// Parameters:
//   - key: The key to store
//   - value: The value to store
//
// Returns an error if the transaction is closed or read-only.
//
// Example:
//
//	err := tx.Put([]byte("key"), []byte("value"))
//	if err != nil {
//	    tx.Rollback()
//	    log.Fatal(err)
//	}
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
// The operation is not persisted until Commit is called.
//
// Parameters:
//   - key: The key to delete
//
// Returns an error if the transaction is closed or read-only.
//
// Example:
//
//	err := tx.Delete([]byte("key"))
//	if err != nil {
//	    tx.Rollback()
//	    log.Fatal(err)
//	}
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
// For read-only transactions, this simply marks the transaction as closed.
//
// Returns an error if the transaction is already closed or if the commit fails.
//
// Example:
//
//	if err := tx.Commit(); err != nil {
//	    log.Fatal(err)
//	}
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
// This is safe to call even if the transaction is already closed.
//
// Returns nil if the rollback succeeds.
//
// Example:
//
//	defer func() {
//	    if err != nil {
//	        tx.Rollback()
//	    }
//	}()
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
//
// Example:
//
//	if tx.IsReadOnly() {
//	    fmt.Println("This is a read-only transaction")
//	}
func (t *Transaction) IsReadOnly() bool {
	return t.readOnly
}

// IsClosed returns true if the transaction has been closed.
// A transaction is closed after Commit or Rollback is called.
//
// Example:
//
//	if !tx.IsClosed() {
//	    tx.Rollback()
//	}
func (t *Transaction) IsClosed() bool {
	return t.closed
}
