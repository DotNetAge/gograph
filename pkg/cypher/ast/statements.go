package ast

// Query represents a complete Cypher query.
// It can contain multiple statements separated by semicolons.
//
// Example Cypher:
//
//	MATCH (n) RETURN n;
//	CREATE (m:Person {name: 'Alice'})
//
// Fields:
//   - Statements: List of statements in the query
//   - Start: Position in source where this query begins
//   - EndPos: Position in source where this query ends
type Query struct {
	Statements []Stmt
	Start      Pos
	EndPos     Pos
}

// Position returns the starting position of this query in the source.
func (q *Query) Position() Pos { return q.Start }

// End returns the ending position of this query in the source.
func (q *Query) End() Pos { return q.EndPos }

// String returns "Query".
func (q *Query) String() string { return "Query" }

// MatchStmt represents a MATCH statement in a Cypher query.
// It matches patterns in the graph and can include WHERE, SET, RETURN, and DELETE clauses.
//
// Example Cypher:
//
//	MATCH (n:Person)-[:KNOWS]->(m:Person)
//	WHERE n.age > 25
//	SET m.visited = true
//	RETURN n, m
//
// Fields:
//   - Optional: If true, this is an OPTIONAL MATCH
//   - Clauses: List of clauses in the statement
//   - Start: Position in source where this statement begins
//   - EndPos: Position in source where this statement ends
type MatchStmt struct {
	Optional bool
	Clauses  []Clause
	Start    Pos
	EndPos   Pos
}

// Position returns the starting position of this statement in the source.
func (s *MatchStmt) Position() Pos { return s.Start }

// End returns the ending position of this statement in the source.
func (s *MatchStmt) End() Pos { return s.EndPos }

// String returns "MATCH".
func (s *MatchStmt) String() string { return "MATCH" }

// CreateStmt represents a CREATE statement in a Cypher query.
// It creates new nodes and relationships in the graph.
//
// Example Cypher:
//
//	CREATE (n:Person {name: 'Alice', age: 30})
//	CREATE (a)-[:KNOWS {since: 2020}]->(b)
//
// Fields:
//   - Pattern: The pattern describing what to create
//   - Clauses: Additional clauses (e.g., RETURN)
//   - Start: Position in source where this statement begins
//   - EndPos: Position in source where this statement ends
type CreateStmt struct {
	Pattern *PatternExpr
	Clauses []Clause
	Start   Pos
	EndPos  Pos
}

// Position returns the starting position of this statement in the source.
func (s *CreateStmt) Position() Pos { return s.Start }

// End returns the ending position of this statement in the source.
func (s *CreateStmt) End() Pos { return s.EndPos }

// String returns "CREATE".
func (s *CreateStmt) String() string { return "CREATE" }

// MergeStmt represents a MERGE statement in a Cypher query.
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
//   - Pattern: The pattern to merge
//   - Clause: The merge clause with ON CREATE/MATCH actions
//   - OnCreate: Actions to perform when creating new elements
//   - OnMatch: Actions to perform when matching existing elements
//   - Start: Position in source where this statement begins
//   - EndPos: Position in source where this statement ends
type MergeStmt struct {
	Pattern  *PatternExpr
	Clause   *MergeClause
	OnCreate []*SetItem
	OnMatch  []*SetItem
	Start    Pos
	EndPos   Pos
}

// Position returns the starting position of this statement in the source.
func (s *MergeStmt) Position() Pos { return s.Start }

// End returns the ending position of this statement in the source.
func (s *MergeStmt) End() Pos { return s.EndPos }

// String returns "MERGE".
func (s *MergeStmt) String() string { return "MERGE" }

// SetStmt represents a SET statement in a Cypher query.
// It sets properties on nodes and relationships.
//
// Example Cypher:
//
//	SET n.name = 'Alice'
//	SET n += {age: 30, city: 'NYC'}
//	SET n:Person:Employee
//
// Fields:
//   - Items: List of SET operations
//   - Start: Position in source where this statement begins
//   - EndPos: Position in source where this statement ends
type SetStmt struct {
	Items  []*SetItem
	Start  Pos
	EndPos Pos
}

// Position returns the starting position of this statement in the source.
func (s *SetStmt) Position() Pos { return s.Start }

// End returns the ending position of this statement in the source.
func (s *SetStmt) End() Pos { return s.EndPos }

// String returns "SET".
func (s *SetStmt) String() string { return "SET" }

// SetItem represents a single SET operation.
// It can set a property, add labels, or perform a map update.
//
// Example Cypher:
//
//	SET n.name = 'Alice'           // Property assignment
//	SET n += {age: 30}             // Map update
//	SET n:Person:Employee          // Label addition
//
// Fields:
//   - Target: The target expression (property or node)
//   - Value: The value to set
//   - Operator: The operator (= or +=)
//   - IsLabel: If true, this is a label operation
//   - Start: Position in source where this item begins
//   - EndPos: Position in source where this item ends
type SetItem struct {
	Target   Expr
	Value    Expr
	Operator string
	IsLabel  bool
	Start    Pos
	EndPos   Pos
}

// Position returns the starting position of this item in the source.
func (s *SetItem) Position() Pos { return s.Start }

// End returns the ending position of this item in the source.
func (s *SetItem) End() Pos { return s.EndPos }

// String returns "SetItem".
func (s *SetItem) String() string { return "SetItem" }

