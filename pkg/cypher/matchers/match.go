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
	"sort"
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
	return m.ExecuteMatchWithContext(stmt, params, nil)
}

func (m *Matcher) ExecuteMatchWithContext(stmt *ast.MatchStmt, params map[string]interface{}, contextRows []map[string]interface{}) (rows []map[string]interface{}, columns []string, err error) {
	if contextRows != nil && len(contextRows) > 0 {
		return m.executeMatchWithUnwindContext(stmt, params, contextRows)
	}

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

	if returnExpr != nil && m.hasAggregation(returnExpr.Items) {
		pathsForAgg := make([]map[string]interface{}, len(matchedPaths))
		for i, p := range matchedPaths {
			pathsForAgg[i] = p
		}
		return m.computeAggregation(pathsForAgg, returnExpr)
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

	if returnExpr != nil {
		if returnExpr.OrderBy != nil && len(rows) > 0 {
			rows = m.sortRows(rows, returnExpr.OrderBy, params)
		}

		skip := 0
		if returnExpr.Skip != nil {
			skip = int(m.resolveIntValue(returnExpr.Skip, params))
			if skip < 0 {
				skip = 0
			}
			if skip > len(rows) {
				skip = len(rows)
			}
		}

		if skip > 0 {
			rows = rows[skip:]
		}

		if returnExpr.Limit != nil {
			limit := int(m.resolveIntValue(returnExpr.Limit, params))
			if limit > 0 && limit < len(rows) {
				rows = rows[:limit]
			}
		}
	}

	return rows, columns, nil
}

func (m *Matcher) executeMatchWithUnwindContext(stmt *ast.MatchStmt, params map[string]interface{}, contextRows []map[string]interface{}) (rows []map[string]interface{}, columns []string, err error) {
	var allMatchedPaths []map[string]interface{}
	var returnExpr *ast.ReturnExpr

	for _, clause := range stmt.Clauses {
		switch c := clause.(type) {
		case *ast.WhereExpr:
		case *ast.ReturnExpr:
			returnExpr = c
		}
	}

	for _, ctxRow := range contextRows {
		mergedParams := make(map[string]interface{})
		for k, v := range params {
			mergedParams[k] = v
		}
		for k, v := range ctxRow {
			mergedParams[k] = v
		}

		for _, clause := range stmt.Clauses {
			switch c := clause.(type) {
			case *ast.MatchClause:
				paths, pathErr := m.executeMatchClauseWithContext(c, mergedParams, ctxRow)
				if pathErr != nil {
					return nil, nil, pathErr
				}
				allMatchedPaths = append(allMatchedPaths, paths...)
			}
		}
	}

	if returnExpr != nil && m.hasAggregation(returnExpr.Items) {
		pathsForAgg := make([]map[string]interface{}, len(allMatchedPaths))
		for i, p := range allMatchedPaths {
			pathsForAgg[i] = p
		}
		return m.computeAggregation(pathsForAgg, returnExpr)
	}

	if returnExpr != nil {
		for i := range returnExpr.Items {
			columns = append(columns, m.getColumnName(returnExpr.Items[i]))
		}
	}

	for _, path := range allMatchedPaths {
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

	return rows, columns, nil
}

func (m *Matcher) executeMatchClauseWithContext(clause *ast.MatchClause, params map[string]interface{}, ctxRow map[string]interface{}) ([]map[string]interface{}, error) {
	if clause.Pattern == nil {
		return nil, nil
	}

	var matchedPaths []map[string]interface{}

	for _, part := range clause.Pattern.Parts {
		if part.Path == nil {
			continue
		}

		paths, err := m.executePathWithContext(part.Path, params, ctxRow)
		if err != nil {
			return nil, err
		}
		matchedPaths = append(matchedPaths, paths...)
	}

	if clause.Where != nil {
		var filteredPaths []map[string]interface{}
		for _, path := range matchedPaths {
			if m.EvaluateExpressionWithContext(path, clause.Where.Expr, params, ctxRow) {
				filteredPaths = append(filteredPaths, path)
			}
		}
		matchedPaths = filteredPaths
	}

	return matchedPaths, nil
}

func (m *Matcher) executePathWithContext(path *ast.PathExpr, params map[string]interface{}, ctxRow map[string]interface{}) ([]map[string]interface{}, error) {
	if len(path.Nodes) == 0 {
		return nil, nil
	}

	var matchedPaths []map[string]interface{}

	if len(path.Relationships) == 0 {
		node := path.Nodes[0]
		nodes := m.findNodesWithContext(node, params, ctxRow)
		for _, n := range nodes {
			p := make(map[string]interface{})
			if node.Variable != "" {
				p[node.Variable] = n
			}
			for k, v := range ctxRow {
				p[k] = v
			}
			matchedPaths = append(matchedPaths, p)
		}
		return matchedPaths, nil
	}

	startNode := path.Nodes[0]
	startNodes := m.findNodesWithContext(startNode, params, ctxRow)

	for _, start := range startNodes {
		paths := m.traversePath(start, path, 0, make(map[string]bool))
		for _, p := range paths {
			for k, v := range ctxRow {
				p[k] = v
			}
			matchedPaths = append(matchedPaths, p)
		}
	}

	return matchedPaths, nil
}

func (m *Matcher) findNodesWithContext(nodePattern *ast.NodePattern, params map[string]interface{}, ctxRow map[string]interface{}) []*graph.Node {
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
				if m.nodeMatchesPropertiesWithContext(&node, nodePattern.Labels, nodePattern.Properties, params, ctxRow) {
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
				if m.nodeMatchesPropertiesWithContext(&node, nil, nodePattern.Properties, params, ctxRow) {
					nodes = append(nodes, &node)
				}
			}
		}
	}

	return nodes
}

