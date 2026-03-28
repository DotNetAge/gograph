package modifiers

import (
	"os"
	"testing"

	"github.com/DotNetAge/gograph/pkg/cypher/ast"
	"github.com/DotNetAge/gograph/pkg/cypher/matchers"
	"github.com/DotNetAge/gograph/pkg/graph"
	"github.com/DotNetAge/gograph/pkg/storage"
	"github.com/DotNetAge/gograph/pkg/tx"
)

func TestModifierExecuteSet(t *testing.T) {
	path := "/tmp/gograph_modifier_set_test"
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

	node := &graph.Node{
		ID:     "node:1",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Alice"),
		},
	}

	data, err := storage.Marshal(node)
	if err != nil {
		t.Fatalf("failed to marshal node: %v", err)
	}
	err = transaction.Put(storage.NodeKey(node.ID), data)
	if err != nil {
		t.Fatalf("failed to put node: %v", err)
	}

	matcher := matchers.NewMatcher(store, graph.NewIndex(store))
	modifier := NewModifier(store, matcher)

	varVars := map[string]interface{}{
		"n": node,
	}

	clause := &ast.SetClause{
		Assignments: []ast.Assignment{
			{
				Property: ast.PropertyAccess{
					Node:     "n",
					Property: "name",
				},
				Value: &ast.Literal{
					Value: "Bob",
				},
			},
		},
	}

	affected, err := modifier.ExecuteSet(transaction, clause, varVars, nil)
	if err != nil {
		t.Fatalf("failed to execute set: %v", err)
	}

	if affected != 1 {
		t.Errorf("expected 1 affected node, got %d", affected)
	}

	if node.Properties["name"].StringValue() != "Bob" {
		t.Errorf("expected name 'Bob', got %s", node.Properties["name"].StringValue())
	}

	err = transaction.Commit()
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}
}

func TestModifierExecuteDelete(t *testing.T) {
	path := "/tmp/gograph_modifier_delete_test"
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

	node := &graph.Node{
		ID:     "node:1",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Alice"),
		},
	}

	data, err := storage.Marshal(node)
	if err != nil {
		t.Fatalf("failed to marshal node: %v", err)
	}
	err = transaction.Put(storage.NodeKey(node.ID), data)
	if err != nil {
		t.Fatalf("failed to put node: %v", err)
	}

	matcher := matchers.NewMatcher(store, graph.NewIndex(store))
	modifier := NewModifier(store, matcher)

	varVars := map[string]interface{}{
		"n": node,
	}

	clause := &ast.DeleteClause{
		Expressions: []ast.Expression{
			&ast.PropertyLookup{
				Node: "n",
			},
		},
	}

	nodes, rels, err := modifier.ExecuteDelete(transaction, clause, varVars, nil)
	if err != nil {
		t.Fatalf("failed to execute delete: %v", err)
	}

	if nodes != 1 {
		t.Errorf("expected 1 deleted node, got %d", nodes)
	}

	if rels != 0 {
		t.Errorf("expected 0 deleted relationships, got %d", rels)
	}

	err = transaction.Commit()
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}
}

func TestModifierExecuteRemoveLabel(t *testing.T) {
	path := "/tmp/gograph_modifier_remove_label_test"
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

	node := &graph.Node{
		ID:     "node:1",
		Labels: []string{"User", "VIP"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Alice"),
		},
	}

	data, err := storage.Marshal(node)
	if err != nil {
		t.Fatalf("failed to marshal node: %v", err)
	}
	err = transaction.Put(storage.NodeKey(node.ID), data)
	if err != nil {
		t.Fatalf("failed to put node: %v", err)
	}

	matcher := matchers.NewMatcher(store, graph.NewIndex(store))
	modifier := NewModifier(store, matcher)

	varVars := map[string]interface{}{
		"n": node,
	}

	clause := &ast.RemoveClause{
		Removals: []ast.RemoveItem{
			{
				Type:  ast.RemoveItemTypeLabel,
				Label: "VIP",
				Property: ast.PropertyAccess{
					Node: "n",
				},
			},
		},
	}

	affected, err := modifier.ExecuteRemove(transaction, clause, varVars, nil)
	if err != nil {
		t.Fatalf("failed to execute remove: %v", err)
	}

	if affected != 1 {
		t.Errorf("expected 1 affected node, got %d", affected)
	}

	if node.HasLabel("VIP") {
		t.Error("expected VIP label to be removed")
	}

	if !node.HasLabel("User") {
		t.Error("expected User label to remain")
	}

	err = transaction.Commit()
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}
}

func TestModifierExecuteRemoveProperty(t *testing.T) {
	path := "/tmp/gograph_modifier_remove_prop_test"
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

	node := &graph.Node{
		ID:     "node:1",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Alice"),
			"age":  graph.NewIntProperty(30),
		},
	}

	data, err := storage.Marshal(node)
	if err != nil {
		t.Fatalf("failed to marshal node: %v", err)
	}
	err = transaction.Put(storage.NodeKey(node.ID), data)
	if err != nil {
		t.Fatalf("failed to put node: %v", err)
	}

	matcher := matchers.NewMatcher(store, graph.NewIndex(store))
	modifier := NewModifier(store, matcher)

	varVars := map[string]interface{}{
		"n": node,
	}

	clause := &ast.RemoveClause{
		Removals: []ast.RemoveItem{
			{
				Type: ast.RemoveItemTypeProperty,
				Property: ast.PropertyAccess{
					Node:     "n",
					Property: "age",
				},
			},
		},
	}

	affected, err := modifier.ExecuteRemove(transaction, clause, varVars, nil)
	if err != nil {
		t.Fatalf("failed to execute remove: %v", err)
	}

	if affected != 1 {
		t.Errorf("expected 1 affected node, got %d", affected)
	}

	if _, exists := node.Properties["age"]; exists {
		t.Error("expected age property to be removed")
	}

	if _, exists := node.Properties["name"]; !exists {
		t.Error("expected name property to remain")
	}

	err = transaction.Commit()
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}
}

