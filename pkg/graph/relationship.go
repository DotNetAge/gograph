// Package graph provides core data structures and interfaces for the gograph database.
package graph

// Relationship represents a directed relationship between two nodes in a graph.
// Relationships connect nodes and can also have their own properties, allowing
// you to model complex connections with attributes.
//
// Each relationship has:
//   - A unique identifier (ID)
//   - A start node ID (where the relationship originates)
//   - An end node ID (where the relationship points to)
//   - A type that categorizes the relationship
//   - Properties that store data as key-value pairs
//
// Relationships are directed, meaning they have a clear direction from
// start node to end node. For bidirectional relationships, you typically
// create two relationships or query in both directions.
type Relationship struct {
	// ID is the unique identifier for this relationship.
	ID string

	// StartNodeID is the ID of the node where this relationship originates.
	StartNodeID string

	// EndNodeID is the ID of the node where this relationship points to.
	EndNodeID string

	// Type is the category of this relationship (e.g., "KNOWS", "WORKS_AT").
	Type string

	// Properties are the key-value pairs associated with this relationship.
	// Values are stored as PropertyValue which supports string, int, float, and bool.
	Properties map[string]PropertyValue
}

// NewRelationship creates a new Relationship with the given parameters.
//
// Parameters:
//   - startNodeID: The ID of the node where the relationship originates
//   - endNodeID: The ID of the node where the relationship points to
//   - relType: The type/category of the relationship (e.g., "KNOWS")
//   - properties: A map of property names to values. Values can be string, int,
//     int64, float64, or bool. Other types will be converted to strings.
//
// Returns a new Relationship with a generated unique ID.
//
// Example:
//
//	rel := graph.NewRelationship(
//	    alice.ID,    // start node
//	    bob.ID,      // end node
//	    "KNOWS",     // relationship type
//	    map[string]interface{}{
//	        "since": "2020-01-01",
//	        "strength": 0.8,
//	    },
//	)
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

// GetProperty returns the property value and true if it exists.
//
// Parameters:
//   - key: The property name to look up
//
// Returns the PropertyValue and true if found, or zero value and false if not found.
//
// Example:
//
//	if value, ok := rel.GetProperty("since"); ok {
//	    fmt.Printf("Relationship started: %s\n", value.StringValue())
//	}
func (r *Relationship) GetProperty(key string) (PropertyValue, bool) {
	v, ok := r.Properties[key]
	return v, ok
}

// SetProperty sets a property value on the relationship.
// If the property already exists, it will be overwritten.
//
// Parameters:
//   - key: The property name
//   - value: The PropertyValue to set
//
// Example:
//
//	rel.SetProperty("strength", graph.NewFloatProperty(0.9))
func (r *Relationship) SetProperty(key string, value PropertyValue) {
	r.Properties[key] = value
}

// RemoveProperty removes a property from the relationship.
// If the property doesn't exist, this is a no-op.
//
// Parameters:
//   - key: The property name to remove
//
// Example:
//
//	rel.RemoveProperty("temporary")
func (r *Relationship) RemoveProperty(key string) {
	delete(r.Properties, key)
}
