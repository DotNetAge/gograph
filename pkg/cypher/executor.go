package cypher

import (
	"context"

	"github.com/DotNetAge/gograph/pkg/cypher/ast"
	"github.com/DotNetAge/gograph/pkg/cypher/creators"
	"github.com/DotNetAge/gograph/pkg/cypher/matchers"
	"github.com/DotNetAge/gograph/pkg/cypher/modifiers"
	"github.com/DotNetAge/gograph/pkg/cypher/parser"
	"github.com/DotNetAge/gograph/pkg/graph"
	"github.com/DotNetAge/gograph/pkg/storage"
	"github.com/DotNetAge/gograph/pkg/tx"
)

type Executor struct {
	Store    *storage.DB
	Index    *graph.Index
	txMgr    *tx.Manager
	matcher  *matchers.Matcher
	creator  *creators.Creator
	modifier *modifiers.Modifier
}

func NewExecutor(store *storage.DB) *Executor {
	index := graph.NewIndex(store)
	return &Executor{
		Store:    store,
		Index:    index,
		txMgr:    tx.NewManager(store),
		matcher:  matchers.NewMatcher(store, index),
		creator:  creators.NewCreator(store),
		modifier: modifiers.NewModifier(store),
	}
}

func (e *Executor) Execute(ctx context.Context, query string, params map[string]interface{}) (Result, error) {
	p := parser.New(query)
	q, err := p.Parse()
	if err != nil {
		return Result{}, err
	}
	return e.ExecuteAST(ctx, q, params)
}

func (e *Executor) withTx(fn func(*tx.Transaction) error) error {
	t, err := e.txMgr.Begin(false)
	if err != nil {
		return err
	}
	if err := fn(t); err != nil {
		t.Rollback()
		return err
	}
	return t.Commit()
}

func (e *Executor) ExecuteWithTx(t *tx.Transaction, query *ast.Query, params map[string]interface{}) (Result, error) {
	result := Result{}
	varVars := make(map[string]interface{})

	for _, stmt := range query.Statements {
		var execErr error
		switch s := stmt.(type) {
		case *ast.CreateStmt:
			an, ar, err := e.creator.ExecuteCreate(t, s, params)
			if err != nil {
				return Result{}, err
			}
			result.AddAffected(an, ar)

		case *ast.MatchStmt:
			rows, cols, err := e.matcher.ExecuteMatch(s, params)
			execErr = err
			if execErr == nil {
				result.Rows = rows
				result.Columns = cols
			}
			execErr = e.handleMatchDelete(t, s, &result, params)

		case *ast.SetStmt:
			an, err := e.modifier.ExecuteSet(t, s, varVars, params)
			if err != nil {
				return Result{}, err
			}
			result.AddAffected(an, 0)

		case *ast.DeleteStmt:
			an, ar, err := e.modifier.ExecuteDelete(t, s, varVars, params)
			if err != nil {
				return Result{}, err
			}
			result.AddAffected(an, ar)

		case *ast.RemoveStmt:
			an, err := e.modifier.ExecuteRemove(t, s, varVars, params)
			if err != nil {
				return Result{}, err
			}
			result.AddAffected(an, 0)

		case *ast.MergeStmt:
			an, ar, err := e.creator.ExecuteMerge(t, s, params)
			if err != nil {
				return Result{}, err
			}
			result.AddAffected(an, ar)
		}

		if execErr != nil {
			return Result{}, execErr
		}
	}

	return result, nil
}

func (e *Executor) handleMatchDelete(t *tx.Transaction, s *ast.MatchStmt, result *Result, params map[string]interface{}) error {
	var deleteClause *ast.DeleteClause
	for _, clause := range s.Clauses {
		if dc, ok := clause.(*ast.DeleteClause); ok {
			deleteClause = dc
			break
		}
	}

	if deleteClause == nil {
		return nil
	}

	for _, row := range result.Rows {
		an, ar, err := e.modifier.ExecuteDelete(t, &ast.DeleteStmt{
			Detach: deleteClause.Detach,
			Items:  deleteClause.Items,
		}, row, params)
		if err != nil {
			return err
		}
		result.AddAffected(an, ar)
	}
	return nil
}

func (e *Executor) ExecuteAST(ctx context.Context, query *ast.Query, params map[string]interface{}) (Result, error) {
	result := Result{}
	varVars := make(map[string]interface{})

	for _, stmt := range query.Statements {
		var execErr error
		switch s := stmt.(type) {
		case *ast.CreateStmt:
			execErr = e.withTx(func(t *tx.Transaction) error {
				an, ar, err := e.creator.ExecuteCreate(t, s, params)
				if err != nil {
					return err
				}
				result.AddAffected(an, ar)
				return nil
			})

		case *ast.MatchStmt:
			rows, cols, err := e.matcher.ExecuteMatch(s, params)
			execErr = err
			if execErr == nil {
				result.Rows = rows
				result.Columns = cols
			}
			if execErr == nil {
				execErr = e.handleMatchDeleteAST(s, &result, params)
			}

		case *ast.SetStmt:
			execErr = e.withTx(func(t *tx.Transaction) error {
				an, err := e.modifier.ExecuteSet(t, s, varVars, params)
				if err != nil {
					return err
				}
				result.AddAffected(an, 0)
				return nil
			})

		case *ast.DeleteStmt:
			execErr = e.withTx(func(t *tx.Transaction) error {
				an, ar, err := e.modifier.ExecuteDelete(t, s, varVars, params)
				if err != nil {
					return err
				}
				result.AddAffected(an, ar)
				return nil
			})

		case *ast.RemoveStmt:
			execErr = e.withTx(func(t *tx.Transaction) error {
				an, err := e.modifier.ExecuteRemove(t, s, varVars, params)
				if err != nil {
					return err
				}
				result.AddAffected(an, 0)
				return nil
			})

		case *ast.MergeStmt:
			execErr = e.withTx(func(t *tx.Transaction) error {
				an, ar, err := e.creator.ExecuteMerge(t, s, params)
				if err != nil {
					return err
				}
				result.AddAffected(an, ar)
				return nil
			})
		}

		if execErr != nil {
			return Result{}, execErr
		}
	}

	return result, nil
}

func (e *Executor) handleMatchDeleteAST(s *ast.MatchStmt, result *Result, params map[string]interface{}) error {
	var deleteClause *ast.DeleteClause
	for _, clause := range s.Clauses {
		if dc, ok := clause.(*ast.DeleteClause); ok {
			deleteClause = dc
			break
		}
	}

	if deleteClause == nil {
		return nil
	}

	for _, row := range result.Rows {
		err := e.withTx(func(t *tx.Transaction) error {
			an, ar, err := e.modifier.ExecuteDelete(t, &ast.DeleteStmt{
				Detach: deleteClause.Detach,
				Items:  deleteClause.Items,
			}, row, params)
			if err != nil {
				return err
			}
			result.AddAffected(an, ar)
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
