package cypher

import (
	"testing"
)

func TestAddAffected(t *testing.T) {
	result := &Result{}

	// Test initial state
	if result.AffectedNodes != 0 {
		t.Errorf("expected initial AffectedNodes to be 0, got %d", result.AffectedNodes)
	}
	if result.AffectedRels != 0 {
		t.Errorf("expected initial AffectedRels to be 0, got %d", result.AffectedRels)
	}

	// Test adding affected nodes and relationships
	result.AddAffected(2, 3)
	if result.AffectedNodes != 2 {
		t.Errorf("expected AffectedNodes to be 2, got %d", result.AffectedNodes)
	}
	if result.AffectedRels != 3 {
		t.Errorf("expected AffectedRels to be 3, got %d", result.AffectedRels)
	}

	// Test adding more affected nodes and relationships
	result.AddAffected(1, 2)
	if result.AffectedNodes != 3 {
		t.Errorf("expected AffectedNodes to be 3, got %d", result.AffectedNodes)
	}
	if result.AffectedRels != 5 {
		t.Errorf("expected AffectedRels to be 5, got %d", result.AffectedRels)
	}

	// Test adding zero values
	result.AddAffected(0, 0)
	if result.AffectedNodes != 3 {
		t.Errorf("expected AffectedNodes to remain 3, got %d", result.AffectedNodes)
	}
	if result.AffectedRels != 5 {
		t.Errorf("expected AffectedRels to remain 5, got %d", result.AffectedRels)
	}
}
