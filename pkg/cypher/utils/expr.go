// Package utils provides utility functions for the Cypher query engine.
// It includes helper functions for converting AST expressions to values.
package utils

import (
	"fmt"

	"github.com/DotNetAge/gograph/pkg/cypher/ast"
)

// ExprToValue converts an AST expression to its corresponding Go value.
// It handles literal expressions (integers, floats, strings, booleans)
// and parameter references.
//
// Parameters:
//   - expr: The AST expression to convert
//   - params: Query parameters for parameterized expressions
//
// Returns the Go value represented by the expression, or nil if the expression
// is nil or of an unsupported type.
//
// Example:
//
//	value := utils.ExprToValue(&ast.StringLit{Value: "hello"}, nil)
//	// value == "hello"
//
//	value := utils.ExprToValue(&ast.Param{Name: "name"}, map[string]interface{}{"name": "Alice"})
//	// value == "Alice"
func ExprToValue(expr ast.Expr, params map[string]interface{}) interface{} {
	if expr == nil {
		return nil
	}
	switch e := expr.(type) {
	case *ast.IntegerLit:
		return e.Value
	case *ast.FloatLit:
		return e.Value
	case *ast.StringLit:
		return e.Value
	case *ast.BoolLit:
		return e.Value
	case *ast.Param:
		return params[e.Name]
	case *ast.ParamExpr:
		return params[e.Name]
	}
	return nil
}

// ExprToString converts an AST expression to a string value.
// It handles string literals, identifiers, and other expression types
// by converting them to their string representation.
//
// Parameters:
//   - expr: The AST expression to convert
//   - params: Query parameters for parameterized expressions
//
// Returns the string representation of the expression, or an empty string
// if the expression is nil.
//
// Example:
//
//	str := utils.ExprToString(&ast.StringLit{Value: "hello"}, nil)
//	// str == "hello"
//
//	str := utils.ExprToString(&ast.Ident{Name: "name"}, nil)
//	// str == "name"
func ExprToString(expr ast.Expr, params map[string]interface{}) string {
	if expr == nil {
		return ""
	}
	switch e := expr.(type) {
	case *ast.StringLit:
		return e.Value
	case *ast.Ident:
		return e.Name
	}
	return fmt.Sprintf("%v", ExprToValue(expr, params))
}
