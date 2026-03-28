package cypher

import (
	"testing"

	"github.com/DotNetAge/gograph/pkg/cypher/ast"
)

func TestParserCreate(t *testing.T) {
	parser := NewParser("CREATE (n:User {name: 'Alice'})")
	result, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if len(result.Statements) != 1 {
		t.Errorf("expected 1 statement, got %d", len(result.Statements))
	}

	create, ok := result.Statements[0].Clause.(*ast.CreateClause)
	if !ok {
		t.Error("expected CreateClause")
	}

	if len(create.Pattern.Elements) != 1 {
		t.Errorf("expected 1 element, got %d", len(create.Pattern.Elements))
	}

	node := create.Pattern.Elements[0].Node
	if node.Variable != "n" {
		t.Errorf("expected variable n, got %s", node.Variable)
	}

	if len(node.Labels) != 1 || node.Labels[0] != "User" {
		t.Errorf("expected label User, got %v", node.Labels)
	}
}

func TestParserMatch(t *testing.T) {
	parser := NewParser("MATCH (n:User) RETURN n")
	result, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	match, ok := result.Statements[0].Clause.(*ast.MatchClause)
	if !ok {
		t.Error("expected MatchClause")
	}

	if match.Return == nil {
		t.Error("expected Return clause")
	}

	if len(match.Return.Items) != 1 {
		t.Errorf("expected 1 return item, got %d", len(match.Return.Items))
	}
}

func TestParserSet(t *testing.T) {
	parser := NewParser("SET n.name = 'Bob'")
	result, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	set, ok := result.Statements[0].Clause.(*ast.SetClause)
	if !ok {
		t.Error("expected SetClause")
		return
	}

	if len(set.Assignments) != 1 {
		t.Errorf("expected 1 assignment, got %d", len(set.Assignments))
	}

	if set.Assignments[0].Property.Property != "name" {
		t.Errorf("expected property name, got %s", set.Assignments[0].Property.Property)
	}
}

func TestParserSetMultiple(t *testing.T) {
	parser := NewParser("SET n.name = 'Bob', n.age = 30, n.active = true")
	result, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	set, ok := result.Statements[0].Clause.(*ast.SetClause)
	if !ok {
		t.Error("expected SetClause")
		return
	}

	if len(set.Assignments) != 3 {
		t.Errorf("expected 3 assignments, got %d", len(set.Assignments))
	}

	expectedProperties := []string{"name", "age", "active"}
	for i, expected := range expectedProperties {
		if set.Assignments[i].Property.Property != expected {
			t.Errorf("expected property %s, got %s", expected, set.Assignments[i].Property.Property)
		}
	}
}

func TestParserDelete(t *testing.T) {
	parser := NewParser("MATCH (n:User) DELETE n")
	result, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	match, ok := result.Statements[0].Clause.(*ast.MatchClause)
	if !ok {
		t.Error("expected MatchClause")
	}

	if match.Delete == nil {
		t.Error("expected Delete clause")
	}

	if len(match.Delete.Expressions) != 1 {
		t.Errorf("expected 1 expression, got %d", len(match.Delete.Expressions))
	}
}

func TestParserDetachDelete(t *testing.T) {
	parser := NewParser("MATCH (n:User) DETACH DELETE n")
	result, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	match, ok := result.Statements[0].Clause.(*ast.MatchClause)
	if !ok {
		t.Error("expected MatchClause")
	}

	if match.Delete == nil {
		t.Error("expected Delete clause")
	}

	if !match.Delete.Detach {
		t.Error("expected Detach to be true")
	}
}

func TestParserRemove(t *testing.T) {
	parser := NewParser("REMOVE n:VIP")
	result, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	remove, ok := result.Statements[0].Clause.(*ast.RemoveClause)
	if !ok {
		t.Error("expected RemoveClause")
		return
	}

	if len(remove.Removals) != 1 {
		t.Errorf("expected 1 removal, got %d", len(remove.Removals))
	}

	if remove.Removals[0].Type != ast.RemoveItemTypeLabel {
		t.Error("expected label removal")
	}
}

func TestParserWhere(t *testing.T) {
	parser := NewParser("MATCH (n:User) WHERE n.age > 18 RETURN n")
	result, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	match, ok := result.Statements[0].Clause.(*ast.MatchClause)
	if !ok {
		t.Error("expected MatchClause")
	}

	if match.Where == nil {
		t.Error("expected Where clause")
	}

	if match.Where.Expression == nil {
		t.Error("expected Where expression")
	}
}

