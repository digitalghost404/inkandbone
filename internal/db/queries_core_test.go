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

func TestCloseCampaign(t *testing.T) {
	d := newTestDB(t)
	rs, _ := d.GetRulesetByName("dnd5e")
	id, err := d.CreateCampaign(rs.ID, "Test", "")
	require.NoError(t, err)

	require.NoError(t, d.CloseCampaign(id))

	c, err := d.GetCampaign(id)
	require.NoError(t, err)
	assert.False(t, c.Active)

	assert.Error(t, d.CloseCampaign(99999))
}

func TestReopenCampaign(t *testing.T) {
	d := newTestDB(t)
	rs, _ := d.GetRulesetByName("dnd5e")
	id, err := d.CreateCampaign(rs.ID, "Test", "")
	require.NoError(t, err)
	require.NoError(t, d.CloseCampaign(id))

	c, err := d.GetCampaign(id)
	require.NoError(t, err)
	assert.False(t, c.Active)

	require.NoError(t, d.ReopenCampaign(id))

	c, err = d.GetCampaign(id)
	require.NoError(t, err)
	assert.True(t, c.Active)
}

func TestGetCampaignStats(t *testing.T) {
	d := newTestDB(t)
	rs, _ := d.GetRulesetByName("dnd5e")
	campID, err := d.CreateCampaign(rs.ID, "Test", "")
	require.NoError(t, err)

	stats, err := d.GetCampaignStats(campID)
	require.NoError(t, err)
	assert.Equal(t, CampaignStats{}, stats)

	_, err = d.CreateCharacter(campID, "Hero")
	require.NoError(t, err)
	_, err = d.CreateSession(campID, "S1", "2026-04-01")
	require.NoError(t, err)

	stats, err = d.GetCampaignStats(campID)
	require.NoError(t, err)
	assert.Equal(t, 1, stats.Sessions)
	assert.Equal(t, 1, stats.Characters)
	assert.Equal(t, 0, stats.WorldNotes)
	assert.Equal(t, 0, stats.Maps)
}

func TestCharacterCurrency(t *testing.T) {
	d, err := Open(":memory:")
	require.NoError(t, err)
	defer d.Close()

	rsID, err := d.CreateRuleset("test", "{}", "1")
	require.NoError(t, err)
	campID, err := d.CreateCampaign(rsID, "Camp", "")
	require.NoError(t, err)
	charID, err := d.CreateCharacter(campID, "Hero")
	require.NoError(t, err)

	// Defaults
	c, err := d.GetCharacter(charID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), c.CurrencyBalance)
	assert.Equal(t, "Gold", c.CurrencyLabel)

	// Update balance
	err = d.UpdateCharacterCurrencyBalance(charID, 50)
	require.NoError(t, err)
	c, err = d.GetCharacter(charID)
	require.NoError(t, err)
	assert.Equal(t, int64(50), c.CurrencyBalance)

	// Update label
	err = d.UpdateCharacterCurrencyLabel(charID, "Coin")
	require.NoError(t, err)
	c, err = d.GetCharacter(charID)
	require.NoError(t, err)
	assert.Equal(t, "Coin", c.CurrencyLabel)

	// ListCharacters includes currency
	chars, err := d.ListCharacters(campID)
	require.NoError(t, err)
	require.Len(t, chars, 1)
	assert.Equal(t, int64(50), chars[0].CurrencyBalance)
	assert.Equal(t, "Coin", chars[0].CurrencyLabel)
}

func TestDeleteCampaign(t *testing.T) {
	d := newTestDB(t)
	rs, _ := d.GetRulesetByName("dnd5e")
	campID, err := d.CreateCampaign(rs.ID, "Test", "")
	require.NoError(t, err)

	charID, err := d.CreateCharacter(campID, "Hero")
	require.NoError(t, err)
	sessID, err := d.CreateSession(campID, "S1", "2026-04-01")
	require.NoError(t, err)
	_, err = d.CreateMessage(sessID, "user", "hello", false)
	require.NoError(t, err)

	require.NoError(t, d.DeleteCampaign(campID))

	c, err := d.GetCampaign(campID)
	require.NoError(t, err)
	assert.Nil(t, c)

	ch, err := d.GetCharacter(charID)
	require.NoError(t, err)
	assert.Nil(t, ch)

	sess, err := d.GetSession(sessID)
	require.NoError(t, err)
	assert.Nil(t, sess)
}
