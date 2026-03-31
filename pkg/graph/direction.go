// Package graph provides core data structures and interfaces for the gograph database.
package graph

import (
	"fmt"
	"strings"
)

// Direction represents the direction of a relationship in a graph.
// It is used when traversing relationships to specify which direction
// to follow from a starting node.
//
// The three possible directions are:
//   - DirectionOutgoing: From start node to end node (->)
//   - DirectionIncoming: From end node to start node (<-)
//   - DirectionBoth: Both directions (-)
//
// Example:
//
//	// Query outgoing relationships
//	related, _ := adjacencyList.GetRelatedNodes(nodeID, "KNOWS", DirectionOutgoing)
//
//	// Query incoming relationships
//	related, _ := adjacencyList.GetRelatedNodes(nodeID, "FOLLOWS", DirectionIncoming)
//
//	// Query relationships in both directions
//	related, _ := adjacencyList.GetRelatedNodes(nodeID, "CONNECTED", DirectionBoth)
type Direction string

const (
	// DirectionOutgoing represents relationships going out from a node.
	// In Cypher syntax, this is represented as "->".
	DirectionOutgoing Direction = "outgoing"

	// DirectionIncoming represents relationships coming into a node.
	// In Cypher syntax, this is represented as "<-".
	DirectionIncoming Direction = "incoming"

	// DirectionBoth represents relationships in either direction.
	// In Cypher syntax, this is represented as "-".
	DirectionBoth Direction = "both"
)

var _ fmt.Stringer = DirectionOutgoing

// String returns the string representation of the direction.
// This implements the fmt.Stringer interface.
//
// Returns "outgoing", "incoming", or "both".
//
// Example:
//
//	fmt.Println(DirectionOutgoing) // Output: outgoing
func (d Direction) String() string {
	return string(d)
}

// ParseDirection parses a direction string and returns the corresponding Direction.
// It supports multiple formats for each direction:
//
//   - Outgoing: "outgoing", "->", "out"
//   - Incoming: "incoming", "<-", "in"
//   - Both: "both", "-"
//
// Parameters:
//   - s: The direction string to parse
//
// Returns the Direction constant and nil error on success, or DirectionOutgoing
// and an error if the string cannot be parsed.
//
// Example:
//
//	dir, err := graph.ParseDirection("->")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(dir) // Output: outgoing
func ParseDirection(s string) (Direction, error) {
	switch strings.ToLower(s) {
	case "outgoing", "->", "out":
		return DirectionOutgoing, nil
	case "incoming", "<-", "in":
		return DirectionIncoming, nil
	case "both", "-":
		return DirectionBoth, nil
	default:
		return DirectionOutgoing, fmt.Errorf("invalid direction: %s", s)
	}
}
