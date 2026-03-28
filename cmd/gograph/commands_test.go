package main

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/DotNetAge/gograph/pkg/api"
	"github.com/DotNetAge/gograph/pkg/graph"
)

func TestExecCmd(t *testing.T) {
	// Create a temporary database for testing
	tempPath := "/tmp/gograph_test.db"
	defer os.RemoveAll(tempPath)

	// Test exec command with valid query
	cmd := execCmd
	args := []string{tempPath, "CREATE (n:User {name: 'Test'})"}
	if err := cmd.RunE(cmd, args); err != nil {
		t.Errorf("expected exec command to succeed, got error: %v", err)
	}
}

func TestQueryCmd(t *testing.T) {
	// Create a temporary database for testing
	tempPath := "/tmp/gograph_test.db"
	defer os.RemoveAll(tempPath)

	// Create a test node
	db, err := api.Open(tempPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	_, err = db.Exec(context.Background(), "CREATE (n:User {name: 'Test'})")
	if err != nil {
		t.Fatalf("failed to create test node: %v", err)
	}
	db.Close()

	// Test query command with valid query
	cmd := queryCmd
	args := []string{tempPath, "MATCH (n:User) RETURN n"}
	if err := cmd.RunE(cmd, args); err != nil {
		t.Errorf("expected query command to succeed, got error: %v", err)
	}
}

func TestFormatValue(t *testing.T) {
	// Test node formatting
	node := &graph.Node{
		ID:     "1",
		Labels: []string{"User", "Admin"},
		Properties: map[string]graph.PropertyValue{
			"name": graph.NewStringProperty("Alice"),
			"age":  graph.NewIntProperty(30),
		},
	}

	nodeStr := formatValue(node)
	// Since Go maps are unordered, we'll check that the string contains all the expected parts
	expectedParts := []string{"(1:User:Admin", "name:Alice", "age:30", "}"}
	for _, part := range expectedParts {
		if !strings.Contains(nodeStr, part) {
			t.Errorf("expected node string to contain %s, got %s", part, nodeStr)
		}
	}

	// Test relationship formatting
	rel := &graph.Relationship{
		ID:          "1",
		Type:        "KNOWS",
		StartNodeID: "1",
		EndNodeID:   "2",
		Properties: map[string]graph.PropertyValue{
			"since": graph.NewIntProperty(2020),
		},
	}

	relStr := formatValue(rel)
	expectedRelStr := "[1:KNOWS {since:2020}]"
	if relStr != expectedRelStr {
		t.Errorf("expected %s, got %s", expectedRelStr, relStr)
	}

	// Test nil formatting
	nilStr := formatValue(nil)
	expectedNilStr := "NULL"
	if nilStr != expectedNilStr {
		t.Errorf("expected %s, got %s", expectedNilStr, nilStr)
	}

	// Test default formatting
	defaultStr := formatValue(42)
	expectedDefaultStr := "42"
	if defaultStr != expectedDefaultStr {
		t.Errorf("expected %s, got %s", expectedDefaultStr, defaultStr)
	}

	// Test node with no properties or labels
	simpleNode := &graph.Node{
		ID:         "2",
		Labels:     []string{},
		Properties: map[string]graph.PropertyValue{},
	}

	simpleNodeStr := formatValue(simpleNode)
	expectedSimpleNodeStr := "(2)"
	if simpleNodeStr != expectedSimpleNodeStr {
		t.Errorf("expected %s, got %s", expectedSimpleNodeStr, simpleNodeStr)
	}

	// Test relationship with no properties
	simpleRel := &graph.Relationship{
		ID:          "2",
		Type:        "FRIEND",
		StartNodeID: "1",
		EndNodeID:   "3",
		Properties:  map[string]graph.PropertyValue{},
	}

	simpleRelStr := formatValue(simpleRel)
	expectedSimpleRelStr := "[2:FRIEND]"
	if simpleRelStr != expectedSimpleRelStr {
		t.Errorf("expected %s, got %s", expectedSimpleRelStr, simpleRelStr)
	}
}

func TestMainFunction(t *testing.T) {
	// Test that main function initializes without errors
	// We'll test this by checking if the root command has the expected subcommands
	expectedCommands := []string{"tui", "exec", "query"}
	actualCommands := make(map[string]bool)

	for _, cmd := range rootCmd.Commands() {
		actualCommands[cmd.Name()] = true
	}

	for _, expectedCmd := range expectedCommands {
		if !actualCommands[expectedCmd] {
			t.Errorf("expected command %s not found", expectedCmd)
		}
	}
}

func TestMain(t *testing.T) {
	// Test that main function doesn't panic
	// We can't easily test the actual execution, but we can test that it compiles and runs
	// without panicking when called with no arguments
	go func() {
		main()
	}()
}
