// Package context provides parsing context management for the Cypher query parser.
// It tracks the current parsing state, variable scopes, and error accumulation
// during the parsing process.
//
// The context maintains a stack of scopes for variable binding and tracks
// the current clause being parsed (MATCH, CREATE, WHERE, etc.).
//
// Example:
//
//	ctx := context.New()
//	ctx.EnterScope()
//	ctx.BindVariable("n", context.VarNode, 0)
//	if binding := ctx.LookupVariable("n"); binding != nil {
//	    fmt.Printf("Variable n is a %v\n", binding.Type)
//	}
//	ctx.ExitScope()
package context

// ParseState represents the current state of the parser.
type ParseState int

// Parse state constants.
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

// String returns the string representation of the parse state.
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

// Context manages the parsing state and variable scopes.
type Context struct {
	scopes []*Scope    // Stack of scopes
	state  ParseState  // Current parse state
	errors []error     // Accumulated errors
}

// New creates a new parsing context with a global scope.
//
// Returns a new Context instance.
//
// Example:
//
//	ctx := context.New()
//	ctx.EnterScope()
func New() *Context {
	globalScope := NewScope(nil)
	return &Context{
		scopes: []*Scope{globalScope},
		state:  StateInitial,
	}
}

// EnterScope creates and enters a new scope.
// The new scope becomes the current scope.
//
// Example:
//
//	ctx.EnterScope()
//	defer ctx.ExitScope()
func (c *Context) EnterScope() {
	newScope := NewScope(c.CurrentScope())
	c.scopes = append(c.scopes, newScope)
}

// ExitScope exits the current scope and returns to the parent scope.
// The global scope cannot be exited.
func (c *Context) ExitScope() {
	if len(c.scopes) > 1 {
		c.scopes = c.scopes[:len(c.scopes)-1]
	}
}

// CurrentScope returns the current scope.
//
// Returns the current scope, or nil if no scopes exist.
func (c *Context) CurrentScope() *Scope {
	if len(c.scopes) == 0 {
		return nil
	}
	return c.scopes[len(c.scopes)-1]
}

// BindVariable binds a variable in the current scope.
//
// Parameters:
//   - name: The variable name
//   - typ: The variable type
//   - offset: The variable offset (for code generation)
func (c *Context) BindVariable(name string, typ VarType, offset int) {
	scope := c.CurrentScope()
	if scope != nil {
		scope.Bind(name, typ, offset)
	}
}

// LookupVariable looks up a variable in the current scope.
//
// Parameters:
//   - name: The variable name
//
// Returns the variable binding, or nil if not found.
func (c *Context) LookupVariable(name string) *VarBinding {
	scope := c.CurrentScope()
	if scope != nil {
		return scope.Lookup(name)
	}
	return nil
}

// IsVariableDeclared checks if a variable is declared in the current scope.
//
// Parameters:
//   - name: The variable name
//
// Returns true if the variable is declared.
func (c *Context) IsVariableDeclared(name string) bool {
	scope := c.CurrentScope()
	if scope != nil {
		return scope.IsBound(name)
	}
	return false
}

// MarkVariableUsed marks a variable as used in the current scope.
//
// Parameters:
//   - name: The variable name
func (c *Context) MarkVariableUsed(name string) {
	scope := c.CurrentScope()
	if scope != nil {
		scope.MarkUsed(name)
	}
}

// SetState sets the current parse state.
//
// Parameters:
//   - state: The new parse state
func (c *Context) SetState(state ParseState) {
	c.state = state
}

// GetState returns the current parse state.
//
// Returns the current parse state.
func (c *Context) GetState() ParseState {
	return c.state
}

// InClauseState returns true if the parser is currently in a clause state.
// This includes MATCH, CREATE, WHERE, RETURN, etc.
//
// Returns true if in a clause state.
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

// AddError adds an error to the context.
//
// Parameters:
//   - err: The error to add
func (c *Context) AddError(err error) {
	c.errors = append(c.errors, err)
}

// Errors returns all accumulated errors.
//
// Returns the slice of errors.
func (c *Context) Errors() []error {
	return c.errors
}

// HasErrors returns true if any errors have been accumulated.
//
// Returns true if there are errors.
func (c *Context) HasErrors() bool {
	return len(c.errors) > 0
}
