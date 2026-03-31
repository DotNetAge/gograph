package ast

// Ident represents an identifier expression in Cypher.
// It is used for variable names, property names, and labels.
//
// Example Cypher:
//
//	n (variable)
//	name (property)
//	Person (label)
//
// Fields:
//   - Name: The identifier name
//   - Start: Position in source where this identifier begins
//   - EndPos: Position in source where this identifier ends
type Ident struct {
	Name   string
	Start  Pos
	EndPos Pos
}

// Position returns the starting position of this identifier in the source.
func (e *Ident) Position() Pos { return e.Start }

// End returns the ending position of this identifier in the source.
func (e *Ident) End() Pos { return e.EndPos }

// String returns the identifier name.
func (e *Ident) String() string { return e.Name }

// PropertyAccessExpr represents a property access expression (e.g., n.name).
// It accesses a property of a node, relationship, or map.
//
// Example Cypher:
//
//	n.name
//	user.email
//	person.address.city (nested access)
//
// Fields:
//   - Target: The expression whose property is being accessed
//   - Property: The property name
//   - Start: Position in source where this expression begins
//   - EndPos: Position in source where this expression ends
type PropertyAccessExpr struct {
	Target   Expr
	Property string
	Start    Pos
	EndPos   Pos
}

// Position returns the starting position of this expression in the source.
func (e *PropertyAccessExpr) Position() Pos { return e.Start }

// End returns the ending position of this expression in the source.
func (e *PropertyAccessExpr) End() Pos { return e.EndPos }

// String returns the property name.
func (e *PropertyAccessExpr) String() string {
	return e.Property
}

// BinaryExpr represents a binary expression with two operands and an operator.
// It is used for arithmetic, comparison, and logical operations.
//
// Example Cypher:
//
//	n.age > 25 (comparison)
//	a + b (arithmetic)
//	x AND y (logical)
//
// Fields:
//   - Left: The left operand
//   - Operator: The binary operator (e.g., +, -, *, /, =, <>, <, >, AND, OR)
//   - Right: The right operand
//   - Start: Position in source where this expression begins
//   - EndPos: Position in source where this expression ends
type BinaryExpr struct {
	Left     Expr
	Operator string
	Right    Expr
	Start    Pos
	EndPos   Pos
}

// Position returns the starting position of this expression in the source.
func (e *BinaryExpr) Position() Pos { return e.Start }

// End returns the ending position of this expression in the source.
func (e *BinaryExpr) End() Pos { return e.EndPos }

// String returns the operator.
func (e *BinaryExpr) String() string { return e.Operator }

// UnaryExpr represents a unary expression with an operator and operand.
// It is used for negation and other unary operations.
//
// Example Cypher:
//
//	-n (negation)
//	NOT x (logical not)
//
// Fields:
//   - Operator: The unary operator (e.g., -, NOT)
//   - Operand: The operand expression
//   - Start: Position in source where this expression begins
//   - EndPos: Position in source where this expression ends
type UnaryExpr struct {
	Operator string
	Operand  Expr
	Start    Pos
	EndPos   Pos
}

// Position returns the starting position of this expression in the source.
func (e *UnaryExpr) Position() Pos { return e.Start }

// End returns the ending position of this expression in the source.
func (e *UnaryExpr) End() Pos { return e.EndPos }

// String returns the operator.
func (e *UnaryExpr) String() string { return e.Operator }

// FuncCall represents a function call expression.
// It calls a built-in or user-defined function with arguments.
//
// Example Cypher:
//
//	count(*)
//	toUpper(name)
//	coalesce(a, b, c)
//
// Fields:
//   - Name: The function name
//   - Args: The function arguments
//   - Distinct: If true, applies DISTINCT to the arguments
//   - Start: Position in source where this expression begins
//   - EndPos: Position in source where this expression ends
type FuncCall struct {
	Name     string
	Args     []Expr
	Distinct bool
	Start    Pos
	EndPos   Pos
}

// Position returns the starting position of this expression in the source.
func (e *FuncCall) Position() Pos { return e.Start }

// End returns the ending position of this expression in the source.
func (e *FuncCall) End() Pos { return e.EndPos }

// String returns the function name.
func (e *FuncCall) String() string { return e.Name }

