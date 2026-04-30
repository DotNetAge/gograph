package graph

import (
	"container/list"
	"sync"
)

// AdjCacheEntry represents a cached adjacency query result.
type AdjCacheEntry struct {
	NodeID      string
	RelType     string
	Direction   Direction
	RelatedNodes []string
}

// cacheKey generates a unique key for cache lookup.
func (e *AdjCacheEntry) cacheKey() string {
	return e.NodeID + ":" + e.RelType + ":" + string(e.Direction)
}

// AdjacencyCache provides an LRU cache for adjacency list queries.
// It significantly improves performance for repeated relationship traversals.
type AdjacencyCache struct {
	mu       sync.RWMutex
	capacity int
	items    map[string]*list.Element
	order    *list.List
}

// NewAdjacencyCache creates a new adjacency cache with the given capacity.
//
// Parameters:
//   - capacity: Maximum number of entries to keep in cache
//
// Example:
//
//	cache := graph.NewAdjacencyCache(1000)
func NewAdjacencyCache(capacity int) *AdjacencyCache {
	return &AdjacencyCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		order:    list.New(),
	}
}

// Get retrieves cached adjacency results if available.
//
// Parameters:
//   - nodeID: The node ID to look up
//   - relType: The relationship type
//   - direction: The direction of traversal
//
// Returns the cached node IDs and true if found, nil and false otherwise.
func (c *AdjacencyCache) Get(nodeID, relType string, direction Direction) ([]string, bool) {
	key := nodeID + ":" + relType + ":" + string(direction)

	c.mu.RLock()
	elem, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		return nil, false
	}

	// Move to front (most recently used)
	c.mu.Lock()
	c.order.MoveToFront(elem)
	c.mu.Unlock()

	entry := elem.Value.(*AdjCacheEntry)
	return entry.RelatedNodes, true
}

// Put stores adjacency results in the cache.
//
// Parameters:
//   - nodeID: The node ID
//   - relType: The relationship type
//   - direction: The direction of traversal
//   - relatedNodes: The related node IDs to cache
func (c *AdjacencyCache) Put(nodeID, relType string, direction Direction, relatedNodes []string) {
	key := nodeID + ":" + relType + ":" + string(direction)

	c.mu.Lock()
	defer c.mu.Unlock()

	// If key already exists, update it and move to front
	if elem, exists := c.items[key]; exists {
		c.order.MoveToFront(elem)
		entry := elem.Value.(*AdjCacheEntry)
		entry.RelatedNodes = relatedNodes
		return
	}

	// Add new entry
	entry := &AdjCacheEntry{
		NodeID:       nodeID,
		RelType:      relType,
		Direction:    direction,
		RelatedNodes: relatedNodes,
	}
	elem := c.order.PushFront(entry)
	c.items[key] = elem

	// Evict oldest if over capacity
	if c.order.Len() > c.capacity {
		oldest := c.order.Back()
		if oldest != nil {
			oldEntry := oldest.Value.(*AdjCacheEntry)
			delete(c.items, oldEntry.cacheKey())
			c.order.Remove(oldest)
		}
	}
}

// Invalidate removes a specific entry from the cache.
//
// Parameters:
//   - nodeID: The node ID to invalidate
//   - relType: The relationship type (optional, "" means all types)
//   - direction: The direction (optional, use empty string for all directions)
func (c *AdjacencyCache) Invalidate(nodeID string, relType string, direction Direction) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Collect keys to delete
	var keysToDelete []string
	for key, elem := range c.items {
		entry := elem.Value.(*AdjCacheEntry)
		if entry.NodeID == nodeID {
			if relType == "" || entry.RelType == relType {
				if direction == "" || entry.Direction == direction {
					keysToDelete = append(keysToDelete, key)
				}
			}
		}
	}

	// Delete collected keys
	for _, key := range keysToDelete {
		if elem, exists := c.items[key]; exists {
			c.order.Remove(elem)
			delete(c.items, key)
		}
	}
}

// Clear removes all entries from the cache.
func (c *AdjacencyCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*list.Element)
	c.order = list.New()
}

// Size returns the current number of cached entries.
func (c *AdjacencyCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.order.Len()
}

// HitRate returns the cache hit rate (not implemented in basic version).
// For production use, add hit/miss counters.
func (c *AdjacencyCache) HitRate() float64 {
	// TODO: Implement hit/miss tracking
	return 0.0
}
