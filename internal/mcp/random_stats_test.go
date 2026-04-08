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

func TestRollStats_knownSystems(t *testing.T) {
	systems := []struct {
		name      string
		checkKeys []string
	}{
		{"dnd5e", []string{"str", "dex", "con", "int", "wis", "cha", "level"}},
		{"ironsworn", []string{"edge", "heart", "iron", "shadow", "wits", "health", "spirit"}},
		{"vtm", []string{"generation", "humanity", "hunger", "blood_potency", "stains"}},
		{"coc", []string{"str", "con", "pow", "sanity", "luck"}},
		{"cyberpunk_red", []string{"int", "ref", "body", "emp", "humanity"}},
		{"shadowrun", []string{"body", "agility", "essence"}},
		{"wfrp", []string{"ws", "bs", "s", "t", "wounds"}},
		{"starwars", []string{"brawn", "agility", "wounds_threshold", "credits"}},
		{"l5r", []string{"air", "earth", "fire", "water", "void", "honor"}},
		{"theonering", []string{"body", "heart", "wits", "endurance_max", "hope_max"}},
		{"wrath_glory", []string{"strength", "toughness", "wounds", "ruin"}},
		{"blades", []string{"hunt", "prowl", "skirmish", "stress", "coin"}},
		{"paranoia", []string{"violence", "moxie", "clone_number", "security_clearance"}},
	}

	for _, tc := range systems {
		t.Run(tc.name, func(t *testing.T) {
			stats := rollStats(tc.name, "")
			assert.NotEmpty(t, stats, "rollStats(%q) returned empty map", tc.name)
			for _, key := range tc.checkKeys {
				assert.Contains(t, stats, key, "missing field %q for system %q", key, tc.name)
			}
		})
	}
}

func TestRollStats_unknownSystem(t *testing.T) {
	assert.Empty(t, rollStats("pathfinder", ""))
}

func TestRollStats_dnd5eRanges(t *testing.T) {
	for i := 0; i < 20; i++ {
		stats := rollStats("dnd5e", "")
		for _, attr := range []string{"str", "dex", "con", "int", "wis", "cha"} {
			v, ok := stats[attr].(int)
			require.True(t, ok, "stat %q should be int", attr)
			assert.GreaterOrEqual(t, v, 3, "stat %q below minimum", attr)
			assert.LessOrEqual(t, v, 18, "stat %q above maximum", attr)
		}
	}
}

func TestRollStats_ironswornPointTotal(t *testing.T) {
	for i := 0; i < 20; i++ {
		stats := rollStats("ironsworn", "")
		total := 0
		for _, attr := range []string{"edge", "heart", "iron", "shadow", "wits"} {
			v, ok := stats[attr].(int)
			require.True(t, ok)
			assert.GreaterOrEqual(t, v, 1)
			assert.LessOrEqual(t, v, 3)
			total += v
		}
		assert.Equal(t, 7, total, "Ironsworn attributes must sum to 7")
	}
}

func TestRollStats_bladesPointTotal(t *testing.T) {
	actions := []string{
		"hunt", "study", "survey", "tinker",
		"finesse", "prowl", "skirmish", "wreck",
		"attune", "command", "consort", "sway",
	}
	for i := 0; i < 20; i++ {
		stats := rollStats("blades", "")
		total := 0
		for _, a := range actions {
			v, ok := stats[a].(int)
			require.True(t, ok)
			assert.GreaterOrEqual(t, v, 0)
			assert.LessOrEqual(t, v, 2)
			total += v
		}
		assert.Equal(t, 4, total, "Blades action ratings must sum to 4")
	}
}

func TestCreateCharacter_rollsStatsAutomatically(t *testing.T) {
	s := newTestMCP(t)
	campID, _, _ := setupCampaign(t, s)
	require.NoError(t, s.db.SetSetting("active_campaign_id", strconv.FormatInt(campID, 10)))

	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{"name": "Rollena"}
	result, err := s.handleCreateCharacter(context.Background(), req)
	require.NoError(t, err)
	require.False(t, result.IsError)

	// Verify stats were automatically stored on the character.
	charIDStr, _ := s.db.GetSetting("active_character_id")
	charID, _ := strconv.ParseInt(charIDStr, 10, 64)
	char, err := s.db.GetCharacter(charID)
	require.NoError(t, err)
	require.NotNil(t, char)

	var data map[string]any
	require.NoError(t, json.Unmarshal([]byte(char.DataJSON), &data))
	// setupCampaign uses dnd5e — str must be present
	assert.Contains(t, data, "str")
}
