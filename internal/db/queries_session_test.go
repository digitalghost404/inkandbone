package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupCampaign(t *testing.T, d *DB) int64 {
	t.Helper()
	rsID, err := d.CreateRuleset("dnd5e", `{}`, "1.0")
	require.NoError(t, err)
	campID, err := d.CreateCampaign(rsID, "Test", "")
	require.NoError(t, err)
	return campID
}

func TestSessions(t *testing.T) {
	d := newTestDB(t)
	campID := setupCampaign(t, d)

	sessID, err := d.CreateSession(campID, "Session 1", "2026-04-03")
	require.NoError(t, err)

	s, err := d.GetSession(sessID)
	require.NoError(t, err)
	assert.Equal(t, "Session 1", s.Title)
	assert.Empty(t, s.Summary)

	require.NoError(t, d.UpdateSessionSummary(sessID, "Found the dungeon"))
	s, err = d.GetSession(sessID)
	require.NoError(t, err)
	assert.Equal(t, "Found the dungeon", s.Summary)

	list, err := d.ListSessions(campID)
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestMessages(t *testing.T) {
	d := newTestDB(t)
	campID := setupCampaign(t, d)
	sessID, err := d.CreateSession(campID, "S1", "2026-04-03")
	require.NoError(t, err)

	_, err = d.CreateMessage(sessID, "assistant", "You enter the dungeon")
	require.NoError(t, err)
	_, err = d.CreateMessage(sessID, "user", "I draw my sword")
	require.NoError(t, err)

	// ORDER BY created_at, id — id tiebreaker ensures determinism when timestamps share the same second
	msgs, err := d.ListMessages(sessID)
	require.NoError(t, err)
	assert.Len(t, msgs, 2)
	assert.Equal(t, "assistant", msgs[0].Role)
	assert.Equal(t, "user", msgs[1].Role)

	recent, err := d.RecentMessages(sessID, 1)
	require.NoError(t, err)
	assert.Len(t, recent, 1)
	assert.Equal(t, "user", recent[0].Role)
}
