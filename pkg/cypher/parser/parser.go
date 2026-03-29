package parser

import (
	"fmt"
	"strconv"

	"github.com/DotNetAge/gograph/pkg/cypher/ast"
	"github.com/DotNetAge/gograph/pkg/cypher/context"
	"github.com/DotNetAge/gograph/pkg/cypher/lexer"
)

const (
	precOr         = 1
	precXor        = 2
	precAnd        = 3
	precNot        = 4
	precEquality   = 5
	precComparison = 6
	precAddSub     = 7
	precMulDiv     = 8
	precUnary      = 9
	precPower      = 10
	precPostfix    = 11
)

var precedences = map[lexer.TokenType]int{
	lexer.TokenOr:      precOr,
	lexer.TokenXor:     precXor,
	lexer.TokenAnd:     precAnd,
	lexer.TokenEq:      precEquality,
	lexer.TokenNeq:     precEquality,
	lexer.TokenLt:      precComparison,
	lexer.TokenLe:      precComparison,
	lexer.TokenGt:      precComparison,
	lexer.TokenGe:      precComparison,
	lexer.TokenPlus:    precAddSub,
	lexer.TokenDash:    precAddSub,
	lexer.TokenStar:    precMulDiv,
	lexer.TokenSlash:   precMulDiv,
	lexer.TokenPercent: precMulDiv,
	lexer.TokenCaret:   precPower,
}

type Parser struct {
	tokens []lexer.Token
	pos    int
	ctx    *context.Context
	errors []error
}

func New(input string) *Parser {
	l := lexer.New(input)
	tokens, err := l.Tokenize()
	if err != nil {
		return &Parser{
			tokens: tokens,
			ctx:    context.New(),
			errors: []error{err},
		}
	}
	return &Parser{
		tokens: tokens,
		ctx:    context.New(),
	}
}

func NewFromTokens(tokens []lexer.Token) *Parser {
	return &Parser{
		tokens: tokens,
		ctx:    context.New(),
	}
}

func (p *Parser) Parse() (*ast.Query, error) {
	if len(p.errors) > 0 {
		return nil, fmt.Errorf("lexer errors: %v", p.errors)
	}

	query := &ast.Query{
		Start: p.currentPos(),
	}

	for !p.atEnd() {
		if p.check(lexer.TokenSemicolon) {
			p.advance()
			continue
		}

		stmt, err := p.parseStatement()
		if err != nil {
			p.errors = append(p.errors, err)
			p.synchronize()
			continue
		}
		if stmt != nil {
			query.Statements = append(query.Statements, stmt)
		}
		if p.check(lexer.TokenSemicolon) {
			p.advance()
		}
	}

	query.EndPos = p.currentPos()

	if len(p.errors) > 0 {
		return query, fmt.Errorf("parse errors: %v", p.errors)
	}
	return query, nil
}

func (p *Parser) parseStatement() (ast.Stmt, error) {
	p.ctx.EnterScope()
	defer p.ctx.ExitScope()

	if p.matchKeyword("OPTIONAL") {
		if !p.matchKeyword("MATCH") {
			return nil, p.errorf("expected MATCH after OPTIONAL")
		}
		return p.parseMatchStmtWithOptional(true)
	}
	if p.matchKeyword("MATCH") {
		return p.parseMatchStmtWithOptional(false)
	}
	if p.matchKeyword("CREATE") {
		return p.parseCreateStmt()
	}
	if p.matchKeyword("MERGE") {
		return p.parseMergeStmt()
	}
	if p.matchKeyword("SET") {
		return p.parseSetStmt()
	}
	if p.matchKeyword("DELETE") || p.matchKeyword("DETACH") {
		return p.parseDeleteStmt()
	}
	if p.matchKeyword("REMOVE") {
		return p.parseRemoveStmt()
	}
	if p.matchKeyword("WITH") {
		return p.parseWithStmt()
	}
	if p.matchKeyword("UNWIND") {
		return p.parseUnwindStmt()
	}
	if p.matchKeyword("RETURN") {
		return p.parseReturnStmt()
	}

	return nil, p.errorf("expected statement, got %s", p.peek().Value)
}

