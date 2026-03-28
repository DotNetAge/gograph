package matchers

import (
	"os"
	"testing"

	"github.com/DotNetAge/gograph/pkg/cypher/ast"
	"github.com/DotNetAge/gograph/pkg/graph"
	"github.com/DotNetAge/gograph/pkg/storage"
)

func setupUndirectedTest(t *testing.T) (*storage.DB, *graph.Index, *Matcher) {
	path := "/tmp/gograph_undirected_test_" + t.Name()
	defer os.RemoveAll(path)

	store, err := storage.Open(path)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}

	index := graph.NewIndex(store)
	matcher := NewMatcher(store, index)

	return store, index, matcher
}

func createUndirectedTestData(t *testing.T, store *storage.DB, index *graph.Index) {
	node1 := &graph.Node{
		ID:     "1",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Alice"),
		},
	}

	node2 := &graph.Node{
		ID:     "2",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Bob"),
		},
	}

	node3 := &graph.Node{
		ID:     "3",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Charlie"),
		},
	}

	nodes := []*graph.Node{node1, node2, node3}
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
		Type:        "KNOWS",
		StartNodeID: node1.ID,
		EndNodeID:   node2.ID,
		Properties:  map[string]graph.PropertyValue{},
	}

	rel2to3 := &graph.Relationship{
		ID:          "rel2",
		Type:        "KNOWS",
		StartNodeID: node2.ID,
		EndNodeID:   node3.ID,
		Properties:  map[string]graph.PropertyValue{},
	}

	rel3to1 := &graph.Relationship{
		ID:          "rel3",
		Type:        "KNOWS",
		StartNodeID: node3.ID,
		EndNodeID:   node1.ID,
		Properties:  map[string]graph.PropertyValue{},
	}

	rels := []*graph.Relationship{rel1to2, rel2to3, rel3to1}
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

func TestMatcherUndirectedRelationship(t *testing.T) {
	store, index, matcher := setupUndirectedTest(t)
	defer store.Close()
	createUndirectedTestData(t, store, index)

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
					Dir:      ast.RelDirBoth,
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
				{Expression: &ast.Identifier{Name: "a"}},
				{Expression: &ast.Identifier{Name: "r"}},
				{Expression: &ast.Identifier{Name: "b"}},
			},
		},
	}

	rows, columns, err := matcher.Execute(matchClause, make(map[string]interface{}), make(map[string]interface{}))
	if err != nil {
		t.Fatalf("failed to execute undirected match: %v", err)
	}

	if len(rows) < 3 {
		t.Errorf("expected at least 3 rows for undirected KNOWS relationships, got %d", len(rows))
	}

	if len(columns) != 3 {
		t.Errorf("expected 3 columns, got %d", len(columns))
	}
}

func TestMatcherUndirectedRelationshipWithSelfLoop(t *testing.T) {
	store, index, matcher := setupUndirectedTest(t)
	defer store.Close()

	node1 := &graph.Node{
		ID:     "1",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Alice"),
		},
	}

	data, err := storage.Marshal(node1)
	if err != nil {
		t.Fatalf("failed to marshal node: %v", err)
	}
	if err := store.Put(storage.NodeKey(node1.ID), data); err != nil {
		t.Fatalf("failed to store node: %v", err)
	}
	if err := index.BuildLabelIndex(store, node1); err != nil {
		t.Fatalf("failed to build label index: %v", err)
	}

	selfRel := &graph.Relationship{
		ID:          "self1",
		Type:        "KNOWS",
		StartNodeID: node1.ID,
		EndNodeID:   node1.ID,
		Properties:  map[string]graph.PropertyValue{},
	}

	data, err = storage.Marshal(selfRel)
	if err != nil {
		t.Fatalf("failed to marshal relationship: %v", err)
	}
	if err := store.Put(storage.RelKey(selfRel.ID), data); err != nil {
		t.Fatalf("failed to store relationship: %v", err)
	}

	outKey := storage.AdjKey(selfRel.StartNodeID, selfRel.Type, "out", selfRel.ID)
	if err := store.Put(outKey, []byte(selfRel.EndNodeID)); err != nil {
		t.Fatalf("failed to store out adjacency: %v", err)
	}

	inKey := storage.AdjKey(selfRel.EndNodeID, selfRel.Type, "in", selfRel.ID)
	if err := store.Put(inKey, []byte(selfRel.StartNodeID)); err != nil {
		t.Fatalf("failed to store in adjacency: %v", err)
	}

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
					Dir:      ast.RelDirBoth,
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
				{Expression: &ast.Identifier{Name: "a"}},
				{Expression: &ast.Identifier{Name: "r"}},
				{Expression: &ast.Identifier{Name: "b"}},
			},
		},
	}

	rows, _, err := matcher.Execute(matchClause, make(map[string]interface{}), make(map[string]interface{}))
	if err != nil {
		t.Fatalf("failed to execute undirected match with self-loop: %v", err)
	}

	selfLoopFound := false
	for _, row := range rows {
		if a, ok := row["a"].(*graph.Node); ok {
			if b, ok := row["b"].(*graph.Node); ok {
				if a.ID == b.ID && a.ID == node1.ID {
					selfLoopFound = true
					break
				}
			}
		}
	}

	if !selfLoopFound {
		t.Error("expected to find self-loop relationship")
	}
}

