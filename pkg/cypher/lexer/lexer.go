// Package lexer provides lexical analysis for Cypher query strings.
// It tokenizes input strings into a stream of tokens that can be
// consumed by the parser.
//
// The lexer supports:
//   - Identifiers and keywords
//   - Numbers (integers and floats)
//   - Strings (single and double quoted)
//   - Operators (+, -, *, /, %, ^, =, !=, <, <=, >, >=)
//   - Delimiters (parentheses, braces, brackets, commas, etc.)
//   - Comments (single-line // and multi-line /* */)
//
// Basic Usage:
//
//	l := lexer.New("MATCH (n:Person) RETURN n.name")
//	tokens, err := l.Tokenize()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, tok := range tokens {
//	    fmt.Printf("%s: %s\n", tok.Type, tok.Value)
//	}
package lexer

import (
	"fmt"
	"strings"
	"unicode"
)

// Lexer tokenizes Cypher query strings into a stream of tokens.
type Lexer struct {
	input   string      // Input string to tokenize
	pos     int         // Current position in input
	line    int         // Current line number
	column  int         // Current column number
	tokens  []Token     // Accumulated tokens
	errors  []error     // Accumulated errors
	lastTok *Token      // Last token produced
}

// New creates a new Lexer for the given input string.
//
// Parameters:
//   - input: The Cypher query string to tokenize
//
// Returns a new Lexer instance.
//
// Example:
//
//	l := lexer.New("MATCH (n:Person) RETURN n.name")
//	tokens, err := l.Tokenize()
func New(input string) *Lexer {
	return &Lexer{
		input:  strings.TrimSpace(input),
		line:   1,
		column: 1,
	}
}

// Tokenize processes the entire input and returns all tokens.
// It returns an error if any lexical errors are encountered.
//
// Returns the slice of tokens and any error encountered.
//
// Example:
//
//	l := lexer.New("MATCH (n:Person) RETURN n.name")
//	tokens, err := l.Tokenize()
//	if err != nil {
//	    log.Fatal(err)
//	}
func (l *Lexer) Tokenize() ([]Token, error) {
	for !l.atEnd() {
		l.skipWhitespace()
		if l.atEnd() {
			break
		}

		tok, err := l.nextToken()
		if err != nil {
			l.errors = append(l.errors, err)
			continue
		}
		if tok.Type != TokenError {
			l.tokens = append(l.tokens, tok)
			l.lastTok = &tok
		}
	}

	l.tokens = append(l.tokens, Token{
		Type:     TokenEOF,
		Line:     l.line,
		Column:   l.column,
		Position: l.pos,
	})

	if len(l.errors) > 0 {
		return l.tokens, &LexerErrors{Errors: l.errors}
	}
	return l.tokens, nil
}

