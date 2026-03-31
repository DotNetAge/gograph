// Package graph provides core data structures and interfaces for the gograph database.
package graph

import (
	"sync/atomic"
)

// idCounter is a global atomic counter used for generating unique IDs.
// It is incremented atomically to ensure thread-safety across concurrent operations.
var idCounter uint64

// GenerateID generates a unique ID with the given prefix.
// It uses an atomic counter to ensure uniqueness across concurrent operations.
//
// The generated ID has the format: "prefix:XY" where:
//   - prefix is the provided prefix string
//   - X is a letter (a-z) based on the counter value
//   - Y is a digit (0-9) based on the counter value
//
// Parameters:
//   - prefix: The prefix to use for the ID (e.g., "node", "rel")
//
// Returns a unique string identifier.
//
// Example:
//
//	id1 := graph.GenerateID("node") // e.g., "node:a1"
//	id2 := graph.GenerateID("rel")  // e.g., "rel:b2"
func GenerateID(prefix string) string {
	id := atomic.AddUint64(&idCounter, 1)
	return prefix + ":" + string(rune('a'+id%26)) + string(rune('0'+id%10))
}

// SetIDCounter sets the ID counter to a specific value.
// This is primarily useful for testing purposes to ensure predictable IDs.
//
// Parameters:
//   - counter: The value to set the counter to
//
// Example:
//
//	// In test setup
//	graph.SetIDCounter(0)
//	id := graph.GenerateID("node") // Will generate "node:a1"
func SetIDCounter(counter uint64) {
	atomic.StoreUint64(&idCounter, counter)
}
