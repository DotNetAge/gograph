// Package kv defines the key-value storage abstraction for gograph.
// It provides interfaces that decouple the graph engine from specific
// storage implementations (e.g., PebbleDB, MemoryDB, mock stores).
package kv

// Store defines the interface for the underlying key-value storage.
type Store interface {
	// Get retrieves the value for the given key.
	Get(key []byte) ([]byte, error)

	// Put stores a key-value pair.
	Put(key, value []byte) error

	// Delete removes the value for the given key.
	Delete(key []byte) error

	// NewIter creates a new iterator for scanning the storage.
	// The opts parameter is implementation-specific (e.g., *pebble.IterOptions).
	NewIter(opts interface{}) (Iterator, error)

	// Close closes the storage and releases resources.
	Close() error
}

// Iterator defines the interface for iterating over storage entries.
type Iterator interface {
	// Valid returns true if the iterator is positioned at a valid entry.
	Valid() bool

	// Key returns the key at the current position.
	Key() []byte

	// Value returns the value at the current position.
	Value() []byte

	// Next advances the iterator to the next entry.
	Next() bool

	// Prev moves the iterator to the previous entry.
	Prev() bool

	// SeekGE positions the iterator at the first key that is >= the given key.
	SeekGE(key []byte) bool

	// SeekLT positions the iterator at the last key that is < the given key.
	SeekLT(key []byte) bool

	// Error returns any accumulated error during iteration.
	Error() error

	// Close closes the iterator and releases resources.
	Close() error
}