func (m *Matcher) nodeMatchesPropertiesWithContext(node *graph.Node, labels []string, props map[string]ast.Expr, params map[string]interface{}, ctxRow map[string]interface{}) bool {
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
		expected := m.exprToValueWithContext(expr, params, ctxRow)
		if !m.propertyMatches(prop, expected) {
			return false
		}
	}
	return true
}

func (m *Matcher) exprToValueWithContext(expr ast.Expr, params map[string]interface{}, ctxRow map[string]interface{}) interface{} {
	switch e := expr.(type) {
	case *ast.Ident:
		if val, ok := ctxRow[e.Name]; ok {
			return val
		}
		if val, ok := params[e.Name]; ok {
			return val
		}
	case *ast.IntegerLit:
		return e.Value
	case *ast.FloatLit:
		return e.Value
	case *ast.StringLit:
		return e.Value
	case *ast.BoolLit:
		return e.Value
	case *ast.Param:
		return params[e.Name]
	case *ast.ParamExpr:
		return params[e.Name]
	}
	return nil
}

func (m *Matcher) EvaluateExpressionWithContext(path map[string]interface{}, expr ast.Expr, params map[string]interface{}, ctxRow map[string]interface{}) bool {
	switch e := expr.(type) {
	case *ast.BinaryExpr:
		return m.evaluateBinaryExprWithContext(path, e, params, ctxRow)
	default:
		return true
	}
}

func (m *Matcher) evaluateBinaryExprWithContext(path map[string]interface{}, expr *ast.BinaryExpr, params map[string]interface{}, ctxRow map[string]interface{}) bool {
	op := expr.Operator
	if op == "AND" {
		return m.EvaluateExpressionWithContext(path, expr.Left, params, ctxRow) && m.EvaluateExpressionWithContext(path, expr.Right, params, ctxRow)
	}
	if op == "OR" {
		return m.EvaluateExpressionWithContext(path, expr.Left, params, ctxRow) || m.EvaluateExpressionWithContext(path, expr.Right, params, ctxRow)
	}

	leftVal := m.resolveExprValueWithContext(path, expr.Left, params, ctxRow)
	rightVal := m.resolveExprValueWithContext(path, expr.Right, params, ctxRow)

	if leftVal == nil || rightVal == nil {
		return false
	}

	return m.compareValues(leftVal, rightVal, op)
}

