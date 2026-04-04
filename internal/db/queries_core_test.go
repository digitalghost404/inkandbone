package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestDB(t *testing.T) *DB {
	t.Helper()
	d, err := Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { d.Close() })
	return d
}

func TestSettings(t *testing.T) {
	d := newTestDB(t)

	val, err := d.GetSetting("missing")
	require.NoError(t, err)
	assert.Empty(t, val)

	require.NoError(t, d.SetSetting("active_campaign_id", "3"))
	val, err = d.GetSetting("active_campaign_id")
	require.NoError(t, err)
	assert.Equal(t, "3", val)

	require.NoError(t, d.SetSetting("active_campaign_id", "7"))
	val, _ = d.GetSetting("active_campaign_id")
	assert.Equal(t, "7", val) // upsert
}

func TestRulesets(t *testing.T) {
	d := newTestDB(t)

	// "dnd5e" is already seeded by migration 002; look it up instead of inserting.
	r, err := d.GetRulesetByName("dnd5e")
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Positive(t, r.ID)
	assert.Equal(t, "dnd5e", r.Name)

	list, err := d.ListRulesets()
	require.NoError(t, err)
	assert.Len(t, list, 13)
}

func TestCampaigns(t *testing.T) {
	d := newTestDB(t)
	rs, err := d.GetRulesetByName("ironsworn")
	require.NoError(t, err)
	require.NotNil(t, rs)
	rsID := rs.ID

	id, err := d.CreateCampaign(rsID, "The Ironlands", "A grim world")
	require.NoError(t, err)

	c, err := d.GetCampaign(id)
	require.NoError(t, err)
	require.NotNil(t, c)
	assert.Equal(t, "The Ironlands", c.Name)
	assert.True(t, c.Active)

	list, _ := d.ListCampaigns()
	assert.Len(t, list, 1)
}

func TestCharacters(t *testing.T) {
	d := newTestDB(t)
	rs, err := d.GetRulesetByName("ironsworn")
	require.NoError(t, err)
	require.NotNil(t, rs)
	campID, err := d.CreateCampaign(rs.ID, "Test Campaign", "")
	require.NoError(t, err)

	charID, err := d.CreateCharacter(campID, "Kael")
	require.NoError(t, err)

	ch, err := d.GetCharacter(charID)
	require.NoError(t, err)
	assert.Equal(t, "Kael", ch.Name)
	assert.Equal(t, "{}", ch.DataJSON)

	require.NoError(t, d.UpdateCharacterData(charID, `{"hp":20}`))
	ch, _ = d.GetCharacter(charID)
	assert.Equal(t, `{"hp":20}`, ch.DataJSON)

	list, _ := d.ListCharacters(campID)
	assert.Len(t, list, 1)
}
