// Package ast defines the Abstract Syntax Tree (AST) nodes for Cypher queries.
// It provides type definitions for all Cypher language constructs including
// statements, clauses, expressions, patterns, and literals.
//
// The AST is organized into several categories:
//
//   - Node: The base interface for all AST nodes
//   - Stmt: Statements (Query, MatchStmt, CreateStmt, etc.)
//   - Clause: Clauses (MatchClause, WhereClause, ReturnClause, etc.)
//   - Expr: Expressions (BinaryExpr, FuncCall, Ident, etc.)
//   - Pattern: Patterns (PatternExpr, PathExpr, NodePattern, etc.)
//
// Each node type implements the Node interface with Position(), End(), and String() methods.
// Additional interfaces (Stmt, Clause, Expr, Pattern) are used to categorize nodes.
//
// Example:
//
//	query := &ast.Query{
//	    Statements: []ast.Stmt{
//	        &ast.MatchStmt{
//	            Clauses: []ast.Clause{
//	                &ast.MatchClause{
//	                    Pattern: pattern,
//	                },
//	            },
//	        },
//	    },
//	}
package ast

// MatchClause represents a MATCH clause in a Cypher query.
// It matches patterns in the graph and can optionally be an OPTIONAL MATCH.
//
// Example Cypher:
//
//	MATCH (n:Person)-[:KNOWS]->(m:Person)
//	OPTIONAL MATCH (n)-[:WORKS_AT]->(c:Company)
//
// Fields:
//   - Optional: If true, this is an OPTIONAL MATCH (returns null for non-matches)
//   - Pattern: The pattern to match in the graph
//   - Where: Optional WHERE filter expression
//   - Return: Optional RETURN clause
//   - Delete: Optional DELETE clause
//   - Start: Position in source where this clause begins
//   - EndPos: Position in source where this clause ends
type MatchClause struct {
	Optional bool
	Pattern  *PatternExpr
	Where    *WhereExpr
	Return   *ReturnExpr
	Delete   *DeleteClause
	Start    Pos
	EndPos   Pos
}

// Position returns the starting position of this clause in the source.
func (c *MatchClause) Position() Pos { return c.Start }

// End returns the ending position of this clause in the source.
func (c *MatchClause) End() Pos { return c.EndPos }

// String returns a string representation of this clause.
func (c *MatchClause) String() string { return "MATCH" }

// CreateClause represents a CREATE clause in a Cypher query.
// It creates new nodes and relationships in the graph.
//
// Example Cypher:
//
//	CREATE (n:Person {name: 'Alice'})
//	CREATE (a)-[:KNOWS]->(b)
//
// Fields:
//   - Pattern: The pattern describing what to create
//   - Start: Position in source where this clause begins
//   - EndPos: Position in source where this clause ends
type CreateClause struct {
	Pattern *PatternExpr
	Start   Pos
	EndPos  Pos
}

// Position returns the starting position of this clause in the source.
func (c *CreateClause) Position() Pos { return c.Start }

// End returns the ending position of this clause in the source.
func (c *CreateClause) End() Pos { return c.EndPos }

// String returns a string representation of this clause.
func (c *CreateClause) String() string { return "CREATE" }

// MergeClause represents a MERGE clause in a Cypher query.
// It ensures a pattern exists in the graph, creating it if necessary.
// Supports ON CREATE and ON MATCH actions.
//
// Example Cypher:
//
//	MERGE (n:Person {name: 'Alice'})
//	ON CREATE SET n.created = timestamp()
//	ON MATCH SET n.lastSeen = timestamp()
//
// Fields:
//   - Pattern: The pattern to merge (create if doesn't exist)
//   - OnCreate: Actions to perform when creating new elements
//   - OnMatch: Actions to perform when matching existing elements
//   - Start: Position in source where this clause begins
//   - EndPos: Position in source where this clause ends
type MergeClause struct {
	Pattern  *PatternExpr
	OnCreate []*SetItem
	OnMatch  []*SetItem
	Start    Pos
	EndPos   Pos
}