func (m *Matcher) resolveExprValueWithContext(path map[string]interface{}, expr ast.Expr, params map[string]interface{}, ctxRow map[string]interface{}) interface{} {
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
			if val, ok := ctxRow[ident.Name]; ok {
				return val
			}
		}
	case *ast.Ident:
		if val, ok := ctxRow[e.Name]; ok {
			return val
		}
		if val, ok := path[e.Name]; ok {
			return val
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
			items = append(items, m.resolveExprValueWithContext(path, item, params, ctxRow))
		}
		return items
	case *ast.FuncCall:
		if len(e.Args) > 0 {
			resolved := m.resolveFuncArg(path, e.Args[0])
			if resolved != nil {
				return m.evaluateStandardFunc(e.Name, resolved)
			}
		}
	}
	return nil
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
// If statistics are available, it selects the most selective label
// (the one with the fewest nodes) to minimize the initial scan.
func (m *Matcher) findNodes(nodePattern *ast.NodePattern, params map[string]interface{}) []*graph.Node {
	var nodes []*graph.Node

	if len(nodePattern.Labels) > 0 {
		// Use statistics to select the most selective label
		label := nodePattern.Labels[0]
		if stats := m.Index.Stats(); stats != nil && len(nodePattern.Labels) > 1 {
			label = stats.SelectBestLabel(nodePattern.Labels)
		}
		ids, _ := m.Index.LookupByLabel(label)
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
	switch comp.Operator {
	case "AND", "and", "&&":
		return m.EvaluateExpression(path, comp.Left, params) && m.EvaluateExpression(path, comp.Right, params)
	case "OR", "or", "||":
		return m.EvaluateExpression(path, comp.Left, params) || m.EvaluateExpression(path, comp.Right, params)
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
	case "STARTS WITH":
		return strings.HasPrefix(leftStr, rightStr)
	case "ENDS WITH":
		return strings.HasSuffix(leftStr, rightStr)
	case "CONTAINS":
		return strings.Contains(leftStr, rightStr)
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
	case *ast.FuncCall:
		if len(e.Args) > 0 {
			resolved := m.resolveFuncArg(path, e.Args[0])
			if resolved != nil {
				return m.evaluateStandardFunc(e.Name, resolved)
			}
		}
	}
	return nil
}

func (m *Matcher) resolveIntValue(expr ast.Expr, params map[string]interface{}) int64 {
	switch e := expr.(type) {
	case *ast.IntegerLit:
		return e.Value
	case *ast.Param:
		if v, ok := params[e.Name]; ok {
			if i, ok := ToInt64(v); ok {
				return i
			}
		}
	case *ast.ParamExpr:
		if v, ok := params[e.Name]; ok {
			if i, ok := ToInt64(v); ok {
				return i
			}
		}
	case *ast.Ident:
		if v, ok := params[e.Name]; ok {
			if i, ok := ToInt64(v); ok {
				return i
			}
		}
	}
	return 0
}

func (m *Matcher) sortRows(rows []map[string]interface{}, orderBy *ast.OrderByExpr, params map[string]interface{}) []map[string]interface{} {
	if len(orderBy.Items) == 0 {
		return rows
	}

	sort.Slice(rows, func(i, j int) bool {
		for _, item := range orderBy.Items {
			valI := m.getSortValue(rows[i], item.Expr, params)
			valJ := m.getSortValue(rows[j], item.Expr, params)

			if valI == nil && valJ == nil {
				continue
			}
			if valI == nil {
				return !item.Descending
			}
			if valJ == nil {
				return item.Descending
			}

			less := m.compareSortValues(valI, valJ)
			if item.Descending {
				less = !less
			}
			if less {
				return true
			}
			if m.compareSortValues(valI, valJ) && !m.compareSortValues(valJ, valI) {
				continue
			}
		}
		return false
	})

	return rows
}

func (m *Matcher) getSortValue(row map[string]interface{}, expr ast.Expr, params map[string]interface{}) interface{} {
	switch e := expr.(type) {
	case *ast.PropertyAccessExpr:
		if ident, ok := e.Target.(*ast.Ident); ok {
			fullKey := ident.Name + "." + e.Property
			if val, exists := row[fullKey]; exists {
				return val
			}
			obj := row[ident.Name]
			if obj == nil {
				return nil
			}
			if node, ok := obj.(*graph.Node); ok {
				if prop, exists := node.Properties[e.Property]; exists {
					return m.PropertyToInterface(prop)
				}
				return nil
			}
			if rel, ok := obj.(*graph.Relationship); ok {
				if prop, exists := rel.Properties[e.Property]; exists {
					return m.PropertyToInterface(prop)
				}
				return nil
			}
			return nil
		}
	case *ast.Ident:
		if val, ok := row[e.Name]; ok {
			return val
		}
	}
	return nil
}

func (m *Matcher) compareSortValues(a, b interface{}) bool {
	aFloat, aOk := ToFloat64(a)
	bFloat, bOk := ToFloat64(b)

	if aOk && bOk {
		return aFloat < bFloat
	}

	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return aStr < bStr
}

func (m *Matcher) ExecuteUnwind(stmt *ast.UnwindStmt, params map[string]interface{}) (rows []map[string]interface{}, columns []string, err error) {
	listVal := m.resolveExprValue(nil, stmt.List, params)
	if listVal == nil {
		return rows, columns, nil
	}

	var list []interface{}
	switch v := listVal.(type) {
	case []interface{}:
		list = v
	case []string:
		for _, s := range v {
			list = append(list, s)
		}
	default:
		return rows, columns, nil
	}

	for _, item := range list {
		row := make(map[string]interface{})
		row[stmt.Variable] = item
		rows = append(rows, row)
	}

	if stmt.Return != nil {
		for i := range stmt.Return.Items {
			columns = append(columns, m.getColumnName(stmt.Return.Items[i]))
		}
		var processedRows []map[string]interface{}
		for _, row := range rows {
			newRow := make(map[string]interface{})
			for i := range stmt.Return.Items {
				m.fillRow(newRow, stmt.Return.Items[i], row)
			}
			processedRows = append(processedRows, newRow)
		}
		return processedRows, columns, nil
	}

	if len(rows) > 0 {
		columns = []string{stmt.Variable}
	}

	return rows, columns, nil
}

func (m *Matcher) ProcessReturnStmt(rows []map[string]interface{}, returnExpr *ast.ReturnExpr) ([]map[string]interface{}, []string, error) {
	if returnExpr == nil {
		return rows, nil, nil
	}

	hasAggregation := m.hasAggregation(returnExpr.Items)
	if hasAggregation {
		aggResult, cols, err := m.computeAggregation(rows, returnExpr)
		return aggResult, cols, err
	}

	var processedRows []map[string]interface{}
	for _, row := range rows {
		newRow := make(map[string]interface{})
		for i := range returnExpr.Items {
			m.fillRow(newRow, returnExpr.Items[i], row)
		}
		processedRows = append(processedRows, newRow)
	}

	var columns []string
	for i := range returnExpr.Items {
		columns = append(columns, m.getColumnName(returnExpr.Items[i]))
	}

	if returnExpr.OrderBy != nil && len(processedRows) > 0 {
		processedRows = m.sortRows(processedRows, returnExpr.OrderBy, nil)
	}

	skip := 0
	if returnExpr.Skip != nil {
		skip = int(m.resolveIntValue(returnExpr.Skip, nil))
		if skip < 0 {
			skip = 0
		}
		if skip > len(processedRows) {
			skip = len(processedRows)
		}
	}

	if skip > 0 {
		processedRows = processedRows[skip:]
	}

	if returnExpr.Limit != nil {
		limit := int(m.resolveIntValue(returnExpr.Limit, nil))
		if limit > 0 && limit < len(processedRows) {
			processedRows = processedRows[:limit]
		}
	}

	return processedRows, columns, nil
}
func (m *Matcher) fillRow(row map[string]interface{}, item *ast.ReturnItemExpr, path map[string]interface{}) {
	// Determine the output key: use alias if provided, otherwise use the expression's natural name.
	outputKey := m.getColumnName(item)

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
					row[outputKey] = o
				} else if prop, exists := o.Properties[expr.Property]; exists {
					row[outputKey] = m.PropertyToInterface(prop)
				}
			case *graph.Relationship:
				if expr.Property == "" {
					row[outputKey] = o
				} else if prop, exists := o.Properties[expr.Property]; exists {
					row[outputKey] = m.PropertyToInterface(prop)
				}
			}
		}
	case *ast.Ident:
		if obj, ok := path[expr.Name]; ok {
			row[outputKey] = obj
		}
	case *ast.FuncCall:
		// Handle standard Cypher built-in functions on the current path.
		if len(expr.Args) > 0 {
			resolved := m.resolveFuncArg(path, expr.Args[0])
			if resolved != nil {
				val := m.evaluateStandardFunc(expr.Name, resolved)
				if val != nil {
					row[outputKey] = val
				}
			}
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
	case *ast.FuncCall:
		return expr.Name
	}
	return ""
}

