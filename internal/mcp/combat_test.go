package mcp

import (
	"context"
	"strconv"
	"testing"

	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupActiveSession(t *testing.T, s *Server) int64 {
	t.Helper()
	rsID, err := s.db.CreateRuleset("dnd5e", `{}`, "1.0")
	require.NoError(t, err)
	campID, err := s.db.CreateCampaign(rsID, "Camp", "")
	require.NoError(t, err)
	sessID, err := s.db.CreateSession(campID, "S1", "2026-04-01")
	require.NoError(t, err)
	require.NoError(t, s.db.SetSetting("active_session_id", strconv.FormatInt(sessID, 10)))
	return sessID
}

func TestStartCombat(t *testing.T) {
	s := newTestMCP(t)
	sessID := setupActiveSession(t, s)

	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"name":       "Goblin Ambush",
		"combatants": `[{"name":"Goblin","initiative":14,"hp_max":7,"is_player":false},{"name":"Hero","initiative":18,"hp_max":20,"is_player":true}]`,
	}
	result, err := s.handleStartCombat(context.Background(), req)
	require.NoError(t, err)
	require.False(t, result.IsError)

	enc, err := s.db.GetActiveEncounter(sessID)
	require.NoError(t, err)
	require.NotNil(t, enc)
	assert.Equal(t, "Goblin Ambush", enc.Name)

	combatants, err := s.db.ListCombatants(enc.ID)
	require.NoError(t, err)
	assert.Len(t, combatants, 2)
	// ListCombatants returns ORDER BY initiative DESC; Hero(18) first
	assert.Equal(t, "Hero", combatants[0].Name)
	assert.Equal(t, "Goblin", combatants[1].Name)
}

func TestUpdateCombatant(t *testing.T) {
	s := newTestMCP(t)
	sessID := setupActiveSession(t, s)

	encID, err := s.db.CreateEncounter(sessID, "Fight")
	require.NoError(t, err)
	combID, err := s.db.AddCombatant(encID, "Orc", 12, 15, false, nil)
	require.NoError(t, err)

	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"combatant_id": float64(combID),
		"hp_current":   float64(8),
		"conditions":   `["poisoned"]`,
	}
	result, err := s.handleUpdateCombatant(context.Background(), req)
	require.NoError(t, err)
	require.False(t, result.IsError)

	combatants, err := s.db.ListCombatants(encID)
	require.NoError(t, err)
	require.Len(t, combatants, 1)
	assert.Equal(t, 8, combatants[0].HPCurrent)
	assert.Equal(t, `["poisoned"]`, combatants[0].ConditionsJSON)
}

func TestUpdateCombatant_missingHP(t *testing.T) {
	s := newTestMCP(t)
	setupActiveSession(t, s)

	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"combatant_id": float64(1),
		"conditions":   `["prone"]`,
		// hp_current intentionally omitted
	}
	result, err := s.handleUpdateCombatant(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestEndCombat(t *testing.T) {
	s := newTestMCP(t)
	sessID := setupActiveSession(t, s)

	_, err := s.db.CreateEncounter(sessID, "Fight")
	require.NoError(t, err)

	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{}
	result, err := s.handleEndCombat(context.Background(), req)
	require.NoError(t, err)
	require.False(t, result.IsError)

	enc, err := s.db.GetActiveEncounter(sessID)
	require.NoError(t, err)
	assert.Nil(t, enc)
}
