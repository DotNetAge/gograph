// Package matchers provides executors for Cypher MATCH clauses and pattern matching.
package matchers

import (
	"fmt"
	"strings"

	"github.com/DotNetAge/gograph/pkg/cypher/ast"
	"github.com/DotNetAge/gograph/pkg/graph"
	"github.com/DotNetAge/gograph/pkg/storage"
)

// Matcher executes MATCH clauses to find existing graph patterns.
// It supports pattern matching on nodes and relationships with optional
// WHERE filters and RETURN projections.
type Matcher struct {
	Store *storage.DB
	Index *graph.Index
}

// NewMatcher creates a new Matcher instance with the given storage and index.
func NewMatcher(store *storage.DB, index *graph.Index) *Matcher {
	return &Matcher{Store: store, Index: index}
}

// Execute finds graph elements matching the MATCH clause pattern.
// It returns rows of matched data, column names, and any error encountered.
// The varVars map is used to store matched variables for use by subsequent clauses.
func (m *Matcher) Execute(clause *ast.MatchClause, varVars map[string]interface{}, params map[string]interface{}) (rows []map[string]interface{}, columns []string, err error) {
	var matchedPaths []map[string]interface{}

	if len(clause.Pattern.Elements) > 0 {
		elem := clause.Pattern.Elements[0]
		if elem.Node != nil && elem.Relation != nil {
			// Find start nodes using index if possible
			var startNodes []*graph.Node
			if len(elem.Node.Labels) > 0 {
				ids, _ := m.Index.LookupByLabel(elem.Node.Labels[0])
				for _, id := range ids {
					data, err := m.Store.Get(storage.NodeKey(id))
					if err != nil {
						continue
					}
					var node graph.Node
					if err := storage.Unmarshal(data, &node); err == nil {
						if m.NodeMatchesProperties(&node, elem.Node.Labels, elem.Node.Properties) {
							startNodes = append(startNodes, &node)
						}
					}
				}
			} else {
				// Fallback to full node scan
				iter, _ := m.Store.NewIter(nil)
				for iter.SeekGE([]byte(storage.KeyPrefixNode)); iter.Valid(); iter.Next() {
					key := iter.Key()
					if len(key) < 5 || string(key)[:5] != storage.KeyPrefixNode {
						break
					}
					var node graph.Node
					if err := storage.Unmarshal(iter.Value(), &node); err == nil {
						if m.NodeMatchesProperties(&node, nil, elem.Node.Properties) {
							startNodes = append(startNodes, &node)
						}
					}
				}
				iter.Close()
			}

			// For each start node, follow adjacency list
			adj := graph.NewAdjacencyList(m.Store)
			for _, startNode := range startNodes {
				relIDs, _ := adj.GetAllRelated(startNode.ID)
				for _, relID := range relIDs {
					relData, err := m.Store.Get(storage.RelKey(relID))
					if err != nil {
						continue
					}
					var rel graph.Relationship
					if err := storage.Unmarshal(relData, &rel); err != nil {
						continue
					}

					// Check rel type and direction
					if elem.Relation.RelType != "" && rel.Type != elem.Relation.RelType {
						continue
					}
					// Only follow outgoing relationships in current simple implementation
					if rel.StartNodeID != startNode.ID {
						continue
					}

					// Load end node
					endData, err := m.Store.Get(storage.NodeKey(rel.EndNodeID))
					if err != nil {
						continue
					}
					var endNode graph.Node
					if err := storage.Unmarshal(endData, &endNode); err != nil {
						continue
					}

					if elem.Relation.EndNode != nil {
						if !m.NodeMatchesProperties(&endNode, elem.Relation.EndNode.Labels, elem.Relation.EndNode.Properties) {
							continue
						}
					}

					path := make(map[string]interface{})
					if elem.Node.Variable != "" {
						path[elem.Node.Variable] = startNode
					}
					if elem.Relation.Variable != "" {
						path[elem.Relation.Variable] = &rel
					}
					if elem.Relation.EndNode != nil && elem.Relation.EndNode.Variable != "" {
						path[elem.Relation.EndNode.Variable] = &endNode
					}
					matchedPaths = append(matchedPaths, path)
				}
			}
		} else if elem.Node != nil {
			if len(elem.Node.Labels) > 0 {
				// Use index
				ids, _ := m.Index.LookupByLabel(elem.Node.Labels[0])
				for _, id := range ids {
					data, err := m.Store.Get(storage.NodeKey(id))
					if err != nil {
						continue
					}
					var node graph.Node
					if err := storage.Unmarshal(data, &node); err == nil {
						if m.NodeMatchesProperties(&node, elem.Node.Labels, elem.Node.Properties) {
							path := make(map[string]interface{})
							if elem.Node.Variable != "" {
								path[elem.Node.Variable] = &node
							} else {
								path["n"] = &node
							}
							matchedPaths = append(matchedPaths, path)
						}
					}
				}
			} else {
				// Fallback to full node scan
				nodeIter, err := m.Store.NewIter(nil)
				if err != nil {
					return nil, nil, err
				}
				defer nodeIter.Close()

				for nodeIter.SeekGE([]byte(storage.KeyPrefixNode)); nodeIter.Valid(); nodeIter.Next() {
					key := nodeIter.Key()
					if len(key) < 5 || string(key)[:5] != storage.KeyPrefixNode {
						break
					}
					var node graph.Node
					if err := storage.Unmarshal(nodeIter.Value(), &node); err == nil {
						if m.NodeMatchesProperties(&node, nil, elem.Node.Properties) {
							path := make(map[string]interface{})
							if elem.Node.Variable != "" {
								path[elem.Node.Variable] = &node
							} else {
								path["n"] = &node
							}
							matchedPaths = append(matchedPaths, path)
						}
					}
				}
			}
		}
	}

	if clause.Where != nil {
		var filteredPaths []map[string]interface{}
		for _, path := range matchedPaths {
			if m.EvaluateExpression(path, clause.Where.Expression, params) {
				filteredPaths = append(filteredPaths, path)
			}
		}
		matchedPaths = filteredPaths
	}

	for _, path := range matchedPaths {
		row := make(map[string]interface{})
		if clause.Return != nil {
			for _, item := range clause.Return.Items {
				m.fillRow(row, item, path)
			}
		} else {
			for k, v := range path {
				row[k] = v
			}
		}
		rows = append(rows, row)

		// This updates varVars, but it only keeps the last row's state for subsequent modifiers
		// unless we fix the executor to loop.
		for k, v := range path {
			varVars[k] = v
		}
	}

	if clause.Return != nil {
		for _, item := range clause.Return.Items {
			if lookup, ok := item.Expression.(*ast.PropertyLookup); ok {
				if lookup.Property == "" {
					columns = append(columns, lookup.Node)
				} else {
					columns = append(columns, lookup.Node+"."+lookup.Property)
				}
			} else if ident, ok := item.Expression.(*ast.Identifier); ok {
				columns = append(columns, ident.Name)
			}
		}
	} else if len(rows) > 0 {
		for k := range rows[0] {
			columns = append(columns, k)
		}
	}

	return rows, columns, nil
}

