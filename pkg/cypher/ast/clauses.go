package ast

type MatchClause struct {
	Optional bool
	Pattern  *PatternExpr
	Where    *WhereExpr
	Return   *ReturnExpr
	Delete   *DeleteClause
	Start    Pos
	EndPos   Pos
}

func (c *MatchClause) Position() Pos { return c.Start }
func (c *MatchClause) End() Pos      { return c.EndPos }
func (c *MatchClause) String() string { return "MATCH" }

type CreateClause struct {
	Pattern *PatternExpr
	Start   Pos
	EndPos  Pos
}

func (c *CreateClause) Position() Pos { return c.Start }
func (c *CreateClause) End() Pos      { return c.EndPos }
func (c *CreateClause) String() string { return "CREATE" }

type MergeClause struct {
	Pattern    *PatternExpr
	OnCreate   []*SetItem
	OnMatch    []*SetItem
	Start      Pos
	EndPos     Pos
}

func (c *MergeClause) Position() Pos { return c.Start }
func (c *MergeClause) End() Pos      { return c.EndPos }
func (c *MergeClause) String() string { return "MERGE" }

type SetClause struct {
	Items []*SetItem
	Start Pos
	EndPos Pos
}

func (c *SetClause) Position() Pos { return c.Start }
func (c *SetClause) End() Pos      { return c.EndPos }
func (c *SetClause) String() string { return "SET" }

type DeleteClause struct {
	Detach bool
	Items  []Expr
	Start  Pos
	EndPos Pos
}

func (c *DeleteClause) Position() Pos { return c.Start }
func (c *DeleteClause) End() Pos      { return c.EndPos }
func (c *DeleteClause) String() string {
	if c.Detach {
		return "DETACH DELETE"
	}
	return "DELETE"
}

type RemoveClause struct {
	Items []*RemoveItemExpr
	Start Pos
	EndPos Pos
}

func (c *RemoveClause) Position() Pos { return c.Start }
func (c *RemoveClause) End() Pos      { return c.EndPos }
func (c *RemoveClause) String() string { return "REMOVE" }

type WhereClause struct {
	Expr   Expr
	Start  Pos
	EndPos Pos
}

func (c *WhereClause) Position() Pos { return c.Start }
func (c *WhereClause) End() Pos      { return c.EndPos }
func (c *WhereClause) String() string { return "WHERE" }

type WhereExpr struct {
	Expr   Expr
	Start  Pos
	EndPos Pos
}

func (c *WhereExpr) Position() Pos { return c.Start }
func (c *WhereExpr) End() Pos      { return c.EndPos }
func (c *WhereExpr) String() string { return "WHERE" }

type ReturnClause struct {
	Items    []*ReturnItemExpr
	Distinct bool
	OrderBy  *OrderByExpr
	Skip     Expr
	Limit    Expr
	Start    Pos
	EndPos   Pos
}

func (c *ReturnClause) Position() Pos { return c.Start }
func (c *ReturnClause) End() Pos      { return c.EndPos }
func (c *ReturnClause) String() string { return "RETURN" }

type ReturnExpr struct {
	Items    []*ReturnItemExpr
	Distinct bool
	OrderBy  *OrderByExpr
	Skip     Expr
	Limit    Expr
	Start    Pos
	EndPos   Pos
}

func (c *ReturnExpr) Position() Pos { return c.Start }
func (c *ReturnExpr) End() Pos      { return c.EndPos }
func (c *ReturnExpr) String() string { return "RETURN" }

type ReturnItemExpr struct {
	Expr   Expr
	Alias  string
	Start  Pos
	EndPos Pos
}

func (c *ReturnItemExpr) Position() Pos { return c.Start }
func (c *ReturnItemExpr) End() Pos      { return c.EndPos }
func (c *ReturnItemExpr) String() string {
	if c.Alias != "" {
		return c.Alias
	}
	return "ReturnItem"
}

type OrderByExpr struct {
	Items  []*OrderByItem
	Start  Pos
	EndPos Pos
}

func (c *OrderByExpr) Position() Pos { return c.Start }
func (c *OrderByExpr) End() Pos      { return c.EndPos }
func (c *OrderByExpr) String() string { return "ORDER BY" }

type OrderByItem struct {
	Expr       Expr
	Descending bool
	Ascending  bool
	Start      Pos
	EndPos     Pos
}

func (c *OrderByItem) Position() Pos { return c.Start }
func (c *OrderByItem) End() Pos      { return c.EndPos }
func (c *OrderByItem) String() string { return "OrderByItem" }

type UnwindClause struct {
	Expr   Expr
	Var    string
	Start  Pos
	EndPos Pos
}

func (c *UnwindClause) Position() Pos { return c.Start }
func (c *UnwindClause) End() Pos      { return c.EndPos }
func (c *UnwindClause) String() string { return "UNWIND" }

type UnionClause struct {
	All    bool
	Start  Pos
	EndPos Pos
}

func (c *UnionClause) Position() Pos { return c.Start }
func (c *UnionClause) End() Pos      { return c.EndPos }
func (c *UnionClause) String() string {
	if c.All {
		return "UNION ALL"
	}
	return "UNION"
}

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

func (c *WithClause) Position() Pos { return c.Start }
func (c *WithClause) End() Pos      { return c.EndPos }
func (c *WithClause) String() string { return "WITH" }
