package graph

import (
	"strings"
	"testing"
)

func TestGenerateID(t *testing.T) {
	// Generate IDs and verify format (prefix:ULID)
	id1 := GenerateID("node")
	if !strings.HasPrefix(id1, "node:") {
		t.Errorf("expected ID to start with 'node:', got %s", id1)
	}
	if len(id1) <= len("node:") {
		t.Errorf("expected ID to have ULID suffix, got %s", id1)
	}

	// Generate second ID - should be different
	id2 := GenerateID("node")
	if id1 == id2 {
		t.Errorf("expected unique IDs, got same ID: %s", id1)
	}

	// Generate third ID with different prefix
	id3 := GenerateID("rel")
	if !strings.HasPrefix(id3, "rel:") {
		t.Errorf("expected ID to start with 'rel:', got %s", id3)
	}

	// Test with different prefix
	id4 := GenerateID("test")
	if !strings.HasPrefix(id4, "test:") {
		t.Errorf("expected ID to start with 'test:', got %s", id4)
	}
}

func TestSetIDCounter(t *testing.T) {
	// SetIDCounter is deprecated and should be a no-op with ULID generation.
	// This test verifies that IDs are still generated correctly after calling it.
	SetIDCounter(100)

	id := GenerateID("node")
	if !strings.HasPrefix(id, "node:") {
		t.Errorf("expected ID to start with 'node:', got %s", id)
	}

	SetIDCounter(200)

	id2 := GenerateID("node")
	if !strings.HasPrefix(id2, "node:") {
		t.Errorf("expected ID to start with 'node:', got %s", id2)
	}
	if id == id2 {
		t.Errorf("expected unique IDs after counter reset, got same ID: %s", id)
	}
}
