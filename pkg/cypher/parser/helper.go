package parser

import (
	"fmt"

	"github.com/DotNetAge/gograph/pkg/cypher/ast"
	"github.com/DotNetAge/gograph/pkg/cypher/lexer"
)

// currentPos returns the current position in the source as an ast.Pos.
// It converts the token position information to AST position format.
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

// atEnd returns true if the parser has reached the end of the token stream.
func (p *Parser) atEnd() bool {
	return p.tokens == nil || p.pos < 0 || p.pos >= len(p.tokens) || p.tokens[p.pos].Type == lexer.TokenEOF
}

// peek returns the current token without consuming it.
// Returns EOF token if at the end of the stream.
func (p *Parser) peek() lexer.Token {
	if p.atEnd() {
		return lexer.Token{Type: lexer.TokenEOF}
	}
	return p.tokens[p.pos]
}

// peekNext returns the next token without consuming any tokens.
// Returns EOF token if there is no next token.
func (p *Parser) peekNext() lexer.Token {
	if p.tokens == nil || p.pos+1 < 0 || p.pos+1 >= len(p.tokens) {
		return lexer.Token{Type: lexer.TokenEOF}
	}
	return p.tokens[p.pos+1]
}

// advance consumes and returns the current token, moving to the next one.
// Returns EOF token if at the end of the stream.
func (p *Parser) advance() lexer.Token {
	if p.atEnd() {
		return lexer.Token{Type: lexer.TokenEOF}
	}
	tok := p.tokens[p.pos]
	p.pos++
	return tok
}

// check returns true if the current token has the given type.
func (p *Parser) check(typ lexer.TokenType) bool {
	return !p.atEnd() && p.peek().Type == typ
}

// checkKeyword returns true if the current token is an identifier with the given name.
func (p *Parser) checkKeyword(name string) bool {
	return !p.atEnd() && p.peek().Value == name
}

// checkIdentifier returns true if the current token is an identifier.
func (p *Parser) checkIdentifier() bool {
	return !p.atEnd() && p.peek().Type == lexer.TokenIdentifier
}

// checkNext returns true if the next token has the given type.
func (p *Parser) checkNext(typ lexer.TokenType) bool {
	return !p.atEnd() && p.peekNext().Type == typ
}

// checkNextKeyword returns true if the next token is an identifier with the given name.
func (p *Parser) checkNextKeyword(name string) bool {
	next := p.peekNext()
	return !p.atEnd() && next.Type == lexer.TokenIdentifier && next.Value == name
}

// match consumes the current token if it has the given type and returns true.
// Returns false if the token doesn't match.
func (p *Parser) match(typ lexer.TokenType) bool {
	if p.check(typ) {
		p.advance()
		return true
	}
	return false
}

// matchKeyword consumes the current token if it's an identifier with the given name.
// Returns false if the token doesn't match.
func (p *Parser) matchKeyword(name string) bool {
	if p.checkKeyword(name) {
		p.advance()
		return true
	}
	return false
}

// expect consumes the current token if it has the given type, otherwise returns an error.
func (p *Parser) expect(typ lexer.TokenType) error {
	if !p.check(typ) {
		return p.errorf("expected %v, got %s", typ, p.peek().Value)
	}
	p.advance()
	return nil
}

// expectKeyword consumes the current token if it's an identifier with the given name.
// Returns an error if the token doesn't match.
func (p *Parser) expectKeyword(name string) error {
	if !p.checkKeyword(name) {
		return p.errorf("expected %s, got %s", name, p.peek().Value)
	}
	p.advance()
	return nil
}

// errorf creates a formatted parse error with position information.
func (p *Parser) errorf(format string, args ...interface{}) error {
	tok := p.peek()
	return fmt.Errorf("parse error at line %d, column %d: %s",
		tok.Line, tok.Column, fmt.Sprintf(format, args...))
}

// synchronize attempts to recover from a parse error by skipping tokens
// until a synchronization point is found (statement boundary or semicolon).
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
