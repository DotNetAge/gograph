package parser

import (
	"github.com/DotNetAge/gograph/pkg/cypher/ast"
	"github.com/DotNetAge/gograph/pkg/cypher/lexer"
)

func (p *Parser) parseExpr() (ast.Expr, error) {
	return p.parseBinaryExpr(0)
}

func (p *Parser) parseBinaryExpr(minPrec int) (ast.Expr, error) {
	left, err := p.parseUnaryExpr()
	if err != nil {
		return nil, err
	}

	for {
		tok := p.peek()

		if tok.Type == lexer.TokenIdentifier {
			switch tok.Value {
			case "IS":
				if minPrec > precEquality {
					break
				}
				p.advance()
				isNot := p.matchKeyword("NOT")
				if err := p.expectKeyword("NULL"); err != nil {
					return nil, err
				}
				left = &ast.IsNullExpr{
					Expr:   left,
					IsNot:  isNot,
					Start:  left.Position(),
					EndPos: p.currentPos(),
				}
				continue
			case "IN":
				if minPrec > precComparison {
					break
				}
				p.advance()
				right, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				left = &ast.InExpr{
					Left:   left,
					Right:  right,
					Start:  left.Position(),
					EndPos: right.End(),
				}
				continue
			case "AND", "OR", "XOR":
				prec := precAnd
				if tok.Value == "OR" {
					prec = precOr
				} else if tok.Value == "XOR" {
					prec = precXor
				}
				if prec < minPrec {
					break
				}
				p.advance()
				right, err := p.parseBinaryExpr(prec + 1)
				if err != nil {
					return nil, err
				}
				left = &ast.BinaryExpr{
					Left:     left,
					Operator: tok.Value,
					Right:    right,
					Start:    left.Position(),
					EndPos:   right.End(),
				}
				continue
			case "CONTAINS", "STARTS", "ENDS":
				if minPrec > precComparison {
					break
				}
				p.advance()
				var op string
				switch tok.Value {
				case "CONTAINS":
					op = "CONTAINS"
				case "STARTS":
					if !p.matchKeyword("WITH") {
						return nil, p.errorf("expected WITH after STARTS")
					}
					op = "STARTS WITH"
				case "ENDS":
					if !p.matchKeyword("WITH") {
						return nil, p.errorf("expected WITH after ENDS")
					}
					op = "ENDS WITH"
				}
				right, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				left = &ast.BinaryExpr{
					Left:     left,
					Operator: op,
					Right:    right,
					Start:    left.Position(),
					EndPos:   right.End(),
				}
				continue
			}
		}

		prec, ok := precedences[tok.Type]
		if !ok || prec < minPrec {
			break
		}

		op := p.advance()
		right, err := p.parseBinaryExpr(prec + 1)
		if err != nil {
			return nil, err
		}

		left = &ast.BinaryExpr{
			Left:     left,
			Operator: op.Value,
			Right:    right,
			Start:    left.Position(),
			EndPos:   right.End(),
		}
	}

	return left, nil
}

func (p *Parser) parseUnaryExpr() (ast.Expr, error) {
	start := p.currentPos()

	if p.matchKeyword("NOT") {
		operand, err := p.parseUnaryExpr()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpr{
			Operator: "NOT",
			Operand:  operand,
			Start:    start,
			EndPos:   operand.End(),
		}, nil
	}

	if p.match(lexer.TokenDash) {
		operand, err := p.parseUnaryExpr()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpr{
			Operator: "-",
			Operand:  operand,
			Start:    start,
			EndPos:   operand.End(),
		}, nil
	}

	if p.match(lexer.TokenPlus) {
		return p.parseUnaryExpr()
	}

	return p.parsePostfixExpr()
}

