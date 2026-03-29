package context

type VarType int

const (
	VarUnknown VarType = iota
	VarNode
	VarRelationship
	VarPath
	VarMap
	VarList
	VarScalar
)

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

type VarBinding struct {
	Name     string
	Type     VarType
	Declared bool
	Used     bool
	Offset   int
}

type Scope struct {
	parent    *Scope
	variables map[string]*VarBinding
}

func NewScope(parent *Scope) *Scope {
	return &Scope{
		parent:    parent,
		variables: make(map[string]*VarBinding),
	}
}

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

func (s *Scope) Lookup(name string) *VarBinding {
	if binding, ok := s.variables[name]; ok {
		return binding
	}
	if s.parent != nil {
		return s.parent.Lookup(name)
	}
	return nil
}

func (s *Scope) IsBound(name string) bool {
	return s.Lookup(name) != nil
}

func (s *Scope) MarkUsed(name string) {
	if binding, ok := s.variables[name]; ok {
		binding.Used = true
	}
}
