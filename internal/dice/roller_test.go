package dice

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoll_validExpressions(t *testing.T) {
	cases := []struct {
		expr      string
		diceCount int
		sides     int
		mod       int
	}{
		{"d6", 1, 6, 0},
		{"1d6", 1, 6, 0},
		{"2d6", 2, 6, 0},
		{"2d6+3", 2, 6, 3},
		{"1d8-1", 1, 8, -1},
		{"d20", 1, 20, 0},
		{"4d6+0", 4, 6, 0},
	}
	for _, tc := range cases {
		t.Run(tc.expr, func(t *testing.T) {
			total, breakdown, err := Roll(tc.expr)
			require.NoError(t, err)
			assert.Len(t, breakdown, tc.diceCount)
			for _, d := range breakdown {
				assert.GreaterOrEqual(t, d, 1)
				assert.LessOrEqual(t, d, tc.sides)
			}
			sum := tc.mod
			for _, d := range breakdown {
				sum += d
			}
			assert.Equal(t, sum, total)
		})
	}
}

func TestRoll_invalidExpressions(t *testing.T) {
	cases := []string{"", "abc", "2x6", "0d6", "2d0", "d"}
	for _, expr := range cases {
		t.Run(expr, func(t *testing.T) {
			_, _, err := Roll(expr)
			assert.Error(t, err)
		})
	}
}
