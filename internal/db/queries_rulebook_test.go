package db

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetRuleset(t *testing.T) {
	d := newTestDB(t)
	rulesets, err := d.ListRulesets()
	require.NoError(t, err)
	require.NotEmpty(t, rulesets)

	r, err := d.GetRuleset(rulesets[0].ID)
	require.NoError(t, err)
	require.Equal(t, rulesets[0].ID, r.ID)
	require.Equal(t, rulesets[0].Name, r.Name)
}

func TestCreateAndSearchRulebookChunks(t *testing.T) {
	d := newTestDB(t)
	rulesets, _ := d.ListRulesets()
	rulesetID := rulesets[0].ID

	chunks := []RulebookChunk{
		{Heading: "Stealth Rules", Content: "Characters can hide in shadows."},
		{Heading: "Combat", Content: "Initiative is rolled at the start of combat."},
	}
	err := d.CreateRulebookChunks(rulesetID, chunks)
	require.NoError(t, err)

	results, err := d.SearchRulebookChunks(rulesetID, "stealth")
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "Stealth Rules", results[0].Heading)
}

func TestSearchRulebookChunks_empty(t *testing.T) {
	d := newTestDB(t)
	rulesets, _ := d.ListRulesets()
	results, err := d.SearchRulebookChunks(rulesets[0].ID, "nonexistent")
	require.NoError(t, err)
	require.Empty(t, results)
}
