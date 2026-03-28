package api

import (
	"context"
	"os"
	"testing"

	"github.com/DotNetAge/gograph/pkg/cypher"
)

func TestDBClose(t *testing.T) {
	path := "/tmp/gograph_close_test"
	defer os.RemoveAll(path)

	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	err = db.Close()
	if err != nil {
		t.Fatalf("failed to close db: %v", err)
	}

	if !db.IsClosed() {
		t.Error("expected db to be closed")
	}

	err = db.Close()
	if err != ErrDBClosed {
		t.Errorf("expected ErrDBClosed, got %v", err)
	}
}

func TestDBClosedOperations(t *testing.T) {
	path := "/tmp/gograph_closed_ops_test"
	defer os.RemoveAll(path)

	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	db.Close()

	ctx := context.Background()

	_, err = db.Exec(ctx, "CREATE (n:User)")
	if err != ErrDBClosed {
		t.Errorf("expected ErrDBClosed for Exec, got %v", err)
	}

	_, err = db.Query(ctx, "MATCH (n) RETURN n")
	if err != ErrDBClosed {
		t.Errorf("expected ErrDBClosed for Query, got %v", err)
	}

	_, err = db.BeginTx(ctx, nil)
	if err != ErrDBClosed {
		t.Errorf("expected ErrDBClosed for BeginTx, got %v", err)
	}
}

func TestTxClosed(t *testing.T) {
	path := "/tmp/gograph_tx_closed_test"
	defer os.RemoveAll(path)

	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("failed to commit tx: %v", err)
	}

	_, err = tx.Exec("CREATE (n:User)")
	if err == nil {
		t.Error("expected error for Exec on closed transaction")
	}

	_, err = tx.Query("MATCH (n) RETURN n")
	if err == nil {
		t.Error("expected error for Query on closed transaction")
	}

	err = tx.Commit()
	if err == nil {
		t.Error("expected error for Commit on closed transaction")
	}

	err = tx.Rollback()
	if err != nil {
		t.Errorf("expected nil for Rollback on closed transaction, got %v", err)
	}
}

func TestTxRollbackAfterCommit(t *testing.T) {
	path := "/tmp/gograph_tx_rollback_test"
	defer os.RemoveAll(path)

	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("failed to commit tx: %v", err)
	}

	err = tx.Rollback()
	if err != nil {
		t.Errorf("expected nil for Rollback after Commit, got %v", err)
	}
}

func TestExtractParams(t *testing.T) {
	tests := []struct {
		name     string
		args     []interface{}
		expected map[string]interface{}
	}{
		{"no args", nil, nil},
		{"empty args", []interface{}{}, nil},
		{"map params", []interface{}{map[string]interface{}{"key": "value"}}, map[string]interface{}{"key": "value"}},
		{"non-map args", []interface{}{"string", 123}, nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := extractParams(tc.args)
			if tc.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
			} else {
				if len(result) != len(tc.expected) {
					t.Errorf("expected %d params, got %d", len(tc.expected), len(result))
				}
			}
		})
	}
}

func TestReadOnlyTransaction(t *testing.T) {
	path := "/tmp/gograph_ro_tx_test"
	defer os.RemoveAll(path)

	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	opts := &TxOptions{ReadOnly: true}
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		t.Fatalf("failed to begin read-only tx: %v", err)
	}

	_, err = tx.Exec("CREATE (n:User)")
	if err == nil {
		t.Error("expected error for Exec in read-only transaction")
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("failed to commit read-only tx: %v", err)
	}
}

func TestOpenWithObservability(t *testing.T) {
	path := "/tmp/gograph_obs_test"
	defer os.RemoveAll(path)

	obs := cypher.NewObservability()
	db, err := Open(path, WithObservability(obs))
	if err != nil {
		t.Fatalf("failed to open db with observability: %v", err)
	}
	defer db.Close()

	if db.obs != obs {
		t.Error("expected observability to be set")
	}
}

func TestDBExecWithParams(t *testing.T) {
	path := "/tmp/gograph_exec_params_test"
	defer os.RemoveAll(path)

	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Test with parameters
	params := map[string]interface{}{"name": "Test", "age": 30}
	result, err := db.Exec(ctx, "CREATE (n:User {name: $name, age: $age})", params)
	if err != nil {
		t.Fatalf("failed to exec with params: %v", err)
	}

	if result.AffectedNodes != 1 {
		t.Errorf("expected 1 affected node, got %d", result.AffectedNodes)
	}
}

