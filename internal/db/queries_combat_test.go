package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSession(t *testing.T, d *DB) int64 {
	t.Helper()
	campID := setupCampaign(t, d)
	sessID, err := d.CreateSession(campID, "S1", "2026-04-03")
	require.NoError(t, err)
	return sessID
}

func TestCombatEncounters(t *testing.T) {
	d := newTestDB(t)
	sessID := setupSession(t, d)

	encID, err := d.CreateEncounter(sessID, "Goblin Ambush")
	require.NoError(t, err)

	active, err := d.GetActiveEncounter(sessID)
	require.NoError(t, err)
	require.NotNil(t, active)
	assert.Equal(t, "Goblin Ambush", active.Name)

	require.NoError(t, d.EndEncounter(encID))
	active, _ = d.GetActiveEncounter(sessID)
	assert.Nil(t, active)
}

func TestCombatants(t *testing.T) {
	d := newTestDB(t)
	sessID := setupSession(t, d)
	encID, err := d.CreateEncounter(sessID, "Fight")
	require.NoError(t, err)

	cID, err := d.AddCombatant(encID, "Orc Warrior", 15, 20, false, nil)
	require.NoError(t, err)

	list, err := d.ListCombatants(encID)
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "Orc Warrior", list[0].Name)
	assert.Equal(t, 20, list[0].HPMax)
	assert.Equal(t, 20, list[0].HPCurrent)

	require.NoError(t, d.UpdateCombatant(cID, 12, `["poisoned"]`))
	list, err = d.ListCombatants(encID)
	require.NoError(t, err)
	assert.Equal(t, 12, list[0].HPCurrent)
	assert.Equal(t, `["poisoned"]`, list[0].ConditionsJSON)
}
