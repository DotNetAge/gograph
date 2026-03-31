// Package api provides the public database API for gograph, a graph database engine.
// It offers a high-level interface for executing Cypher queries and managing
// graph data including nodes, relationships, and their properties.
package api

import (
	"context"
	"errors"
	"sync"

	"github.com/DotNetAge/gograph/pkg/cypher"
	"github.com/DotNetAge/gograph/pkg/storage"
	"github.com/DotNetAge/gograph/pkg/tx"
)

// ErrDBClosed is returned when attempting to perform operations on a closed database.
var ErrDBClosed = errors.New("database is closed")

// DB represents a graph database instance. It provides methods for executing
// Cypher queries and managing transactions. A DB instance is safe for
// concurrent use by multiple goroutines.
//
// DB manages the lifecycle of the underlying storage and provides a high-level
// interface for graph operations. It handles query parsing, execution, and
// result formatting.
//
// Example:
//
//	db, err := api.Open("/path/to/db")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer db.Close()
//
//	// Execute queries
//	result, err := db.Exec(ctx, "CREATE (n:Person {name: 'Alice'})")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Query data
//	rows, err := db.Query(ctx, "MATCH (n:Person) RETURN n.name")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer rows.Close()
type DB struct {
	store    *storage.DB
	executor *cypher.Executor
	closed   bool
	mu       sync.RWMutex
	obs      *cypher.Observability
}

// DBOption configures optional parameters for database operations.
// It is used with the Open function to customize database behavior.
type DBOption func(*DB)

// WithObservability enables observability features for the database executor,
// allowing tracing and metrics collection during query execution.
//
// Parameters:
//   - o: The Observability instance to use
//
// Returns a DBOption that can be passed to Open.
//
// Example:
//
//	obs := cypher.NewObservability()
//	db, err := api.Open("/path/to/db", api.WithObservability(obs))
func WithObservability(o *cypher.Observability) DBOption {
	return func(db *DB) {
		db.obs = o
	}
}

// Open opens a database at the specified path and returns a DB instance.
// The path is used as the storage location for the underlying Pebble database.
// Optional DBOption functions can be provided to configure the database.
//
// Parameters:
//   - path: The directory path where the database files will be stored
//   - opts: Optional configuration options
//
// Returns a new DB instance or an error if the database cannot be opened.
//
// Example:
//
//	// Open with default options
//	db, err := api.Open("/var/lib/gograph/mydb")
//
//	// Open with observability
//	obs := cypher.NewObservability()
//	db, err := api.Open("/var/lib/gograph/mydb", api.WithObservability(obs))
func Open(path string, opts ...DBOption) (*DB, error) {
	store, err := storage.Open(path)
	if err != nil {
		return nil, err
	}
	db := &DB{
		store: store,
		obs:   cypher.NewObservability(),
	}
	for _, opt := range opts {
		opt(db)
	}
	db.executor = cypher.NewExecutor(store)
	return db, nil
}

// Close closes the database and releases all associated resources.
// Any pending operations will be completed before closing.
//
// Returns an error if the database is already closed.
//
// Example:
//
//	db, err := api.Open("/path/to/db")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer db.Close()
func (db *DB) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.closed {
		return ErrDBClosed
	}
	db.closed = true
	return db.store.Close()
}

// Exec executes a Cypher query that modifies data (CREATE, SET, DELETE, REMOVE)
// and returns a Result containing the count of affected nodes and relationships.
// The query can include optional positional parameters passed as arguments.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - cypherQuery: The Cypher query string to execute
//   - args: Optional query parameters. If the first argument is a map[string]interface{},
//     it is used as the parameter map.
//
// Returns a Result containing affected counts, or an error if the query fails.
//
// Example:
//
//	// Simple create
//	result, err := db.Exec(ctx, "CREATE (n:Person {name: 'Alice'})")
//
//	// With parameters
//	result, err := db.Exec(ctx,
//	    "CREATE (n:Person {name: $name, age: $age})",
//	    map[string]interface{}{
//	        "name": "Alice",
//	        "age": 30,
//	    },
//	)
//
//	fmt.Printf("Created %d nodes\n", result.NodesCreated)
func (db *DB) Exec(ctx context.Context, cypherQuery string, args ...interface{}) (Result, error) {
	db.mu.RLock()
	if db.closed {
		db.mu.RUnlock()
		return Result{}, ErrDBClosed
	}
	db.mu.RUnlock()

	params := extractParams(args)

	db.mu.RLock()
	defer db.mu.RUnlock()
	if db.closed {
		return Result{}, ErrDBClosed
	}

	cypherResult, err := db.executor.Execute(ctx, cypherQuery, params)
	if err != nil {
		return Result{}, err
	}

	return cypherResult, nil
}

