package ast

import (
	"testing"
)

func TestRelDirectionString(t *testing.T) {
	tests := []struct {
		direction RelDirection
		expected  string
	}{
		{RelDirOutgoing, "->"},
		{RelDirIncoming, "<-"},
		{RelDirBoth, "-"},
	}

	for _, tc := range tests {
		result := tc.direction.String()
		if result != tc.expected {
			t.Errorf("expected %s, got %s", tc.expected, result)
		}
	}
}
