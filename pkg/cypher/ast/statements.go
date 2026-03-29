package ast

type Query struct {
	Statements []Stmt
	Start      Pos
	EndPos     Pos
}

func (q *Query) Position() Pos { return q.Start }
func (q *Query) End() Pos      { return q.EndPos }
func (q *Query) String() string { return "Query" }

type MatchStmt struct {
	Optional bool
	Clauses  []Clause
	Start    Pos
	EndPos   Pos
}

func (s *MatchStmt) Position() Pos { return s.Start }
func (s *MatchStmt) End() Pos      { return s.EndPos }
func (s *MatchStmt) String() string { return "MATCH" }

type CreateStmt struct {
	Pattern *PatternExpr
	Clauses []Clause
	Start   Pos
	EndPos  Pos
}

func (s *CreateStmt) Position() Pos { return s.Start }
func (s *CreateStmt) End() Pos      { return s.EndPos }
func (s *CreateStmt) String() string { return "CREATE" }

type MergeStmt struct {
	Pattern  *PatternExpr
	Clause   *MergeClause
	OnCreate []*SetItem
	OnMatch  []*SetItem
	Start    Pos
	EndPos   Pos
}

func (s *MergeStmt) Position() Pos { return s.Start }
func (s *MergeStmt) End() Pos      { return s.EndPos }
func (s *MergeStmt) String() string { return "MERGE" }

type SetStmt struct {
	Items  []*SetItem
	Start  Pos
	EndPos Pos
}

func (s *SetStmt) Position() Pos { return s.Start }
func (s *SetStmt) End() Pos      { return s.EndPos }
func (s *SetStmt) String() string { return "SET" }

type SetItem struct {
	Target   Expr
	Value    Expr
	Operator string
	IsLabel  bool
	Start    Pos
	EndPos   Pos
}

func (s *SetItem) Position() Pos { return s.Start }
func (s *SetItem) End() Pos      { return s.EndPos }
func (s *SetItem) String() string { return "SetItem" }

type DeleteStmt struct {
	Detach bool
	Items  []Expr
	Start  Pos
	EndPos Pos
}

func (s *DeleteStmt) Position() Pos { return s.Start }
func (s *DeleteStmt) End() Pos      { return s.EndPos }
func (s *DeleteStmt) String() string {
	if s.Detach {
		return "DETACH DELETE"
	}
	return "DELETE"
}

type RemoveStmt struct {
	Items  []*RemoveItemExpr
	Start  Pos
	EndPos Pos
}

func (s *RemoveStmt) Position() Pos { return s.Start }
func (s *RemoveStmt) End() Pos      { return s.EndPos }
func (s *RemoveStmt) String() string { return "REMOVE" }

type RemoveItemExpr struct {
	Target  Expr
	IsLabel bool
	Label   string
	Start   Pos
	EndPos  Pos
}

func (r *RemoveItemExpr) Position() Pos { return r.Start }
func (r *RemoveItemExpr) End() Pos      { return r.EndPos }
func (r *RemoveItemExpr) String() string { return "RemoveItem" }

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

func (s *WithStmt) Position() Pos { return s.Start }
func (s *WithStmt) End() Pos      { return s.EndPos }
func (s *WithStmt) String() string { return "WITH" }

type UnwindStmt struct {
	List     Expr
	Variable string
	Return   *ReturnExpr
	Start    Pos
	EndPos   Pos
}

func (s *UnwindStmt) Position() Pos { return s.Start }
func (s *UnwindStmt) End() Pos      { return s.EndPos }
func (s *UnwindStmt) String() string { return "UNWIND" }

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

func (s *ReturnStmt) Position() Pos { return s.Start }
func (s *ReturnStmt) End() Pos      { return s.EndPos }
func (s *ReturnStmt) String() string { return "RETURN" }
