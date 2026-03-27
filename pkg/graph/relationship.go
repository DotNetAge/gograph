// Package graph provides core data structures and interfaces for the gograph database.
package graph

// Relationship represents a directed relationship between two nodes in a graph.
type Relationship struct {
	ID         string
	StartNodeID string
	EndNodeID   string
	Type       string
	Properties map[string]PropertyValue
}

// NewRelationship creates a new Relationship with the given parameters.
func NewRelationship(startNodeID, endNodeID, relType string, properties map[string]interface{}) *Relationship {
	props := make(map[string]PropertyValue)
	for k, v := range properties {
		props[k] = ToPropertyValue(v)
	}
	return &Relationship{
		ID:          GenerateID("rel"),
		StartNodeID: startNodeID,
		EndNodeID:   endNodeID,
		Type:        relType,
		Properties:  props,
	}
}

// GetProperty returns the property value and true if it exists, or zero value and false.
func (r *Relationship) GetProperty(key string) (PropertyValue, bool) {
	v, ok := r.Properties[key]
	return v, ok
}

// SetProperty sets a property value on the relationship.
func (r *Relationship) SetProperty(key string, value PropertyValue) {
	r.Properties[key] = value
}

// RemoveProperty removes a property from the relationship.
func (r *Relationship) RemoveProperty(key string) {
	delete(r.Properties, key)
}
