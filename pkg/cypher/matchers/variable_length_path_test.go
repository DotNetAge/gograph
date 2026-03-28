package matchers

import (
	"os"
	"testing"

	"github.com/DotNetAge/gograph/pkg/cypher/ast"
	"github.com/DotNetAge/gograph/pkg/graph"
	"github.com/DotNetAge/gograph/pkg/storage"
)

func setupVariableLengthTest(t *testing.T) (*storage.DB, *graph.Index, *Matcher) {
	path := "/tmp/gograph_varpath_test_" + t.Name()
	defer os.RemoveAll(path)

	store, err := storage.Open(path)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}

	index := graph.NewIndex(store)
	matcher := NewMatcher(store, index)

	return store, index, matcher
}

func createVariableLengthTestData(t *testing.T, store *storage.DB, index *graph.Index) {
	node1 := &graph.Node{
		ID:     "1",
		Labels: []string{"Person"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Alice"),
		},
	}

	node2 := &graph.Node{
		ID:     "2",
		Labels: []string{"Person"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Bob"),
		},
	}

	node3 := &graph.Node{
		ID:     "3",
		Labels: []string{"Person"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Charlie"),
		},
	}

	node4 := &graph.Node{
		ID:     "4",
		Labels: []string{"Person"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("David"),
		},
	}

	nodes := []*graph.Node{node1, node2, node3, node4}
	for _, node := range nodes {
		data, err := storage.Marshal(node)
		if err != nil {
			t.Fatalf("failed to marshal node: %v", err)
		}
		if err := store.Put(storage.NodeKey(node.ID), data); err != nil {
			t.Fatalf("failed to store node: %v", err)
		}
		if err := index.BuildLabelIndex(store, node); err != nil {
			t.Fatalf("failed to build label index: %v", err)
		}
	}

	rel1to2 := &graph.Relationship{
		ID:          "rel1",
		Type:        "FRIEND",
		StartNodeID: node1.ID,
		EndNodeID:   node2.ID,
		Properties:  map[string]graph.PropertyValue{},
	}

	rel2to3 := &graph.Relationship{
		ID:          "rel2",
		Type:        "FRIEND",
		StartNodeID: node2.ID,
		EndNodeID:   node3.ID,
		Properties:  map[string]graph.PropertyValue{},
	}

	rel3to4 := &graph.Relationship{
		ID:          "rel3",
		Type:        "FRIEND",
		StartNodeID: node3.ID,
		EndNodeID:   node4.ID,
		Properties:  map[string]graph.PropertyValue{},
	}

	rels := []*graph.Relationship{rel1to2, rel2to3, rel3to4}
	for _, rel := range rels {
		data, err := storage.Marshal(rel)
		if err != nil {
			t.Fatalf("failed to marshal relationship: %v", err)
		}
		if err := store.Put(storage.RelKey(rel.ID), data); err != nil {
			t.Fatalf("failed to store relationship: %v", err)
		}

		outKey := storage.AdjKey(rel.StartNodeID, rel.Type, "out", rel.ID)
		if err := store.Put(outKey, []byte(rel.EndNodeID)); err != nil {
			t.Fatalf("failed to store out adjacency: %v", err)
		}

		inKey := storage.AdjKey(rel.EndNodeID, rel.Type, "in", rel.ID)
		if err := store.Put(inKey, []byte(rel.StartNodeID)); err != nil {
			t.Fatalf("failed to store in adjacency: %v", err)
		}
	}
}

