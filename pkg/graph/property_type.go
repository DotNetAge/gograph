// Package graph provides core data structures and interfaces for the gograph database.
package graph

// PropertyType represents the type of a property value.
type PropertyType string

const (
	PropertyTypeString PropertyType = "string"
	PropertyTypeInt    PropertyType = "int"
	PropertyTypeFloat  PropertyType = "float"
	PropertyTypeBool   PropertyType = "bool"
	PropertyTypeList   PropertyType = "list"
)
