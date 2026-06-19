package api

import (
	"errors"
	"fmt"

	"github.com/DotNetAge/gograph/pkg/graph"
	"github.com/DotNetAge/gograph/pkg/storage"
)

var (
	ErrNodeNotFound    = errors.New("node not found")
	ErrInvalidNodeID   = errors.New("invalid node id")
	ErrInvalidEdgeData = errors.New("invalid edge data")
)

type GraphStore struct {
	db    *DB
	store *storage.DB
	index *graph.Index
	adj   *graph.AdjacencyList
}

func NewGraphStore(db *DB) *GraphStore {
	store := db.Store()
	return &GraphStore{
		db:    db,
		store: store,
		index: graph.NewIndex(store),
		adj:   graph.NewAdjacencyList(store),
	}
}

type NodeData struct {
	ID         string
	Labels     []string
	Properties map[string]interface{}
}

type EdgeData struct {
	FromNodeID string
	ToNodeID   string
	Type       string
	Properties map[string]interface{}
}

type NeighborResult struct {
	Node *graph.Node
	Edge *graph.Relationship
}

func (gs *GraphStore) UpsertNodes(nodes []*NodeData) error {
	gs.db.Lock()
	defer gs.db.Unlock()
	if gs.db.IsClosedLocked() {
		return ErrDBClosed
	}

	batch := gs.store.NewBatch()
	defer batch.Close()

	for _, nodeData := range nodes {
		if nodeData.ID == "" {
			return ErrInvalidNodeID
		}

		node := &graph.Node{
			ID:         nodeData.ID,
			Labels:     nodeData.Labels,
			Properties: make(map[string]graph.PropertyValue),
		}

		for k, v := range nodeData.Properties {
			node.Properties[k] = graph.ToPropertyValue(v)
		}

		data, err := storage.Marshal(node)
		if err != nil {
			return err
		}

		if err := batch.Put(storage.NodeKey(node.ID), data); err != nil {
			return err
		}

		if err := gs.index.BuildLabelIndex(batch, node); err != nil {
			return err
		}
		if err := gs.index.BuildPropertyIndex(batch, node); err != nil {
			return err
		}
	}

	return batch.Commit()
}

func (gs *GraphStore) UpsertEdges(edges []*EdgeData) error {
	gs.db.Lock()
	defer gs.db.Unlock()
	if gs.db.IsClosedLocked() {
		return ErrDBClosed
	}

	batch := gs.store.NewBatch()
	defer batch.Close()

	rels := make([]*graph.Relationship, 0, len(edges))
	for _, edgeData := range edges {
		if edgeData.FromNodeID == "" || edgeData.ToNodeID == "" {
			return ErrInvalidEdgeData
		}

		rel := graph.NewRelationship(edgeData.FromNodeID, edgeData.ToNodeID, edgeData.Type, edgeData.Properties)
		rels = append(rels, rel)

		relData, err := storage.Marshal(rel)
		if err != nil {
			return err
		}

		if err := batch.Put(storage.RelKey(rel.ID), relData); err != nil {
			return err
		}
	}

	if err := gs.adj.AddRelationships(batch, rels); err != nil {
		return err
	}

	return batch.Commit()
}

func (gs *GraphStore) GetNode(nodeID string) (*graph.Node, error) {
	gs.db.RLock()
	defer gs.db.RUnlock()
	if gs.db.IsClosedLocked() {
		return nil, ErrDBClosed
	}

	if nodeID == "" {
		return nil, ErrInvalidNodeID
	}

	data, err := gs.store.Get(storage.NodeKey(nodeID))
	if err != nil {
		return nil, ErrNodeNotFound
	}

	var node graph.Node
	if err := storage.Unmarshal(data, &node); err != nil {
		return nil, err
	}

	return &node, nil
}

func (gs *GraphStore) GetNeighbors(nodeID string, depth int, limit int) ([]*NeighborResult, error) {
	gs.db.RLock()
	defer gs.db.RUnlock()
	if gs.db.IsClosedLocked() {
		return nil, ErrDBClosed
	}

	if nodeID == "" {
		return nil, ErrInvalidNodeID
	}

	var results []*NeighborResult
	visited := make(map[string]bool)
	visited[nodeID] = true

	currentLevel := []string{nodeID}
	count := 0

	for d := 0; d < depth; d++ {
		var nextLevel []string

		for _, currentID := range currentLevel {
			relIDs, err := gs.adj.GetAllRelated(currentID)
			if err != nil {
				continue
			}

			for _, relID := range relIDs {
				relData, err := gs.store.Get(storage.RelKey(relID))
				if err != nil {
					continue
				}

				var rel graph.Relationship
				if err := storage.Unmarshal(relData, &rel); err != nil {
					continue
				}

				var neighborID string
				if rel.StartNodeID == currentID {
					neighborID = rel.EndNodeID
				} else if rel.EndNodeID == currentID {
					neighborID = rel.StartNodeID
				} else {
					continue
				}

				if visited[neighborID] {
					continue
				}

				visited[neighborID] = true

				nodeData, err := gs.store.Get(storage.NodeKey(neighborID))
				if err != nil {
					continue
				}

				var neighborNode graph.Node
				if err := storage.Unmarshal(nodeData, &neighborNode); err != nil {
					continue
				}

				results = append(results, &NeighborResult{
					Node: &neighborNode,
					Edge: &rel,
				})
				count++

				if limit > 0 && count >= limit {
					return results, nil
				}

				nextLevel = append(nextLevel, neighborID)
			}
		}

		currentLevel = nextLevel
		if len(currentLevel) == 0 {
			break
		}
	}

	return results, nil
}

