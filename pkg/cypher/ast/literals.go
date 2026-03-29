package ast

type IntegerLit struct {
	Value   int64
	Start   Pos
	EndPos  Pos
}

func (e *IntegerLit) Position() Pos { return e.Start }
func (e *IntegerLit) End() Pos      { return e.EndPos }
func (e *IntegerLit) String() string { return "Integer" }

type FloatLit struct {
	Value   float64
	Start   Pos
	EndPos  Pos
}

func (e *FloatLit) Position() Pos { return e.Start }
func (e *FloatLit) End() Pos      { return e.EndPos }
func (e *FloatLit) String() string { return "Float" }

type StringLit struct {
	Value   string
	Start   Pos
	EndPos  Pos
}

func (e *StringLit) Position() Pos { return e.Start }
func (e *StringLit) End() Pos      { return e.EndPos }
func (e *StringLit) String() string { return "String" }

type BoolLit struct {
	Value   bool
	Start   Pos
	EndPos  Pos
}

func (e *BoolLit) Position() Pos { return e.Start }
func (e *BoolLit) End() Pos      { return e.EndPos }
func (e *BoolLit) String() string { return "Bool" }

type NullLit struct {
	Start  Pos
	EndPos Pos
}

func (e *NullLit) Position() Pos { return e.Start }
func (e *NullLit) End() Pos      { return e.EndPos }
func (e *NullLit) String() string { return "NULL" }

type Param struct {
	Name   string
	Start  Pos
	EndPos Pos
}

func (e *Param) Position() Pos { return e.Start }
func (e *Param) End() Pos      { return e.EndPos }
func (e *Param) String() string { return "$" + e.Name }

type ParamExpr struct {
	Name   string
	Start  Pos
	EndPos Pos
}

func (e *ParamExpr) Position() Pos { return e.Start }
func (e *ParamExpr) End() Pos      { return e.EndPos }
func (e *ParamExpr) String() string { return "$" + e.Name }

type ListLit struct {
	Elements []Expr
	Start    Pos
	EndPos   Pos
}

func (e *ListLit) Position() Pos { return e.Start }
func (e *ListLit) End() Pos      { return e.EndPos }
func (e *ListLit) String() string { return "List" }

type MapLit struct {
	Properties map[string]Expr
	Start      Pos
	EndPos     Pos
}

func (e *MapLit) Position() Pos { return e.Start }
func (e *MapLit) End() Pos      { return e.EndPos }
func (e *MapLit) String() string { return "Map" }
