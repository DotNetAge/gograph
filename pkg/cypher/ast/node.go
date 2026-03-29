package ast

type Pos struct {
	Line   int
	Column int
	Offset int
}

type Node interface {
	Position() Pos
	End() Pos
	String() string
}

type Expr interface {
	Node
	exprNode()
}

type Stmt interface {
	Node
	stmtNode()
}

type Clause interface {
	Node
	clauseNode()
}

type Pattern interface {
	Node
	patternNode()
}

func (e *Ident) exprNode()               {}
func (e *PropertyAccessExpr) exprNode() {}
func (e *BinaryExpr) exprNode()         {}
func (e *UnaryExpr) exprNode()          {}
func (e *IntegerLit) exprNode()         {}
func (e *FloatLit) exprNode()           {}
func (e *StringLit) exprNode()          {}
func (e *BoolLit) exprNode()            {}
func (e *NullLit) exprNode()            {}
func (e *Param) exprNode()              {}
func (e *ParamExpr) exprNode()          {}
func (e *ListLit) exprNode()            {}
func (e *MapLit) exprNode()             {}
func (e *FuncCall) exprNode()           {}
func (e *CaseExpr) exprNode()           {}
func (e *ExistsExpr) exprNode()         {}
func (e *ListComprehension) exprNode()  {}
func (e *PatternComprehension) exprNode() {}
func (e *IsNullExpr) exprNode()         {}
func (e *InExpr) exprNode()             {}
func (e *ListSliceExpr) exprNode()      {}
func (e *ListIndexExpr) exprNode()      {}
func (e *ListExpr) exprNode()           {}
func (e *MapExpr) exprNode()            {}
func (e *MapPair) exprNode()            {}

func (s *MatchStmt) stmtNode()  {}
func (s *CreateStmt) stmtNode() {}
func (s *MergeStmt) stmtNode()  {}
func (s *SetStmt) stmtNode()    {}
func (s *DeleteStmt) stmtNode() {}
func (s *RemoveStmt) stmtNode() {}
func (s *WithStmt) stmtNode()   {}
func (s *UnwindStmt) stmtNode() {}
func (s *ReturnStmt) stmtNode() {}

func (c *MatchClause) clauseNode()  {}
func (c *CreateClause) clauseNode() {}
func (c *MergeClause) clauseNode()  {}
func (c *SetClause) clauseNode()    {}
func (c *DeleteClause) clauseNode() {}
func (c *RemoveClause) clauseNode() {}
func (c *WhereClause) clauseNode()  {}
func (c *ReturnClause) clauseNode() {}
func (c *WhereExpr) clauseNode()  {}
func (c *WhereExpr) exprNode()    {}
func (c *ReturnExpr) clauseNode()   {}
func (c *OrderByExpr) clauseNode()  {}
func (c *UnwindClause) clauseNode() {}
func (c *UnionClause) clauseNode()  {}
func (c *WithClause) clauseNode()   {}

func (p *PatternExpr) patternNode()   {}
func (p *PathExpr) patternNode()      {}
func (p *NodePattern) patternNode()   {}
func (p *RelationPattern) patternNode() {}
func (p *PatternExprItem) patternNode() {}