func TestParserRelationship(t *testing.T) {
	parser := NewParser("CREATE (a:User)-[:KNOWS]->(b:User)")
	result, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	create, ok := result.Statements[0].Clause.(*ast.CreateClause)
	if !ok {
		t.Error("expected CreateClause")
	}

	elem := create.Pattern.Elements[0]
	if elem.Relation == nil {
		t.Error("expected relationship")
	}

	if elem.Relation.RelType != "KNOWS" {
		t.Errorf("expected relationship type KNOWS, got %s", elem.Relation.RelType)
	}
}

func TestParserMultipleStatements(t *testing.T) {
	parser := NewParser("CREATE (n:User); CREATE (m:Product)")
	result, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if len(result.Statements) != 2 {
		t.Errorf("expected 2 statements, got %d", len(result.Statements))
	}
}

func TestParserEmpty(t *testing.T) {
	parser := NewParser("")
	result, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse empty string: %v", err)
	}

	if len(result.Statements) != 0 {
		t.Errorf("expected 0 statements, got %d", len(result.Statements))
	}
}

func TestParserWhitespace(t *testing.T) {
	parser := NewParser("  CREATE   (n:User)  ")
	result, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse with whitespace: %v", err)
	}

	if len(result.Statements) != 1 {
		t.Errorf("expected 1 statement, got %d", len(result.Statements))
	}
}

func TestParserLogicalOperators(t *testing.T) {
	// Test AND operator
	parser := NewParser("MATCH (n:User) WHERE n.age > 18 AND n.name = 'Alice' RETURN n")
	result, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse AND operator: %v", err)
	}

	match, ok := result.Statements[0].Clause.(*ast.MatchClause)
	if !ok {
		t.Error("expected MatchClause")
	}

	if match.Where == nil {
		t.Error("expected Where clause")
	}

	// Test OR operator
	parser = NewParser("MATCH (n:User) WHERE n.age > 18 OR n.name = 'Alice' RETURN n")
	result, err = parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse OR operator: %v", err)
	}

	match, ok = result.Statements[0].Clause.(*ast.MatchClause)
	if !ok {
		t.Error("expected MatchClause")
	}

	if match.Where == nil {
		t.Error("expected Where clause")
	}

	// Test combined AND/OR
	parser = NewParser("MATCH (n:User) WHERE (n.age > 18 AND n.name = 'Alice') OR (n.age < 10 AND n.name = 'Bob') RETURN n")
	result, err = parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse combined AND/OR: %v", err)
	}

	match, ok = result.Statements[0].Clause.(*ast.MatchClause)
	if !ok {
		t.Error("expected MatchClause")
	}

	if match.Where == nil {
		t.Error("expected Where clause")
	}
}

func TestParserDeleteExpression(t *testing.T) {
	// Test deleting node
	parser := NewParser("DELETE n")
	result, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse DELETE n: %v", err)
	}

	deleteClause, ok := result.Statements[0].Clause.(*ast.DeleteClause)
	if !ok {
		t.Error("expected DeleteClause")
	}

	if len(deleteClause.Expressions) != 1 {
		t.Errorf("expected 1 expression, got %d", len(deleteClause.Expressions))
	}

	// Test deleting relationship
	parser = NewParser("DELETE r")
	result, err = parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse DELETE r: %v", err)
	}

	deleteClause, ok = result.Statements[0].Clause.(*ast.DeleteClause)
	if !ok {
		t.Error("expected DeleteClause")
	}

	if len(deleteClause.Expressions) != 1 {
		t.Errorf("expected 1 expression, got %d", len(deleteClause.Expressions))
	}

	// Test deleting multiple items
	parser = NewParser("DELETE n, r")
	result, err = parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse DELETE n, r: %v", err)
	}

	deleteClause, ok = result.Statements[0].Clause.(*ast.DeleteClause)
	if !ok {
		t.Error("expected DeleteClause")
	}

	if len(deleteClause.Expressions) != 2 {
		t.Errorf("expected 2 expressions, got %d", len(deleteClause.Expressions))
	}
}

