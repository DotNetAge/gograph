package creators

import (
	"github.com/DotNetAge/gograph/pkg/cypher/ast"
	"github.com/DotNetAge/gograph/pkg/cypher/matchers"
	"github.com/DotNetAge/gograph/pkg/cypher/utils"
	"github.com/DotNetAge/gograph/pkg/graph"
	"github.com/DotNetAge/gograph/pkg/storage"
	"github.com/DotNetAge/gograph/pkg/tx"
)

type Creator struct {
	Store *storage.DB
	index *graph.Index
	adj   *graph.AdjacencyList
}

func NewCreator(store *storage.DB) *Creator {
	return &Creator{
		Store: store,
		index: graph.NewIndex(store),
		adj:   graph.NewAdjacencyList(store),
	}
}

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

func (c *Creator) exprToValue(expr ast.Expr, params map[string]interface{}) interface{} {
	return utils.ExprToValue(expr, params)
}