func (p *Parser) parseMatchStmtWithOptional(optional bool) (*ast.MatchStmt, error) {
	start := p.currentPos()
	p.ctx.SetState(context.StateInMatch)

	stmt := &ast.MatchStmt{
		Start: start,
	}

	clause, err := p.parseMatchClauseWithOptional(optional)
	if err != nil {
		return nil, err
	}
	stmt.Clauses = append(stmt.Clauses, clause)

	for !p.atEnd() {
		if p.matchKeyword("WHERE") {
			where, err := p.parseWhereClause()
			if err != nil {
				return nil, err
			}
			stmt.Clauses = append(stmt.Clauses, where)
		} else if p.matchKeyword("SET") {
			setClause, err := p.parseSetClause()
			if err != nil {
				return nil, err
			}
			stmt.Clauses = append(stmt.Clauses, setClause)
		} else if p.matchKeyword("RETURN") {
			ret, err := p.parseReturnClause()
			if err != nil {
				return nil, err
			}
			stmt.Clauses = append(stmt.Clauses, ret)
		} else if p.checkKeyword("DELETE") {
			p.advance()
			del, err := p.parseDeleteClause(false)
			if err != nil {
				return nil, err
			}
			stmt.Clauses = append(stmt.Clauses, del)
		} else if p.checkKeyword("DETACH") {
			p.advance()
			if !p.matchKeyword("DELETE") {
				return nil, p.errorf("expected DELETE after DETACH")
			}
			del, err := p.parseDeleteClause(true)
			if err != nil {
				return nil, err
			}
			stmt.Clauses = append(stmt.Clauses, del)
		} else {
			break
		}
	}

	stmt.EndPos = p.currentPos()
	return stmt, nil
}

func (p *Parser) parseMatchClauseWithOptional(optional bool) (*ast.MatchClause, error) {
	start := p.currentPos()
	clause := &ast.MatchClause{Start: start, Optional: optional}

	pattern, err := p.parsePattern()
	if err != nil {
		return nil, err
	}
	clause.Pattern = pattern
	clause.EndPos = p.currentPos()

	return clause, nil
}

func (p *Parser) parseMatchClause() (*ast.MatchClause, error) {
	return p.parseMatchClauseWithOptional(false)
}

func (p *Parser) parseCreateStmt() (*ast.CreateStmt, error) {
	start := p.currentPos()
	p.ctx.SetState(context.StateInCreate)

	clause, err := p.parseCreateClause()
	if err != nil {
		return nil, err
	}

	return &ast.CreateStmt{
		Pattern: clause.Pattern,
		Start:   start,
		EndPos:  p.currentPos(),
	}, nil
}

func (p *Parser) parseCreateClause() (*ast.CreateClause, error) {
	start := p.currentPos()
	pattern, err := p.parsePattern()
	if err != nil {
		return nil, err
	}
	return &ast.CreateClause{
		Pattern: pattern,
		Start:   start,
		EndPos:  p.currentPos(),
	}, nil
}

func (p *Parser) parseMergeStmt() (*ast.MergeStmt, error) {
	start := p.currentPos()
	p.ctx.SetState(context.StateInMerge)

	clause, err := p.parseMergeClause()
	if err != nil {
		return nil, err
	}

	return &ast.MergeStmt{
		Pattern:  clause.Pattern,
		OnCreate: clause.OnCreate,
		OnMatch:  clause.OnMatch,
		Start:    start,
		EndPos:   p.currentPos(),
	}, nil
}

func (p *Parser) parseMergeClause() (*ast.MergeClause, error) {
	start := p.currentPos()
	pattern, err := p.parsePattern()
	if err != nil {
		return nil, err
	}

	clause := &ast.MergeClause{
		Pattern: pattern,
		Start:   start,
	}

	for !p.atEnd() {
		if p.matchKeyword("ON") {
			if p.matchKeyword("CREATE") {
				if !p.matchKeyword("SET") {
					return nil, p.errorf("expected SET after ON CREATE")
				}
				items, err := p.parseSetItems()
				if err != nil {
					return nil, err
				}
				clause.OnCreate = items
			} else if p.matchKeyword("MATCH") {
				if !p.matchKeyword("SET") {
					return nil, p.errorf("expected SET after ON MATCH")
				}
				items, err := p.parseSetItems()
				if err != nil {
					return nil, err
				}
				clause.OnMatch = items
			}
		} else {
			break
		}
	}

	clause.EndPos = p.currentPos()
	return clause, nil
}