// Query executes a Cypher query that returns rows of data (MATCH ... RETURN)
// and returns a Rows iterator for scanning the results.
// The query can include optional positional parameters passed as arguments.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - cypherQuery: The Cypher query string to execute
//   - args: Optional query parameters. If the first argument is a map[string]interface{},
//     it is used as the parameter map.
//
// Returns a Rows iterator for scanning results, or an error if the query fails.
//
// Example:
//
//	rows, err := db.Query(ctx, "MATCH (n:Person) RETURN n.name, n.age")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer rows.Close()
//
//	for rows.Next() {
//	    var name string
//	    var age int
//	    if err := rows.Scan(&name, &age); err != nil {
//	        log.Fatal(err)
//	    }
//	    fmt.Printf("Name: %s, Age: %d\n", name, age)
//	}
func (db *DB) Query(ctx context.Context, cypherQuery string, args ...interface{}) (*Rows, error) {
	db.mu.RLock()
	if db.closed {
		db.mu.RUnlock()
		return nil, ErrDBClosed
	}
	db.mu.RUnlock()

	params := extractParams(args)

	db.mu.RLock()
	defer db.mu.RUnlock()
	if db.closed {
		return nil, ErrDBClosed
	}

	result, err := db.executor.Execute(ctx, cypherQuery, params)
	if err != nil {
		return nil, err
	}

	return &Rows{
		result:  result.Rows,
		columns: result.Columns,
		index:   -1,
	}, nil
}

// BeginTx starts a new transaction with the specified options.
// If opts is nil, a read-write transaction is started.
// A transaction allows grouping multiple operations into a single atomic unit.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - opts: Transaction options. Pass nil for default read-write transaction.
//
// Returns a new Tx transaction object, or an error if the transaction cannot be started.
//
// Example:
//
//	// Begin a transaction
//	tx, err := db.BeginTx(ctx, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Perform operations
//	_, err = tx.Exec("CREATE (n:Person {name: 'Alice'})")
//	if err != nil {
//	    tx.Rollback()
//	    log.Fatal(err)
//	}
//
//	// Commit
//	if err := tx.Commit(); err != nil {
//	    log.Fatal(err)
//	}
func (db *DB) BeginTx(ctx context.Context, opts *TxOptions) (*Tx, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	if db.closed {
		return nil, ErrDBClosed
	}

	readOnly := opts != nil && opts.ReadOnly
	t, err := tx.NewManager(db.store).Begin(readOnly)
	if err != nil {
		return nil, err
	}

	return &Tx{
		db:     db,
		ctx:    ctx,
		tx:     t,
		closed: false,
	}, nil
}

// IsClosed returns true if the database has been closed.
//
// Example:
//
//	if !db.IsClosed() {
//	    db.Close()
//	}
func (db *DB) IsClosed() bool {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.closed
}

// Store returns the underlying storage database.
// This is primarily for advanced use cases and testing.
func (db *DB) Store() *storage.DB {
	return db.store
}

// Lock acquires a write lock on the database.
// This is used internally for synchronization.
func (db *DB) Lock() {
	db.mu.Lock()
}

// Unlock releases the write lock on the database.
// This is used internally for synchronization.
func (db *DB) Unlock() {
	db.mu.Unlock()
}

// RLock acquires a read lock on the database.
// This is used internally for synchronization.
func (db *DB) RLock() {
	db.mu.RLock()
}

// RUnlock releases the read lock on the database.
// This is used internally for synchronization.
func (db *DB) RUnlock() {
	db.mu.RUnlock()
}

// IsClosedLocked returns true if the database is closed.
// This should only be called while holding a lock.
func (db *DB) IsClosedLocked() bool {
	return db.closed
}

// extractParams extracts query parameters from argument list.
// If the first argument is a map[string]interface{}, it is used as the
// parameter map; otherwise, returns nil.
//
// Parameters:
//   - args: The argument list to extract parameters from
//
// Returns a map of parameter names to values, or nil if no parameters found.
func extractParams(args []interface{}) map[string]interface{} {
	if len(args) == 0 {
		return nil
	}
	if params, ok := args[0].(map[string]interface{}); ok {
		return params
	}
	return nil
}

