package graph

type Node struct {
	ID         string
	Labels     []string
	Properties map[string]PropertyValue
}

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

func (n *Node) HasLabel(label string) bool {
	for _, l := range n.Labels {
		if l == label {
			return true
		}
	}
	return false
}

func (n *Node) GetProperty(key string) (PropertyValue, bool) {
	v, ok := n.Properties[key]
	return v, ok
}

func (n *Node) SetProperty(key string, value PropertyValue) {
	n.Properties[key] = value
}

func (n *Node) RemoveProperty(key string) {
	delete(n.Properties, key)
}

func (n *Node) RemoveLabel(label string) {
	newLabels := make([]string, 0, len(n.Labels))
	for _, l := range n.Labels {
		if l != label {
			newLabels = append(newLabels, l)
		}
	}
	n.Labels = newLabels
}
