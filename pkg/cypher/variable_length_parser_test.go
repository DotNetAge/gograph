package cypher

import (
	"testing"

	"github.com/DotNetAge/gograph/pkg/cypher/ast"
)

func TestParserVariableLengthPathMinMax(t *testing.T) {
	parser := NewParser("MATCH (a)-[r:FRIEND*1..3]->(b) RETURN a, b")
	result, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if len(result.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(result.Statements))
	}

	match, ok := result.Statements[0].Clause.(*ast.MatchClause)
	if !ok {
		t.Fatal("expected MatchClause")
	}

	if len(match.Pattern.Elements) != 1 {
		t.Fatalf("expected 1 pattern element, got %d", len(match.Pattern.Elements))
	}

	elem := match.Pattern.Elements[0]
	if elem.Relation == nil {
		t.Fatal("expected relation pattern")
	}

	if elem.Relation.MinHops != 1 {
		t.Errorf("expected MinHops=1, got %d", elem.Relation.MinHops)
	}

	if elem.Relation.MaxHops != 3 {
		t.Errorf("expected MaxHops=3, got %d", elem.Relation.MaxHops)
	}

	if elem.Relation.RelType != "FRIEND" {
		t.Errorf("expected RelType=FRIEND, got %s", elem.Relation.RelType)
	}
}

func TestParserVariableLengthPathExactHop(t *testing.T) {
	parser := NewParser("MATCH (a)-[r:KNOWS*3]->(b) RETURN a, b")
	result, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	match := result.Statements[0].Clause.(*ast.MatchClause)
	elem := match.Pattern.Elements[0]

	if elem.Relation.MinHops != 3 {
		t.Errorf("expected MinHops=3, got %d", elem.Relation.MinHops)
	}

	if elem.Relation.MaxHops != 3 {
		t.Errorf("expected MaxHops=3, got %d", elem.Relation.MaxHops)
	}
}

func TestParserVariableLengthPathMinOnly(t *testing.T) {
	parser := NewParser("MATCH (a)-[r:KNOWS*2..]->(b) RETURN a, b")
	result, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	match := result.Statements[0].Clause.(*ast.MatchClause)
	elem := match.Pattern.Elements[0]

	if elem.Relation.MinHops != 2 {
		t.Errorf("expected MinHops=2, got %d", elem.Relation.MinHops)
	}

	if elem.Relation.MaxHops != -1 {
		t.Errorf("expected MaxHops=-1 (unlimited), got %d", elem.Relation.MaxHops)
	}
}

func TestParserVariableLengthPathMaxOnly(t *testing.T) {
	parser := NewParser("MATCH (a)-[r:KNOWS*..5]->(b) RETURN a, b")
	result, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	match := result.Statements[0].Clause.(*ast.MatchClause)
	elem := match.Pattern.Elements[0]

	if elem.Relation.MinHops != 1 {
		t.Errorf("expected MinHops=1 (default), got %d", elem.Relation.MinHops)
	}

	if elem.Relation.MaxHops != 5 {
		t.Errorf("expected MaxHops=5, got %d", elem.Relation.MaxHops)
	}
}

func TestParserVariableLengthPathWithoutVariable(t *testing.T) {
	parser := NewParser("MATCH (a)-[:FRIEND*1..3]->(b) RETURN a, b")
	result, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	match := result.Statements[0].Clause.(*ast.MatchClause)
	elem := match.Pattern.Elements[0]

	if elem.Relation.Variable != "" {
		t.Errorf("expected empty Variable, got %s", elem.Relation.Variable)
	}

	if elem.Relation.MinHops != 1 {
		t.Errorf("expected MinHops=1, got %d", elem.Relation.MinHops)
	}

	if elem.Relation.MaxHops != 3 {
		t.Errorf("expected MaxHops=3, got %d", elem.Relation.MaxHops)
	}
}

func TestParserVariableLengthPathWithoutType(t *testing.T) {
	parser := NewParser("MATCH (a)-[r*1..3]->(b) RETURN a, b")
	result, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	match := result.Statements[0].Clause.(*ast.MatchClause)
	elem := match.Pattern.Elements[0]

	if elem.Relation.RelType != "" {
		t.Errorf("expected empty RelType, got %s", elem.Relation.RelType)
	}

	if elem.Relation.MinHops != 1 {
		t.Errorf("expected MinHops=1, got %d", elem.Relation.MinHops)
	}

	if elem.Relation.MaxHops != 3 {
		t.Errorf("expected MaxHops=3, got %d", elem.Relation.MaxHops)
	}
}

func TestParserVariableLengthPathWithProperties(t *testing.T) {
	parser := NewParser("MATCH (a)-[r:FRIEND*1..3 {since: 2020}]->(b) RETURN a, b")
	result, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	match := result.Statements[0].Clause.(*ast.MatchClause)
	elem := match.Pattern.Elements[0]

	if elem.Relation.MinHops != 1 {
		t.Errorf("expected MinHops=1, got %d", elem.Relation.MinHops)
	}

	if elem.Relation.MaxHops != 3 {
		t.Errorf("expected MaxHops=3, got %d", elem.Relation.MaxHops)
	}

	if len(elem.Relation.Properties) == 0 {
		t.Error("expected properties to be parsed")
	}
}

func TestParserVariableLengthPathWithOutgoingDirection(t *testing.T) {
	parser := NewParser("MATCH (a)-[:FRIEND*1..3]->(b) RETURN a, b")
	result, err := parser.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	match := result.Statements[0].Clause.(*ast.MatchClause)
	elem := match.Pattern.Elements[0]

	if elem.Relation.MinHops != 1 {
		t.Errorf("expected MinHops=1, got %d", elem.Relation.MinHops)
	}

	if elem.Relation.MaxHops != 3 {
		t.Errorf("expected MaxHops=3, got %d", elem.Relation.MaxHops)
	}

	if elem.Relation.RelType != "FRIEND" {
		t.Errorf("expected RelType=FRIEND, got %s", elem.Relation.RelType)
	}
}

func TestParserHopRangeValidation(t *testing.T) {
	parser := NewParser("MATCH (a)-[r*0..3]->(b) RETURN a, b")
	_, err := parser.Parse()
	if err == nil {
		t.Error("expected error for max hops = 0")
	}
}

func TestParserHopRangeMinGreaterThanMax(t *testing.T) {
	parser := NewParser("MATCH (a)-[r*5..3]->(b) RETURN a, b")
	_, err := parser.Parse()
	if err == nil {
		t.Error("expected error for min > max")
	}
}