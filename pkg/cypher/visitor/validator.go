// Package visitor provides AST visitor implementations for validating
// and inspecting Cypher query syntax trees.
//
// It includes:
//   - Function validation
//   - AST printing/debugging
//   - Custom visitor support
//
// Example:
//
//	validator := visitor.NewValidator()
//	ast.Walk(validator, query)
//	if validator.HasErrors() {
//	    for _, err := range validator.Errors() {
//	        log.Println(err)
//	    }
//	}
package visitor

import (
	"fmt"
	"sync"

	"github.com/DotNetAge/gograph/pkg/cypher/ast"
)

// FunctionCategory represents the category of a Cypher function.
type FunctionCategory int

// Function category constants.
const (
	FunctionAggregate FunctionCategory = iota
	FunctionList
	FunctionString
	FunctionMath
	FunctionTemporal
	FunctionCoalesce
	FunctionPath
)

// FunctionInfo contains metadata about a Cypher function.
type FunctionInfo struct {
	Name     string           // Function name
	Category FunctionCategory // Function category
	MinArgs  int              // Minimum number of arguments (-1 for unlimited)
	MaxArgs  int              // Maximum number of arguments (-1 for unlimited)
}

// FunctionRegistry provides an interface for function registration and lookup.
type FunctionRegistry interface {
	// IsValid returns true if the function name is valid.
	IsValid(name string) bool

	// GetInfo returns function information for the given name.
	GetInfo(name string) *FunctionInfo

	// Register registers a new function.
	Register(info FunctionInfo)
}

// defaultRegistry is the default function registry implementation.
type defaultRegistry struct {
	mu        sync.RWMutex
	functions map[string]*FunctionInfo
}

// newDefaultRegistry creates a new default registry with built-in functions.
func newDefaultRegistry() *defaultRegistry {
	r := &defaultRegistry{
		functions: make(map[string]*FunctionInfo),
	}
	r.registerDefaults()
	return r
}

