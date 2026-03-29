package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser_MatchStatement(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name:     "simple match",
			input:    "MATCH (n:Person) RETURN n",
			hasError: false,
		},
		{
			name:     "match with where",
			input:    "MATCH (n:Person) WHERE n.age > 18 RETURN n",
			hasError: false,
		},
		{
			name:     "match with relationship",
			input:    "MATCH (a:Person)-[:KNOWS]->(b:Person) RETURN a, b",
			hasError: false,
		},
		{
			name:     "match with variable length path",
			input:    "MATCH (a)-[:KNOWS*1..3]->(b) RETURN a, b",
			hasError: false,
		},
		{
			name:     "match with multiple labels",
			input:    "MATCH (n:Person:Employee) RETURN n",
			hasError: false,
		},
		{
			name:     "match with properties",
			input:    "MATCH (n:Person {name: 'Alice', age: 30}) RETURN n",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.input)
			_, err := p.Parse()
			if tt.hasError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParser_CreateStatement(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name:     "create single node",
			input:    "CREATE (n:Person {name: 'Alice'})",
			hasError: false,
		},
		{
			name:     "create with relationship",
			input:    "CREATE (a:Person)-[:KNOWS]->(b:Person)",
			hasError: false,
		},
		{
			name:     "create multiple nodes",
			input:    "CREATE (a:Person), (b:Company)",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.input)
			_, err := p.Parse()
			if tt.hasError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParser_SetStatement(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name:     "set single property",
			input:    "MATCH (n:Person) SET n.age = 30",
			hasError: false,
		},
		{
			name:     "set multiple properties",
			input:    "MATCH (n:Person) SET n.age = 30, n.city = 'Beijing'",
			hasError: false,
		},
		{
			name:     "set with plus equals",
			input:    "MATCH (n:Person) SET n.tags += ['new']",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.input)
			_, err := p.Parse()
			if tt.hasError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParser_DeleteStatement(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name:     "delete node",
			input:    "MATCH (n:Person) DELETE n",
			hasError: false,
		},
		{
			name:     "detach delete",
			input:    "MATCH (n:Person) DETACH DELETE n",
			hasError: false,
		},
		{
			name:     "delete relationship",
			input:    "MATCH (a)-[r:KNOWS]->(b) DELETE r",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.input)
			_, err := p.Parse()
			if tt.hasError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParser_ReturnClause(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name:     "return with alias",
			input:    "MATCH (n:Person) RETURN n.name AS userName",
			hasError: false,
		},
		{
			name:     "return distinct",
			input:    "MATCH (n:Person) RETURN DISTINCT n.city",
			hasError: false,
		},
		{
			name:     "return with order by",
			input:    "MATCH (n:Person) RETURN n ORDER BY n.age DESC",
			hasError: false,
		},
		{
			name:     "return with pagination",
			input:    "MATCH (n:Person) RETURN n ORDER BY n.name SKIP 10 LIMIT 5",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.input)
			_, err := p.Parse()
			if tt.hasError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParser_WhereClause(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name:     "where with comparison",
			input:    "MATCH (n:Person) WHERE n.age > 18 RETURN n",
			hasError: false,
		},
		{
			name:     "where with and",
			input:    "MATCH (n:Person) WHERE n.age > 18 AND n.city = 'Beijing' RETURN n",
			hasError: false,
		},
		{
			name:     "where with or",
			input:    "MATCH (n:Person) WHERE n.city = 'Beijing' OR n.city = 'Shanghai' RETURN n",
			hasError: false,
		},
		{
			name:     "where with in",
			input:    "MATCH (n:Person) WHERE n.city IN ['Beijing', 'Shanghai'] RETURN n",
			hasError: false,
		},
		{
			name:     "where with is null",
			input:    "MATCH (n:Person) WHERE n.email IS NOT NULL RETURN n",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.input)
			_, err := p.Parse()
			if tt.hasError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParser_Expressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name:     "arithmetic expression",
			input:    "MATCH (n:Person) RETURN n.age * 2 + 10",
			hasError: false,
		},
		{
			name:     "function call",
			input:    "MATCH (n:Person) RETURN COUNT(n) AS total",
			hasError: false,
		},
		{
			name:     "list literal",
			input:    "MATCH (n:Person) RETURN [1, 2, 3] AS numbers",
			hasError: false,
		},
		{
			name:     "map literal",
			input:    "MATCH (n:Person) RETURN {name: n.name, age: n.age} AS info",
			hasError: false,
		},
		{
			name:     "parameter",
			input:    "MATCH (n:Person) WHERE n.name = $name RETURN n",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.input)
			_, err := p.Parse()
			if tt.hasError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
