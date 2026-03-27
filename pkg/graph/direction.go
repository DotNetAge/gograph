// Package graph provides core data structures and interfaces for the gograph database.
package graph

import (
	"fmt"
	"strings"
)

// Direction represents the direction of a relationship in a graph.
type Direction string

const (
	DirectionOutgoing Direction = "outgoing"
	DirectionIncoming Direction = "incoming"
	DirectionBoth     Direction = "both"
)

var _ fmt.Stringer = DirectionOutgoing

// String returns the string representation of the direction.
func (d Direction) String() string {
	return string(d)
}

// ParseDirection parses a direction string and returns the corresponding Direction.
// Supported formats: "outgoing", "->", "out" for outgoing;
// "incoming", "<-", "in" for incoming; "both", "-" for both.
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
