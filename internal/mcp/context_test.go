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

func TestGetContext_empty(t *testing.T) {
	s := newTestMCP(t)
	result, err := s.handleGetContext(context.Background(), mcplib.CallToolRequest{})
	require.NoError(t, err)
	require.False(t, result.IsError)
	tc, ok := result.Content[0].(mcplib.TextContent)
	require.True(t, ok)
	var snap map[string]any
	require.NoError(t, json.Unmarshal([]byte(tc.Text), &snap))
	assert.Nil(t, snap["campaign"])
	assert.Nil(t, snap["character"])
	assert.Nil(t, snap["session"])
	assert.Nil(t, snap["active_combat"])
}

func TestGetContext_withActiveState(t *testing.T) {
	s := newTestMCP(t)
	d := s.db

	rsID, err := d.CreateRuleset("dnd5e", `{}`, "1.0")
	require.NoError(t, err)
	campID, err := d.CreateCampaign(rsID, "Test Campaign", "")
	require.NoError(t, err)
	charID, err := d.CreateCharacter(campID, "Arin")
	require.NoError(t, err)
	sessID, err := d.CreateSession(campID, "Session 1", "2026-04-01")
	require.NoError(t, err)

	require.NoError(t, d.SetSetting("active_campaign_id", strconv.FormatInt(campID, 10)))
	require.NoError(t, d.SetSetting("active_character_id", strconv.FormatInt(charID, 10)))
	require.NoError(t, d.SetSetting("active_session_id", strconv.FormatInt(sessID, 10)))

	result, err := s.handleGetContext(context.Background(), mcplib.CallToolRequest{})
	require.NoError(t, err)
	require.False(t, result.IsError)
	tc, ok := result.Content[0].(mcplib.TextContent)
	require.True(t, ok)

	var snap map[string]any
	require.NoError(t, json.Unmarshal([]byte(tc.Text), &snap))
	camp := snap["campaign"].(map[string]any)
	assert.Equal(t, "Test Campaign", camp["Name"])
	char := snap["character"].(map[string]any)
	assert.Equal(t, "Arin", char["Name"])
	sess := snap["session"].(map[string]any)
	assert.Equal(t, "Session 1", sess["Title"])
	assert.Nil(t, snap["active_combat"])
}
