package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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