// nextToken reads and returns the next token from the input.
func (l *Lexer) nextToken() (Token, error) {
	ch := l.peek()

	switch {
	case isDigit(ch):
		return l.readNumber()
	case isStringQuote(ch):
		return l.readString()
	case isLetter(ch) || ch == '_':
		return l.readIdentifier()
	case ch == '$':
		l.advance()
		if isLetter(l.peek()) || l.peek() == '_' {
			return l.readParameter()
		}
		return Token{Type: TokenDollar, Value: "$", Line: l.line, Column: l.column - 1}, nil
	case ch == '-':
		return l.readDashOrArrow()
	case ch == '<':
		return l.readLessOrArrow()
	case ch == '>':
		return l.readGreater()
	case ch == '=':
		return l.readEqual()
	case ch == '!':
		return l.readNotEqual()
	case ch == '+':
		return l.readPlusOrPlusEq()
	case ch == '.':
		return l.readDotOrRange()
	case ch == '*':
		l.advance()
		return Token{Type: TokenStar, Value: "*", Line: l.line, Column: l.column - 1}, nil
	case ch == '/':
		if l.peekNext() == '/' || l.peekNext() == '*' {
			l.skipComment()
			return Token{Type: TokenError}, nil
		}
		l.advance()
		return Token{Type: TokenSlash, Value: "/", Line: l.line, Column: l.column - 1}, nil
	case ch == '%':
		l.advance()
		return Token{Type: TokenPercent, Value: "%", Line: l.line, Column: l.column - 1}, nil
	case ch == '^':
		l.advance()
		return Token{Type: TokenCaret, Value: "^", Line: l.line, Column: l.column - 1}, nil
	case ch == '(':
		l.advance()
		return Token{Type: TokenLParen, Value: "(", Line: l.line, Column: l.column - 1}, nil
	case ch == ')':
		l.advance()
		return Token{Type: TokenRParen, Value: ")", Line: l.line, Column: l.column - 1}, nil
	case ch == '{':
		l.advance()
		return Token{Type: TokenLBrace, Value: "{", Line: l.line, Column: l.column - 1}, nil
	case ch == '}':
		l.advance()
		return Token{Type: TokenRBrace, Value: "}", Line: l.line, Column: l.column - 1}, nil
	case ch == '[':
		l.advance()
		return Token{Type: TokenLBracket, Value: "[", Line: l.line, Column: l.column - 1}, nil
	case ch == ']':
		l.advance()
		return Token{Type: TokenRBracket, Value: "]", Line: l.line, Column: l.column - 1}, nil
	case ch == ':':
		l.advance()
		return Token{Type: TokenColon, Value: ":", Line: l.line, Column: l.column - 1}, nil
	case ch == ',':
		l.advance()
		return Token{Type: TokenComma, Value: ",", Line: l.line, Column: l.column - 1}, nil
	case ch == ';':
		l.advance()
		return Token{Type: TokenSemicolon, Value: ";", Line: l.line, Column: l.column - 1}, nil
	case ch == '|':
		l.advance()
		return Token{Type: TokenPipe, Value: "|", Line: l.line, Column: l.column - 1}, nil
	default:
		l.advance()
		return Token{
			Type:     TokenError,
			Value:    string(ch),
			Line:     l.line,
			Column:   l.column - 1,
			Position: l.pos - 1,
		}, fmt.Errorf("unexpected character: %c at line %d, column %d", ch, l.line, l.column-1)
	}
}

// readNumber reads a numeric literal (integer or float).
func (l *Lexer) readNumber() (Token, error) {
	startLine := l.line
	startCol := l.column
	startPos := l.pos

	var isFloat bool

	for isDigit(l.peek()) {
		l.advance()
	}

	if l.peek() == '.' && isDigit(l.peekNext()) {
		isFloat = true
		l.advance()
		for isDigit(l.peek()) {
			l.advance()
		}
	}

	if l.peek() == 'e' || l.peek() == 'E' {
		isFloat = true
		l.advance()
		if l.peek() == '+' || l.peek() == '-' {
			l.advance()
		}
		for isDigit(l.peek()) {
			l.advance()
		}
	}

	value := l.input[startPos:l.pos]
	tokType := TokenInteger
	if isFloat {
		tokType = TokenFloat
	}

	return Token{
		Type:     tokType,
		Value:    value,
		Line:     startLine,
		Column:   startCol,
		Position: startPos,
	}, nil
}

// readString reads a string literal (single or double quoted).
func (l *Lexer) readString() (Token, error) {
	startLine := l.line
	startCol := l.column
	startPos := l.pos

	quote := l.peek()
	l.advance()

	var buf strings.Builder
	for !l.atEnd() && l.peek() != quote {
		if l.peek() == '\\' {
			l.advance()
			if l.atEnd() {
				break
			}
			switch l.peek() {
			case 'n':
				buf.WriteByte('\n')
			case 't':
				buf.WriteByte('\t')
			case 'r':
				buf.WriteByte('\r')
			case '\\':
				buf.WriteByte('\\')
			case '\'':
				buf.WriteByte('\'')
			case '"':
				buf.WriteByte('"')
			default:
				buf.WriteByte(l.peek())
			}
			l.advance()
		} else {
			if l.peek() == '\n' {
				l.line++
				l.column = 0
			}
			buf.WriteByte(l.peek())
			l.advance()
		}
	}

	if l.atEnd() {
		return Token{
			Type:     TokenError,
			Value:    l.input[startPos:l.pos],
			Line:     startLine,
			Column:   startCol,
			Position: startPos,
		}, fmt.Errorf("unterminated string at line %d, column %d", startLine, startCol)
	}

	l.advance()

	return Token{
		Type:     TokenString,
		Value:    buf.String(),
		Line:     startLine,
		Column:   startCol,
		Position: startPos,
	}, nil
}

