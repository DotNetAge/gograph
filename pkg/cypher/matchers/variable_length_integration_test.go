package matchers

import (
	"os"
	"testing"

	"github.com/DotNetAge/gograph/pkg/cypher/ast"
	"github.com/DotNetAge/gograph/pkg/graph"
	"github.com/DotNetAge/gograph/pkg/storage"
)

func setupVariableLengthPathTest(t *testing.T) (*storage.DB, *graph.Index, *Matcher) {
	path := "/tmp/gograph_varlen_test_" + t.Name()
	defer os.RemoveAll(path)

	store, err := storage.Open(path)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}

	index := graph.NewIndex(store)
	matcher := NewMatcher(store, index)

	return store, index, matcher
}

func createVariableLengthPathTestData(t *testing.T, store *storage.DB, index *graph.Index) {
	nodes := []*graph.Node{
		{ID: "1", Labels: []string{"Person"}, Properties: map[string]graph.PropertyValue{"name": graph.NewStringProperty("Alice")}},
		{ID: "2", Labels: []string{"Person"}, Properties: map[string]graph.PropertyValue{"name": graph.NewStringProperty("Bob")}},
		{ID: "3", Labels: []string{"Person"}, Properties: map[string]graph.PropertyValue{"name": graph.NewStringProperty("Charlie")}},
		{ID: "4", Labels: []string{"Person"}, Properties: map[string]graph.PropertyValue{"name": graph.NewStringProperty("David")}},
	}

	for _, node := range nodes {
		data, _ := storage.Marshal(node)
		store.Put(storage.NodeKey(node.ID), data)
		index.BuildLabelIndex(store, node)
	}

	rels := []*graph.Relationship{
		{ID: "r1", Type: "FRIEND", StartNodeID: "1", EndNodeID: "2"},
		{ID: "r2", Type: "FRIEND", StartNodeID: "2", EndNodeID: "3"},
		{ID: "r3", Type: "FRIEND", StartNodeID: "3", EndNodeID: "4"},
	}

	for _, rel := range rels {
		data, _ := storage.Marshal(rel)
		store.Put(storage.RelKey(rel.ID), data)
		store.Put(storage.AdjKey(rel.StartNodeID, rel.Type, "out", rel.ID), []byte(rel.EndNodeID))
		store.Put(storage.AdjKey(rel.EndNodeID, rel.Type, "in", rel.ID), []byte(rel.StartNodeID))
	}
}

func TestMatcherVariableLengthPathTwoHops(t *testing.T) {
	store, index, matcher := setupVariableLengthPathTest(t)
	defer store.Close()
	createVariableLengthPathTestData(t, store, index)

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
					MinHops:  2,
					MaxHops:  2,
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
				{Expression: &ast.Identifier{Name: "b"}},
			},
		},
	}

	rows, _, err := matcher.Execute(matchClause, make(map[string]interface{}), make(map[string]interface{}))
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	if len(rows) == 0 {
		t.Error("expected at least one path with 2 hops")
	}

	for _, row := range rows {
		if a, ok := row["a"].(*graph.Node); ok {
			if b, ok := row["b"].(*graph.Node); ok {
				t.Logf("Path from %s to %s (2 hops)", a.Properties["name"].StringValue(), b.Properties["name"].StringValue())
			}
		}
	}
}

func TestMatcherVariableLengthPathMin2Max3(t *testing.T) {
	store, index, matcher := setupVariableLengthPathTest(t)
	defer store.Close()
	createVariableLengthPathTestData(t, store, index)

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
					MinHops:  2,
					MaxHops:  3,
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
				{Expression: &ast.Identifier{Name: "b"}},
			},
		},
	}

	rows, _, err := matcher.Execute(matchClause, make(map[string]interface{}), make(map[string]interface{}))
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	if len(rows) == 0 {
		t.Error("expected at least one path with 2-3 hops")
	}

	t.Logf("Found %d paths with 2-3 hops", len(rows))
}

