package creators

import (
	"github.com/DotNetAge/gograph/internal/cypher/ast"
	"github.com/DotNetAge/gograph/internal/graph"
	"github.com/DotNetAge/gograph/internal/storage"
	"github.com/DotNetAge/gograph/internal/tx"
)

type Creator struct {
	Store *storage.DB
}

func NewCreator(store *storage.DB) *Creator {
	return &Creator{Store: store}
}

func (c *Creator) Execute(t *tx.Transaction, clause *ast.CreateClause) (affectedNodes, affectedRels int, err error) {
	for _, elem := range clause.Pattern.Elements {
		if elem.Node != nil {
			node := &graph.Node{
				ID:         graph.GenerateID("node"),
				Labels:     elem.Node.Labels,
				Properties: make(map[string]graph.PropertyValue),
			}

			for k, v := range elem.Node.Properties {
				node.Properties[k] = graph.ToPropertyValue(v)
			}

			data, err := storage.Marshal(node)
			if err != nil {
				return 0, 0, err
			}
			if err := t.Put(storage.NodeKey(node.ID), data); err != nil {
				return 0, 0, err
			}

			if err := c.indexWithTx(t, node); err != nil {
				return 0, 0, err
			}

			affectedNodes++

			if elem.Relation != nil && elem.Relation.EndNode != nil {
				endNode := &graph.Node{
					ID:         graph.GenerateID("node"),
					Labels:     elem.Relation.EndNode.Labels,
					Properties: make(map[string]graph.PropertyValue),
				}

				for k, v := range elem.Relation.EndNode.Properties {
					endNode.Properties[k] = graph.ToPropertyValue(v)
				}

				endData, err := storage.Marshal(endNode)
				if err != nil {
					return 0, 0, err
				}
				if err := t.Put(storage.NodeKey(endNode.ID), endData); err != nil {
					return 0, 0, err
				}

				if err := c.indexWithTx(t, endNode); err != nil {
					return 0, 0, err
				}

				rel := graph.NewRelationship(node.ID, endNode.ID, elem.Relation.RelType, elem.Relation.Properties)

				relData, err := storage.Marshal(rel)
				if err != nil {
					return 0, 0, err
				}
				if err := t.Put(storage.RelKey(rel.ID), relData); err != nil {
					return 0, 0, err
				}

				if err := c.adjWithTx(t, rel); err != nil {
					return 0, 0, err
				}

				affectedNodes++
				affectedRels++
			}
		}
	}

	return affectedNodes, affectedRels, nil
}

func (c *Creator) indexWithTx(t *tx.Transaction, node *graph.Node) error {
	for _, label := range node.Labels {
		key := storage.LabelKey(label, node.ID)
		if err := t.Put(key, []byte(node.ID)); err != nil {
			return err
		}
	}
	for _, label := range node.Labels {
		for propName, propValue := range node.Properties {
			encodedValue := graph.EncodePropertyValue(propValue)
			key := storage.PropertyKey(label, propName, encodedValue)
			if err := t.Put(key, []byte(node.ID)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Creator) adjWithTx(t *tx.Transaction, rel *graph.Relationship) error {
	outKey := storage.AdjKey(rel.StartNodeID, rel.Type, "out")
	if err := t.Put(outKey, []byte(rel.EndNodeID)); err != nil {
		return err
	}

	inKey := storage.AdjKey(rel.EndNodeID, rel.Type, "in")
	if err := t.Put(inKey, []byte(rel.StartNodeID)); err != nil {
		return err
	}

	return nil
}
