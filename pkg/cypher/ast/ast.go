// Package ast defines the abstract syntax tree (AST) representation for Cypher queries.
// It provides data structures that represent the parsed structure of Cypher
// statements including CREATE, MATCH, SET, DELETE, and REMOVE clauses.
package ast

// AST represents a parsed Cypher query containing multiple statements.
type AST struct {
	Statements []Statement
}

// Statement represents a single statement in a Cypher query.
type Statement struct {
	Clause Clause
}

// Clause is implemented by all Cypher clause types.
// It is used to identify the type of clause in a statement.
type Clause interface {
	isClause()
}

// CreateClause represents a CREATE clause that creates new nodes and relationships.
type CreateClause struct {
	Pattern Pattern
}

// MatchClause represents a MATCH clause that queries for existing graph patterns.
// It may include WHERE conditions, RETURN projections, and DELETE operations.
type MatchClause struct {
	Pattern Pattern
	Where   *WhereClause
	Return  *ReturnClause
	Delete  *DeleteClause
}

// SetClause represents a SET clause that updates node properties.
type SetClause struct {
	Assignments []Assignment
	Where       *WhereClause
}

// DeleteClause represents a DELETE clause that removes nodes and relationships.
// The Detach field indicates whether to also delete connected relationships.
type DeleteClause struct {
	Detach      bool
	Expressions []Expression
	Where       *WhereClause
}

// RemoveClause represents a REMOVE clause that removes labels or properties from nodes.
type RemoveClause struct {
	Removals []RemoveItem
	Where    *WhereClause
}

func (c *CreateClause) isClause() {}
func (c *MatchClause) isClause()  {}
func (c *SetClause) isClause()    {}
func (c *DeleteClause) isClause() {}
func (c *RemoveClause) isClause() {}

// Pattern represents a graph pattern consisting of connected nodes and relationships.
type Pattern struct {
	Elements []PatternElement
}

// PatternElement represents a node pattern optionally followed by a relationship pattern.
type PatternElement struct {
	Node     *NodePattern
	Relation *RelationPattern
}

// NodePattern represents a node in a Cypher pattern, with optional variable, labels, and properties.
type NodePattern struct {
	Variable   string
	Labels     []string
	Properties map[string]interface{}
}

// RelationPattern represents a relationship pattern between two nodes.
// It includes the relationship type, direction, variable, and properties.
type RelationPattern struct {
	Variable   string
	RelType    string
	Dir        RelDirection
	StartNode  *NodePattern
	EndNode    *NodePattern
	Properties map[string]interface{}
}

// RelDirection represents the direction of a relationship pattern.
type RelDirection string

const (
	RelDirOutgoing RelDirection = "->"
	RelDirIncoming RelDirection = "<-"
	RelDirBoth     RelDirection = "-"
)

// WhereClause represents a WHERE clause containing a filter expression.
type WhereClause struct {
	Expression Expression
}

// ReturnClause represents a RETURN clause specifying which expressions to return.
type ReturnClause struct {
	Items []ReturnItem
}

// ReturnItem represents a single return expression with an optional alias.
type ReturnItem struct {
	Expression Expression
	Alias      string
}

// Assignment represents a property assignment in a SET clause.
type Assignment struct {
	Property PropertyAccess
	Value    Expression
}

// RemoveItem represents a single removal operation in a REMOVE clause.
type RemoveItem struct {
	Type     RemoveItemType
	Label    string
	Property PropertyAccess
}

// RemoveItemType specifies the type of removal operation.
type RemoveItemType int

const (
	RemoveItemTypeLabel RemoveItemType = iota
	RemoveItemTypeProperty
)

// Expression is implemented by all expression types in the AST.
type Expression interface {
	isExpression()
}

// PropertyAccess represents a property lookup expression (e.g., n.property).
type PropertyAccess struct {
	Node     string
	Property string
}

// Literal represents a literal value expression.
type Literal struct {
	Value interface{}
}

// PropertyLookup represents a property lookup on a node or relationship.
type PropertyLookup struct {
	Node     string
	Property string
}

// ComparisonOp represents a comparison expression with a left operand, operator, and right operand.
type ComparisonOp struct {
	Left     Expression
	Operator string
	Right    Expression
}

// Identifier represents an identifier expression.
type Identifier struct {
	Name string
}

// RelationVariable represents a relationship variable reference.
type RelationVariable struct {
	Name string
}

func (p *PropertyAccess) isExpression()   {}
func (l *Literal) isExpression()          {}
func (p *PropertyLookup) isExpression()   {}
func (c *ComparisonOp) isExpression()     {}
func (i *Identifier) isExpression()       {}
func (r *RelationVariable) isExpression() {}

// String returns the string representation of the relationship direction.
func (r RelDirection) String() string {
	return string(r)
}