// GetNeighborsByTypes returns neighbors of a node filtered by specific relationship types.
// This is more efficient than filtering after GetNeighbors because it uses adjacency
// list prefix scanning instead of loading all relationships.
//
// Parameters:
//   - nodeID: The ID of the starting node
//   - depth: The maximum traversal depth (1 = direct neighbors only)
//   - limit: Maximum number of results (0 = unlimited)
//   - relTypes: Only follow edges matching these relationship types (empty = all types)
func (gs *GraphStore) GetNeighborsByTypes(nodeID string, depth int, limit int, relTypes []string) ([]*NeighborResult, error) {
	gs.db.RLock()
	defer gs.db.RUnlock()
	if gs.db.IsClosedLocked() {
		return nil, ErrDBClosed
	}

	if nodeID == "" {
		return nil, ErrInvalidNodeID
	}

	var results []*NeighborResult
	visited := make(map[string]bool)
	visited[nodeID] = true

	currentLevel := []string{nodeID}
	count := 0

	for d := 0; d < depth; d++ {
		var nextLevel []string

		for _, currentID := range currentLevel {
			var relIDs []string
			var err error

			if len(relTypes) > 0 {
				relIDs, err = gs.getRelatedByTypes(currentID, relTypes)
			} else {
				relIDs, err = gs.adj.GetAllRelated(currentID)
			}
			if err != nil {
				continue
			}

			for _, relID := range relIDs {
				relData, err := gs.store.Get(storage.RelKey(relID))
				if err != nil {
					continue
				}

				var rel graph.Relationship
				if err := storage.Unmarshal(relData, &rel); err != nil {
					continue
				}

				var neighborID string
				if rel.StartNodeID == currentID {
					neighborID = rel.EndNodeID
				} else if rel.EndNodeID == currentID {
					neighborID = rel.StartNodeID
				} else {
					continue
				}

				if visited[neighborID] {
					continue
				}

				visited[neighborID] = true

				nodeData, err := gs.store.Get(storage.NodeKey(neighborID))
				if err != nil {
					continue
				}

				var neighborNode graph.Node
				if err := storage.Unmarshal(nodeData, &neighborNode); err != nil {
					continue
				}

				results = append(results, &NeighborResult{
					Node: &neighborNode,
					Edge: &rel,
				})
				count++

				if limit > 0 && count >= limit {
					return results, nil
				}

				nextLevel = append(nextLevel, neighborID)
			}
		}

		currentLevel = nextLevel
		if len(currentLevel) == 0 {
			break
		}
	}

	return results, nil
}

// ListNodes returns all nodes in the graph store.
// Uses storage prefix iteration to list all node: prefixed keys.
//
// Returns a slice of all Node objects, or an error if iteration fails.
func (gs *GraphStore) ListNodes() ([]*graph.Node, error) {
	gs.db.RLock()
	defer gs.db.RUnlock()
	if gs.db.IsClosedLocked() {
		return nil, ErrDBClosed
	}

	iter, err := gs.store.NewIter(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create node iterator: %w", err)
	}
	defer iter.Close()

	prefix := []byte(storage.KeyPrefixNode)
	var nodes []*graph.Node
	for iter.SeekGE(prefix); iter.Valid(); iter.Next() {
		if len(iter.Key()) < len(prefix) || string(iter.Key()[:len(prefix)]) != storage.KeyPrefixNode {
			break
		}
		data := make([]byte, len(iter.Value()))
		copy(data, iter.Value())
		var node graph.Node
		if err := storage.Unmarshal(data, &node); err != nil {
			continue // skip corrupt entries
		}
		nodes = append(nodes, &node)
	}
	return nodes, nil
}

// ListEdges returns all relationships in the graph store.
// Uses storage prefix iteration to list all rel: prefixed keys.
//
// Returns a slice of all Relationship objects, or an error if iteration fails.
func (gs *GraphStore) ListEdges() ([]*graph.Relationship, error) {
	gs.db.RLock()
	defer gs.db.RUnlock()
	if gs.db.IsClosedLocked() {
		return nil, ErrDBClosed
	}

	iter, err := gs.store.NewIter(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create edge iterator: %w", err)
	}
	defer iter.Close()

	prefix := []byte(storage.KeyPrefixRel)
	var edges []*graph.Relationship
	for iter.SeekGE(prefix); iter.Valid(); iter.Next() {
		if len(iter.Key()) < len(prefix) || string(iter.Key()[:len(prefix)]) != storage.KeyPrefixRel {
			break
		}
		data := make([]byte, len(iter.Value()))
		copy(data, iter.Value())
		var rel graph.Relationship
		if err := storage.Unmarshal(data, &rel); err != nil {
			continue // skip corrupt entries
		}
		edges = append(edges, &rel)
	}
	return edges, nil
}

// getRelatedByTypes returns relationship IDs for the given node filtered by relationship types.
func (gs *GraphStore) getRelatedByTypes(nodeID string, relTypes []string) ([]string, error) {
	var relIDs []string

	for _, relType := range relTypes {
		for _, dir := range []string{"out", "in"} {
			key := storage.AdjGroupKey(nodeID, relType, dir)
			data, err := gs.store.Get(key)
			if err != nil {
				continue
			}
			var entries []graph.AdjacencyEntry
			if err := storage.Unmarshal(data, &entries); err != nil {
				continue
			}
			for _, e := range entries {
				relIDs = append(relIDs, e.RelID)
			}
		}
	}

	return relIDs, nil
}
