// Package storage provides the底层 storage layer for gograph using Pebble as the
// underlying key-value store. It handles data marshaling, key generation,
// and basic CRUD operations.
package storage

const (
	KeyPrefixNode        = "node:"
	KeyPrefixRel         = "rel:"
	KeyPrefixLabel       = "label:"
	KeyPrefixProp        = "prop:"
	KeyPrefixAdj         = "adj:"
	KeyPrefixMeta        = "meta:"
	KeyPrefixIndexCount  = "meta:index_count"
)

// NodeKey returns the storage key for a node with the given ID.
func NodeKey(nodeID string) []byte {
	return []byte(KeyPrefixNode + nodeID)
}

// RelKey returns the storage key for a relationship with the given ID.
func RelKey(relID string) []byte {
	return []byte(KeyPrefixRel + relID)
}

// LabelKey returns the storage key for a label index entry.
func LabelKey(labelName, nodeID string) []byte {
	return []byte(KeyPrefixLabel + labelName + ":" + nodeID)
}

// LabelKeyPrefix returns the key prefix for all entries with the given label.
func LabelKeyPrefix(labelName string) []byte {
	return []byte(KeyPrefixLabel + labelName + ":")
}

// PropertyKey returns the storage key for a property index entry.
func PropertyKey(labelName, propName, propValue string) []byte {
	return []byte(KeyPrefixProp + labelName + ":" + propName + ":" + propValue)
}

// PropertyKeyPrefix returns the key prefix for all entries with the given label and property name.
func PropertyKeyPrefix(labelName, propName string) []byte {
	return []byte(KeyPrefixProp + labelName + ":" + propName + ":")
}

// AdjKey returns the storage key for an adjacency entry.
func AdjKey(nodeID, relType, direction, relID string) []byte {
	return []byte(KeyPrefixAdj + nodeID + ":" + relType + ":" + direction + ":" + relID)
}

// AdjKeyPrefix returns the key prefix for all adjacency entries of a node.
func AdjKeyPrefix(nodeID string) []byte {
	return []byte(KeyPrefixAdj + nodeID + ":")
}

// AdjKeyPrefixNodeAndType returns the key prefix for all adjacency entries of a node with a specific relationship type.
func AdjKeyPrefixNodeAndType(nodeID, relType string) []byte {
	return []byte(KeyPrefixAdj + nodeID + ":" + relType + ":")
}

// AdjKeyPrefixNodeAndTypeAndDir returns the key prefix for all adjacency entries of a node with a specific relationship type and direction.
func AdjKeyPrefixNodeAndTypeAndDir(nodeID, relType, direction string) []byte {
	return []byte(KeyPrefixAdj + nodeID + ":" + relType + ":" + direction + ":")
}
