package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleOracleRoll(t *testing.T) {
	s := newTestServer(t)

	req := httptest.NewRequest("POST", "/api/oracle/roll", strings.NewReader(`{"table":"action","roll":1}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "Attack", resp["result"])
	assert.Equal(t, "action", resp["table"])
	assert.EqualValues(t, 1, resp["roll"])
}

func TestHandleGetTension(t *testing.T) {
	s := newTestServer(t)
	campID, sessID := seedCampaign(t, s.db)
	_ = campID

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/sessions/%d/tension", sessID), nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.EqualValues(t, 5, resp["tension_level"]) // default is 5
}

func TestHandlePatchTension(t *testing.T) {
	s := newTestServer(t)
	campID, sessID := seedCampaign(t, s.db)
	_ = campID

	req := httptest.NewRequest("PATCH", fmt.Sprintf("/api/sessions/%d/tension", sessID), strings.NewReader(`{"tension_level":8}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	level, err := s.db.GetTension(sessID)
	require.NoError(t, err)
	assert.Equal(t, 8, level)
}

func TestRelationshipCRUD_API(t *testing.T) {
	s := newTestServer(t)
	campaignID, _ := seedCampaign(t, s.db)

	// Create
	body := `{"from_name":"Elara","to_name":"Doran","relationship_type":"ally","description":"Old friends"}`
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/campaigns/%d/relationships", campaignID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var created map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&created))
	id := int64(created["id"].(float64))

	// List
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/campaigns/%d/relationships", campaignID), nil)
	w = httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var list []map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&list))
	assert.Len(t, list, 1)

	// Update
	req = httptest.NewRequest("PATCH", fmt.Sprintf("/api/relationships/%d", id), strings.NewReader(`{"relationship_type":"rival","description":"They fell out"}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Delete
	req = httptest.NewRequest("DELETE", fmt.Sprintf("/api/relationships/%d", id), nil)
	w = httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)

	rels, _ := s.db.ListRelationships(campaignID)
	assert.Len(t, rels, 0)
}

func TestGetMasqueradeIntegrity_default(t *testing.T) {
	s := newTestServer(t)
	campID, sessID := seedCampaign(t, s.db)
	_ = campID

	req := httptest.NewRequest(http.MethodGet, "/api/sessions/"+strconv.FormatInt(sessID, 10)+"/masquerade", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, float64(10), resp["masquerade_integrity"])
}

func TestPatchMasqueradeIntegrity(t *testing.T) {
	s := newTestServer(t)
	campID, sessID := seedCampaign(t, s.db)
	_ = campID

	body, _ := json.Marshal(map[string]any{"masquerade_integrity": 7})
	req := httptest.NewRequest(http.MethodPatch, "/api/sessions/"+strconv.FormatInt(sessID, 10)+"/masquerade", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify via GET
	req2 := httptest.NewRequest(http.MethodGet, "/api/sessions/"+strconv.FormatInt(sessID, 10)+"/masquerade", nil)
	w2 := httptest.NewRecorder()
	s.ServeHTTP(w2, req2)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w2.Body.Bytes(), &resp))
	assert.Equal(t, float64(7), resp["masquerade_integrity"])
}
