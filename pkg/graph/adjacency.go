// Package graph provides core data structures and interfaces for the gograph database.
// It includes Node, Relationship, and Property types along with utilities for
// indexing and adjacency list management.
package graph

import (
	"github.com/DotNetAge/gograph/pkg/storage"
)

// AdjacencyEntry represents a single relationship in a merged adjacency group.
type AdjacencyEntry struct {
	RelID  string `msgpack:"r"`
	NodeID string `msgpack:"n"`
}

// AdjacencyList manages the adjacency relationships between nodes.
// It provides methods to build and query relationships between graph nodes,
// enabling efficient traversal of the graph structure.
//
// The adjacency list stores relationships in merged groups per (nodeID, relType, direction)
// to reduce the number of small keys in the LSM-Tree. Each group is stored as
// a MessagePack-encoded list of AdjacencyEntry.
type AdjacencyList struct {
	store *storage.DB
	cache *AdjacencyCache
}

// NewAdjacencyList creates a new AdjacencyList instance.
func NewAdjacencyList(store *storage.DB) *AdjacencyList {
	return &AdjacencyList{
		store: store,
		cache: NewAdjacencyCache(10000),
	}
}

// loadGroup loads an adjacency group from storage.
func (adj *AdjacencyList) loadGroup(nodeID, relType, direction string) ([]AdjacencyEntry, error) {
	key := storage.AdjGroupKey(nodeID, relType, direction)
	data, err := adj.store.Get(key)
	if err != nil {
		// If not found, return empty slice
		return []AdjacencyEntry{}, nil
	}
	var entries []AdjacencyEntry
	if err := storage.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// saveGroup saves an adjacency group to storage via the mutator.
func (adj *AdjacencyList) saveGroup(m Mutator, nodeID, relType, direction string, entries []AdjacencyEntry) error {
	key := storage.AdjGroupKey(nodeID, relType, direction)
	if len(entries) == 0 {
		return m.Delete(key)
	}
	data, err := storage.Marshal(entries)
	if err != nil {
		return err
	}
	return m.Put(key, data)
}

// AddRelationship creates adjacency entries for a relationship within a mutation context.
// It appends to the merged groups for both outgoing and incoming directions.
func (adj *AdjacencyList) AddRelationship(m Mutator, rel *Relationship) error {
	return adj.AddRelationships(m, []*Relationship{rel})
}

// AddRelationships batch-creates adjacency entries for multiple relationships.
// It properly merges relationships into groups even when operating within a
// single uncommitted batch by aggregating in memory before writing.
func (adj *AdjacencyList) AddRelationships(m Mutator, rels []*Relationship) error {
	// groupKey -> entries (loaded from storage + new ones)
	type groupKey struct {
		nodeID    string
		relType   string
		direction string
	}
	groups := make(map[groupKey][]AdjacencyEntry)

	// Load existing groups from storage
	for _, rel := range rels {
		outKey := groupKey{rel.StartNodeID, rel.Type, "out"}
		if _, ok := groups[outKey]; !ok {
			entries, _ := adj.loadGroup(rel.StartNodeID, rel.Type, "out")
			groups[outKey] = entries
		}
		inKey := groupKey{rel.EndNodeID, rel.Type, "in"}
		if _, ok := groups[inKey]; !ok {
			entries, _ := adj.loadGroup(rel.EndNodeID, rel.Type, "in")
			groups[inKey] = entries
		}
	}

	// Append new relationships to groups
	for _, rel := range rels {
		outKey := groupKey{rel.StartNodeID, rel.Type, "out"}
		groups[outKey] = append(groups[outKey], AdjacencyEntry{RelID: rel.ID, NodeID: rel.EndNodeID})

		inKey := groupKey{rel.EndNodeID, rel.Type, "in"}
		groups[inKey] = append(groups[inKey], AdjacencyEntry{RelID: rel.ID, NodeID: rel.StartNodeID})
	}

	// Write all groups back
	for gk, entries := range groups {
		if err := adj.saveGroup(m, gk.nodeID, gk.relType, gk.direction, entries); err != nil {
			return err
		}
	}

	// Invalidate cache entries
	invalidated := make(map[groupKey]bool)
	for _, rel := range rels {
		outKey := groupKey{rel.StartNodeID, rel.Type, ""}
		if !invalidated[outKey] {
			adj.cache.Invalidate(rel.StartNodeID, rel.Type, "")
			invalidated[outKey] = true
		}
		inKey := groupKey{rel.EndNodeID, rel.Type, ""}
		if !invalidated[inKey] {
			adj.cache.Invalidate(rel.EndNodeID, rel.Type, "")
			invalidated[inKey] = true
		}
	}

	return nil
}

// RemoveRelationship removes adjacency entries for a relationship within a mutation context.
// It removes entries from the merged groups for both directions.
func (adj *AdjacencyList) RemoveRelationship(m Mutator, rel *Relationship) error {
	// Remove from outgoing group
	outEntries, err := adj.loadGroup(rel.StartNodeID, rel.Type, "out")
	if err != nil {
		return err
	}
	outEntries = filterEntries(outEntries, rel.ID)
	if err := adj.saveGroup(m, rel.StartNodeID, rel.Type, "out", outEntries); err != nil {
		return err
	}

	// Remove from incoming group
	inEntries, err := adj.loadGroup(rel.EndNodeID, rel.Type, "in")
	if err != nil {
		return err
	}
	inEntries = filterEntries(inEntries, rel.ID)
	if err := adj.saveGroup(m, rel.EndNodeID, rel.Type, "in", inEntries); err != nil {
		return err
	}

	// Invalidate cache entries for both nodes
	adj.cache.Invalidate(rel.StartNodeID, rel.Type, "")
	adj.cache.Invalidate(rel.EndNodeID, rel.Type, "")

	return nil
}

// filterEntries removes the entry with the given relID from the slice.
func filterEntries(entries []AdjacencyEntry, relID string) []AdjacencyEntry {
	filtered := make([]AdjacencyEntry, 0, len(entries))
	for _, e := range entries {
		if e.RelID != relID {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// GetRelatedNodes returns the IDs of nodes related to the given node with the
// specified relationship type and direction.
func (adj *AdjacencyList) GetRelatedNodes(nodeID, relType string, direction Direction) ([]string, error) {
	// Check cache first
	if cached, found := adj.cache.Get(nodeID, relType, direction); found {
		return cached, nil
	}

	var nodeIDs []string

	var directions []string
	switch direction {
	case DirectionOutgoing:
		directions = []string{"out"}
	case DirectionIncoming:
		directions = []string{"in"}
	case DirectionBoth:
		directions = []string{"out", "in"}
	}

	for _, dir := range directions {
		entries, err := adj.loadGroup(nodeID, relType, dir)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			nodeIDs = append(nodeIDs, e.NodeID)
		}
	}

	// Store in cache for future queries
	adj.cache.Put(nodeID, relType, direction, nodeIDs)

	return nodeIDs, nil
}

// GetAllRelated returns all Relationship IDs related to the given node, regardless of type.
// This is critical for efficient `Detach Delete` operations.
func (adj *AdjacencyList) GetAllRelated(nodeID string) ([]string, error) {
	var relIDs []string
	prefix := storage.AdjKeyPrefix(nodeID)

	iter, err := adj.store.NewIter(nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	for iter.SeekGE(prefix); iter.Valid(); iter.Next() {
		key := iter.Key()
		if !hasAdjPrefix(key, prefix) {
			break
		}

		data := iter.Value()
		var entries []AdjacencyEntry
		if err := storage.Unmarshal(data, &entries); err != nil {
			continue
		}
		for _, e := range entries {
			relIDs = append(relIDs, e.RelID)
		}
	}

	return relIDs, nil
}

// hasAdjPrefix checks if the key has the given prefix.
func hasAdjPrefix(key, prefix []byte) bool {
	return len(key) >= len(prefix) && string(key[:len(prefix)]) == string(prefix)
}
