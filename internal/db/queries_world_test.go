package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorldNotes(t *testing.T) {
	d := newTestDB(t)
	campID := setupCampaign(t, d)

	id, err := d.CreateWorldNote(campID, "Gareth the Guard", "A surly dwarf", "npc")
	require.NoError(t, err)
	assert.Positive(t, id)

	results, err := d.SearchWorldNotes(campID, "Gareth", "", "")
	require.NoError(t, err)
	assert.Len(t, results, 1)

	results, err = d.SearchWorldNotes(campID, "", "npc", "")
	require.NoError(t, err)
	assert.Len(t, results, 1)

	results, err = d.SearchWorldNotes(campID, "", "location", "")
	require.NoError(t, err)
	assert.Empty(t, results)

	require.NoError(t, d.UpdateWorldNote(id, "Gareth the Guard", "A surly but kind dwarf", ""))
	results, err = d.SearchWorldNotes(campID, "kind", "", "")
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestSearchWorldNotes_tagFilter(t *testing.T) {
	d := newTestDB(t)
	campID := setupCampaign(t, d)

	id1, err := d.CreateWorldNote(campID, "Goblin Den", "Dark cave", "location")
	require.NoError(t, err)
	id2, err := d.CreateWorldNote(campID, "Orc Warlord", "Fierce enemy", "npc")
	require.NoError(t, err)

	require.NoError(t, d.UpdateWorldNote(id1, "Goblin Den", "Dark cave", `["dungeon","encounter"]`))
	require.NoError(t, d.UpdateWorldNote(id2, "Orc Warlord", "Fierce enemy", `["encounter","boss"]`))

	results, err := d.SearchWorldNotes(campID, "", "", "dungeon")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, id1, results[0].ID)

	results, err = d.SearchWorldNotes(campID, "", "", "encounter")
	require.NoError(t, err)
	assert.Len(t, results, 2)

	results, err = d.SearchWorldNotes(campID, "", "", "boss")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, id2, results[0].ID)

	results, err = d.SearchWorldNotes(campID, "", "", "")
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestUpdateWorldNote_setsTagsJSON(t *testing.T) {
	d := newTestDB(t)
	campID := setupCampaign(t, d)

	id, err := d.CreateWorldNote(campID, "Mira", "A merchant.", "npc")
	require.NoError(t, err)

	require.NoError(t, d.UpdateWorldNote(id, "Mira", "A merchant.", `["npc","ally"]`))
	results, err := d.SearchWorldNotes(campID, "", "", "ally")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Contains(t, results[0].TagsJSON, "ally")
}

func TestMapsAndPins(t *testing.T) {
	d := newTestDB(t)
	campID := setupCampaign(t, d)

	mapID, err := d.CreateMap(campID, "The Keep", "/maps/keep.png")
	require.NoError(t, err)

	m, err := d.GetMap(mapID)
	require.NoError(t, err)
	assert.Equal(t, "The Keep", m.Name)

	_, err = d.AddMapPin(mapID, 0.25, 0.75, "Entrance", "Main gate", "#ff0000")
	require.NoError(t, err)

	pins, err := d.ListMapPins(mapID)
	require.NoError(t, err)
	assert.Len(t, pins, 1)
	assert.InDelta(t, 0.25, pins[0].X, 0.001)
	assert.Equal(t, "Entrance", pins[0].Label)
}

func TestDiceRolls(t *testing.T) {
	d := newTestDB(t)
	sessID := setupSession(t, d)

	_, err := d.LogDiceRoll(sessID, "1d20+5", 18, `{"rolls":[13],"modifier":5}`)
	require.NoError(t, err)

	rolls, err := d.ListDiceRolls(sessID)
	require.NoError(t, err)
	assert.Len(t, rolls, 1)
	assert.Equal(t, "1d20+5", rolls[0].Expression)
	assert.Equal(t, 18, rolls[0].Result)
}
