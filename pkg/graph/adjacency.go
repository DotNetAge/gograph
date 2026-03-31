// Package graph provides core data structures and interfaces for the gograph database.
// It includes Node, Relationship, and Property types along with utilities for
// indexing and adjacency list management.
package graph

import (
	"strings"

	"github.com/DotNetAge/gograph/pkg/storage"
)

// AdjacencyList manages the adjacency relationships between nodes.
// It provides methods to build and query relationships between graph nodes,
// enabling efficient traversal of the graph structure.
//
// The adjacency list stores relationships in both directions:
//   - Outgoing: From start node to end node
//   - Incoming: From end node to start node
//
// This allows for efficient querying of relationships in either direction.
//
// Example:
//
//	adj := graph.NewAdjacencyList(store)
//
//	// Add a relationship to the adjacency list
//	err := adj.AddRelationship(mutator, relationship)
//
//	// Query related nodes
//	related, err := adj.GetRelatedNodes(nodeID, "KNOWS", graph.DirectionOutgoing)
type AdjacencyList struct {
	store *storage.DB
}

// NewAdjacencyList creates a new AdjacencyList instance.
//
// Parameters:
//   - store: The storage database to use for adjacency data
//
// Returns a new AdjacencyList ready to manage relationships.
//
// Example:
//
//	store, _ := storage.Open("/path/to/db")
//	adj := graph.NewAdjacencyList(store)
func NewAdjacencyList(store *storage.DB) *AdjacencyList {
	return &AdjacencyList{store: store}
}

// AddRelationship creates adjacency entries for a relationship within a mutation context.
// It creates both outgoing and incoming entries to enable bidirectional traversal.
//
// Parameters:
//   - m: The Mutator to use for the operation
//   - rel: The relationship to add to the adjacency list
//
// Returns an error if the operation fails.
//
// Example:
//
//	err := adj.AddRelationship(mutator, relationship)
//	if err != nil {
//	    log.Fatal(err)
//	}
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
// It removes both outgoing and incoming entries.
//
// Parameters:
//   - m: The Mutator to use for the operation
//   - rel: The relationship to remove from the adjacency list
//
// Returns an error if the operation fails.
//
// Example:
//
//	err := adj.RemoveRelationship(mutator, relationship)
//	if err != nil {
//	    log.Fatal(err)
//	}
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
//
// Parameters:
//   - nodeID: The ID of the node to find relationships for
//   - relType: The type of relationship to look for
//   - direction: The direction of relationships to query (outgoing, incoming, or both)
//
// Returns a slice of node IDs that are related to the given node, or an error if the query fails.
//
// Example:
//
//	// Get all nodes that this node knows
//	related, err := adj.GetRelatedNodes(nodeID, "KNOWS", graph.DirectionOutgoing)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, id := range related {
//	    fmt.Printf("Related node: %s\n", id)
//	}
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
// This is critical for efficient `Detach Delete` operations.
//
// Parameters:
//   - nodeID: The ID of the node to find all relationships for
//
// Returns a slice of relationship IDs, or an error if the query fails.
//
// Example:
//
//	// Get all relationships for a node (for detach delete)
//	relIDs, err := adj.GetAllRelated(nodeID)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, relID := range relIDs {
//	    fmt.Printf("Relationship: %s\n", relID)
//	}
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

// hasAdjPrefix checks if the key has the given prefix.
//
// Parameters:
//   - key: The key to check
//   - prefix: The prefix to look for
//
// Returns true if the key starts with the prefix.
func hasAdjPrefix(key, prefix []byte) bool {
	return len(key) >= len(prefix) && string(key[:len(prefix)]) == string(prefix)
}
