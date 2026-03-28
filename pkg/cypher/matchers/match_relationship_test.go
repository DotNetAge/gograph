package matchers

import (
	"os"
	"testing"

	"github.com/DotNetAge/gograph/pkg/cypher/ast"
	"github.com/DotNetAge/gograph/pkg/graph"
	"github.com/DotNetAge/gograph/pkg/storage"
)

func TestExecuteWithRelationshipMatch(t *testing.T) {
	// Create a temporary directory for storage
	path := "/tmp/gograph_relationship_test"
	defer os.RemoveAll(path)

	// Create storage and index
	store, err := storage.Open(path)
	if err != nil {
		t.Fatalf("failed to open storage: %v", err)
	}
	defer store.Close()

	index := graph.NewIndex(store)

	// Create start node
	startNode := &graph.Node{
		ID:     "1",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Alice"),
		},
	}

	// Create end node
	endNode := &graph.Node{
		ID:     "2",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Bob"),
		},
	}

	// Create relationship
	rel := &graph.Relationship{
		ID:          "1",
		Type:        "KNOWS",
		StartNodeID: startNode.ID,
		EndNodeID:   endNode.ID,
		Properties:  map[string]graph.PropertyValue{},
	}

	// Store nodes and relationship
	data, err := storage.Marshal(startNode)
	if err != nil {
		t.Fatalf("failed to marshal start node: %v", err)
	}
	if err := store.Put(storage.NodeKey(startNode.ID), data); err != nil {
		t.Fatalf("failed to store start node: %v", err)
	}

	data, err = storage.Marshal(endNode)
	if err != nil {
		t.Fatalf("failed to marshal end node: %v", err)
	}
	if err := store.Put(storage.NodeKey(endNode.ID), data); err != nil {
		t.Fatalf("failed to store end node: %v", err)
	}

	data, err = storage.Marshal(rel)
	if err != nil {
		t.Fatalf("failed to marshal relationship: %v", err)
	}
	if err := store.Put(storage.RelKey(rel.ID), data); err != nil {
		t.Fatalf("failed to store relationship: %v", err)
	}

	// Build adjacency list for the relationship
	adjKey := storage.AdjKey(startNode.ID, rel.Type, "out", rel.ID)
	if err := store.Put(adjKey, []byte(endNode.ID)); err != nil {
		t.Fatalf("failed to store adjacency key: %v", err)
	}

	adjKey = storage.AdjKey(endNode.ID, rel.Type, "in", rel.ID)
	if err := store.Put(adjKey, []byte(startNode.ID)); err != nil {
		t.Fatalf("failed to store reverse adjacency key: %v", err)
	}

	// Build label indexes
	if err := index.BuildLabelIndex(store, startNode); err != nil {
		t.Fatalf("failed to build label index for start node: %v", err)
	}
	if err := index.BuildLabelIndex(store, endNode); err != nil {
		t.Fatalf("failed to build label index for end node: %v", err)
	}

	// Create matcher
	matcher := NewMatcher(store, index)

	// Create a MATCH clause with relationship
	pattern := ast.Pattern{
		Elements: []ast.PatternElement{
			{
				Node: &ast.NodePattern{
					Variable: "a",
					Labels:   []string{"User"},
				},
				Relation: &ast.RelationPattern{
					Variable: "r",
					RelType:  "KNOWS",
					EndNode: &ast.NodePattern{
						Variable: "b",
						Labels:   []string{"User"},
					},
				},
			},
		},
	}

	matchClause := &ast.MatchClause{
		Pattern: pattern,
		Return: &ast.ReturnClause{
			Items: []ast.ReturnItem{
				{
					Expression: &ast.Identifier{Name: "a"},
				},
				{
					Expression: &ast.Identifier{Name: "r"},
				},
				{
					Expression: &ast.Identifier{Name: "b"},
				},
			},
		},
	}

	// Execute the match
	rows, columns, err := matcher.Execute(matchClause, make(map[string]interface{}), make(map[string]interface{}))
	if err != nil {
		t.Fatalf("failed to execute match: %v", err)
	}

	// Check results
	if len(rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(rows))
	}

	if len(columns) != 3 || columns[0] != "a" || columns[1] != "r" || columns[2] != "b" {
		t.Errorf("expected columns ['a', 'r', 'b'], got %v", columns)
	}

	if _, ok := rows[0]["a"]; !ok {
		t.Error("expected 'a' in row")
	}

	if _, ok := rows[0]["r"]; !ok {
		t.Error("expected 'r' in row")
	}

	if _, ok := rows[0]["b"]; !ok {
		t.Error("expected 'b' in row")
	}
}

func TestExecuteWithWhereClause(t *testing.T) {
	// Create a temporary directory for storage
	path := "/tmp/gograph_where_test"
	defer os.RemoveAll(path)

	// Create storage and index
	store, err := storage.Open(path)
	if err != nil {
		t.Fatalf("failed to open storage: %v", err)
	}
	defer store.Close()

	index := graph.NewIndex(store)

	// Create node
	node := &graph.Node{
		ID:     "1",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Alice"),
			"age":  graph.NewIntProperty(30),
		},
	}

	// Store node
	data, err := storage.Marshal(node)
	if err != nil {
		t.Fatalf("failed to marshal node: %v", err)
	}
	if err := store.Put(storage.NodeKey(node.ID), data); err != nil {
		t.Fatalf("failed to store node: %v", err)
	}

	// Build label index
	if err := index.BuildLabelIndex(store, node); err != nil {
		t.Fatalf("failed to build label index: %v", err)
	}

	// Create matcher
	matcher := NewMatcher(store, index)

	// Create a MATCH clause with WHERE
	pattern := ast.Pattern{
		Elements: []ast.PatternElement{
			{
				Node: &ast.NodePattern{
					Variable: "n",
					Labels:   []string{"User"},
				},
			},
		},
	}

	// Create WHERE clause
	where := &ast.WhereClause{
		Expression: &ast.ComparisonOp{
			Left: &ast.PropertyLookup{
				Node:     "n",
				Property: "age",
			},
			Operator: ">",
			Right: &ast.Literal{
				Value: 25,
			},
		},
	}

	matchClause := &ast.MatchClause{
		Pattern: pattern,
		Where:   where,
		Return: &ast.ReturnClause{
			Items: []ast.ReturnItem{
				{
					Expression: &ast.Identifier{Name: "n"},
				},
			},
		},
	}

	// Execute the match
	rows, columns, err := matcher.Execute(matchClause, make(map[string]interface{}), make(map[string]interface{}))
	if err != nil {
		t.Fatalf("failed to execute match: %v", err)
	}

	// Check results
	if len(rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(rows))
	}

	if len(columns) != 1 || columns[0] != "n" {
		t.Errorf("expected column 'n', got %v", columns)
	}

	if _, ok := rows[0]["n"]; !ok {
		t.Error("expected 'n' in row")
	}

	// Test WHERE clause that doesn't match
	where.Expression = &ast.ComparisonOp{
		Left: &ast.PropertyLookup{
			Node:     "n",
			Property: "age",
		},
		Operator: ">",
		Right: &ast.Literal{
			Value: 35,
		},
	}

	rows, columns, err = matcher.Execute(matchClause, make(map[string]interface{}), make(map[string]interface{}))
	if err != nil {
		t.Fatalf("failed to execute match: %v", err)
	}

	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}
}
