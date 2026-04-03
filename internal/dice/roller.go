package dice

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
)

// diceRe matches: optional N, literal 'd', Y sides, optional +/-modifier
var diceRe = regexp.MustCompile(`^(\d*)d(\d+)([+-]\d+)?$`)

// Roll parses and evaluates a dice expression (e.g. "2d6+3", "d20", "1d8-1").
// Returns the total, the per-die results (modifier excluded), and any parse error.
func Roll(expr string) (total int, breakdown []int, err error) {
	expr = strings.TrimSpace(strings.ToLower(expr))
	m := diceRe.FindStringSubmatch(expr)
	if m == nil {
		return 0, nil, fmt.Errorf("invalid dice expression %q: use NdY, NdY+M, or NdY-M", expr)
	}

	count := 1
	if m[1] != "" {
		count, err = strconv.Atoi(m[1])
		if err != nil || count < 1 {
			return 0, nil, fmt.Errorf("invalid die count in %q", expr)
		}
	}

	sides, err := strconv.Atoi(m[2])
	if err != nil || sides < 1 {
		return 0, nil, fmt.Errorf("invalid die sides in %q", expr)
	}

	mod := 0
	if m[3] != "" {
		mod, err = strconv.Atoi(m[3]) // includes the sign character
		if err != nil {
			return 0, nil, fmt.Errorf("invalid modifier in %q", expr)
		}
	}

	breakdown = make([]int, count)
	sum := 0
	for i := range breakdown {
		breakdown[i] = rand.Intn(sides) + 1
		sum += breakdown[i]
	}
	return sum + mod, breakdown, nil
}
