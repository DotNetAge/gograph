package lexer_test

import (
	"testing"

	"github.com/DotNetAge/gograph/pkg/cypher/lexer"
)

func TestLexer_TokenTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []struct {
			typ   lexer.TokenType
			value string
		}
	}{
		{
			name:  "integer literal",
			input: "12345",
			expected: []struct {
				typ   lexer.TokenType
				value string
			}{
				{lexer.TokenInteger, "12345"},
			},
		},
		{
			name:  "float literal",
			input: "3.14159",
			expected: []struct {
				typ   lexer.TokenType
				value string
			}{
				{lexer.TokenFloat, "3.14159"},
			},
		},
		{
			name:  "scientific notation",
			input: "1.5e10 2.5E-5",
			expected: []struct {
				typ   lexer.TokenType
				value string
			}{
				{lexer.TokenFloat, "1.5e10"},
				{lexer.TokenFloat, "2.5E-5"},
			},
		},
		{
			name:  "string single quote",
			input: "'hello world'",
			expected: []struct {
				typ   lexer.TokenType
				value string
			}{
				{lexer.TokenString, "hello world"},
			},
		},
		{
			name:  "string double quote",
			input: `"hello world"`,
			expected: []struct {
				typ   lexer.TokenType
				value string
			}{
				{lexer.TokenString, "hello world"},
			},
		},
		{
			name:  "boolean true",
			input: "true TRUE True",
			expected: []struct {
				typ   lexer.TokenType
				value string
			}{
				{lexer.TokenBoolean, "TRUE"},
				{lexer.TokenBoolean, "TRUE"},
				{lexer.TokenBoolean, "TRUE"},
			},
		},
		{
			name:  "boolean false",
			input: "false FALSE False",
			expected: []struct {
				typ   lexer.TokenType
				value string
			}{
				{lexer.TokenBoolean, "FALSE"},
				{lexer.TokenBoolean, "FALSE"},
				{lexer.TokenBoolean, "FALSE"},
			},
		},
		{
			name:  "null",
			input: "null NULL Null",
			expected: []struct {
				typ   lexer.TokenType
				value string
			}{
				{lexer.TokenNull, "NULL"},
				{lexer.TokenNull, "NULL"},
				{lexer.TokenNull, "NULL"},
			},
		},
		{
			name:  "identifier",
			input: "name _var $param",
			expected: []struct {
				typ   lexer.TokenType
				value string
			}{
				{lexer.TokenIdentifier, "name"},
				{lexer.TokenIdentifier, "_var"},
				{lexer.TokenIdentifier, "$param"},
			},
		},
		{
			name:  "keywords",
			input: "MATCH CREATE MERGE SET DELETE REMOVE RETURN WHERE WITH UNWIND UNION",
			expected: []struct {
				typ   lexer.TokenType
				value string
			}{
				{lexer.TokenIdentifier, "MATCH"},
				{lexer.TokenIdentifier, "CREATE"},
				{lexer.TokenIdentifier, "MERGE"},
				{lexer.TokenIdentifier, "SET"},
				{lexer.TokenIdentifier, "DELETE"},
				{lexer.TokenIdentifier, "REMOVE"},
				{lexer.TokenIdentifier, "RETURN"},
				{lexer.TokenIdentifier, "WHERE"},
				{lexer.TokenIdentifier, "WITH"},
				{lexer.TokenIdentifier, "UNWIND"},
				{lexer.TokenIdentifier, "UNION"},
			},
		},
		{
			name:  "arithmetic operators",
			input: "+ - * / % ^",
			expected: []struct {
				typ   lexer.TokenType
				value string
			}{
				{lexer.TokenPlus, "+"},
				{lexer.TokenDash, "-"},
				{lexer.TokenStar, "*"},
				{lexer.TokenSlash, "/"},
				{lexer.TokenPercent, "%"},
				{lexer.TokenCaret, "^"},
			},
		},
		{
			name:  "comparison operators",
			input: "= != < <= > >=",
			expected: []struct {
				typ   lexer.TokenType
				value string
			}{
				{lexer.TokenEq, "="},
				{lexer.TokenNeq, "!="},
				{lexer.TokenLt, "<"},
				{lexer.TokenLe, "<="},
				{lexer.TokenGt, ">"},
				{lexer.TokenGe, ">="},
			},
		},
		{
			name:  "logical keywords",
			input: "AND OR NOT XOR",
			expected: []struct {
				typ   lexer.TokenType
				value string
			}{
				{lexer.TokenIdentifier, "AND"},
				{lexer.TokenIdentifier, "OR"},
				{lexer.TokenIdentifier, "NOT"},
				{lexer.TokenIdentifier, "XOR"},
			},
		},
		{
			name:  "delimiters",
			input: "() {} [] : , . ; |",
			expected: []struct {
				typ   lexer.TokenType
				value string
			}{
				{lexer.TokenLParen, "("},
				{lexer.TokenRParen, ")"},
				{lexer.TokenLBrace, "{"},
				{lexer.TokenRBrace, "}"},
				{lexer.TokenLBracket, "["},
				{lexer.TokenRBracket, "]"},
				{lexer.TokenColon, ":"},
				{lexer.TokenComma, ","},
				{lexer.TokenDot, "."},
				{lexer.TokenSemicolon, ";"},
				{lexer.TokenPipe, "|"},
			},
		},
		{
			name:  "arrow operators",
			input: "-> <- --",
			expected: []struct {
				typ   lexer.TokenType
				value string
			}{
				{lexer.TokenArrowRight, "->"},
				{lexer.TokenArrowLeft, "<-"},
				{lexer.TokenDash, "-"},
			},
		},
		{
			name:  "plus equals",
			input: "+=",
			expected: []struct {
				typ   lexer.TokenType
				value string
			}{
				{lexer.TokenPlusEq, "+="},
			},
		},
		{
			name:  "range operator",
			input: "..",
			expected: []struct {
				typ   lexer.TokenType
				value string
			}{
				{lexer.TokenRange, ".."},
			},
		},
		{
			name:  "parameter",
			input: "$name $param1",
			expected: []struct {
				typ   lexer.TokenType
				value string
			}{
				{lexer.TokenIdentifier, "$name"},
				{lexer.TokenIdentifier, "$param1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			tokens, err := l.Tokenize()
			if err != nil {
				t.Fatalf("lexer error: %v", err)
			}

			for i, exp := range tt.expected {
				if i >= len(tokens)-1 {
					t.Errorf("not enough tokens: expected %d, got %d", len(tt.expected), len(tokens)-1)
					break
				}
				if tokens[i].Type != exp.typ {
					t.Errorf("token[%d] type: got %v, want %v", i, tokens[i].Type, exp.typ)
				}
				if tokens[i].Value != exp.value {
					t.Errorf("token[%d] value: got %q, want %q", i, tokens[i].Value, exp.value)
				}
			}
		})
	}
}

