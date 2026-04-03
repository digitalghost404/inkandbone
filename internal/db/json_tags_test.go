package db

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCampaignJSONKeys(t *testing.T) {
	c := Campaign{ID: 1, RulesetID: 2, Name: "Greyhawk", Description: "desc", Active: true, CreatedAt: "2026-04-03"}
	b, err := json.Marshal(c)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(b, &m))
	assert.Contains(t, m, "id")
	assert.Contains(t, m, "ruleset_id")
	assert.Contains(t, m, "name")
	assert.Contains(t, m, "description")
	assert.Contains(t, m, "active")
	assert.Contains(t, m, "created_at")
	assert.NotContains(t, m, "ID")
	assert.NotContains(t, m, "Name")
}

func TestMessageJSONKeys(t *testing.T) {
	msg := Message{ID: 1, SessionID: 2, Role: "user", Content: "hello", CreatedAt: "2026-04-03"}
	b, err := json.Marshal(msg)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(b, &m))
	assert.Contains(t, m, "id")
	assert.Contains(t, m, "session_id")
	assert.Contains(t, m, "role")
	assert.Contains(t, m, "content")
	assert.Contains(t, m, "created_at")
	assert.NotContains(t, m, "SessionID")
}

func TestCombatantJSONKeys(t *testing.T) {
	c := Combatant{ID: 1, EncounterID: 2, Name: "Goblin", Initiative: 12, HPCurrent: 7, HPMax: 7, ConditionsJSON: "[]", IsPlayer: false}
	b, err := json.Marshal(c)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(b, &m))
	assert.Contains(t, m, "id")
	assert.Contains(t, m, "encounter_id")
	assert.Contains(t, m, "initiative")
	assert.Contains(t, m, "hp_current")
	assert.Contains(t, m, "hp_max")
	assert.Contains(t, m, "is_player")
	assert.NotContains(t, m, "HPCurrent")
}
