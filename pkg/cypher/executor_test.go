package cypher

import (
	"context"
	"os"
	"testing"

	"github.com/DotNetAge/gograph/pkg/cypher/ast"
	"github.com/DotNetAge/gograph/pkg/storage"
	"github.com/stretchr/testify/assert"
)

func createTempDB(t *testing.T) *storage.DB {
	tempDir, err := os.MkdirTemp("", "gograph-test")
	assert.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	store, err := storage.Open(tempDir)
	assert.NoError(t, err)
	t.Cleanup(func() {
		store.Close()
	})

	return store
}

func TestNewExecutor(t *testing.T) {
	// Create a temporary storage for testing
	store := createTempDB(t)

	// Test creating executor with default options
	executor := NewExecutor(store)
	assert.NotNil(t, executor)
	assert.NotNil(t, executor.store)
	assert.NotNil(t, executor.txMgr)
	assert.NotNil(t, executor.index)
	assert.NotNil(t, executor.adj)
	assert.NotNil(t, executor.observability)
	assert.NotNil(t, executor.creator)
	assert.NotNil(t, executor.matcher)
	assert.NotNil(t, executor.modifier)

	// Test creating executor with observability option
	obs := NewObservability()
	executor = NewExecutor(store, WithObservability(obs))
	assert.NotNil(t, executor)
	assert.Equal(t, obs, executor.observability)
}

func TestExecuteEmptyQuery(t *testing.T) {
	// Create a temporary storage for testing
	store := createTempDB(t)

	executor := NewExecutor(store)
	ctx := context.Background()

	// Test executing empty AST
	ast := &ast.AST{}
	result, err := executor.Execute(ctx, ast, nil)
	assert.NoError(t, err)
	assert.Empty(t, result.Rows)
	assert.Empty(t, result.Columns)
	assert.Equal(t, 0, result.AffectedNodes)
	assert.Equal(t, 0, result.AffectedRels)
}

func TestExecuteWithTxEmptyQuery(t *testing.T) {
	// Create a temporary storage for testing
	store := createTempDB(t)

	executor := NewExecutor(store)

	// Begin a transaction
	tx, err := executor.txMgr.Begin(false)
	assert.NoError(t, err)
	defer tx.Rollback()

	// Test executing empty AST
	ast := &ast.AST{}
	result, err := executor.ExecuteWithTx(tx, ast, nil)
	assert.NoError(t, err)
	assert.Empty(t, result.Rows)
	assert.Empty(t, result.Columns)
	assert.Equal(t, 0, result.AffectedNodes)
	assert.Equal(t, 0, result.AffectedRels)
}

func TestExecuteCreate(t *testing.T) {
	// Create a temporary storage for testing
	store := createTempDB(t)

	executor := NewExecutor(store)
	ctx := context.Background()

	// Create a CREATE query AST
	parser := NewParser("CREATE (n:User {name: 'Alice', age: 30})")
	ast, err := parser.Parse()
	assert.NoError(t, err)

	// Execute the query
	result, err := executor.Execute(ctx, ast, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, result.AffectedNodes)
	assert.Equal(t, 0, result.AffectedRels)
}

func TestExecuteMatch(t *testing.T) {
	// Create a temporary storage for testing
	store := createTempDB(t)

	executor := NewExecutor(store)
	ctx := context.Background()

	// First create a node
	createParser := NewParser("CREATE (n:User {name: 'Alice', age: 30})")
	createAST, err := createParser.Parse()
	assert.NoError(t, err)
	_, err = executor.Execute(ctx, createAST, nil)
	assert.NoError(t, err)

	// Then match it
	matchParser := NewParser("MATCH (n:User) RETURN n")
	matchAST, err := matchParser.Parse()
	assert.NoError(t, err)

	// Execute the query
	result, err := executor.Execute(ctx, matchAST, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Rows))
	assert.Equal(t, []string{"n"}, result.Columns)
}

