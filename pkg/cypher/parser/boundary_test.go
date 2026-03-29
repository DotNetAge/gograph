package parser_test

import (
	"testing"

	"github.com/DotNetAge/gograph/pkg/cypher/parser"
)

func TestParser_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldError bool
	}{
		{"empty input", "", false},
		{"whitespace only", "   \n\t   ", false},
		{"unmatched parenthesis", "MATCH (n:Person RETURN n", true},
		{"unmatched bracket", "MATCH (n)-[r:KNOWS->(m) RETURN n", true},
		{"unmatched brace", "MATCH (n {name: 'Alice') RETURN n", true},
		{"invalid keyword", "INVALID (n:Person) RETURN n", true},
		{"missing return", "MATCH (n:Person)", false},
		{"invalid expression", "MATCH (n) RETURN n.", true},
		{"invalid property access", "MATCH (n) RETURN n..name", true},
		{"invalid relationship", "MATCH (a)-[]->(b) RETURN a", false},
		{"empty label", "MATCH (n:) RETURN n", true},
		{"empty relationship type", "MATCH (a)-[:]->(b) RETURN a", true},
		{"invalid variable length", "MATCH (a)-[*abc]->(b) RETURN a", true},
		{"missing where condition", "MATCH (n) WHERE RETURN n", true},
		{"invalid set target", "MATCH (n) SET = 5", true},
		{"missing set value", "MATCH (n) SET n.age =", true},
		{"invalid delete target", "MATCH (n) DELETE", true},
		{"missing unwind variable", "UNWIND [1, 2, 3] RETURN 1", true},
		{"invalid case", "RETURN CASE WHEN THEN END", true},
		{"missing case end", "RETURN CASE WHEN 1 THEN 2", true},
		{"invalid function call", "RETURN COUNT()", false},
		{"missing closing quote", "MATCH (n) WHERE n.name = 'Alice RETURN n", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(tt.input)
			_, err := p.Parse()
			
			if tt.shouldError && err == nil {
				t.Errorf("expected error for input %q but got none", tt.input)
			}
			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error for input %q: %v", tt.input, err)
			}
		})
	}
}

func TestParser_BoundaryValues(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"max integer", "RETURN 9223372036854775807"},
		{"min integer", "RETURN -9223372036854775808"},
		{"large float", "RETURN 1.7976931348623157e308"},
		{"small float", "RETURN 4.9e-324"},
		{"empty string", "RETURN ''"},
		{"long string", "RETURN 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'"},
		{"empty list", "RETURN []"},
		{"nested empty lists", "RETURN [[[]]]"},
		{"large list", "RETURN [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20]"},
		{"empty map", "RETURN {}"},
		{"deeply nested map", "RETURN {a: {b: {c: {d: {e: 1}}}}}"},
		{"many labels", "MATCH (n:A:B:C:D:E:F:G:H:I:J) RETURN n"},
		{"deep property access", "MATCH (n) RETURN n.a.b.c.d.e.f.g.h.i.j"},
		{"deep list access", "RETURN [[[[[[[[[[1]]]]]]]]]][0][0][0][0][0][0][0][0][0][0]"},
		{"many conditions", "MATCH (n) WHERE n.a = 1 AND n.b = 2 AND n.c = 3 AND n.d = 4 AND n.e = 5 RETURN n"},
		{"many order by", "MATCH (n) RETURN n ORDER BY n.a, n.b, n.c, n.d, n.e"},
		{"large skip limit", "MATCH (n) RETURN n SKIP 999999999 LIMIT 999999999"},
		{"variable length max", "MATCH (a)-[*999]->(b) RETURN a, b"},
		{"variable length zero", "MATCH (a)-[*0]->(b) RETURN a, b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(tt.input)
			_, err := p.Parse()
			if err != nil {
				t.Errorf("failed to parse %q: %v", tt.input, err)
			}
		})
	}
}