// Position returns the starting position of this clause in the source.
func (c *MergeClause) Position() Pos { return c.Start }

// End returns the ending position of this clause in the source.
func (c *MergeClause) End() Pos { return c.EndPos }

// String returns a string representation of this clause.
func (c *MergeClause) String() string { return "MERGE" }

// SetClause represents a SET clause in a Cypher query.
// It sets properties on nodes and relationships.
//
// Example Cypher:
//
//	SET n.name = 'Alice'
//	SET n += {age: 30, city: 'NYC'}
//	SET n:Person:Employee
//
// Fields:
//   - Items: List of SET operations to perform
//   - Start: Position in source where this clause begins
//   - EndPos: Position in source where this clause ends
type SetClause struct {
	Items  []*SetItem
	Start  Pos
	EndPos Pos
}

// Position returns the starting position of this clause in the source.
func (c *SetClause) Position() Pos { return c.Start }

// End returns the ending position of this clause in the source.
func (c *SetClause) End() Pos { return c.EndPos }

// String returns a string representation of this clause.
func (c *SetClause) String() string { return "SET" }

// DeleteClause represents a DELETE clause in a Cypher query.
// It deletes nodes and relationships. Use Detach to delete nodes
// without requiring deletion of their relationships first.
//
// Example Cypher:
//
//	DELETE r
//	DETACH DELETE n
//
// Fields:
//   - Detach: If true, delete nodes and their relationships
//   - Items: Expressions identifying what to delete
//   - Start: Position in source where this clause begins
//   - EndPos: Position in source where this clause ends
type DeleteClause struct {
	Detach bool
	Items  []Expr
	Start  Pos
	EndPos Pos
}

// Position returns the starting position of this clause in the source.
func (c *DeleteClause) Position() Pos { return c.Start }

// End returns the ending position of this clause in the source.
func (c *DeleteClause) End() Pos { return c.EndPos }

// String returns a string representation of this clause.
func (c *DeleteClause) String() string {
	if c.Detach {
		return "DETACH DELETE"
	}
	return "DELETE"
}

// RemoveClause represents a REMOVE clause in a Cypher query.
// It removes properties and labels from nodes and relationships.
//
// Example Cypher:
//
//	REMOVE n.age
//	REMOVE n:Person
//
// Fields:
//   - Items: List of items to remove
//   - Start: Position in source where this clause begins
//   - EndPos: Position in source where this clause ends
type RemoveClause struct {
	Items  []*RemoveItemExpr
	Start  Pos
	EndPos Pos
}

// Position returns the starting position of this clause in the source.
func (c *RemoveClause) Position() Pos { return c.Start }

// End returns the ending position of this clause in the source.
func (c *RemoveClause) End() Pos { return c.EndPos }

// String returns a string representation of this clause.
func (c *RemoveClause) String() string { return "REMOVE" }

// WhereClause represents a WHERE clause in a Cypher query.
// It filters results based on a boolean expression.
//
// Example Cypher:
//
//	WHERE n.age > 25 AND n.name STARTS WITH 'A'
//
// Fields:
//   - Expr: The boolean expression to evaluate
//   - Start: Position in source where this clause begins
//   - EndPos: Position in source where this clause ends
type WhereClause struct {
	Expr   Expr
	Start  Pos
	EndPos Pos
}

// Position returns the starting position of this clause in the source.
func (c *WhereClause) Position() Pos { return c.Start }

// End returns the ending position of this clause in the source.
func (c *WhereClause) End() Pos { return c.EndPos }

// String returns a string representation of this clause.
func (c *WhereClause) String() string { return "WHERE" }

// WhereExpr represents a WHERE expression in a Cypher query.
// It is similar to WhereClause but used in different contexts.
//
// Fields:
//   - Expr: The boolean expression to evaluate
//   - Start: Position in source where this expression begins
//   - EndPos: Position in source where this expression ends
type WhereExpr struct {
	Expr   Expr
	Start  Pos
	EndPos Pos
}