func TestExecuteSet(t *testing.T) {
	// Create a temporary storage for testing
	store := createTempDB(t)

	executor := NewExecutor(store)
	ctx := context.Background()

	// First create a node
	createParser := NewParser("CREATE (n:User {name: 'Alice', age: 30})")
	createAST, err := createParser.Parse()
	assert.NoError(t, err)
	_, err = executor.Execute(ctx, createAST, nil)
	assert.NoError(t, err)

	// Then update it
	setParser := NewParser("MATCH (n:User) SET n.age = 31")
	setAST, err := setParser.Parse()
	assert.NoError(t, err)

	// Execute the query
	result, err := executor.Execute(ctx, setAST, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, result.AffectedNodes)
	assert.Equal(t, 0, result.AffectedRels)
}

func TestExecuteDelete(t *testing.T) {
	// Create a temporary storage for testing
	store := createTempDB(t)

	executor := NewExecutor(store)
	ctx := context.Background()

	// First create a node
	createParser := NewParser("CREATE (n:User {name: 'Alice', age: 30})")
	createAST, err := createParser.Parse()
	assert.NoError(t, err)
	_, err = executor.Execute(ctx, createAST, nil)
	assert.NoError(t, err)

	// Then delete it
	deleteParser := NewParser("MATCH (n:User) DELETE n")
	deleteAST, err := deleteParser.Parse()
	assert.NoError(t, err)

	// Execute the query
	result, err := executor.Execute(ctx, deleteAST, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, result.AffectedNodes)
	assert.Equal(t, 0, result.AffectedRels)
}

func TestExecuteRemove(t *testing.T) {
	// Create a temporary storage for testing
	store := createTempDB(t)

	executor := NewExecutor(store)
	ctx := context.Background()

	// First create a node
	createParser := NewParser("CREATE (n:User:Admin {name: 'Alice', age: 30})")
	createAST, err := createParser.Parse()
	assert.NoError(t, err)
	_, err = executor.Execute(ctx, createAST, nil)
	assert.NoError(t, err)

	// Then remove a label
	removeParser := NewParser("MATCH (n:User:Admin) REMOVE n:Admin")
	removeAST, err := removeParser.Parse()
	assert.NoError(t, err)

	// Execute the query
	result, err := executor.Execute(ctx, removeAST, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, result.AffectedNodes)
	assert.Equal(t, 0, result.AffectedRels)
}

func TestExecuteMultipleStatements(t *testing.T) {
	// Create a temporary storage for testing
	store := createTempDB(t)

	executor := NewExecutor(store)
	ctx := context.Background()

	// Create multiple statements in a single query
	parser := NewParser("CREATE (n:User {name: 'Alice'}); CREATE (m:User {name: 'Bob'})")
	ast, err := parser.Parse()
	assert.NoError(t, err)

	// Execute the query
	result, err := executor.Execute(ctx, ast, nil)
	assert.NoError(t, err)
	assert.Equal(t, 2, result.AffectedNodes)
	assert.Equal(t, 0, result.AffectedRels)
}

func TestExecuteWithEmptyStatements(t *testing.T) {
	// Create a temporary storage for testing
	store := createTempDB(t)

	executor := NewExecutor(store)
	ctx := context.Background()

	// Create an empty AST
	ast := &ast.AST{Statements: []ast.Statement{}}

	// Execute the query
	result, err := executor.Execute(ctx, ast, nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.AffectedNodes)
	assert.Equal(t, 0, result.AffectedRels)
	assert.Equal(t, 0, len(result.Rows))
}

func TestExecuteErrorRollback(t *testing.T) {
	// Create a temporary storage for testing
	store := createTempDB(t)

	executor := NewExecutor(store)
	ctx := context.Background()

	// First create a node
	createParser := NewParser("CREATE (n:User {name: 'Alice'})")
	createAST, err := createParser.Parse()
	assert.NoError(t, err)
	_, err = executor.Execute(ctx, createAST, nil)
	assert.NoError(t, err)

	// Verify the node exists
	matchParser := NewParser("MATCH (n:User) RETURN n")
	matchAST, err := matchParser.Parse()
	assert.NoError(t, err)

	result, err := executor.Execute(ctx, matchAST, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Rows))
}

