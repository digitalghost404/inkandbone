package mcp

import (
	"context"
	"testing"

	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddMapPin(t *testing.T) {
	s := newTestMCP(t)
	campID := setupActiveCampaign(t, s)

	mapID, err := s.db.CreateMap(campID, "World Map", "/maps/world.png")
	require.NoError(t, err)

	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"map_id": float64(mapID),
		"x":      0.42,
		"y":      0.73,
		"label":  "Ruins",
		"note":   "Ancient ruins of the old empire.",
		"color":  "#ff0000",
	}
	result, err := s.handleAddMapPin(context.Background(), req)
	require.NoError(t, err)
	require.False(t, result.IsError)

	pins, err := s.db.ListMapPins(mapID)
	require.NoError(t, err)
	require.Len(t, pins, 1)
	assert.Equal(t, "Ruins", pins[0].Label)
	assert.InDelta(t, 0.42, pins[0].X, 0.0001)
}
