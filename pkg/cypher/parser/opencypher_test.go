package parser_test

import (
	"testing"

	"github.com/DotNetAge/gograph/pkg/cypher/parser"
)

func TestOpenCypher_BasicSyntax(t *testing.T) {
	t.Run("case_insensitivity", func(t *testing.T) {
		tests := []string{
			"MATCH (n:Person) RETURN n",
			"match (n:Person) return n",
			"Match (n:Person) Return n",
			"MATCH (n:person) RETURN n",
		}
		for _, input := range tests {
			p := parser.New(input)
			_, err := p.Parse()
			if err != nil {
				t.Errorf("failed to parse %q: %v", input, err)
			}
		}
	})

	t.Run("comments", func(t *testing.T) {
		tests := []string{
			"MATCH // single line comment\n(n:Person) RETURN n",
			"MATCH /* multi\nline\ncomment */ (n:Person) RETURN n",
		}
		for _, input := range tests {
			p := parser.New(input)
			_, err := p.Parse()
			if err != nil {
				t.Errorf("failed to parse %q: %v", input, err)
			}
		}
	})
}

func TestOpenCypher_DataTypes(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"integer", "RETURN 12345"},
		{"negative integer", "RETURN -5"},
		{"float", "RETURN 3.14"},
		{"scientific notation", "RETURN 1.5e10"},
		{"string single quote", "RETURN 'Alice'"},
		{"string double quote", `RETURN "Alice"`},
		{"boolean true", "RETURN true"},
		{"boolean false", "RETURN false"},
		{"null", "RETURN null"},
		{"list", "RETURN [1, 2, 3]"},
		{"nested list", "RETURN [[1, 2], [3, 4]]"},
		{"map", "RETURN {name: 'Alice', age: 30}"},
		{"nested map", "RETURN {person: {name: 'Alice'}}"},
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

