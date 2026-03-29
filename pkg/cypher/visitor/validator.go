package visitor

import (
	"fmt"
	"sync"

	"github.com/DotNetAge/gograph/pkg/cypher/ast"
)

type FunctionCategory int

const (
	FunctionAggregate FunctionCategory = iota
	FunctionList
	FunctionString
	FunctionMath
	FunctionTemporal
	FunctionCoalesce
	FunctionPath
)

type FunctionInfo struct {
	Name     string
	Category FunctionCategory
	MinArgs  int
	MaxArgs  int
}

type FunctionRegistry interface {
	IsValid(name string) bool
	GetInfo(name string) *FunctionInfo
	Register(info FunctionInfo)
}

type defaultRegistry struct {
	mu        sync.RWMutex
	functions map[string]*FunctionInfo
}

func newDefaultRegistry() *defaultRegistry {
	r := &defaultRegistry{
		functions: make(map[string]*FunctionInfo),
	}
	r.registerDefaults()
	return r
}

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

func (r *defaultRegistry) IsValid(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.functions[name]
	return ok
}

func (r *defaultRegistry) GetInfo(name string) *FunctionInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.functions[name]
}

func (r *defaultRegistry) Register(info FunctionInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.functions[info.Name] = &info
}

var globalRegistry = newDefaultRegistry()

func GlobalRegistry() FunctionRegistry {
	return globalRegistry
}

func RegisterFunction(info FunctionInfo) {
	globalRegistry.Register(info)
}

type Validator struct {
	registry FunctionRegistry
	errors   []error
}

type ValidatorOption func(*Validator)

func WithRegistry(r FunctionRegistry) ValidatorOption {
	return func(v *Validator) {
		v.registry = r
	}
}

func NewValidator(opts ...ValidatorOption) *Validator {
	v := &Validator{
		registry: globalRegistry,
	}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

func (v *Validator) Visit(node ast.Node) (ast.Visitor, error) {
	switch n := node.(type) {
	case *ast.Ident:
		v.validateIdent(n)
	case *ast.FuncCall:
		v.validateFuncCall(n)
	}
	return v, nil
}

func (v *Validator) validateIdent(ident *ast.Ident) {
	if ident.Name == "" {
		v.errors = append(v.errors, fmt.Errorf("empty identifier at %v", ident.Position()))
	}
}

func (v *Validator) validateFuncCall(call *ast.FuncCall) {
	if v.registry == nil {
		return
	}
	if !v.registry.IsValid(call.Name) {
		v.errors = append(v.errors, fmt.Errorf("unknown function: %s at %v", call.Name, call.Position()))
	}
}

func (v *Validator) Errors() []error {
	return v.errors
}

func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

type Printer struct {
	indent int
	output string
}

func NewPrinter() *Printer {
	return &Printer{}
}

func (p *Printer) Visit(node ast.Node) (ast.Visitor, error) {
	p.output += p.prefix() + node.String() + "\n"
	p.indent++
	return p, nil
}

func (p *Printer) prefix() string {
	result := ""
	for i := 0; i < p.indent; i++ {
		result += "  "
	}
	return result
}

func (p *Printer) String() string {
	return p.output
}

func Print(node ast.Node) string {
	printer := NewPrinter()
	_ = ast.Walk(printer, node)
	return printer.String()
}
