// Package cypher provides Cypher query parsing and execution capabilities for gograph.
package cypher

import (
	"fmt"
	"strings"

	"github.com/DotNetAge/gograph/pkg/cypher/ast"
)

// Parser parses Cypher query strings into AST structures.
type Parser struct {
	input  string
	pos    int
	length int
}

// NewParser creates a new Parser for the given Cypher query string.
func NewParser(input string) *Parser {
	return &Parser{
		input:  strings.TrimSpace(input),
		length: len(input),
	}
}

// Parse parses the Cypher query and returns the AST representation.
func (p *Parser) Parse() (*ast.AST, error) {
	p.pos = 0
	var stmts []ast.Statement

	for p.pos < p.length {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
		p.skipWhitespace()
		if p.peek() == ';' {
			p.pos++
		}
		p.skipWhitespace()
	}

	return &ast.AST{Statements: stmts}, nil
}

// parseStatement parses a single Cypher statement.
func (p *Parser) parseStatement() (ast.Statement, error) {
	p.skipWhitespace()
	upper := strings.ToUpper(p.peekRemaining())
	var clause ast.Clause
	var err error

	if strings.HasPrefix(upper, "CREATE") {
		clause, err = p.parseCreate()
	} else if strings.HasPrefix(upper, "MATCH") {
		clause, err = p.parseMatch()
	} else if strings.HasPrefix(upper, "SET") {
		clause, err = p.parseSet()
	} else if strings.HasPrefix(upper, "DELETE") {
		clause, err = p.parseDelete()
	} else if strings.HasPrefix(upper, "REMOVE") {
		clause, err = p.parseRemove()
	} else {
		return ast.Statement{}, fmt.Errorf("unsupported statement: %s", p.peekRemaining())
	}

	if err != nil {
		return ast.Statement{}, err
	}

	return ast.Statement{Clause: clause}, nil
}

// parseCreate parses a CREATE clause.
func (p *Parser) parseCreate() (*ast.CreateClause, error) {
	if err := p.expectKeyword("CREATE"); err != nil {
		return nil, err
	}

	pattern, err := p.parsePattern()
	if err != nil {
		return nil, err
	}

	return &ast.CreateClause{Pattern: pattern}, nil
}

// parseMatch parses a MATCH clause with optional WHERE, RETURN, and DELETE.
func (p *Parser) parseMatch() (*ast.MatchClause, error) {
	if err := p.expectKeyword("MATCH"); err != nil {
		return nil, err
	}

	pattern, err := p.parsePattern()
	if err != nil {
		return nil, err
	}

	clause := &ast.MatchClause{Pattern: pattern}

	p.skipWhitespace()
	if strings.HasPrefix(strings.ToUpper(p.peekRemaining()), "WHERE") {
		where, err := p.parseWhere()
		if err != nil {
			return nil, err
		}
		clause.Where = where
	}

	p.skipWhitespace()
	if strings.HasPrefix(strings.ToUpper(p.peekRemaining()), "RETURN") {
		ret, err := p.parseReturn()
		if err != nil {
			return nil, err
		}
		clause.Return = ret
	}

	p.skipWhitespace()
	if strings.HasPrefix(strings.ToUpper(p.peekRemaining()), "DELETE") ||
		strings.HasPrefix(strings.ToUpper(p.peekRemaining()), "DETACH") {
		deleteClause, err := p.parseDelete()
		if err != nil {
			return nil, err
		}
		clause.Delete = deleteClause
	}

	return clause, nil
}

// parseSet parses a SET clause.
func (p *Parser) parseSet() (*ast.SetClause, error) {
	if err := p.expectKeyword("SET"); err != nil {
		return nil, err
	}

	var assignments []ast.Assignment
	for {
		p.skipWhitespace()
		if p.isAtEnd() {
			break
		}
		prop, err := p.parsePropertyAccess()
		if err != nil {
			return nil, err
		}
		if err := p.expect("="); err != nil {
			return nil, err
		}
		val, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, ast.Assignment{
			Property: ast.PropertyAccess{Node: prop.Node, Property: prop.Property},
			Value:    val,
		})
		p.skipWhitespace()
		if p.peek() == ',' {
			p.pos++
		} else {
			break
		}
	}

	return &ast.SetClause{Assignments: assignments}, nil
}