func TestOpenCypher_NodePattern(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty node", "MATCH () RETURN 1"},
		{"variable only", "MATCH (n) RETURN n"},
		{"single label", "MATCH (n:Person) RETURN n"},
		{"multiple labels", "MATCH (n:Person:Employee:Manager) RETURN n"},
		{"properties", "MATCH (n {name: 'Alice', age: 30}) RETURN n"},
		{"label and properties", "MATCH (n:Person {name: 'Alice'}) RETURN n"},
		{"variable label properties", "MATCH (p:Person:Employee {name: 'Alice', active: true}) RETURN p"},
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

func TestOpenCypher_RelationshipPattern(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"undirected", "MATCH (a)--(b) RETURN a, b"},
		{"outgoing", "MATCH (a)->(b) RETURN a, b"},
		{"incoming", "MATCH (a)<-(b) RETURN a, b"},
		{"with type", "MATCH (a)-[:KNOWS]->(b) RETURN a, b"},
		{"with variable", "MATCH (a)-[r]->(b) RETURN r"},
		{"with properties", "MATCH (a)-[:KNOWS {since: 2020}]->(b) RETURN a, b"},
		{"full relationship", "MATCH (a)-[r:KNOWS {since: 2020, status: 'close'}]->(b) RETURN r"},
		{"variable length min", "MATCH (a)-[:KNOWS*1]->(b) RETURN a, b"},
		{"variable length range", "MATCH (a)-[:KNOWS*1..5]->(b) RETURN a, b"},
		{"variable length unbounded", "MATCH (a)-[:KNOWS*]->(b) RETURN a, b"},
		{"variable length with properties", "MATCH (a)-[:KNOWS*1..3 {active: true}]->(b) RETURN a, b"},
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

func TestOpenCypher_Pattern(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple path", "MATCH (a:Person)-[:KNOWS]->(b:Person) RETURN a, b"},
		{"multi hop", "MATCH (a)-[:KNOWS]->(b)-[:FRIENDS_WITH]->(c) RETURN a, c"},
		{"multiple patterns", "MATCH (a:Person), (b:Company) RETURN a, b"},
		{"path variable", "MATCH path = (a)-[:KNOWS]->(b) RETURN path"},
		{"complex path", "MATCH (a:Person {name: 'Alice'})-[:WORKS_FOR]->(c:Company {name: 'TechCorp'})<-[:WORKS_FOR]-(b:Person) RETURN a, b, c"},
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

func TestOpenCypher_MatchClause(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"basic match", "MATCH (n:Person) RETURN n"},
		{"match with where", "MATCH (n:Person) WHERE n.age > 18 RETURN n"},
		{"optional match", "OPTIONAL MATCH (n:Person) RETURN n"},
		{"match with return", "MATCH (n:Person) RETURN n.name, n.age"},
		{"match with delete", "MATCH (n:Person) DELETE n"},
		{"match with detach delete", "MATCH (n:Person) DETACH DELETE n"},
		{"match with set", "MATCH (n:Person) SET n.updated = true"},
		{"match where return", "MATCH (n:Person) WHERE n.active = true RETURN n"},
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

func TestOpenCypher_WhereClause(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"comparison", "MATCH (n) WHERE n.age > 18 RETURN n"},
		{"equality", "MATCH (n) WHERE n.name = 'Alice' RETURN n"},
		{"inequality", "MATCH (n) WHERE n.name != 'Bob' RETURN n"},
		{"and", "MATCH (n) WHERE n.age > 18 AND n.active = true RETURN n"},
		{"or", "MATCH (n) WHERE n.city = 'Beijing' OR n.city = 'Shanghai' RETURN n"},
		{"not", "MATCH (n) WHERE NOT n.deleted RETURN n"},
		{"parentheses", "MATCH (n) WHERE (n.age > 18 OR n.vip = true) AND n.active = true RETURN n"},
		{"in list", "MATCH (n) WHERE n.city IN ['Beijing', 'Shanghai'] RETURN n"},
		{"is null", "MATCH (n) WHERE n.email IS NULL RETURN n"},
		{"is not null", "MATCH (n) WHERE n.email IS NOT NULL RETURN n"},
		{"exists", "MATCH (n) WHERE EXISTS(n.email) RETURN n"},
		{"contains", "MATCH (n) WHERE n.name CONTAINS 'Li' RETURN n"},
		{"starts with", "MATCH (n) WHERE n.name STARTS WITH 'A' RETURN n"},
		{"ends with", "MATCH (n) WHERE n.name ENDS WITH 'e' RETURN n"},
		{"complex condition", "MATCH (n) WHERE n.age >= 18 AND n.age < 65 AND n.active = true RETURN n"},
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

func TestOpenCypher_ReturnClause(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"single property", "MATCH (n) RETURN n.name"},
		{"multiple properties", "MATCH (n) RETURN n.name, n.age, n.city"},
		{"with alias", "MATCH (n) RETURN n.name AS userName"},
		{"distinct", "MATCH (n) RETURN DISTINCT n.city"},
		{"order by asc", "MATCH (n) RETURN n ORDER BY n.name ASC"},
		{"order by desc", "MATCH (n) RETURN n ORDER BY n.age DESC"},
		{"order by multiple", "MATCH (n) RETURN n ORDER BY n.age DESC, n.name ASC"},
		{"skip", "MATCH (n) RETURN n SKIP 10"},
		{"limit", "MATCH (n) RETURN n LIMIT 5"},
		{"skip and limit", "MATCH (n) RETURN n SKIP 10 LIMIT 5"},
		{"order skip limit", "MATCH (n) RETURN n ORDER BY n.name SKIP 10 LIMIT 5"},
		{"expression", "MATCH (n) RETURN n.age * 2 AS doubleAge"},
		{"function call", "MATCH (n) RETURN COUNT(n) AS total"},
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

func TestOpenCypher_WithClause(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"basic with", "MATCH (n) WITH n RETURN n"},
		{"with alias", "MATCH (n) WITH n AS node RETURN node"},
		{"with aggregation", "MATCH (n)-[r]->(m) WITH n, COUNT(r) AS relCount RETURN n, relCount"},
		{"with where", "MATCH (n) WITH n WHERE n.age > 18 RETURN n"},
		{"with order by", "MATCH (n) WITH n ORDER BY n.name RETURN n"},
		{"with skip limit", "MATCH (n) WITH n SKIP 5 LIMIT 10 RETURN n"},
		{"with distinct", "MATCH (n) WITH DISTINCT n.city AS city RETURN city"},
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

func TestOpenCypher_CreateClause(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"single node", "CREATE (n:Person {name: 'Alice'})"},
		{"multiple nodes", "CREATE (a:Person), (b:Company)"},
		{"node with relationship", "CREATE (a:Person)-[:KNOWS]->(b:Person)"},
		{"multi label", "CREATE (n:Person:Employee {name: 'Alice'})"},
		{"complex create", "CREATE (a:Person {name: 'Alice'})-[:WORKS_FOR {since: 2020}]->(c:Company {name: 'TechCorp'})"},
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

func TestOpenCypher_MergeClause(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"basic merge", "MERGE (n:Person {name: 'Alice'})"},
		{"merge on create", "MERGE (n:Person {name: 'Alice'}) ON CREATE SET n.createdAt = timestamp()"},
		{"merge on match", "MERGE (n:Person {name: 'Alice'}) ON MATCH SET n.updatedAt = timestamp()"},
		{"merge both", "MERGE (n:Person {name: 'Alice'}) ON CREATE SET n.new = true ON MATCH SET n.existing = true"},
		{"merge relationship", "MERGE (a)-[:KNOWS]->(b)"},
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

func TestOpenCypher_SetClause(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"single property", "MATCH (n) SET n.age = 30"},
		{"multiple properties", "MATCH (n) SET n.age = 30, n.city = 'Beijing'"},
		{"plus equals", "MATCH (n) SET n.tags += ['new']"},
		{"add label", "MATCH (n) SET n:Employee"},
		{"multiple labels", "MATCH (n) SET n:Employee:Manager"},
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

func TestOpenCypher_DeleteClause(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"delete node", "MATCH (n) DELETE n"},
		{"detach delete", "MATCH (n) DETACH DELETE n"},
		{"delete relationship", "MATCH ()-[r]->() DELETE r"},
		{"delete multiple", "MATCH (a)-[r]->(b) DELETE a, r, b"},
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

func TestOpenCypher_RemoveClause(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"remove property", "MATCH (n) REMOVE n.temp"},
		{"remove multiple properties", "MATCH (n) REMOVE n.temp, n.flag"},
		{"remove label", "MATCH (n) REMOVE n:Employee"},
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

func TestOpenCypher_UnwindClause(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"unwind list", "UNWIND [1, 2, 3] AS n RETURN n"},
		{"unwind with operation", "UNWIND [1, 2, 3] AS n RETURN n * 2"},
		{"unwind property", "MATCH (n) UNWIND n.tags AS tag RETURN tag"},
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

func TestOpenCypher_Expressions(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"arithmetic", "RETURN 1 + 2 * 3 - 4 / 2"},
		{"modulo", "RETURN 10 % 3"},
		{"power", "RETURN 2 ^ 10"},
		{"parentheses", "RETURN (1 + 2) * 3"},
		{"unary minus", "RETURN -5"},
		{"property access", "MATCH (n) RETURN n.name"},
		{"nested property", "MATCH (n) RETURN n.address.city"},
		{"list index", "RETURN [1, 2, 3][0]"},
		{"list slice", "RETURN [1, 2, 3, 4, 5][1..3]"},
		{"case simple", "RETURN CASE n.age WHEN 18 THEN 'adult' ELSE 'other' END"},
		{"case searched", "RETURN CASE WHEN n.age < 18 THEN 'minor' WHEN n.age >= 18 THEN 'adult' END"},
		{"parameter", "MATCH (n) WHERE n.name = $name RETURN n"},
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

func TestOpenCypher_Functions(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"count", "MATCH (n) RETURN COUNT(n)"},
		{"sum", "MATCH (n) RETURN SUM(n.age)"},
		{"avg", "MATCH (n) RETURN AVG(n.age)"},
		{"min", "MATCH (n) RETURN MIN(n.age)"},
		{"max", "MATCH (n) RETURN MAX(n.age)"},
		{"collect", "MATCH (n) RETURN COLLECT(n.name)"},
		{"head", "RETURN HEAD([1, 2, 3])"},
		{"last", "RETURN LAST([1, 2, 3])"},
		{"tail", "RETURN TAIL([1, 2, 3])"},
		{"size", "RETURN SIZE([1, 2, 3])"},
		{"range", "RETURN RANGE(1, 10)"},
		{"reverse", "RETURN REVERSE([1, 2, 3])"},
		{"abs", "RETURN ABS(-5)"},
		{"ceil", "RETURN CEIL(3.14)"},
		{"floor", "RETURN FLOOR(3.99)"},
		{"round", "RETURN ROUND(3.5)"},
		{"rand", "RETURN RAND()"},
		{"toupper", "RETURN TOUPPER('hello')"},
		{"tolower", "RETURN TOLOWER('HELLO')"},
		{"replace", "RETURN REPLACE('hello', 'l', 'L')"},
		{"substring", "RETURN SUBSTRING('hello', 1, 3)"},
		{"trim", "RETURN TRIM('  hello  ')"},
		{"id", "MATCH (n) RETURN ID(n)"},
		{"labels", "MATCH (n) RETURN LABELS(n)"},
		{"type", "MATCH ()-[r]->() RETURN TYPE(r)"},
		{"properties", "MATCH (n) RETURN PROPERTIES(n)"},
		{"coalesce", "RETURN COALESCE(NULL, 'default')"},
		{"nullif", "RETURN NULLIF('value', 'value')"},
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

func TestOpenCypher_ComplexQueries(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "multi clause query",
			input: "MATCH (n:Person) WHERE n.age > 18 WITH n ORDER BY n.name SKIP 10 LIMIT 5 RETURN n.name, n.age",
		},
		{
			name:  "aggregation with grouping",
			input: "MATCH (n:Person)-[:WORKS_FOR]->(c:Company) WITH c, COUNT(n) AS employees RETURN c.name, employees",
		},
		{
			name:  "complex filter",
			input: "MATCH (n:Person) WHERE (n.age > 18 AND n.active = true) OR n.vip = true RETURN n",
		},
		{
			name:  "path analysis",
			input: "MATCH path = (a)-[:KNOWS*1..3]->(b) WHERE a.name = 'Alice' RETURN path",
		},
		{
			name:  "create and return",
			input: "CREATE (n:Person {name: 'Alice', age: 30}) RETURN n",
		},
		{
			name:  "merge with conditions",
			input: "MERGE (n:Person {name: 'Alice'}) ON CREATE SET n.createdAt = timestamp() ON MATCH SET n.updatedAt = timestamp() RETURN n",
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
