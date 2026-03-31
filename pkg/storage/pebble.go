// Package storage provides a key-value storage layer for gograph using PebbleDB.
// It wraps the Pebble database to provide a simplified interface for storing
// and retrieving graph data including nodes, relationships, and indices.
//
// The storage layer uses the following key prefixes:
//   - "n:" - Node data
//   - "r:" - Relationship data
//   - "i:" - Index data
//   - "a:" - Adjacency list data
//
// Basic Usage:
//
//	db, err := storage.Open("/path/to/db")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer db.Close()
//
//	// Store data
//	err = db.Put([]byte("key"), []byte("value"))
//
//	// Retrieve data
//	value, err := db.Get([]byte("key"))
//
// Thread Safety:
//
// DB is safe for concurrent use by multiple goroutines.
package storage

import (
	"github.com/cockroachdb/pebble"
)

// noopLogger is a logger that discards all log output.
// It's used to suppress Pebble's default logging.
type noopLogger struct{}

// Infof implements the pebble.Logger interface, discarding info messages.
func (l noopLogger) Infof(format string, args ...interface{}) {}

// Fatalf implements the pebble.Logger interface, discarding fatal messages.
// Note: In production, you might want to handle fatal errors differently.
func (l noopLogger) Fatalf(format string, args ...interface{}) {}

// DB wraps a Pebble database instance.
// It provides a simplified interface for key-value operations and
// manages the underlying Pebble database lifecycle.
//
// DB is safe for concurrent use by multiple goroutines.
type DB struct {
	db *pebble.DB
}

// Open opens a database at the given path and returns a DB instance.
//
// Parameters:
//   - path: The directory path where the database files will be stored.
//     If the directory doesn't exist, it will be created.
//
// Returns a new DB instance or an error if the database cannot be opened.
//
// Example:
//
//	db, err := storage.Open("/var/lib/gograph/mydb")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer db.Close()
func Open(path string) (*DB, error) {
	db, err := pebble.Open(path, &pebble.Options{
		Logger:                   noopLogger{},
		MemTableSize:             4 << 20,
		MaxConcurrentCompactions: func() int { return 2 },
	})
	if err != nil {
		return nil, err
	}
	return &DB{db: db}, nil
}

// Close closes the database and releases all associated resources.
// Any pending writes will be flushed before closing.
//
// Returns an error if the close operation fails.
func (db *DB) Close() error {
	return db.db.Close()
}

// Get retrieves the value for the given key.
//
// Parameters:
//   - key: The key to look up
//
// Returns the value as a byte slice, or an error if the key is not found
// or if a database error occurs.
//
// Example:
//
//	value, err := db.Get([]byte("mykey"))
//	if err != nil {
//	    log.Printf("Key not found: %v", err)
//	} else {
//	    fmt.Printf("Value: %s\n", value)
//	}
func (db *DB) Get(key []byte) ([]byte, error) {
	val, closer, err := db.db.Get(key)
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	ret := make([]byte, len(val))
	copy(ret, val)
	return ret, nil
}

// Put stores a key-value pair in the database.
// If the key already exists, its value will be overwritten.
//
// Parameters:
//   - key: The key to store
//   - value: The value to store
//
// Returns an error if the operation fails.
//
// Example:
//
//	err := db.Put([]byte("mykey"), []byte("myvalue"))
//	if err != nil {
//	    log.Fatal(err)
//	}
func (db *DB) Put(key, value []byte) error {
	return db.db.Set(key, value, pebble.Sync)
}

// Delete removes the value for the given key.
// If the key doesn't exist, this is a no-op.
//
// Parameters:
//   - key: The key to delete
//
// Returns an error if the operation fails.
//
// Example:
//
//	err := db.Delete([]byte("mykey"))
//	if err != nil {
//	    log.Fatal(err)
//	}
func (db *DB) Delete(key []byte) error {
	return db.db.Delete(key, pebble.Sync)
}

// NewBatch creates a new batch for atomic write operations.
// Batches allow multiple operations to be committed atomically.
//
// Returns a new Batch instance.
//
// Example:
//
//	batch := db.NewBatch()
//	batch.Put([]byte("key1"), []byte("value1"))
//	batch.Put([]byte("key2"), []byte("value2"))
//	batch.Delete([]byte("key3"))
//	err := batch.Commit()
func (db *DB) NewBatch() *Batch {
	return &Batch{batch: db.db.NewBatch()}
}