func TestMatcherVariableLengthPathOneHop(t *testing.T) {
	store, index, matcher := setupVariableLengthTest(t)
	defer store.Close()
	createVariableLengthTestData(t, store, index)

	pattern := ast.Pattern{
		Elements: []ast.PatternElement{
			{
				Node: &ast.NodePattern{
					Variable: "a",
					Labels:   []string{"Person"},
				},
				Relation: &ast.RelationPattern{
					Variable: "r",
					RelType:  "FRIEND",
					Dir:      ast.RelDirOutgoing,
					EndNode: &ast.NodePattern{
						Variable: "b",
						Labels:   []string{"Person"},
					},
				},
			},
		},
	}

	matchClause := &ast.MatchClause{
		Pattern: pattern,
		Return: &ast.ReturnClause{
			Items: []ast.ReturnItem{
				{Expression: &ast.Identifier{Name: "a"}},
				{Expression: &ast.Identifier{Name: "r"}},
				{Expression: &ast.Identifier{Name: "b"}},
			},
		},
	}

	rows, _, err := matcher.Execute(matchClause, make(map[string]interface{}), make(map[string]interface{}))
	if err != nil {
		t.Fatalf("failed to execute variable length path match: %v", err)
	}

	if len(rows) != 3 {
		t.Errorf("expected 3 rows for 1-hop FRIEND relationships, got %d", len(rows))
	}
}

func TestMatcherVariableLengthPathWithCycleDetection(t *testing.T) {
	store, index, matcher := setupVariableLengthTest(t)
	defer store.Close()

	node1 := &graph.Node{
		ID:     "1",
		Labels: []string{"Person"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Alice"),
		},
	}

	node2 := &graph.Node{
		ID:     "2",
		Labels: []string{"Person"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Bob"),
		},
	}

	for _, node := range []*graph.Node{node1, node2} {
		data, err := storage.Marshal(node)
		if err != nil {
			t.Fatalf("failed to marshal node: %v", err)
		}
		if err := store.Put(storage.NodeKey(node.ID), data); err != nil {
			t.Fatalf("failed to store node: %v", err)
		}
		if err := index.BuildLabelIndex(store, node); err != nil {
			t.Fatalf("failed to build label index: %v", err)
		}
	}

	rel1to2 := &graph.Relationship{
		ID:          "rel1",
		Type:        "FRIEND",
		StartNodeID: node1.ID,
		EndNodeID:   node2.ID,
		Properties:  map[string]graph.PropertyValue{},
	}

	rel2to1 := &graph.Relationship{
		ID:          "rel2",
		Type:        "FRIEND",
		StartNodeID: node2.ID,
		EndNodeID:   node1.ID,
		Properties:  map[string]graph.PropertyValue{},
	}

	for _, rel := range []*graph.Relationship{rel1to2, rel2to1} {
		data, err := storage.Marshal(rel)
		if err != nil {
			t.Fatalf("failed to marshal relationship: %v", err)
		}
		if err := store.Put(storage.RelKey(rel.ID), data); err != nil {
			t.Fatalf("failed to store relationship: %v", err)
		}

		outKey := storage.AdjKey(rel.StartNodeID, rel.Type, "out", rel.ID)
		if err := store.Put(outKey, []byte(rel.EndNodeID)); err != nil {
			t.Fatalf("failed to store out adjacency: %v", err)
		}

		inKey := storage.AdjKey(rel.EndNodeID, rel.Type, "in", rel.ID)
		if err := store.Put(inKey, []byte(rel.StartNodeID)); err != nil {
			t.Fatalf("failed to store in adjacency: %v", err)
		}
	}

	pattern := ast.Pattern{
		Elements: []ast.PatternElement{
			{
				Node: &ast.NodePattern{
					Variable: "a",
					Labels:   []string{"Person"},
				},
				Relation: &ast.RelationPattern{
					Variable: "r",
					RelType:  "FRIEND",
					Dir:      ast.RelDirOutgoing,
					EndNode: &ast.NodePattern{
						Variable: "b",
						Labels:   []string{"Person"},
					},
				},
			},
		},
	}

	matchClause := &ast.MatchClause{
		Pattern: pattern,
		Return: &ast.ReturnClause{
			Items: []ast.ReturnItem{
				{Expression: &ast.Identifier{Name: "a"}},
				{Expression: &ast.Identifier{Name: "r"}},
				{Expression: &ast.Identifier{Name: "b"}},
			},
		},
	}

	rows, _, err := matcher.Execute(matchClause, make(map[string]interface{}), make(map[string]interface{}))
	if err != nil {
		t.Fatalf("failed to execute match with cycle: %v", err)
	}

	if len(rows) != 2 {
		t.Errorf("expected 2 rows for 1-hop with cycle, got %d", len(rows))
	}
}