// parseDelete parses a DELETE or DETACH DELETE clause.
func (p *Parser) parseDelete() (*ast.DeleteClause, error) {
	isDetach := false
	if strings.HasPrefix(strings.ToUpper(p.peekRemaining()), "DETACH") {
		if err := p.expectKeyword("DETACH"); err != nil {
			return nil, err
		}
		isDetach = true
	}
	if err := p.expectKeyword("DELETE"); err != nil {
		return nil, err
	}

	var exprs []ast.Expression
	for {
		p.skipWhitespace()
		if p.isAtEnd() {
			break
		}
		expr, err := p.parseDeleteExpression()
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, expr)
		p.skipWhitespace()
		if p.peek() == ',' {
			p.pos++
		} else {
			break
		}
	}

	return &ast.DeleteClause{Detach: isDetach, Expressions: exprs}, nil
}

// parseDeleteExpression parses a single expression in a DELETE clause.
func (p *Parser) parseDeleteExpression() (ast.Expression, error) {
	p.skipWhitespace()

	if p.peek() == '(' {
		return p.parsePropertyAccess()
	}

	if p.peek() == 'r' {
		startPos := p.pos
		p.pos++
		nextChar := p.peek()
		if nextChar == ':' {
			p.pos++
			_, err := p.parseIdentifier()
			if err != nil {
				return nil, err
			}
			return &ast.RelationVariable{Name: "r"}, nil
		}
		p.pos = startPos
	}

	ident, err := p.parseIdentifier()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	if p.peek() == '.' {
		p.pos++
		prop, err := p.parseIdentifier()
		if err != nil {
			return nil, err
		}
		return &ast.PropertyLookup{Node: ident.Name, Property: prop.Name}, nil
	}

	if ident.Name == "r" {
		return &ast.RelationVariable{Name: ident.Name}, nil
	}

	return &ast.PropertyLookup{Node: ident.Name, Property: ""}, nil
}

// parseRemove parses a REMOVE clause.
func (p *Parser) parseRemove() (*ast.RemoveClause, error) {
	if err := p.expectKeyword("REMOVE"); err != nil {
		return nil, err
	}

	var removals []ast.RemoveItem
	for {
		p.skipWhitespace()
		if p.isAtEnd() {
			break
		}

		if p.peek() == '(' {
			prop, err := p.parsePropertyAccess()
			if err != nil {
				return nil, err
			}
			removals = append(removals, ast.RemoveItem{
				Type:     ast.RemoveItemTypeProperty,
				Property: ast.PropertyAccess{Node: prop.Node, Property: prop.Property},
			})
		} else if p.isIdentifierStart() {
			ident, err := p.parseIdentifier()
			if err != nil {
				return nil, err
			}
			p.skipWhitespace()
			if p.peek() == ':' {
				p.pos++
				label, err := p.parseIdentifier()
				if err != nil {
					return nil, err
				}
				removals = append(removals, ast.RemoveItem{
					Type:  ast.RemoveItemTypeLabel,
					Label: label.Name,
					Property: ast.PropertyAccess{Node: ident.Name},
				})
			} else if p.peek() == '.' {				p.pos++
				prop, err := p.parseIdentifier()
				if err != nil {
					return nil, err
				}
				removals = append(removals, ast.RemoveItem{
					Type:     ast.RemoveItemTypeProperty,
					Property: ast.PropertyAccess{Node: ident.Name, Property: prop.Name},
				})
			}
		}

		p.skipWhitespace()
		if p.peek() == ',' {
			p.pos++
		} else {
			break
		}
	}

	return &ast.RemoveClause{Removals: removals}, nil
}

// parseReturn parses a RETURN clause.
func (p *Parser) parseReturn() (*ast.ReturnClause, error) {
	if err := p.expectKeyword("RETURN"); err != nil {
		return nil, err
	}

	var items []ast.ReturnItem
	for {
		p.skipWhitespace()
		if p.isAtEnd() {
			break
		}
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		alias := ""
		p.skipWhitespace()
		if strings.HasPrefix(strings.ToUpper(p.peekRemaining()), "AS") {
			p.pos += 2
			p.skipWhitespace()
			id, err := p.parseIdentifier()
			if err != nil {
				return nil, err
			}
			alias = id.Name
		}
		items = append(items, ast.ReturnItem{Expression: expr, Alias: alias})
		p.skipWhitespace()
		if p.peek() == ',' {
			p.pos++
		} else {
			break
		}
	}

	return &ast.ReturnClause{Items: items}, nil
}

