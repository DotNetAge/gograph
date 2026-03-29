package modifiers

import (
	"github.com/DotNetAge/gograph/pkg/cypher/ast"
	"github.com/DotNetAge/gograph/pkg/cypher/utils"
	"github.com/DotNetAge/gograph/pkg/graph"
	"github.com/DotNetAge/gograph/pkg/storage"
	"github.com/DotNetAge/gograph/pkg/tx"
)

type Modifier struct {
	Store *storage.DB
	index *graph.Index
	adj   *graph.AdjacencyList
}

func NewModifier(store *storage.DB) *Modifier {
	return &Modifier{
		Store: store,
		index: graph.NewIndex(store),
		adj:   graph.NewAdjacencyList(store),
	}
}

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

func (m *Modifier) saveRel(t *tx.Transaction, rel *graph.Relationship) error {
	data, err := storage.Marshal(rel)
	if err != nil {
		return err
	}
	return t.Put(storage.RelKey(rel.ID), data)
}

func (m *Modifier) deleteNode(t *tx.Transaction, node *graph.Node) error {
	if err := t.Delete(storage.NodeKey(node.ID)); err != nil {
		return err
	}
	if err := m.index.RemoveLabelIndex(t, node); err != nil {
		return err
	}
	return m.index.RemovePropertyIndex(t, node)
}

func (m *Modifier) deleteRel(t *tx.Transaction, rel *graph.Relationship) error {
	if err := t.Delete(storage.RelKey(rel.ID)); err != nil {
		return err
	}
	return m.adj.RemoveRelationship(t, rel)
}

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

func (m *Modifier) exprToValue(expr ast.Expr, params map[string]interface{}) interface{} {
	return utils.ExprToValue(expr, params)
}

func (m *Modifier) exprToString(expr ast.Expr, params map[string]interface{}) string {
	return utils.ExprToString(expr, params)
}