func (m *Matcher) hasAggregation(items []*ast.ReturnItemExpr) bool {
	for _, item := range items {
		if m.exprHasAggregation(item.Expr) {
			return true
		}
	}
	return false
}

func (m *Matcher) exprHasAggregation(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.FuncCall:
		switch strings.ToLower(e.Name) {
		case "count", "sum", "avg", "min", "max", "collect", "size":
			return true
		}
	case *ast.BinaryExpr:
		if m.exprHasAggregation(e.Left) || m.exprHasAggregation(e.Right) {
			return true
		}
	}
	return false
}

func (m *Matcher) computeAggregation(rows []map[string]interface{}, returnExpr *ast.ReturnExpr) ([]map[string]interface{}, []string, error) {
	var columns []string
	for _, item := range returnExpr.Items {
		columns = append(columns, m.getColumnName(item))
	}

	groupedItems := make(map[string][]map[string]interface{})
	nonAggKeys := m.getNonAggregatedKeys(returnExpr.Items)

	if len(nonAggKeys) == 0 {
		resultRow := make(map[string]interface{})
		for _, item := range returnExpr.Items {
			colName := m.getColumnName(item)
			val := m.evaluateAggregation(item.Expr, rows)
			resultRow[colName] = val
		}
		return []map[string]interface{}{resultRow}, columns, nil
	}

	for _, row := range rows {
		key := m.computeGroupKey(row, nonAggKeys)
		groupedItems[key] = append(groupedItems[key], row)
	}

	var result []map[string]interface{}
	for _, groupRows := range groupedItems {
		resultRow := make(map[string]interface{})
		for _, item := range returnExpr.Items {
			colName := m.getColumnName(item)
			val := m.evaluateAggregation(item.Expr, groupRows)
			resultRow[colName] = val
		}
		for _, key := range nonAggKeys {
			if val := m.getRowValue(groupRows[0], key); val != nil {
				resultRow[key] = val
			}
		}
		result = append(result, resultRow)
	}

	return result, columns, nil
}