// parseWhere parses a WHERE clause.
func (p *Parser) parseWhere() (*ast.WhereClause, error) {
	if err := p.expectKeyword("WHERE"); err != nil {
		return nil, err
	}
	p.skipWhitespace()

	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return &ast.WhereClause{Expression: expr}, nil
}

// parsePattern parses a graph pattern.
func (p *Parser) parsePattern() (ast.Pattern, error) {
	var elements []ast.PatternElement

	for {
		p.skipWhitespace()
		if p.peek() == '(' {
			node, err := p.parseNodePattern()
			if err != nil {
				return ast.Pattern{}, err
			}
			elem := ast.PatternElement{Node: node}

			p.skipWhitespace()
			if p.peek() == '-' {
				rel, dir, err := p.parseRelationPattern()
				if err == nil && rel != nil {
					elem.Relation = rel

					if dir == ast.RelDirIncoming {
						p.skipWhitespace()
						node2, err := p.parseNodePattern()
						if err != nil {
							return ast.Pattern{}, err
						}
						rel.EndNode = node2
					} else if dir == ast.RelDirOutgoing {
						p.skipWhitespace()
						node2, err := p.parseNodePattern()
						if err != nil {
							return ast.Pattern{}, err
						}
						rel.EndNode = node2
					}
				}
			}

			elements = append(elements, elem)
		}
		p.skipWhitespace()
		if p.peek() == ',' {
			p.pos++
			continue
		}
		break
	}

	return ast.Pattern{Elements: elements}, nil
}

// parseNodePattern parses a node pattern like (n:Label {prop: value}).
func (p *Parser) parseNodePattern() (*ast.NodePattern, error) {
	if err := p.expect("("); err != nil {
		return nil, err
	}

	node := &ast.NodePattern{Labels: []string{}, Properties: make(map[string]interface{})}

	p.skipWhitespace()

	if p.isIdentifierStart() && p.peek() != ':' {
		ident, err := p.parseIdentifier()
		if err == nil {
			node.Variable = ident.Name
			p.skipWhitespace()
		}
	}

	for p.peek() == ':' {
		p.pos++
		label, err := p.parseIdentifier()
		if err != nil {
			return nil, err
		}
		node.Labels = append(node.Labels, label.Name)
		p.skipWhitespace()
	}

	if p.peek() == '{' {
		props, err := p.parseProperties()
		if err != nil {
			return nil, err
		}
		node.Properties = props
	}

	if err := p.expect(")"); err != nil {
		return nil, err
	}

	return node, nil
}

// parseRelationPattern parses a relationship pattern like -[r:TYPE]->.
func (p *Parser) parseRelationPattern() (*ast.RelationPattern, ast.RelDirection, error) {
	if err := p.expect("-"); err != nil {
		return nil, "", err
	}

	dir := ast.RelDirBoth
	rel := &ast.RelationPattern{Properties: make(map[string]interface{})}

	p.skipWhitespace()

	if p.peek() == '<' {
		p.pos++
		if err := p.expect("-"); err != nil {
			return nil, "", err
		}
		dir = ast.RelDirIncoming
	}

	p.skipWhitespace()

	if p.peek() == '[' {
		p.pos++
		p.skipWhitespace()

		if p.isIdentifierStart() && p.peek() != ':' {
			ident, err := p.parseIdentifier()
			if err == nil {
				rel.Variable = ident.Name
				p.skipWhitespace()
			}
		}

		if p.peek() == ':' {
			p.pos++
			relType, err := p.parseIdentifier()
			if err != nil {
				return nil, "", err
			}
			rel.RelType = relType.Name
		}

		p.skipWhitespace()
		if p.peek() == '{' {
			props, err := p.parseProperties()
			if err != nil {
				return nil, "", err
			}
			rel.Properties = props
		}
		if err := p.expect("]"); err != nil {
			return nil, "", err
		}
	}

	p.skipWhitespace()

	if p.peek() == '>' {
		p.pos++
		dir = ast.RelDirOutgoing
	} else if p.peek() == '-' {
		p.pos++
		if p.peek() == '>' {
			p.pos++
			dir = ast.RelDirOutgoing
		}
	}

	return rel, dir, nil
}

