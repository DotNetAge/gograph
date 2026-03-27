// Package graph provides core data structures and interfaces for the gograph database.
package graph

import (
	"sync/atomic"
)

var idCounter uint64

// GenerateID generates a unique ID with the given prefix.
// It uses an atomic counter to ensure uniqueness across concurrent operations.
func GenerateID(prefix string) string {
	id := atomic.AddUint64(&idCounter, 1)
	return prefix + ":" + string(rune('a'+id%26)) + string(rune('0'+id%10))
}

// SetIDCounter sets the ID counter to a specific value.
// This is primarily useful for testing purposes.
func SetIDCounter(counter uint64) {
	atomic.StoreUint64(&idCounter, counter)
}