func TestResolveValue(t *testing.T) {
	modifier := NewModifier(nil, nil)

	tests := []struct {
		name     string
		expr     ast.Expression
		params   map[string]interface{}
		expected interface{}
	}{
		{
			name:     "literal",
			expr:     &ast.Literal{Value: "test"},
			params:   nil,
			expected: "test",
		},
		{
			name:     "parameter",
			expr:     &ast.Identifier{Name: "$param"},
			params:   map[string]interface{}{"param": 42},
			expected: 42,
		},
		{
			name:     "unknown parameter",
			expr:     &ast.Identifier{Name: "$unknown"},
			params:   map[string]interface{}{},
			expected: nil,
		},
		{
			name:     "default case",
			expr:     &ast.PropertyLookup{Node: "n", Property: "prop"},
			params:   nil,
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := modifier.resolveValue(tc.expr, tc.params)
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestModifierExecuteDeleteWithRelationship(t *testing.T) {
	path := "/tmp/gograph_modifier_delete_rel_test"
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

	// Create relationship
	rel := &graph.Relationship{
		ID:          "rel:1",
		Type:        "KNOWS",
		StartNodeID: "node:1",
		EndNodeID:   "node:2",
	}

	data, err := storage.Marshal(rel)
	if err != nil {
		t.Fatalf("failed to marshal relationship: %v", err)
	}
	err = transaction.Put(storage.RelKey(rel.ID), data)
	if err != nil {
		t.Fatalf("failed to put relationship: %v", err)
	}

	// Create adjacency list entries
	adj := graph.NewAdjacencyList(store)
	if err := adj.AddRelationship(transaction, rel); err != nil {
		t.Fatalf("failed to add relationship to adjacency list: %v", err)
	}

	matcher := matchers.NewMatcher(store, graph.NewIndex(store))
	modifier := NewModifier(store, matcher)

	varVars := map[string]interface{}{
		"r": rel,
	}

	clause := &ast.DeleteClause{
		Expressions: []ast.Expression{
			&ast.RelationVariable{Name: "r"},
		},
	}

	nodes, rels, err := modifier.ExecuteDelete(transaction, clause, varVars, nil)
	if err != nil {
		t.Fatalf("failed to execute delete: %v", err)
	}

	if nodes != 0 {
		t.Errorf("expected 0 deleted nodes, got %d", nodes)
	}

	if rels != 1 {
		t.Errorf("expected 1 deleted relationship, got %d", rels)
	}

	err = transaction.Commit()
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}
}

func TestModifierExecuteDeleteWithDetach(t *testing.T) {
	path := "/tmp/gograph_modifier_delete_detach_test"
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

	// Create nodes
	node1 := &graph.Node{
		ID:     "node:1",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Alice"),
		},
	}

	node2 := &graph.Node{
		ID:     "node:2",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Bob"),
		},
	}

	data1, err := storage.Marshal(node1)
	if err != nil {
		t.Fatalf("failed to marshal node1: %v", err)
	}
	err = transaction.Put(storage.NodeKey(node1.ID), data1)
	if err != nil {
		t.Fatalf("failed to put node1: %v", err)
	}

	data2, err := storage.Marshal(node2)
	if err != nil {
		t.Fatalf("failed to marshal node2: %v", err)
	}
	err = transaction.Put(storage.NodeKey(node2.ID), data2)
	if err != nil {
		t.Fatalf("failed to put node2: %v", err)
	}

	// Create relationship
	rel := &graph.Relationship{
		ID:          "rel:1",
		Type:        "KNOWS",
		StartNodeID: node1.ID,
		EndNodeID:   node2.ID,
	}

	dataRel, err := storage.Marshal(rel)
	if err != nil {
		t.Fatalf("failed to marshal relationship: %v", err)
	}
	err = transaction.Put(storage.RelKey(rel.ID), dataRel)
	if err != nil {
		t.Fatalf("failed to put relationship: %v", err)
	}

	// Create adjacency list entries
	adj := graph.NewAdjacencyList(store)
	if err := adj.AddRelationship(transaction, rel); err != nil {
		t.Fatalf("failed to add relationship to adjacency list: %v", err)
	}

	matcher := matchers.NewMatcher(store, graph.NewIndex(store))
	modifier := NewModifier(store, matcher)

	varVars := map[string]interface{}{
		"n": node1,
	}

	clause := &ast.DeleteClause{
		Expressions: []ast.Expression{
			&ast.PropertyLookup{
				Node: "n",
			},
		},
		Detach: true,
	}

	nodes, rels, err := modifier.ExecuteDelete(transaction, clause, varVars, nil)
	if err != nil {
		t.Fatalf("failed to execute delete: %v", err)
	}

	if nodes != 1 {
		t.Errorf("expected 1 deleted node, got %d", nodes)
	}

	if rels != 0 {
		t.Errorf("expected 0 deleted relationships (detach counts are handled internally), got %d", rels)
	}

	err = transaction.Commit()
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}
}