// EvaluateExpression evaluates a Cypher expression against a matched path.
func (m *Matcher) EvaluateExpression(path map[string]interface{}, expr ast.Expression, params map[string]interface{}) bool {
	comp, ok := expr.(*ast.ComparisonOp)
	if !ok {
		return true
	}

	lookup, ok := comp.Left.(*ast.PropertyLookup)
	if !ok {
		return true
	}

	obj := path[lookup.Node]
	if obj == nil {
		return false
	}

	var prop graph.PropertyValue
	var exists bool
	if n, ok := obj.(*graph.Node); ok {
		prop, exists = n.Properties[lookup.Property]
	} else if r, ok := obj.(*graph.Relationship); ok {
		prop, exists = r.Properties[lookup.Property]
	}

	if !exists {
		return false
	}

	var rightVal interface{}
	if lit, ok := comp.Right.(*ast.Literal); ok {
		rightVal = lit.Value
	} else if ident, ok := comp.Right.(*ast.Identifier); ok {
		paramName := strings.TrimPrefix(ident.Name, "$")
		rightVal = params[paramName]
	}

	switch prop.Type() {
	case graph.PropertyTypeInt:
		leftVal := prop.IntValue()
		switch comp.Operator {
		case ">":
			if v, ok := ToInt64(rightVal); ok {
				return leftVal > v
			}
			if v, ok := ToFloat64(rightVal); ok {
				return float64(leftVal) > v
			}
		case ">=":
			if v, ok := ToInt64(rightVal); ok {
				return leftVal >= v
			}
			if v, ok := ToFloat64(rightVal); ok {
				return float64(leftVal) >= v
			}
		case "<":
			if v, ok := ToInt64(rightVal); ok {
				return leftVal < v
			}
			if v, ok := ToFloat64(rightVal); ok {
				return float64(leftVal) < v
			}
		case "<=":
			if v, ok := ToInt64(rightVal); ok {
				return leftVal <= v
			}
			if v, ok := ToFloat64(rightVal); ok {
				return float64(leftVal) <= v
			}
		case "=":
			if v, ok := ToInt64(rightVal); ok {
				return leftVal == v
			}
			if v, ok := ToFloat64(rightVal); ok {
				return float64(leftVal) == v
			}
		case "!=":
			if v, ok := ToInt64(rightVal); ok {
				return leftVal != v
			}
			if v, ok := ToFloat64(rightVal); ok {
				return float64(leftVal) != v
			}
		}
	case graph.PropertyTypeFloat:
		leftVal := prop.FloatValue()
		rightFloat, ok := ToFloat64(rightVal)
		if !ok {
			return false
		}
		switch comp.Operator {
		case ">":
			return leftVal > rightFloat
		case ">=":
			return leftVal >= rightFloat
		case "<":
			return leftVal < rightFloat
		case "<=":
			return leftVal <= rightFloat
		case "=":
			return leftVal == rightFloat
		case "!=":
			return leftVal != rightFloat
		}
	case graph.PropertyTypeString:
		leftVal := prop.StringValue()
		rightStr := fmt.Sprintf("%v", rightVal)
		switch comp.Operator {
		case "=":
			return leftVal == rightStr
		case "!=":
			return leftVal != rightStr
		}
	}
	return false
}