// NewIter creates a new iterator for scanning the database.
//
// Parameters:
//   - opts: Iterator options. Pass nil for default options.
//
// Returns a new Iterator or an error if creation fails.
//
// Example:
//
//	iter, err := db.NewIter(nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer iter.Close()
//
//	for iter.SeekGE([]byte("prefix")); iter.Valid(); iter.Next() {
//	    fmt.Printf("%s: %s\n", iter.Key(), iter.Value())
//	}
func (db *DB) NewIter(opts *pebble.IterOptions) (*Iterator, error) {
	iter, err := db.db.NewIter(opts)
	if err != nil {
		return nil, err
	}
	return &Iterator{iter: iter}, nil
}

// Batch represents a collection of write operations that can be committed atomically.
// Use NewBatch to create a batch, add operations to it, then call Commit.
type Batch struct {
	batch *pebble.Batch
}

// Put adds a put operation to the batch.
// The operation will not be applied until Commit is called.
//
// Parameters:
//   - key: The key to store
//   - value: The value to store
//
// Returns an error if the operation cannot be added to the batch.
func (b *Batch) Put(key, value []byte) error {
	return b.batch.Set(key, value, nil)
}

// Delete adds a delete operation to the batch.
// The operation will not be applied until Commit is called.
//
// Parameters:
//   - key: The key to delete
//
// Returns an error if the operation cannot be added to the batch.
func (b *Batch) Delete(key []byte) error {
	return b.batch.Delete(key, nil)
}

// Commit applies all operations in the batch atomically.
// If any operation fails, none of the operations will be applied.
//
// Returns an error if the commit fails.
func (b *Batch) Commit() error {
	return b.batch.Commit(pebble.Sync)
}

// Close releases resources associated with the batch.
// This should be called after the batch is no longer needed.
//
// Returns an error if closing fails.
func (b *Batch) Close() error {
	return b.batch.Close()
}

// Iterator provides ordered iteration over the database.
// Iterators must be closed when no longer needed to release resources.
type Iterator struct {
	iter *pebble.Iterator
}

// SeekGE positions the iterator at the first key that is greater than or equal to the given key.
//
// Parameters:
//   - key: The key to seek to
//
// Returns true if a valid key was found, false otherwise.
func (i *Iterator) SeekGE(key []byte) bool { return i.iter.SeekGE(key) }

// SeekLT positions the iterator at the last key that is less than the given key.
//
// Parameters:
//   - key: The key to seek to
//
// Returns true if a valid key was found, false otherwise.
func (i *Iterator) SeekLT(key []byte) bool { return i.iter.SeekLT(key) }

// Next advances the iterator to the next key.
//
// Returns true if a valid key was found, false if the iterator is exhausted.
func (i *Iterator) Next() bool { return i.iter.Next() }

// Prev moves the iterator to the previous key.
//
// Returns true if a valid key was found, false if the iterator is exhausted.
func (i *Iterator) Prev() bool { return i.iter.Prev() }

// Valid returns true if the iterator is positioned at a valid key.
func (i *Iterator) Valid() bool { return i.iter.Valid() }

// Key returns the current key.
// The returned slice is only valid until the next iterator operation.
func (i *Iterator) Key() []byte { return i.iter.Key() }

// Value returns the current value.
// The returned slice is only valid until the next iterator operation.
func (i *Iterator) Value() []byte { return i.iter.Value() }

// Error returns any accumulated error during iteration.
func (i *Iterator) Error() error { return i.iter.Error() }

// Close releases resources associated with the iterator.
// This must be called when the iterator is no longer needed.
//
// Returns an error if closing fails.
func (i *Iterator) Close() error { return i.iter.Close() }

// DumpAll returns all key-value pairs in the database.
// This is primarily useful for debugging and testing.
//
// Returns a map of all keys to their values, or an error if the operation fails.
//
// Warning: This method loads all data into memory and should not be used
// on large databases.
//
// Example:
//
//	data, err := db.DumpAll()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for key, value := range data {
//	    fmt.Printf("%s: %s\n", key, value)
//	}
func (db *DB) DumpAll() (map[string][]byte, error) {
	result := make(map[string][]byte)
	iter, err := db.NewIter(nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	for iter.SeekGE([]byte{}); iter.Valid(); iter.Next() {
		key := string(iter.Key())
		value := make([]byte, len(iter.Value()))
		copy(value, iter.Value())
		result[key] = value
	}
	return result, nil
}
