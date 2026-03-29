package matchers

import (
	"fmt"
	"strings"

	"github.com/DotNetAge/gograph/pkg/cypher/ast"
	"github.com/DotNetAge/gograph/pkg/cypher/utils"
	"github.com/DotNetAge/gograph/pkg/graph"
	"github.com/DotNetAge/gograph/pkg/storage"
)

type Matcher struct {
	Store *storage.DB
	Index *graph.Index
}

func NewMatcher(store *storage.DB, index *graph.Index) *Matcher {
	return &Matcher{Store: store, Index: index}
}

func (m *Matcher) ExecuteMatch(stmt *ast.MatchStmt, params map[string]interface{}) (rows []map[string]interface{}, columns []string, err error) {
	var matchedPaths []map[string]interface{}

	for _, clause := range stmt.Clauses {
		switch c := clause.(type) {
		case *ast.MatchClause:
			paths, err := m.executeMatchClause(c, params)
			if err != nil {
				return nil, nil, err
			}
			matchedPaths = append(matchedPaths, paths...)
		}
	}

	var whereExpr ast.Expr
	var returnExpr *ast.ReturnExpr

	for _, clause := range stmt.Clauses {
		switch c := clause.(type) {
		case *ast.WhereExpr:
			whereExpr = c.Expr
		case *ast.ReturnExpr:
			returnExpr = c
		}
	}

	if whereExpr != nil {
		var filteredPaths []map[string]interface{}
		for _, path := range matchedPaths {
			if m.EvaluateExpression(path, whereExpr, params) {
				filteredPaths = append(filteredPaths, path)
			}
		}
		matchedPaths = filteredPaths
	}

	for _, path := range matchedPaths {
		row := make(map[string]interface{})
		if returnExpr != nil {
			for i := range returnExpr.Items {
				m.fillRow(row, returnExpr.Items[i], path)
			}
		} else {
			for k, v := range path {
				row[k] = v
			}
		}
		rows = append(rows, row)
	}

	if returnExpr != nil {
		for i := range returnExpr.Items {
			columns = append(columns, m.getColumnName(returnExpr.Items[i]))
		}
	} else if len(rows) > 0 {
		for k := range rows[0] {
			columns = append(columns, k)
		}
	}

	return rows, columns, nil
}

func (m *Matcher) executeMatchClause(clause *ast.MatchClause, params map[string]interface{}) ([]map[string]interface{}, error) {
	if clause.Pattern == nil {
		return nil, nil
	}

	var matchedPaths []map[string]interface{}

	for _, part := range clause.Pattern.Parts {
		if part.Path == nil {
			continue
		}

		paths, err := m.executePath(part.Path, params)
		if err != nil {
			return nil, err
		}
		matchedPaths = append(matchedPaths, paths...)
	}

	if clause.Where != nil {
		var filteredPaths []map[string]interface{}
		for _, path := range matchedPaths {
			if m.EvaluateExpression(path, clause.Where.Expr, params) {
				filteredPaths = append(filteredPaths, path)
			}
		}
		matchedPaths = filteredPaths
	}

	return matchedPaths, nil
}

func (m *Matcher) executePath(path *ast.PathExpr, params map[string]interface{}) ([]map[string]interface{}, error) {
	if len(path.Nodes) == 0 {
		return nil, nil
	}

	var matchedPaths []map[string]interface{}

	if len(path.Relationships) == 0 {
		node := path.Nodes[0]
		nodes := m.findNodes(node, params)
		for _, n := range nodes {
			p := make(map[string]interface{})
			if node.Variable != "" {
				p[node.Variable] = n
			}
			matchedPaths = append(matchedPaths, p)
		}
		return matchedPaths, nil
	}

	startNode := path.Nodes[0]
	startNodes := m.findNodes(startNode, params)

	for _, start := range startNodes {
		paths := m.traversePath(start, path, 0, make(map[string]bool))
		matchedPaths = append(matchedPaths, paths...)
	}

	return matchedPaths, nil
}

func (m *Matcher) findNodes(nodePattern *ast.NodePattern, params map[string]interface{}) []*graph.Node {
	var nodes []*graph.Node

	if len(nodePattern.Labels) > 0 {
		ids, _ := m.Index.LookupByLabel(nodePattern.Labels[0])
		for _, id := range ids {
			data, err := m.Store.Get(storage.NodeKey(id))
			if err != nil {
				continue
			}
			var node graph.Node
			if err := storage.Unmarshal(data, &node); err == nil {
				if m.nodeMatchesProperties(&node, nodePattern.Labels, nodePattern.Properties, params) {
					nodes = append(nodes, &node)
				}
			}
		}
	} else {
		iter, _ := m.Store.NewIter(nil)
		defer iter.Close()
		for iter.SeekGE([]byte(storage.KeyPrefixNode)); iter.Valid(); iter.Next() {
			key := iter.Key()
			if len(key) < 5 || string(key)[:5] != storage.KeyPrefixNode {
				break
			}
			var node graph.Node
			if err := storage.Unmarshal(iter.Value(), &node); err == nil {
				if m.nodeMatchesProperties(&node, nil, nodePattern.Properties, params) {
					nodes = append(nodes, &node)
				}
			}
		}
	}

	return nodes
}

