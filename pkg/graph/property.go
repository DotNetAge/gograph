// Package graph provides core data structures and interfaces for the gograph database.
package graph

import (
	"fmt"
	"strconv"
)

// PropertyValue represents a typed property value that can hold string, int, float, or bool.
// It uses a union-style structure where only one field is populated at a time.
// Use the New*Property constructors to create PropertyValue instances.
//
// PropertyValue supports:
//   - String values
//   - Integer values (int64)
//   - Floating point values (float64)
//   - Boolean values
//
// Example:
//
//	// Create different types of property values
//	name := graph.NewStringProperty("Alice")
//	age := graph.NewIntProperty(30)
//	balance := graph.NewFloatProperty(1234.56)
//	active := graph.NewBoolProperty(true)
//
//	// Check the type
//	switch name.Type() {
//	case graph.PropertyTypeString:
//	    fmt.Println("It's a string")
//	case graph.PropertyTypeInt:
//	    fmt.Println("It's an integer")
//	}
type PropertyValue struct {
	// String holds the value if this is a string property.
	String *string

	// Int holds the value if this is an integer property.
	Int *int64

	// Float holds the value if this is a float property.
	Float *float64

	// Bool holds the value if this is a boolean property.
	Bool *bool
}

// NewStringProperty creates a PropertyValue holding a string.
//
// Parameters:
//   - v: The string value to store
//
// Returns a PropertyValue containing the string.
//
// Example:
//
//	prop := graph.NewStringProperty("hello")
func NewStringProperty(v string) PropertyValue {
	return PropertyValue{String: &v}
}

// NewIntProperty creates a PropertyValue holding an int64.
//
// Parameters:
//   - v: The int64 value to store
//
// Returns a PropertyValue containing the integer.
//
// Example:
//
//	prop := graph.NewIntProperty(42)
func NewIntProperty(v int64) PropertyValue {
	return PropertyValue{Int: &v}
}

// NewFloatProperty creates a PropertyValue holding a float64.
//
// Parameters:
//   - v: The float64 value to store
//
// Returns a PropertyValue containing the float.
//
// Example:
//
//	prop := graph.NewFloatProperty(3.14)
func NewFloatProperty(v float64) PropertyValue {
	return PropertyValue{Float: &v}
}

// NewBoolProperty creates a PropertyValue holding a bool.
//
// Parameters:
//   - v: The bool value to store
//
// Returns a PropertyValue containing the boolean.
//
// Example:
//
//	prop := graph.NewBoolProperty(true)
func NewBoolProperty(v bool) PropertyValue {
	return PropertyValue{Bool: &v}
}

// Type returns the type of the property value.
// It checks which field is populated and returns the corresponding type.
// If no field is populated, it returns PropertyTypeString as a default.
//
// Returns the PropertyType of this value.
//
// Example:
//
//	prop := graph.NewIntProperty(42)
//	fmt.Println(prop.Type()) // Output: int
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
//
// Returns the string value if this PropertyValue holds a string, otherwise "".
//
// Example:
//
//	prop := graph.NewStringProperty("hello")
//	fmt.Println(prop.StringValue()) // Output: hello
func (p PropertyValue) StringValue() string {
	if p.String != nil {
		return *p.String
	}
	return ""
}

// IntValue returns the int value, or 0 if not an int type.
//
// Returns the int64 value if this PropertyValue holds an integer, otherwise 0.
//
// Example:
//
//	prop := graph.NewIntProperty(42)
//	fmt.Println(prop.IntValue()) // Output: 42
func (p PropertyValue) IntValue() int64 {
	if p.Int != nil {
		return *p.Int
	}
	return 0
}

// FloatValue returns the float value, or 0 if not a float type.
//
// Returns the float64 value if this PropertyValue holds a float, otherwise 0.
//
// Example:
//
//	prop := graph.NewFloatProperty(3.14)
//	fmt.Println(prop.FloatValue()) // Output: 3.14
func (p PropertyValue) FloatValue() float64 {
	if p.Float != nil {
		return *p.Float
	}
	return 0
}

// BoolValue returns the bool value, or false if not a bool type.
//
// Returns the bool value if this PropertyValue holds a boolean, otherwise false.
//
// Example:
//
//	prop := graph.NewBoolProperty(true)
//	fmt.Println(prop.BoolValue()) // Output: true
func (p PropertyValue) BoolValue() bool {
	if p.Bool != nil {
		return *p.Bool
	}
	return false
}

// ToPropertyValue converts a Go value to a PropertyValue.
// It supports the following types:
//   - string: Stored as PropertyTypeString
//   - int, int64: Stored as PropertyTypeInt
//   - float64: Stored as PropertyTypeFloat
//   - bool: Stored as PropertyTypeBool
//   - Other types: Converted to string using fmt.Sprintf
//
// Parameters:
//   - v: The value to convert
//
// Returns a PropertyValue containing the converted value.
//
// Example:
//
//	// Various conversions
//	strProp := graph.ToPropertyValue("hello")
//	intProp := graph.ToPropertyValue(42)
//	floatProp := graph.ToPropertyValue(3.14)
//	boolProp := graph.ToPropertyValue(true)
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
// This is used internally for creating index keys.
//
// Parameters:
//   - v: The PropertyValue to encode
//
// Returns a string representation suitable for indexing.
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
