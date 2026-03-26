package graph

import (
	"testing"
)

func TestNode(t *testing.T) {
	node := &Node{
		ID:     "node:1",
		Labels: []string{"User", "Admin"},
		Properties: map[string]PropertyValue{
			"name": NewStringProperty("Alice"),
			"age":  NewIntProperty(30),
		},
	}

	if !node.HasLabel("User") {
		t.Error("expected label User")
	}
	if !node.HasLabel("Admin") {
		t.Error("expected label Admin")
	}
	if node.HasLabel("Guest") {
		t.Error("did not expect label Guest")
	}

	prop, ok := node.GetProperty("name")
	if !ok || prop.StringValue() != "Alice" {
		t.Error("expected property name to be Alice")
	}

	node.SetProperty("age", NewIntProperty(31))
	prop, _ = node.GetProperty("age")
	if prop.IntValue() != 31 {
		t.Error("expected age to be 31")
	}

	node.RemoveLabel("Admin")
	if node.HasLabel("Admin") {
		t.Error("did not expect label Admin after removal")
	}

	node.RemoveProperty("name")
	_, ok = node.GetProperty("name")
	if ok {
		t.Error("did not expect property name after removal")
	}
}

func TestRelationship(t *testing.T) {
	rel := NewRelationship("node:1", "node:2", "KNOWS", map[string]interface{}{"since": 2020})

	if rel.StartNodeID != "node:1" {
		t.Error("expected start node node:1")
	}
	if rel.EndNodeID != "node:2" {
		t.Error("expected end node node:2")
	}
	if rel.Type != "KNOWS" {
		t.Error("expected type KNOWS")
	}

	prop, ok := rel.GetProperty("since")
	if !ok || prop.IntValue() != 2020 {
		t.Error("expected property since to be 2020")
	}

	rel.SetProperty("since", NewIntProperty(2021))
	prop, _ = rel.GetProperty("since")
	if prop.IntValue() != 2021 {
		t.Error("expected since to be 2021")
	}

	rel.RemoveProperty("since")
	_, ok = rel.GetProperty("since")
	if ok {
		t.Error("did not expect property since after removal")
	}
}

func TestPropertyValue(t *testing.T) {
	s := NewStringProperty("test")
	if s.Type() != PropertyTypeString || s.StringValue() != "test" {
		t.Error("invalid string property")
	}

	i := NewIntProperty(123)
	if i.Type() != PropertyTypeInt || i.IntValue() != 123 {
		t.Error("invalid int property")
	}

	f := NewFloatProperty(3.14)
	if f.Type() != PropertyTypeFloat || f.FloatValue() != 3.14 {
		t.Error("invalid float property")
	}

	b := NewBoolProperty(true)
	if b.Type() != PropertyTypeBool || !b.BoolValue() {
		t.Error("invalid bool property")
	}

	// Test ToPropertyValue
	if ToPropertyValue("hello").StringValue() != "hello" {
		t.Error("ToPropertyValue failed for string")
	}
	if ToPropertyValue(int64(100)).IntValue() != 100 {
		t.Error("ToPropertyValue failed for int64")
	}
	if ToPropertyValue(3.14).FloatValue() != 3.14 {
		t.Error("ToPropertyValue failed for float64")
	}
	if !ToPropertyValue(true).BoolValue() {
		t.Error("ToPropertyValue failed for bool")
	}

	// Test EncodePropertyValue
	if EncodePropertyValue(NewStringProperty("abc")) != "abc" {
		t.Error("EncodePropertyValue failed for string")
	}
	if EncodePropertyValue(NewIntProperty(123)) != "123" {
		t.Error("EncodePropertyValue failed for int")
	}
	if EncodePropertyValue(NewBoolProperty(true)) != "1" {
		t.Error("EncodePropertyValue failed for bool true")
	}
	if EncodePropertyValue(NewBoolProperty(false)) != "0" {
		t.Error("EncodePropertyValue failed for bool false")
	}
}
