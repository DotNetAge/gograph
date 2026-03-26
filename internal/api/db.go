package api

import (
	"context"
	"errors"
	"sync"

	"github.com/DotNetAge/gograph/internal/cypher"
	"github.com/DotNetAge/gograph/internal/storage"
	"github.com/DotNetAge/gograph/internal/tx"
)

var (
	ErrDBClosed = errors.New("database is closed")
)

type DB struct {
	store    *storage.DB
	executor *cypher.Executor
	closed   bool
	mu       sync.RWMutex
	obs      *cypher.Observability
}

type DBOption func(*DB)

func WithObservability(o *cypher.Observability) DBOption {
	return func(db *DB) {
		db.obs = o
	}
}

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
	db.executor = cypher.NewExecutor(store, cypher.WithObservability(db.obs))
	return db, nil
}

func (db *DB) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.closed {
		return ErrDBClosed
	}
	db.closed = true
	return db.store.Close()
}

func (db *DB) Exec(ctx context.Context, cypherQuery string, args ...interface{}) (Result, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	if db.closed {
		return Result{}, ErrDBClosed
	}

	params := extractParams(args)
	ast, err := cypher.NewParser(cypherQuery).Parse()
	if err != nil {
		return Result{}, err
	}

	cypherResult, err := db.executor.Execute(ctx, ast, params)
	if err != nil {
		return Result{}, err
	}

	return Result{
		AffectedNodes: cypherResult.AffectedNodes,
		AffectedRels:  cypherResult.AffectedRels,
	}, nil
}

func (db *DB) Query(ctx context.Context, cypherQuery string, args ...interface{}) (*Rows, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	if db.closed {
		return nil, ErrDBClosed
	}

	params := extractParams(args)
	ast, err := cypher.NewParser(cypherQuery).Parse()
	if err != nil {
		return nil, err
	}

	result, err := db.executor.Execute(ctx, ast, params)
	if err != nil {
		return nil, err
	}

	return &Rows{
		result:  result.Rows,
		columns: result.Columns,
		index:   -1,
	}, nil
}

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

func (db *DB) IsClosed() bool {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.closed
}

func extractParams(args []interface{}) map[string]interface{} {
	if len(args) == 0 {
		return nil
	}
	if params, ok := args[0].(map[string]interface{}); ok {
		return params
	}
	return nil
}

type TxOptions struct {
	ReadOnly bool
}

type Tx struct {
	db     *DB
	ctx    context.Context
	tx     *tx.Transaction
	closed bool
	mu     sync.Mutex
}

func (tx *Tx) Exec(cypherQuery string, args ...interface{}) (Result, error) {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.closed {
		return Result{}, errors.New("transaction closed")
	}

	params := extractParams(args)
	ast, err := cypher.NewParser(cypherQuery).Parse()
	if err != nil {
		return Result{}, err
	}

	cypherResult, err := tx.db.executor.ExecuteWithTx(tx.tx, ast, params)
	if err != nil {
		return Result{}, err
	}

	return Result{
		AffectedNodes: cypherResult.AffectedNodes,
		AffectedRels:  cypherResult.AffectedRels,
	}, nil
}

func (tx *Tx) Query(cypherQuery string, args ...interface{}) (*Rows, error) {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.closed {
		return nil, errors.New("transaction closed")
	}

	params := extractParams(args)
	ast, err := cypher.NewParser(cypherQuery).Parse()
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

func (tx *Tx) Commit() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.closed {
		return errors.New("transaction closed")
	}
	tx.closed = true
	return tx.tx.Commit()
}

func (tx *Tx) Rollback() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.closed {
		return nil
	}
	tx.closed = true
	return tx.tx.Rollback()
}
