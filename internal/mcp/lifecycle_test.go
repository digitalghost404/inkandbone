package mcp

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/digitalghost404/inkandbone/internal/db"
	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCampaign(t *testing.T) {
	s := newTestMCP(t)
	// "dnd5e" is seeded by migration 002
	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"ruleset": "dnd5e",
		"name":    "Dragon Campaign",
	}
	result, err := s.handleCreateCampaign(context.Background(), req)
	require.NoError(t, err)
	require.False(t, result.IsError)

	got, _ := s.db.GetSetting("active_campaign_id")
	assert.NotEmpty(t, got)
}

func TestCreateCampaign_unknownRuleset(t *testing.T) {
	s := newTestMCP(t)
	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{"ruleset": "pathfinder", "name": "Test"}
	result, err := s.handleCreateCampaign(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestListCampaigns_empty(t *testing.T) {
	s := newTestMCP(t)
	result, err := s.handleListCampaigns(context.Background(), mcplib.CallToolRequest{})
	require.NoError(t, err)
	require.False(t, result.IsError)

	tc, ok := result.Content[0].(mcplib.TextContent)
	require.True(t, ok)
	var campaigns []db.Campaign
	require.NoError(t, json.Unmarshal([]byte(tc.Text), &campaigns))
	assert.Empty(t, campaigns)
}

func TestCreateCharacter(t *testing.T) {
	s := newTestMCP(t)
	campID, _, _ := setupCampaign(t, s)
	require.NoError(t, s.db.SetSetting("active_campaign_id", strconv.FormatInt(campID, 10)))

	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{"name": "Talia"}
	result, err := s.handleCreateCharacter(context.Background(), req)
	require.NoError(t, err)
	require.False(t, result.IsError)

	got, _ := s.db.GetSetting("active_character_id")
	assert.NotEmpty(t, got)
}

func TestCreateCharacter_noCampaign(t *testing.T) {
	s := newTestMCP(t)
	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{"name": "Talia"}
	result, err := s.handleCreateCharacter(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestListCharacters(t *testing.T) {
	s := newTestMCP(t)
	campID, charID, _ := setupCampaign(t, s)
	require.NoError(t, s.db.SetSetting("active_campaign_id", strconv.FormatInt(campID, 10)))

	result, err := s.handleListCharacters(context.Background(), mcplib.CallToolRequest{})
	require.NoError(t, err)
	require.False(t, result.IsError)

	tc, ok := result.Content[0].(mcplib.TextContent)
	require.True(t, ok)
	var chars []db.Character
	require.NoError(t, json.Unmarshal([]byte(tc.Text), &chars))
	require.Len(t, chars, 1)
	assert.Equal(t, charID, chars[0].ID)
}

func TestListSessions(t *testing.T) {
	s := newTestMCP(t)
	campID, _, sessID := setupCampaign(t, s)
	require.NoError(t, s.db.SetSetting("active_campaign_id", strconv.FormatInt(campID, 10)))

	result, err := s.handleListSessions(context.Background(), mcplib.CallToolRequest{})
	require.NoError(t, err)
	require.False(t, result.IsError)

	tc, ok := result.Content[0].(mcplib.TextContent)
	require.True(t, ok)
	var sessions []db.Session
	require.NoError(t, json.Unmarshal([]byte(tc.Text), &sessions))
	require.Len(t, sessions, 1)
	assert.Equal(t, sessID, sessions[0].ID)
}