func (p *Parser) parseSetStmt() (*ast.SetStmt, error) {
	start := p.currentPos()
	p.ctx.SetState(context.StateInSet)

	clause, err := p.parseSetClause()
	if err != nil {
		return nil, err
	}

	return &ast.SetStmt{
		Items:  clause.Items,
		Start:  start,
		EndPos: p.currentPos(),
	}, nil
}

func (p *Parser) parseSetClause() (*ast.SetClause, error) {
	start := p.currentPos()
	items, err := p.parseSetItems()
	if err != nil {
		return nil, err
	}
	return &ast.SetClause{
		Items:  items,
		Start:  start,
		EndPos: p.currentPos(),
	}, nil
}

func (p *Parser) parseSetItems() ([]*ast.SetItem, error) {
	var items []*ast.SetItem

	for {
		item, err := p.parseSetItem()
		if err != nil {
			return nil, err
		}
		items = append(items, &item)

		if !p.match(lexer.TokenComma) {
			break
		}
	}

	return items, nil
}

func (p *Parser) parseSetItem() (ast.SetItem, error) {
	start := p.currentPos()
	item := ast.SetItem{Start: start}

	if p.checkIdentifier() && p.checkNext(lexer.TokenColon) {
		name := p.advance().Value
		p.advance()
		item.Target = &ast.Ident{Name: name, Start: start}
		item.IsLabel = true

		var labels []string
		for {
			if !p.checkIdentifier() {
				break
			}
			label := p.advance().Value
			labels = append(labels, label)
			if !p.match(lexer.TokenColon) {
				break
			}
		}
		if len(labels) == 1 {
			item.Value = &ast.StringLit{Value: labels[0], Start: p.currentPos()}
		} else if len(labels) > 1 {
			item.Value = &ast.ListExpr{
				Elements: func() []ast.Expr {
					var elems []ast.Expr
					for _, l := range labels {
						elems = append(elems, &ast.StringLit{Value: l})
					}
					return elems
				}(),
				Start:  p.currentPos(),
				EndPos: p.currentPos(),
			}
		}
		item.EndPos = p.currentPos()
		return item, nil
	}

	target, err := p.parsePropertyAccessExpr()
	if err != nil {
		return item, err
	}
	item.Target = target

	if p.match(lexer.TokenPlusEq) {
		item.Operator = "+="
	} else if p.match(lexer.TokenEq) {
		item.Operator = "="
	} else {
		return item, p.errorf("expected '=' or '+=' in SET clause")
	}

	value, err := p.parseExpr()
	if err != nil {
		return item, err
	}
	item.Value = value
	item.EndPos = p.currentPos()

	return item, nil
}

func (p *Parser) parsePropertyAccessExpr() (ast.Expr, error) {
	start := p.currentPos()

	if !p.checkIdentifier() {
		return nil, p.errorf("expected identifier")
	}

	name := p.advance().Value
	var expr ast.Expr = &ast.Ident{Name: name, Start: start, EndPos: p.currentPos()}

	for p.match(lexer.TokenDot) {
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
	}

	return expr, nil
}

func (p *Parser) parseDeleteStmt() (*ast.DeleteStmt, error) {
	start := p.currentPos()
	p.ctx.SetState(context.StateInDelete)

	detach := p.matchKeyword("DETACH")
	if !p.matchKeyword("DELETE") {
		return nil, p.errorf("expected DELETE")
	}
	clause, err := p.parseDeleteClause(detach)
	if err != nil {
		return nil, err
	}

	return &ast.DeleteStmt{
		Detach: clause.Detach,
		Items:  clause.Items,
		Start:  start,
		EndPos: p.currentPos(),
	}, nil
}

func (p *Parser) parseDeleteClause(detach bool) (*ast.DeleteClause, error) {
	start := p.currentPos()
	clause := &ast.DeleteClause{Start: start, Detach: detach}

	var items []ast.Expr
	for {
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		items = append(items, expr)

		if !p.match(lexer.TokenComma) {
			break
		}
	}

	clause.Items = items
	clause.EndPos = p.currentPos()
	return clause, nil
}