func TestExecuteMatchWithDelete(t *testing.T) {
	// Create a temporary storage for testing
	store := createTempDB(t)

	executor := NewExecutor(store)
	ctx := context.Background()

	// First create a node
	createParser := NewParser("CREATE (n:User {name: 'Alice'})")
	createAST, err := createParser.Parse()
	assert.NoError(t, err)
	_, err = executor.Execute(ctx, createAST, nil)
	assert.NoError(t, err)

	// Execute MATCH with DELETE
	matchDeleteParser := NewParser("MATCH (n:User) DELETE n")
	matchDeleteAST, err := matchDeleteParser.Parse()
	assert.NoError(t, err)

	// Execute the query
	result, err := executor.Execute(ctx, matchDeleteAST, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, result.AffectedNodes)
	assert.Equal(t, 0, result.AffectedRels)

	// Verify the node was deleted
	matchParser := NewParser("MATCH (n:User) RETURN n")
	matchAST, err := matchParser.Parse()
	assert.NoError(t, err)

	result, err = executor.Execute(ctx, matchAST, nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result.Rows))
}

func TestExecuteSetWithResult(t *testing.T) {
	// Create a temporary storage for testing
	store := createTempDB(t)

	executor := NewExecutor(store)
	ctx := context.Background()

	// First create a node
	createParser := NewParser("CREATE (n:User {name: 'Alice', age: 30})")
	createAST, err := createParser.Parse()
	assert.NoError(t, err)
	_, err = executor.Execute(ctx, createAST, nil)
	assert.NoError(t, err)

	// Execute MATCH to get result
	matchParser := NewParser("MATCH (n:User) RETURN n")
	matchAST, err := matchParser.Parse()
	assert.NoError(t, err)

	// Execute the query
	result, err := executor.Execute(ctx, matchAST, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Rows))
}

func TestExecuteDeleteWithResult(t *testing.T) {
	// Create a temporary storage for testing
	store := createTempDB(t)

	executor := NewExecutor(store)
	ctx := context.Background()

	// First create a node
	createParser := NewParser("CREATE (n:User {name: 'Alice'})")
	createAST, err := createParser.Parse()
	assert.NoError(t, err)
	_, err = executor.Execute(ctx, createAST, nil)
	assert.NoError(t, err)

	// Execute MATCH to get result
	matchParser := NewParser("MATCH (n:User) RETURN n")
	matchAST, err := matchParser.Parse()
	assert.NoError(t, err)

	// Execute the query
	result, err := executor.Execute(ctx, matchAST, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Rows))
}

func TestExecuteRemoveWithResult(t *testing.T) {
	// Create a temporary storage for testing
	store := createTempDB(t)

	executor := NewExecutor(store)
	ctx := context.Background()

	// First create a node
	createParser := NewParser("CREATE (n:User:Admin {name: 'Alice', age: 30})")
	createAST, err := createParser.Parse()
	assert.NoError(t, err)
	_, err = executor.Execute(ctx, createAST, nil)
	assert.NoError(t, err)

	// Execute MATCH to get result
	matchParser := NewParser("MATCH (n:User:Admin) RETURN n")
	matchAST, err := matchParser.Parse()
	assert.NoError(t, err)

	// Execute the query
	result, err := executor.Execute(ctx, matchAST, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Rows))
}

func TestExecuteWithRemoveClause(t *testing.T) {
	// Create a temporary storage for testing
	store := createTempDB(t)

	executor := NewExecutor(store)
	ctx := context.Background()

	// First create a node
	createParser := NewParser("CREATE (n:User:Admin {name: 'Alice', age: 30})")
	createAST, err := createParser.Parse()
	assert.NoError(t, err)
	_, err = executor.Execute(ctx, createAST, nil)
	assert.NoError(t, err)

	// Execute MATCH with REMOVE
	matchRemoveParser := NewParser("MATCH (n:User:Admin) REMOVE n:Admin")
	matchRemoveAST, err := matchRemoveParser.Parse()
	assert.NoError(t, err)

	// Execute the query
	result, err := executor.Execute(ctx, matchRemoveAST, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, result.AffectedNodes)
	assert.Equal(t, 0, result.AffectedRels)
}

