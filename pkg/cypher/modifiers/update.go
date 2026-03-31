// Package modifiers provides functionality for modifying nodes and relationships
// in the graph database. It implements the SET, DELETE, and REMOVE clause execution
// for Cypher queries.
//
// The modifier supports:
//   - Setting properties on nodes and relationships
//   - Adding labels to nodes
//   - Deleting nodes and relationships
//   - Removing properties and labels
//   - Detach delete (delete node with all relationships)
//
// Example:
//
//	modifier := modifiers.NewModifier(store)
//	nodes, err := modifier.ExecuteSet(tx, setStmt, path, params)
//	if err != nil {
//	    log.Fatal(err)
//	}
package modifiers

import (
	"github.com/DotNetAge/gograph/pkg/cypher/ast"
	"github.com/DotNetAge/gograph/pkg/cypher/utils"
	"github.com/DotNetAge/gograph/pkg/graph"
	"github.com/DotNetAge/gograph/pkg/storage"
	"github.com/DotNetAge/gograph/pkg/tx"
)

// Modifier modifies nodes and relationships in the graph database.
// It handles the execution of SET, DELETE, and REMOVE statements.
type Modifier struct {
	// Store is the underlying storage database.
	Store *storage.DB

	// index provides efficient lookups for nodes and relationships.
	index *graph.Index

	// adj manages adjacency lists for relationship traversal.
	adj *graph.AdjacencyList
}

// NewModifier creates a new Modifier instance.
//
// Parameters:
//   - store: The storage database
//
// Returns a new Modifier instance.
//
// Example:
//
//	modifier := modifiers.NewModifier(store)
func NewModifier(store *storage.DB) *Modifier {
	return &Modifier{
		Store: store,
		index: graph.NewIndex(store),
		adj:   graph.NewAdjacencyList(store),
	}
}

// ExecuteSet executes a SET statement and returns the number of affected nodes.
//
// Parameters:
//   - t: The transaction for atomic operations
//   - stmt: The SET statement AST node
//   - path: The matched path containing nodes and relationships
//   - params: Query parameters for parameterized queries
//
// Returns the number of affected nodes and any error encountered.
//
// Example:
//
//	affected, err := modifier.ExecuteSet(tx, setStmt, path, map[string]interface{}{
//	    "name": "Alice",
//	})
func (m *Modifier) ExecuteSet(t *tx.Transaction, stmt *ast.SetStmt, path map[string]interface{}, params map[string]interface{}) (affectedNodes int, err error) {
	for _, item := range stmt.Items {
		if item.Target == nil {
			continue
		}

		var node *graph.Node
		var rel *graph.Relationship

		if ident, ok := item.Target.(*ast.Ident); ok {
			if obj, exists := path[ident.Name]; exists {
				switch o := obj.(type) {
				case *graph.Node:
					node = o
				case *graph.Relationship:
					rel = o
				}
			}
		} else if pa, ok := item.Target.(*ast.PropertyAccessExpr); ok {
			if ident, ok := pa.Target.(*ast.Ident); ok {
				if obj, exists := path[ident.Name]; exists {
					switch o := obj.(type) {
					case *graph.Node:
						node = o
					case *graph.Relationship:
						rel = o
					}
				}
			}

			if item.IsLabel {
				if node != nil {
					label := m.exprToString(item.Value, params)
					if label != "" {
						node.Labels = append(node.Labels, label)
						if err := m.saveNode(t, node); err != nil {
							return 0, err
						}
						affectedNodes++
					}
				}
				continue
			}

			value := m.exprToValue(item.Value, params)
			if node != nil {
				node.Properties[pa.Property] = graph.ToPropertyValue(value)
				if err := m.saveNode(t, node); err != nil {
					return 0, err
				}
				affectedNodes++
			} else if rel != nil {
				rel.Properties[pa.Property] = graph.ToPropertyValue(value)
				if err := m.saveRel(t, rel); err != nil {
					return 0, err
				}
			}
		}
	}

	return affectedNodes, nil
}

// ExecuteDelete executes a DELETE statement and returns the number of affected elements.
//
// Parameters:
//   - t: The transaction for atomic operations
//   - stmt: The DELETE statement AST node
//   - path: The matched path containing nodes and relationships
//   - params: Query parameters for parameterized queries
//
// Returns the number of affected nodes, affected relationships, and any error encountered.
//
// Example:
//
//	nodes, rels, err := modifier.ExecuteDelete(tx, deleteStmt, path, params)
func (m *Modifier) ExecuteDelete(t *tx.Transaction, stmt *ast.DeleteStmt, path map[string]interface{}, params map[string]interface{}) (affectedNodes, affectedRels int, err error) {
	for _, expr := range stmt.Items {
		var node *graph.Node
		var rel *graph.Relationship

		switch e := expr.(type) {
		case *ast.Ident:
			if obj, exists := path[e.Name]; exists {
				switch o := obj.(type) {
				case *graph.Node:
					node = o
				case *graph.Relationship:
					rel = o
				}
			}
		case *ast.PropertyAccessExpr:
			if ident, ok := e.Target.(*ast.Ident); ok {
				if obj, exists := path[ident.Name]; exists {
					switch o := obj.(type) {
					case *graph.Node:
						node = o
					case *graph.Relationship:
						rel = o
					}
				}
			}
		}

		if rel != nil {
			if err := m.deleteRel(t, rel); err != nil {
				return 0, 0, err
			}
			affectedRels++
		} else if node != nil {
			if stmt.Detach {
				if err := m.detachDeleteNode(t, node); err != nil {
					return 0, 0, err
				}
			} else {
				if err := m.deleteNode(t, node); err != nil {
					return 0, 0, err
				}
			}
			affectedNodes++
		}
	}

	return affectedNodes, affectedRels, nil
}

