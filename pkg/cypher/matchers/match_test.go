package matchers

import (
	"os"
	"testing"

	"github.com/DotNetAge/gograph/pkg/cypher/ast"
	"github.com/DotNetAge/gograph/pkg/graph"
	"github.com/DotNetAge/gograph/pkg/storage"
)

func TestMatcher(t *testing.T) {
	path := "/tmp/gograph_matcher_test"
	defer os.RemoveAll(path)

	store, err := storage.Open(path)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer store.Close()

	matcher := NewMatcher(store, graph.NewIndex(store))

	clause := &ast.MatchClause{
		Pattern: ast.Pattern{
			Elements: []ast.PatternElement{
				{
					Node: &ast.NodePattern{
						Variable: "n",
						Labels:   []string{"User"},
					},
				},
			},
		},
	}

	rows, columns, err := matcher.Execute(clause, make(map[string]interface{}), nil)
	if err != nil {
		t.Fatalf("failed to execute match: %v", err)
	}

	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}

	if len(columns) != 0 {
		t.Errorf("expected 0 columns, got %d", len(columns))
	}
}

func TestMatcherWithReturn(t *testing.T) {
	path := "/tmp/gograph_matcher_return_test"
	defer os.RemoveAll(path)

	store, err := storage.Open(path)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer store.Close()

	matcher := NewMatcher(store, graph.NewIndex(store))

	clause := &ast.MatchClause{
		Pattern: ast.Pattern{
			Elements: []ast.PatternElement{
				{
					Node: &ast.NodePattern{
						Variable: "n",
						Labels:   []string{"User"},
					},
				},
			},
		},
		Return: &ast.ReturnClause{
			Items: []ast.ReturnItem{
				{
					Expression: &ast.PropertyLookup{
						Node: "n",
					},
				},
			},
		},
	}

	rows, columns, err := matcher.Execute(clause, make(map[string]interface{}), nil)
	if err != nil {
		t.Fatalf("failed to execute match: %v", err)
	}

	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}

	if len(columns) != 1 || columns[0] != "n" {
		t.Errorf("expected column 'n', got %v", columns)
	}
}

func TestToInt64(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected int64
		ok       bool
	}{
		{int64(42), 42, true},
		{int(42), 42, true},
		{int32(42), 42, true},
		{float64(42.0), 42, true},
		{float64(42.5), 0, false},
		{"42", 0, false},
	}

	for _, tc := range tests {
		result, ok := ToInt64(tc.input)
		if ok != tc.ok {
			t.Errorf("expected ok=%v for %v, got %v", tc.ok, tc.input, ok)
		}
		if ok && result != tc.expected {
			t.Errorf("expected %d, got %d", tc.expected, result)
		}
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected float64
		ok       bool
	}{
		{float64(3.14), 3.14, true},
		{int64(42), 42.0, true},
		{int(42), 42.0, true},
		{"3.14", 0.0, false},
	}

	for _, tc := range tests {
		result, ok := ToFloat64(tc.input)
		if ok != tc.ok {
			t.Errorf("expected ok=%v for %v, got %v", tc.ok, tc.input, ok)
		}
		if ok && result != tc.expected {
			t.Errorf("expected %f, got %f", tc.expected, result)
		}
	}

	// Test float32 conversion separately
	result, ok := ToFloat64(float32(3.14))
	if !ok {
		t.Error("expected ok=true for float32")
	}
	// Check if the result is approximately 3.14
	if result < 3.139999 || result > 3.140001 {
		t.Errorf("expected ~3.14, got %f", result)
	}
}