// CaseExpr represents a CASE expression for conditional logic.
// It supports both simple and searched CASE forms.
//
// Example Cypher:
//
//	CASE n.age
//	    WHEN 18 THEN 'young'
//	    WHEN 65 THEN 'senior'
//	    ELSE 'adult'
//	END
//
//	CASE
//	    WHEN n.age < 18 THEN 'minor'
//	    WHEN n.age >= 65 THEN 'senior'
//	    ELSE 'adult'
//	END
//
// Fields:
//   - Operand: Optional expression for simple CASE (nil for searched CASE)
//   - Whens: List of WHEN clauses
//   - Else: Optional ELSE expression
//   - Start: Position in source where this expression begins
//   - EndPos: Position in source where this expression ends
type CaseExpr struct {
	Operand Expr
	Whens   []*WhenClause
	Else    Expr
	Start   Pos
	EndPos  Pos
}

// Position returns the starting position of this expression in the source.
func (e *CaseExpr) Position() Pos { return e.Start }

// End returns the ending position of this expression in the source.
func (e *CaseExpr) End() Pos { return e.EndPos }

// String returns "CASE".
func (e *CaseExpr) String() string { return "CASE" }

// ExistsExpr represents an EXISTS subquery expression.
// It checks if a pattern exists in the graph.
//
// Example Cypher:
//
//	EXISTS { (n)-[:KNOWS]->(m) }
//	EXISTS { MATCH (n)-[:FRIENDS_WITH]->(m) WHERE m.active = true }
//
// Fields:
//   - Pattern: The pattern to check for existence
//   - Expr: Alternative expression form
//   - Start: Position in source where this expression begins
//   - EndPos: Position in source where this expression ends
type ExistsExpr struct {
	Pattern *PatternExpr
	Expr    Expr
	Start   Pos
	EndPos  Pos
}

// Position returns the starting position of this expression in the source.
func (e *ExistsExpr) Position() Pos { return e.Start }

// End returns the ending position of this expression in the source.
func (e *ExistsExpr) End() Pos { return e.EndPos }

// String returns "EXISTS".
func (e *ExistsExpr) String() string { return "EXISTS" }

// ListComprehension represents a list comprehension expression.
// It creates a new list by transforming and filtering an existing list.
//
// Example Cypher:
//
//	[x IN list WHERE x > 0 | x * 2]
//	[n IN nodes(p) | n.name]
//
// Fields:
//   - Var: The iteration variable name
//   - Variable: Alternative variable name field
//   - InExpr: The expression being iterated (input list)
//   - List: Alternative input list field
//   - Where: Optional filter condition
//   - Filter: Alternative filter field
//   - Expr: The projection expression
//   - Projection: Alternative projection field
//   - Start: Position in source where this expression begins
//   - EndPos: Position in source where this expression ends
type ListComprehension struct {
	Var        string
	Variable   string
	InExpr     Expr
	List       Expr
	Where      Expr
	Filter     Expr
	Expr       Expr
	Projection Expr
	Start      Pos
	EndPos     Pos
}

// Position returns the starting position of this expression in the source.
func (e *ListComprehension) Position() Pos { return e.Start }

// End returns the ending position of this expression in the source.
func (e *ListComprehension) End() Pos { return e.EndPos }

// String returns "ListComprehension".
func (e *ListComprehension) String() string { return "ListComprehension" }

// InExpr represents an IN expression for membership testing.
// It checks if a value is contained in a list.
//
// Example Cypher:
//
//	n.name IN ['Alice', 'Bob', 'Charlie']
//	x IN range(1, 10)
//
// Fields:
//   - Left: The value to check for membership
//   - Right: The list expression
//   - Start: Position in source where this expression begins
//   - EndPos: Position in source where this expression ends
type InExpr struct {
	Left  Expr
	Right Expr
	Start Pos
	EndPos Pos
}

// Position returns the starting position of this expression in the source.
func (e *InExpr) Position() Pos { return e.Start }

// End returns the ending position of this expression in the source.
func (e *InExpr) End() Pos { return e.EndPos }

// String returns "IN".
func (e *InExpr) String() string { return "IN" }

// IsNullExpr represents an IS NULL or IS NOT NULL expression.
// It checks if a value is null or not null.
//
// Example Cypher:
//
//	n.name IS NULL
//	n.email IS NOT NULL
//
// Fields:
//   - Expr: The expression to check
//   - IsNot: If true, this is IS NOT NULL
//   - Negate: Alternative negation flag
//   - Start: Position in source where this expression begins
//   - EndPos: Position in source where this expression ends
type IsNullExpr struct {
	Expr   Expr
	IsNot  bool
	Negate bool
	Start  Pos
	EndPos Pos
}

// Position returns the starting position of this expression in the source.
func (e *IsNullExpr) Position() Pos { return e.Start }

// End returns the ending position of this expression in the source.
func (e *IsNullExpr) End() Pos { return e.EndPos }