func TestParserRemoveProperty(t *testing.T) {
	// Test removing property
	parser := NewParser("REMOVE n.age")
	result, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse REMOVE n.age: %v", err)
	}

	removeClause, ok := result.Statements[0].Clause.(*ast.RemoveClause)
	if !ok {
		t.Error("expected RemoveClause")
	}

	if len(removeClause.Removals) != 1 {
		t.Errorf("expected 1 removal, got %d", len(removeClause.Removals))
	}

	if removeClause.Removals[0].Type != ast.RemoveItemTypeProperty {
		t.Error("expected property removal")
	}

	// Test removing multiple items
	parser = NewParser("REMOVE n.age, n:VIP")
	result, err = parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse REMOVE n.age, n:VIP: %v", err)
	}

	removeClause, ok = result.Statements[0].Clause.(*ast.RemoveClause)
	if !ok {
		t.Error("expected RemoveClause")
	}

	if len(removeClause.Removals) != 2 {
		t.Errorf("expected 2 removals, got %d", len(removeClause.Removals))
	}
}

func TestParserValue(t *testing.T) {
	// Test parsing string value
	parser := NewParser("'Alice'")
	val, err := parser.parseValue()
	if err != nil {
		t.Fatalf("failed to parse string value: %v", err)
	}
	if val != "Alice" {
		t.Errorf("expected 'Alice', got %v", val)
	}

	// Test parsing numeric value (int)
	parser = NewParser("42")
	val, err = parser.parseValue()
	if err != nil {
		t.Fatalf("failed to parse int value: %v", err)
	}
	if val != int64(42) {
		t.Errorf("expected 42, got %v", val)
	}

	// Test parsing numeric value (float)
	parser = NewParser("3.14")
	val, err = parser.parseValue()
	if err != nil {
		t.Fatalf("failed to parse float value: %v", err)
	}
	if val != 3.14 {
		t.Errorf("expected 3.14, got %v", val)
	}

	// Test parsing boolean value (true)
	parser = NewParser("true")
	val, err = parser.parseValue()
	if err != nil {
		t.Fatalf("failed to parse true value: %v", err)
	}
	if val != true {
		t.Errorf("expected true, got %v", val)
	}

	// Test parsing boolean value (false)
	parser = NewParser("false")
	val, err = parser.parseValue()
	if err != nil {
		t.Fatalf("failed to parse false value: %v", err)
	}
	if val != false {
		t.Errorf("expected false, got %v", val)
	}

	// Test parsing parameter
	parser = NewParser("$param")
	val, err = parser.parseValue()
	if err != nil {
		t.Fatalf("failed to parse parameter: %v", err)
	}
	if val != "$param" {
		t.Errorf("expected $param, got %v", val)
	}

	// Test parsing identifier
	parser = NewParser("variable")
	val, err = parser.parseValue()
	if err != nil {
		t.Fatalf("failed to parse identifier: %v", err)
	}
	if val != "variable" {
		t.Errorf("expected variable, got %v", val)
	}
}

func TestParserPropertyAccess(t *testing.T) {
	// Test parsing simple property access
	parser := NewParser("n")
	prop, err := parser.parsePropertyAccess()
	if err != nil {
		t.Fatalf("failed to parse property access: %v", err)
	}
	if prop.Node != "n" || prop.Property != "" {
		t.Errorf("expected Node: n, Property: '', got Node: %s, Property: %s", prop.Node, prop.Property)
	}

	// Test parsing property access with property
	parser = NewParser("n.name")
	prop, err = parser.parsePropertyAccess()
	if err != nil {
		t.Fatalf("failed to parse property access with property: %v", err)
	}
	if prop.Node != "n" || prop.Property != "name" {
		t.Errorf("expected Node: n, Property: name, got Node: %s, Property: %s", prop.Node, prop.Property)
	}
}

