// Package graph provides core data structures and interfaces for the gograph database.
package graph

import (
	"github.com/DotNetAge/gograph/pkg/storage"
)

// Index manages label and property indexes for efficient node lookups.
type Index struct {
	store *storage.DB
}

// NewIndex creates a new Index instance for the given storage.
func NewIndex(store *storage.DB) *Index {
	return &Index{store: store}
}

// BuildLabelIndex creates index entries for all labels on a node.
func (idx *Index) BuildLabelIndex(m Mutator, node *Node) error {
	for _, label := range node.Labels {
		key := storage.LabelKey(label, node.ID)
		if err := m.Put(key, []byte(node.ID)); err != nil {
			return err
		}
	}
	return nil
}

// RemoveLabelIndex removes index entries for all labels on a node.
func (idx *Index) RemoveLabelIndex(m Mutator, node *Node) error {
	for _, label := range node.Labels {
		key := storage.LabelKey(label, node.ID)
		if err := m.Delete(key); err != nil {
			return err
		}
	}
	return nil
}

// BuildPropertyIndex creates index entries for all properties on a node.
func (idx *Index) BuildPropertyIndex(m Mutator, node *Node) error {
	for _, label := range node.Labels {
		for propName, propValue := range node.Properties {
			encodedValue := encodePropertyValue(propValue)
			key := storage.PropertyKey(label, propName, encodedValue)
			if err := m.Put(key, []byte(node.ID)); err != nil {
				return err
			}
		}
	}
	return nil
}

// RemovePropertyIndex removes index entries for all properties on a node.
func (idx *Index) RemovePropertyIndex(m Mutator, node *Node) error {
	for _, label := range node.Labels {
		for propName, propValue := range node.Properties {
			encodedValue := encodePropertyValue(propValue)
			key := storage.PropertyKey(label, propName, encodedValue)
			if err := m.Delete(key); err != nil {
				return err
			}
		}
	}
	return nil
}
// LookupByLabel returns all node IDs that have the given label.
func (idx *Index) LookupByLabel(label string) ([]string, error) {
	var nodeIDs []string
	prefix := storage.LabelKeyPrefix(label)
	iter, err := idx.store.NewIter(nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	for iter.SeekGE(prefix); iter.Valid(); iter.Next() {
		key := iter.Key()
		if !hasPrefix(key, prefix) {
			break
		}
		nodeIDs = append(nodeIDs, string(iter.Value()))
	}
	return nodeIDs, nil
}

// LookupByProperty returns all node IDs that have the given label and property value.
func (idx *Index) LookupByProperty(label, propName, propValue string) ([]string, error) {
	var nodeIDs []string
	prefix := storage.PropertyKeyPrefix(label, propName)
	searchKey := storage.PropertyKey(label, propName, propValue)

	iter, err := idx.store.NewIter(nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	for iter.SeekGE(prefix); iter.Valid(); iter.Next() {
		key := iter.Key()
		if !hasPrefix(key, prefix) {
			break
		}
		if string(key) == string(searchKey) {
			nodeIDs = append(nodeIDs, string(iter.Value()))
		}
	}
	return nodeIDs, nil
}

func hasPrefix(key, prefix []byte) bool {
	return len(key) >= len(prefix) && string(key[:len(prefix)]) == string(prefix)
}

func encodePropertyValue(v PropertyValue) string {
	return EncodePropertyValue(v)
}
