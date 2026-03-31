package ast

// Pos represents a position in the source code.
// It is used for error reporting and source mapping.
//
// Fields:
//   - Line: The 1-based line number
//   - Column: The 1-based column number
//   - Offset: The 0-based byte offset from the start of the file
type Pos struct {
	Line   int
	Column int
	Offset int
}

// Node is the base interface for all AST nodes.
// Every node in the AST must implement this interface.
//
// Methods:
//   - Position(): Returns the starting position of the node
//   - End(): Returns the ending position of the node
//   - String(): Returns a string representation of the node
type Node interface {
	Position() Pos
	End() Pos
	String() string
}

// Expr is the interface for all expression nodes.
// Expressions can be evaluated to produce values.
//
// In addition to Node methods, Expr implementations must have
// an exprNode() method to distinguish them from other node types.
type Expr interface {
	Node
	exprNode()
}

// Stmt is the interface for all statement nodes.
// Statements represent executable actions in Cypher queries.
//
// In addition to Node methods, Stmt implementations must have
// a stmtNode() method to distinguish them from other node types.
type Stmt interface {
	Node
	stmtNode()
}

// Clause is the interface for all clause nodes.
// Clauses are components of statements (e.g., WHERE, RETURN).
//
// In addition to Node methods, Clause implementations must have
// a clauseNode() method to distinguish them from other node types.
type Clause interface {
	Node
	clauseNode()
}

// Pattern is the interface for all pattern nodes.
// Patterns describe graph structures to match or create.
//
// In addition to Node methods, Pattern implementations must have
// a patternNode() method to distinguish them from other node types.
type Pattern interface {
	Node
	patternNode()
}

// exprNode implementations mark types as expression nodes.
// These methods are empty and serve only as type markers.

func (e *Ident) exprNode()                {}
func (e *PropertyAccessExpr) exprNode()   {}
func (e *BinaryExpr) exprNode()           {}
func (e *UnaryExpr) exprNode()            {}
func (e *IntegerLit) exprNode()           {}
func (e *FloatLit) exprNode()             {}
func (e *StringLit) exprNode()            {}
func (e *BoolLit) exprNode()              {}
func (e *NullLit) exprNode()              {}
func (e *Param) exprNode()                {}
func (e *ParamExpr) exprNode()            {}
func (e *ListLit) exprNode()              {}
func (e *MapLit) exprNode()               {}
func (e *FuncCall) exprNode()             {}
func (e *CaseExpr) exprNode()             {}
func (e *ExistsExpr) exprNode()           {}
func (e *ListComprehension) exprNode()    {}
func (e *PatternComprehension) exprNode() {}
func (e *IsNullExpr) exprNode()           {}
func (e *InExpr) exprNode()               {}
func (e *ListSliceExpr) exprNode()        {}
func (e *ListIndexExpr) exprNode()        {}
func (e *ListExpr) exprNode()             {}
func (e *MapExpr) exprNode()              {}
func (e *MapPair) exprNode()              {}

// stmtNode implementations mark types as statement nodes.
// These methods are empty and serve only as type markers.

func (s *MatchStmt) stmtNode()  {}
func (s *CreateStmt) stmtNode() {}
func (s *MergeStmt) stmtNode()  {}
func (s *SetStmt) stmtNode()    {}
func (s *DeleteStmt) stmtNode() {}
func (s *RemoveStmt) stmtNode() {}
func (s *WithStmt) stmtNode()   {}
func (s *UnwindStmt) stmtNode() {}
func (s *ReturnStmt) stmtNode() {}

// clauseNode implementations mark types as clause nodes.
// These methods are empty and serve only as type markers.

func (c *MatchClause) clauseNode()  {}
func (c *CreateClause) clauseNode() {}
func (c *MergeClause) clauseNode()  {}
func (c *SetClause) clauseNode()    {}
func (c *DeleteClause) clauseNode() {}
func (c *RemoveClause) clauseNode() {}
func (c *WhereClause) clauseNode()  {}
func (c *ReturnClause) clauseNode() {}
func (c *WhereExpr) clauseNode()    {}
func (c *WhereExpr) exprNode()      {}
func (c *ReturnExpr) clauseNode()   {}
func (c *OrderByExpr) clauseNode()  {}
func (c *UnwindClause) clauseNode() {}
func (c *UnionClause) clauseNode()  {}
func (c *WithClause) clauseNode()   {}

// patternNode implementations mark types as pattern nodes.
// These methods are empty and serve only as type markers.

func (p *PatternExpr) patternNode()     {}
func (p *PathExpr) patternNode()        {}
func (p *NodePattern) patternNode()     {}
func (p *RelationPattern) patternNode() {}
func (p *PatternExprItem) patternNode() {}