func TestMatcherUndirectedRelationshipOnlyIncoming(t *testing.T) {
	store, index, matcher := setupUndirectedTest(t)
	defer store.Close()
	createUndirectedTestData(t, store, index)

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
					Dir:      ast.RelDirIncoming,
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
				{Expression: &ast.Identifier{Name: "a"}},
				{Expression: &ast.Identifier{Name: "r"}},
				{Expression: &ast.Identifier{Name: "b"}},
			},
		},
	}

	rows, _, err := matcher.Execute(matchClause, make(map[string]interface{}), make(map[string]interface{}))
	if err != nil {
		t.Fatalf("failed to execute incoming match: %v", err)
	}

	for _, row := range rows {
		if r, ok := row["r"].(*graph.Relationship); ok {
			if a, ok := row["a"].(*graph.Node); ok {
				if r.EndNodeID != a.ID {
					t.Errorf("expected incoming relationship to end at a, got start=%s end=%s", r.StartNodeID, r.EndNodeID)
				}
			}
		}
	}
}

func TestMatcherUndirectedRelationshipOnlyOutgoing(t *testing.T) {
	store, index, matcher := setupUndirectedTest(t)
	defer store.Close()
	createUndirectedTestData(t, store, index)

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
					Dir:      ast.RelDirOutgoing,
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
				{Expression: &ast.Identifier{Name: "a"}},
				{Expression: &ast.Identifier{Name: "r"}},
				{Expression: &ast.Identifier{Name: "b"}},
			},
		},
	}

	rows, _, err := matcher.Execute(matchClause, make(map[string]interface{}), make(map[string]interface{}))
	if err != nil {
		t.Fatalf("failed to execute outgoing match: %v", err)
	}

	for _, row := range rows {
		if r, ok := row["r"].(*graph.Relationship); ok {
			if a, ok := row["a"].(*graph.Node); ok {
				if r.StartNodeID != a.ID {
					t.Errorf("expected outgoing relationship to start at a, got start=%s end=%s", r.StartNodeID, r.EndNodeID)
				}
			}
		}
	}
}

func TestMatcherUndirectedRelationshipMixedDirections(t *testing.T) {
	store, index, matcher := setupUndirectedTest(t)
	defer store.Close()
	createUndirectedTestData(t, store, index)

	pattern := ast.Pattern{
		Elements: []ast.PatternElement{
			{
				Node: &ast.NodePattern{
					Variable: "a",
					Labels:   []string{"User"},
					Properties: map[string]interface{}{
						"name": "Bob",
					},
				},
				Relation: &ast.RelationPattern{
					Variable: "r",
					RelType:  "KNOWS",
					Dir:      ast.RelDirBoth,
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
				{Expression: &ast.Identifier{Name: "a"}},
				{Expression: &ast.Identifier{Name: "r"}},
				{Expression: &ast.Identifier{Name: "b"}},
			},
		},
	}

	rows, _, err := matcher.Execute(matchClause, make(map[string]interface{}), make(map[string]interface{}))
	if err != nil {
		t.Fatalf("failed to execute mixed direction match: %v", err)
	}

	for _, row := range rows {
		if a, ok := row["a"].(*graph.Node); ok {
			if a.Properties["name"].StringValue() != "Bob" {
				t.Error("expected 'a' to be Bob")
			}
		}
	}
}

func TestMatcherUndirectedRelationshipWithoutVariable(t *testing.T) {
	store, index, matcher := setupUndirectedTest(t)
	defer store.Close()
	createUndirectedTestData(t, store, index)

	pattern := ast.Pattern{
		Elements: []ast.PatternElement{
			{
				Node: &ast.NodePattern{
					Variable: "a",
					Labels:   []string{"User"},
				},
				Relation: &ast.RelationPattern{
					RelType: "KNOWS",
					Dir:     ast.RelDirBoth,
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
				{Expression: &ast.Identifier{Name: "a"}},
				{Expression: &ast.Identifier{Name: "b"}},
			},
		},
	}

	rows, _, err := matcher.Execute(matchClause, make(map[string]interface{}), make(map[string]interface{}))
	if err != nil {
		t.Fatalf("failed to execute undirected match without rel variable: %v", err)
	}

	if len(rows) < 3 {
		t.Errorf("expected at least 3 rows, got %d", len(rows))
	}
}

func TestMatcherUndirectedRelationshipWithWhere(t *testing.T) {
	store, index, matcher := setupUndirectedTest(t)
	defer store.Close()
	createUndirectedTestData(t, store, index)

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
					Dir:      ast.RelDirBoth,
					EndNode: &ast.NodePattern{
						Variable: "b",
						Labels:   []string{"User"},
					},
				},
			},
		},
	}

	where := &ast.WhereClause{
		Expression: &ast.ComparisonOp{
			Left: &ast.PropertyLookup{
				Node:     "a",
				Property: "name",
			},
			Operator: "=",
			Right: &ast.Literal{
				Value: "Alice",
			},
		},
	}

	matchClause := &ast.MatchClause{
		Pattern: pattern,
		Where:   where,
		Return: &ast.ReturnClause{
			Items: []ast.ReturnItem{
				{Expression: &ast.Identifier{Name: "a"}},
				{Expression: &ast.Identifier{Name: "b"}},
			},
		},
	}

	rows, _, err := matcher.Execute(matchClause, make(map[string]interface{}), make(map[string]interface{}))
	if err != nil {
		t.Fatalf("failed to execute undirected match with where: %v", err)
	}

	for _, row := range rows {
		if a, ok := row["a"].(*graph.Node); ok {
			if a.Properties["name"].StringValue() != "Alice" {
				t.Errorf("expected name Alice, got %s", a.Properties["name"].StringValue())
			}
		}
	}
}