// readIdentifier reads an identifier or keyword.
func (l *Lexer) readIdentifier() (Token, error) {
	startLine := l.line
	startCol := l.column
	startPos := l.pos

	for isLetter(l.peek()) || isDigit(l.peek()) || l.peek() == '_' {
		l.advance()
	}

	value := l.input[startPos:l.pos]
	upper := strings.ToUpper(value)

	tokType := TokenIdentifier
	if IsKeyword(upper) {
		tokType = TokenIdentifier
		value = upper
	}
	if upper == "TRUE" || upper == "FALSE" {
		tokType = TokenBoolean
		value = upper
	}
	if upper == "NULL" {
		tokType = TokenNull
		value = upper
	}

	return Token{
		Type:     tokType,
		Value:    value,
		Line:     startLine,
		Column:   startCol,
		Position: startPos,
	}, nil
}

// readParameter reads a parameter reference (e.g., $name).
func (l *Lexer) readParameter() (Token, error) {
	startLine := l.line
	startCol := l.column
	startPos := l.pos - 1

	for isLetter(l.peek()) || isDigit(l.peek()) || l.peek() == '_' {
		l.advance()
	}

	return Token{
		Type:     TokenIdentifier,
		Value:    l.input[startPos:l.pos],
		Line:     startLine,
		Column:   startCol,
		Position: startPos,
	}, nil
}

// readDashOrArrow reads a dash or right arrow (->).
func (l *Lexer) readDashOrArrow() (Token, error) {
	startLine := l.line
	startCol := l.column
	startPos := l.pos

	l.advance()

	if l.peek() == '>' {
		l.advance()
		return Token{
			Type:     TokenArrowRight,
			Value:    "->",
			Line:     startLine,
			Column:   startCol,
			Position: startPos,
		}, nil
	}

	return Token{
		Type:     TokenDash,
		Value:    "-",
		Line:     startLine,
		Column:   startCol,
		Position: startPos,
	}, nil
}

// readLessOrArrow reads a less-than sign or left arrow (<-).
func (l *Lexer) readLessOrArrow() (Token, error) {
	startLine := l.line
	startCol := l.column
	startPos := l.pos

	l.advance()

	if l.peek() == '-' {
		l.advance()
		return Token{
			Type:     TokenArrowLeft,
			Value:    "<-",
			Line:     startLine,
			Column:   startCol,
			Position: startPos,
		}, nil
	}

	if l.peek() == '=' {
		l.advance()
		return Token{
			Type:     TokenLe,
			Value:    "<=",
			Line:     startLine,
			Column:   startCol,
			Position: startPos,
		}, nil
	}

	return Token{
		Type:     TokenLt,
		Value:    "<",
		Line:     startLine,
		Column:   startCol,
		Position: startPos,
	}, nil
}

// readGreater reads a greater-than sign or greater-than-or-equal (>=).
func (l *Lexer) readGreater() (Token, error) {
	startLine := l.line
	startCol := l.column
	startPos := l.pos

	l.advance()

	if l.peek() == '=' {
		l.advance()
		return Token{
			Type:     TokenGe,
			Value:    ">=",
			Line:     startLine,
			Column:   startCol,
			Position: startPos,
		}, nil
	}

	return Token{
		Type:     TokenGt,
		Value:    ">",
		Line:     startLine,
		Column:   startCol,
		Position: startPos,
	}, nil
}

