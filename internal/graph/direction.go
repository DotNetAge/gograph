package graph

import (
	"fmt"
	"strings"
)

type Direction string

const (
	DirectionOutgoing Direction = "outgoing"
	DirectionIncoming Direction = "incoming"
	DirectionBoth     Direction = "both"
)

var _ fmt.Stringer = DirectionOutgoing

func (d Direction) String() string {
	return string(d)
}

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
