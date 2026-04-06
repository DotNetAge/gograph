package storage

import (
	"sync"
)

// MemoryDB is an in-memory key-value store.
type MemoryDB struct {
	mu   sync.RWMutex
	data map[string][]byte
}

// NewMemoryDB creates a new MemoryDB.
func NewMemoryDB() *MemoryDB {
	return &MemoryDB{
		data: make(map[string][]byte),
	}
}

// Get retrieves a value from the MemoryDB.
func (db *MemoryDB) Get(key []byte) ([]byte, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	val, ok := db.data[string(key)]
	if !ok {
		return nil, nil
	}
	ret := make([]byte, len(val))
	copy(ret, val)
	return ret, nil
}

// Put stores a key-value pair in the MemoryDB.
func (db *MemoryDB) Put(key, value []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	val := make([]byte, len(value))
	copy(val, value)
	db.data[string(key)] = val
	return nil
}

// Delete removes a key-value pair from the MemoryDB.
func (db *MemoryDB) Delete(key []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	delete(db.data, string(key))
	return nil
}

// MemoryBatch represents a batch of operations for MemoryDB.
type MemoryBatch struct {
	db   *MemoryDB
	ops  []func()
}

// NewMemoryBatch creates a new MemoryBatch.
func (db *MemoryDB) NewMemoryBatch() *MemoryBatch {
	return &MemoryBatch{
		db: db,
	}
}

// Put adds a put operation to the batch.
func (b *MemoryBatch) Put(key, value []byte) error {
	k := make([]byte, len(key))
	copy(k, key)
	v := make([]byte, len(value))
	copy(v, value)
	b.ops = append(b.ops, func() {
		b.db.Put(k, v)
	})
	return nil
}

// Delete adds a delete operation to the batch.
func (b *MemoryBatch) Delete(key []byte) error {
	k := make([]byte, len(key))
	copy(k, key)
	b.ops = append(b.ops, func() {
		b.db.Delete(k)
	})
	return nil
}

// Commit applies the batch operations to the MemoryDB.
func (b *MemoryBatch) Commit() error {
	b.db.mu.Lock()
	defer b.db.mu.Unlock()
	for _, op := range b.ops {
		op()
	}
	return nil
}

// Close closes the batch.
func (b *MemoryBatch) Close() error {
	b.ops = nil
	return nil
}
