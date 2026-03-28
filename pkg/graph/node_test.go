package graph

import (
	"testing"
)

func TestNewNode(t *testing.T) {
	// Set a known counter value for testing
	SetIDCounter(0)

	// Test creating a node with labels and properties
	labels := []string{"User", "Admin"}
	properties := map[string]interface{}{
		"name": "Alice",
		"age":  30,
		"active": true,
		"score": 95.5,
	}

	node := NewNode(labels, properties)

	// Check that the node has the correct ID
	if node.ID != "node:b1" {
		t.Errorf("expected 'node:b1', got %s", node.ID)
	}

	// Check that the node has the correct labels
	if len(node.Labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(node.Labels))
	}

	if node.Labels[0] != "User" {
		t.Errorf("expected first label 'User', got %s", node.Labels[0])
	}

	if node.Labels[1] != "Admin" {
		t.Errorf("expected second label 'Admin', got %s", node.Labels[1])
	}

	// Check that the node has the correct properties
	if len(node.Properties) != 4 {
		t.Errorf("expected 4 properties, got %d", len(node.Properties))
	}

	// Check string property
	if nameProp, ok := node.Properties["name"]; ok {
		if nameProp.Type() != PropertyTypeString {
			t.Errorf("expected name property to be string, got %v", nameProp.Type())
		}
		if nameProp.StringValue() != "Alice" {
			t.Errorf("expected name to be 'Alice', got %s", nameProp.StringValue())
		}
	} else {
		t.Error("expected 'name' property to exist")
	}

	// Check int property
	if ageProp, ok := node.Properties["age"]; ok {
		if ageProp.Type() != PropertyTypeInt {
			t.Errorf("expected age property to be int, got %v", ageProp.Type())
		}
		if ageProp.IntValue() != 30 {
			t.Errorf("expected age to be 30, got %d", ageProp.IntValue())
		}
	} else {
		t.Error("expected 'age' property to exist")
	}

	// Check bool property
	if activeProp, ok := node.Properties["active"]; ok {
		if activeProp.Type() != PropertyTypeBool {
			t.Errorf("expected active property to be bool, got %v", activeProp.Type())
		}
		if activeProp.BoolValue() != true {
			t.Errorf("expected active to be true, got %v", activeProp.BoolValue())
		}
	} else {
		t.Error("expected 'active' property to exist")
	}

	// Check float property
	if scoreProp, ok := node.Properties["score"]; ok {
		if scoreProp.Type() != PropertyTypeFloat {
			t.Errorf("expected score property to be float, got %v", scoreProp.Type())
		}
		if scoreProp.FloatValue() != 95.5 {
			t.Errorf("expected score to be 95.5, got %f", scoreProp.FloatValue())
		}
	} else {
		t.Error("expected 'score' property to exist")
	}
}

func TestNewNodeWithNoProperties(t *testing.T) {
	// Set a known counter value for testing
	SetIDCounter(100)

	// Test creating a node with no properties
	labels := []string{"User"}
	properties := map[string]interface{}{}

	node := NewNode(labels, properties)

	// Check that the node has the correct ID
	if node.ID != "node:x1" {
		t.Errorf("expected 'node:x1', got %s", node.ID)
	}

	// Check that the node has the correct labels
	if len(node.Labels) != 1 {
		t.Errorf("expected 1 label, got %d", len(node.Labels))
	}

	if node.Labels[0] != "User" {
		t.Errorf("expected label 'User', got %s", node.Labels[0])
	}

	// Check that the node has no properties
	if len(node.Properties) != 0 {
		t.Errorf("expected 0 properties, got %d", len(node.Properties))
	}
}

func TestNewNodeWithNoLabels(t *testing.T) {
	// Set a known counter value for testing
	SetIDCounter(200)

	// Test creating a node with no labels
	labels := []string{}
	properties := map[string]interface{}{
		"name": "Bob",
	}

	node := NewNode(labels, properties)

	// Check that the node has the correct ID
	if node.ID != "node:t1" {
		t.Errorf("expected 'node:t1', got %s", node.ID)
	}

	// Check that the node has no labels
	if len(node.Labels) != 0 {
		t.Errorf("expected 0 labels, got %d", len(node.Labels))
	}

	// Check that the node has the correct property
	if len(node.Properties) != 1 {
		t.Errorf("expected 1 property, got %d", len(node.Properties))
	}

	if nameProp, ok := node.Properties["name"]; ok {
		if nameProp.StringValue() != "Bob" {
			t.Errorf("expected name to be 'Bob', got %s", nameProp.StringValue())
		}
	} else {
		t.Error("expected 'name' property to exist")
	}
}
