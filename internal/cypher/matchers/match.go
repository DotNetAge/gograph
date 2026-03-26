package matchers

import (
	"fmt"
	"strings"

	"github.com/DotNetAge/gograph/internal/cypher/ast"
	"github.com/DotNetAge/gograph/internal/graph"
	"github.com/DotNetAge/gograph/internal/storage"
)

type Matcher struct {
	Store *storage.DB
	Index *graph.Index
}

func NewMatcher(store *storage.DB, index *graph.Index) *Matcher {
	return &Matcher{Store: store, Index: index}
}

func (m *Matcher) Execute(clause *ast.MatchClause, varVars map[string]interface{}, params map[string]interface{}) (rows []map[string]interface{}, columns []string, err error) {
	var matchedPaths []map[string]interface{}

	if len(clause.Pattern.Elements) > 0 {
		elem := clause.Pattern.Elements[0]
		if elem.Node != nil && elem.Relation != nil {
			// (n)-[r]->(m)
			relType := elem.Relation.RelType
			iter, err := m.Store.NewIter(nil)
			if err != nil {
				return nil, nil, err
			}
			defer iter.Close()

			for iter.SeekGE([]byte(storage.KeyPrefixRel)); iter.Valid(); iter.Next() {
				key := iter.Key()
				if len(key) < 4 || string(key)[:4] != storage.KeyPrefixRel {
					break
				}
				var rel graph.Relationship
				if err := storage.Unmarshal(iter.Value(), &rel); err != nil {
					continue
				}
				if relType != "" && rel.Type != relType {
					continue
				}

				// Fetch start and end nodes
				startData, err := m.Store.Get(storage.NodeKey(rel.StartNodeID))
				if err != nil {
					continue
				}
				var startNode graph.Node
				storage.Unmarshal(startData, &startNode)

				endData, err := m.Store.Get(storage.NodeKey(rel.EndNodeID))
				if err != nil {
					continue
				}
				var endNode graph.Node
				storage.Unmarshal(endData, &endNode)

				// Check if nodes match patterns
				if !m.NodeMatchesProperties(&startNode, elem.Node.Labels, elem.Node.Properties) {
					continue
				}
				if elem.Relation.EndNode != nil {
					if !m.NodeMatchesProperties(&endNode, elem.Relation.EndNode.Labels, elem.Relation.EndNode.Properties) {
						continue
					}
				}

				path := make(map[string]interface{})
				if elem.Node.Variable != "" {
					path[elem.Node.Variable] = &startNode
				}
				if elem.Relation.Variable != "" {
					path[elem.Relation.Variable] = &rel
				}
				if elem.Relation.EndNode != nil && elem.Relation.EndNode.Variable != "" {
					path[elem.Relation.EndNode.Variable] = &endNode
				}
				matchedPaths = append(matchedPaths, path)
			}
		} else if elem.Node != nil {
			// (n)
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
				if err := storage.Unmarshal(nodeIter.Value(), &node); err != nil {
					continue
				}
				if !m.NodeMatchesProperties(&node, elem.Node.Labels, elem.Node.Properties) {
					continue
				}
				path := make(map[string]interface{})
				if elem.Node.Variable != "" {
					path[elem.Node.Variable] = &node
				} else {
					path["n"] = &node // default if no variable
				}
				matchedPaths = append(matchedPaths, path)
			}
		}
	}

	// Apply WHERE filter
	if clause.Where != nil {
		var filteredPaths []map[string]interface{}
		for _, path := range matchedPaths {
			if m.EvaluateExpression(path, clause.Where.Expression, params) {
				filteredPaths = append(filteredPaths, path)
			}
		}
		matchedPaths = filteredPaths
	}

	// Build rows and columns
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
		
		// Also update varVars for the LAST match (needed for DELETE r etc)
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
					columns = append(columns, lookup.Property)
				}
			}
		}
	} else if len(rows) > 0 {
		// Just use keys of the first row
		for k := range rows[0] {
			columns = append(columns, k)
		}
	}

	return rows, columns, nil
}

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
		case ">=":
			if v, ok := ToInt64(rightVal); ok {
				return leftVal >= v
			}
		case "<":
			if v, ok := ToInt64(rightVal); ok {
				return leftVal < v
			}
		case "<=":
			if v, ok := ToInt64(rightVal); ok {
				return leftVal <= v
			}
		case "=":
			if v, ok := ToInt64(rightVal); ok {
				return leftVal == v
			}
		case "!=":
			if v, ok := ToInt64(rightVal); ok {
				return leftVal != v
			}
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

		if lookup.Property == "" {
			row[lookup.Node] = obj
		} else {
			if n, ok := obj.(*graph.Node); ok {
				if prop, ok := n.Properties[lookup.Property]; ok {
					row[lookup.Property] = m.PropertyToInterface(prop)
				}
			} else if r, ok := obj.(*graph.Relationship); ok {
				if prop, ok := r.Properties[lookup.Property]; ok {
					row[lookup.Property] = m.PropertyToInterface(prop)
				}
			}
		}
	}
}

func (m *Matcher) NodeMatchesProperties(node *graph.Node, labels []string, props map[string]interface{}) bool {
	if len(labels) > 0 {
		matchedLabel := false
		for _, l1 := range node.Labels {
			for _, l2 := range labels {
				if l1 == l2 {
					matchedLabel = true
					break
				}
			}
		}
		if !matchedLabel {
			return false
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
			var expectedVal int64
			switch val := v.(type) {
			case int64:
				expectedVal = val
			case int:
				expectedVal = int64(val)
			}
			if nodeVal != expectedVal {
				return false
			}
		}
	}
	return true
}

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
	}
	return nil
}

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
