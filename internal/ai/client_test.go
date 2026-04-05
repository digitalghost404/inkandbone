package ai_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/digitalghost404/inkandbone/internal/ai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Generate_ok(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key", r.Header.Get("x-api-key"))
		assert.Equal(t, "2023-06-01", r.Header.Get("anthropic-version"))
		json.NewEncoder(w).Encode(map[string]any{
			"content": []map[string]string{{"type": "text", "text": "generated text"}},
		})
	}))
	defer srv.Close()

	client := ai.NewClientWithURL("test-key", srv.URL)
	result, err := client.Generate(context.Background(), "test prompt", 256)
	require.NoError(t, err)
	assert.Equal(t, "generated text", result)
}

func TestClient_Generate_emptyContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"content": []any{}})
	}))
	defer srv.Close()

	client := ai.NewClientWithURL("test-key", srv.URL)
	_, err := client.Generate(context.Background(), "prompt", 256)
	assert.ErrorContains(t, err, "empty response")
}

func TestClient_Generate_httpError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := ai.NewClientWithURL("bad-key", srv.URL)
	_, err := client.Generate(context.Background(), "prompt", 256)
	assert.ErrorContains(t, err, "401")
}