func TestMatcherUndirectedWithIncomingOnlyData(t *testing.T) {
	store, index, matcher := setupUndirectedTest(t)
	defer store.Close()

	node1 := &graph.Node{
		ID:     "1",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Alice"),
		},
	}

	node2 := &graph.Node{
		ID:     "2",
		Labels: []string{"User"},
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

	rel := &graph.Relationship{
		ID:          "rel1",
		Type:        "KNOWS",
		StartNodeID: node2.ID,
		EndNodeID:   node1.ID,
		Properties:  map[string]graph.PropertyValue{},
	}

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
					Dir:      ast.RelDirBoth,
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
				{Expression: &ast.Identifier{Name: "a"}},
				{Expression: &ast.Identifier{Name: "r"}},
				{Expression: &ast.Identifier{Name: "b"}},
			},
		},
	}

	rows, _, err := matcher.Execute(matchClause, make(map[string]interface{}), make(map[string]interface{}))
	if err != nil {
		t.Fatalf("failed to execute undirected match: %v", err)
	}

	if len(rows) < 1 {
		t.Error("expected at least 1 row for undirected match")
	}
}

func TestMatcherUndirectedRelationshipTypeFilter(t *testing.T) {
	store, index, matcher := setupUndirectedTest(t)
	defer store.Close()

	node1 := &graph.Node{
		ID:     "1",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Alice"),
		},
	}

	node2 := &graph.Node{
		ID:     "2",
		Labels: []string{"User"},
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

	knowsRel := &graph.Relationship{
		ID:          "rel1",
		Type:        "KNOWS",
		StartNodeID: node1.ID,
		EndNodeID:   node2.ID,
		Properties:  map[string]graph.PropertyValue{},
	}

	likesRel := &graph.Relationship{
		ID:          "rel2",
		Type:        "LIKES",
		StartNodeID: node1.ID,
		EndNodeID:   node2.ID,
		Properties:  map[string]graph.PropertyValue{},
	}

	for _, rel := range []*graph.Relationship{knowsRel, likesRel} {
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
					Labels:   []string{"User"},
				},
				Relation: &ast.RelationPattern{
					Variable: "r",
					RelType:  "KNOWS",
					Dir:      ast.RelDirBoth,
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
				{Expression: &ast.Identifier{Name: "a"}},
				{Expression: &ast.Identifier{Name: "r"}},
				{Expression: &ast.Identifier{Name: "b"}},
			},
		},
	}

	rows, _, err := matcher.Execute(matchClause, make(map[string]interface{}), make(map[string]interface{}))
	if err != nil {
		t.Fatalf("failed to execute undirected match with type filter: %v", err)
	}

	if len(rows) != 1 {
		t.Errorf("expected exactly 1 row for KNOWS type filter, got %d", len(rows))
	}

	for _, row := range rows {
		if r, ok := row["r"].(*graph.Relationship); ok {
			if r.Type != "KNOWS" {
				t.Errorf("expected relationship type KNOWS, got %s", r.Type)
			}
		}
	}
}
