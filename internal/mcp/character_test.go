package mcp

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"

	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupActiveCharacter(t *testing.T, s *Server) int64 {
	t.Helper()
	rsID, err := s.db.CreateRuleset("dnd5e", `{}`, "1.0")
	require.NoError(t, err)
	campID, err := s.db.CreateCampaign(rsID, "Camp", "")
	require.NoError(t, err)
	charID, err := s.db.CreateCharacter(campID, "Lyra")
	require.NoError(t, err)
	require.NoError(t, s.db.SetSetting("active_character_id", strconv.FormatInt(charID, 10)))
	require.NoError(t, s.db.SetSetting("active_campaign_id", strconv.FormatInt(campID, 10)))
	return charID
}

func TestGetCharacterSheet(t *testing.T) {
	s := newTestMCP(t)
	charID := setupActiveCharacter(t, s)

	result, err := s.handleGetCharacterSheet(context.Background(), mcplib.CallToolRequest{})
	require.NoError(t, err)
	require.False(t, result.IsError)
	tc, ok := result.Content[0].(mcplib.TextContent)
	require.True(t, ok)
	var char map[string]any
	require.NoError(t, json.Unmarshal([]byte(tc.Text), &char))
	assert.Equal(t, float64(charID), char["ID"])
	assert.Equal(t, "Lyra", char["Name"])
}

func TestUpdateCharacter(t *testing.T) {
	s := newTestMCP(t)
	charID := setupActiveCharacter(t, s)

	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"updates": `{"hp":18,"level":2}`,
	}
	result, err := s.handleUpdateCharacter(context.Background(), req)
	require.NoError(t, err)
	require.False(t, result.IsError)

	char, err := s.db.GetCharacter(charID)
	require.NoError(t, err)
	var data map[string]any
	require.NoError(t, json.Unmarshal([]byte(char.DataJSON), &data))
	assert.Equal(t, float64(18), data["hp"])
	assert.Equal(t, float64(2), data["level"])
}

func TestAddAndRemoveItem(t *testing.T) {
	s := newTestMCP(t)
	charID := setupActiveCharacter(t, s)

	addReq := mcplib.CallToolRequest{}
	addReq.Params.Arguments = map[string]any{"item_name": "Longsword"}
	result, err := s.handleAddItem(context.Background(), addReq)
	require.NoError(t, err)
	require.False(t, result.IsError)

	char, err := s.db.GetCharacter(charID)
	require.NoError(t, err)
	var data map[string]any
	require.NoError(t, json.Unmarshal([]byte(char.DataJSON), &data))
	inv := data["inventory"].([]any)
	assert.Contains(t, inv, "Longsword")

	removeReq := mcplib.CallToolRequest{}
	removeReq.Params.Arguments = map[string]any{"item_name": "Longsword"}
	result, err = s.handleRemoveItem(context.Background(), removeReq)
	require.NoError(t, err)
	require.False(t, result.IsError)

	char, err = s.db.GetCharacter(charID)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal([]byte(char.DataJSON), &data))
	inv = data["inventory"].([]any)
	assert.NotContains(t, inv, "Longsword")
}
