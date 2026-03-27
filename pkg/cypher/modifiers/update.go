// Package modifiers provides executors for Cypher data modification clauses
// including SET, DELETE, and REMOVE operations.
package modifiers

import (
	"strings"

	"github.com/DotNetAge/gograph/pkg/cypher/ast"
	"github.com/DotNetAge/gograph/pkg/cypher/matchers"
	"github.com/DotNetAge/gograph/pkg/graph"
	"github.com/DotNetAge/gograph/pkg/storage"
	"github.com/DotNetAge/gograph/pkg/tx"
)

// Modifier executes data modification clauses (SET, DELETE, REMOVE).
// It coordinates with the Matcher to find existing graph elements to modify.
type Modifier struct {
	Store   *storage.DB
	Matcher *matchers.Matcher
	index   *graph.Index
	adj     *graph.AdjacencyList
}

// NewModifier creates a new Modifier instance with the given storage and matcher.
func NewModifier(store *storage.DB, matcher *matchers.Matcher) *Modifier {
	return &Modifier{
		Store:   store,
		Matcher: matcher,
		index:   graph.NewIndex(store),
		adj:     graph.NewAdjacencyList(store),
	}
}

// ExecuteSet executes a SET clause to update node properties.
// It returns the count of affected nodes.
func (m *Modifier) ExecuteSet(t *tx.Transaction, clause *ast.SetClause, varVars map[string]interface{}, params map[string]interface{}) (affectedNodes int, err error) {
	for _, assignment := range clause.Assignments {
		nodeID := assignment.Property.Node
		if node, ok := varVars[nodeID].(*graph.Node); ok {
			// Remove old index before modifying properties
			m.index.RemovePropertyIndex(t, node)

			val := m.resolveValue(assignment.Value, params)
			node.Properties[assignment.Property.Property] = graph.ToPropertyValue(val)

			data, err := storage.Marshal(node)
			if err != nil {
				return 0, err
			}
			if err := t.Put(storage.NodeKey(node.ID), data); err != nil {
				return 0, err
			}

			// Rebuild index with new properties
			if err := m.index.BuildPropertyIndex(t, node); err != nil {
				return 0, err
			}

			affectedNodes++
		}
	}
	return affectedNodes, nil
}

// ExecuteDelete executes a DELETE clause to remove nodes and relationships.
// If Detach is true, it also removes all relationships connected to deleted nodes.
// It returns the count of affected nodes and relationships.
func (m *Modifier) ExecuteDelete(t *tx.Transaction, clause *ast.DeleteClause, varVars map[string]interface{}, params map[string]interface{}) (affectedNodes, affectedRels int, err error) {
	for _, expr := range clause.Expressions {
		switch v := expr.(type) {
		case *ast.PropertyLookup:
			if v.Property == "" {
				if node, ok := varVars[v.Node].(*graph.Node); ok {
					if clause.Detach {
						if err := m.deleteRelationshipsOfNode(t, node.ID); err != nil {
							continue
						}
					}
					if err := m.index.RemoveLabelIndex(t, node); err != nil {
						continue
					}
					if err := m.index.RemovePropertyIndex(t, node); err != nil {
						continue
					}
					if err := t.Delete(storage.NodeKey(node.ID)); err != nil {
						continue
					}
					affectedNodes++
				}
			}
		case *ast.RelationVariable:
			if rel, ok := varVars[v.Name].(*graph.Relationship); ok {
				if err := m.adj.RemoveRelationship(t, rel); err != nil {
					continue
				}
				if err := t.Delete(storage.RelKey(rel.ID)); err != nil {
					continue
				}
				affectedRels++
			}
		}
	}
	return affectedNodes, affectedRels, nil
}

// ExecuteRemove executes a REMOVE clause to remove labels or properties from nodes.
// It returns the count of affected nodes.
func (m *Modifier) ExecuteRemove(t *tx.Transaction, clause *ast.RemoveClause, varVars map[string]interface{}, params map[string]interface{}) (affectedNodes int, err error) {
	for _, remove := range clause.Removals {
		var targetNode *graph.Node

		if remove.Type == ast.RemoveItemTypeLabel {
			if node, ok := varVars[remove.Property.Node].(*graph.Node); ok {
				targetNode = node
			}
			if targetNode != nil {
				// Remove old indexes
				m.index.RemoveLabelIndex(t, targetNode)
				m.index.RemovePropertyIndex(t, targetNode)

				targetNode.RemoveLabel(remove.Label)
				data, err := storage.Marshal(targetNode)
				if err != nil {
					return 0, err
				}
				if err := t.Put(storage.NodeKey(targetNode.ID), data); err != nil {
					return 0, err
				}

				// Rebuild indexes
				if err := m.index.BuildLabelIndex(t, targetNode); err != nil {
					return 0, err
				}
				if err := m.index.BuildPropertyIndex(t, targetNode); err != nil {
					return 0, err
				}

				affectedNodes++
			}
		} else if remove.Type == ast.RemoveItemTypeProperty {
			if node, ok := varVars[remove.Property.Node].(*graph.Node); ok {
				targetNode = node
			}
			if targetNode != nil {
				m.index.RemovePropertyIndex(t, targetNode)

				delete(targetNode.Properties, remove.Property.Property)

				data, err := storage.Marshal(targetNode)
				if err != nil {
					return 0, err
				}
				if err := t.Put(storage.NodeKey(targetNode.ID), data); err != nil {
					return 0, err
				}

				if err := m.index.BuildPropertyIndex(t, targetNode); err != nil {
					return 0, err
				}

				affectedNodes++
			}
		}
	}
	return affectedNodes, nil
}
// resolveValue resolves an expression to its value, considering parameters.
func (m *Modifier) resolveValue(expr ast.Expression, params map[string]interface{}) interface{} {
	switch v := expr.(type) {
	case *ast.Literal:
		return v.Value
	case *ast.Identifier:
		paramName := strings.TrimPrefix(v.Name, "$")
		if val, ok := params[paramName]; ok {
			return val
		}
		return nil
	default:
		return nil
	}
}

// deleteRelationshipsOfNode removes all relationships connected to a node.
func (m *Modifier) deleteRelationshipsOfNode(t *tx.Transaction, nodeID string) error {
	relsToDelete, err := m.adj.GetAllRelated(nodeID)
	if err != nil {
		return err
	}

	for _, relID := range relsToDelete {
		data, err := t.Get(storage.RelKey(relID))
		if err != nil {
			continue
		}
		var rel graph.Relationship
		if err := storage.Unmarshal(data, &rel); err != nil {
			continue
		}
		if err := m.adj.RemoveRelationship(t, &rel); err != nil {
			continue
		}
		if err := t.Delete(storage.RelKey(rel.ID)); err != nil {
			continue
		}
	}

	return nil
}
