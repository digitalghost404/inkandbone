package mcp

import (
	"context"
	"strconv"
	"testing"

	"github.com/digitalghost404/inkandbone/internal/api"
	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mcpStubCompleter struct{ response string }

func (s *mcpStubCompleter) Generate(_ context.Context, _ string) (string, error) {
	return s.response, nil
}

func TestGenerateSessionRecap(t *testing.T) {
	stub := &mcpStubCompleter{response: "The heroes fought valiantly."}
	s := newTestMCPWithAI(t, stub)

	// Seed campaign and session
	rs, err := s.db.GetRulesetByName("dnd5e")
	require.NoError(t, err)
	require.NotNil(t, rs)
	campID, err := s.db.CreateCampaign(rs.ID, "Test Campaign", "")
	require.NoError(t, err)
	sessID, err := s.db.CreateSession(campID, "S1", "2026-04-03")
	require.NoError(t, err)
	require.NoError(t, s.db.SetSetting("active_session_id", strconv.FormatInt(sessID, 10)))

	// Collect WS events
	events := []api.Event{}
	ch := s.bus.Subscribe()

	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{"session_id": float64(sessID)}
	result, err := s.handleGenerateSessionRecap(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)

	// Drain one event
	select {
	case e := <-ch:
		events = append(events, e)
	default:
	}
	require.Len(t, events, 1)
	assert.Equal(t, api.EventSessionUpdated, events[0].Type)
	payload := events[0].Payload.(map[string]any)
	assert.Equal(t, sessID, payload["session_id"])
	assert.Equal(t, "The heroes fought valiantly.", payload["summary"])

	// Verify DB updated
	sess, err := s.db.GetSession(sessID)
	require.NoError(t, err)
	assert.Equal(t, "The heroes fought valiantly.", sess.Summary)
}

func TestGenerateSessionRecap_noAI(t *testing.T) {
	s := newTestMCP(t) // aiClient is nil
	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{"session_id": float64(1)}
	result, err := s.handleGenerateSessionRecap(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}
