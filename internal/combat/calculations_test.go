package combat

import (
	"fmt"
	"math"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/characters"
)

// The unit tests
func TestAlignmentChange(t *testing.T) {
	tests := []struct {
		killerAlignment int8
		killedAlignment int8
		expectedChange  int
	}{

		{0, 0, 0},
		{0, 5, 0},
		{5, 0, 0},

		{0, 15, 0},
		{0, -15, 0},

		{15, -15, 0},
		{-15, 15, 0},

		{15, 25, 0},
		{25, 15, 0},

		{-20, -25, 2},
		{-25, -20, 2},

		{50, -10, -1},
		{-50, 10, -1},

		{50, -50, 2},
		{-50, 50, -2},

		{90, 0, -2},
		{-90, 0, -2},

		{100, 20, -4},
		{-100, -20, 4},

		{90, -90, 4},
		{-90, 90, -4},
	}

	for _, test := range tests {
		desc := fmt.Sprintf(`%s kills %s`, characters.AlignmentToString(test.killerAlignment), characters.AlignmentToString(test.killedAlignment))
		delta := int(math.Abs(math.Max(float64(test.killerAlignment), float64(test.killedAlignment))-math.Min(float64(test.killerAlignment), float64(test.killedAlignment))) * 0.5)
		result := AlignmentChange(test.killerAlignment, test.killedAlignment)
		if result != test.expectedChange {
			t.Errorf("%s [Delta: %d]: AlignmentChange(%d, %d) = %d; want %d",
				desc, delta, test.killerAlignment, test.killedAlignment, result, test.expectedChange)
		}
	}
}