func TestParserPrimary(t *testing.T) {
	// Test parsing string literal
	parser := NewParser("'Alice'")
	expr, err := parser.parsePrimary()
	if err != nil {
		t.Fatalf("failed to parse string literal: %v", err)
	}
	lit, ok := expr.(*ast.Literal)
	if !ok {
		t.Error("expected Literal")
	}
	if lit.Value != "Alice" {
		t.Errorf("expected 'Alice', got %v", lit.Value)
	}

	// Test parsing number literal
	parser = NewParser("42")
	expr, err = parser.parsePrimary()
	if err != nil {
		t.Fatalf("failed to parse number literal: %v", err)
	}
	lit, ok = expr.(*ast.Literal)
	if !ok {
		t.Error("expected Literal")
	}
	if lit.Value != int64(42) {
		t.Errorf("expected 42, got %v", lit.Value)
	}

	// Test parsing boolean literal
	parser = NewParser("true")
	expr, err = parser.parsePrimary()
	if err != nil {
		t.Fatalf("failed to parse boolean literal: %v", err)
	}
	lit, ok = expr.(*ast.Literal)
	if !ok {
		t.Error("expected Literal")
	}
	if lit.Value != true {
		t.Errorf("expected true, got %v", lit.Value)
	}

	// Test parsing parameter
	parser = NewParser("$param")
	expr, err = parser.parsePrimary()
	if err != nil {
		t.Fatalf("failed to parse parameter: %v", err)
	}
	ident, ok := expr.(*ast.Identifier)
	if !ok {
		t.Error("expected Identifier")
	}
	if ident.Name != "$param" {
		t.Errorf("expected $param, got %s", ident.Name)
	}

	// Test parsing parenthesized expression
	parser = NewParser("(n.age > 18)")
	expr, err = parser.parsePrimary()
	if err != nil {
		t.Fatalf("failed to parse parenthesized expression: %v", err)
	}
	_, ok = expr.(*ast.ComparisonOp)
	if !ok {
		t.Error("expected ComparisonOp")
	}
}

func TestParserComparison(t *testing.T) {
	// Test parsing equality comparison
	parser := NewParser("n.name = 'Alice'")
	expr, err := parser.parseComparison()
	if err != nil {
		t.Fatalf("failed to parse equality comparison: %v", err)
	}
	comp, ok := expr.(*ast.ComparisonOp)
	if !ok {
		t.Error("expected ComparisonOp")
	}
	if comp.Operator != "=" {
		t.Errorf("expected =, got %s", comp.Operator)
	}

	// Test parsing inequality comparison
	parser = NewParser("n.age != 18")
	expr, err = parser.parseComparison()
	if err != nil {
		t.Fatalf("failed to parse inequality comparison: %v", err)
	}
	comp, ok = expr.(*ast.ComparisonOp)
	if !ok {
		t.Error("expected ComparisonOp")
	}
	if comp.Operator != "!=" {
		t.Errorf("expected !=, got %s", comp.Operator)
	}

	// Test parsing greater than comparison
	parser = NewParser("n.age > 18")
	expr, err = parser.parseComparison()
	if err != nil {
		t.Fatalf("failed to parse greater than comparison: %v", err)
	}
	comp, ok = expr.(*ast.ComparisonOp)
	if !ok {
		t.Error("expected ComparisonOp")
	}
	if comp.Operator != ">" {
		t.Errorf("expected >, got %s", comp.Operator)
	}

	// Test parsing less than comparison
	parser = NewParser("n.age < 18")
	expr, err = parser.parseComparison()
	if err != nil {
		t.Fatalf("failed to parse less than comparison: %v", err)
	}
	comp, ok = expr.(*ast.ComparisonOp)
	if !ok {
		t.Error("expected ComparisonOp")
	}
	if comp.Operator != "<" {
		t.Errorf("expected <, got %s", comp.Operator)
	}

	// Test parsing greater than or equal comparison
	parser = NewParser("n.age >= 18")
	expr, err = parser.parseComparison()
	if err != nil {
		t.Fatalf("failed to parse greater than or equal comparison: %v", err)
	}
	comp, ok = expr.(*ast.ComparisonOp)
	if !ok {
		t.Error("expected ComparisonOp")
	}
	if comp.Operator != ">=" {
		t.Errorf("expected >=, got %s", comp.Operator)
	}

	// Test parsing less than or equal comparison
	parser = NewParser("n.age <= 18")
	expr, err = parser.parseComparison()
	if err != nil {
		t.Fatalf("failed to parse less than or equal comparison: %v", err)
	}
	comp, ok = expr.(*ast.ComparisonOp)
	if !ok {
		t.Error("expected ComparisonOp")
	}
	if comp.Operator != "<=" {
		t.Errorf("expected <=, got %s", comp.Operator)
	}
}