func (m *Matcher) nodeMatchesProperties(node *graph.Node, labels []string, props map[string]ast.Expr, params map[string]interface{}) bool {
	for _, label := range labels {
		found := false
		for _, nodeLabel := range node.Labels {
			if nodeLabel == label {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	for k, expr := range props {
		prop, exists := node.Properties[k]
		if !exists {
			return false
		}
		expected := m.exprToValue(expr, params)
		if !m.propertyMatches(prop, expected) {
			return false
		}
	}
	return true
}

func (m *Matcher) traversePath(start *graph.Node, path *ast.PathExpr, relIndex int, visited map[string]bool) []map[string]interface{} {
	if relIndex >= len(path.Relationships) {
		p := make(map[string]interface{})
		if path.Nodes[0].Variable != "" {
			p[path.Nodes[0].Variable] = start
		}
		return []map[string]interface{}{p}
	}

	rel := path.Relationships[relIndex]
	nextNodePattern := path.Nodes[relIndex+1]

	if visited[start.ID] {
		return nil
	}
	visited[start.ID] = true

	var results []map[string]interface{}

	adj := graph.NewAdjacencyList(m.Store)
	relIDs, _ := adj.GetAllRelated(start.ID)

	for _, relID := range relIDs {
		relData, err := m.Store.Get(storage.RelKey(relID))
		if err != nil {
			continue
		}
		var relationship graph.Relationship
		if err := storage.Unmarshal(relData, &relationship); err != nil {
			continue
		}

		if len(rel.Types) > 0 {
			typeMatch := false
			for _, t := range rel.Types {
				if relationship.Type == t {
					typeMatch = true
					break
				}
			}
			if !typeMatch {
				continue
			}
		}

		var endNodeID string
		switch rel.Direction {
		case ast.DirectionOutgoing:
			if relationship.StartNodeID != start.ID {
				continue
			}
			endNodeID = relationship.EndNodeID
		case ast.DirectionIncoming:
			if relationship.EndNodeID != start.ID {
				continue
			}
			endNodeID = relationship.StartNodeID
		default:
			if relationship.StartNodeID == start.ID {
				endNodeID = relationship.EndNodeID
			} else if relationship.EndNodeID == start.ID {
				endNodeID = relationship.StartNodeID
			} else {
				continue
			}
		}

		endData, err := m.Store.Get(storage.NodeKey(endNodeID))
		if err != nil {
			continue
		}
		var endNode graph.Node
		if err := storage.Unmarshal(endData, &endNode); err != nil {
			continue
		}

		if !m.nodeMatchesProperties(&endNode, nextNodePattern.Labels, nextNodePattern.Properties, nil) {
			continue
		}

		newVisited := make(map[string]bool)
		for k, v := range visited {
			newVisited[k] = v
		}

		subPaths := m.traversePath(&endNode, path, relIndex+1, newVisited)
		for _, subPath := range subPaths {
			p := make(map[string]interface{})
			for k, v := range subPath {
				p[k] = v
			}
			if path.Nodes[0].Variable != "" {
				p[path.Nodes[0].Variable] = start
			}
			if rel.Variable != "" {
				p[rel.Variable] = &relationship
			}
			if nextNodePattern.Variable != "" {
				p[nextNodePattern.Variable] = &endNode
			}
			results = append(results, p)
		}
	}

	return results
}

func (m *Matcher) EvaluateExpression(path map[string]interface{}, expr ast.Expr, params map[string]interface{}) bool {
	comp, ok := expr.(*ast.BinaryExpr)
	if !ok {
		return true
	}

	var leftVal interface{}
	var leftProp string

	switch left := comp.Left.(type) {
	case *ast.PropertyAccessExpr:
		if ident, ok := left.Target.(*ast.Ident); ok {
			leftProp = left.Property
			if obj, ok := path[ident.Name]; ok {
				switch o := obj.(type) {
				case *graph.Node:
					if prop, exists := o.Properties[leftProp]; exists {
						leftVal = m.PropertyToInterface(prop)
					}
				case *graph.Relationship:
					if prop, exists := o.Properties[leftProp]; exists {
						leftVal = m.PropertyToInterface(prop)
					}
				}
			}
		}
	case *ast.Ident:
		if obj, ok := path[left.Name]; ok {
			leftVal = obj
		}
	}

	if leftVal == nil {
		return false
	}

	var rightVal interface{}
	switch right := comp.Right.(type) {
	case *ast.IntegerLit:
		rightVal = right.Value
	case *ast.FloatLit:
		rightVal = right.Value
	case *ast.StringLit:
		rightVal = right.Value
	case *ast.BoolLit:
		rightVal = right.Value
	case *ast.Param:
		rightVal = params[right.Name]
	case *ast.ParamExpr:
		rightVal = params[right.Name]
	}

	return m.compareValues(leftVal, rightVal, comp.Operator)
}

func (m *Matcher) compareValues(left, right interface{}, op string) bool {
	leftFloat, leftIsFloat := ToFloat64(left)
	rightFloat, rightIsFloat := ToFloat64(right)

	if leftIsFloat && rightIsFloat {
		switch op {
		case ">":
			return leftFloat > rightFloat
		case ">=":
			return leftFloat >= rightFloat
		case "<":
			return leftFloat < rightFloat
		case "<=":
			return leftFloat <= rightFloat
		case "=", "==":
			return leftFloat == rightFloat
		case "!=", "<>":
			return leftFloat != rightFloat
		}
	}

	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)

	switch op {
	case "=", "==":
		return leftStr == rightStr
	case "!=", "<>":
		return leftStr != rightStr
	case ">":
		return leftStr > rightStr
	case ">=":
		return leftStr >= rightStr
	case "<":
		return leftStr < rightStr
	case "<=":
		return leftStr <= rightStr
	}

	return false
}

func (m *Matcher) fillRow(row map[string]interface{}, item *ast.ReturnItemExpr, path map[string]interface{}) {
	switch expr := item.Expr.(type) {
	case *ast.PropertyAccessExpr:
		if ident, ok := expr.Target.(*ast.Ident); ok {
			obj := path[ident.Name]
			if obj == nil {
				return
			}
			switch o := obj.(type) {
			case *graph.Node:
				if expr.Property == "" {
					row[ident.Name] = o
				} else if prop, exists := o.Properties[expr.Property]; exists {
					row[ident.Name+"."+expr.Property] = m.PropertyToInterface(prop)
				}
			case *graph.Relationship:
				if expr.Property == "" {
					row[ident.Name] = o
				} else if prop, exists := o.Properties[expr.Property]; exists {
					row[ident.Name+"."+expr.Property] = m.PropertyToInterface(prop)
				}
			}
		}
	case *ast.Ident:
		if obj, ok := path[expr.Name]; ok {
			row[expr.Name] = obj
		}
	}
}

func (m *Matcher) getColumnName(item *ast.ReturnItemExpr) string {
	if item.Alias != "" {
		return item.Alias
	}
	switch expr := item.Expr.(type) {
	case *ast.PropertyAccessExpr:
		if ident, ok := expr.Target.(*ast.Ident); ok {
			if expr.Property == "" {
				return ident.Name
			}
			return ident.Name + "." + expr.Property
		}
	case *ast.Ident:
		return expr.Name
	}
	return ""
}

func (m *Matcher) exprToValue(expr ast.Expr, params map[string]interface{}) interface{} {
	return utils.ExprToValue(expr, params)
}

func (m *Matcher) propertyMatches(prop graph.PropertyValue, expected interface{}) bool {
	switch prop.Type() {
	case graph.PropertyTypeString:
		return prop.StringValue() == fmt.Sprintf("%v", expected)
	case graph.PropertyTypeInt:
		if v, ok := ToInt64(expected); ok {
			return prop.IntValue() == v
		}
	case graph.PropertyTypeFloat:
		if v, ok := ToFloat64(expected); ok {
			return prop.FloatValue() == v
		}
	case graph.PropertyTypeBool:
		if v, ok := expected.(bool); ok {
			return prop.BoolValue() == v
		}
	}
	return false
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
	default:
		return nil
	}
}

func init() {
	_ = strings.Builder{}
}

func ToFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case uint:
		return float64(val), true
	case uint32:
		return float64(val), true
	case uint64:
		return float64(val), true
	}
	return 0, false
}

func ToInt64(v interface{}) (int64, bool) {
	switch val := v.(type) {
	case int:
		return int64(val), true
	case int32:
		return int64(val), true
	case int64:
		return val, true
	case uint:
		return int64(val), true
	case uint32:
		return int64(val), true
	case uint64:
		return int64(val), true
	}
	return 0, false
}

func NewMatcherForMerge(store *storage.DB, index *graph.Index) *Matcher {
	return &Matcher{Store: store, Index: index}
}

func (m *Matcher) ExecutePathForMerge(path *ast.PathExpr, params map[string]interface{}) ([]map[string]interface{}, error) {
	return m.executePath(path, params)
}
