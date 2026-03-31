// Package creators provides functionality for creating nodes and relationships
// in the graph database. It implements the CREATE and MERGE clause execution
// for Cypher queries.
//
// The creator supports:
//   - Creating nodes with labels and properties
//   - Creating relationships between nodes
//   - MERGE operations (create if not exists)
//   - Parameterized property values
//
// Example:
//
//	creator := creators.NewCreator(store)
//	nodes, rels, err := creator.ExecuteCreate(tx, createStmt, params)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Created %d nodes and %d relationships\n", nodes, rels)
package creators

import (
	"github.com/DotNetAge/gograph/pkg/cypher/ast"
	"github.com/DotNetAge/gograph/pkg/cypher/matchers"
	"github.com/DotNetAge/gograph/pkg/cypher/utils"
	"github.com/DotNetAge/gograph/pkg/graph"
	"github.com/DotNetAge/gograph/pkg/storage"
	"github.com/DotNetAge/gograph/pkg/tx"
)

// Creator creates nodes and relationships in the graph database.
// It handles the execution of CREATE and MERGE statements.
type Creator struct {
	// Store is the underlying storage database.
	Store *storage.DB

	// index provides efficient lookups for nodes and relationships.
	index *graph.Index

	// adj manages adjacency lists for relationship traversal.
	adj *graph.AdjacencyList
}

// NewCreator creates a new Creator instance.
//
// Parameters:
//   - store: The storage database
//
// Returns a new Creator instance.
//
// Example:
//
//	creator := creators.NewCreator(store)
func NewCreator(store *storage.DB) *Creator {
	return &Creator{
		Store: store,
		index: graph.NewIndex(store),
		adj:   graph.NewAdjacencyList(store),
	}
}

// ExecuteCreate executes a CREATE statement and returns the number of affected elements.
//
// Parameters:
//   - t: The transaction for atomic operations
//   - stmt: The CREATE statement AST node
//   - params: Query parameters for parameterized queries
//
// Returns the number of affected nodes, affected relationships, and any error encountered.
//
// Example:
//
//	nodes, rels, err := creator.ExecuteCreate(tx, createStmt, map[string]interface{}{
//	    "name": "Alice",
//	})
func (c *Creator) ExecuteCreate(t *tx.Transaction, stmt *ast.CreateStmt, params map[string]interface{}) (affectedNodes, affectedRels int, err error) {
	if stmt.Pattern == nil {
		return 0, 0, nil
	}

	for _, part := range stmt.Pattern.Parts {
		if part.Path == nil {
			continue
		}

		nodes, rels, err := c.createPath(t, part.Path, params)
		if err != nil {
			return 0, 0, err
		}
		affectedNodes += nodes
		affectedRels += rels
	}

	return affectedNodes, affectedRels, nil
}

// createPath creates nodes and relationships for a single path pattern.
func (c *Creator) createPath(t *tx.Transaction, path *ast.PathExpr, params map[string]interface{}) (nodes, rels int, err error) {
	if len(path.Nodes) == 0 {
		return 0, 0, nil
	}

	var createdNodes []*graph.Node
	for _, nodePattern := range path.Nodes {
		node := &graph.Node{
			ID:         graph.GenerateID("node"),
			Labels:     nodePattern.Labels,
			Properties: make(map[string]graph.PropertyValue),
		}

		if nodePattern.Properties != nil {
			for k, v := range nodePattern.Properties {
				node.Properties[k] = graph.ToPropertyValue(c.exprToValue(v, params))
			}
		}

		if nodePattern.PropertyExpr != nil {
			if param, ok := nodePattern.PropertyExpr.(*ast.ParamExpr); ok {
				if props, ok := params[param.Name].(map[string]interface{}); ok {
					for k, v := range props {
						node.Properties[k] = graph.ToPropertyValue(v)
					}
				}
			}
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

		createdNodes = append(createdNodes, node)
		nodes++
	}

	for i, relPattern := range path.Relationships {
		if i >= len(createdNodes)-1 {
			break
		}

		startNode := createdNodes[i]
		endNode := createdNodes[i+1]

		relType := ""
		if len(relPattern.Types) > 0 {
			relType = relPattern.Types[0]
		}

		props := make(map[string]interface{})
		if relPattern.Properties != nil {
			for k, v := range relPattern.Properties {
				props[k] = c.exprToValue(v, params)
			}
		}

		rel := graph.NewRelationship(startNode.ID, endNode.ID, relType, props)

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

		rels++
	}

	return nodes, rels, nil
}

// ExecuteMerge executes a MERGE statement and returns the number of affected elements.
// MERGE creates elements only if they don't already exist.
//
// Parameters:
//   - t: The transaction for atomic operations
//   - stmt: The MERGE statement AST node
//   - params: Query parameters for parameterized queries
//
// Returns the number of affected nodes, affected relationships, and any error encountered.
//
// Example:
//
//	nodes, rels, err := creator.ExecuteMerge(tx, mergeStmt, params)
func (c *Creator) ExecuteMerge(t *tx.Transaction, stmt *ast.MergeStmt, params map[string]interface{}) (affectedNodes, affectedRels int, err error) {
	if stmt.Pattern == nil {
		return 0, 0, nil
	}

	for _, part := range stmt.Pattern.Parts {
		if part.Path == nil || len(part.Path.Nodes) == 0 {
			continue
		}

		matcher := matchers.NewMatcherForMerge(c.Store, c.index)
		paths, err := matcher.ExecutePathForMerge(part.Path, params)
		if err != nil {
			return 0, 0, err
		}

		if len(paths) == 0 {
			nodes, rels, err := c.createPath(t, part.Path, params)
			if err != nil {
				return 0, 0, err
			}
			affectedNodes += nodes
			affectedRels += rels
		}
	}

	return affectedNodes, affectedRels, nil
}

// exprToValue converts an expression to its value.
func (c *Creator) exprToValue(expr ast.Expr, params map[string]interface{}) interface{} {
	return utils.ExprToValue(expr, params)
}
