package ast

type Ident struct {
	Name   string
	Start  Pos
	EndPos Pos
}

func (e *Ident) Position() Pos { return e.Start }
func (e *Ident) End() Pos      { return e.EndPos }
func (e *Ident) String() string { return e.Name }

type PropertyAccessExpr struct {
	Target   Expr
	Property string
	Start    Pos
	EndPos   Pos
}

func (e *PropertyAccessExpr) Position() Pos { return e.Start }
func (e *PropertyAccessExpr) End() Pos      { return e.EndPos }
func (e *PropertyAccessExpr) String() string {
	return e.Property
}

type BinaryExpr struct {
	Left     Expr
	Operator string
	Right    Expr
	Start    Pos
	EndPos   Pos
}

func (e *BinaryExpr) Position() Pos { return e.Start }
func (e *BinaryExpr) End() Pos      { return e.EndPos }
func (e *BinaryExpr) String() string { return e.Operator }

type UnaryExpr struct {
	Operator string
	Operand  Expr
	Start    Pos
	EndPos   Pos
}

func (e *UnaryExpr) Position() Pos { return e.Start }
func (e *UnaryExpr) End() Pos      { return e.EndPos }
func (e *UnaryExpr) String() string { return e.Operator }

type FuncCall struct {
	Name     string
	Args     []Expr
	Distinct bool
	Start    Pos
	EndPos   Pos
}

func (e *FuncCall) Position() Pos { return e.Start }
func (e *FuncCall) End() Pos      { return e.EndPos }
func (e *FuncCall) String() string { return e.Name }

type CaseExpr struct {
	Operand Expr
	Whens   []*WhenClause
	Else    Expr
	Start   Pos
	EndPos  Pos
}

func (e *CaseExpr) Position() Pos { return e.Start }
func (e *CaseExpr) End() Pos      { return e.EndPos }
func (e *CaseExpr) String() string { return "CASE" }

type ExistsExpr struct {
	Pattern *PatternExpr
	Expr    Expr
	Start   Pos
	EndPos  Pos
}

func (e *ExistsExpr) Position() Pos { return e.Start }
func (e *ExistsExpr) End() Pos      { return e.EndPos }
func (e *ExistsExpr) String() string { return "EXISTS" }

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

func (e *ListComprehension) Position() Pos { return e.Start }
func (e *ListComprehension) End() Pos      { return e.EndPos }
func (e *ListComprehension) String() string { return "ListComprehension" }

type InExpr struct {
	Left  Expr
	Right Expr
	Start Pos
	EndPos Pos
}

func (e *InExpr) Position() Pos { return e.Start }
func (e *InExpr) End() Pos      { return e.EndPos }
func (e *InExpr) String() string { return "IN" }

type IsNullExpr struct {
	Expr   Expr
	IsNot  bool
	Negate bool
	Start  Pos
	EndPos Pos
}

func (e *IsNullExpr) Position() Pos { return e.Start }
func (e *IsNullExpr) End() Pos      { return e.EndPos }
func (e *IsNullExpr) String() string {
	if e.IsNot || e.Negate {
		return "IS NOT NULL"
	}
	return "IS NULL"
}

type ListSliceExpr struct {
	List   Expr
	From   Expr
	To     Expr
	Start  Pos
	EndPos Pos
}

func (e *ListSliceExpr) Position() Pos { return e.Start }
func (e *ListSliceExpr) End() Pos      { return e.EndPos }
func (e *ListSliceExpr) String() string { return "ListSlice" }

type ListIndexExpr struct {
	List   Expr
	Index  Expr
	Start  Pos
	EndPos Pos
}

func (e *ListIndexExpr) Position() Pos { return e.Start }
func (e *ListIndexExpr) End() Pos      { return e.EndPos }
func (e *ListIndexExpr) String() string { return "ListIndex" }

type ListExpr struct {
	Elements []Expr
	Start    Pos
	EndPos   Pos
}

func (e *ListExpr) Position() Pos { return e.Start }
func (e *ListExpr) End() Pos      { return e.EndPos }
func (e *ListExpr) String() string { return "List" }

type MapExpr struct {
	Pairs  []*MapPair
	Start  Pos
	EndPos Pos
}

func (e *MapExpr) Position() Pos { return e.Start }
func (e *MapExpr) End() Pos      { return e.EndPos }
func (e *MapExpr) String() string { return "Map" }

type MapPair struct {
	Key    string
	Value  Expr
	Start  Pos
	EndPos Pos
}

func (p *MapPair) Position() Pos { return p.Start }
func (p *MapPair) End() Pos      { return p.EndPos }
func (p *MapPair) String() string { return p.Key }

type WhenClause struct {
	Condition Expr
	Result    Expr
	Start     Pos
	EndPos    Pos
}

func (c *WhenClause) Position() Pos { return c.Start }
func (c *WhenClause) End() Pos      { return c.EndPos }
func (c *WhenClause) String() string { return "WHEN" }
