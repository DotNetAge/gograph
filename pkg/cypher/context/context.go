package context

type ParseState int

const (
	StateInitial ParseState = iota
	StateInMatch
	StateInCreate
	StateInMerge
	StateInWhere
	StateInReturn
	StateInWith
	StateInSet
	StateInDelete
	StateInRemove
	StateInPattern
	StateInExpression
	StateInListComprehension
	StateInPatternComprehension
)

func (s ParseState) String() string {
	switch s {
	case StateInMatch:
		return "InMatch"
	case StateInCreate:
		return "InCreate"
	case StateInMerge:
		return "InMerge"
	case StateInWhere:
		return "InWhere"
	case StateInReturn:
		return "InReturn"
	case StateInWith:
		return "InWith"
	case StateInSet:
		return "InSet"
	case StateInDelete:
		return "InDelete"
	case StateInRemove:
		return "InRemove"
	case StateInPattern:
		return "InPattern"
	case StateInExpression:
		return "InExpression"
	case StateInListComprehension:
		return "InListComprehension"
	case StateInPatternComprehension:
		return "InPatternComprehension"
	default:
		return "Initial"
	}
}

type Context struct {
	scopes []*Scope
	state  ParseState
	errors []error
}

func New() *Context {
	globalScope := NewScope(nil)
	return &Context{
		scopes: []*Scope{globalScope},
		state:  StateInitial,
	}
}

func (c *Context) EnterScope() {
	newScope := NewScope(c.CurrentScope())
	c.scopes = append(c.scopes, newScope)
}

func (c *Context) ExitScope() {
	if len(c.scopes) > 1 {
		c.scopes = c.scopes[:len(c.scopes)-1]
	}
}

func (c *Context) CurrentScope() *Scope {
	if len(c.scopes) == 0 {
		return nil
	}
	return c.scopes[len(c.scopes)-1]
}

func (c *Context) BindVariable(name string, typ VarType, offset int) {
	scope := c.CurrentScope()
	if scope != nil {
		scope.Bind(name, typ, offset)
	}
}

func (c *Context) LookupVariable(name string) *VarBinding {
	scope := c.CurrentScope()
	if scope != nil {
		return scope.Lookup(name)
	}
	return nil
}

func (c *Context) IsVariableDeclared(name string) bool {
	scope := c.CurrentScope()
	if scope != nil {
		return scope.IsBound(name)
	}
	return false
}

func (c *Context) MarkVariableUsed(name string) {
	scope := c.CurrentScope()
	if scope != nil {
		scope.MarkUsed(name)
	}
}

func (c *Context) SetState(state ParseState) {
	c.state = state
}

func (c *Context) GetState() ParseState {
	return c.state
}

func (c *Context) InClauseState() bool {
	switch c.state {
	case StateInMatch, StateInCreate, StateInMerge,
		StateInWhere, StateInReturn, StateInWith,
		StateInSet, StateInDelete, StateInRemove:
		return true
	default:
		return false
	}
}

func (c *Context) AddError(err error) {
	c.errors = append(c.errors, err)
}

func (c *Context) Errors() []error {
	return c.errors
}

func (c *Context) HasErrors() bool {
	return len(c.errors) > 0
}
