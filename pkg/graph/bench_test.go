package graph

import (
	"fmt"
	"os"
	"testing"

	"github.com/DotNetAge/gograph/pkg/storage"
)

func BenchmarkBuildPropertyIndex(b *testing.B) {
	path := "/tmp/gograph_bench_index"
	defer os.RemoveAll(path)

	store, err := storage.Open(path)
	if err != nil {
		b.Fatalf("failed to open store: %v", err)
	}
	defer store.Close()

	idx := NewIndex(store)
	node := &Node{
		ID:     "node:bench",
		Labels: []string{"User", "Admin", "Verified"},
		Properties: map[string]PropertyValue{
			"name":  NewStringProperty("Alice"),
			"email": NewStringProperty("alice@example.com"),
			"age":   NewIntProperty(30),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batch := store.NewBatch()
		_ = idx.BuildPropertyIndex(batch, node)
		_ = batch.Commit()
		batch = store.NewBatch()
		_ = idx.RemovePropertyIndex(batch, node)
		_ = batch.Commit()
	}
}

func BenchmarkAdjacencyListAddRelationship(b *testing.B) {
	path := "/tmp/gograph_bench_adj"
	defer os.RemoveAll(path)

	store, err := storage.Open(path)
	if err != nil {
		b.Fatalf("failed to open store: %v", err)
	}
	defer store.Close()

	adj := NewAdjacencyList(store)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batch := store.NewBatch()
		rel := NewRelationship(fmt.Sprintf("node:%d", i), fmt.Sprintf("node:%d", i+1), "KNOWS", nil)
		_ = adj.AddRelationship(batch, rel)
		_ = batch.Commit()
	}
}

func BenchmarkAdjacencyListGetRelatedNodes(b *testing.B) {
	path := "/tmp/gograph_bench_adj_get"
	defer os.RemoveAll(path)

	store, err := storage.Open(path)
	if err != nil {
		b.Fatalf("failed to open store: %v", err)
	}
	defer store.Close()

	adj := NewAdjacencyList(store)

	// Pre-populate with 100 relationships
	batch := store.NewBatch()
	for i := 0; i < 100; i++ {
		rel := NewRelationship("node:center", fmt.Sprintf("node:%d", i), "KNOWS", nil)
		_ = adj.AddRelationship(batch, rel)
	}
	_ = batch.Commit()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = adj.GetRelatedNodes("node:center", "KNOWS", DirectionOutgoing)
	}
}

func BenchmarkMarshalNode(b *testing.B) {
	node := &Node{
		ID:     "node:bench",
		Labels: []string{"User"},
		Properties: map[string]PropertyValue{
			"name":  NewStringProperty("Alice"),
			"email": NewStringProperty("alice@example.com"),
			"age":   NewIntProperty(30),
			"score": NewFloatProperty(95.5),
			"active": NewBoolProperty(true),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, _ := storage.Marshal(node)
		_ = data
	}
}

func BenchmarkUnmarshalNode(b *testing.B) {
	node := &Node{
		ID:     "node:bench",
		Labels: []string{"User"},
		Properties: map[string]PropertyValue{
			"name":  NewStringProperty("Alice"),
			"email": NewStringProperty("alice@example.com"),
			"age":   NewIntProperty(30),
			"score": NewFloatProperty(95.5),
			"active": NewBoolProperty(true),
		},
	}
	data, _ := storage.Marshal(node)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var n Node
		_ = storage.Unmarshal(data, &n)
	}
}
