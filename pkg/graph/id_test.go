package graph

import (
	"testing"
)

func TestGenerateID(t *testing.T) {
	// Set a known counter value for testing
	SetIDCounter(0)

	// Generate first ID
	id1 := GenerateID("node")
	if id1 != "node:b1" {
		t.Errorf("expected 'node:b1', got %s", id1)
	}

	// Generate second ID
	id2 := GenerateID("node")
	if id2 != "node:c2" {
		t.Errorf("expected 'node:c2', got %s", id2)
	}

	// Generate third ID
	id3 := GenerateID("rel")
	if id3 != "rel:d3" {
		t.Errorf("expected 'rel:d3', got %s", id3)
	}

	// Test with different prefix
	id4 := GenerateID("test")
	if id4 != "test:e4" {
		t.Errorf("expected 'test:e4', got %s", id4)
	}
}

func TestSetIDCounter(t *testing.T) {
	// Set counter to a specific value
	SetIDCounter(100)

	// Generate ID and check if it uses the new counter value
	id := GenerateID("node")
	if id != "node:x1" {
		t.Errorf("expected 'node:x1', got %s", id)
	}

	// Set counter to another value
	SetIDCounter(200)

	// Generate another ID
	id2 := GenerateID("node")
	if id2 != "node:t1" {
		t.Errorf("expected 'node:t1', got %s", id2)
	}
}
