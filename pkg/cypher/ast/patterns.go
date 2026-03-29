package ast

type Direction int

const (
	DirectionBoth Direction = iota
	DirectionOutgoing
	DirectionIncoming
)

type PatternExpr struct {
	Parts  []*PatternPart
	Paths  []*PathExpr
	Start  Pos
	EndPos Pos
}

func (p *PatternExpr) Position() Pos { return p.Start }
func (p *PatternExpr) End() Pos      { return p.EndPos }
func (p *PatternExpr) String() string { return "Pattern" }

type PatternPart struct {
	Variable string
	Path     *PathExpr
	Start    Pos
	EndPos   Pos
}

func (p *PatternPart) Position() Pos { return p.Start }
func (p *PatternPart) End() Pos      { return p.EndPos }
func (p *PatternPart) String() string {
	if p.Variable != "" {
		return p.Variable
	}
	return "PatternPart"
}

func (p *PatternPart) patternNode() {}

type PathExpr struct {
	Nodes         []*NodePattern
	Relationships []*RelationPattern
	Start         Pos
	EndPos        Pos
}

func (p *PathExpr) Position() Pos { return p.Start }
func (p *PathExpr) End() Pos      { return p.EndPos }
func (p *PathExpr) String() string { return "Path" }

type NodePattern struct {
	Variable     string
	Labels       []string
	Properties   map[string]Expr
	PropertyExpr Expr
	Start        Pos
	EndPos       Pos
}

func (n *NodePattern) Position() Pos { return n.Start }
func (n *NodePattern) End() Pos      { return n.EndPos }
func (n *NodePattern) String() string {
	if n.Variable != "" {
		return n.Variable
	}
	return "Node"
}

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

func (r *RelationPattern) Position() Pos { return r.Start }
func (r *RelationPattern) End() Pos      { return r.EndPos }
func (r *RelationPattern) String() string {
	if r.Variable != "" {
		return r.Variable
	}
	return "Relation"
}

type PatternExprItem struct {
	Variable string
	Path     *PathExpr
	Start    Pos
	EndPos   Pos
}

func (p *PatternExprItem) Position() Pos { return p.Start }
func (p *PatternExprItem) End() Pos      { return p.EndPos }
func (p *PatternExprItem) String() string {
	if p.Variable != "" {
		return p.Variable
	}
	return "PatternItem"
}

type PatternComprehension struct {
	Pattern *PatternExpr
	Where   Expr
	Expr    Expr
	Start   Pos
	EndPos  Pos
}

func (p *PatternComprehension) Position() Pos { return p.Start }
func (p *PatternComprehension) End() Pos      { return p.EndPos }
func (p *PatternComprehension) String() string { return "PatternComprehension" }
