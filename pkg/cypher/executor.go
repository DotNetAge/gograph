// Package cypher provides Cypher query parsing and execution capabilities for gograph.
// It includes a parser for Cypher syntax and an executor that coordinates the various
// clause handlers to process queries.
package cypher

import (
	"context"

	"github.com/DotNetAge/gograph/pkg/cypher/ast"
	"github.com/DotNetAge/gograph/pkg/cypher/creators"
	"github.com/DotNetAge/gograph/pkg/cypher/matchers"
	"github.com/DotNetAge/gograph/pkg/cypher/modifiers"
	"github.com/DotNetAge/gograph/pkg/graph"
	"github.com/DotNetAge/gograph/pkg/storage"
	"github.com/DotNetAge/gograph/pkg/tx"
)

// Executor coordinates the execution of Cypher queries by delegating to specialized
// clause handlers for CREATE, MATCH, SET, DELETE, and REMOVE operations.
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

// ExecutorOption configures optional parameters for the executor.
type ExecutorOption func(*Executor)

// WithObservability sets the observability handler for tracing and metrics.
func WithObservability(o *Observability) ExecutorOption {
	return func(e *Executor) {
		e.observability = o
	}
}

// NewExecutor creates a new Executor instance that operates on the given storage.
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

// Execute runs a Cypher query and returns the result.
// It automatically manages transactions, committing write operations and
// rolling back read-only operations.
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
	defer t.Rollback()

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
				if len(result.Rows) > 0 {
					for _, row := range result.Rows {
						an, ar, err := e.modifier.ExecuteDelete(t, clause.Delete, row, params)
						if err != nil {
							execErr = err
							break
						}
						result.AddAffected(an, ar)
					}
				} else {
					an, ar, err := e.modifier.ExecuteDelete(t, clause.Delete, varVars, params)
					execErr = err
					if execErr == nil {
						result.AddAffected(an, ar)
					}
				}
				hasWrite = true
			}
		case *ast.SetClause:
			if len(result.Rows) > 0 {
				for _, row := range result.Rows {
					an, err := e.modifier.ExecuteSet(t, clause, row, params)
					if err != nil {
						execErr = err
						break
					}
					result.AddAffected(an, 0)
				}
			} else {
				an, err := e.modifier.ExecuteSet(t, clause, varVars, params)
				execErr = err
				if execErr == nil {
					result.AddAffected(an, 0)
				}
			}
			hasWrite = true
		case *ast.DeleteClause:
			if len(result.Rows) > 0 {
				for _, row := range result.Rows {
					an, ar, err := e.modifier.ExecuteDelete(t, clause, row, params)
					if err != nil {
						execErr = err
						break
					}
					result.AddAffected(an, ar)
				}
			} else {
				an, ar, err := e.modifier.ExecuteDelete(t, clause, varVars, params)
				execErr = err
				if execErr == nil {
					result.AddAffected(an, ar)
				}
			}
			hasWrite = true
		case *ast.RemoveClause:
			if len(result.Rows) > 0 {
				for _, row := range result.Rows {
					an, err := e.modifier.ExecuteRemove(t, clause, row, params)
					if err != nil {
						execErr = err
						break
					}
					result.AddAffected(an, 0)
				}
			} else {
				an, err := e.modifier.ExecuteRemove(t, clause, varVars, params)
				execErr = err
				if execErr == nil {
					result.AddAffected(an, 0)
				}
			}
			hasWrite = true
		}

		if execErr != nil {
			e.observability.Logger.Error("execution failed", "error", execErr)
			return Result{}, execErr
		}
	}

	if hasWrite {
		if err := t.Commit(); err != nil {
			e.observability.Logger.Error("failed to commit transaction", "error", err)
			return Result{}, err
		}
	}

	return result, nil
}

// ExecuteWithTx runs a Cypher query within an existing transaction.
// It is intended for advanced use cases where transaction management is handled externally.
func (e *Executor) ExecuteWithTx(t *tx.Transaction, a *ast.AST, params map[string]interface{}) (Result, error) {
	result := Result{}
	varVars := make(map[string]interface{})

	for _, stmt := range a.Statements {
		var execErr error
		switch clause := stmt.Clause.(type) {
		case *ast.CreateClause:
			an, ar, err := e.creator.Execute(t, clause)
			execErr = err
			if execErr == nil {
				result.AddAffected(an, ar)
			}
		case *ast.MatchClause:
			rows, cols, err := e.matcher.Execute(clause, varVars, params)
			execErr = err
			if execErr == nil {
				result.Rows = rows
				result.Columns = cols
			}
			if clause.Delete != nil && execErr == nil {
				if len(result.Rows) > 0 {
					for _, row := range result.Rows {
						an, ar, err := e.modifier.ExecuteDelete(t, clause.Delete, row, params)
						if err != nil {
							execErr = err
							break
						}
						result.AddAffected(an, ar)
					}
				} else {
					an, ar, err := e.modifier.ExecuteDelete(t, clause.Delete, varVars, params)
					execErr = err
					if execErr == nil {
						result.AddAffected(an, ar)
					}
				}
			}
		case *ast.SetClause:
			if len(result.Rows) > 0 {
				for _, row := range result.Rows {
					an, err := e.modifier.ExecuteSet(t, clause, row, params)
					if err != nil {
						execErr = err
						break
					}
					result.AddAffected(an, 0)
				}
			} else {
				an, err := e.modifier.ExecuteSet(t, clause, varVars, params)
				execErr = err
				if execErr == nil {
					result.AddAffected(an, 0)
				}
			}
		case *ast.DeleteClause:
			if len(result.Rows) > 0 {
				for _, row := range result.Rows {
					an, ar, err := e.modifier.ExecuteDelete(t, clause, row, params)
					if err != nil {
						execErr = err
						break
					}
					result.AddAffected(an, ar)
				}
			} else {
				an, ar, err := e.modifier.ExecuteDelete(t, clause, varVars, params)
				execErr = err
				if execErr == nil {
					result.AddAffected(an, ar)
				}
			}
		case *ast.RemoveClause:
			if len(result.Rows) > 0 {
				for _, row := range result.Rows {
					an, err := e.modifier.ExecuteRemove(t, clause, row, params)
					if err != nil {
						execErr = err
						break
					}
					result.AddAffected(an, 0)
				}
			} else {
				an, err := e.modifier.ExecuteRemove(t, clause, varVars, params)
				execErr = err
				if execErr == nil {
					result.AddAffected(an, 0)
				}
			}
		}

		if execErr != nil {
			return Result{}, execErr
		}
	}

	return result, nil
}
