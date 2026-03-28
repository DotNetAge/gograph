// Package creators provides executors for Cypher data creation clauses.
package creators

import (
	"github.com/DotNetAge/gograph/pkg/cypher/ast"
	"github.com/DotNetAge/gograph/pkg/graph"
	"github.com/DotNetAge/gograph/pkg/storage"
	"github.com/DotNetAge/gograph/pkg/tx"
)

// Creator executes CREATE clauses to create new nodes and relationships in the graph.
type Creator struct {
	Store *storage.DB
	index *graph.Index
	adj   *graph.AdjacencyList
}

// NewCreator creates a new Creator instance that operates on the given storage.
func NewCreator(store *storage.DB) *Creator {
	return &Creator{
		Store: store,
		index: graph.NewIndex(store),
		adj:   graph.NewAdjacencyList(store),
	}
}

// Execute creates nodes and relationships as specified by the CREATE clause.
// It returns the count of affected nodes and relationships.
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

			if err := c.index.BuildLabelIndex(t, node); err != nil {
				return 0, 0, err
			}
			if err := c.index.BuildPropertyIndex(t, node); err != nil {
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

				if err := c.index.BuildLabelIndex(t, endNode); err != nil {
					return 0, 0, err
				}
				if err := c.index.BuildPropertyIndex(t, endNode); err != nil {
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

				if err := c.adj.AddRelationship(t, rel); err != nil {
				return 0, 0, err
			}

			affectedNodes++
			affectedRels++
			}
		}
	}

	return affectedNodes, affectedRels, nil
}