func (m *Matcher) fillRow(row map[string]interface{}, item ast.ReturnItem, path map[string]interface{}) {
	if lookup, ok := item.Expression.(*ast.PropertyLookup); ok {
		obj := path[lookup.Node]
		if obj == nil {
			return
		}

		var prop graph.PropertyValue
		var exists bool
		if n, ok := obj.(*graph.Node); ok {
			if lookup.Property == "" {
				row[lookup.Node] = n
				return
			}
			prop, exists = n.Properties[lookup.Property]
		} else if r, ok := obj.(*graph.Relationship); ok {
			if lookup.Property == "" {
				row[lookup.Node] = r
				return
			}
			prop, exists = r.Properties[lookup.Property]
		}

		if exists {
			row[lookup.Node+"."+lookup.Property] = m.PropertyToInterface(prop)
		}
	} else if ident, ok := item.Expression.(*ast.Identifier); ok {
		if obj, ok := path[ident.Name]; ok {
			row[ident.Name] = obj
		}
	}
}

// NodeMatchesProperties checks if a node matches the specified labels and properties.
func (m *Matcher) NodeMatchesProperties(node *graph.Node, labels []string, props map[string]interface{}) bool {
	if len(labels) > 0 {
		for _, label := range labels {
			matchedLabel := false
			for _, nodeLabel := range node.Labels {
				if nodeLabel == label {
					matchedLabel = true
					break
				}
			}
			if !matchedLabel {
				return false
			}
		}
	}

	for k, v := range props {
		prop, exists := node.Properties[k]
		if !exists {
			return false
		}
		switch prop.Type() {
		case graph.PropertyTypeString:
			if prop.StringValue() != fmt.Sprintf("%v", v) {
				return false
			}
		case graph.PropertyTypeInt:
			nodeVal := prop.IntValue()
			expectedVal, ok := ToInt64(v)
			if !ok || nodeVal != expectedVal {
				return false
			}
		case graph.PropertyTypeFloat:
			nodeVal := prop.FloatValue()
			expectedVal, ok := ToFloat64(v)
			if !ok || nodeVal != expectedVal {
				return false
			}
		case graph.PropertyTypeBool:
			if b, ok := v.(bool); ok {
				if prop.BoolValue() != b {
					return false
				}
			} else {
				return false
			}
		}
	}
	return true
}

// PropertyToInterface converts a PropertyValue to a Go interface{}.
func (m *Matcher) PropertyToInterface(prop graph.PropertyValue) interface{} {
	switch prop.Type() {
	case graph.PropertyTypeString:
		return prop.StringValue()
	case graph.PropertyTypeInt:
		return prop.IntValue()
	case graph.PropertyTypeFloat:
		return prop.FloatValue()
	case graph.PropertyTypeBool:
		return prop.BoolValue()
	default:
		return nil
	}
}

// ToInt64 converts various numeric types to int64.
func ToInt64(v interface{}) (int64, bool) {
	switch val := v.(type) {
	case int64:
		return val, true
	case int:
		return int64(val), true
	case int32:
		return int64(val), true
	case float64:
		if float64(int64(val)) == val {
			return int64(val), true
		}
	}
	return 0, false
}

// ToFloat64 converts various numeric types to float64.
func ToFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int64:
		return float64(val), true
	case int:
		return float64(val), true
	}
	return 0.0, false
}

// FindNodesByVariableAndLabel is a helper method to find nodes by variable name or label.
// (Keeping it for compatibility if needed, but it should be optimized or removed)
func (m *Matcher) FindNodesByVariableAndLabel(varName string, where *ast.WhereClause) ([]*graph.Node, error) {
	var nodes []*graph.Node

	iter, err := m.Store.NewIter(nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	for iter.SeekGE([]byte(storage.KeyPrefixNode)); iter.Valid(); iter.Next() {
		key := iter.Key()
		if len(key) < 5 || string(key)[:5] != storage.KeyPrefixNode {
			break
		}
		var node graph.Node
		if err := storage.Unmarshal(iter.Value(), &node); err != nil {
			continue
		}
		nodes = append(nodes, &node)
	}

	return nodes, nil
}
