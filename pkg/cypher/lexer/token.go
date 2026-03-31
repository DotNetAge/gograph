package lexer

import (
	"strings"
)

// TokenType represents the type of a lexical token.
type TokenType int

// Token type constants.
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

// keywords contains all Cypher reserved keywords.
var keywords = map[string]bool{
	"MATCH":            true,
	"OPTIONAL":         true,
	"WHERE":            true,
	"RETURN":           true,
	"CREATE":           true,
	"DELETE":           true,
	"DETACH":           true,
	"SET":              true,
	"REMOVE":           true,
	"MERGE":            true,
	"WITH":             true,
	"UNWIND":           true,
	"AS":               true,
	"DISTINCT":         true,
	"ORDER":            true,
	"BY":               true,
	"ASC":              true,
	"DESC":             true,
	"LIMIT":            true,
	"SKIP":             true,
	"AND":              true,
	"OR":               true,
	"XOR":              true,
	"NOT":              true,
	"IN":               true,
	"IS":               true,
	"NULL":             true,
	"TRUE":             true,
	"FALSE":            true,
	"CONTAINS":         true,
	"STARTS":           true,
	"ENDS":             true,
	"WITHIN":           true,
	"CASE":             true,
	"WHEN":             true,
	"THEN":             true,
	"ELSE":             true,
	"END":              true,
	"UNION":            true,
	"ALL":              true,
	"CALL":             true,
	"YIELD":            true,
	"LOAD":             true,
	"CSV":              true,
	"FROM":             true,
	"FOREACH":          true,
	"ON":               true,
	"INDEX":            true,
	"CONSTRAINT":       true,
	"ASSERT":           true,
	"DROP":             true,
	"USING":            true,
	"PERIODIC":         true,
	"COMMIT":           true,
	"SCHEMA":           true,
	"AWAIT":            true,
	"POINT":            true,
	"RANGE":            true,
	"LOOKUP":           true,
	"JOIN":             true,
	"SHORTESTPATH":     true,
	"ALLSHORTESTPATHS": true,
	"COUNT":            true,
	"EXISTS":           true,
	"ANY":              true,
	"NONE":             true,
	"SINGLE":           true,
	"REDUCE":           true,
	"EXTRACT":          true,
	"FILTER":           true,
	"TAIL":             true,
	"HEAD":             true,
	"LAST":             true,
	"NODES":            true,
	"RELATIONSHIPS":    true,
	"LABELS":           true,
	"KEYS":             true,
	"PROPERTIES":       true,
	"LENGTH":           true,
	"SIZE":             true,
	"TYPE":             true,
	"ID":               true,
	"COALESCE":         true,
	"IF":               true,
}

// IsKeyword returns true if the given string is a Cypher keyword.
//
// Parameters:
//   - word: The string to check
//
// Returns true if the word is a keyword.
//
// Example:
//
//	if lexer.IsKeyword("MATCH") {
//	    // Handle keyword
//	}
func IsKeyword(word string) bool {
	return keywords[word]
}

// Token represents a lexical token with its type, value, and position.
type Token struct {
	Type     TokenType // The type of the token
	Value    string    // The string value of the token
	Line     int       // Line number where the token appears
	Column   int       // Column number where the token appears
	Position int       // Byte position in the input
}

// String returns the string value of the token.
func (t Token) String() string {
	return t.Value
}

// IsKeyword returns true if the token matches the given keyword.
func (t Token) IsKeyword(name string) bool {
	return t.Value == name
}

// IsEOF returns true if the token is an EOF token.
func (t Token) IsEOF() bool {
	return t.Type == TokenEOF
}

// IsError returns true if the token is an error token.
func (t Token) IsError() bool {
	return t.Type == TokenError
}

// LexerError represents a lexical error with position information.
type LexerError struct {
	Line    int    // Line number where the error occurred
	Column  int    // Column number where the error occurred
	Message string // Error message
}

// Error returns the error message.
func (e *LexerError) Error() string {
	return e.Message
}

// LexerErrors aggregates multiple lexical errors.
type LexerErrors struct {
	Errors []error // Slice of errors
}

// Error returns a string containing all error messages.
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