func (p *Parser) parseRemoveStmt() (*ast.RemoveStmt, error) {
	start := p.currentPos()
	p.ctx.SetState(context.StateInRemove)

	clause, err := p.parseRemoveClause()
	if err != nil {
		return nil, err
	}

	return &ast.RemoveStmt{
		Items:  clause.Items,
		Start:  start,
		EndPos: p.currentPos(),
	}, nil
}

func (p *Parser) parseRemoveClause() (*ast.RemoveClause, error) {
	start := p.currentPos()
	var items []*ast.RemoveItemExpr

	for {
		item, err := p.parseRemoveItem()
		if err != nil {
			return nil, err
		}
		items = append(items, &item)

		if !p.match(lexer.TokenComma) {
			break
		}
	}

	return &ast.RemoveClause{
		Items:  items,
		Start:  start,
		EndPos: p.currentPos(),
	}, nil
}

func (p *Parser) parseRemoveItem() (ast.RemoveItemExpr, error) {
	start := p.currentPos()
	item := ast.RemoveItemExpr{Start: start}

	if p.checkIdentifier() && p.checkNext(lexer.TokenColon) {
		name := p.advance().Value
		p.advance()
		_ = p.advance().Value
		item.Target = &ast.Ident{Name: name, Start: start}
		item.IsLabel = true
		item.EndPos = p.currentPos()
		return item, nil
	}

	target, err := p.parseExpr()
	if err != nil {
		return item, err
	}
	item.Target = target
	item.EndPos = p.currentPos()

	return item, nil
}

func (p *Parser) parseWithStmt() (*ast.WithStmt, error) {
	start := p.currentPos()
	p.ctx.SetState(context.StateInWith)

	clause, err := p.parseWithClause()
	if err != nil {
		return nil, err
	}

	return &ast.WithStmt{
		Distinct: clause.Distinct,
		Items:    clause.Items,
		OrderBy:  clause.OrderBy,
		Skip:     clause.Skip,
		Limit:    clause.Limit,
		Where:    clause.Where,
		Start:    start,
		EndPos:   p.currentPos(),
	}, nil
}

func (p *Parser) parseWithClause() (*ast.WithClause, error) {
	start := p.currentPos()
	clause := &ast.WithClause{Start: start}

	if p.matchKeyword("DISTINCT") {
		clause.Distinct = true
	}

	items, err := p.parseReturnItems()
	if err != nil {
		return nil, err
	}
	clause.Items = items

	if p.matchKeyword("ORDER") {
		if !p.matchKeyword("BY") {
			return nil, p.errorf("expected BY after ORDER")
		}
		orderBy, err := p.parseOrderByItems()
		if err != nil {
			return nil, err
		}
		clause.OrderBy = orderBy
	}

	if p.matchKeyword("SKIP") {
		skip, err := p.parseInteger()
		if err != nil {
			return nil, err
		}
		clause.Skip = &ast.IntegerLit{Value: skip}
	}

	if p.matchKeyword("LIMIT") {
		limit, err := p.parseInteger()
		if err != nil {
			return nil, err
		}
		clause.Limit = &ast.IntegerLit{Value: limit}
	}

	if p.matchKeyword("WHERE") {
		where, err := p.parseWhereClause()
		if err != nil {
			return nil, err
		}
		clause.Where = where
	}

	clause.EndPos = p.currentPos()
	return clause, nil
}

func (p *Parser) parseUnwindStmt() (*ast.UnwindStmt, error) {
	start := p.currentPos()

	list, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	if !p.matchKeyword("AS") {
		return nil, p.errorf("expected AS in UNWIND clause")
	}

	variable := p.advance().Value
	p.ctx.BindVariable(variable, context.VarScalar, p.currentPos().Offset)

	return &ast.UnwindStmt{
		List:     list,
		Variable: variable,
		Start:    start,
		EndPos:   p.currentPos(),
	}, nil
}