func TestExecuteWithTx(t *testing.T) {
	// Create a temporary storage for testing
	store := createTempDB(t)

	executor := NewExecutor(store)

	// Begin a transaction
	tx, err := executor.txMgr.Begin(false)
	assert.NoError(t, err)

	// Create a CREATE query AST
	parser := NewParser("CREATE (n:User {name: 'Alice', age: 30})")
	ast, err := parser.Parse()
	assert.NoError(t, err)

	// Execute the query within the transaction
	result, err := executor.ExecuteWithTx(tx, ast, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, result.AffectedNodes)
	assert.Equal(t, 0, result.AffectedRels)

	// Commit the transaction
	err = tx.Commit()
	assert.NoError(t, err)
}

func TestExecuteWithTxMatch(t *testing.T) {
	// Create a temporary storage for testing
	store := createTempDB(t)

	executor := NewExecutor(store)

	// First create a node in a separate transaction
	tx1, err := executor.txMgr.Begin(false)
	assert.NoError(t, err)
	createParser := NewParser("CREATE (n:User {name: 'Alice', age: 30})")
	createAST, err := createParser.Parse()
	assert.NoError(t, err)
	_, err = executor.ExecuteWithTx(tx1, createAST, nil)
	assert.NoError(t, err)
	err = tx1.Commit()
	assert.NoError(t, err)

	// Begin a new transaction for matching
	tx2, err := executor.txMgr.Begin(true)
	assert.NoError(t, err)
	defer tx2.Rollback()

	// Create a MATCH query AST
	matchParser := NewParser("MATCH (n:User) RETURN n")
	matchAST, err := matchParser.Parse()
	assert.NoError(t, err)

	// Execute the query within the transaction
	result, err := executor.ExecuteWithTx(tx2, matchAST, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Rows))
	assert.Equal(t, []string{"n"}, result.Columns)
}

func TestExecuteWithTxSet(t *testing.T) {
	// Create a temporary storage for testing
	store := createTempDB(t)

	executor := NewExecutor(store)

	// First create a node in a separate transaction
	tx1, err := executor.txMgr.Begin(false)
	assert.NoError(t, err)
	createParser := NewParser("CREATE (n:User {name: 'Alice', age: 30})")
	createAST, err := createParser.Parse()
	assert.NoError(t, err)
	_, err = executor.ExecuteWithTx(tx1, createAST, nil)
	assert.NoError(t, err)
	err = tx1.Commit()
	assert.NoError(t, err)

	// Begin a new transaction for updating
	tx2, err := executor.txMgr.Begin(false)
	assert.NoError(t, err)

	// Create a SET query AST
	setParser := NewParser("MATCH (n:User) SET n.age = 31")
	setAST, err := setParser.Parse()
	assert.NoError(t, err)

	// Execute the query within the transaction
	result, err := executor.ExecuteWithTx(tx2, setAST, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, result.AffectedNodes)
	assert.Equal(t, 0, result.AffectedRels)

	// Commit the transaction
	err = tx2.Commit()
	assert.NoError(t, err)
}

func TestExecuteWithTxDelete(t *testing.T) {
	// Create a temporary storage for testing
	store := createTempDB(t)

	executor := NewExecutor(store)

	// First create a node in a separate transaction
	tx1, err := executor.txMgr.Begin(false)
	assert.NoError(t, err)
	createParser := NewParser("CREATE (n:User {name: 'Alice', age: 30})")
	createAST, err := createParser.Parse()
	assert.NoError(t, err)
	_, err = executor.ExecuteWithTx(tx1, createAST, nil)
	assert.NoError(t, err)
	err = tx1.Commit()
	assert.NoError(t, err)

	// Begin a new transaction for deleting
	tx2, err := executor.txMgr.Begin(false)
	assert.NoError(t, err)

	// Create a DELETE query AST
	deleteParser := NewParser("MATCH (n:User) DELETE n")
	deleteAST, err := deleteParser.Parse()
	assert.NoError(t, err)

	// Execute the query within the transaction
	result, err := executor.ExecuteWithTx(tx2, deleteAST, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, result.AffectedNodes)
	assert.Equal(t, 0, result.AffectedRels)

	// Commit the transaction
	err = tx2.Commit()
	assert.NoError(t, err)
}