func TestMatcherVariableLengthPathThreeHops(t *testing.T) {
	store, index, matcher := setupVariableLengthPathTest(t)
	defer store.Close()
	createVariableLengthPathTestData(t, store, index)

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
					MinHops:  3,
					MaxHops:  3,
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
				{Expression: &ast.Identifier{Name: "b"}},
			},
		},
	}

	rows, _, err := matcher.Execute(matchClause, make(map[string]interface{}), make(map[string]interface{}))
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	if len(rows) == 0 {
		t.Error("expected at least one path with 3 hops")
	}

	for _, row := range rows {
		if a, ok := row["a"].(*graph.Node); ok {
			if b, ok := row["b"].(*graph.Node); ok {
				t.Logf("Path from %s to %s (3 hops)", a.Properties["name"].StringValue(), b.Properties["name"].StringValue())
			}
		}
	}
}

func TestMatcherVariableLengthPathWithCycle(t *testing.T) {
	store, index, matcher := setupVariableLengthPathTest(t)
	defer store.Close()

	nodes := []*graph.Node{
		{ID: "1", Labels: []string{"Person"}, Properties: map[string]graph.PropertyValue{"name": graph.NewStringProperty("Alice")}},
		{ID: "2", Labels: []string{"Person"}, Properties: map[string]graph.PropertyValue{"name": graph.NewStringProperty("Bob")}},
	}

	for _, node := range nodes {
		data, _ := storage.Marshal(node)
		store.Put(storage.NodeKey(node.ID), data)
		index.BuildLabelIndex(store, node)
	}

	rels := []*graph.Relationship{
		{ID: "r1", Type: "FRIEND", StartNodeID: "1", EndNodeID: "2"},
		{ID: "r2", Type: "FRIEND", StartNodeID: "2", EndNodeID: "1"},
	}

	for _, rel := range rels {
		data, _ := storage.Marshal(rel)
		store.Put(storage.RelKey(rel.ID), data)
		store.Put(storage.AdjKey(rel.StartNodeID, rel.Type, "out", rel.ID), []byte(rel.EndNodeID))
		store.Put(storage.AdjKey(rel.EndNodeID, rel.Type, "in", rel.ID), []byte(rel.StartNodeID))
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
					MinHops:  2,
					MaxHops:  3,
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
				{Expression: &ast.Identifier{Name: "b"}},
			},
		},
	}

	rows, _, err := matcher.Execute(matchClause, make(map[string]interface{}), make(map[string]interface{}))
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	t.Logf("Found %d paths (should not include cycles due to visited tracking)", len(rows))
}

func TestMatcherVariableLengthPathWithEndNodeFilter(t *testing.T) {
	store, index, matcher := setupVariableLengthPathTest(t)
	defer store.Close()
	createVariableLengthPathTestData(t, store, index)

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
					MinHops:  2,
					MaxHops:  3,
					EndNode: &ast.NodePattern{
						Variable: "b",
						Labels:   []string{"Person"},
						Properties: map[string]interface{}{
							"name": "David",
						},
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
		t.Fatalf("failed to execute: %v", err)
	}

	for _, row := range rows {
		if b, ok := row["b"].(*graph.Node); ok {
			if b.Properties["name"].StringValue() != "David" {
				t.Errorf("expected end node name David, got %s", b.Properties["name"].StringValue())
			}
		}
	}
}

func TestMatcherVariableLengthPathWithoutVariable(t *testing.T) {
	store, index, matcher := setupVariableLengthPathTest(t)
	defer store.Close()
	createVariableLengthPathTestData(t, store, index)

	pattern := ast.Pattern{
		Elements: []ast.PatternElement{
			{
				Node: &ast.NodePattern{
					Variable: "a",
					Labels:   []string{"Person"},
				},
				Relation: &ast.RelationPattern{
					RelType: "FRIEND",
					Dir:     ast.RelDirOutgoing,
					MinHops: 2,
					MaxHops: 2,
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
				{Expression: &ast.Identifier{Name: "b"}},
			},
		},
	}

	rows, _, err := matcher.Execute(matchClause, make(map[string]interface{}), make(map[string]interface{}))
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	if len(rows) == 0 {
		t.Error("expected at least one path with 2 hops")
	}
}
