package ast

// IntegerLit represents an integer literal expression.
// It holds a signed 64-bit integer value.
//
// Example Cypher:
//
//	42
//	-100
//	0
//
// Fields:
//   - Value: The integer value
//   - Start: Position in source where this literal begins
//   - EndPos: Position in source where this literal ends
type IntegerLit struct {
	Value  int64
	Start  Pos
	EndPos Pos
}

// Position returns the starting position of this literal in the source.
func (e *IntegerLit) Position() Pos { return e.Start }

// End returns the ending position of this literal in the source.
func (e *IntegerLit) End() Pos { return e.EndPos }

// String returns "Integer".
func (e *IntegerLit) String() string { return "Integer" }

// FloatLit represents a floating-point literal expression.
// It holds a 64-bit floating-point value.
//
// Example Cypher:
//
//	3.14159
//	-0.5
//	1.0e10
//
// Fields:
//   - Value: The floating-point value
//   - Start: Position in source where this literal begins
//   - EndPos: Position in source where this literal ends
type FloatLit struct {
	Value  float64
	Start  Pos
	EndPos Pos
}

// Position returns the starting position of this literal in the source.
func (e *FloatLit) Position() Pos { return e.Start }

// End returns the ending position of this literal in the source.
func (e *FloatLit) End() Pos { return e.EndPos }

// String returns "Float".
func (e *FloatLit) String() string { return "Float" }

// StringLit represents a string literal expression.
// It holds a string value (quotes are not included).
//
// Example Cypher:
//
//	'hello'
//	"world"
//	'it\'s a test'
//
// Fields:
//   - Value: The string value (without quotes)
//   - Start: Position in source where this literal begins
//   - EndPos: Position in source where this literal ends
type StringLit struct {
	Value  string
	Start  Pos
	EndPos Pos
}

// Position returns the starting position of this literal in the source.
func (e *StringLit) Position() Pos { return e.Start }

// End returns the ending position of this literal in the source.
func (e *StringLit) End() Pos { return e.EndPos }

// String returns "String".
func (e *StringLit) String() string { return "String" }

// BoolLit represents a boolean literal expression.
// It holds a true or false value.
//
// Example Cypher:
//
//	true
//	false
//
// Fields:
//   - Value: The boolean value
//   - Start: Position in source where this literal begins
//   - EndPos: Position in source where this literal ends
type BoolLit struct {
	Value  bool
	Start  Pos
	EndPos Pos
}

// Position returns the starting position of this literal in the source.
func (e *BoolLit) Position() Pos { return e.Start }

// End returns the ending position of this literal in the source.
func (e *BoolLit) End() Pos { return e.EndPos }

// String returns "Bool".
func (e *BoolLit) String() string { return "Bool" }

// NullLit represents a NULL literal expression.
// It represents the absence of a value.
//
// Example Cypher:
//
//	NULL
//
// Fields:
//   - Start: Position in source where this literal begins
//   - EndPos: Position in source where this literal ends
type NullLit struct {
	Start  Pos
	EndPos Pos
}

// Position returns the starting position of this literal in the source.
func (e *NullLit) Position() Pos { return e.Start }

// End returns the ending position of this literal in the source.
func (e *NullLit) End() Pos { return e.EndPos }

// String returns "NULL".
func (e *NullLit) String() string { return "NULL" }

// Param represents a parameter placeholder expression.
// It references a parameter by name without the $ prefix.
//
// Example Cypher:
//
//	$name
//	{age: $minAge}
//
// Fields:
//   - Name: The parameter name (without $)
//   - Start: Position in source where this parameter begins
//   - EndPos: Position in source where this parameter ends
type Param struct {
	Name   string
	Start  Pos
	EndPos Pos
}

// Position returns the starting position of this parameter in the source.
func (e *Param) Position() Pos { return e.Start }

// End returns the ending position of this parameter in the source.
func (e *Param) End() Pos { return e.EndPos }

// String returns "$" + name.
func (e *Param) String() string { return "$" + e.Name }

// ParamExpr represents a parameter expression.
// It is functionally equivalent to Param.
//
// Example Cypher:
//
//	$name
//
// Fields:
//   - Name: The parameter name (without $)
//   - Start: Position in source where this parameter begins
//   - EndPos: Position in source where this parameter ends
type ParamExpr struct {
	Name   string
	Start  Pos
	EndPos Pos
}

// Position returns the starting position of this parameter in the source.
func (e *ParamExpr) Position() Pos { return e.Start }

// End returns the ending position of this parameter in the source.
func (e *ParamExpr) End() Pos { return e.EndPos }

// String returns "$" + name.
func (e *ParamExpr) String() string { return "$" + e.Name }

// ListLit represents a list literal expression.
// It holds a list of expressions.
//
// Example Cypher:
//
//	[1, 2, 3]
//	['a', 'b', 'c']
//	[] (empty list)
//
// Fields:
//   - Elements: The list elements
//   - Start: Position in source where this literal begins
//   - EndPos: Position in source where this literal ends
type ListLit struct {
	Elements []Expr
	Start    Pos
	EndPos   Pos
}

// Position returns the starting position of this literal in the source.
func (e *ListLit) Position() Pos { return e.Start }

// End returns the ending position of this literal in the source.
func (e *ListLit) End() Pos { return e.EndPos }

// String returns "List".
func (e *ListLit) String() string { return "List" }

// MapLit represents a map literal expression.
// It holds a map of string keys to expression values.
//
// Example Cypher:
//
//	{name: 'Alice', age: 30}
//	{} (empty map)
//
// Fields:
//   - Properties: The map properties (key-value pairs)
//   - Start: Position in source where this literal begins
//   - EndPos: Position in source where this literal ends
type MapLit struct {
	Properties map[string]Expr
	Start      Pos
	EndPos     Pos
}

// Position returns the starting position of this literal in the source.
func (e *MapLit) Position() Pos { return e.Start }

// End returns the ending position of this literal in the source.
func (e *MapLit) End() Pos { return e.EndPos }

// String returns "Map".
func (e *MapLit) String() string { return "Map" }
