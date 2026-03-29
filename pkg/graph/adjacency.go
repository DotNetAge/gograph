// Package graph provides core data structures and interfaces for the gograph database.
// It includes Node, Relationship, and Property types along with utilities for
// indexing and adjacency list management.
package graph

import (
	"strings"
	"github.com/DotNetAge/gograph/pkg/storage"
)

// AdjacencyList manages the adjacency relationships between nodes.
// It provides methods to build and query relationships between graph nodes.
type AdjacencyList struct {
	store *storage.DB
}

// NewAdjacencyList creates a new AdjacencyList instance.
func NewAdjacencyList(store *storage.DB) *AdjacencyList {
	return &AdjacencyList{store: store}
}

// AddRelationship creates adjacency entries for a relationship within a mutation context.
func (adj *AdjacencyList) AddRelationship(m Mutator, rel *Relationship) error {
	outKey := storage.AdjKey(rel.StartNodeID, rel.Type, "out", rel.ID)
	if err := m.Put(outKey, []byte(rel.EndNodeID)); err != nil {
		return err
	}

	inKey := storage.AdjKey(rel.EndNodeID, rel.Type, "in", rel.ID)
	if err := m.Put(inKey, []byte(rel.StartNodeID)); err != nil {
		return err
	}

	return nil
}

// RemoveRelationship removes adjacency entries for a relationship within a mutation context.
func (adj *AdjacencyList) RemoveRelationship(m Mutator, rel *Relationship) error {
	outKey := storage.AdjKey(rel.StartNodeID, rel.Type, "out", rel.ID)
	if err := m.Delete(outKey); err != nil {
		return err
	}

	inKey := storage.AdjKey(rel.EndNodeID, rel.Type, "in", rel.ID)
	if err := m.Delete(inKey); err != nil {
		return err
	}

	return nil
}

// GetRelatedNodes returns the IDs of nodes related to the given node with the
// specified relationship type and direction.
func (adj *AdjacencyList) GetRelatedNodes(nodeID, relType string, direction Direction) ([]string, error) {
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
		prefix := storage.AdjKeyPrefixNodeAndTypeAndDir(nodeID, relType, dir)
		iter, err := adj.store.NewIter(nil)
		if err != nil {
			return nil, err
		}

		func() {
			defer iter.Close()
			for iter.SeekGE(prefix); iter.Valid(); iter.Next() {
				key := iter.Key()
				if !hasAdjPrefix(key, prefix) {
					break
				}
				nodeIDs = append(nodeIDs, string(iter.Value()))
			}
		}()
	}

	return nodeIDs, nil
}

// GetAllRelated returns all Relationship IDs related to the given node, regardless of type.
// This is critical for efficient `Detach Delete`.
func (adj *AdjacencyList) GetAllRelated(nodeID string) ([]string, error) {
	var relIDs []string
	prefix := storage.AdjKeyPrefix(nodeID)

	iter, err := adj.store.NewIter(nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	prefixStr := string(prefix)
	for iter.SeekGE(prefix); iter.Valid(); iter.Next() {
		key := iter.Key()
		keyStr := string(key)
		if !hasAdjPrefix(key, prefix) {
			break
		}

		remainder := strings.TrimPrefix(keyStr, prefixStr)
		parts := strings.Split(remainder, ":")
		if len(parts) >= 3 {
			relID := strings.Join(parts[2:], ":")
			relIDs = append(relIDs, relID)
		}
	}

	return relIDs, nil
}

func hasAdjPrefix(key, prefix []byte) bool {
	return len(key) >= len(prefix) && string(key[:len(prefix)]) == string(prefix)
}