// registerDefaults registers all built-in Cypher functions.
func (r *defaultRegistry) registerDefaults() {
	aggregates := []FunctionInfo{
		{Name: "COUNT", Category: FunctionAggregate, MinArgs: 1, MaxArgs: 1},
		{Name: "SUM", Category: FunctionAggregate, MinArgs: 1, MaxArgs: 1},
		{Name: "AVG", Category: FunctionAggregate, MinArgs: 1, MaxArgs: 1},
		{Name: "MIN", Category: FunctionAggregate, MinArgs: 1, MaxArgs: 1},
		{Name: "MAX", Category: FunctionAggregate, MinArgs: 1, MaxArgs: 1},
		{Name: "COLLECT", Category: FunctionAggregate, MinArgs: 1, MaxArgs: 1},
	}
	listFuncs := []FunctionInfo{
		{Name: "HEAD", Category: FunctionList, MinArgs: 1, MaxArgs: 1},
		{Name: "LAST", Category: FunctionList, MinArgs: 1, MaxArgs: 1},
		{Name: "TAIL", Category: FunctionList, MinArgs: 1, MaxArgs: 1},
		{Name: "SIZE", Category: FunctionList, MinArgs: 1, MaxArgs: 1},
		{Name: "RANGE", Category: FunctionList, MinArgs: 2, MaxArgs: 3},
		{Name: "REVERSE", Category: FunctionList, MinArgs: 1, MaxArgs: 1},
	}
	stringFuncs := []FunctionInfo{
		{Name: "TOUPPER", Category: FunctionString, MinArgs: 1, MaxArgs: 1},
		{Name: "TOLOWER", Category: FunctionString, MinArgs: 1, MaxArgs: 1},
		{Name: "REPLACE", Category: FunctionString, MinArgs: 3, MaxArgs: 3},
		{Name: "SUBSTRING", Category: FunctionString, MinArgs: 2, MaxArgs: 3},
		{Name: "TRIM", Category: FunctionString, MinArgs: 1, MaxArgs: 1},
		{Name: "LTRIM", Category: FunctionString, MinArgs: 1, MaxArgs: 1},
		{Name: "RTRIM", Category: FunctionString, MinArgs: 1, MaxArgs: 1},
	}
	mathFuncs := []FunctionInfo{
		{Name: "ABS", Category: FunctionMath, MinArgs: 1, MaxArgs: 1},
		{Name: "CEIL", Category: FunctionMath, MinArgs: 1, MaxArgs: 1},
		{Name: "FLOOR", Category: FunctionMath, MinArgs: 1, MaxArgs: 1},
		{Name: "ROUND", Category: FunctionMath, MinArgs: 1, MaxArgs: 1},
		{Name: "RAND", Category: FunctionMath, MinArgs: 0, MaxArgs: 0},
	}
	temporalFuncs := []FunctionInfo{
		{Name: "DATE", Category: FunctionTemporal, MinArgs: 0, MaxArgs: 1},
		{Name: "DATETIME", Category: FunctionTemporal, MinArgs: 0, MaxArgs: 1},
		{Name: "TIME", Category: FunctionTemporal, MinArgs: 0, MaxArgs: 1},
		{Name: "TIMESTAMP", Category: FunctionTemporal, MinArgs: 0, MaxArgs: 0},
		{Name: "DURATION", Category: FunctionTemporal, MinArgs: 1, MaxArgs: 1},
		{Name: "LOCALDATETIME", Category: FunctionTemporal, MinArgs: 0, MaxArgs: 1},
		{Name: "LOCALTIME", Category: FunctionTemporal, MinArgs: 0, MaxArgs: 1},
	}
	coalesceFuncs := []FunctionInfo{
		{Name: "COALESCE", Category: FunctionCoalesce, MinArgs: 1, MaxArgs: -1},
		{Name: "NULLIF", Category: FunctionCoalesce, MinArgs: 2, MaxArgs: 2},
	}
	pathFuncs := []FunctionInfo{
		{Name: "ID", Category: FunctionPath, MinArgs: 1, MaxArgs: 1},
		{Name: "LABELS", Category: FunctionPath, MinArgs: 1, MaxArgs: 1},
		{Name: "TYPE", Category: FunctionPath, MinArgs: 1, MaxArgs: 1},
		{Name: "PROPERTIES", Category: FunctionPath, MinArgs: 1, MaxArgs: 1},
		{Name: "NODES", Category: FunctionPath, MinArgs: 1, MaxArgs: 1},
		{Name: "RELATIONSHIPS", Category: FunctionPath, MinArgs: 1, MaxArgs: 1},
		{Name: "LENGTH", Category: FunctionPath, MinArgs: 1, MaxArgs: 1},
	}

	for _, f := range aggregates {
		r.functions[f.Name] = &f
	}
	for _, f := range listFuncs {
		r.functions[f.Name] = &f
	}
	for _, f := range stringFuncs {
		r.functions[f.Name] = &f
	}
	for _, f := range mathFuncs {
		r.functions[f.Name] = &f
	}
	for _, f := range temporalFuncs {
		r.functions[f.Name] = &f
	}
	for _, f := range coalesceFuncs {
		r.functions[f.Name] = &f
	}
	for _, f := range pathFuncs {
		r.functions[f.Name] = &f
	}
}

// IsValid returns true if the function name is registered.
func (r *defaultRegistry) IsValid(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.functions[name]
	return ok
}

// GetInfo returns function information for the given name.
func (r *defaultRegistry) GetInfo(name string) *FunctionInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.functions[name]
}

// Register registers a new function in the registry.
func (r *defaultRegistry) Register(info FunctionInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.functions[info.Name] = &info
}

// globalRegistry is the global function registry instance.
var globalRegistry = newDefaultRegistry()

// GlobalRegistry returns the global function registry.
//
// Returns the global FunctionRegistry instance.
func GlobalRegistry() FunctionRegistry {
	return globalRegistry
}

