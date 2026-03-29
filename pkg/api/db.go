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
type DB struct {
	store    *storage.DB
	executor *cypher.Executor
	closed   bool
	mu       sync.RWMutex
	obs      *cypher.Observability
}

// DBOption configures optional parameters for database operations.
type DBOption func(*DB)

// WithObservability enables observability features for the database executor,
// allowing tracing and metrics collection during query execution.
func WithObservability(o *cypher.Observability) DBOption {
	return func(db *DB) {
		db.obs = o
	}
}

// Open opens a database at the specified path and returns a DB instance.
// The path is used as the storage location for the underlying Pebble database.
// Optional DBOption functions can be provided to configure the database.
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
// It returns an error if the database is already closed.
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
func (db *DB) IsClosed() bool {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.closed
}

// extractParams extracts query parameters from argument list.
// If the first argument is a map[string]interface{}, it is used as the
// parameter map; otherwise, returns nil.
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
// Returns an error if the transaction is already closed or if the commit fails.
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
func (tx *Tx) Rollback() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.closed {
		return nil
	}
	tx.closed = true
	return tx.tx.Rollback()
}