// TxOptions specifies configuration options for a transaction.
type TxOptions struct {
	// ReadOnly indicates whether the transaction should be read-only.
	// A read-only transaction cannot modify data but may have better performance.
	ReadOnly bool
}

// Tx represents a database transaction. It provides methods for executing
// Cypher queries within a transaction. A transaction is safe for concurrent
// use by multiple goroutines, though serial execution is recommended.
//
// Transactions provide ACID guarantees:
//   - Atomicity: All operations succeed or all fail
//   - Consistency: Database remains in a consistent state
//   - Isolation: Transactions don't interfere with each other
//   - Durability: Committed changes persist
//
// Example:
//
//	tx, err := db.BeginTx(ctx, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer func() {
//	    if err != nil {
//	        tx.Rollback()
//	    }
//	}()
//
//	// Perform multiple operations
//	_, err = tx.Exec("CREATE (n:Person {name: 'Alice'})")
//	_, err = tx.Exec("CREATE (n:Person {name: 'Bob'})")
//
//	// Commit all operations
//	err = tx.Commit()
type Tx struct {
	db     *DB
	ctx    context.Context
	tx     *tx.Transaction
	closed bool
	mu     sync.Mutex
}

// Exec executes a Cypher query within the transaction and returns a Result
// containing the count of affected nodes and relationships.
// The query can include optional positional parameters passed as arguments.
//
// Parameters:
//   - cypherQuery: The Cypher query string to execute
//   - args: Optional query parameters
//
// Returns a Result containing affected counts, or an error if the query fails
// or the transaction is closed.
//
// Example:
//
//	result, err := tx.Exec("CREATE (n:Person {name: 'Alice'})")
//	if err != nil {
//	    tx.Rollback()
//	    log.Fatal(err)
//	}
func (tx *Tx) Exec(cypherQuery string, args ...interface{}) (Result, error) {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.closed {
		return Result{}, errors.New("transaction closed")
	}

	params := extractParams(args)

	p := cypher.NewParser(cypherQuery)
	ast, err := p.Parse()
	if err != nil {
		return Result{}, err
	}

	cypherResult, err := tx.db.executor.ExecuteWithTx(tx.tx, ast, params)
	if err != nil {
		return Result{}, err
	}

	return cypherResult, nil
}

// Query executes a Cypher query within the transaction and returns a Rows
// iterator for scanning the results.
// The query can include optional positional parameters passed as arguments.
//
// Parameters:
//   - cypherQuery: The Cypher query string to execute
//   - args: Optional query parameters
//
// Returns a Rows iterator for scanning results, or an error if the query fails
// or the transaction is closed.
//
// Example:
//
//	rows, err := tx.Query("MATCH (n:Person) RETURN n.name")
//	if err != nil {
//	    tx.Rollback()
//	    log.Fatal(err)
//	}
//	defer rows.Close()
//
//	for rows.Next() {
//	    var name string
//	    rows.Scan(&name)
//	    fmt.Println(name)
//	}
func (tx *Tx) Query(cypherQuery string, args ...interface{}) (*Rows, error) {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.closed {
		return nil, errors.New("transaction closed")
	}

	params := extractParams(args)

	p := cypher.NewParser(cypherQuery)
	ast, err := p.Parse()
	if err != nil {
		return nil, err
	}

	result, err := tx.db.executor.ExecuteWithTx(tx.tx, ast, params)
	if err != nil {
		return nil, err
	}

	return &Rows{
		result:  result.Rows,
		columns: result.Columns,
		index:   -1,
	}, nil
}

// Commit commits the transaction, making all changes permanent.
// After commit, the transaction cannot be used anymore.
//
// Returns an error if the transaction is already closed or if the commit fails.
//
// Example:
//
//	if err := tx.Commit(); err != nil {
//	    log.Fatal(err)
//	}
func (tx *Tx) Commit() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.closed {
		return errors.New("transaction closed")
	}
	tx.closed = true
	return tx.tx.Commit()
}

// Rollback aborts the transaction and discards all changes.
// If the transaction is already closed, Rollback returns nil.
// Rollback is safe to call multiple times.
//
// Example:
//
//	defer func() {
//	    if err != nil {
//	        tx.Rollback()
//	    }
//	}()
func (tx *Tx) Rollback() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.closed {
		return nil
	}
	tx.closed = true
	return tx.tx.Rollback()
}
