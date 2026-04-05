package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedCampaign(t *testing.T, db *DB) int64 {
	t.Helper()
	rs, err := db.GetRulesetByName("dnd5e")
	require.NoError(t, err)
	require.NotNil(t, rs)
	id, err := db.CreateCampaign(rs.ID, "Seed Campaign", "")
	require.NoError(t, err)
	return id
}

func seedSession(t *testing.T, db *DB, campaignID int64) int64 {
	t.Helper()
	id, err := db.CreateSession(campaignID, "Seed Session", "2026-04-04")
	require.NoError(t, err)
	return id
}

func TestGetTension_Default(t *testing.T) {
	db := newTestDB(t)
	campaignID := seedCampaign(t, db)
	sessionID := seedSession(t, db, campaignID)

	level, err := db.GetTension(sessionID)
	require.NoError(t, err)
	assert.Equal(t, 5, level) // default is 5
}

func TestUpdateTension_Clamp(t *testing.T) {
	db := newTestDB(t)
	campaignID := seedCampaign(t, db)
	sessionID := seedSession(t, db, campaignID)

	// Set to 8
	err := db.UpdateTension(sessionID, 8)
	require.NoError(t, err)
	level, _ := db.GetTension(sessionID)
	assert.Equal(t, 8, level)

	// Clamp to max 10
	err = db.UpdateTension(sessionID, 15)
	require.NoError(t, err)
	level, _ = db.GetTension(sessionID)
	assert.Equal(t, 10, level)

	// Clamp to min 1
	err = db.UpdateTension(sessionID, -5)
	require.NoError(t, err)
	level, _ = db.GetTension(sessionID)
	assert.Equal(t, 1, level)
}