func (m *Matcher) getNonAggregatedKeys(items []*ast.ReturnItemExpr) []string {
	var keys []string
	for _, item := range items {
		if !m.exprHasAggregation(item.Expr) {
			key := m.getColumnName(item)
			if key != "" {
				keys = append(keys, key)
			}
		}
	}
	return keys
}

func (m *Matcher) computeGroupKey(row map[string]interface{}, keys []string) string {
	var parts []string
	for _, key := range keys {
		val := m.getRowValue(row, key)
		if val != nil {
			parts = append(parts, fmt.Sprintf("%v", val))
		} else {
			parts = append(parts, "")
		}
	}
	return strings.Join(parts, "|")
}

func (m *Matcher) getRowValue(row map[string]interface{}, key string) interface{} {
	if strings.Contains(key, ".") {
		parts := strings.SplitN(key, ".", 2)
		if val, ok := row[parts[0]]; ok {
			if node, ok := val.(*graph.Node); ok {
				propVal, found := node.GetProperty(parts[1])
				if found {
					return propVal.InterfaceValue()
				}
				return nil
			}
			if m, ok := val.(map[string]interface{}); ok {
				return m[parts[1]]
			}
		}
		return nil
	}
	return row[key]
}

func (m *Matcher) evaluateAggregation(expr ast.Expr, rows []map[string]interface{}) interface{} {
	switch e := expr.(type) {
	case *ast.FuncCall:
		return m.evaluateFuncCall(e, rows)
	case *ast.Ident:
		if len(rows) > 0 {
			if val, ok := rows[0][e.Name]; ok {
				return val
			}
		}
		return nil
	}
	return nil
}

