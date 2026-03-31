// Package cypher provides Cypher query parsing and execution for gograph.
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

// Executor executes Cypher queries against the graph database.
// It coordinates parsing, planning, and execution of queries.
//
// The executor manages:
//   - Query parsing and AST generation
//   - Index management for efficient lookups
//   - Transaction coordination
//   - Match, create, and modify operations
//
// Example:
//
//	executor := cypher.NewExecutor(store)
//	result, err := executor.Execute(ctx, "MATCH (n:Person) RETURN n.name", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, row := range result.Rows {
//	    fmt.Println(row["n.name"])
//	}
type Executor struct {
	// Store is the underlying storage database.
	Store *storage.DB

	// Index provides efficient lookups for nodes and relationships.
	Index *graph.Index

	txMgr    *tx.Manager
	matcher  *matchers.Matcher
	creator  *creators.Creator
	modifier *modifiers.Modifier
}

// NewExecutor creates a new query executor for the given storage.
//
// Parameters:
//   - store: The storage database to execute queries against
//
// Returns a new Executor instance ready to execute queries.
//
// Example:
//
//	store, _ := storage.Open("/path/to/db")
//	executor := cypher.NewExecutor(store)
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

// Execute parses and executes a Cypher query string.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - query: The Cypher query string to execute
//   - params: Optional query parameters (e.g., {"name": "Alice"})
//
// Returns the query result or an error if execution fails.
//
// Example:
//
//	// Execute a CREATE query
//	result, err := executor.Execute(ctx, "CREATE (n:Person {name: 'Alice'})", nil)
//
//	// Execute with parameters
//	result, err := executor.Execute(ctx,
//	    "CREATE (n:Person {name: $name})",
//	    map[string]interface{}{"name": "Alice"},
//	)
func (e *Executor) Execute(ctx context.Context, query string, params map[string]interface{}) (Result, error) {
	p := parser.New(query)
	q, err := p.Parse()
	if err != nil {
		return Result{}, err
	}
	return e.ExecuteAST(ctx, q, params)
}

// withTx executes a function within a transaction.
// It handles transaction lifecycle including begin, commit, and rollback on error.
//
// Parameters:
//   - fn: The function to execute within the transaction
//
// Returns an error if the transaction fails.
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

// ExecuteWithTx executes a parsed query within an existing transaction.
//
// Parameters:
//   - t: The transaction to execute within
//   - query: The parsed AST query
//   - params: Optional query parameters
//
// Returns the query result or an error if execution fails.
//
// Example:
//
//	tx, _ := manager.Begin(false)
//	result, err := executor.ExecuteWithTx(tx, ast, params)
//	if err != nil {
//	    tx.Rollback()
//	} else {
//	    tx.Commit()
//	}
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

// handleMatchDelete handles DELETE clauses in MATCH statements.
//
// Parameters:
//   - t: The transaction to execute within
//   - s: The MATCH statement
//   - result: The current result to update
//   - params: Query parameters
//
// Returns an error if deletion fails.
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

// ExecuteAST executes a parsed AST query.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - query: The parsed AST query
//   - params: Optional query parameters
//
// Returns the query result or an error if execution fails.
//
// Example:
//
//	p := parser.New("MATCH (n:Person) RETURN n.name")
//	ast, _ := p.Parse()
//	result, err := executor.ExecuteAST(ctx, ast, nil)
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

// handleMatchDeleteAST handles DELETE clauses in MATCH statements for AST execution.
//
// Parameters:
//   - s: The MATCH statement
//   - result: The current result to update
//   - params: Query parameters
//
// Returns an error if deletion fails.
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
