package matchers

import (
	"github.com/DotNetAge/gograph/pkg/cypher/ast"
	"github.com/DotNetAge/gograph/pkg/graph"
	"github.com/DotNetAge/gograph/pkg/storage"
)

type pathState struct {
	currentNode *graph.Node
	path       []interface{}
	hopCount   int
}

func (m *Matcher) findVariableLengthPaths(
	startNode *graph.Node,
	elem *ast.PatternElement,
	visited map[string]bool,
) []map[string]interface{} {
	var results []map[string]interface{}

	if elem.Relation.MinHops == 0 {
		path := make(map[string]interface{})
		if elem.Node.Variable != "" {
			path[elem.Node.Variable] = startNode
		}
		results = append(results, path)
	}

	minHops := 1
	maxHops := 1
	if elem.Relation.MinHops > 0 {
		minHops = elem.Relation.MinHops
	}
	if elem.Relation.MaxHops > 0 {
		maxHops = elem.Relation.MaxHops
	}

	initialVisited := make(map[string]bool)
	initialVisited[startNode.ID] = true

	queue := []pathState{
		{
			currentNode: startNode,
			path:       []interface{}{},
			hopCount:   0,
		},
	}

	for len(queue) > 0 {
		state := queue[0]
		queue = queue[1:]

		if state.hopCount >= maxHops {
			continue
		}

		adj := graph.NewAdjacencyList(m.Store)
		relIDs, _ := adj.GetAllRelated(state.currentNode.ID)

		for _, relID := range relIDs {
			relData, err := m.Store.Get(storage.RelKey(relID))
			if err != nil {
				continue
			}
			var rel graph.Relationship
			if err := storage.Unmarshal(relData, &rel); err != nil {
				continue
			}

			if elem.Relation.RelType != "" && rel.Type != elem.Relation.RelType {
				continue
			}

			var nextNodeID string
			switch elem.Relation.Dir {
			case ast.RelDirOutgoing, "":
				if rel.StartNodeID != state.currentNode.ID {
					continue
				}
				nextNodeID = rel.EndNodeID
			case ast.RelDirIncoming:
				if rel.EndNodeID != state.currentNode.ID {
					continue
				}
				nextNodeID = rel.StartNodeID
			case ast.RelDirBoth:
				if rel.StartNodeID == state.currentNode.ID {
					nextNodeID = rel.EndNodeID
				} else if rel.EndNodeID == state.currentNode.ID {
					nextNodeID = rel.StartNodeID
				} else {
					continue
				}
			default:
				continue
			}

			if initialVisited[nextNodeID] {
				continue
			}

			nextNodeData, err := m.Store.Get(storage.NodeKey(nextNodeID))
			if err != nil {
				continue
			}
			var nextNode graph.Node
			if err := storage.Unmarshal(nextNodeData, &nextNode); err != nil {
				continue
			}

			if elem.Relation.EndNode != nil {
				if !m.NodeMatchesProperties(&nextNode, elem.Relation.EndNode.Labels, elem.Relation.EndNode.Properties) {
					continue
				}
			}

			newPath := make([]interface{}, len(state.path))
			copy(newPath, state.path)
			if elem.Relation.Variable != "" {
				newPath = append(newPath, &rel)
			}

			newState := pathState{
				currentNode: &nextNode,
				path:       newPath,
				hopCount:   state.hopCount + 1,
			}

			if newState.hopCount >= minHops {
				result := make(map[string]interface{})
				if elem.Node.Variable != "" {
					result[elem.Node.Variable] = startNode
				}
				if elem.Relation.Variable != "" && len(newPath) > 0 {
					result[elem.Relation.Variable] = newPath
				}
				if elem.Relation.EndNode != nil && elem.Relation.EndNode.Variable != "" {
					result[elem.Relation.EndNode.Variable] = &nextNode
				}
				results = append(results, result)
			}

			newVisited := make(map[string]bool)
			for k, v := range initialVisited {
				newVisited[k] = v
			}
			newVisited[nextNodeID] = true

			queue = append(queue, newState)
		}
	}

	return results
}