func (p *Parser) parsePostfixExpr() (ast.Expr, error) {
	expr, err := p.parsePrimaryExpr()
	if err != nil {
		return nil, err
	}

	for {
		if p.match(lexer.TokenDot) {
			if !p.checkIdentifier() {
				return nil, p.errorf("expected property name after '.'")
			}
			prop := p.advance().Value
			expr = &ast.PropertyAccessExpr{
				Target:   expr,
				Property: prop,
				Start:    expr.Position(),
				EndPos:   p.currentPos(),
			}
			continue
		}

		if p.match(lexer.TokenLBracket) {
			if p.match(lexer.TokenRange) {
				var to ast.Expr
				if !p.check(lexer.TokenRBracket) {
					idx, err := p.parseExpr()
					if err != nil {
						return nil, err
					}
					to = idx
				}
				if err := p.expect(lexer.TokenRBracket); err != nil {
					return nil, err
				}
				expr = &ast.ListSliceExpr{
					List:   expr,
					From:   nil,
					To:     to,
					Start:  expr.Position(),
					EndPos: p.currentPos(),
				}
				continue
			}

			index, err := p.parseExpr()
			if err != nil {
				return nil, err
			}

			if p.match(lexer.TokenRange) {
				var to ast.Expr
				if !p.check(lexer.TokenRBracket) {
					idx, err := p.parseExpr()
					if err != nil {
						return nil, err
					}
					to = idx
				}
				if err := p.expect(lexer.TokenRBracket); err != nil {
					return nil, err
				}
				expr = &ast.ListSliceExpr{
					List:   expr,
					From:   index,
					To:     to,
					Start:  expr.Position(),
					EndPos: p.currentPos(),
				}
				continue
			}

			if err := p.expect(lexer.TokenRBracket); err != nil {
				return nil, err
			}
			expr = &ast.ListIndexExpr{
				List:   expr,
				Index:  index,
				Start:  expr.Position(),
				EndPos: p.currentPos(),
			}
			continue
		}

		break
	}

	return expr, nil
}

func (p *Parser) parsePrimaryExpr() (ast.Expr, error) {
	start := p.currentPos()

	if p.matchKeyword("CASE") {
		return p.parseCaseExpr()
	}

	if p.matchKeyword("EXISTS") {
		if err := p.expect(lexer.TokenLParen); err != nil {
			return nil, err
		}

		if p.check(lexer.TokenLParen) {
			pattern, err := p.parsePattern()
			if err != nil {
				return nil, err
			}
			if err := p.expect(lexer.TokenRParen); err != nil {
				return nil, err
			}
			return &ast.ExistsExpr{
				Pattern: pattern,
				Start:   start,
				EndPos:  p.currentPos(),
			}, nil
		}

		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if err := p.expect(lexer.TokenRParen); err != nil {
			return nil, err
		}
		return &ast.ExistsExpr{
			Expr:    expr,
			Start:   start,
			EndPos:  p.currentPos(),
		}, nil
	}

	if p.check(lexer.TokenInteger) {
		val, err := p.parseInteger()
		if err != nil {
			return nil, err
		}
		return &ast.IntegerLit{
			Value:  val,
			Start:  start,
			EndPos: p.currentPos(),
		}, nil
	}

	if p.check(lexer.TokenFloat) {
		val := p.advance().Value
		return &ast.FloatLit{
			Value:  parseFloat(val),
			Start:  start,
			EndPos: p.currentPos(),
		}, nil
	}

	if p.check(lexer.TokenString) {
		val := p.advance().Value
		return &ast.StringLit{
			Value:  val,
			Start:  start,
			EndPos: p.currentPos(),
		}, nil
	}

	if p.matchKeyword("TRUE") {
		return &ast.BoolLit{
			Value:  true,
			Start:  start,
			EndPos: p.currentPos(),
		}, nil
	}

	if p.matchKeyword("FALSE") {
		return &ast.BoolLit{
			Value:  false,
			Start:  start,
			EndPos: p.currentPos(),
		}, nil
	}

	if p.matchKeyword("NULL") {
		return &ast.NullLit{
			Start:  start,
			EndPos: p.currentPos(),
		}, nil
	}

	if p.match(lexer.TokenDollar) {
		name := p.advance().Value
		return &ast.Param{
			Name:   name,
			Start:  start,
			EndPos: p.currentPos(),
		}, nil
	}

	if p.match(lexer.TokenLBracket) {
		return p.parseListExpr()
	}

	if p.match(lexer.TokenLBrace) {
		return p.parseMapExpr()
	}

	if p.match(lexer.TokenLParen) {
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if err := p.expect(lexer.TokenRParen); err != nil {
			return nil, err
		}
		return expr, nil
	}

	if p.checkIdentifier() {
		name := p.advance().Value

		if p.check(lexer.TokenLParen) {
			p.advance()
			distinct := p.matchKeyword("DISTINCT")
			args, err := p.parseArgList()
			if err != nil {
				return nil, err
			}
			if err := p.expect(lexer.TokenRParen); err != nil {
				return nil, err
			}
			return &ast.FuncCall{
				Name:     name,
				Args:     args,
				Distinct: distinct,
				Start:    start,
				EndPos:   p.currentPos(),
			}, nil
		}

		return &ast.Ident{
			Name:   name,
			Start:  start,
			EndPos: p.currentPos(),
		}, nil
	}

	return nil, p.errorf("expected expression, got %s", p.peek().Value)
}

