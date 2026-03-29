package parser

import (
	"fmt"

	"github.com/DotNetAge/gograph/pkg/cypher/ast"
	"github.com/DotNetAge/gograph/pkg/cypher/lexer"
)

func (p *Parser) currentPos() ast.Pos {
	if p.tokens == nil || p.pos < 0 || p.pos >= len(p.tokens) {
		return ast.Pos{}
	}
	tok := p.tokens[p.pos]
	return ast.Pos{
		Line:   tok.Line,
		Column: tok.Column,
		Offset: tok.Position,
	}
}

func (p *Parser) atEnd() bool {
	return p.tokens == nil || p.pos < 0 || p.pos >= len(p.tokens) || p.tokens[p.pos].Type == lexer.TokenEOF
}

func (p *Parser) peek() lexer.Token {
	if p.atEnd() {
		return lexer.Token{Type: lexer.TokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) peekNext() lexer.Token {
	if p.tokens == nil || p.pos+1 < 0 || p.pos+1 >= len(p.tokens) {
		return lexer.Token{Type: lexer.TokenEOF}
	}
	return p.tokens[p.pos+1]
}

func (p *Parser) advance() lexer.Token {
	if p.atEnd() {
		return lexer.Token{Type: lexer.TokenEOF}
	}
	tok := p.tokens[p.pos]
	p.pos++
	return tok
}

func (p *Parser) check(typ lexer.TokenType) bool {
	return !p.atEnd() && p.peek().Type == typ
}

func (p *Parser) checkKeyword(name string) bool {
	return !p.atEnd() && p.peek().Value == name
}

func (p *Parser) checkIdentifier() bool {
	return !p.atEnd() && p.peek().Type == lexer.TokenIdentifier
}

func (p *Parser) checkNext(typ lexer.TokenType) bool {
	return !p.atEnd() && p.peekNext().Type == typ
}

func (p *Parser) checkNextKeyword(name string) bool {
	next := p.peekNext()
	return !p.atEnd() && next.Type == lexer.TokenIdentifier && next.Value == name
}

func (p *Parser) match(typ lexer.TokenType) bool {
	if p.check(typ) {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) matchKeyword(name string) bool {
	if p.checkKeyword(name) {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) expect(typ lexer.TokenType) error {
	if !p.check(typ) {
		return p.errorf("expected %v, got %s", typ, p.peek().Value)
	}
	p.advance()
	return nil
}

func (p *Parser) expectKeyword(name string) error {
	if !p.checkKeyword(name) {
		return p.errorf("expected %s, got %s", name, p.peek().Value)
	}
	p.advance()
	return nil
}

func (p *Parser) errorf(format string, args ...interface{}) error {
	tok := p.peek()
	return fmt.Errorf("parse error at line %d, column %d: %s",
		tok.Line, tok.Column, fmt.Sprintf(format, args...))
}

func (p *Parser) synchronize() {
	for !p.atEnd() {
		if p.peek().Type == lexer.TokenSemicolon {
			return
		}
		switch p.peek().Value {
		case "MATCH", "CREATE", "MERGE", "SET", "DELETE", "REMOVE", "WITH", "RETURN", "UNWIND":
			return
		}
		p.advance()
	}
}