// parseProperties parses a property map like {key: value, key2: value2}.
func (p *Parser) parseProperties() (map[string]interface{}, error) {
	if err := p.expect("{"); err != nil {
		return nil, err
	}

	props := make(map[string]interface{})

	p.skipWhitespace()
	if p.peek() == '}' {
		p.pos++
		return props, nil
	}

	for {
		key, err := p.parseIdentifier()
		if err != nil {
			return nil, err
		}
		p.skipWhitespace()
		if err := p.expect(":"); err != nil {
			return nil, err
		}
		p.skipWhitespace()
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		props[key.Name] = val
		p.skipWhitespace()
		if p.peek() == ',' {
			p.pos++
		} else {
			break
		}
	}

	if err := p.expect("}"); err != nil {
		return nil, err
	}

	return props, nil
}

// parseExpression parses an expression.
func (p *Parser) parseExpression() (ast.Expression, error) {
	p.skipWhitespace()

	left, err := p.parseLogicalOr()
	if err != nil {
		return nil, err
	}

	return left, nil
}

// parseLogicalOr parses an OR expression.
func (p *Parser) parseLogicalOr() (ast.Expression, error) {
	left, err := p.parseLogicalAnd()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	for strings.HasPrefix(p.peekRemaining(), "OR") {
		p.pos += 2
		p.skipWhitespace()
		right, err := p.parseLogicalAnd()
		if err != nil {
			return nil, err
		}
		left = &ast.ComparisonOp{Left: left, Operator: "OR", Right: right}
		p.skipWhitespace()
	}

	return left, nil
}

// parseLogicalAnd parses an AND expression.
func (p *Parser) parseLogicalAnd() (ast.Expression, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	for strings.HasPrefix(p.peekRemaining(), "AND") {
		p.pos += 3
		p.skipWhitespace()
		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		left = &ast.ComparisonOp{Left: left, Operator: "AND", Right: right}
		p.skipWhitespace()
	}

	return left, nil
}

// parseComparison parses a comparison expression.
func (p *Parser) parseComparison() (ast.Expression, error) {
	p.skipWhitespace()

	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	ops := []string{">=", "<=", "!=", "=", ">", "<"}
	for _, op := range ops {
		if strings.HasPrefix(p.peekRemaining(), op) {
			p.pos += len(op)
			p.skipWhitespace()
			right, err := p.parsePrimary()
			if err != nil {
				return nil, err
			}
			return &ast.ComparisonOp{Left: left, Operator: op, Right: right}, nil
		}
	}

	return left, nil
}

// parsePrimary parses a primary expression.
func (p *Parser) parsePrimary() (ast.Expression, error) {
	p.skipWhitespace()

	if p.peek() == '$' {
		p.pos++
		ident, err := p.parseIdentifier()
		if err != nil {
			return nil, err
		}
		return &ast.Identifier{Name: "$" + ident.Name}, nil
	}

	if p.peek() == '\'' || p.peek() == '"' {
		return p.parseStringLiteral()
	}

	if p.isDigit() {
		return p.parseNumberLiteral()
	}

	if strings.HasPrefix(p.peekRemaining(), "true") {
		p.pos += 4
		return &ast.Literal{Value: true}, nil
	}
	if strings.HasPrefix(p.peekRemaining(), "false") {
		p.pos += 5
		return &ast.Literal{Value: false}, nil
	}

	if p.peek() == '(' {
		p.pos++
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if err := p.expect(")"); err != nil {
			return nil, err
		}
		return expr, nil
	}

	return p.parsePropertyAccess()
}

// parsePropertyAccess parses a property access expression like n.property.
func (p *Parser) parsePropertyAccess() (*ast.PropertyLookup, error) {
	ident, err := p.parseIdentifier()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	if p.peek() == '.' {
		p.pos++
		prop, err := p.parseIdentifier()
		if err != nil {
			return nil, err
		}
		return &ast.PropertyLookup{Node: ident.Name, Property: prop.Name}, nil
	}

	return &ast.PropertyLookup{Node: ident.Name, Property: ""}, nil
}

// parseIdentifier parses an identifier.
func (p *Parser) parseIdentifier() (ast.Identifier, error) {
	p.skipWhitespace()
	start := p.pos

	if !p.isIdentifierStart() {
		return ast.Identifier{}, fmt.Errorf("expected identifier start at position %d", p.pos)
	}
	p.pos++
	for p.isIdentifierPart() {
		p.pos++
	}

	name := p.input[start:p.pos]
	return ast.Identifier{Name: name}, nil
}

