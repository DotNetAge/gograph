package context

// VarType represents the type of a variable in the Cypher query.
type VarType int

// Variable type constants.
const (
	VarUnknown VarType = iota
	VarNode
	VarRelationship
	VarPath
	VarMap
	VarList
	VarScalar
)

// String returns the string representation of the variable type.
func (v VarType) String() string {
	switch v {
	case VarNode:
		return "Node"
	case VarRelationship:
		return "Relationship"
	case VarPath:
		return "Path"
	case VarMap:
		return "Map"
	case VarList:
		return "List"
	case VarScalar:
		return "Scalar"
	default:
		return "Unknown"
	}
}

// VarBinding represents a variable binding in a scope.
type VarBinding struct {
	Name     string  // Variable name
	Type     VarType // Variable type
	Declared bool    // Whether the variable is declared
	Used     bool    // Whether the variable is used
	Offset   int     // Variable offset for code generation
}

// Scope represents a variable scope in the Cypher query.
// Scopes can be nested, with child scopes inheriting from parent scopes.
type Scope struct {
	parent    *Scope                // Parent scope (nil for global scope)
	variables map[string]*VarBinding // Variables in this scope
}

// NewScope creates a new scope with the given parent.
//
// Parameters:
//   - parent: The parent scope, or nil for a global scope
//
// Returns a new Scope instance.
//
// Example:
//
//	global := NewScope(nil)
//	local := NewScope(global)
func NewScope(parent *Scope) *Scope {
	return &Scope{
		parent:    parent,
		variables: make(map[string]*VarBinding),
	}
}

// Bind creates a new variable binding in this scope.
//
// Parameters:
//   - name: The variable name
//   - typ: The variable type
//   - offset: The variable offset
//
// Returns the created VarBinding.
//
// Example:
//
//	binding := scope.Bind("n", VarNode, 0)
func (s *Scope) Bind(name string, typ VarType, offset int) *VarBinding {
	binding := &VarBinding{
		Name:     name,
		Type:     typ,
		Declared: true,
		Offset:   offset,
	}
	s.variables[name] = binding
	return binding
}

// Lookup searches for a variable in this scope and parent scopes.
//
// Parameters:
//   - name: The variable name
//
// Returns the VarBinding if found, or nil if not found.
//
// Example:
//
//	if binding := scope.Lookup("n"); binding != nil {
//	    fmt.Printf("Found variable: %s\n", binding.Name)
//	}
func (s *Scope) Lookup(name string) *VarBinding {
	if binding, ok := s.variables[name]; ok {
		return binding
	}
	if s.parent != nil {
		return s.parent.Lookup(name)
	}
	return nil
}

// IsBound checks if a variable is bound in this scope or any parent scope.
//
// Parameters:
//   - name: The variable name
//
// Returns true if the variable is bound.
func (s *Scope) IsBound(name string) bool {
	return s.Lookup(name) != nil
}

// MarkUsed marks a variable as used in this scope.
// Note: This only marks variables in the current scope, not parent scopes.
//
// Parameters:
//   - name: The variable name
func (s *Scope) MarkUsed(name string) {
	if binding, ok := s.variables[name]; ok {
		binding.Used = true
	}
}
