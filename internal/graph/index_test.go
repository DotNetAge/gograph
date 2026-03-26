package graph

import (
	"os"
	"testing"

	"github.com/DotNetAge/gograph/internal/storage"
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

	err = idx.BuildLabelIndex(node)
	if err != nil {
		t.Fatalf("failed to build label index: %v", err)
	}

	ids, err := idx.LookupByLabel("User")
	if err != nil {
		t.Fatalf("failed to lookup label: %v", err)
	}
	if len(ids) != 1 || ids[0] != "node:1" {
		t.Errorf("expected node:1, got %v", ids)
	}

	err = idx.BuildPropertyIndex(node)
	if err != nil {
		t.Fatalf("failed to build property index: %v", err)
	}

	ids, err = idx.LookupByProperty("User", "name", "Alice")
	if err != nil {
		t.Fatalf("failed to lookup property: %v", err)
	}
	if len(ids) != 1 || ids[0] != "node:1" {
		t.Errorf("expected node:1, got %v", ids)
	}

	idx.RemoveLabelIndex(node)
	idx.RemovePropertyIndex(node)
}
