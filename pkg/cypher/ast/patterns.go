package ast

// Direction represents the direction of a relationship in a pattern.
// It determines how the relationship connects nodes.
type Direction int

const (
	// DirectionBoth indicates a relationship in either direction (no arrow).
	// Example: (a)--(b)
	DirectionBoth Direction = iota

	// DirectionOutgoing indicates an outgoing relationship (right arrow).
	// Example: (a)-->(b)
	DirectionOutgoing

	// DirectionIncoming indicates an incoming relationship (left arrow).
	// Example: (a)<--(b)
	DirectionIncoming
)

// PatternExpr represents a complete pattern expression.
// It consists of one or more pattern parts that describe
// the graph structure to match or create.
//
// Example Cypher:
//
//	(n:Person)-[:KNOWS]->(m:Person), (m)-[:WORKS_AT]->(c:Company)
//
// Fields:
//   - Parts: The pattern parts (comma-separated in Cypher)
//   - Paths: Alternative path representations
//   - Start: Position in source where this pattern begins
//   - EndPos: Position in source where this pattern ends
type PatternExpr struct {
	Parts  []*PatternPart
	Paths  []*PathExpr
	Start  Pos
	EndPos Pos
}

// Position returns the starting position of this pattern in the source.
func (p *PatternExpr) Position() Pos { return p.Start }

// End returns the ending position of this pattern in the source.
func (p *PatternExpr) End() Pos { return p.EndPos }

// String returns "Pattern".
func (p *PatternExpr) String() string { return "Pattern" }

// PatternPart represents a single part of a pattern expression.
// It can be named with a variable and contains a path.
//
// Example Cypher:
//
//	p = (n:Person)-[:KNOWS]->(m:Person)
//
// Fields:
//   - Variable: Optional variable name for the entire pattern part
//   - Path: The path expression within this part
//   - Start: Position in source where this part begins
//   - EndPos: Position in source where this part ends
type PatternPart struct {
	Variable string
	Path     *PathExpr
	Start    Pos
	EndPos   Pos
}

// Position returns the starting position of this part in the source.
func (p *PatternPart) Position() Pos { return p.Start }

// End returns the ending position of this part in the source.
func (p *PatternPart) End() Pos { return p.EndPos }

// String returns the variable name or "PatternPart".
func (p *PatternPart) String() string {
	if p.Variable != "" {
		return p.Variable
	}
	return "PatternPart"
}

// patternNode marks PatternPart as a pattern node.
func (p *PatternPart) patternNode() {}

// PathExpr represents a path expression in a pattern.
// It describes a sequence of nodes connected by relationships.
//
// Example Cypher:
//
//	(n:Person)-[:KNOWS]->(m:Person)<-[:FRIENDS_WITH]-(o:Person)
//
// Fields:
//   - Nodes: The nodes in the path
//   - Relationships: The relationships connecting the nodes
//   - Start: Position in source where this path begins
//   - EndPos: Position in source where this path ends
type PathExpr struct {
	Nodes         []*NodePattern
	Relationships []*RelationPattern
	Start         Pos
	EndPos        Pos
}

// Position returns the starting position of this path in the source.
func (p *PathExpr) Position() Pos { return p.Start }

// End returns the ending position of this path in the source.
func (p *PathExpr) End() Pos { return p.EndPos }

// String returns "Path".
func (p *PathExpr) String() string { return "Path" }

// NodePattern represents a node pattern in a Cypher query.
// It describes a node with optional variable name, labels, and properties.
//
// Example Cypher:
//
//	(n:Person:Employee {name: 'Alice', age: 30})
//	(:Company)
//	(n)
//
// Fields:
//   - Variable: Optional variable name for the node
//   - Labels: List of labels (e.g., Person, Employee)
//   - Properties: Map of property names to expressions
//   - PropertyExpr: Alternative property expression
//   - Start: Position in source where this node begins
//   - EndPos: Position in source where this node ends
type NodePattern struct {
	Variable     string
	Labels       []string
	Properties   map[string]Expr
	PropertyExpr Expr
	Start        Pos
	EndPos       Pos
}

