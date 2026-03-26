package modifiers

import (
	"github.com/DotNetAge/gograph/internal/cypher/ast"
	"github.com/DotNetAge/gograph/internal/cypher/matchers"
	"github.com/DotNetAge/gograph/internal/graph"
	"github.com/DotNetAge/gograph/internal/storage"
	"github.com/DotNetAge/gograph/internal/tx"
)

type Modifier struct {
	Store   *storage.DB
	Matcher *matchers.Matcher
}

func NewModifier(store *storage.DB, matcher *matchers.Matcher) *Modifier {
	return &Modifier{Store: store, Matcher: matcher}
}

func (m *Modifier) ExecuteSet(t *tx.Transaction, clause *ast.SetClause, varVars map[string]interface{}, params map[string]interface{}) (affectedNodes int, err error) {
	for _, assignment := range clause.Assignments {
		nodeID := assignment.Property.Node
		if node, ok := varVars[nodeID].(*graph.Node); ok {
			val := m.resolveValue(assignment.Value, params)
			node.Properties[assignment.Property.Property] = graph.ToPropertyValue(val)

			data, err := storage.Marshal(node)
			if err != nil {
				return 0, err
			}
			if err := t.Put(storage.NodeKey(node.ID), data); err != nil {
				return 0, err
			}

			affectedNodes++
			continue
		}
		nodesFromMatch, err := m.Matcher.FindNodesByVariableAndLabel(nodeID, nil)
		if err != nil || len(nodesFromMatch) == 0 {
			continue
		}
		for _, node := range nodesFromMatch {
			val := m.resolveValue(assignment.Value, params)
			node.Properties[assignment.Property.Property] = graph.ToPropertyValue(val)

			data, err := storage.Marshal(node)
			if err != nil {
				return 0, err
			}
			if err := t.Put(storage.NodeKey(node.ID), data); err != nil {
				return 0, err
			}

			affectedNodes++
		}
	}
	return affectedNodes, nil
}

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
					if err := m.removeIndexWithTx(t, node); err != nil {
						continue
					}
					if err := t.Delete(storage.NodeKey(node.ID)); err != nil {
						continue
					}
					affectedNodes++
				} else {
					nodes, err := m.Matcher.FindNodesByVariableAndLabel(v.Node, nil)
					if err != nil {
						continue
					}
					for _, node := range nodes {
						if clause.Detach {
							if err := m.deleteRelationshipsOfNode(t, node.ID); err != nil {
								continue
							}
						}
						if err := m.removeIndexWithTx(t, node); err != nil {
							continue
						}
						if err := t.Delete(storage.NodeKey(node.ID)); err != nil {
							continue
						}
						affectedNodes++
					}
				}
			} else {
				// Handle other property lookups?
			}
		case *ast.RelationVariable:
			if rel, ok := varVars[v.Name].(*graph.Relationship); ok {
				if err := m.adjRemoveWithTx(t, rel); err != nil {
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

func (m *Modifier) ExecuteRemove(t *tx.Transaction, clause *ast.RemoveClause, varVars map[string]interface{}, params map[string]interface{}) (affectedNodes int, err error) {
	for _, remove := range clause.Removals {
		if remove.Type == ast.RemoveItemTypeLabel {
			nodes, err := m.Matcher.FindNodesByVariableAndLabel(remove.Label, nil)
			if err != nil {
				continue
			}
			for _, node := range nodes {
				node.RemoveLabel(remove.Label)

				data, err := storage.Marshal(node)
				if err != nil {
					return 0, err
				}
				if err := t.Put(storage.NodeKey(node.ID), data); err != nil {
					return 0, err
				}

				affectedNodes++
			}
		} else if remove.Type == ast.RemoveItemTypeProperty {
			nodes, err := m.Matcher.FindNodesByVariableAndLabel(remove.Property.Node, nil)
			if err != nil {
				continue
			}
			for _, node := range nodes {
				delete(node.Properties, remove.Property.Property)

				data, err := storage.Marshal(node)
				if err != nil {
					return 0, err
				}
				if err := t.Put(storage.NodeKey(node.ID), data); err != nil {
					return 0, err
				}

				affectedNodes++
			}
		}
	}
	return affectedNodes, nil
}

func (m *Modifier) resolveValue(expr ast.Expression, params map[string]interface{}) interface{} {
	switch v := expr.(type) {
	case *ast.Literal:
		return v.Value
	case *ast.Identifier:
		if val, ok := params[v.Name]; ok {
			return val
		}
		return nil
	default:
		return nil
	}
}

func (m *Modifier) deleteRelationshipsOfNode(t *tx.Transaction, nodeID string) error {
	iter, err := m.Store.NewIter(nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	var relsToDelete []string

	for iter.SeekGE([]byte(storage.KeyPrefixRel)); iter.Valid(); iter.Next() {
		key := iter.Key()
		if len(key) < 4 || string(key)[:4] != storage.KeyPrefixRel {
			break
		}
		var rel graph.Relationship
		if err := storage.Unmarshal(iter.Value(), &rel); err != nil {
			continue
		}
		if rel.StartNodeID == nodeID || rel.EndNodeID == nodeID {
			relsToDelete = append(relsToDelete, rel.ID)
		}
	}

	for _, relID := range relsToDelete {
		data, err := m.Store.Get(storage.RelKey(relID))
		if err != nil {
			continue
		}
		var rel graph.Relationship
		if err := storage.Unmarshal(data, &rel); err != nil {
			continue
		}
		if err := m.adjRemoveWithTx(t, &rel); err != nil {
			continue
		}
		if err := t.Delete(storage.RelKey(rel.ID)); err != nil {
			continue
		}
	}

	return nil
}

func (m *Modifier) removeIndexWithTx(t *tx.Transaction, node *graph.Node) error {
	for _, label := range node.Labels {
		key := storage.LabelKey(label, node.ID)
		if err := t.Delete(key); err != nil {
			return err
		}
	}
	for _, label := range node.Labels {
		for propName, propValue := range node.Properties {
			encodedValue := graph.EncodePropertyValue(propValue)
			key := storage.PropertyKey(label, propName, encodedValue)
			if err := t.Delete(key); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Modifier) adjRemoveWithTx(t *tx.Transaction, rel *graph.Relationship) error {
	outKey := storage.AdjKey(rel.StartNodeID, rel.Type, "out")
	if err := t.Delete(outKey); err != nil {
		return err
	}

	inKey := storage.AdjKey(rel.EndNodeID, rel.Type, "in")
	if err := t.Delete(inKey); err != nil {
		return err
	}

	return nil
}