// Position returns the starting position of this expression in the source.
func (c *WhereExpr) Position() Pos { return c.Start }

// End returns the ending position of this expression in the source.
func (c *WhereExpr) End() Pos { return c.EndPos }

// String returns a string representation of this expression.
func (c *WhereExpr) String() string { return "WHERE" }

// ReturnClause represents a RETURN clause in a Cypher query.
// It specifies what data to return from the query.
//
// Example Cypher:
//
//	RETURN n, m.name AS friendName
//	RETURN DISTINCT n.age ORDER BY n.age DESC LIMIT 10
//
// Fields:
//   - Items: List of expressions to return
//   - Distinct: If true, return only distinct rows
//   - OrderBy: Optional ORDER BY clause
//   - Skip: Optional number of rows to skip
//   - Limit: Optional maximum number of rows to return
//   - Start: Position in source where this clause begins
//   - EndPos: Position in source where this clause ends
type ReturnClause struct {
	Items    []*ReturnItemExpr
	Distinct bool
	OrderBy  *OrderByExpr
	Skip     Expr
	Limit    Expr
	Start    Pos
	EndPos   Pos
}

// Position returns the starting position of this clause in the source.
func (c *ReturnClause) Position() Pos { return c.Start }

// End returns the ending position of this clause in the source.
func (c *ReturnClause) End() Pos { return c.EndPos }

// String returns a string representation of this clause.
func (c *ReturnClause) String() string { return "RETURN" }

// ReturnExpr represents a RETURN expression in a Cypher query.
// It is functionally equivalent to ReturnClause.
//
// Fields:
//   - Items: List of expressions to return
//   - Distinct: If true, return only distinct rows
//   - OrderBy: Optional ORDER BY clause
//   - Skip: Optional number of rows to skip
//   - Limit: Optional maximum number of rows to return
//   - Start: Position in source where this expression begins
//   - EndPos: Position in source where this expression ends
type ReturnExpr struct {
	Items    []*ReturnItemExpr
	Distinct bool
	OrderBy  *OrderByExpr
	Skip     Expr
	Limit    Expr
	Start    Pos
	EndPos   Pos
}

// Position returns the starting position of this expression in the source.
func (c *ReturnExpr) Position() Pos { return c.Start }

// End returns the ending position of this expression in the source.
func (c *ReturnExpr) End() Pos { return c.EndPos }

// String returns a string representation of this expression.
func (c *ReturnExpr) String() string { return "RETURN" }

// ReturnItemExpr represents a single item in a RETURN clause.
// It can have an optional alias (AS clause).
//
// Example Cypher:
//
//	RETURN n.name AS personName
//
// Fields:
//   - Expr: The expression to return
//   - Alias: Optional alias name (empty if no AS clause)
//   - Start: Position in source where this item begins
//   - EndPos: Position in source where this item ends
type ReturnItemExpr struct {
	Expr   Expr
	Alias  string
	Start  Pos
	EndPos Pos
}

// Position returns the starting position of this item in the source.
func (c *ReturnItemExpr) Position() Pos { return c.Start }

// End returns the ending position of this item in the source.
func (c *ReturnItemExpr) End() Pos { return c.EndPos }

// String returns a string representation of this item.
func (c *ReturnItemExpr) String() string {
	if c.Alias != "" {
		return c.Alias
	}
	return "ReturnItem"
}

// OrderByExpr represents an ORDER BY clause in a Cypher query.
// It specifies how to sort the results.
//
// Example Cypher:
//
//	ORDER BY n.age DESC, n.name ASC
//
// Fields:
//   - Items: List of sort specifications
//   - Start: Position in source where this clause begins
//   - EndPos: Position in source where this clause ends
type OrderByExpr struct {
	Items  []*OrderByItem
	Start  Pos
	EndPos Pos
}

// Position returns the starting position of this clause in the source.
func (c *OrderByExpr) Position() Pos { return c.Start }

