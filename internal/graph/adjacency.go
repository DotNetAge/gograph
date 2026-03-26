package graph

import (
	"github.com/DotNetAge/gograph/internal/storage"
)

type AdjacencyList struct {
	store *storage.DB
}

func NewAdjacencyList(store *storage.DB) *AdjacencyList {
	return &AdjacencyList{store: store}
}

func (adj *AdjacencyList) BuildFromRelationship(rel *Relationship) error {
	outKey := storage.AdjKey(rel.StartNodeID, rel.Type, "out")
	if err := adj.store.Put(outKey, []byte(rel.EndNodeID)); err != nil {
		return err
	}

	inKey := storage.AdjKey(rel.EndNodeID, rel.Type, "in")
	if err := adj.store.Put(inKey, []byte(rel.StartNodeID)); err != nil {
		return err
	}

	return nil
}

func (adj *AdjacencyList) RemoveRelationship(rel *Relationship) error {
	outKey := storage.AdjKey(rel.StartNodeID, rel.Type, "out")
	if err := adj.store.Delete(outKey); err != nil {
		return err
	}

	inKey := storage.AdjKey(rel.EndNodeID, rel.Type, "in")
	if err := adj.store.Delete(inKey); err != nil {
		return err
	}

	return nil
}

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
		prefix := storage.AdjKeyPrefixNodeAndType(nodeID, relType)
		iter, err := adj.store.NewIter(nil)
		if err != nil {
			return nil, err
		}

		for iter.SeekGE(prefix); iter.Valid(); iter.Next() {
			key := iter.Key()
			if !hasAdjPrefix(key, prefix) {
				break
			}
			if direction == DirectionBoth || string(key)[len(key)-2:len(key)-1] == dir {
				nodeIDs = append(nodeIDs, string(iter.Value()))
			}
		}
		iter.Close()
	}

	return nodeIDs, nil
}

func (adj *AdjacencyList) GetAllRelated(nodeID string) ([]string, error) {
	var nodeIDs []string
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
		nodeIDs = append(nodeIDs, string(iter.Value()))
	}

	return nodeIDs, nil
}

func hasAdjPrefix(key, prefix []byte) bool {
	return len(key) >= len(prefix) && string(key[:len(prefix)]) == string(prefix)
}
