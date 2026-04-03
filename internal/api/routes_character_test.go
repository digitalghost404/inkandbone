package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/digitalghost404/inkandbone/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRuleset(t *testing.T) {
	s := newTestServer(t)
	// seedCampaign creates a ruleset; find it via ListRulesets
	seedCampaign(t, s.db)
	rulesets, err := s.db.ListRulesets()
	require.NoError(t, err)
	require.NotEmpty(t, rulesets)
	rs := rulesets[0]

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/rulesets/%d", rs.ID), nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var got db.Ruleset
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, rs.ID, got.ID)
	assert.Equal(t, rs.Name, got.Name)
}

func TestGetRuleset_notFound(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/rulesets/9999", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestPatchCharacter(t *testing.T) {
	s := newTestServer(t)
	campID, _ := seedCampaign(t, s.db)
	charID, err := s.db.CreateCharacter(campID, "Kael")
	require.NoError(t, err)

	// Subscribe before the request so we capture the event
	ch := s.bus.Subscribe()

	body := `{"data_json":"{\"hp\":10}"}`
	req := httptest.NewRequest(http.MethodPatch,
		fmt.Sprintf("/api/characters/%d", charID),
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify DB updated
	char, err := s.db.GetCharacter(charID)
	require.NoError(t, err)
	assert.Equal(t, `{"hp":10}`, char.DataJSON)

	// Verify event published
	var got Event
	select {
	case got = <-ch:
	default:
		t.Fatal("expected character_updated event, got none")
	}
	assert.Equal(t, EventCharacterUpdated, got.Type)
	payload, ok := got.Payload.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, charID, payload["id"])
}

func TestUploadPortrait(t *testing.T) {
	dir := t.TempDir()
	s := newTestServerWithDir(t, dir)
	campID, _ := seedCampaign(t, s.db)
	charID, err := s.db.CreateCharacter(campID, "Mira")
	require.NoError(t, err)

	// Subscribe before the request so we capture the event
	ch := s.bus.Subscribe()

	// Build multipart body
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("portrait", "avatar.jpg")
	require.NoError(t, err)
	_, err = io.WriteString(fw, "fake-image-bytes")
	require.NoError(t, err)
	mw.Close()

	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/characters/%d/portrait", charID),
		&body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		PortraitPath string `json:"portrait_path"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp.PortraitPath)
	assert.True(t, strings.HasPrefix(resp.PortraitPath, "portraits/"))

	// Verify DB updated
	char, err := s.db.GetCharacter(charID)
	require.NoError(t, err)
	assert.Equal(t, resp.PortraitPath, char.PortraitPath)

	// Verify event published
	var got Event
	select {
	case got = <-ch:
	default:
		t.Fatal("expected character_updated event, got none")
	}
	assert.Equal(t, EventCharacterUpdated, got.Type)
	payload, ok := got.Payload.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, charID, payload["id"])
	assert.Equal(t, resp.PortraitPath, payload["portrait_path"])
}
