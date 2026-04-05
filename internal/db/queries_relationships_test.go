package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelationshipCRUD(t *testing.T) {
	db := newTestDB(t)
	campaignID := seedCampaign(t, db)

	// Create
	id, err := db.CreateRelationship(campaignID, "Elara", "Doran", "ally", "Old friends from the war")
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// List
	rels, err := db.ListRelationships(campaignID)
	require.NoError(t, err)
	require.Len(t, rels, 1)
	assert.Equal(t, "Elara", rels[0].FromName)
	assert.Equal(t, "Doran", rels[0].ToName)
	assert.Equal(t, "ally", rels[0].RelationshipType)

	// Update
	err = db.UpdateRelationship(id, "rival", "They fell out over the treasure")
	require.NoError(t, err)

	rels, _ = db.ListRelationships(campaignID)
	assert.Equal(t, "rival", rels[0].RelationshipType)
	assert.Equal(t, "They fell out over the treasure", rels[0].Description)

	// Delete
	err = db.DeleteRelationship(id)
	require.NoError(t, err)

	rels, _ = db.ListRelationships(campaignID)
	assert.Len(t, rels, 0)
}
