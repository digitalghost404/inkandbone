package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRollOracle_Generic(t *testing.T) {
	db := newTestDB(t)

	// Roll 1 against 'action' table (generic) — should return 'Attack' (roll_min=1, roll_max=2)
	result, err := db.RollOracle(nil, "action", 1)
	require.NoError(t, err)
	assert.Equal(t, "Attack", result)
}

func TestRollOracle_Mid(t *testing.T) {
	db := newTestDB(t)

	// Roll 25 against 'action' — should return 'Create' (roll_min=25, roll_max=26)
	result, err := db.RollOracle(nil, "action", 25)
	require.NoError(t, err)
	assert.Equal(t, "Create", result)
}

func TestRollOracle_NotFound(t *testing.T) {
	db := newTestDB(t)

	// Roll 99 against 'action' (max is 50) — should return empty string, no error
	result, err := db.RollOracle(nil, "action", 99)
	require.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestRollOracle_RulesetFallback(t *testing.T) {
	db := newTestDB(t)

	// No ruleset-specific rows exist, so fallback to generic
	rulesetID := int64(999) // non-existent ruleset
	result, err := db.RollOracle(&rulesetID, "action", 1)
	require.NoError(t, err)
	assert.Equal(t, "Attack", result)
}