func TestPropertyToInterface(t *testing.T) {
	matcher := NewMatcher(nil, nil)

	tests := []struct {
		name     string
		prop     graph.PropertyValue
		expected interface{}
	}{
		{"string", graph.NewStringProperty("test"), "test"},
		{"int", graph.NewIntProperty(42), int64(42)},
		{"float", graph.NewFloatProperty(3.14), 3.14},
		{"bool", graph.NewBoolProperty(true), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := matcher.PropertyToInterface(tc.prop)
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestNodeMatchesProperties(t *testing.T) {
	matcher := NewMatcher(nil, nil)

	// Create a node with properties
	node := &graph.Node{
		ID:     "1",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Alice"),
			"age":  graph.NewIntProperty(30),
		},
	}

	// Test with matching labels and properties
	labels := []string{"User"}
	props := map[string]interface{}{
		"name": "Alice",
	}

	if !matcher.NodeMatchesProperties(node, labels, props) {
		t.Error("expected node to match properties")
	}

	// Test with non-matching properties
	props = map[string]interface{}{
		"name": "Bob",
	}

	if matcher.NodeMatchesProperties(node, labels, props) {
		t.Error("expected node to not match properties")
	}

	// Test with non-matching labels
	labels = []string{"Admin"}
	props = map[string]interface{}{
		"name": "Alice",
	}

	if matcher.NodeMatchesProperties(node, labels, props) {
		t.Error("expected node to not match labels")
	}
}

func TestExecuteWithSimpleMatch(t *testing.T) {
	// Create a temporary directory for storage
	path := "/tmp/gograph_execute_test"
	defer os.RemoveAll(path)

	// Create storage and index
	store, err := storage.Open(path)
	if err != nil {
		t.Fatalf("failed to open storage: %v", err)
	}
	defer store.Close()

	index := graph.NewIndex(store)

	// Create a node and store it
	node := &graph.Node{
		ID:     "1",
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

	if err := store.Put(storage.NodeKey(node.ID), data); err != nil {
		t.Fatalf("failed to store node: %v", err)
	}

	// Build label index
	if err := index.BuildLabelIndex(store, node); err != nil {
		t.Fatalf("failed to build label index: %v", err)
	}

	// Create matcher
	matcher := NewMatcher(store, index)

	// Create a simple MATCH clause
	pattern := ast.Pattern{
		Elements: []ast.PatternElement{
			{
				Node: &ast.NodePattern{
					Variable: "n",
					Labels:   []string{"User"},
					Properties: map[string]interface{}{
						"name": "Alice",
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
}

func TestEvaluateExpression(t *testing.T) {
	matcher := NewMatcher(nil, nil)

	// Create a node with properties
	node := &graph.Node{
		ID:     "1",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Alice"),
			"age":  graph.NewIntProperty(30),
		},
	}

	// Create a path with the node
	path := map[string]interface{}{
		"n": node,
	}

	// Test string equality
	comparison := &ast.ComparisonOp{
		Left: &ast.PropertyLookup{
			Node:     "n",
			Property: "name",
		},
		Operator: "=",
		Right: &ast.Literal{
			Value: "Alice",
		},
	}

	result := matcher.EvaluateExpression(path, comparison, make(map[string]interface{}))
	if !result {
		t.Error("expected expression to evaluate to true")
	}

	// Test string inequality
	comparison.Operator = "!="
	comparison.Right = &ast.Literal{
		Value: "Bob",
	}

	result = matcher.EvaluateExpression(path, comparison, make(map[string]interface{}))
	if !result {
		t.Error("expected expression to evaluate to true")
	}

	// Test numeric comparison
	comparison.Left = &ast.PropertyLookup{
		Node:     "n",
		Property: "age",
	}
	comparison.Operator = ">"
	comparison.Right = &ast.Literal{
		Value: 25,
	}

	result = matcher.EvaluateExpression(path, comparison, make(map[string]interface{}))
	if !result {
		t.Error("expected expression to evaluate to true")
	}

	// Test numeric comparison with float
	comparison.Operator = "<"
	comparison.Right = &ast.Literal{
		Value: 35.5,
	}

	result = matcher.EvaluateExpression(path, comparison, make(map[string]interface{}))
	if !result {
		t.Error("expected expression to evaluate to true")
	}
}

func TestFillRow(t *testing.T) {
	matcher := NewMatcher(nil, nil)

	// Create a node with properties
	node := &graph.Node{
		ID:     "1",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Alice"),
			"age":  graph.NewIntProperty(30),
		},
	}

	// Create a path with the node
	path := map[string]interface{}{
		"n": node,
	}

	// Test filling row with node
	row := make(map[string]interface{})
	item := ast.ReturnItem{
		Expression: &ast.Identifier{Name: "n"},
	}

	matcher.fillRow(row, item, path)
	if row["n"] != node {
		t.Error("expected node in row")
	}

	// Test filling row with property
	row = make(map[string]interface{})
	item = ast.ReturnItem{
		Expression: &ast.PropertyLookup{
			Node:     "n",
			Property: "name",
		},
	}

	matcher.fillRow(row, item, path)
	if row["n.name"] != "Alice" {
		t.Errorf("expected 'Alice', got %v", row["n.name"])
	}

	// Test filling row with non-existent property
	row = make(map[string]interface{})
	item = ast.ReturnItem{
		Expression: &ast.PropertyLookup{
			Node:     "n",
			Property: "nonExistent",
		},
	}

	matcher.fillRow(row, item, path)
	if _, ok := row["n.nonExistent"]; ok {
		t.Error("expected non-existent property to not be in row")
	}

	// Test filling row with non-existent node
	row = make(map[string]interface{})
	item = ast.ReturnItem{
		Expression: &ast.PropertyLookup{
			Node:     "nonExistent",
			Property: "name",
		},
	}

	matcher.fillRow(row, item, path)
	if _, ok := row["nonExistent.name"]; ok {
		t.Error("expected non-existent node to not be in row")
	}
}

func TestFillRowWithRelationship(t *testing.T) {
	matcher := NewMatcher(nil, nil)

	// Create a relationship with properties
	rel := &graph.Relationship{
		ID:          "1",
		Type:        "KNOWS",
		StartNodeID: "1",
		EndNodeID:   "2",
		Properties: map[string]graph.PropertyValue{
			"since": graph.NewIntProperty(2020),
		},
	}

	// Create a path with the relationship
	path := map[string]interface{}{
		"r": rel,
	}

	// Test filling row with relationship
	row := make(map[string]interface{})
	item := ast.ReturnItem{
		Expression: &ast.Identifier{Name: "r"},
	}

	matcher.fillRow(row, item, path)
	if row["r"] != rel {
		t.Error("expected relationship in row")
	}

	// Test filling row with relationship property
	row = make(map[string]interface{})
	item = ast.ReturnItem{
		Expression: &ast.PropertyLookup{
			Node:     "r",
			Property: "since",
		},
	}

	matcher.fillRow(row, item, path)
	if row["r.since"] != int64(2020) {
		t.Errorf("expected 2020, got %v", row["r.since"])
	}
}

func TestFindNodesByVariableAndLabel(t *testing.T) {
	// Create a temporary directory for storage
	path := "/tmp/gograph_find_nodes_test"
	defer os.RemoveAll(path)

	// Create storage
	store, err := storage.Open(path)
	if err != nil {
		t.Fatalf("failed to open storage: %v", err)
	}
	defer store.Close()

	// Create a node and store it
	node := &graph.Node{
		ID:     "1",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Alice"),
		},
	}

	data, err := storage.Marshal(node)
	if err != nil {
		t.Fatalf("failed to marshal node: %v", err)
	}

	if err := store.Put(storage.NodeKey(node.ID), data); err != nil {
		t.Fatalf("failed to store node: %v", err)
	}

	// Create matcher
	matcher := NewMatcher(store, graph.NewIndex(store))

	// Test finding nodes
	nodes, err := matcher.FindNodesByVariableAndLabel("n", nil)
	if err != nil {
		t.Fatalf("failed to find nodes: %v", err)
	}

	if len(nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(nodes))
	}

	if nodes[0].ID != "1" {
		t.Errorf("expected node ID '1', got %s", nodes[0].ID)
	}
}

func TestExecuteWithRelationship(t *testing.T) {
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

	// Create nodes
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

	// Store nodes
	data1, err := storage.Marshal(node1)
	if err != nil {
		t.Fatalf("failed to marshal node1: %v", err)
	}

	data2, err := storage.Marshal(node2)
	if err != nil {
		t.Fatalf("failed to marshal node2: %v", err)
	}

	if err := store.Put(storage.NodeKey(node1.ID), data1); err != nil {
		t.Fatalf("failed to store node1: %v", err)
	}

	if err := store.Put(storage.NodeKey(node2.ID), data2); err != nil {
		t.Fatalf("failed to store node2: %v", err)
	}

	// Build label indexes
	if err := index.BuildLabelIndex(store, node1); err != nil {
		t.Fatalf("failed to build label index for node1: %v", err)
	}

	if err := index.BuildLabelIndex(store, node2); err != nil {
		t.Fatalf("failed to build label index for node2: %v", err)
	}

	// Create relationship
	rel := &graph.Relationship{
		ID:          "1",
		Type:        "KNOWS",
		StartNodeID: node1.ID,
		EndNodeID:   node2.ID,
		Properties: map[string]graph.PropertyValue{
			"since": graph.NewIntProperty(2020),
		},
	}

	// Store relationship
	dataRel, err := storage.Marshal(rel)
	if err != nil {
		t.Fatalf("failed to marshal relationship: %v", err)
	}

	if err := store.Put(storage.RelKey(rel.ID), dataRel); err != nil {
		t.Fatalf("failed to store relationship: %v", err)
	}

	// Build adjacency list
	adj := graph.NewAdjacencyList(store)
	if err := adj.AddRelationship(store, rel); err != nil {
		t.Fatalf("failed to add relationship to adjacency list: %v", err)
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

func TestEvaluateExpressionExtended(t *testing.T) {
	matcher := NewMatcher(nil, nil)

	// Test with int property
	node := &graph.Node{
		ID:     "1",
		Labels: []string{"User"},
		Properties: map[string]graph.PropertyValue{
			"age":  graph.NewIntProperty(30),
			"name": graph.NewStringProperty("Alice"),
		},
	}

	path := map[string]interface{}{
		"n": node,
	}

	// Test int comparisons
	tests := []struct {
		name     string
		operator string
		value    interface{}
		expected bool
	}{
		{"age > 25", ">", 25, true},
		{"age >= 30", ">=", 30, true},
		{"age < 35", "<", 35, true},
		{"age <= 30", "<=", 30, true},
		{"age = 30", "=", 30, true},
		{"age != 25", "!=", 25, true},
		{"age > 35", ">", 35, false},
		{"age < 25", "<", 25, false},
		{"age = 25", "=", 25, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			expr := &ast.ComparisonOp{
				Left: &ast.PropertyLookup{
					Node:     "n",
					Property: "age",
				},
				Operator: tc.operator,
				Right:    &ast.Literal{Value: tc.value},
			}
			result := matcher.EvaluateExpression(path, expr, nil)
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}

	// Test string comparisons
	stringTests := []struct {
		name     string
		operator string
		value    string
		expected bool
	}{
		{"name = Alice", "=", "Alice", true},
		{"name != Bob", "!=", "Bob", true},
		{"name = Bob", "=", "Bob", false},
	}

	for _, tc := range stringTests {
		t.Run(tc.name, func(t *testing.T) {
			expr := &ast.ComparisonOp{
				Left: &ast.PropertyLookup{
					Node:     "n",
					Property: "name",
				},
				Operator: tc.operator,
				Right:    &ast.Literal{Value: tc.value},
			}
			result := matcher.EvaluateExpression(path, expr, nil)
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}

	// Test with parameter
	paramTests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{"parameter age = 30", 30, true},
		{"parameter age = 25", 25, false},
	}

	for _, tc := range paramTests {
		t.Run(tc.name, func(t *testing.T) {
			expr := &ast.ComparisonOp{
				Left: &ast.PropertyLookup{
					Node:     "n",
					Property: "age",
				},
				Operator: "=",
				Right:    &ast.Identifier{Name: "$age"},
			}
			result := matcher.EvaluateExpression(path, expr, map[string]interface{}{"age": tc.value})
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}

	// Test with relationship
	rel := &graph.Relationship{
		ID:          "1",
		Type:        "KNOWS",
		StartNodeID: "1",
		EndNodeID:   "2",
		Properties: map[string]graph.PropertyValue{
			"since": graph.NewIntProperty(2020),
		},
	}

	relPath := map[string]interface{}{
		"r": rel,
	}

	t.Run("relationship property", func(t *testing.T) {
		expr := &ast.ComparisonOp{
			Left: &ast.PropertyLookup{
				Node:     "r",
				Property: "since",
			},
			Operator: "=",
			Right:    &ast.Literal{Value: 2020},
		}
		result := matcher.EvaluateExpression(relPath, expr, nil)
		if !result {
			t.Error("expected relationship property to match")
		}
	})
}

func TestNodeMatchesPropertiesExtended(t *testing.T) {
	matcher := NewMatcher(nil, nil)

	tests := []struct {
		name     string
		node     *graph.Node
		labels   []string
		props    map[string]interface{}
		expected bool
	}{
		{
			name: "node matches labels and properties",
			node: &graph.Node{
				ID:     "1",
				Labels: []string{"User", "Admin"},
				Properties: map[string]graph.PropertyValue{
					"name":   graph.NewStringProperty("Alice"),
					"age":    graph.NewIntProperty(30),
					"active": graph.NewBoolProperty(true),
				},
			},
			labels: []string{"User"},
			props: map[string]interface{}{
				"name": "Alice",
				"age":  30,
			},
			expected: true,
		},
		{
			name: "node doesn't match label",
			node: &graph.Node{
				ID:     "1",
				Labels: []string{"User"},
				Properties: map[string]graph.PropertyValue{
					"name": graph.NewStringProperty("Alice"),
				},
			},
			labels:   []string{"Admin"},
			props:    map[string]interface{}{"name": "Alice"},
			expected: false,
		},
		{
			name: "node doesn't match property",
			node: &graph.Node{
				ID:     "1",
				Labels: []string{"User"},
				Properties: map[string]graph.PropertyValue{
					"name": graph.NewStringProperty("Alice"),
				},
			},
			labels:   []string{"User"},
			props:    map[string]interface{}{"name": "Bob"},
			expected: false,
		},
		{
			name: "node with float property",
			node: &graph.Node{
				ID:     "1",
				Labels: []string{"Product"},
				Properties: map[string]graph.PropertyValue{
					"price": graph.NewFloatProperty(9.99),
				},
			},
			labels:   []string{"Product"},
			props:    map[string]interface{}{"price": 9.99},
			expected: true,
		},
		{
			name: "node with bool property",
			node: &graph.Node{
				ID:     "1",
				Labels: []string{"User"},
				Properties: map[string]graph.PropertyValue{
					"active": graph.NewBoolProperty(true),
				},
			},
			labels:   []string{"User"},
			props:    map[string]interface{}{"active": true},
			expected: true,
		},
		{
			name: "node missing property",
			node: &graph.Node{
				ID:     "1",
				Labels: []string{"User"},
				Properties: map[string]graph.PropertyValue{
					"name": graph.NewStringProperty("Alice"),
				},
			},
			labels:   []string{"User"},
			props:    map[string]interface{}{"name": "Alice", "age": 30},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := matcher.NodeMatchesProperties(tc.node, tc.labels, tc.props)
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}
