package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/digitalghost404/inkandbone/internal/db"
	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchRulebook(t *testing.T) {
	s := newTestMCP(t)
	// Seed: campaign → ruleset → chunks
	rs, err := s.db.GetRulesetByName("dnd5e")
	require.NoError(t, err)
	require.NotNil(t, rs)
	campID, err := s.db.CreateCampaign(rs.ID, "Test Campaign", "")
	require.NoError(t, err)
	camp, err := s.db.GetCampaign(campID)
	require.NoError(t, err)

	// Import a couple chunks into the ruleset
	chunks := []db.RulebookChunk{
		{Heading: "Spellcasting", Content: "A wizard can cast spells using slots."},
		{Heading: "Combat", Content: "On your turn you may move and take one action."},
	}
	err = s.db.CreateRulebookChunks(camp.RulesetID, chunks)
	require.NoError(t, err)

	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"query":      "spell",
		"ruleset_id": float64(camp.RulesetID),
	}
	result, err := s.handleSearchRulebook(context.Background(), req)
	require.NoError(t, err)
	require.False(t, result.IsError)

	var got []map[string]any
	text := result.Content[0].(mcplib.TextContent).Text
	require.NoError(t, json.Unmarshal([]byte(text), &got))
	require.Len(t, got, 1)
	assert.Equal(t, "Spellcasting", got[0]["heading"])
}

func TestSearchRulebook_noQuery(t *testing.T) {
	s := newTestMCP(t)
	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{}
	result, err := s.handleSearchRulebook(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}
