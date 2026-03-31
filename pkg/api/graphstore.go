package api

import (
	"errors"

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

		for _, label := range node.Labels {
			key := storage.LabelKey(label, node.ID)
			if err := batch.Put(key, []byte(node.ID)); err != nil {
				return err
			}
		}

		for _, label := range node.Labels {
			for propName, propValue := range node.Properties {
				encodedValue := graph.EncodePropertyValue(propValue)
				key := storage.PropertyKey(label, propName, encodedValue)
				if err := batch.Put(key, []byte(node.ID)); err != nil {
					return err
				}
			}
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

	for _, edgeData := range edges {
		if edgeData.FromNodeID == "" || edgeData.ToNodeID == "" {
			return ErrInvalidEdgeData
		}

		rel := graph.NewRelationship(edgeData.FromNodeID, edgeData.ToNodeID, edgeData.Type, edgeData.Properties)

		relData, err := storage.Marshal(rel)
		if err != nil {
			return err
		}

		if err := batch.Put(storage.RelKey(rel.ID), relData); err != nil {
			return err
		}

		outKey := storage.AdjKey(rel.StartNodeID, rel.Type, "out", rel.ID)
		if err := batch.Put(outKey, []byte(rel.EndNodeID)); err != nil {
			return err
		}

		inKey := storage.AdjKey(rel.EndNodeID, rel.Type, "in", rel.ID)
		if err := batch.Put(inKey, []byte(rel.StartNodeID)); err != nil {
			return err
		}
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
