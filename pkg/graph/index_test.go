package graph

import (
	"os"
	"testing"

	"github.com/DotNetAge/gograph/pkg/storage"
)

func TestIndex(t *testing.T) {
	path := "/tmp/gograph_index_test"
	defer os.RemoveAll(path)

	store, err := storage.Open(path)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer store.Close()

	idx := NewIndex(store)
	node := &Node{
		ID:     "node:1",
		Labels: []string{"User"},
		Properties: map[string]PropertyValue{
			"name": NewStringProperty("Alice"),
		},
	}

	batch := store.NewBatch()
	err = idx.BuildLabelIndex(batch, node)
	if err != nil {
		t.Fatalf("failed to build label index: %v", err)
	}
	batch.Commit()

	ids, err := idx.LookupByLabel("User")
	if err != nil {
		t.Fatalf("failed to lookup label: %v", err)
	}
	if len(ids) != 1 || ids[0] != "node:1" {
		t.Errorf("expected node:1, got %v", ids)
	}

	batch = store.NewBatch()
	err = idx.BuildPropertyIndex(batch, node)
	if err != nil {
		t.Fatalf("failed to build property index: %v", err)
	}
	batch.Commit()

	ids, err = idx.LookupByProperty("User", "name", "Alice")
	if err != nil {
		t.Fatalf("failed to lookup property: %v", err)
	}
	if len(ids) != 1 || ids[0] != "node:1" {
		t.Errorf("expected node:1, got %v", ids)
	}

	batch = store.NewBatch()
	idx.RemoveLabelIndex(batch, node)
	idx.RemovePropertyIndex(batch, node)
	batch.Commit()

	ids, err = idx.LookupByLabel("User")
	if err != nil {
		t.Fatalf("failed to lookup label after removal: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("expected no nodes, got %v", ids)
	}

	ids, err = idx.LookupByProperty("User", "name", "Alice")
	if err != nil {
		t.Fatalf("failed to lookup property after removal: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("expected no nodes, got %v", ids)
	}
}

func TestAdjacencyList(t *testing.T) {
	path := "/tmp/gograph_adj_test"
	defer os.RemoveAll(path)

	store, err := storage.Open(path)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer store.Close()

	adj := NewAdjacencyList(store)

	rel := NewRelationship("node:1", "node:2", "KNOWS", map[string]interface{}{"since": 2020})

	batch := store.NewBatch()
	err = adj.AddRelationship(batch, rel)
	if err != nil {
		t.Fatalf("failed to add relationship: %v", err)
	}
	batch.Commit()

	nodes, err := adj.GetRelatedNodes("node:1", "KNOWS", DirectionOutgoing)
	if err != nil {
		t.Fatalf("failed to get related nodes: %v", err)
	}
	if len(nodes) != 1 || nodes[0] != "node:2" {
		t.Errorf("expected node:2, got %v", nodes)
	}

	nodes, err = adj.GetRelatedNodes("node:2", "KNOWS", DirectionIncoming)
	if err != nil {
		t.Fatalf("failed to get incoming nodes: %v", err)
	}
	if len(nodes) != 1 || nodes[0] != "node:1" {
		t.Errorf("expected node:1, got %v", nodes)
	}

	nodes, err = adj.GetRelatedNodes("node:1", "KNOWS", DirectionBoth)
	if err != nil {
		t.Fatalf("failed to get both direction nodes: %v", err)
	}
	if len(nodes) != 1 || nodes[0] != "node:2" {
		t.Errorf("expected node:2, got %v", nodes)
	}

	relIDs, err := adj.GetAllRelated("node:1")
	if err != nil {
		t.Fatalf("failed to get all related: %v", err)
	}
	if len(relIDs) != 1 {
		t.Errorf("expected 1 relationship, got %d", len(relIDs))
	}

	batch = store.NewBatch()
	err = adj.RemoveRelationship(batch, rel)
	if err != nil {
		t.Fatalf("failed to remove relationship: %v", err)
	}
	batch.Commit()

	nodes, err = adj.GetRelatedNodes("node:1", "KNOWS", DirectionOutgoing)
	if err != nil {
		t.Fatalf("failed to get related nodes after removal: %v", err)
	}
	if len(nodes) != 0 {
		t.Errorf("expected no nodes, got %v", nodes)
	}
}