func TestLexer_EscapeSequences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"newline", `'hello\nworld'`, "hello\nworld"},
		{"tab", `'hello\tworld'`, "hello\tworld"},
		{"carriage return", `'hello\rworld'`, "hello\rworld"},
		{"backslash", `'hello\\world'`, "hello\\world"},
		{"single quote", `'hello\'world'`, "hello'world"},
		{"double quote", `"hello\"world"`, "hello\"world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			tokens, err := l.Tokenize()
			if err != nil {
				t.Fatalf("lexer error: %v", err)
			}
			if len(tokens) < 1 {
				t.Fatal("no tokens")
			}
			if tokens[0].Value != tt.expected {
				t.Errorf("value: got %q, want %q", tokens[0].Value, tt.expected)
			}
		})
	}
}

func TestLexer_Comments(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"single line comment", "MATCH // this is a comment\n(n:Person)"},
		{"multi line comment", "MATCH /* this is\na comment */ (n:Person)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			tokens, err := l.Tokenize()
			if err != nil {
				t.Fatalf("lexer error: %v", err)
			}
			
			hasMatch := false
			for _, tok := range tokens {
				if tok.Value == "MATCH" {
					hasMatch = true
					break
				}
			}
			if !hasMatch {
				t.Error("expected MATCH token")
			}
		})
	}
}

func TestLexer_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"unterminated string", `'hello`},
		{"unterminated string double", `"hello`},
		{"invalid character", "MATCH (n:Person) @ "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			_, err := l.Tokenize()
			if err == nil {
				t.Error("expected error but got none")
			}
		})
	}
}

func TestLexer_Position(t *testing.T) {
	input := "MATCH\n(n:Person)"
	l := lexer.New(input)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	expected := []struct {
		value      string
		line, col  int
	}{
		{"MATCH", 1, 1},
		{"(", 2, 1},
		{"n", 2, 2},
		{":", 2, 3},
		{"Person", 2, 4},
		{")", 2, 10},
	}

	for i, exp := range expected {
		if i >= len(tokens) {
			break
		}
		if tokens[i].Value != exp.value {
			t.Errorf("token[%d] value: got %q, want %q", i, tokens[i].Value, exp.value)
		}
		if tokens[i].Line != exp.line {
			t.Errorf("token[%d] line: got %d, want %d", i, tokens[i].Line, exp.line)
		}
		if tokens[i].Column != exp.col {
			t.Errorf("token[%d] column: got %d, want %d", i, tokens[i].Column, exp.col)
		}
	}
}
