// Package storage provides the底层 storage layer for gograph using Pebble as the
// underlying key-value store. It handles data marshaling, key generation,
// and basic CRUD operations.
package storage

// Key prefixes for different types of data in the storage layer.
// These prefixes ensure that different data types are stored in separate
// key ranges for efficient iteration and querying.
const (
	// KeyPrefixNode is the prefix for node data keys.
	// Format: "node:<nodeID>"
	KeyPrefixNode = "node:"

	// KeyPrefixRel is the prefix for relationship data keys.
	// Format: "rel:<relID>"
	KeyPrefixRel = "rel:"

	// KeyPrefixLabel is the prefix for label index keys.
	// Format: "label:<labelName>:<nodeID>"
	KeyPrefixLabel = "label:"

	// KeyPrefixProp is the prefix for property index keys.
	// Format: "prop:<labelName>:<propName>:<propValue>:<nodeID>"
	KeyPrefixProp = "prop:"

	// KeyPrefixAdj is the prefix for adjacency list keys.
	// Format: "adj:<nodeID>:<relType>:<direction>:<relID>"
	KeyPrefixAdj = "adj:"

	// KeyPrefixMeta is the prefix for metadata keys.
	KeyPrefixMeta = "meta:"

	// KeyPrefixIndexCount is the key for storing the index count.
	KeyPrefixIndexCount = "meta:index_count"
)

// NodeKey returns the storage key for a node with the given ID.
//
// Parameters:
//   - nodeID: The unique identifier of the node
//
// Returns the storage key as a byte slice.
//
// Example:
//
//	key := storage.NodeKey("node:a1")
//	// key = []byte("node:node:a1")
func NodeKey(nodeID string) []byte {
	return []byte(KeyPrefixNode + nodeID)
}

// RelKey returns the storage key for a relationship with the given ID.
//
// Parameters:
//   - relID: The unique identifier of the relationship
//
// Returns the storage key as a byte slice.
//
// Example:
//
//	key := storage.RelKey("rel:b2")
//	// key = []byte("rel:rel:b2")
func RelKey(relID string) []byte {
	return []byte(KeyPrefixRel + relID)
}

// LabelKey returns the storage key for a label index entry.
//
// Parameters:
//   - labelName: The name of the label
//   - nodeID: The ID of the node with this label
//
// Returns the storage key as a byte slice.
//
// Example:
//
//	key := storage.LabelKey("Person", "node:a1")
//	// key = []byte("label:Person:node:a1")
func LabelKey(labelName, nodeID string) []byte {
	return []byte(KeyPrefixLabel + labelName + ":" + nodeID)
}

// LabelKeyPrefix returns the key prefix for all entries with the given label.
// This is used for iterating over all nodes with a specific label.
//
// Parameters:
//   - labelName: The name of the label
//
// Returns the key prefix as a byte slice.
//
// Example:
//
//	prefix := storage.LabelKeyPrefix("Person")
//	// prefix = []byte("label:Person:")
func LabelKeyPrefix(labelName string) []byte {
	return []byte(KeyPrefixLabel + labelName + ":")
}

// PropertyKey returns the storage key for a property index entry.
//
// Parameters:
//   - labelName: The name of the label
//   - propName: The name of the property
//   - propValue: The value of the property (encoded as string)
//
// Returns the storage key as a byte slice.
//
// Example:
//
//	key := storage.PropertyKey("Person", "name", "Alice")
//	// key = []byte("prop:Person:name:Alice")
func PropertyKey(labelName, propName, propValue string) []byte {
	return []byte(KeyPrefixProp + labelName + ":" + propName + ":" + propValue)
}

// PropertyKeyPrefix returns the key prefix for all entries with the given label and property name.
// This is used for iterating over all nodes with a specific label and property.
//
// Parameters:
//   - labelName: The name of the label
//   - propName: The name of the property
//
// Returns the key prefix as a byte slice.
//
// Example:
//
//	prefix := storage.PropertyKeyPrefix("Person", "name")
//	// prefix = []byte("prop:Person:name:")
func PropertyKeyPrefix(labelName, propName string) []byte {
	return []byte(KeyPrefixProp + labelName + ":" + propName + ":")
}

// AdjKey returns the storage key for an adjacency entry.
//
// Parameters:
//   - nodeID: The ID of the node
//   - relType: The type of relationship
//   - direction: The direction ("out" or "in")
//   - relID: The ID of the relationship
//
// Returns the storage key as a byte slice.
//
// Example:
//
//	key := storage.AdjKey("node:a1", "KNOWS", "out", "rel:b2")
//	// key = []byte("adj:node:a1:KNOWS:out:rel:b2")
func AdjKey(nodeID, relType, direction, relID string) []byte {
	return []byte(KeyPrefixAdj + nodeID + ":" + relType + ":" + direction + ":" + relID)
}

// AdjKeyPrefix returns the key prefix for all adjacency entries of a node.
// This is used for finding all relationships connected to a node.
//
// Parameters:
//   - nodeID: The ID of the node
//
// Returns the key prefix as a byte slice.
//
// Example:
//
//	prefix := storage.AdjKeyPrefix("node:a1")
//	// prefix = []byte("adj:node:a1:")
func AdjKeyPrefix(nodeID string) []byte {
	return []byte(KeyPrefixAdj + nodeID + ":")
}

// AdjKeyPrefixNodeAndType returns the key prefix for all adjacency entries of a node with a specific relationship type.
// This is used for finding all relationships of a specific type connected to a node.
//
// Parameters:
//   - nodeID: The ID of the node
//   - relType: The type of relationship
//
// Returns the key prefix as a byte slice.
//
// Example:
//
//	prefix := storage.AdjKeyPrefixNodeAndType("node:a1", "KNOWS")
//	// prefix = []byte("adj:node:a1:KNOWS:")
func AdjKeyPrefixNodeAndType(nodeID, relType string) []byte {
	return []byte(KeyPrefixAdj + nodeID + ":" + relType + ":")
}

// AdjKeyPrefixNodeAndTypeAndDir returns the key prefix for all adjacency entries of a node with a specific relationship type and direction.
// This is used for finding relationships in a specific direction.
//
// Parameters:
//   - nodeID: The ID of the node
//   - relType: The type of relationship
//   - direction: The direction ("out" or "in")
//
// Returns the key prefix as a byte slice.
//
// Example:
//
//	prefix := storage.AdjKeyPrefixNodeAndTypeAndDir("node:a1", "KNOWS", "out")
//	// prefix = []byte("adj:node:a1:KNOWS:out:")
func AdjKeyPrefixNodeAndTypeAndDir(nodeID, relType, direction string) []byte {
	return []byte(KeyPrefixAdj + nodeID + ":" + relType + ":" + direction + ":")
}