// DeleteStmt represents a DELETE statement in a Cypher query.
// It deletes nodes and relationships.
//
// Example Cypher:
//
//	DELETE r
//	DETACH DELETE n
//
// Fields:
//   - Detach: If true, delete nodes and their relationships
//   - Items: Expressions identifying what to delete
//   - Start: Position in source where this statement begins
//   - EndPos: Position in source where this statement ends
type DeleteStmt struct {
	Detach bool
	Items  []Expr
	Start  Pos
	EndPos Pos
}

// Position returns the starting position of this statement in the source.
func (s *DeleteStmt) Position() Pos { return s.Start }

// End returns the ending position of this statement in the source.
func (s *DeleteStmt) End() Pos { return s.EndPos }

// String returns "DELETE" or "DETACH DELETE".
func (s *DeleteStmt) String() string {
	if s.Detach {
		return "DETACH DELETE"
	}
	return "DELETE"
}

// RemoveStmt represents a REMOVE statement in a Cypher query.
// It removes properties and labels from nodes and relationships.
//
// Example Cypher:
//
//	REMOVE n.age
//	REMOVE n:Person
//
// Fields:
//   - Items: List of items to remove
//   - Start: Position in source where this statement begins
//   - EndPos: Position in source where this statement ends
type RemoveStmt struct {
	Items  []*RemoveItemExpr
	Start  Pos
	EndPos Pos
}

// Position returns the starting position of this statement in the source.
func (s *RemoveStmt) Position() Pos { return s.Start }

// End returns the ending position of this statement in the source.
func (s *RemoveStmt) End() Pos { return s.EndPos }

// String returns "REMOVE".
func (s *RemoveStmt) String() string { return "REMOVE" }

// RemoveItemExpr represents a single item in a REMOVE statement.
// It can remove a property or a label.
//
// Example Cypher:
//
//	REMOVE n.age      // Property
//	REMOVE n:Person   // Label
//
// Fields:
//   - Target: The target expression
//   - IsLabel: If true, this is a label removal
//   - Label: The label name (if IsLabel is true)
//   - Start: Position in source where this item begins
//   - EndPos: Position in source where this item ends
type RemoveItemExpr struct {
	Target  Expr
	IsLabel bool
	Label   string
	Start   Pos
	EndPos  Pos
}

// Position returns the starting position of this item in the source.
func (r *RemoveItemExpr) Position() Pos { return r.Start }

// End returns the ending position of this item in the source.
func (r *RemoveItemExpr) End() Pos { return r.EndPos }

// String returns "RemoveItem".
func (r *RemoveItemExpr) String() string { return "RemoveItem" }

// WithStmt represents a WITH statement in a Cypher query.
// It chains query parts, passing results to the next part.
//
// Example Cypher:
//
//	WITH n, count(*) AS cnt
//	ORDER BY cnt DESC
//	LIMIT 10
//
// Fields:
//   - Items: List of expressions to pass forward
//   - Distinct: If true, return only distinct rows
//   - OrderBy: Optional ORDER BY clause
//   - Skip: Optional number of rows to skip
//   - Limit: Optional maximum number of rows
//   - Where: Optional WHERE filter
//   - Start: Position in source where this statement begins
//   - EndPos: Position in source where this statement ends
type WithStmt struct {
	Items    []*ReturnItemExpr
	Distinct bool
	OrderBy  *OrderByExpr
	Skip     Expr
	Limit    Expr
	Where    Expr
	Start    Pos
	EndPos   Pos
}

// Position returns the starting position of this statement in the source.
func (s *WithStmt) Position() Pos { return s.Start }

// End returns the ending position of this statement in the source.
func (s *WithStmt) End() Pos { return s.EndPos }

// String returns "WITH".
func (s *WithStmt) String() string { return "WITH" }

// UnwindStmt represents an UNWIND statement in a Cypher query.
// It expands a list into a sequence of rows.
//
// Example Cypher:
//
//	UNWIND [1, 2, 3] AS num
//	UNWIND $names AS name
//
// Fields:
//   - List: The list expression to unwind
//   - Variable: The variable name for each element
//   - Return: Optional RETURN clause
//   - Start: Position in source where this statement begins
//   - EndPos: Position in source where this statement ends
type UnwindStmt struct {
	List     Expr
	Variable string
	Return   *ReturnExpr
	Start    Pos
	EndPos   Pos
}

// Position returns the starting position of this statement in the source.
func (s *UnwindStmt) Position() Pos { return s.Start }

// End returns the ending position of this statement in the source.
func (s *UnwindStmt) End() Pos { return s.EndPos }

// String returns "UNWIND".
func (s *UnwindStmt) String() string { return "UNWIND" }

// ReturnStmt represents a RETURN statement in a Cypher query.
// It specifies what data to return.
//
// Example Cypher:
//
//	RETURN n, m.name AS friendName
//	RETURN DISTINCT n.age ORDER BY n.age DESC LIMIT 10
//
// Fields:
//   - Distinct: If true, return only distinct rows
//   - Items: List of expressions to return
//   - OrderBy: Optional ORDER BY clause
//   - Skip: Optional number of rows to skip
//   - Limit: Optional maximum number of rows
//   - Return: The return expression
//   - Start: Position in source where this statement begins
//   - EndPos: Position in source where this statement ends
type ReturnStmt struct {
	Distinct bool
	Items    []*ReturnItemExpr
	OrderBy  *OrderByExpr
	Skip     Expr
	Limit    Expr
	Return   *ReturnExpr
	Start    Pos
	EndPos   Pos
}

// Position returns the starting position of this statement in the source.
func (s *ReturnStmt) Position() Pos { return s.Start }

// End returns the ending position of this statement in the source.
func (s *ReturnStmt) End() Pos { return s.EndPos }

// String returns "RETURN".
func (s *ReturnStmt) String() string { return "RETURN" }