// End returns the ending position of this clause in the source.
func (c *OrderByExpr) End() Pos { return c.EndPos }

// String returns a string representation of this clause.
func (c *OrderByExpr) String() string { return "ORDER BY" }

// OrderByItem represents a single sort specification in ORDER BY.
//
// Example Cypher:
//
//	ORDER BY n.age DESC
//
// Fields:
//   - Expr: The expression to sort by
//   - Descending: If true, sort in descending order
//   - Ascending: If true, sort in ascending order
//   - Start: Position in source where this item begins
//   - EndPos: Position in source where this item ends
type OrderByItem struct {
	Expr       Expr
	Descending bool
	Ascending  bool
	Start      Pos
	EndPos     Pos
}

// Position returns the starting position of this item in the source.
func (c *OrderByItem) Position() Pos { return c.Start }

// End returns the ending position of this item in the source.
func (c *OrderByItem) End() Pos { return c.EndPos }

// String returns a string representation of this item.
func (c *OrderByItem) String() string { return "OrderByItem" }

// UnwindClause represents an UNWIND clause in a Cypher query.
// It expands a list into a sequence of rows.
//
// Example Cypher:
//
//	UNWIND [1, 2, 3] AS num
//	UNWIND $names AS name
//
// Fields:
//   - Expr: The list expression to unwind
//   - Var: The variable name for each element
//   - Start: Position in source where this clause begins
//   - EndPos: Position in source where this clause ends
type UnwindClause struct {
	Expr   Expr
	Var    string
	Start  Pos
	EndPos Pos
}

// Position returns the starting position of this clause in the source.
func (c *UnwindClause) Position() Pos { return c.Start }

// End returns the ending position of this clause in the source.
func (c *UnwindClause) End() Pos { return c.EndPos }

// String returns a string representation of this clause.
func (c *UnwindClause) String() string { return "UNWIND" }

// UnionClause represents a UNION clause in a Cypher query.
// It combines results from multiple queries.
//
// Example Cypher:
//
//	MATCH (n) RETURN n.name
//	UNION
//	MATCH (m) RETURN m.name
//
//	MATCH (n) RETURN n.name
//	UNION ALL
//	MATCH (m) RETURN m.name
//
// Fields:
//   - All: If true, this is UNION ALL (keeps duplicates)
//   - Start: Position in source where this clause begins
//   - EndPos: Position in source where this clause ends
type UnionClause struct {
	All    bool
	Start  Pos
	EndPos Pos
}

// Position returns the starting position of this clause in the source.
func (c *UnionClause) Position() Pos { return c.Start }

// End returns the ending position of this clause in the source.
func (c *UnionClause) End() Pos { return c.EndPos }

// String returns a string representation of this clause.
func (c *UnionClause) String() string {
	if c.All {
		return "UNION ALL"
	}
	return "UNION"
}

// WithClause represents a WITH clause in a Cypher query.
// It chains query parts, passing results to the next part.
//
// Example Cypher:
//
//	WITH n, count(*) AS cnt ORDER BY cnt DESC LIMIT 10
//
// Fields:
//   - Items: List of expressions to pass forward
//   - Distinct: If true, return only distinct rows
//   - OrderBy: Optional ORDER BY clause
//   - Skip: Optional number of rows to skip
//   - Limit: Optional maximum number of rows to return
//   - Where: Optional WHERE filter
//   - Start: Position in source where this clause begins
//   - EndPos: Position in source where this clause ends
type WithClause struct {
	Items    []*ReturnItemExpr
	Distinct bool
	OrderBy  *OrderByExpr
	Skip     Expr
	Limit    Expr
	Where    Expr
	Start    Pos
	EndPos   Pos
}

// Position returns the starting position of this clause in the source.
func (c *WithClause) Position() Pos { return c.Start }

// End returns the ending position of this clause in the source.
func (c *WithClause) End() Pos { return c.EndPos }

// String returns a string representation of this clause.
func (c *WithClause) String() string { return "WITH" }
