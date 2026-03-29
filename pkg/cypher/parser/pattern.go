package parser

import (
	"github.com/DotNetAge/gograph/pkg/cypher/ast"
	"github.com/DotNetAge/gograph/pkg/cypher/context"
	"github.com/DotNetAge/gograph/pkg/cypher/lexer"
)

func (p *Parser) parsePattern() (*ast.PatternExpr, error) {
	start := p.currentPos()
	p.ctx.SetState(context.StateInPattern)

	var parts []*ast.PatternPart

	for {
		part, err := p.parsePatternPart()
		if err != nil {
			return nil, err
		}
		parts = append(parts, part)

		if !p.match(lexer.TokenComma) {
			break
		}
	}

	return &ast.PatternExpr{
		Parts:   parts,
		Start:   start,
		EndPos:  p.currentPos(),
	}, nil
}

func (p *Parser) parsePatternPart() (*ast.PatternPart, error) {
	start := p.currentPos()
	part := &ast.PatternPart{Start: start}

	if p.checkIdentifier() && p.checkNext(lexer.TokenEq) {
		name := p.advance().Value
		p.advance()
		part.Variable = name
		p.ctx.BindVariable(name, context.VarPath, start.Offset)
	}

	path, err := p.parsePath()
	if err != nil {
		return nil, err
	}
	part.Path = path
	part.EndPos = p.currentPos()

	return part, nil
}

func (p *Parser) parsePath() (*ast.PathExpr, error) {
	start := p.currentPos()
	path := &ast.PathExpr{Start: start}

	node, err := p.parseNodePattern()
	if err != nil {
		return nil, err
	}
	path.Nodes = append(path.Nodes, node)

	for {
		rel, err := p.parseRelationPattern()
		if err != nil {
			break
		}
		if rel == nil {
			break
		}

		node, err := p.parseNodePattern()
		if err != nil {
			return nil, err
		}

		path.Relationships = append(path.Relationships, rel)
		path.Nodes = append(path.Nodes, node)
	}

	path.EndPos = p.currentPos()
	return path, nil
}

func (p *Parser) parseNodePattern() (*ast.NodePattern, error) {
	start := p.currentPos()
	node := &ast.NodePattern{Start: start}

	if err := p.expect(lexer.TokenLParen); err != nil {
		return nil, err
	}

	if p.checkIdentifier() {
		name := p.advance().Value
		node.Variable = name
		p.ctx.BindVariable(name, context.VarNode, start.Offset)
	}

	for p.match(lexer.TokenColon) {
		label := p.advance().Value
		node.Labels = append(node.Labels, label)
	}

	if p.match(lexer.TokenLBrace) {
		props, err := p.parsePropertyMap()
		if err != nil {
			return nil, err
		}
		node.Properties = props
	} else if p.checkIdentifier() && len(p.peek().Value) > 0 && p.peek().Value[0] == '$' {
		param := p.advance()
		node.PropertyExpr = &ast.ParamExpr{
			Name: param.Value[1:],
			Start: ast.Pos{
				Line:   param.Line,
				Column: param.Column,
				Offset: param.Position,
			},
			EndPos: ast.Pos{
				Line:   param.Line,
				Column: param.Column + len(param.Value),
				Offset: param.Position + len(param.Value),
			},
		}
	}

	if err := p.expect(lexer.TokenRParen); err != nil {
		return nil, err
	}

	node.EndPos = p.currentPos()
	return node, nil
}

func (p *Parser) parseRelationPattern() (*ast.RelationPattern, error) {
	start := p.currentPos()
	rel := &ast.RelationPattern{Start: start}

	if p.match(lexer.TokenArrowRight) {
		rel.Direction = ast.DirectionOutgoing
		rel.RightArrow = true
		rel.EndPos = p.currentPos()
		return rel, nil
	}

	if p.match(lexer.TokenArrowLeft) {
		rel.Direction = ast.DirectionIncoming
		rel.LeftArrow = true

		if p.match(lexer.TokenLBracket) {
			if p.checkIdentifier() {
				name := p.advance().Value
				rel.Variable = name
				p.ctx.BindVariable(name, context.VarRelationship, start.Offset)
			}

			if p.match(lexer.TokenColon) {
				relType := p.advance().Value
				rel.Types = []string{relType}
			}

			if p.match(lexer.TokenLBrace) {
				props, err := p.parsePropertyMap()
				if err != nil {
					return nil, err
				}
				rel.Properties = props
			}

			if err := p.expect(lexer.TokenRBracket); err != nil {
				return nil, err
			}

			p.match(lexer.TokenDash)
		}

		rel.EndPos = p.currentPos()
		return rel, nil
	}

	if p.match(lexer.TokenDash) {
		rel.Direction = ast.DirectionBoth
	} else {
		return nil, nil
	}

	if p.match(lexer.TokenLBracket) {
		if p.checkIdentifier() {
			name := p.advance().Value
			rel.Variable = name
			p.ctx.BindVariable(name, context.VarRelationship, start.Offset)
		}

		if p.match(lexer.TokenColon) {
			relType := p.advance().Value
			rel.Types = []string{relType}
		}

		if p.match(lexer.TokenStar) {
			rel.VariableLength = true
			minVal := 1
			maxVal := -1
			rel.MinHops = &minVal
			rel.MaxHops = &maxVal

			if p.check(lexer.TokenInteger) {
				min, err := p.parseInteger()
				if err != nil {
					return nil, err
				}
				minVal := int(min)
				rel.MinHops = &minVal
			}

			if p.match(lexer.TokenRange) {
				if p.check(lexer.TokenInteger) {
					max, err := p.parseInteger()
					if err != nil {
						return nil, err
					}
					maxVal := int(max)
					rel.MaxHops = &maxVal
				}
			} else {
				rel.MaxHops = rel.MinHops
			}
		}

		if p.match(lexer.TokenLBrace) {
			props, err := p.parsePropertyMap()
			if err != nil {
				return nil, err
			}
			rel.Properties = props
		}

		if err := p.expect(lexer.TokenRBracket); err != nil {
			return nil, err
		}
	}

	if p.match(lexer.TokenArrowRight) {
		rel.Direction = ast.DirectionOutgoing
		rel.RightArrow = true
	} else if p.match(lexer.TokenDash) {
		if rel.Direction == ast.DirectionIncoming {
			rel.Direction = ast.DirectionIncoming
		} else {
			rel.Direction = ast.DirectionBoth
		}
	}

	rel.EndPos = p.currentPos()
	return rel, nil
}

func (p *Parser) parsePropertyMap() (map[string]ast.Expr, error) {
	props := make(map[string]ast.Expr)

	for !p.check(lexer.TokenRBrace) && !p.atEnd() {
		key := p.advance().Value

		if err := p.expect(lexer.TokenColon); err != nil {
			return nil, err
		}

		value, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		props[key] = value

		if !p.match(lexer.TokenComma) {
			break
		}
	}

	if err := p.expect(lexer.TokenRBrace); err != nil {
		return nil, err
	}

	return props, nil
}
