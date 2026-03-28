package creators

import (
	"os"
	"testing"

	"github.com/DotNetAge/gograph/pkg/cypher/ast"
	"github.com/DotNetAge/gograph/pkg/storage"
	"github.com/DotNetAge/gograph/pkg/tx"
)

func TestCreator(t *testing.T) {
	path := "/tmp/gograph_creator_test"
	defer os.RemoveAll(path)

	store, err := storage.Open(path)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer store.Close()

	mgr := tx.NewManager(store)
	transaction, err := mgr.Begin(false)
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}

	creator := NewCreator(store)

	clause := &ast.CreateClause{
		Pattern: ast.Pattern{
			Elements: []ast.PatternElement{
				{
					Node: &ast.NodePattern{
						Variable: "n",
						Labels:   []string{"User"},
						Properties: map[string]interface{}{
							"name": "Alice",
							"age":  30,
						},
					},
				},
			},
		},
	}

	nodes, rels, err := creator.Execute(transaction, clause)
	if err != nil {
		t.Fatalf("failed to execute create: %v", err)
	}

	if nodes != 1 {
		t.Errorf("expected 1 node, got %d", nodes)
	}

	if rels != 0 {
		t.Errorf("expected 0 relationships, got %d", rels)
	}

	err = transaction.Commit()
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}
}

func TestCreatorWithRelationship(t *testing.T) {
	path := "/tmp/gograph_creator_rel_test"
	defer os.RemoveAll(path)

	store, err := storage.Open(path)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer store.Close()

	mgr := tx.NewManager(store)
	transaction, err := mgr.Begin(false)
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}

	creator := NewCreator(store)

	clause := &ast.CreateClause{
		Pattern: ast.Pattern{
			Elements: []ast.PatternElement{
				{
					Node: &ast.NodePattern{
						Variable: "a",
						Labels:   []string{"User"},
						Properties: map[string]interface{}{
							"name": "Alice",
						},
					},
					Relation: &ast.RelationPattern{
						RelType: "KNOWS",
						Dir:     ast.RelDirOutgoing,
						EndNode: &ast.NodePattern{
							Variable: "b",
							Labels:   []string{"User"},
							Properties: map[string]interface{}{
								"name": "Bob",
							},
						},
					},
				},
			},
		},
	}

	nodes, rels, err := creator.Execute(transaction, clause)
	if err != nil {
		t.Fatalf("failed to execute create: %v", err)
	}

	if nodes != 2 {
		t.Errorf("expected 2 nodes, got %d", nodes)
	}

	if rels != 1 {
		t.Errorf("expected 1 relationship, got %d", rels)
	}

	err = transaction.Commit()
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}
}

func TestCreatorMultipleNodes(t *testing.T) {
	path := "/tmp/gograph_creator_multi_test"
	defer os.RemoveAll(path)

	store, err := storage.Open(path)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer store.Close()

	mgr := tx.NewManager(store)
	transaction, err := mgr.Begin(false)
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}

	creator := NewCreator(store)

	clause := &ast.CreateClause{
		Pattern: ast.Pattern{
			Elements: []ast.PatternElement{
				{
					Node: &ast.NodePattern{
						Variable: "a",
						Labels:   []string{"User"},
						Properties: map[string]interface{}{
							"name": "Alice",
						},
					},
				},
				{
					Node: &ast.NodePattern{
						Variable: "b",
						Labels:   []string{"User"},
						Properties: map[string]interface{}{
							"name": "Bob",
						},
					},
				},
			},
		},
	}

	nodes, rels, err := creator.Execute(transaction, clause)
	if err != nil {
		t.Fatalf("failed to execute create: %v", err)
	}

	if nodes != 2 {
		t.Errorf("expected 2 nodes, got %d", nodes)
	}

	if rels != 0 {
		t.Errorf("expected 0 relationships, got %d", rels)
	}

	err = transaction.Commit()
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}
}