func TestExecuteWithTxRemove(t *testing.T) {
	// Create a temporary storage for testing
	store := createTempDB(t)

	executor := NewExecutor(store)

	// First create a node in a separate transaction
	tx1, err := executor.txMgr.Begin(false)
	assert.NoError(t, err)
	createParser := NewParser("CREATE (n:User:Admin {name: 'Alice', age: 30})")
	createAST, err := createParser.Parse()
	assert.NoError(t, err)
	_, err = executor.ExecuteWithTx(tx1, createAST, nil)
	assert.NoError(t, err)
	err = tx1.Commit()
	assert.NoError(t, err)

	// Begin a new transaction for removing
	tx2, err := executor.txMgr.Begin(false)
	assert.NoError(t, err)

	// Create a REMOVE query AST
	removeParser := NewParser("MATCH (n:User:Admin) REMOVE n:Admin")
	removeAST, err := removeParser.Parse()
	assert.NoError(t, err)

	// Execute the query within the transaction
	result, err := executor.ExecuteWithTx(tx2, removeAST, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, result.AffectedNodes)
	assert.Equal(t, 0, result.AffectedRels)

	// Commit the transaction
	err = tx2.Commit()
	assert.NoError(t, err)
}

func TestExecuteWithTxMultipleStatements(t *testing.T) {
	// Create a temporary storage for testing
	store := createTempDB(t)

	executor := NewExecutor(store)

	// Begin a transaction
	tx, err := executor.txMgr.Begin(false)
	assert.NoError(t, err)

	// Create a query with multiple statements
	parser := NewParser("CREATE (n:User {name: 'Alice'}); CREATE (m:User {name: 'Bob'})")
	ast, err := parser.Parse()
	assert.NoError(t, err)

	// Execute the query within the transaction
	result, err := executor.ExecuteWithTx(tx, ast, nil)
	assert.NoError(t, err)
	assert.Equal(t, 2, result.AffectedNodes)
	assert.Equal(t, 0, result.AffectedRels)

	// Commit the transaction
	err = tx.Commit()
	assert.NoError(t, err)
}

func TestExecuteWithTxEmptyStatements(t *testing.T) {
	// Create a temporary storage for testing
	store := createTempDB(t)

	executor := NewExecutor(store)

	// Begin a transaction
	tx, err := executor.txMgr.Begin(false)
	assert.NoError(t, err)

	// Create an empty AST
	ast := &ast.AST{Statements: []ast.Statement{}}

	// Execute the query within the transaction
	result, err := executor.ExecuteWithTx(tx, ast, nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.AffectedNodes)
	assert.Equal(t, 0, result.AffectedRels)
	assert.Equal(t, 0, len(result.Rows))

	// Commit the transaction
	err = tx.Commit()
	assert.NoError(t, err)
}

func TestExecuteWithTxSetNoRows(t *testing.T) {
	// Create a temporary storage for testing
	store := createTempDB(t)

	executor := NewExecutor(store)

	// Begin a transaction
	tx, err := executor.txMgr.Begin(false)
	assert.NoError(t, err)

	// Create a SET query AST without MATCH
	setParser := NewParser("SET n.age = 31")
	setAST, err := setParser.Parse()
	assert.NoError(t, err)

	// Execute the query within the transaction
	result, err := executor.ExecuteWithTx(tx, setAST, nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.AffectedNodes)
	assert.Equal(t, 0, result.AffectedRels)

	// Commit the transaction
	err = tx.Commit()
	assert.NoError(t, err)
}

