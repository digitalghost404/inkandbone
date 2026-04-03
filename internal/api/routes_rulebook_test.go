package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIngestRulebook_textPlain(t *testing.T) {
	s := newTestServer(t)
	rulesets, err := s.db.ListRulesets()
	require.NoError(t, err)
	require.NotEmpty(t, rulesets, "expected at least one seeded ruleset")
	rsID := rulesets[0].ID

	body := "# Chapter 1\nThis is the first chapter content.\n# Chapter 2\nThis is the second chapter content.\n"
	req := httptest.NewRequest(http.MethodPost,
		"/api/rulesets/"+strconv.FormatInt(rsID, 10)+"/rulebook",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		ChunksCreated int `json:"chunks_created"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 2, resp.ChunksCreated)
}

func TestIngestRulebook_noHeadings(t *testing.T) {
	s := newTestServer(t)
	rulesets, err := s.db.ListRulesets()
	require.NoError(t, err)
	require.NotEmpty(t, rulesets, "expected at least one seeded ruleset")
	rsID := rulesets[0].ID

	body := "This is plain text without any headings. It should become one chunk.\n"
	req := httptest.NewRequest(http.MethodPost,
		"/api/rulesets/"+strconv.FormatInt(rsID, 10)+"/rulebook",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		ChunksCreated int `json:"chunks_created"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 1, resp.ChunksCreated)
}

func TestIngestRulebook_pdf(t *testing.T) {
	// TODO: PDF integration test requires fixture file
	t.Skip("PDF integration test requires a valid PDF fixture file")
}