// RegisterFunction registers a function in the global registry.
//
// Parameters:
//   - info: The function information to register
//
// Example:
//
//	visitor.RegisterFunction(visitor.FunctionInfo{
//	    Name: "CUSTOM", Category: visitor.FunctionMath, MinArgs: 1, MaxArgs: 1,
//	})
func RegisterFunction(info FunctionInfo) {
	globalRegistry.Register(info)
}

// Validator validates AST nodes for semantic correctness.
type Validator struct {
	registry FunctionRegistry // Function registry for validation
	errors   []error          // Accumulated errors
}

// ValidatorOption configures the Validator.
type ValidatorOption func(*Validator)

// WithRegistry sets a custom function registry.
//
// Parameters:
//   - r: The function registry to use
//
// Returns a ValidatorOption.
func WithRegistry(r FunctionRegistry) ValidatorOption {
	return func(v *Validator) {
		v.registry = r
	}
}

// NewValidator creates a new AST validator.
//
// Parameters:
//   - opts: Optional configuration options
//
// Returns a new Validator instance.
//
// Example:
//
//	validator := visitor.NewValidator()
//	ast.Walk(validator, query)
func NewValidator(opts ...ValidatorOption) *Validator {
	v := &Validator{
		registry: globalRegistry,
	}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

// Visit implements the ast.Visitor interface.
// It validates the given node and returns the visitor for child nodes.
//
// Parameters:
//   - node: The AST node to visit
//
// Returns the visitor for child nodes and any error encountered.
func (v *Validator) Visit(node ast.Node) (ast.Visitor, error) {
	switch n := node.(type) {
	case *ast.Ident:
		v.validateIdent(n)
	case *ast.FuncCall:
		v.validateFuncCall(n)
	}
	return v, nil
}

// validateIdent validates an identifier node.
func (v *Validator) validateIdent(ident *ast.Ident) {
	if ident.Name == "" {
		v.errors = append(v.errors, fmt.Errorf("empty identifier at %v", ident.Position()))
	}
}

// validateFuncCall validates a function call node.
func (v *Validator) validateFuncCall(call *ast.FuncCall) {
	if v.registry == nil {
		return
	}
	if !v.registry.IsValid(call.Name) {
		v.errors = append(v.errors, fmt.Errorf("unknown function: %s at %v", call.Name, call.Position()))
	}
}

// Errors returns all validation errors.
//
// Returns the slice of errors.
func (v *Validator) Errors() []error {
	return v.errors
}

// HasErrors returns true if any validation errors were found.
//
// Returns true if there are errors.
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// Printer prints the AST structure for debugging.
type Printer struct {
	indent int    // Current indentation level
	output string // Accumulated output
}

// NewPrinter creates a new AST printer.
//
// Returns a new Printer instance.
//
// Example:
//
//	printer := visitor.NewPrinter()
//	ast.Walk(printer, query)
//	fmt.Println(printer.String())
func NewPrinter() *Printer {
	return &Printer{}
}

// Visit implements the ast.Visitor interface.
// It prints the node and returns the visitor for child nodes.
//
// Parameters:
//   - node: The AST node to visit
//
// Returns the visitor for child nodes and any error encountered.
func (p *Printer) Visit(node ast.Node) (ast.Visitor, error) {
	p.output += p.prefix() + node.String() + "\n"
	p.indent++
	return p, nil
}

// prefix returns the indentation prefix for the current level.
func (p *Printer) prefix() string {
	result := ""
	for i := 0; i < p.indent; i++ {
		result += "  "
	}
	return result
}

// String returns the printed AST as a string.
//
// Returns the formatted AST string.
func (p *Printer) String() string {
	return p.output
}

// Print prints the AST structure of a node.
//
// Parameters:
//   - node: The AST node to print
//
// Returns the formatted AST string.
//
// Example:
//
//	fmt.Println(visitor.Print(query))
func Print(node ast.Node) string {
	printer := NewPrinter()
	_ = ast.Walk(printer, node)
	return printer.String()
}