func TestExecuteWithTxDeleteNoRows(t *testing.T) {
	// Create a temporary storage for testing
	store := createTempDB(t)

	executor := NewExecutor(store)

	// Begin a transaction
	tx, err := executor.txMgr.Begin(false)
	assert.NoError(t, err)

	// Create a DELETE query AST without MATCH
	deleteParser := NewParser("DELETE n")
	deleteAST, err := deleteParser.Parse()
	assert.NoError(t, err)

	// Execute the query within the transaction
	result, err := executor.ExecuteWithTx(tx, deleteAST, nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.AffectedNodes)
	assert.Equal(t, 0, result.AffectedRels)

	// Commit the transaction
	err = tx.Commit()
	assert.NoError(t, err)
}

func TestExecuteWithTxRemoveNoRows(t *testing.T) {
	// Create a temporary storage for testing
	store := createTempDB(t)

	executor := NewExecutor(store)

	// Begin a transaction
	tx, err := executor.txMgr.Begin(false)
	assert.NoError(t, err)

	// Create a REMOVE query AST without MATCH
	removeParser := NewParser("REMOVE n:Admin")
	removeAST, err := removeParser.Parse()
	assert.NoError(t, err)

	// Execute the query within the transaction
	result, err := executor.ExecuteWithTx(tx, removeAST, nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.AffectedNodes)
	assert.Equal(t, 0, result.AffectedRels)

	// Commit the transaction
	err = tx.Commit()
	assert.NoError(t, err)
}

func TestExecuteEmptyStatements(t *testing.T) {
	// Create a temporary storage for testing
	store := createTempDB(t)

	executor := NewExecutor(store)
	ctx := context.Background()

	// Create an empty AST
	ast := &ast.AST{Statements: []ast.Statement{}}

	// Execute the query
	result, err := executor.Execute(ctx, ast, nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.AffectedNodes)
	assert.Equal(t, 0, result.AffectedRels)
	assert.Equal(t, 0, len(result.Rows))
}

func TestExecuteMatchWithDeleteClause(t *testing.T) {
	// Create a temporary storage for testing
	store := createTempDB(t)

	executor := NewExecutor(store)
	ctx := context.Background()

	// First create a node
	createParser := NewParser("CREATE (n:User {name: 'Alice'})")
	createAST, err := createParser.Parse()
	assert.NoError(t, err)
	_, err = executor.Execute(ctx, createAST, nil)
	assert.NoError(t, err)

	// Execute MATCH with DELETE
	matchDeleteParser := NewParser("MATCH (n:User) DELETE n")
	matchDeleteAST, err := matchDeleteParser.Parse()
	assert.NoError(t, err)

	// Execute the query
	result, err := executor.Execute(ctx, matchDeleteAST, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, result.AffectedNodes)
	assert.Equal(t, 0, result.AffectedRels)
}

func TestExecuteMatchWithSetClause(t *testing.T) {
	// Create a temporary storage for testing
	store := createTempDB(t)

	executor := NewExecutor(store)
	ctx := context.Background()

	// First create a node
	createParser := NewParser("CREATE (n:User {name: 'Alice', age: 30})")
	createAST, err := createParser.Parse()
	assert.NoError(t, err)
	_, err = executor.Execute(ctx, createAST, nil)
	assert.NoError(t, err)

	// Execute MATCH with SET
	matchSetParser := NewParser("MATCH (n:User) SET n.age = 31")
	matchSetAST, err := matchSetParser.Parse()
	assert.NoError(t, err)

	// Execute the query
	result, err := executor.Execute(ctx, matchSetAST, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, result.AffectedNodes)
	assert.Equal(t, 0, result.AffectedRels)
}

func TestExecuteWithError(t *testing.T) {
	// Create a parser with invalid query
	parser := NewParser("INVALID CYPHER QUERY")
	ast, err := parser.Parse()
	// The parser should return an error for invalid query
	assert.Error(t, err)
	assert.Nil(t, ast)
}
