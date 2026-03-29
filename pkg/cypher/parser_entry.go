package cypher

import (
	"github.com/DotNetAge/gograph/pkg/cypher/ast"
	"github.com/DotNetAge/gograph/pkg/cypher/parser"
)

type Parser struct {
	input string
}

func NewParser(input string) *Parser {
	return &Parser{input: input}
}

func (p *Parser) Parse() (*ast.Query, error) {
	psr := parser.New(p.input)
	return psr.Parse()
}

func Parse(input string) (*ast.Query, error) {
	return NewParser(input).Parse()
}
