package utils

import (
	"fmt"

	"github.com/DotNetAge/gograph/pkg/cypher/ast"
)

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