func TestDBQueryWithParams(t *testing.T) {
	path := "/tmp/gograph_query_params_test"
	defer os.RemoveAll(path)

	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create test node
	_, err = db.Exec(ctx, "CREATE (n:User {name: 'Test', age: 30})")
	if err != nil {
		t.Fatalf("failed to create test node: %v", err)
	}

	// Test with parameters
	params := map[string]interface{}{"minAge": 25}
	rows, err := db.Query(ctx, "MATCH (n:User) WHERE n.age >= $minAge RETURN n", params)
	if err != nil {
		t.Fatalf("failed to query with params: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}

	if count != 1 {
		t.Errorf("expected 1 row, got %d", count)
	}
}

func TestTxExecWithParams(t *testing.T) {
	path := "/tmp/gograph_tx_exec_params_test"
	defer os.RemoveAll(path)

	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}

	// Test with parameters
	params := map[string]interface{}{"name": "Test", "age": 30}
	result, err := tx.Exec("CREATE (n:User {name: $name, age: $age})", params)
	if err != nil {
		t.Fatalf("failed to exec with params: %v", err)
	}

	if result.AffectedNodes != 1 {
		t.Errorf("expected 1 affected node, got %d", result.AffectedNodes)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("failed to commit tx: %v", err)
	}
}

func TestTxQueryWithParams(t *testing.T) {
	path := "/tmp/gograph_tx_query_params_test"
	defer os.RemoveAll(path)

	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create test node in first transaction
	tx1, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}

	_, err = tx1.Exec("CREATE (n:User {name: 'Test', age: 30})")
	if err != nil {
		t.Fatalf("failed to create test node: %v", err)
	}

	err = tx1.Commit()
	if err != nil {
		t.Fatalf("failed to commit tx: %v", err)
	}

	// Query in second transaction
	tx2, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}

	// Test with parameters
	params := map[string]interface{}{"minAge": 25}
	rows, err := tx2.Query("MATCH (n:User) WHERE n.age >= $minAge RETURN n", params)
	if err != nil {
		t.Fatalf("failed to query with params: %v", err)
	}
	defer rows.Close()

	// Test Columns method
	columns := rows.Columns()
	if len(columns) != 1 || columns[0] != "n" {
		t.Errorf("expected columns [n], got %v", columns)
	}

	// Test Scan method
	count := 0
	for rows.Next() {
		var node interface{}
		err := rows.Scan(&node)
		if err != nil {
			t.Fatalf("failed to scan row: %v", err)
		}
		count++
	}

	if count != 1 {
		t.Errorf("expected 1 row, got %d", count)
	}

	err = tx2.Commit()
	if err != nil {
		t.Fatalf("failed to commit tx: %v", err)
	}
}

func TestRowsScan(t *testing.T) {
	path := "/tmp/gograph_rows_scan_test"
	defer os.RemoveAll(path)

	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create test node with different property types
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}

	_, err = tx.Exec("CREATE (n:Test {name: 'Test', age: 30, score: 95.5, active: true})")
	if err != nil {
		t.Fatalf("failed to create test node: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("failed to commit tx: %v", err)
	}

	// Query for specific properties
	rows, err := db.Query(ctx, "MATCH (n:Test) RETURN n.name, n.age, n.score, n.active")
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}
	defer rows.Close()

	// Test Columns method
	columns := rows.Columns()
	expectedColumns := []string{"n.name", "n.age", "n.score", "n.active"}
	if len(columns) != len(expectedColumns) {
		t.Errorf("expected %d columns, got %d", len(expectedColumns), len(columns))
	}
	for i, col := range expectedColumns {
		if columns[i] != col {
			t.Errorf("expected column %d to be %s, got %s", i, col, columns[i])
		}
	}

	// Test Scan into different types
	if rows.Next() {
		var name string
		var age int
		var score float64
		var active bool

		err := rows.Scan(&name, &age, &score, &active)
		if err != nil {
			t.Fatalf("failed to scan row: %v", err)
		}

		if name != "Test" {
			t.Errorf("expected name 'Test', got '%s'", name)
		}
		if age != 30 {
			t.Errorf("expected age 30, got %d", age)
		}
		if score != 95.5 {
			t.Errorf("expected score 95.5, got %f", score)
		}
		if !active {
			t.Errorf("expected active true, got false")
		}
	} else {
		t.Fatalf("expected a row")
	}

	// Test Scan with too few arguments
	if rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			t.Fatalf("failed to scan row with fewer arguments: %v", err)
		}
	}

	// Test Scan after all rows are read
	var name string
	err = rows.Scan(&name)
	if err != ErrNoMoreRows {
		t.Errorf("expected ErrNoMoreRows, got %v", err)
	}
}