// Position returns the starting position of this node in the source.
func (n *NodePattern) Position() Pos { return n.Start }

// End returns the ending position of this node in the source.
func (n *NodePattern) End() Pos { return n.EndPos }

// String returns the variable name or "Node".
func (n *NodePattern) String() string {
	if n.Variable != "" {
		return n.Variable
	}
	return "Node"
}

// RelationPattern represents a relationship pattern in a Cypher query.
// It describes a relationship with optional variable name, types, direction,
// properties, and variable length constraints.
//
// Example Cypher:
//
//	-[r:KNOWS {since: 2020}]->
//	-[:FRIENDS_WITH|COLLEAGUE]->
//	-[*1..3]-> (variable length)
//	-->
//
// Fields:
//   - Variable: Optional variable name for the relationship
//   - Types: List of relationship types
//   - Direction: The relationship direction
//   - Properties: Map of property names to expressions
//   - MinHops: Minimum number of hops (for variable length)
//   - MaxHops: Maximum number of hops (for variable length)
//   - VariableLength: If true, this is a variable-length pattern
//   - RightArrow: If true, has a right arrow
//   - LeftArrow: If true, has a left arrow
//   - Start: Position in source where this relationship begins
//   - EndPos: Position in source where this relationship ends
type RelationPattern struct {
	Variable       string
	Types          []string
	Direction      Direction
	Properties     map[string]Expr
	MinHops        *int
	MaxHops        *int
	VariableLength bool
	RightArrow     bool
	LeftArrow      bool
	Start          Pos
	EndPos         Pos
}

// Position returns the starting position of this relationship in the source.
func (r *RelationPattern) Position() Pos { return r.Start }

// End returns the ending position of this relationship in the source.
func (r *RelationPattern) End() Pos { return r.EndPos }

// String returns the variable name or "Relation".
func (r *RelationPattern) String() string {
	if r.Variable != "" {
		return r.Variable
	}
	return "Relation"
}

// PatternExprItem represents an item within a pattern expression.
// It is similar to PatternPart but used in different contexts.
//
// Fields:
//   - Variable: Optional variable name
//   - Path: The path expression
//   - Start: Position in source where this item begins
//   - EndPos: Position in source where this item ends
type PatternExprItem struct {
	Variable string
	Path     *PathExpr
	Start    Pos
	EndPos   Pos
}

// Position returns the starting position of this item in the source.
func (p *PatternExprItem) Position() Pos { return p.Start }

// End returns the ending position of this item in the source.
func (p *PatternExprItem) End() Pos { return p.EndPos }

// String returns the variable name or "PatternItem".
func (p *PatternExprItem) String() string {
	if p.Variable != "" {
		return p.Variable
	}
	return "PatternItem"
}

// PatternComprehension represents a pattern comprehension expression.
// It creates a list by matching a pattern and projecting results.
//
// Example Cypher:
//
//	[(n)-[:KNOWS]->(m) | m.name]
//
// Fields:
//   - Pattern: The pattern to match
//   - Where: Optional filter condition
//   - Expr: The projection expression
//   - Start: Position in source where this comprehension begins
//   - EndPos: Position in source where this comprehension ends
type PatternComprehension struct {
	Pattern *PatternExpr
	Where   Expr
	Expr    Expr
	Start   Pos
	EndPos  Pos
}

// Position returns the starting position of this comprehension in the source.
func (p *PatternComprehension) Position() Pos { return p.Start }

// End returns the ending position of this comprehension in the source.
func (p *PatternComprehension) End() Pos { return p.EndPos }

// String returns "PatternComprehension".
func (p *PatternComprehension) String() string { return "PatternComprehension" }
