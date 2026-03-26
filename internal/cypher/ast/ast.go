package ast

type AST struct {
	Statements []Statement
}

type Statement struct {
	Clause Clause
}

type Clause interface {
	isClause()
}

type CreateClause struct {
	Pattern Pattern
}

type MatchClause struct {
	Pattern       Pattern
	Where         *WhereClause
	Return        *ReturnClause
	Delete        *DeleteClause
}

type SetClause struct {
	Assignments []Assignment
	Where      *WhereClause
}

type DeleteClause struct {
	Detach     bool
	Expressions []Expression
	Where      *WhereClause
}

type RemoveClause struct {
	Removals []RemoveItem
	Where    *WhereClause
}

func (c *CreateClause) isClause()  {}
func (c *MatchClause) isClause()  {}
func (c *SetClause) isClause()   {}
func (c *DeleteClause) isClause() {}
func (c *RemoveClause) isClause() {}

type Pattern struct {
	Elements []PatternElement
}

type PatternElement struct {
	Node     *NodePattern
	Relation *RelationPattern
}

type NodePattern struct {
	Variable  string
	Labels    []string
	Properties map[string]interface{}
}

type RelationPattern struct {
	Variable   string
	RelType   string
	Dir       RelDirection
	StartNode *NodePattern
	EndNode   *NodePattern
	Properties map[string]interface{}
}

type RelDirection string

const (
	RelDirOutgoing RelDirection = "->"
	RelDirIncoming RelDirection = "<-"
	RelDirBoth     RelDirection = "-"
)

type WhereClause struct {
	Expression Expression
}

type ReturnClause struct {
	Items []ReturnItem
}

type ReturnItem struct {
	Expression Expression
	Alias      string
}

type Assignment struct {
	Property  PropertyAccess
	Value    Expression
}

type RemoveItem struct {
	Type     RemoveItemType
	Label    string
	Property PropertyAccess
}

type RemoveItemType int

const (
	RemoveItemTypeLabel RemoveItemType = iota
	RemoveItemTypeProperty
)

type Expression interface {
	isExpression()
}

type PropertyAccess struct {
	Node     string
	Property string
}

type Literal struct {
	Value interface{}
}

type PropertyLookup struct {
	Node     string
	Property string
}

type ComparisonOp struct {
	Left     Expression
	Operator string
	Right    Expression
}

type Identifier struct {
	Name string
}

type RelationVariable struct {
	Name string
}

func (p *PropertyAccess) isExpression()   {}
func (l *Literal) isExpression()          {}
func (p *PropertyLookup) isExpression()   {}
func (c *ComparisonOp) isExpression()     {}
func (i *Identifier) isExpression()       {}
func (r *RelationVariable) isExpression() {}

func (r RelDirection) String() string {
	return string(r)
}
