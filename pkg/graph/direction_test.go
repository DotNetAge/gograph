package graph

import (
	"testing"
)

func TestDirectionString(t *testing.T) {
	tests := []struct {
		direction Direction
		expected  string
	}{
		{DirectionOutgoing, "outgoing"},
		{DirectionIncoming, "incoming"},
		{DirectionBoth, "both"},
	}

	for _, tc := range tests {
		result := tc.direction.String()
		if result != tc.expected {
			t.Errorf("expected %s, got %s", tc.expected, result)
		}
	}
}

func TestParseDirection(t *testing.T) {
	tests := []struct {
		input     string
		expected  Direction
		expectErr bool
	}{
		// Outgoing directions
		{"outgoing", DirectionOutgoing, false},
		{"->", DirectionOutgoing, false},
		{"out", DirectionOutgoing, false},
		{"OUTGOING", DirectionOutgoing, false},
		{"OUT", DirectionOutgoing, false},
		// Incoming directions
		{"incoming", DirectionIncoming, false},
		{"<-", DirectionIncoming, false},
		{"in", DirectionIncoming, false},
		{"INCOMING", DirectionIncoming, false},
		{"IN", DirectionIncoming, false},
		// Both directions
		{"both", DirectionBoth, false},
		{"-", DirectionBoth, false},
		{"BOTH", DirectionBoth, false},
		// Invalid direction
		{"invalid", DirectionOutgoing, true},
		{"", DirectionOutgoing, true},
	}

	for _, tc := range tests {
		result, err := ParseDirection(tc.input)
		if (err != nil) != tc.expectErr {
			t.Errorf("expected error: %v, got: %v", tc.expectErr, err)
		}
		if result != tc.expected {
			t.Errorf("expected %s, got %s", tc.expected, result)
		}
	}
}
