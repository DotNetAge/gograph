// Package graph provides core data structures and interfaces for the gograph database.
package graph

// Node represents a graph node with an ID, labels, and properties.
type Node struct {
	ID         string
	Labels     []string
	Properties map[string]PropertyValue
}

// NewNode creates a new Node with the given labels and properties.
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
func (n *Node) HasLabel(label string) bool {
	for _, l := range n.Labels {
		if l == label {
			return true
		}
	}
	return false
}

// GetProperty returns the property value and true if it exists, or zero value and false.
func (n *Node) GetProperty(key string) (PropertyValue, bool) {
	v, ok := n.Properties[key]
	return v, ok
}

// SetProperty sets a property value on the node.
func (n *Node) SetProperty(key string, value PropertyValue) {
	n.Properties[key] = value
}

// RemoveProperty removes a property from the node.
func (n *Node) RemoveProperty(key string) {
	delete(n.Properties, key)
}

// RemoveLabel removes a label from the node.
func (n *Node) RemoveLabel(label string) {
	newLabels := make([]string, 0, len(n.Labels))
	for _, l := range n.Labels {
		if l != label {
			newLabels = append(newLabels, l)
		}
	}
	n.Labels = newLabels
}