func TestParser_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"unicode string", "RETURN '你好世界'"},
		{"emoji string", "RETURN '😀🎉🚀'"},
		{"escape sequences", `RETURN 'hello\n\t\r\\world'`},
		{"mixed quotes", `RETURN "it's a test"`},
		{"nested quotes", `RETURN 'he said "hello"'`},
		{"special chars in string", "RETURN 'a\\nb\\tc'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(tt.input)
			_, err := p.Parse()
			if err != nil {
				t.Errorf("failed to parse %q: %v", tt.input, err)
			}
		})
	}
}

func TestParser_UnicodeSupport(t *testing.T) {
	t.Skip("Unicode identifier support requires lexer refactoring to use rune-based processing")

	input := "MATCH (n) WHERE n.名字 = '测试' RETURN n"
	p := parser.New(input)
	_, err := p.Parse()
	if err != nil {
		t.Errorf("failed to parse %q: %v", input, err)
	}
}

func TestParser_ComplexNesting(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "nested case expressions",
			input: "RETURN CASE WHEN CASE WHEN true THEN 1 ELSE 2 END = 1 THEN 'a' ELSE 'b' END",
		},
		{
			name:  "nested list comprehensions",
			input: "MATCH (n) RETURN [x IN [y IN [1, 2, 3] | y * 2] | x + 1]",
		},
		{
			name:  "deeply nested expressions",
			input: "RETURN (((((1 + 2) * 3) - 4) / 5) ^ 2)",
		},
		{
			name:  "complex map with expressions",
			input: "RETURN {a: 1 + 2, b: {c: 3 * 4}, d: [1, 2, 3][0]}",
		},
		{
			name:  "multiple function calls",
			input: "RETURN COUNT(COLLECT(DISTINCT n.name))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(tt.input)
			_, err := p.Parse()
			if err != nil {
				t.Errorf("failed to parse %q: %v", tt.input, err)
			}
		})
	}
}

func TestParser_StatementSeparation(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"multiple statements", "MATCH (n) RETURN n; MATCH (m) RETURN m"},
		{"statement without semicolon", "MATCH (n) RETURN n MATCH (m) RETURN m"},
		{"trailing semicolon", "MATCH (n) RETURN n;"},
		{"multiple semicolons", "MATCH (n) RETURN n;; MATCH (m) RETURN m;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(tt.input)
			_, err := p.Parse()
			if err != nil {
				t.Errorf("failed to parse %q: %v", tt.input, err)
			}
		})
	}
}

func TestParser_ParameterizedQueries(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"single parameter", "MATCH (n) WHERE n.name = $name RETURN n"},
		{"multiple parameters", "MATCH (n) WHERE n.name = $name AND n.age > $minAge RETURN n"},
		{"parameter in create", "CREATE (n:Person $props)"},
		{"parameter in set", "MATCH (n) SET n += $updates"},
		{"parameter in list", "MATCH (n) WHERE n.id IN $ids RETURN n"},
		{"parameter in map", "RETURN {name: $name, age: $age}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(tt.input)
			_, err := p.Parse()
			if err != nil {
				t.Errorf("failed to parse %q: %v", tt.input, err)
			}
		})
	}
}

func TestParser_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"self relationship", "MATCH (a)-[:SELF]->(a) RETURN a"},
		{"zero hop", "MATCH (a)-[*0]-(a) RETURN a"},
		{"negative number", "RETURN -123"},
		{"negative float", "RETURN -3.14"},
		{"double negative", "RETURN --5"},
		{"scientific negative exponent", "RETURN 1e-10"},
		{"boolean in expression", "RETURN true AND false OR true"},
		{"null in expression", "RETURN COALESCE(NULL, 1)"},
		{"empty pattern", "MATCH () RETURN 1"},
		{"anonymous relationship", "MATCH ()-[]->() RETURN 1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(tt.input)
			_, err := p.Parse()
			if err != nil {
				t.Errorf("failed to parse %q: %v", tt.input, err)
			}
		})
	}
}