func (m *Matcher) evaluateFuncCall(call *ast.FuncCall, rows []map[string]interface{}) interface{} {
	switch strings.ToLower(call.Name) {
	case "count":
		return m.evalCount(call, rows)
	case "sum":
		return m.evalSum(call, rows)
	case "avg":
		return m.evalAvg(call, rows)
	case "min":
		return m.evalMin(call, rows)
	case "max":
		return m.evalMax(call, rows)
	case "collect", "size":
		return m.evalCollect(call, rows)
	}
	return nil
}

func (m *Matcher) evalCount(call *ast.FuncCall, rows []map[string]interface{}) interface{} {
	if len(call.Args) == 0 {
		return len(rows)
	}
	arg := call.Args[0]
	switch arg.(type) {
	case *ast.StarLit:
		return len(rows)
	case *ast.Ident, *ast.PropertyAccessExpr:
		count := 0
		for _, row := range rows {
			val := m.resolveExprValue(row, arg, nil)
			if val != nil {
				count++
			}
		}
		return count
	}
	return len(rows)
}

func (m *Matcher) evalSum(call *ast.FuncCall, rows []map[string]interface{}) interface{} {
	if len(call.Args) == 0 {
		return 0
	}
	arg := call.Args[0]
	sum := 0.0
	for _, row := range rows {
		val := m.resolveExprValue(row, arg, nil)
		if val != nil {
			if f, ok := ToFloat64(val); ok {
				sum += f
			}
		}
	}
	return sum
}

func (m *Matcher) evalAvg(call *ast.FuncCall, rows []map[string]interface{}) interface{} {
	if len(call.Args) == 0 || len(rows) == 0 {
		return 0.0
	}
	arg := call.Args[0]
	sum := 0.0
	count := 0
	for _, row := range rows {
		val := m.resolveExprValue(row, arg, nil)
		if val != nil {
			if f, ok := ToFloat64(val); ok {
				sum += f
				count++
			}
		}
	}
	if count == 0 {
		return 0.0
	}
	return sum / float64(count)
}

func (m *Matcher) evalMin(call *ast.FuncCall, rows []map[string]interface{}) interface{} {
	if len(call.Args) == 0 {
		return nil
	}
	arg := call.Args[0]
	var minVal interface{}
	for _, row := range rows {
		val := m.resolveExprValue(row, arg, nil)
		if val != nil {
			if minVal == nil {
				minVal = val
			} else if less, ok := compareValues(val, minVal); ok && less {
				minVal = val
			}
		}
	}
	return minVal
}

