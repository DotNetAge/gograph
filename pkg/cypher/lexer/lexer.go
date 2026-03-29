package lexer

import (
	"fmt"
	"strings"
	"unicode"
)

type Lexer struct {
	input    string
	pos      int
	line     int
	column   int
	tokens   []Token
	errors   []error
	lastTok  *Token
}

func New(input string) *Lexer {
	return &Lexer{
		input:  strings.TrimSpace(input),
		line:   1,
		column: 1,
	}
}

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

func (l *Lexer) skipWhitespace() {
	for !l.atEnd() && unicode.IsSpace(rune(l.peek())) {
		if l.peek() == '\n' {
			l.line++
			l.column = 0
		}
		l.advance()
	}
}

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

func (l *Lexer) atEnd() bool {
	return l.pos >= len(l.input)
}

func (l *Lexer) peek() byte {
	if l.atEnd() {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) peekNext() byte {
	if l.pos+1 >= len(l.input) {
		return 0
	}
	return l.input[l.pos+1]
}

func (l *Lexer) advance() {
	if !l.atEnd() {
		l.pos++
		l.column++
	}
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isStringQuote(ch byte) bool {
	return ch == '\'' || ch == '"'
}
