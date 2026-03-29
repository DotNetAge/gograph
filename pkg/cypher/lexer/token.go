package lexer

import (
	"strings"
)

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenError
	TokenInteger
	TokenFloat
	TokenString
	TokenBoolean
	TokenNull
	TokenIdentifier
	TokenPlus
	TokenStar
	TokenSlash
	TokenPercent
	TokenCaret
	TokenEq
	TokenNeq
	TokenLt
	TokenLe
	TokenGt
	TokenGe
	TokenAnd
	TokenOr
	TokenNot
	TokenXor
	TokenLParen
	TokenRParen
	TokenLBrace
	TokenRBrace
	TokenLBracket
	TokenRBracket
	TokenColon
	TokenComma
	TokenDot
	TokenSemicolon
	TokenPipe
	TokenArrowRight
	TokenArrowLeft
	TokenDash
	TokenPlusEq
	TokenRange
	TokenDollar
)

type Token struct {
	Type     TokenType
	Value    string
	Line     int
	Column   int
	Position int
}

func (t Token) String() string {
	return t.Value
}

func (t Token) IsKeyword(name string) bool {
	return t.Value == name
}

func (t Token) IsEOF() bool {
	return t.Type == TokenEOF
}

func (t Token) IsError() bool {
	return t.Type == TokenError
}

type LexerError struct {
	Line    int
	Column  int
	Message string
}

func (e *LexerError) Error() string {
	return e.Message
}

type LexerErrors struct {
	Errors []error
}

func (e *LexerErrors) Error() string {
	var sb strings.Builder
	for i, err := range e.Errors {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(err.Error())
	}
	return sb.String()
}

func (e *LexerErrors) Unwrap() []error {
	return e.Errors
}

var keywords = map[string]bool{
	"MATCH": true, "CREATE": true, "MERGE": true, "SET": true,
	"DELETE": true, "DETACH": true, "REMOVE": true, "RETURN": true,
	"WHERE": true, "WITH": true, "UNWIND": true, "UNION": true,
	"ALL": true, "DISTINCT": true, "AS": true, "ORDER": true,
	"BY": true, "ASC": true, "DESC": true, "SKIP": true, "LIMIT": true,
	"AND": true, "OR": true, "NOT": true, "XOR": true, "IN": true,
	"IS": true, "NULL": true, "TRUE": true, "FALSE": true,
	"CASE": true, "WHEN": true, "THEN": true, "ELSE": true, "END": true,
	"ON": true, "EXISTS": true, "FOR": true, "EACH": true,
	"ASSERT": true, "CONSTRAINT": true, "INDEX": true, "DROP": true,
	"OPTIONAL": true,
	"COUNT": true, "SUM": true, "AVG": true, "MIN": true, "MAX": true,
	"COLLECT": true, "HEAD": true, "LAST": true, "TAIL": true,
	"SIZE": true, "RANGE": true, "REVERSE": true,
	"ID": true, "LABELS": true, "TYPE": true, "PROPERTIES": true,
	"NODES": true, "RELATIONSHIPS": true, "LENGTH": true,
	"ABS": true, "CEIL": true, "FLOOR": true, "ROUND": true, "RAND": true,
	"TOUPPER": true, "TOLOWER": true, "REPLACE": true, "SUBSTRING": true,
	"TRIM": true, "LTRIM": true, "RTRIM": true,
	"CONTAINS": true, "STARTS": true, "ENDS": true,
	"DATE": true, "DATETIME": true, "TIME": true, "TIMESTAMP": true,
	"DURATION": true, "LOCALDATETIME": true, "LOCALTIME": true,
	"COALESCE": true, "NULLIF": true,
}

func IsKeyword(s string) bool {
	return keywords[s]
}
