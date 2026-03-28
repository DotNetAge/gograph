package main

import (
	"context"
	"os"
	"testing"

	"github.com/DotNetAge/gograph/pkg/api"
)

func TestHandleInternalCommand(t *testing.T) {
	// Create a temporary database for testing
	tempPath := "/tmp/gograph_test.db"
	defer os.RemoveAll(tempPath)

	db, err := api.Open(tempPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Test non-internal command
	if handleInternalCommand("CREATE (n:User)", db) {
		t.Error("expected non-internal command to return false")
	}

	// Test unknown internal command
	if !handleInternalCommand("/unknown", db) {
		t.Error("expected internal command to return true")
	}

	// Test help command
	if !handleInternalCommand("/help", db) {
		t.Error("expected /help command to return true")
	}

	// Test exec command with empty args
	if !handleInternalCommand("/exec", db) {
		t.Error("expected /exec command to return true")
	}

	// Test query command with empty args
	if !handleInternalCommand("/query", db) {
		t.Error("expected /query command to return true")
	}

	// Test exec command with valid query
	if !handleInternalCommand("/exec CREATE (n:User {name: 'Test'})", db) {
		t.Error("expected /exec command to return true")
	}

	// Test query command with valid query
	if !handleInternalCommand("/query MATCH (n:User) RETURN n", db) {
		t.Error("expected /query command to return true")
	}
}

func TestExecuteExec(t *testing.T) {
	// Create a temporary database for testing
	tempPath := "/tmp/gograph_test.db"
	defer os.RemoveAll(tempPath)

	db, err := api.Open(tempPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Test with valid query
	executeExec(db, "CREATE (n:User {name: 'Test'})")

	// Test with invalid query
	executeExec(db, "INVALID CYPHER QUERY")
}

func TestExecuteQuery(t *testing.T) {
	// Create a temporary database for testing
	tempPath := "/tmp/gograph_test.db"
	defer os.RemoveAll(tempPath)

	db, err := api.Open(tempPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Test with empty result set
	executeQuery(db, "MATCH (n:NonExistent) RETURN n")

	// Create a test node
	_, err = db.Exec(context.Background(), "CREATE (n:User {name: 'Test'})")
	if err != nil {
		t.Fatalf("failed to create test node: %v", err)
	}

	// Test with valid query
	executeQuery(db, "MATCH (n:User) RETURN n")

	// Test with invalid query
	executeQuery(db, "INVALID CYPHER QUERY")
}

func TestRunTUI(t *testing.T) {
	// Test that runTUI returns an error for invalid database path
	// Since we can't easily test the interactive part, we'll test the initialization
	// by passing a non-existent directory
	err := runTUI("/nonexistent/directory/db")
	if err == nil {
		t.Error("expected runTUI to return an error for invalid database path")
	}

	// Note: We can't easily test the full interactive TUI behavior in unit tests
	// because it requires user input from stdin. The runTUI function will
	// typically hang or exit gracefully when run in a test environment.
	// We'll skip testing the full TUI execution and focus on the error cases.
}
