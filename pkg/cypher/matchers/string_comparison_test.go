package matchers

import (
	"os"
	"testing"

	"github.com/DotNetAge/gograph/pkg/cypher/ast"
	"github.com/DotNetAge/gograph/pkg/graph"
	"github.com/DotNetAge/gograph/pkg/storage"
)

func TestEvaluateExpressionStringGreaterThan(t *testing.T) {
	matcher := NewMatcher(nil, nil)

	node := &graph.Node{
		ID:     "1",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Zoe"),
		},
	}

	path := map[string]interface{}{
		"n": node,
	}

	comparison := &ast.ComparisonOp{
		Left: &ast.PropertyLookup{
			Node:     "n",
			Property: "name",
		},
		Operator: ">",
		Right: &ast.Literal{
			Value: "Alice",
		},
	}

	result := matcher.EvaluateExpression(path, comparison, make(map[string]interface{}))
	if !result {
		t.Error("expected 'Zoe' > 'Alice' to be true")
	}
}

func TestEvaluateExpressionStringGreaterThanOrEqual(t *testing.T) {
	matcher := NewMatcher(nil, nil)

	node := &graph.Node{
		ID:     "1",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Alice"),
		},
	}

	path := map[string]interface{}{
		"n": node,
	}

	comparison := &ast.ComparisonOp{
		Left: &ast.PropertyLookup{
			Node:     "n",
			Property: "name",
		},
		Operator: ">=",
		Right: &ast.Literal{
			Value: "Alice",
		},
	}

	result := matcher.EvaluateExpression(path, comparison, make(map[string]interface{}))
	if !result {
		t.Error("expected 'Alice' >= 'Alice' to be true")
	}
}

func TestEvaluateExpressionStringLessThan(t *testing.T) {
	matcher := NewMatcher(nil, nil)

	node := &graph.Node{
		ID:     "1",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Bob"),
		},
	}

	path := map[string]interface{}{
		"n": node,
	}

	comparison := &ast.ComparisonOp{
		Left: &ast.PropertyLookup{
			Node:     "n",
			Property: "name",
		},
		Operator: "<",
		Right: &ast.Literal{
			Value: "Zoe",
		},
	}

	result := matcher.EvaluateExpression(path, comparison, make(map[string]interface{}))
	if !result {
		t.Error("expected 'Bob' < 'Zoe' to be true")
	}
}

func TestEvaluateExpressionStringLessThanOrEqual(t *testing.T) {
	matcher := NewMatcher(nil, nil)

	node := &graph.Node{
		ID:     "1",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Bob"),
		},
	}

	path := map[string]interface{}{
		"n": node,
	}

	comparison := &ast.ComparisonOp{
		Left: &ast.PropertyLookup{
			Node:     "n",
			Property: "name",
		},
		Operator: "<=",
		Right: &ast.Literal{
			Value: "Bob",
		},
	}

	result := matcher.EvaluateExpression(path, comparison, make(map[string]interface{}))
	if !result {
		t.Error("expected 'Bob' <= 'Bob' to be true")
	}
}

func TestEvaluateExpressionStringGreaterThanFalse(t *testing.T) {
	matcher := NewMatcher(nil, nil)

	node := &graph.Node{
		ID:     "1",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Alice"),
		},
	}

	path := map[string]interface{}{
		"n": node,
	}

	comparison := &ast.ComparisonOp{
		Left: &ast.PropertyLookup{
			Node:     "n",
			Property: "name",
		},
		Operator: ">",
		Right: &ast.Literal{
			Value: "Zoe",
		},
	}

	result := matcher.EvaluateExpression(path, comparison, make(map[string]interface{}))
	if result {
		t.Error("expected 'Alice' > 'Zoe' to be false")
	}
}

func TestEvaluateExpressionStringLessThanFalse(t *testing.T) {
	matcher := NewMatcher(nil, nil)

	node := &graph.Node{
		ID:     "1",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Zoe"),
		},
	}

	path := map[string]interface{}{
		"n": node,
	}

	comparison := &ast.ComparisonOp{
		Left: &ast.PropertyLookup{
			Node:     "n",
			Property: "name",
		},
		Operator: "<",
		Right: &ast.Literal{
			Value: "Alice",
		},
	}

	result := matcher.EvaluateExpression(path, comparison, make(map[string]interface{}))
	if result {
		t.Error("expected 'Zoe' < 'Alice' to be false")
	}
}

func TestEvaluateExpressionStringWithParameterGreaterThan(t *testing.T) {
	matcher := NewMatcher(nil, nil)

	node := &graph.Node{
		ID:     "1",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Zoe"),
		},
	}

	path := map[string]interface{}{
		"n": node,
	}

	params := map[string]interface{}{
		"threshold": "Alice",
	}

	comparison := &ast.ComparisonOp{
		Left: &ast.PropertyLookup{
			Node:     "n",
			Property: "name",
		},
		Operator: ">",
		Right: &ast.Identifier{
			Name: "$threshold",
		},
	}

	result := matcher.EvaluateExpression(path, comparison, params)
	if !result {
		t.Error("expected 'Zoe' > $threshold (Alice) to be true with parameter")
	}
}