// parseStringLiteral parses a quoted string literal.
func (p *Parser) parseStringLiteral() (*ast.Literal, error) {
	quote := p.peek()
	p.pos++

	start := p.pos
	for p.pos < p.length && p.peek() != quote {
		if p.peek() == '\\' && p.pos+1 < p.length {
			p.pos += 2
			continue
		}
		p.pos++
	}
	value := p.input[start:p.pos]

	if p.pos >= p.length {
		return nil, fmt.Errorf("unterminated string literal")
	}
	p.pos++

	return &ast.Literal{Value: value}, nil
}

// parseNumberLiteral parses a numeric literal.
func (p *Parser) parseNumberLiteral() (*ast.Literal, error) {
	start := p.pos

	for p.isDigit() {
		p.pos++
	}

	if p.peek() == '.' {
		p.pos++
		for p.isDigit() {
			p.pos++
		}
	}

	valueStr := p.input[start:p.pos]
	var value interface{}

	if strings.Contains(valueStr, ".") {
		var f float64
		fmt.Sscanf(valueStr, "%f", &f)
		value = f
	} else {
		var i int64
		fmt.Sscanf(valueStr, "%d", &i)
		value = i
	}

	return &ast.Literal{Value: value}, nil
}

// parseValue parses a value (used in property maps).
func (p *Parser) parseValue() (interface{}, error) {
	p.skipWhitespace()

	if p.peek() == '$' {
		p.pos++
		ident, err := p.parseIdentifier()
		if err != nil {
			return nil, err
		}
		return "$" + ident.Name, nil
	}

	if p.peek() == '\'' || p.peek() == '"' {
		lit, err := p.parseStringLiteral()
		if err != nil {
			return nil, err
		}
		return lit.Value, nil
	}

	if p.isDigit() {
		lit, err := p.parseNumberLiteral()
		if err != nil {
			return nil, err
		}
		return lit.Value, nil
	}

	if strings.HasPrefix(p.peekRemaining(), "true") {
		p.pos += 4
		return true, nil
	}
	if strings.HasPrefix(p.peekRemaining(), "false") {
		p.pos += 5
		return false, nil
	}

	ident, err := p.parseIdentifier()
	if err != nil {
		return nil, err
	}
	return ident.Name, nil
}

// expectKeyword expects a specific keyword at the current position.
func (p *Parser) expectKeyword(keyword string) error {
	p.skipWhitespace()
	upper := strings.ToUpper(p.peekRemaining())
	if !strings.HasPrefix(upper, strings.ToUpper(keyword)) {
		return fmt.Errorf("expected keyword %s at position %d, got %s", keyword, p.pos, p.peekRemaining())
	}
	p.pos += len(keyword)
	return nil
}

// expect expects a specific string at the current position.
func (p *Parser) expect(s string) error {
	p.skipWhitespace()
	if p.pos+len(s) > p.length {
		return fmt.Errorf("expected %s at position %d, got end of input", s, p.pos)
	}
	if p.input[p.pos:p.pos+len(s)] != s {
		return fmt.Errorf("expected %s at position %d, got %s", s, p.pos, p.input[p.pos:p.pos+len(s)])
	}
	p.pos += len(s)
	return nil
}

// peek returns the current character without advancing.
func (p *Parser) peek() byte {
	if p.pos >= p.length {
		return 0
	}
	return p.input[p.pos]
}

// peekRemaining returns the remaining unparsed input.
func (p *Parser) peekRemaining() string {
	if p.pos >= p.length {
		return ""
	}
	return p.input[p.pos:]
}

// skipWhitespace skips whitespace characters.
func (p *Parser) skipWhitespace() {
	for p.pos < p.length && (p.input[p.pos] == ' ' || p.input[p.pos] == '\t' || p.input[p.pos] == '\n' || p.input[p.pos] == '\r') {
		p.pos++
	}
}

// isAtEnd returns true if the parser has reached the end of input.
func (p *Parser) isAtEnd() bool {
	return p.pos >= p.length
}

// isIdentifierStart returns true if the current character can start an identifier.
func (p *Parser) isIdentifierStart() bool {
	if p.pos >= p.length {
		return false
	}
	c := p.input[p.pos]
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

// isIdentifierPart returns true if the current character can be part of an identifier.
func (p *Parser) isIdentifierPart() bool {
	if p.pos >= p.length {
		return false
	}
	c := p.input[p.pos]
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// isDigit returns true if the current character is a digit.
func (p *Parser) isDigit() bool {
	if p.pos >= p.length {
		return false
	}
	c := p.input[p.pos]
	return c >= '0' && c <= '9'
}
