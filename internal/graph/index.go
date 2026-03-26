package graph

import (
	"github.com/DotNetAge/gograph/internal/storage"
)

type Index struct {
	store *storage.DB
}

func NewIndex(store *storage.DB) *Index {
	return &Index{store: store}
}

func (idx *Index) BuildLabelIndex(node *Node) error {
	for _, label := range node.Labels {
		key := storage.LabelKey(label, node.ID)
		if err := idx.store.Put(key, []byte(node.ID)); err != nil {
			return err
		}
	}
	return nil
}

func (idx *Index) RemoveLabelIndex(node *Node) error {
	for _, label := range node.Labels {
		key := storage.LabelKey(label, node.ID)
		if err := idx.store.Delete(key); err != nil {
			return err
		}
	}
	return nil
}

func (idx *Index) BuildPropertyIndex(node *Node) error {
	for _, label := range node.Labels {
		for propName, propValue := range node.Properties {
			encodedValue := encodePropertyValue(propValue)
			key := storage.PropertyKey(label, propName, encodedValue)
			if err := idx.store.Put(key, []byte(node.ID)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (idx *Index) RemovePropertyIndex(node *Node) error {
	for _, label := range node.Labels {
		for propName, propValue := range node.Properties {
			encodedValue := encodePropertyValue(propValue)
			key := storage.PropertyKey(label, propName, encodedValue)
			if err := idx.store.Delete(key); err != nil {
				return err
			}
		}
	}
	return nil
}

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
