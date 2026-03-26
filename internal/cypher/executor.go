package cypher

import (
	"context"

	"github.com/DotNetAge/gograph/internal/cypher/ast"
	"github.com/DotNetAge/gograph/internal/cypher/creators"
	"github.com/DotNetAge/gograph/internal/cypher/matchers"
	"github.com/DotNetAge/gograph/internal/cypher/modifiers"
	"github.com/DotNetAge/gograph/internal/graph"
	"github.com/DotNetAge/gograph/internal/storage"
	"github.com/DotNetAge/gograph/internal/tx"
)

type Executor struct {
	store         *storage.DB
	txMgr         *tx.Manager
	index         *graph.Index
	adj           *graph.AdjacencyList
	observability *Observability

	creator  *creators.Creator
	matcher  *matchers.Matcher
	modifier *modifiers.Modifier
}

type ExecutorOption func(*Executor)

func WithObservability(o *Observability) ExecutorOption {
	return func(e *Executor) {
		e.observability = o
	}
}

func NewExecutor(store *storage.DB, opts ...ExecutorOption) *Executor {
	idx := graph.NewIndex(store)
	matcher := matchers.NewMatcher(store, idx)
	e := &Executor{
		store:         store,
		txMgr:         tx.NewManager(store),
		index:         idx,
		adj:           graph.NewAdjacencyList(store),
		observability: NewObservability(),
		creator:       creators.NewCreator(store),
		matcher:       matcher,
		modifier:      modifiers.NewModifier(store, matcher),
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *Executor) Execute(ctx context.Context, a *ast.AST, params map[string]interface{}) (Result, error) {
	ctx, span := e.observability.Tracer.StartSpan(ctx, "Executor.Execute")
	if span != nil {
		defer span.End()
	}

	if len(a.Statements) == 0 {
		return Result{}, nil
	}

	t, err := e.txMgr.Begin(false)
	if err != nil {
		e.observability.Logger.Error("failed to begin transaction", "error", err)
		return Result{}, err
	}

	result := Result{}
	hasWrite := false
	varVars := make(map[string]interface{})

	for _, stmt := range a.Statements {
		var execErr error
		switch clause := stmt.Clause.(type) {
		case *ast.CreateClause:
			an, ar, err := e.creator.Execute(t, clause)
			execErr = err
			if execErr == nil {
				result.AddAffected(an, ar)
				hasWrite = true
			}
		case *ast.MatchClause:
			e.observability.Logger.Debug("processing MatchClause", "delete", clause.Delete != nil)
			rows, cols, err := e.matcher.Execute(clause, varVars, params)
			execErr = err
			if execErr == nil {
				result.Rows = rows
				result.Columns = cols
			}
			if clause.Delete != nil && execErr == nil {
				e.observability.Logger.Debug("calling executeDeleteWithTx")
				an, ar, err := e.modifier.ExecuteDelete(t, clause.Delete, varVars, params)
				execErr = err
				if execErr == nil {
					result.AddAffected(an, ar)
					hasWrite = true
				}
			}
		case *ast.SetClause:
			an, err := e.modifier.ExecuteSet(t, clause, varVars, params)
			execErr = err
			if execErr == nil {
				result.AddAffected(an, 0)
				hasWrite = true
			}
		case *ast.DeleteClause:
			an, ar, err := e.modifier.ExecuteDelete(t, clause, varVars, params)
			execErr = err
			if execErr == nil {
				result.AddAffected(an, ar)
				hasWrite = true
			}
		case *ast.RemoveClause:
			an, err := e.modifier.ExecuteRemove(t, clause, varVars, params)
			execErr = err
			if execErr == nil {
				result.AddAffected(an, 0)
				hasWrite = true
			}
		}

		if execErr != nil {
			t.Rollback()
			e.observability.Logger.Error("execution error", "error", execErr)
			return Result{}, execErr
		}
	}

	if hasWrite {
		if err := t.Commit(); err != nil {
			e.observability.Logger.Error("failed to commit transaction", "error", err)
			return Result{}, err
		}
	} else {
		t.Rollback()
	}

	return result, nil
}

func (e *Executor) ExecuteWithTx(t *tx.Transaction, a *ast.AST, params map[string]interface{}) (Result, error) {
	if len(a.Statements) == 0 {
		return Result{}, nil
	}

	result := Result{}
	varVars := make(map[string]interface{})

	for _, stmt := range a.Statements {
		switch clause := stmt.Clause.(type) {
		case *ast.CreateClause:
			an, ar, err := e.creator.Execute(t, clause)
			if err != nil {
				return Result{}, err
			}
			result.AddAffected(an, ar)
		case *ast.MatchClause:
			if clause.Delete != nil {
				an, ar, err := e.modifier.ExecuteDelete(t, clause.Delete, varVars, params)
				if err != nil {
					return Result{}, err
				}
				result.AddAffected(an, ar)
			}
			rows, cols, err := e.matcher.Execute(clause, varVars, params)
			if err != nil {
				return Result{}, err
			}
			result.Rows = rows
			result.Columns = cols
		case *ast.SetClause:
			an, err := e.modifier.ExecuteSet(t, clause, varVars, params)
			if err != nil {
				return Result{}, err
			}
			result.AddAffected(an, 0)
		case *ast.DeleteClause:
			an, ar, err := e.modifier.ExecuteDelete(t, clause, varVars, params)
			if err != nil {
				return Result{}, err
			}
			result.AddAffected(an, ar)
		case *ast.RemoveClause:
			an, err := e.modifier.ExecuteRemove(t, clause, varVars, params)
			if err != nil {
				return Result{}, err
			}
			result.AddAffected(an, 0)
		}
	}

	return result, nil
}

type Result struct {
	Rows          []map[string]interface{}
	Columns       []string
	AffectedNodes int
	AffectedRels  int
}

func (r *Result) AddAffected(nodes, rels int) {
	r.AffectedNodes += nodes
	r.AffectedRels += rels
}