func (m *Matcher) evalMax(call *ast.FuncCall, rows []map[string]interface{}) interface{} {
	if len(call.Args) == 0 {
		return nil
	}
	arg := call.Args[0]
	var maxVal interface{}
	for _, row := range rows {
		val := m.resolveExprValue(row, arg, nil)
		if val != nil {
			if maxVal == nil {
				maxVal = val
			} else {
				if greater, ok := compareValues(maxVal, val); ok && greater {
					maxVal = val
				}
			}
		}
	}
	return maxVal
}

func (m *Matcher) evalCollect(call *ast.FuncCall, rows []map[string]interface{}) interface{} {
	if len(call.Args) == 0 {
		result := make([]interface{}, len(rows))
		for i, row := range rows {
			result[i] = row
		}
		return result
	}
	arg := call.Args[0]
	result := make([]interface{}, 0, len(rows))
	for _, row := range rows {
		val := m.resolveExprValue(row, arg, nil)
		result = append(result, val)
	}
	return result
}

func compareValues(a, b interface{}) (bool, bool) {
	aFloat, aOk := ToFloat64(a)
	bFloat, bOk := ToFloat64(b)
	if aOk && bOk {
		return aFloat < bFloat, true
	}

	aInt, aOk := ToInt64(a)
	bInt, bOk := ToInt64(b)
	if aOk && bOk {
		return aInt < bInt, true
	}

	aStr, aOk := a.(string)
	bStr, bOk := b.(string)
	if aOk && bOk {
		return aStr < bStr, true
	}

	return false, false
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

// resolveFuncArg resolves the first argument of a Cypher function call.
// The argument is typically an *ast.Ident (variable name) which is looked up
// in the path to get the actual node or relationship object.
func (m *Matcher) resolveFuncArg(path map[string]interface{}, arg ast.Expr) interface{} {
	switch a := arg.(type) {
	case *ast.Ident:
		return path[a.Name]
	case *ast.PropertyAccessExpr:
		if ident, ok := a.Target.(*ast.Ident); ok {
			if obj := path[ident.Name]; obj != nil {
				switch o := obj.(type) {
				case *graph.Node:
					if prop, exists := o.Properties[a.Property]; exists {
						return m.PropertyToInterface(prop)
					}
				case *graph.Relationship:
					if prop, exists := o.Properties[a.Property]; exists {
						return m.PropertyToInterface(prop)
					}
				}
			}
		}
	}
	return nil
}

// evaluateStandardFunc evaluates a standard Cypher built-in function
// (id, labels, properties, type, keys) on the resolved argument.
func (m *Matcher) evaluateStandardFunc(name string, resolved interface{}) interface{} {
	switch strings.ToLower(name) {
	case "id":
		switch v := resolved.(type) {
		case *graph.Node:
			return v.ID
		case *graph.Relationship:
			return v.ID
		}
	case "labels":
		if node, ok := resolved.(*graph.Node); ok {
			return node.Labels
		}
	case "properties":
		switch v := resolved.(type) {
		case *graph.Node:
			return m.nodePropsToMap(v.Properties)
		case *graph.Relationship:
			return m.nodePropsToMap(v.Properties)
		}
	case "type":
		if rel, ok := resolved.(*graph.Relationship); ok {
			return rel.Type
		}
	case "keys":
		switch v := resolved.(type) {
		case *graph.Node:
			keys := make([]string, 0, len(v.Properties))
			for k := range v.Properties {
				keys = append(keys, k)
			}
			return keys
		case *graph.Relationship:
			keys := make([]string, 0, len(v.Properties))
			for k := range v.Properties {
				keys = append(keys, k)
			}
			return keys
		}
	}
	return nil
}

// nodePropsToMap converts map[string]graph.PropertyValue to map[string]interface{}.
func (m *Matcher) nodePropsToMap(props map[string]graph.PropertyValue) map[string]interface{} {
	result := make(map[string]interface{}, len(props))
	for k, v := range props {
		result[k] = m.PropertyToInterface(v)
	}
	return result
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
