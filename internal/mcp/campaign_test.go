package mcp

import (
	"context"
	"strconv"
	"testing"

	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupCampaign(t *testing.T, s *Server) (campID, charID, sessID int64) {
	t.Helper()
	rsID, err := s.db.CreateRuleset("dnd5e", `{}`, "1.0")
	require.NoError(t, err)
	campID, err = s.db.CreateCampaign(rsID, "Campaign", "")
	require.NoError(t, err)
	charID, err = s.db.CreateCharacter(campID, "Hero")
	require.NoError(t, err)
	sessID, err = s.db.CreateSession(campID, "S1", "2026-04-01")
	require.NoError(t, err)
	return
}

func TestSetActive(t *testing.T) {
	s := newTestMCP(t)
	campID, charID, sessID := setupCampaign(t, s)

	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"campaign_id":  float64(campID),
		"character_id": float64(charID),
		"session_id":   float64(sessID),
	}
	result, err := s.handleSetActive(context.Background(), req)
	require.NoError(t, err)
	require.False(t, result.IsError)

	got, _ := s.db.GetSetting("active_campaign_id")
	assert.Equal(t, strconv.FormatInt(campID, 10), got)
	got, _ = s.db.GetSetting("active_character_id")
	assert.Equal(t, strconv.FormatInt(charID, 10), got)
	got, _ = s.db.GetSetting("active_session_id")
	assert.Equal(t, strconv.FormatInt(sessID, 10), got)
}

func TestStartSession(t *testing.T) {
	s := newTestMCP(t)
	campID, _, _ := setupCampaign(t, s)
	require.NoError(t, s.db.SetSetting("active_campaign_id", strconv.FormatInt(campID, 10)))

	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"title": "New Session",
		"date":  "2026-04-03",
	}
	result, err := s.handleStartSession(context.Background(), req)
	require.NoError(t, err)
	require.False(t, result.IsError)

	idStr, _ := s.db.GetSetting("active_session_id")
	assert.NotEmpty(t, idStr)
	id, _ := strconv.ParseInt(idStr, 10, 64)
	sess, err := s.db.GetSession(id)
	require.NoError(t, err)
	require.NotNil(t, sess)
	assert.Equal(t, "New Session", sess.Title)
}

func TestStartSession_noCampaign(t *testing.T) {
	s := newTestMCP(t)
	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{"title": "X", "date": "2026-04-03"}
	result, err := s.handleStartSession(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestEndSession(t *testing.T) {
	s := newTestMCP(t)
	campID, _, sessID := setupCampaign(t, s)
	require.NoError(t, s.db.SetSetting("active_campaign_id", strconv.FormatInt(campID, 10)))
	require.NoError(t, s.db.SetSetting("active_session_id", strconv.FormatInt(sessID, 10)))

	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{"summary": "A great session."}
	result, err := s.handleEndSession(context.Background(), req)
	require.NoError(t, err)
	require.False(t, result.IsError)

	sess, err := s.db.GetSession(sessID)
	require.NoError(t, err)
	assert.Equal(t, "A great session.", sess.Summary)
}