// String returns "IS NULL" or "IS NOT NULL".
func (e *IsNullExpr) String() string {
	if e.IsNot || e.Negate {
		return "IS NOT NULL"
	}
	return "IS NULL"
}

// ListSliceExpr represents a list slice expression.
// It extracts a portion of a list using start and end indices.
//
// Example Cypher:
//
//	list[0..3] (first 3 elements)
//	list[1..] (from index 1 to end)
//	list[..5] (first 5 elements)
//
// Fields:
//   - List: The list expression to slice
//   - From: The start index (nil means from beginning)
//   - To: The end index (nil means to end)
//   - Start: Position in source where this expression begins
//   - EndPos: Position in source where this expression ends
type ListSliceExpr struct {
	List   Expr
	From   Expr
	To     Expr
	Start  Pos
	EndPos Pos
}

// Position returns the starting position of this expression in the source.
func (e *ListSliceExpr) Position() Pos { return e.Start }

// End returns the ending position of this expression in the source.
func (e *ListSliceExpr) End() Pos { return e.EndPos }

// String returns "ListSlice".
func (e *ListSliceExpr) String() string { return "ListSlice" }

// ListIndexExpr represents a list index expression.
// It accesses an element of a list by index.
//
// Example Cypher:
//
//	list[0] (first element)
//	list[-1] (last element)
//
// Fields:
//   - List: The list expression
//   - Index: The index expression
//   - Start: Position in source where this expression begins
//   - EndPos: Position in source where this expression ends
type ListIndexExpr struct {
	List   Expr
	Index  Expr
	Start  Pos
	EndPos Pos
}

// Position returns the starting position of this expression in the source.
func (e *ListIndexExpr) Position() Pos { return e.Start }

// End returns the ending position of this expression in the source.
func (e *ListIndexExpr) End() Pos { return e.EndPos }

// String returns "ListIndex".
func (e *ListIndexExpr) String() string { return "ListIndex" }

// ListExpr represents a list expression.
// It is used for list literals and list operations.
//
// Example Cypher:
//
//	[1, 2, 3]
//	['a', 'b', 'c']
//
// Fields:
//   - Elements: The list elements
//   - Start: Position in source where this expression begins
//   - EndPos: Position in source where this expression ends
type ListExpr struct {
	Elements []Expr
	Start    Pos
	EndPos   Pos
}

// Position returns the starting position of this expression in the source.
func (e *ListExpr) Position() Pos { return e.Start }

// End returns the ending position of this expression in the source.
func (e *ListExpr) End() Pos { return e.EndPos }

// String returns "List".
func (e *ListExpr) String() string { return "List" }

// MapExpr represents a map expression.
// It is used for map literals with key-value pairs.
//
// Example Cypher:
//
//	{name: 'Alice', age: 30}
//	{key: value, another: other}
//
// Fields:
//   - Pairs: The key-value pairs
//   - Start: Position in source where this expression begins
//   - EndPos: Position in source where this expression ends
type MapExpr struct {
	Pairs  []*MapPair
	Start  Pos
	EndPos Pos
}

// Position returns the starting position of this expression in the source.
func (e *MapExpr) Position() Pos { return e.Start }

// End returns the ending position of this expression in the source.
func (e *MapExpr) End() Pos { return e.EndPos }

// String returns "Map".
func (e *MapExpr) String() string { return "Map" }

// MapPair represents a key-value pair in a map expression.
//
// Fields:
//   - Key: The key name
//   - Value: The value expression
//   - Start: Position in source where this pair begins
//   - EndPos: Position in source where this pair ends
type MapPair struct {
	Key    string
	Value  Expr
	Start  Pos
	EndPos Pos
}

// Position returns the starting position of this pair in the source.
func (p *MapPair) Position() Pos { return p.Start }

// End returns the ending position of this pair in the source.
func (p *MapPair) End() Pos { return p.EndPos }

// String returns the key name.
func (p *MapPair) String() string { return p.Key }

// WhenClause represents a WHEN clause in a CASE expression.
// It pairs a condition with a result.
//
// Example Cypher:
//
//	WHEN 18 THEN 'young'
//
// Fields:
//   - Condition: The condition to match (for searched CASE)
//   - Result: The result expression if the condition matches
//   - Start: Position in source where this clause begins
//   - EndPos: Position in source where this clause ends
type WhenClause struct {
	Condition Expr
	Result    Expr
	Start     Pos
	EndPos    Pos
}

// Position returns the starting position of this clause in the source.
func (c *WhenClause) Position() Pos { return c.Start }

// End returns the ending position of this clause in the source.
func (c *WhenClause) End() Pos { return c.EndPos }

// String returns "WHEN".
func (c *WhenClause) String() string { return "WHEN" }
