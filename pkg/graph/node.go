// Package graph provides core data structures and interfaces for the gograph database.
// It implements a property graph model where data is organized as nodes connected
// by relationships, with both nodes and relationships capable of holding properties.
//
// Graph Model:
//
//	┌─────────────┐         ┌─────────────┐
//	│    Node     │◀───────▶│  Relationship
//	│  (Person)   │         │   (KNOWS)   │
//	├─────────────┤         ├─────────────┤
//	│ ID          │         │ ID          │
//	│ Labels      │         │ Type        │
//	│ Properties  │         │ StartNodeID │
//	└─────────────┘         │ EndNodeID   │
//	                        │ Properties  │
//	                        └─────────────┘
//
// Basic Usage:
//
//	// Create a node with labels and properties
//	node := graph.NewNode(
//	    []string{"Person"},
//	    map[string]interface{}{
//	        "name": "Alice",
//	        "age":  30,
//	    },
//	)
//
//	// Check if node has a label
//	if node.HasLabel("Person") {
//	    fmt.Println("This is a person node")
//	}
//
//	// Get and set properties
//	name, _ := node.GetProperty("name")
//	node.SetProperty("age", graph.NewIntProperty(31))
//
// Thread Safety:
//
// Node and Relationship types are not thread-safe. If you need to access
// them concurrently, use external synchronization mechanisms.
package graph

// Node represents a graph node with an ID, labels, and properties.
// Nodes are the primary entities in a graph database, representing
// objects such as people, places, or things.
//
// Each node has:
//   - A unique identifier (ID)
//   - One or more labels that categorize the node
//   - A map of properties that store data as key-value pairs
type Node struct {
	// ID is the unique identifier for this node.
	ID string

	// Labels are the categories or types this node belongs to.
	// A node can have multiple labels (e.g., ["Person", "Employee"]).
	Labels []string

	// Properties are the key-value pairs associated with this node.
	// Values are stored as PropertyValue which supports string, int, float, and bool.
	Properties map[string]PropertyValue
}

// NewNode creates a new Node with the given labels and properties.
//
// Parameters:
//   - labels: The categories or types for this node (e.g., ["Person"])
//   - properties: A map of property names to values. Values can be string, int,
//     int64, float64, or bool. Other types will be converted to strings.
//
// Returns a new Node with a generated unique ID.
//
// Example:
//
//	node := graph.NewNode(
//	    []string{"Person"},
//	    map[string]interface{}{
//	        "name":    "Alice",
//	        "age":     30,
//	        "active":  true,
//	        "balance": 1234.56,
//	    },
//	)
func NewNode(labels []string, properties map[string]interface{}) *Node {
	props := make(map[string]PropertyValue)
	for k, v := range properties {
		props[k] = ToPropertyValue(v)
	}
	return &Node{
		ID:         GenerateID("node"),
		Labels:     labels,
		Properties: props,
	}
}

// HasLabel returns true if the node has the given label.
//
// Parameters:
//   - label: The label to check for
//
// Returns true if the label exists in the node's Labels slice.
//
// Example:
//
//	if node.HasLabel("Person") {
//	    // Handle person node
//	}
func (n *Node) HasLabel(label string) bool {
	for _, l := range n.Labels {
		if l == label {
			return true
		}
	}
	return false
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
//	if value, ok := node.GetProperty("name"); ok {
//	    fmt.Printf("Name: %s\n", value.StringValue())
//	}
func (n *Node) GetProperty(key string) (PropertyValue, bool) {
	v, ok := n.Properties[key]
	return v, ok
}

// SetProperty sets a property value on the node.
// If the property already exists, it will be overwritten.
//
// Parameters:
//   - key: The property name
//   - value: The PropertyValue to set
//
// Example:
//
//	node.SetProperty("age", graph.NewIntProperty(31))
func (n *Node) SetProperty(key string, value PropertyValue) {
	n.Properties[key] = value
}

// RemoveProperty removes a property from the node.
// If the property doesn't exist, this is a no-op.
//
// Parameters:
//   - key: The property name to remove
//
// Example:
//
//	node.RemoveProperty("temporary")
func (n *Node) RemoveProperty(key string) {
	delete(n.Properties, key)
}

// RemoveLabel removes a label from the node.
// If the label doesn't exist, the Labels slice remains unchanged.
//
// Parameters:
//   - label: The label to remove
//
// Example:
//
//	node.RemoveLabel("Temporary")
func (n *Node) RemoveLabel(label string) {
	newLabels := make([]string, 0, len(n.Labels))
	for _, l := range n.Labels {
		if l != label {
			newLabels = append(newLabels, l)
		}
	}
	n.Labels = newLabels
}
