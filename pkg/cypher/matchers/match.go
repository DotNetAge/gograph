// Package matchers provides pattern matching functionality for Cypher queries.
// It implements the MATCH clause execution by traversing the graph and finding
// nodes and relationships that match the specified patterns.
//
// The matcher supports:
//   - Node pattern matching with labels and properties
//   - Relationship pattern matching with types and directions
//   - Path traversal through the graph
//   - WHERE clause filtering
//   - Property comparisons
//
// Example:
//
//	matcher := matchers.NewMatcher(store, index)
//	rows, columns, err := matcher.ExecuteMatch(matchStmt, params)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, row := range rows {
//	    fmt.Println(row)
//	}
package matchers

import (
	"fmt"
	"strings"

	"github.com/DotNetAge/gograph/pkg/cypher/ast"
	"github.com/DotNetAge/gograph/pkg/cypher/utils"
	"github.com/DotNetAge/gograph/pkg/graph"
	"github.com/DotNetAge/gograph/pkg/storage"
)

// Matcher executes pattern matching for Cypher MATCH queries.
// It traverses the graph to find nodes and relationships that match
// the specified patterns.
type Matcher struct {
	// Store is the underlying storage database.
	Store *storage.DB

	// Index provides efficient lookups for nodes and relationships.
	Index *graph.Index
}

// NewMatcher creates a new Matcher instance.
//
// Parameters:
//   - store: The storage database
//   - index: The graph index for efficient lookups
//
// Returns a new Matcher instance.
//
// Example:
//
//	matcher := matchers.NewMatcher(store, index)
func NewMatcher(store *storage.DB, index *graph.Index) *Matcher {
	return &Matcher{Store: store, Index: index}
}

// ExecuteMatch executes a MATCH statement and returns the matched rows.
//
// Parameters:
//   - stmt: The MATCH statement AST node
//   - params: Query parameters for parameterized queries
//
// Returns the matched rows, column names, and any error encountered.
//
// Example:
//
//	rows, columns, err := matcher.ExecuteMatch(matchStmt, map[string]interface{}{
//	    "name": "Alice",
//	})
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

// executeMatchClause executes a single MATCH clause.
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

// executePath executes a path pattern and returns all matching paths.
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

// findNodes finds all nodes matching the given node pattern.
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

// nodeMatchesProperties checks if a node matches the given labels and properties.
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

// traversePath traverses the graph from the start node following the path pattern.
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

// EvaluateExpression evaluates a WHERE clause expression against a path.
//
// Parameters:
//   - path: The matched path containing nodes and relationships
//   - expr: The expression to evaluate
//   - params: Query parameters
//
// Returns true if the expression evaluates to true for the given path.
func (m *Matcher) EvaluateExpression(path map[string]interface{}, expr ast.Expr, params map[string]interface{}) bool {
	switch e := expr.(type) {
	case *ast.BinaryExpr:
		return m.evaluateBinaryExpr(path, e, params)
	case *ast.InExpr:
		return m.evaluateInExpr(path, e, params)
	default:
		return true
	}
}

// evaluateBinaryExpr evaluates a binary comparison expression (e.g., n.name = "Alice").
func (m *Matcher) evaluateBinaryExpr(path map[string]interface{}, comp *ast.BinaryExpr, params map[string]interface{}) bool {

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

// compareValues compares two values using the given operator.
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

// evaluateInExpr evaluates an IN expression (e.g., $chunkID IN n.source_chunk_ids).
// It supports checking membership of a scalar value in either a string slice property
// or a list literal.
func (m *Matcher) evaluateInExpr(path map[string]interface{}, inExpr *ast.InExpr, params map[string]interface{}) bool {
	leftVal := m.resolveExprValue(path, inExpr.Left, params)
	if leftVal == nil {
		return false
	}

	rightVal := m.resolveExprValue(path, inExpr.Right, params)
	if rightVal == nil {
		return false
	}

	leftStr := fmt.Sprintf("%v", leftVal)

	switch rv := rightVal.(type) {
	case []string:
		for _, item := range rv {
			if item == leftStr {
				return true
			}
		}
		return false
	case string:
		// Property serialized as comma-separated string (legacy fallback)
		for _, item := range strings.Split(rv, ",") {
			if strings.TrimSpace(item) == leftStr {
				return true
			}
		}
		return false
	case []interface{}:
		for _, item := range rv {
			if fmt.Sprintf("%v", item) == leftStr {
				return true
			}
		}
		return false
	}

	return false
}

// resolveExprValue resolves an ast.Expr to a Go value using the path and params.
func (m *Matcher) resolveExprValue(path map[string]interface{}, expr ast.Expr, params map[string]interface{}) interface{} {
	switch e := expr.(type) {
	case *ast.PropertyAccessExpr:
		if ident, ok := e.Target.(*ast.Ident); ok {
			if obj, ok := path[ident.Name]; ok {
				switch o := obj.(type) {
				case *graph.Node:
					if prop, exists := o.Properties[e.Property]; exists {
						return m.PropertyToInterface(prop)
					}
				case *graph.Relationship:
					if prop, exists := o.Properties[e.Property]; exists {
						return m.PropertyToInterface(prop)
					}
				}
			}
		}
	case *ast.Ident:
		if obj, ok := path[e.Name]; ok {
			return obj
		}
	case *ast.Param:
		return params[e.Name]
	case *ast.ParamExpr:
		return params[e.Name]
	case *ast.StringLit:
		return e.Value
	case *ast.IntegerLit:
		return e.Value
	case *ast.FloatLit:
		return e.Value
	case *ast.BoolLit:
		return e.Value
	case *ast.ListExpr:
		var items []interface{}
		for _, item := range e.Elements {
			items = append(items, m.resolveExprValue(path, item, params))
		}
		return items
	}
	return nil
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

// getColumnName returns the column name for a RETURN item.
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

// exprToValue converts an expression to its value.
func (m *Matcher) exprToValue(expr ast.Expr, params map[string]interface{}) interface{} {
	return utils.ExprToValue(expr, params)
}

// propertyMatches checks if a property value matches the expected value.
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

// PropertyToInterface converts a PropertyValue to a Go interface{}.
// Delegates to graph.PropertyValue.InterfaceValue().
func (m *Matcher) PropertyToInterface(prop graph.PropertyValue) interface{} {
	return prop.InterfaceValue()
}

func init() {
	_ = strings.Builder{}
}

// ToFloat64 converts a value to float64.
//
// Parameters:
//   - v: The value to convert
//
// Returns the float64 value and true if conversion succeeded.
//
// Example:
//
//	if f, ok := ToFloat64(42); ok {
//	    fmt.Println(f) // 42.0
//	}
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

// ToInt64 converts a value to int64.
//
// Parameters:
//   - v: The value to convert
//
// Returns the int64 value and true if conversion succeeded.
//
// Example:
//
//	if i, ok := ToInt64(42.5); ok {
//	    fmt.Println(i) // 42
//	}
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

// NewMatcherForMerge creates a new Matcher for MERGE operations.
// This is an alias for NewMatcher.
//
// Parameters:
//   - store: The storage database
//   - index: The graph index for efficient lookups
//
// Returns a new Matcher instance.
func NewMatcherForMerge(store *storage.DB, index *graph.Index) *Matcher {
	return &Matcher{Store: store, Index: index}
}

func (m *Matcher) ExecutePathForMerge(path *ast.PathExpr, params map[string]interface{}) ([]map[string]interface{}, error) {
	return m.executePath(path, params)
}