// readEqual reads an equal sign.
func (l *Lexer) readEqual() (Token, error) {
	startLine := l.line
	startCol := l.column
	startPos := l.pos

	l.advance()

	return Token{
		Type:     TokenEq,
		Value:    "=",
		Line:     startLine,
		Column:   startCol,
		Position: startPos,
	}, nil
}

// readNotEqual reads a not-equal sign (!=).
func (l *Lexer) readNotEqual() (Token, error) {
	startLine := l.line
	startCol := l.column
	startPos := l.pos

	l.advance()

	if l.peek() == '=' {
		l.advance()
		return Token{
			Type:     TokenNeq,
			Value:    "!=",
			Line:     startLine,
			Column:   startCol,
			Position: startPos,
		}, nil
	}

	return Token{
		Type:     TokenError,
		Value:    "!",
		Line:     startLine,
		Column:   startCol,
		Position: startPos,
	}, fmt.Errorf("unexpected character '!' at line %d, column %d", startLine, startCol)
}

// readPlusOrPlusEq reads a plus sign or plus-equal (+=).
func (l *Lexer) readPlusOrPlusEq() (Token, error) {
	startLine := l.line
	startCol := l.column
	startPos := l.pos

	l.advance()

	if l.peek() == '=' {
		l.advance()
		return Token{
			Type:     TokenPlusEq,
			Value:    "+=",
			Line:     startLine,
			Column:   startCol,
			Position: startPos,
		}, nil
	}

	return Token{
		Type:     TokenPlus,
		Value:    "+",
		Line:     startLine,
		Column:   startCol,
		Position: startPos,
	}, nil
}

// readDotOrRange reads a dot or range operator (..).
func (l *Lexer) readDotOrRange() (Token, error) {
	startLine := l.line
	startCol := l.column
	startPos := l.pos

	l.advance()

	if l.peek() == '.' {
		l.advance()
		return Token{
			Type:     TokenRange,
			Value:    "..",
			Line:     startLine,
			Column:   startCol,
			Position: startPos,
		}, nil
	}

	return Token{
		Type:     TokenDot,
		Value:    ".",
		Line:     startLine,
		Column:   startCol,
		Position: startPos,
	}, nil
}

// skipWhitespace skips over whitespace characters.
func (l *Lexer) skipWhitespace() {
	for !l.atEnd() && unicode.IsSpace(rune(l.peek())) {
		if l.peek() == '\n' {
			l.line++
			l.column = 0
		}
		l.advance()
	}
}

// skipComment skips over single-line and multi-line comments.
func (l *Lexer) skipComment() {
	if l.peek() == '/' && l.peekNext() == '/' {
		for !l.atEnd() && l.peek() != '\n' {
			l.advance()
		}
		return
	}

	if l.peek() == '/' && l.peekNext() == '*' {
		l.advance()
		l.advance()
		for !l.atEnd() {
			if l.peek() == '*' && l.peekNext() == '/' {
				l.advance()
				l.advance()
				return
			}
			if l.peek() == '\n' {
				l.line++
				l.column = 0
			}
			l.advance()
		}
	}
}

// atEnd returns true if the lexer has reached the end of input.
func (l *Lexer) atEnd() bool {
	return l.pos >= len(l.input)
}

// peek returns the current character without consuming it.
func (l *Lexer) peek() byte {
	if l.atEnd() {
		return 0
	}
	return l.input[l.pos]
}

// peekNext returns the next character without consuming it.
func (l *Lexer) peekNext() byte {
	if l.pos+1 >= len(l.input) {
		return 0
	}
	return l.input[l.pos+1]
}

// advance consumes and returns the current character.
func (l *Lexer) advance() {
	if !l.atEnd() {
		l.pos++
		l.column++
	}
}

// isDigit returns true if the character is a digit.
func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

// isLetter returns true if the character is a letter.
func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

// isStringQuote returns true if the character is a string quote.
func isStringQuote(ch byte) bool {
	return ch == '\'' || ch == '"'
}