func (p *Parser) parseReturnStmt() (*ast.ReturnStmt, error) {
	start := p.currentPos()
	p.ctx.SetState(context.StateInReturn)

	clause, err := p.parseReturnClause()
	if err != nil {
		return nil, err
	}

	return &ast.ReturnStmt{
		Distinct: clause.Distinct,
		Items:    clause.Items,
		OrderBy:  clause.OrderBy,
		Skip:     clause.Skip,
		Limit:    clause.Limit,
		Start:    start,
		EndPos:   p.currentPos(),
	}, nil
}

func (p *Parser) parseReturnClause() (*ast.ReturnExpr, error) {
	start := p.currentPos()
	clause := &ast.ReturnExpr{Start: start}

	if p.matchKeyword("DISTINCT") {
		clause.Distinct = true
	}

	items, err := p.parseReturnItems()
	if err != nil {
		return nil, err
	}
	clause.Items = items

	if p.matchKeyword("ORDER") {
		if !p.matchKeyword("BY") {
			return nil, p.errorf("expected BY after ORDER")
		}
		orderBy, err := p.parseOrderByItems()
		if err != nil {
			return nil, err
		}
		clause.OrderBy = orderBy
	}

	if p.matchKeyword("SKIP") {
		skip, err := p.parseInteger()
		if err != nil {
			return nil, err
		}
		clause.Skip = &ast.IntegerLit{Value: skip}
	}

	if p.matchKeyword("LIMIT") {
		limit, err := p.parseInteger()
		if err != nil {
			return nil, err
		}
		clause.Limit = &ast.IntegerLit{Value: limit}
	}

	clause.EndPos = p.currentPos()
	return clause, nil
}

func (p *Parser) parseReturnItems() ([]*ast.ReturnItemExpr, error) {
	var items []*ast.ReturnItemExpr

	for {
		item, err := p.parseReturnItem()
		if err != nil {
			return nil, err
		}
		items = append(items, &item)

		if !p.match(lexer.TokenComma) {
			break
		}
	}

	return items, nil
}

func (p *Parser) parseReturnItem() (ast.ReturnItemExpr, error) {
	start := p.currentPos()
	item := ast.ReturnItemExpr{Start: start}

	expr, err := p.parseExpr()
	if err != nil {
		return item, err
	}
	item.Expr = expr

	if p.matchKeyword("AS") {
		item.Alias = p.advance().Value
	}

	item.EndPos = p.currentPos()
	return item, nil
}

func (p *Parser) parseOrderByItems() (*ast.OrderByExpr, error) {
	start := p.currentPos()
	var items []*ast.OrderByItem

	for {
		item, err := p.parseOrderByItem()
		if err != nil {
			return nil, err
		}
		items = append(items, &item)

		if !p.match(lexer.TokenComma) {
			break
		}
	}

	return &ast.OrderByExpr{Items: items, Start: start, EndPos: p.currentPos()}, nil
}

func (p *Parser) parseOrderByItem() (ast.OrderByItem, error) {
	start := p.currentPos()
	item := ast.OrderByItem{Start: start}

	expr, err := p.parseExpr()
	if err != nil {
		return item, err
	}
	item.Expr = expr

	if p.matchKeyword("DESC") || p.matchKeyword("DESCENDING") {
		item.Descending = true
	} else {
		_ = p.matchKeyword("ASC") || p.matchKeyword("ASCENDING")
	}

	item.EndPos = p.currentPos()
	return item, nil
}

func (p *Parser) parseWhereClause() (*ast.WhereExpr, error) {
	start := p.currentPos()
	p.ctx.SetState(context.StateInWhere)

	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	return &ast.WhereExpr{
		Expr:   expr,
		Start:  start,
		EndPos: p.currentPos(),
	}, nil
}

func (p *Parser) parseInteger() (int64, error) {
	tok := p.peek()
	if tok.Type != lexer.TokenInteger {
		return 0, p.errorf("expected integer, got %s", tok.Value)
	}
	p.advance()

	val, err := strconv.ParseInt(tok.Value, 10, 64)
	if err != nil {
		if numErr, ok := err.(*strconv.NumError); ok && numErr.Err == strconv.ErrRange {
			bigVal, bigErr := strconv.ParseUint(tok.Value, 10, 64)
			if bigErr == nil && bigVal == 9223372036854775808 {
				return -9223372036854775808, nil
			}
		}
		return 0, p.errorf("integer value out of range: %s", tok.Value)
	}
	return val, nil
}