func TestEvaluateExpressionStringWithParameterLessThan(t *testing.T) {
	matcher := NewMatcher(nil, nil)

	node := &graph.Node{
		ID:     "1",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Bob"),
		},
	}

	path := map[string]interface{}{
		"n": node,
	}

	params := map[string]interface{}{
		"threshold": "Zoe",
	}

	comparison := &ast.ComparisonOp{
		Left: &ast.PropertyLookup{
			Node:     "n",
			Property: "name",
		},
		Operator: "<",
		Right: &ast.Identifier{
			Name: "$threshold",
		},
	}

	result := matcher.EvaluateExpression(path, comparison, params)
	if !result {
		t.Error("expected 'Bob' < $threshold (Zoe) to be true with parameter")
	}
}

func TestEvaluateExpressionStringEmpty(t *testing.T) {
	matcher := NewMatcher(nil, nil)

	node := &graph.Node{
		ID:     "1",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty(""),
		},
	}

	path := map[string]interface{}{
		"n": node,
	}

	comparison := &ast.ComparisonOp{
		Left: &ast.PropertyLookup{
			Node:     "n",
			Property: "name",
		},
		Operator: ">",
		Right: &ast.Literal{
			Value: "",
		},
	}

	result := matcher.EvaluateExpression(path, comparison, make(map[string]interface{}))
	if result {
		t.Error("expected '' > '' to be false (empty strings are equal)")
	}
}

func TestEvaluateExpressionStringUnicode(t *testing.T) {
	matcher := NewMatcher(nil, nil)

	node := &graph.Node{
		ID:     "1",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("张三"),
		},
	}

	path := map[string]interface{}{
		"n": node,
	}

	comparison := &ast.ComparisonOp{
		Left: &ast.PropertyLookup{
			Node:     "n",
			Property: "name",
		},
		Operator: ">",
		Right: &ast.Literal{
			Value: "李四",
		},
	}

	result := matcher.EvaluateExpression(path, comparison, make(map[string]interface{}))
	if result {
		t.Error("expected Unicode comparison to follow lexical ordering")
	}
}

func TestEvaluateExpressionStringCombinedConditions(t *testing.T) {
	matcher := NewMatcher(nil, nil)

	node := &graph.Node{
		ID:     "1",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Charlie"),
		},
	}

	path := map[string]interface{}{
		"n": node,
	}

	comparison := &ast.ComparisonOp{
		Left: &ast.PropertyLookup{
			Node:     "n",
			Property: "name",
		},
		Operator: ">",
		Right: &ast.Literal{
			Value: "Alice",
		},
	}

	result := matcher.EvaluateExpression(path, comparison, make(map[string]interface{}))
	if !result {
		t.Error("expected 'Charlie' > 'Alice' to be true")
	}

	comparison.Right = &ast.Literal{
		Value: "Zoe",
	}

	result = matcher.EvaluateExpression(path, comparison, make(map[string]interface{}))
	if result {
		t.Error("expected 'Charlie' > 'Zoe' to be false")
	}
}

func TestMatcherExecuteWithStringComparisonInWhere(t *testing.T) {
	path := "/tmp/gograph_string_where_test"
	defer os.RemoveAll(path)

	store, err := storage.Open(path)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer store.Close()

	index := graph.NewIndex(store)

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
			"name": graph.NewStringProperty("Zoe"),
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

	matcher := NewMatcher(store, index)

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

	where := &ast.WhereClause{
		Expression: &ast.ComparisonOp{
			Left: &ast.PropertyLookup{
				Node:     "n",
				Property: "name",
			},
			Operator: ">",
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
				{
					Expression: &ast.Identifier{Name: "n"},
				},
			},
		},
	}

	rows, columns, err := matcher.Execute(matchClause, make(map[string]interface{}), make(map[string]interface{}))
	if err != nil {
		t.Fatalf("failed to execute match: %v", err)
	}

	if len(rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(rows))
	}

	if len(columns) != 1 || columns[0] != "n" {
		t.Errorf("expected column 'n', got %v", columns)
	}
}

func TestMatcherExecuteWithStringComparisonLessThan(t *testing.T) {
	path := "/tmp/gograph_string_where_lt_test"
	defer os.RemoveAll(path)

	store, err := storage.Open(path)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer store.Close()

	index := graph.NewIndex(store)

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

	matcher := NewMatcher(store, index)

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

	where := &ast.WhereClause{
		Expression: &ast.ComparisonOp{
			Left: &ast.PropertyLookup{
				Node:     "n",
				Property: "name",
			},
			Operator: "<",
			Right: &ast.Literal{
				Value: "Zoe",
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

	rows, columns, err := matcher.Execute(matchClause, make(map[string]interface{}), make(map[string]interface{}))
	if err != nil {
		t.Fatalf("failed to execute match: %v", err)
	}

	if len(rows) != 2 {
		t.Errorf("expected 2 rows (Alice and Bob < Zoe), got %d", len(rows))
	}

	if len(columns) != 1 || columns[0] != "n" {
		t.Errorf("expected column 'n', got %v", columns)
	}
}