// ExecuteRemove executes a REMOVE statement and returns the number of affected nodes.
//
// Parameters:
//   - t: The transaction for atomic operations
//   - stmt: The REMOVE statement AST node
//   - path: The matched path containing nodes and relationships
//   - params: Query parameters for parameterized queries
//
// Returns the number of affected nodes and any error encountered.
//
// Example:
//
//	affected, err := modifier.ExecuteRemove(tx, removeStmt, path, params)
func (m *Modifier) ExecuteRemove(t *tx.Transaction, stmt *ast.RemoveStmt, path map[string]interface{}, params map[string]interface{}) (affectedNodes int, err error) {
	for _, item := range stmt.Items {
		if item.Target == nil {
			continue
		}

		var node *graph.Node

		if item.IsLabel {
			if ident, ok := item.Target.(*ast.Ident); ok {
				if obj, exists := path[ident.Name]; exists {
					if n, ok := obj.(*graph.Node); ok {
						node = n
					}
				}
			}
			if node != nil {
				if err := m.index.RemoveLabelIndex(t, node); err != nil {
					return 0, err
				}
				if err := m.saveNode(t, node); err != nil {
					return 0, err
				}
				affectedNodes++
			}
		} else {
			if pa, ok := item.Target.(*ast.PropertyAccessExpr); ok {
				if ident, ok := pa.Target.(*ast.Ident); ok {
					if obj, exists := path[ident.Name]; exists {
						switch o := obj.(type) {
						case *graph.Node:
							node = o
						case *graph.Relationship:
							delete(o.Properties, pa.Property)
							if err := m.saveRel(t, o); err != nil {
								return 0, err
							}
						}
					}
				}
				if node != nil {
					delete(node.Properties, pa.Property)
					if err := m.saveNode(t, node); err != nil {
						return 0, err
					}
					affectedNodes++
				}
			}
		}
	}

	return affectedNodes, nil
}

// saveNode saves a node to the database and updates its indexes.
func (m *Modifier) saveNode(t *tx.Transaction, node *graph.Node) error {
	data, err := storage.Marshal(node)
	if err != nil {
		return err
	}
	if err := t.Put(storage.NodeKey(node.ID), data); err != nil {
		return err
	}
	return m.index.BuildPropertyIndex(t, node)
}

// saveRel saves a relationship to the database.
func (m *Modifier) saveRel(t *tx.Transaction, rel *graph.Relationship) error {
	data, err := storage.Marshal(rel)
	if err != nil {
		return err
	}
	return t.Put(storage.RelKey(rel.ID), data)
}

// deleteNode deletes a node from the database and removes its indexes.
func (m *Modifier) deleteNode(t *tx.Transaction, node *graph.Node) error {
	if err := t.Delete(storage.NodeKey(node.ID)); err != nil {
		return err
	}
	if err := m.index.RemoveLabelIndex(t, node); err != nil {
		return err
	}
	return m.index.RemovePropertyIndex(t, node)
}

// deleteRel deletes a relationship from the database and removes it from adjacency lists.
func (m *Modifier) deleteRel(t *tx.Transaction, rel *graph.Relationship) error {
	if err := t.Delete(storage.RelKey(rel.ID)); err != nil {
		return err
	}
	return m.adj.RemoveRelationship(t, rel)
}

// detachDeleteNode deletes a node and all its relationships.
func (m *Modifier) detachDeleteNode(t *tx.Transaction, node *graph.Node) error {
	relIDs, _ := m.adj.GetAllRelated(node.ID)
	for _, relID := range relIDs {
		relData, err := m.Store.Get(storage.RelKey(relID))
		if err != nil {
			continue
		}
		var rel graph.Relationship
		if err := storage.Unmarshal(relData, &rel); err != nil {
			continue
		}
		if err := m.deleteRel(t, &rel); err != nil {
			return err
		}
	}
	return m.deleteNode(t, node)
}

// exprToValue converts an expression to its value.
func (m *Modifier) exprToValue(expr ast.Expr, params map[string]interface{}) interface{} {
	return utils.ExprToValue(expr, params)
}

// exprToString converts an expression to a string value.
func (m *Modifier) exprToString(expr ast.Expr, params map[string]interface{}) string {
	return utils.ExprToString(expr, params)
}
