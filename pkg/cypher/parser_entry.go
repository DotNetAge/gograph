package cypher

import (
	"github.com/DotNetAge/gograph/pkg/cypher/ast"
	"github.com/DotNetAge/gograph/pkg/cypher/parser"
)

// Parser provides a high-level interface for parsing Cypher queries.
// It wraps the internal parser implementation and provides a simplified API.
//
// The parser converts Cypher query strings into an Abstract Syntax Tree (AST)
// that can be used for query execution or analysis.
//
// Example:
//
//	p := cypher.NewParser("MATCH (n:Person) RETURN n.name")
//	query, err := p.Parse()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// Use the parsed AST

type Parser struct {
	input string
}

// NewParser creates a new Parser for the given Cypher query string.
//
// Parameters:
//   - input: The Cypher query string to parse
//
// Returns a new Parser instance.
//
// Example:
//
//	p := cypher.NewParser("MATCH (n:Person) RETURN n.name")
//	query, err := p.Parse()
func NewParser(input string) *Parser {
	return &Parser{input: input}
}

// Parse parses the Cypher query and returns the AST.
//
// Returns the parsed Query AST, or an error if parsing fails.
//
// Example:
//
//	p := cypher.NewParser("MATCH (n:Person) RETURN n.name")
//	query, err := p.Parse()
//	if err != nil {
//	    log.Fatal(err)
//	}
func (p *Parser) Parse() (*ast.Query, error) {
	psr := parser.New(p.input)
	return psr.Parse()
}

// Parse is a convenience function that parses a Cypher query string.
// It creates a new parser and parses the query in one call.
//
// Parameters:
//   - input: The Cypher query string to parse
//
// Returns the parsed Query AST, or an error if parsing fails.
//
// Example:
//
//	query, err := cypher.Parse("MATCH (n:Person) RETURN n.name")
//	if err != nil {
//	    log.Fatal(err)
//	}
func Parse(input string) (*ast.Query, error) {
	return NewParser(input).Parse()
}
