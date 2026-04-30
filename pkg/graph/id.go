package graph

import (
	"math/rand"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

var (
	entropyLock sync.Mutex
	entropyPool = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// GenerateID generates a unique ID with the given prefix.
// It uses ULID for globally unique, sortable identifiers.
//
// The generated ID has the format: "prefix:XXXXXXXXXXXXXXXXXXXXXXXXXX" where:
//   - prefix is the provided prefix string
//   - X is an ULID string (26 characters, Crockford Base32 encoded)
//
// Parameters:
//   - prefix: The prefix to use for the ID (e.g., "node", "rel")
//
// Returns a unique string identifier.
//
// Example:
//
//	id1 := graph.GenerateID("node") // e.g., "node:01ARZ3NDEKTSV4RRFFQ69G5FAV"
//	id2 := graph.GenerateID("rel")  // e.g., "rel:01ARZ3NDEKTSV4RRFFQ69G5FAW"
func GenerateID(prefix string) string {
	entropyLock.Lock()
	id := ulid.MustNew(ulid.Timestamp(time.Now()), ulid.Monotonic(entropyPool, 0))
	entropyLock.Unlock()
	return prefix + ":" + id.String()
}

// SetIDCounter sets the ID counter to a specific value.
// This is primarily useful for testing purposes to ensure predictable IDs.
//
// Deprecated: ULID generation no longer uses an atomic counter. This function
// is a no-op and will be removed in a future version. Use deterministic ID
// generation in tests by mocking or providing a custom ID generator instead.
//
// Parameters:
//   - counter: The value to set the counter to (ignored)
//
// Example:
//
//	// In test setup
//	graph.SetIDCounter(0)
//	id := graph.GenerateID("node") // Still generates ULID (counter parameter unused)
func SetIDCounter(counter uint64) {
	// ULID generation uses entropy pool instead of atomic counter
	// This function is deprecated and will be removed in a future version.
}
