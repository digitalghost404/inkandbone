package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/digitalghost404/inkandbone/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedCampaign creates a ruleset, campaign, and session for use in route tests.
// It uses t.Name() as the ruleset name to avoid conflicts with pre-seeded rulesets.
func seedCampaign(t *testing.T, d *db.DB) (campID, sessID int64) {
	t.Helper()
	rsID, err := d.CreateRuleset(t.Name(), `{}`, "test")
	require.NoError(t, err)
	campID, err = d.CreateCampaign(rsID, "Test Campaign", "")
	require.NoError(t, err)
	sessID, err = d.CreateSession(campID, "S1", "2026-04-03")
	require.NoError(t, err)
	return
}

func TestListCampaigns_empty(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/campaigns", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var campaigns []db.Campaign
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &campaigns))
	assert.Empty(t, campaigns)
}

func TestListCampaigns_withData(t *testing.T) {
	s := newTestServer(t)
	campID, _ := seedCampaign(t, s.db)
	req := httptest.NewRequest(http.MethodGet, "/api/campaigns", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var campaigns []db.Campaign
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &campaigns))
	require.Len(t, campaigns, 1)
	assert.Equal(t, campID, campaigns[0].ID)
}

func TestListCharacters_empty(t *testing.T) {
	s := newTestServer(t)
	campID, _ := seedCampaign(t, s.db)
	req := httptest.NewRequest(http.MethodGet, "/api/campaigns/"+strconv.FormatInt(campID, 10)+"/characters", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var chars []db.Character
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &chars))
	assert.Empty(t, chars)
}

func TestListCharacters_withData(t *testing.T) {
	s := newTestServer(t)
	campID, _ := seedCampaign(t, s.db)
	charID, err := s.db.CreateCharacter(campID, "Kael")
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodGet, "/api/campaigns/"+strconv.FormatInt(campID, 10)+"/characters", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var chars []db.Character
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &chars))
	require.Len(t, chars, 1)
	assert.Equal(t, charID, chars[0].ID)
}

func TestListSessions_withData(t *testing.T) {
	s := newTestServer(t)
	campID, sessID := seedCampaign(t, s.db)
	req := httptest.NewRequest(http.MethodGet, "/api/campaigns/"+strconv.FormatInt(campID, 10)+"/sessions", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var sessions []db.Session
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &sessions))
	require.Len(t, sessions, 1)
	assert.Equal(t, sessID, sessions[0].ID)
}

func TestListMessages_empty(t *testing.T) {
	s := newTestServer(t)
	_, sessID := seedCampaign(t, s.db)
	req := httptest.NewRequest(http.MethodGet, "/api/sessions/"+strconv.FormatInt(sessID, 10)+"/messages", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var msgs []db.Message
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &msgs))
	assert.Empty(t, msgs)
}

func TestListDiceRolls_empty(t *testing.T) {
	s := newTestServer(t)
	_, sessID := seedCampaign(t, s.db)
	req := httptest.NewRequest(http.MethodGet, "/api/sessions/"+strconv.FormatInt(sessID, 10)+"/dice-rolls", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var rolls []db.DiceRoll
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &rolls))
	assert.Empty(t, rolls)
}

func TestListMapPins_empty(t *testing.T) {
	s := newTestServer(t)
	campID, _ := seedCampaign(t, s.db)
	mapID, err := s.db.CreateMap(campID, "World Map", "")
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodGet, "/api/maps/"+strconv.FormatInt(mapID, 10)+"/pins", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var pins []db.MapPin
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &pins))
	assert.Empty(t, pins)
}

func TestGetContext_empty(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/context", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp contextResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Nil(t, resp.Campaign)
	assert.Nil(t, resp.Character)
	assert.Nil(t, resp.Session)
	assert.Empty(t, resp.RecentMessages)
	assert.Nil(t, resp.ActiveCombat)
}

func TestListWorldNotes_empty(t *testing.T) {
	s := newTestServer(t)
	campID, _ := seedCampaign(t, s.db)
	req := httptest.NewRequest(http.MethodGet, "/api/campaigns/"+strconv.FormatInt(campID, 10)+"/world-notes", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var notes []db.WorldNote
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &notes))
	assert.Empty(t, notes)
}

func TestListWorldNotes_withData(t *testing.T) {
	s := newTestServer(t)
	campID, _ := seedCampaign(t, s.db)
	_, err := s.db.CreateWorldNote(campID, "Tavern", "A seedy place.", "location")
	require.NoError(t, err)
	_, err = s.db.CreateWorldNote(campID, "Dragon", "Ancient red dragon.", "npc")
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodGet, "/api/campaigns/"+strconv.FormatInt(campID, 10)+"/world-notes", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var notes []db.WorldNote
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &notes))
	require.Len(t, notes, 2)
}

func TestListWorldNotes_searchFilter(t *testing.T) {
	s := newTestServer(t)
	campID, _ := seedCampaign(t, s.db)
	_, err := s.db.CreateWorldNote(campID, "Tavern", "A seedy place.", "location")
	require.NoError(t, err)
	_, err = s.db.CreateWorldNote(campID, "Dragon", "Ancient red dragon.", "npc")
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodGet, "/api/campaigns/"+strconv.FormatInt(campID, 10)+"/world-notes?q=tavern", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var notes []db.WorldNote
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &notes))
	require.Len(t, notes, 1)
	assert.Equal(t, "Tavern", notes[0].Title)
}

func TestGetContext_withActiveState(t *testing.T) {
	s := newTestServer(t)
	campID, sessID := seedCampaign(t, s.db)
	charID, err := s.db.CreateCharacter(campID, "Arin")
	require.NoError(t, err)
	require.NoError(t, s.db.SetSetting("active_campaign_id", strconv.FormatInt(campID, 10)))
	require.NoError(t, s.db.SetSetting("active_character_id", strconv.FormatInt(charID, 10)))
	require.NoError(t, s.db.SetSetting("active_session_id", strconv.FormatInt(sessID, 10)))

	req := httptest.NewRequest(http.MethodGet, "/api/context", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp contextResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.NotNil(t, resp.Campaign)
	assert.Equal(t, "Test Campaign", resp.Campaign.Name)
	require.NotNil(t, resp.Character)
	assert.Equal(t, "Arin", resp.Character.Name)
	require.NotNil(t, resp.Session)
	assert.Equal(t, "S1", resp.Session.Title)
	assert.Nil(t, resp.ActiveCombat)
}