func (p *Parser) parseCaseExpr() (ast.Expr, error) {
	start := p.currentPos()
	expr := &ast.CaseExpr{Start: start}

	if !p.checkKeyword("WHEN") {
		operand, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		expr.Operand = operand
	}

	for p.matchKeyword("WHEN") {
		when := &ast.WhenClause{}
		cond, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		when.Condition = cond
		when.Start = cond.Position()

		if err := p.expectKeyword("THEN"); err != nil {
			return nil, err
		}

		result, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		when.Result = result
		when.EndPos = result.End()

		expr.Whens = append(expr.Whens, when)
	}

	if p.matchKeyword("ELSE") {
		elseExpr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		expr.Else = elseExpr
	}

	if err := p.expectKeyword("END"); err != nil {
		return nil, err
	}

	expr.EndPos = p.currentPos()
	return expr, nil
}

func (p *Parser) parseListExpr() (ast.Expr, error) {
	start := p.currentPos()

	if p.checkIdentifier() && p.checkNextKeyword("IN") {
		return p.parseListComprehension(start)
	}

	var elements []ast.Expr

	for !p.check(lexer.TokenRBracket) && !p.atEnd() {
		elem, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		elements = append(elements, elem)

		if !p.match(lexer.TokenComma) {
			break
		}
	}

	if err := p.expect(lexer.TokenRBracket); err != nil {
		return nil, err
	}

	return &ast.ListExpr{
		Elements: elements,
		Start:    start,
		EndPos:   p.currentPos(),
	}, nil
}

func (p *Parser) parseListComprehension(start ast.Pos) (*ast.ListComprehension, error) {
	variable := p.advance().Value

	if !p.matchKeyword("IN") {
		return nil, p.errorf("expected IN in list comprehension")
	}

	list, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	var filter ast.Expr
	if p.matchKeyword("WHERE") {
		filter, err = p.parseExpr()
		if err != nil {
			return nil, err
		}
	}

	var projection ast.Expr
	if p.match(lexer.TokenPipe) {
		projection, err = p.parseExpr()
		if err != nil {
			return nil, err
		}
	}

	if err := p.expect(lexer.TokenRBracket); err != nil {
		return nil, err
	}

	return &ast.ListComprehension{
		Variable:   variable,
		List:       list,
		Filter:     filter,
		Projection: projection,
		Start:      start,
		EndPos:     p.currentPos(),
	}, nil
}

func (p *Parser) parseMapExpr() (ast.Expr, error) {
	start := p.currentPos()
	var pairs []*ast.MapPair

	for !p.check(lexer.TokenRBrace) && !p.atEnd() {
		key := p.advance().Value

		if err := p.expect(lexer.TokenColon); err != nil {
			return nil, err
		}

		value, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		pairs = append(pairs, &ast.MapPair{
			Key:    key,
			Value:  value,
			Start:  start,
			EndPos: p.currentPos(),
		})

		if !p.match(lexer.TokenComma) {
			break
		}
	}

	if err := p.expect(lexer.TokenRBrace); err != nil {
		return nil, err
	}

	return &ast.MapExpr{
		Pairs:  pairs,
		Start:  start,
		EndPos: p.currentPos(),
	}, nil
}

func (p *Parser) parseArgList() ([]ast.Expr, error) {
	var args []ast.Expr

	for !p.check(lexer.TokenRParen) && !p.atEnd() {
		arg, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		if !p.match(lexer.TokenComma) {
			break
		}
	}

	return args, nil
}

func parseFloat(s string) float64 {
	var f float64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			f = f*10 + float64(c-'0')
		} else if c == '.' {
			break
		}
	}
	return f
}
