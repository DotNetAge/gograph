// Package graph provides core data structures and interfaces for the gograph database.
package graph

import (
	"fmt"
	"strconv"
)

// PropertyValue represents a typed property value that can hold string, int, float, or bool.
type PropertyValue struct {
	String *string
	Int    *int64
	Float  *float64
	Bool   *bool
}

// NewStringProperty creates a PropertyValue holding a string.
func NewStringProperty(v string) PropertyValue {
	return PropertyValue{String: &v}
}

// NewIntProperty creates a PropertyValue holding an int64.
func NewIntProperty(v int64) PropertyValue {
	return PropertyValue{Int: &v}
}

// NewFloatProperty creates a PropertyValue holding a float64.
func NewFloatProperty(v float64) PropertyValue {
	return PropertyValue{Float: &v}
}

// NewBoolProperty creates a PropertyValue holding a bool.
func NewBoolProperty(v bool) PropertyValue {
	return PropertyValue{Bool: &v}
}

// Type returns the type of the property value.
func (p PropertyValue) Type() PropertyType {
	if p.String != nil {
		return PropertyTypeString
	}
	if p.Int != nil {
		return PropertyTypeInt
	}
	if p.Float != nil {
		return PropertyTypeFloat
	}
	if p.Bool != nil {
		return PropertyTypeBool
	}
	return PropertyTypeString
}

// StringValue returns the string value, or empty string if not a string type.
func (p PropertyValue) StringValue() string {
	if p.String != nil {
		return *p.String
	}
	return ""
}

// IntValue returns the int value, or 0 if not an int type.
func (p PropertyValue) IntValue() int64 {
	if p.Int != nil {
		return *p.Int
	}
	return 0
}

// FloatValue returns the float value, or 0 if not a float type.
func (p PropertyValue) FloatValue() float64 {
	if p.Float != nil {
		return *p.Float
	}
	return 0
}

// BoolValue returns the bool value, or false if not a bool type.
func (p PropertyValue) BoolValue() bool {
	if p.Bool != nil {
		return *p.Bool
	}
	return false
}

// ToPropertyValue converts a Go value to a PropertyValue.
func ToPropertyValue(v interface{}) PropertyValue {
	switch val := v.(type) {
	case string:
		return NewStringProperty(val)
	case int:
		return NewIntProperty(int64(val))
	case int64:
		return NewIntProperty(val)
	case float64:
		return NewFloatProperty(val)
	case bool:
		return NewBoolProperty(val)
	default:
		return NewStringProperty(fmt.Sprintf("%v", v))
	}
}

// EncodePropertyValue encodes a PropertyValue to a string for indexing.
func EncodePropertyValue(v PropertyValue) string {
	switch v.Type() {
	case PropertyTypeString:
		return v.StringValue()
	case PropertyTypeInt:
		return strconv.FormatInt(v.IntValue(), 10)
	case PropertyTypeFloat:
		return strconv.FormatFloat(v.FloatValue(), 'f', -1, 64)
	case PropertyTypeBool:
		if v.BoolValue() {
			return "1"
		}
		return "0"
	default:
		return v.StringValue()
	}
}
