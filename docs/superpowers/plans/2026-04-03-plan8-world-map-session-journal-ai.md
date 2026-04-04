# World Map + Session Journal + AI Content Generation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add MapPanel (map image + clickable pins), JournalPanel (session summary auto-save + AI recap), "Draft with Claude" world note generation, and the `generate_session_recap` MCP tool.

**Architecture:** A new `internal/ai` package defines a `Completer` interface and a `Client` struct that POST to the Anthropic Messages API directly over HTTP (no SDK dependency). Uploaded map images land in `~/.ttrpg/maps/`, served via `GET /api/files/{path...}` with path-traversal protection. `api.Server` and `mcp.Server` both receive an `ai.Completer` (nil when `ANTHROPIC_API_KEY` is unset). MapPanel fetches `GET /api/campaigns/{id}/maps`, renders the first map image with absolute-positioned pin buttons, and re-fetches pins when a `map_pin_added` WS event arrives. JournalPanel auto-saves on blur via `PATCH /api/sessions/{id}` and updates from `session_updated` WS events. WorldNotesPanel gains a "Draft with Claude" button backed by `POST /api/campaigns/{id}/world-notes/draft`. This plan builds on Plan 6 (useWebSocket returns `lastEvent`) and Plan 7 (timeline endpoint added).

**Tech Stack:** Go (net/http multipart, crypto/rand for filenames, no new dependencies), SQLite (existing `maps`, `map_pins`, `sessions` tables), React + TypeScript, @testing-library/react, vitest.

---

## File Map

**New files:**
- `internal/ai/client.go` — `Completer` interface + `Client` struct (raw HTTP to Anthropic)
- `internal/ai/client_test.go` — stub HTTP server tests
- `internal/mcp/session_recap.go` — `handleGenerateSessionRecap` MCP handler
- `internal/mcp/session_recap_test.go` — test for the MCP handler
- `web/src/MapPanel.tsx` — map image + pin overlay + file upload
- `web/src/MapPanel.test.tsx`
- `web/src/JournalPanel.tsx` — session summary textarea + AI recap button
- `web/src/JournalPanel.test.tsx`

**Modified files:**
- `internal/db/queries_world.go` — add `ListMaps`
- `internal/db/queries_world_test.go` — add `TestListMaps`
- `internal/api/events.go` — add `EventSessionUpdated`
- `internal/api/server.go` — add `dataDir`, `aiClient ai.Completer`; update `NewServer`, `registerRoutes`, `handleHealth`
- `internal/api/server_test.go` — update `newTestServer` signature; add `newTestServerWithDir`, `newTestServerWithAI`
- `internal/api/routes.go` — add `handleServeFile`, `handleListMaps`, `handleGetMap`, `handleUploadMap`, `handlePatchSession`, `handleGenerateRecap`, `handleDraftWorldNote`, `buildRecap`, `parseGeneratedNote`
- `internal/api/routes_test.go` — tests for all new routes
- `internal/mcp/server.go` — add `aiClient ai.Completer`; update `New`, `registerTools`
- `internal/mcp/mcp_test.go` — update `newTestMCP` signature; add `newTestMCPWithAI`
- `cmd/ttrpg/main.go` — create AI client; pass `dataDir` to `NewServer`; pass `aiClient` to `mcp.New`
- `web/src/types.ts` — add `CampaignMap`, `MapPin` interfaces
- `web/src/api.ts` — add `fetchMaps`, `fetchMapPins`, `uploadMap`, `patchSessionSummary`, `generateRecap`, `draftWorldNote`
- `web/src/api.test.ts` — tests for new functions
- `web/src/WorldNotesPanel.tsx` — add "Draft with Claude" button + hint input
- `web/src/WorldNotesPanel.test.tsx` — add draft tests
- `web/src/App.tsx` — add `MapPanel`, `JournalPanel`
- `web/src/App.test.tsx` — update mock fetch to handle new routes
- `web/src/App.css` — styles for MapPanel, JournalPanel, pin, popover, draft UI

---

### Task 1: DB `ListMaps` + `EventSessionUpdated`

**Files:**
- Modify: `internal/db/queries_world.go`
- Modify: `internal/db/queries_world_test.go`
- Modify: `internal/api/events.go`

- [ ] **Step 1: Write failing test for `ListMaps`**

In `internal/db/queries_world_test.go`, add after existing tests:

```go
func TestListMaps(t *testing.T) {
	d := newTestDB(t)
	rsID, err := d.CreateRuleset("testruleset", "{}", "1.0")
	require.NoError(t, err)
	campID, err := d.CreateCampaign(rsID, "Test Campaign", "")
	require.NoError(t, err)

	maps, err := d.ListMaps(campID)
	require.NoError(t, err)
	assert.Empty(t, maps)

	id1, err := d.CreateMap(campID, "World Map", "maps/abc.jpg")
	require.NoError(t, err)
	_, err = d.CreateMap(campID, "Dungeon", "maps/def.jpg")
	require.NoError(t, err)

	maps, err = d.ListMaps(campID)
	require.NoError(t, err)
	require.Len(t, maps, 2)
	assert.Equal(t, id1, maps[0].ID)
	assert.Equal(t, "World Map", maps[0].Name)
	assert.Equal(t, "maps/abc.jpg", maps[0].ImagePath)
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/db/... -run TestListMaps -v
```
Expected: FAIL — `d.ListMaps undefined`

- [ ] **Step 3: Add `ListMaps` to `internal/db/queries_world.go`**

After the existing `GetMap` function, add:

```go
func (d *DB) ListMaps(campaignID int64) ([]Map, error) {
	rows, err := d.db.Query(
		"SELECT id, campaign_id, name, image_path, created_at FROM maps WHERE campaign_id = ? ORDER BY created_at",
		campaignID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Map
	for rows.Next() {
		var m Map
		if err := rows.Scan(&m.ID, &m.CampaignID, &m.Name, &m.ImagePath, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
```

- [ ] **Step 4: Add `EventSessionUpdated` to `internal/api/events.go`**

In the `const` block, add after `EventCharacterCreated`:

```go
EventSessionUpdated EventType = "session_updated"
```

- [ ] **Step 5: Run test to verify it passes**

```bash
go test ./internal/db/... -run TestListMaps -v
```
Expected: PASS

- [ ] **Step 6: Run all DB tests for regressions**

```bash
go test ./internal/db/... -v
```
Expected: all PASS

- [ ] **Step 7: Commit**

```bash
git add internal/db/queries_world.go internal/db/queries_world_test.go internal/api/events.go
git commit -m "feat: add ListMaps DB query and EventSessionUpdated constant"
```

---

### Task 2: AI client package

**Files:**
- Create: `internal/ai/client.go`
- Create: `internal/ai/client_test.go`

- [ ] **Step 1: Write failing test**

Create `internal/ai/client_test.go`:

```go
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
	result, err := client.Generate(context.Background(), "test prompt")
	require.NoError(t, err)
	assert.Equal(t, "generated text", result)
}

func TestClient_Generate_emptyContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"content": []any{}})
	}))
	defer srv.Close()

	client := ai.NewClientWithURL("test-key", srv.URL)
	_, err := client.Generate(context.Background(), "prompt")
	assert.ErrorContains(t, err, "empty response")
}

func TestClient_Generate_httpError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := ai.NewClientWithURL("bad-key", srv.URL)
	_, err := client.Generate(context.Background(), "prompt")
	assert.ErrorContains(t, err, "401")
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/ai/... -v
```
Expected: FAIL — package `ai` does not exist

- [ ] **Step 3: Create `internal/ai/client.go`**

```go
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	defaultURL       = "https://api.anthropic.com/v1/messages"
	anthropicVersion = "2023-06-01"
	model            = "claude-haiku-4-5-20251001"
)

// Completer generates text from a prompt. Implemented by *Client; nil means AI is disabled.
type Completer interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

// Client calls the Anthropic Messages API over plain HTTP.
type Client struct {
	apiKey string
	url    string
	http   *http.Client
}

// NewClient returns a Client using the production Anthropic API URL.
func NewClient(apiKey string) *Client {
	return &Client{apiKey: apiKey, url: defaultURL, http: &http.Client{}}
}

// NewClientWithURL returns a Client using a custom URL (for tests).
func NewClientWithURL(apiKey, url string) *Client {
	return &Client{apiKey: apiKey, url: url, http: &http.Client{}}
}

func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	body, err := json.Marshal(map[string]any{
		"model":      model,
		"max_tokens": 1024,
		"messages":   []map[string]any{{"role": "user", "content": prompt}},
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)
	req.Header.Set("content-type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("anthropic API returned %d", resp.StatusCode)
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Content) == 0 {
		return "", fmt.Errorf("empty response from Anthropic")
	}
	return result.Content[0].Text, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/ai/... -v
```
Expected: all 3 PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ai/
git commit -m "feat: add AI client package with Completer interface"
```

---

### Task 3: Server restructure — add dataDir + AI client + file serving

**Files:**
- Modify: `internal/api/server.go`
- Modify: `internal/api/server_test.go`
- Modify: `internal/api/routes.go`
- Modify: `internal/api/routes_test.go`
- Modify: `cmd/ttrpg/main.go`

- [ ] **Step 1: Write failing tests for the file serving endpoint**

In `internal/api/routes_test.go`, add:

```go
import (
	// existing imports
	"os"
	"path/filepath"
	"strings"
)

func TestServeFile_ok(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("world"), 0600))
	s := newTestServerWithDir(t, dir)
	req := httptest.NewRequest(http.MethodGet, "/api/files/hello.txt", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "world", w.Body.String())
}

func TestServeFile_traversal(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/files/../etc/passwd", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/api/... -run "TestServeFile" -v
```
Expected: FAIL — compilation error (`NewServer` signature unchanged)

- [ ] **Step 3: Rewrite `internal/api/server.go`**

Replace the entire file with:

```go
package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/digitalghost404/inkandbone/internal/ai"
	"github.com/digitalghost404/inkandbone/internal/db"
)

// Server holds dependencies and registers routes.
type Server struct {
	db       *db.DB
	hub      *Hub
	bus      *Bus
	mux      *http.ServeMux
	dataDir  string
	aiClient ai.Completer // nil when ANTHROPIC_API_KEY is unset
}

// NewServer creates the HTTP server. dataDir is the base path for uploaded files
// (e.g. ~/.ttrpg). aiClient may be nil if AI features are disabled.
func NewServer(database *db.DB, dataDir string, aiClient ai.Completer) *Server {
	bus := NewBus()
	hub := NewHub(bus)
	s := &Server{
		db:       database,
		hub:      hub,
		bus:      bus,
		mux:      http.NewServeMux(),
		dataDir:  dataDir,
		aiClient: aiClient,
	}
	s.registerRoutes()
	go hub.Run()
	return s
}

// Bus returns the event bus so the MCP server can publish events.
func (s *Server) Bus() *Bus { return s.bus }

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// ListenAndServe starts the HTTP server on addr (e.g. ":7432").
func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s)
}

// Shutdown is a no-op placeholder.
func (s *Server) Shutdown(_ context.Context) error { return nil }

// RegisterStatic serves the embedded React SPA for all routes not matched by /api/ or /ws.
func (s *Server) RegisterStatic(fsys http.FileSystem) {
	fileServer := http.FileServer(fsys)
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fileServer.ServeHTTP(w, r)
	})
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/ws", s.hub.ServeWS)
	s.mux.HandleFunc("/api/health", s.handleHealth)
	// Existing read routes
	s.mux.HandleFunc("GET /api/campaigns", s.handleListCampaigns)
	s.mux.HandleFunc("GET /api/campaigns/{id}/characters", s.handleListCharacters)
	s.mux.HandleFunc("GET /api/campaigns/{id}/sessions", s.handleListSessions)
	s.mux.HandleFunc("GET /api/campaigns/{id}/world-notes", s.handleListWorldNotes)
	s.mux.HandleFunc("GET /api/sessions/{id}/messages", s.handleListMessages)
	s.mux.HandleFunc("GET /api/sessions/{id}/dice-rolls", s.handleListDiceRolls)
	s.mux.HandleFunc("GET /api/maps/{id}/pins", s.handleListMapPins)
	s.mux.HandleFunc("GET /api/context", s.handleGetContext)
	// Plan 7
	s.mux.HandleFunc("GET /api/sessions/{id}/timeline", s.handleGetTimeline)
	// Plan 8
	s.mux.HandleFunc("GET /api/files/{path...}", s.handleServeFile)
	s.mux.HandleFunc("GET /api/campaigns/{id}/maps", s.handleListMaps)
	s.mux.HandleFunc("POST /api/campaigns/{id}/maps", s.handleUploadMap)
	s.mux.HandleFunc("GET /api/maps/{id}", s.handleGetMap)
	s.mux.HandleFunc("PATCH /api/sessions/{id}", s.handlePatchSession)
	s.mux.HandleFunc("POST /api/sessions/{id}/recap", s.handleGenerateRecap)
	s.mux.HandleFunc("POST /api/campaigns/{id}/world-notes/draft", s.handleDraftWorldNote)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":     "ok",
		"ai_enabled": s.aiClient != nil,
	})
}
```

- [ ] **Step 4: Update `internal/api/server_test.go`**

Replace `newTestServer` and add two more helpers:

```go
package api

import (
	"testing"

	"github.com/digitalghost404/inkandbone/internal/db"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	d, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { d.Close() })
	return NewServer(d, t.TempDir(), nil)
}

func newTestServerWithDir(t *testing.T, dir string) *Server {
	t.Helper()
	d, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { d.Close() })
	return NewServer(d, dir, nil)
}

// completerFunc is a function that satisfies ai.Completer.
type completerFunc func(ctx interface{}, prompt string) (string, error)

// stubCompleter is a test double for ai.Completer.
type stubCompleter struct{ response string }

func (s *stubCompleter) Generate(_ interface{ Done() <-chan struct{} }, _ string) (string, error) {
	return s.response, nil
}
```

Wait — the stub needs to satisfy `ai.Completer` which has signature `Generate(context.Context, string) (string, error)`. Fix:

```go
package api

import (
	"context"
	"testing"

	"github.com/digitalghost404/inkandbone/internal/ai"
	"github.com/digitalghost404/inkandbone/internal/db"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	d, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { d.Close() })
	return NewServer(d, t.TempDir(), nil)
}

func newTestServerWithDir(t *testing.T, dir string) *Server {
	t.Helper()
	d, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { d.Close() })
	return NewServer(d, dir, nil)
}

type stubCompleter struct{ response string }

func (s *stubCompleter) Generate(_ context.Context, _ string) (string, error) {
	return s.response, nil
}

func newTestServerWithAI(t *testing.T, c ai.Completer) *Server {
	t.Helper()
	d, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { d.Close() })
	return NewServer(d, t.TempDir(), c)
}
```

- [ ] **Step 5: Add `handleServeFile` to `internal/api/routes.go`**

Add the following imports and handler. The existing imports block already has `encoding/json`, `net/http`, `strconv`. Add `path/filepath` and `strings`:

```go
import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/digitalghost404/inkandbone/internal/db"
)

func (s *Server) handleServeFile(w http.ResponseWriter, r *http.Request) {
	rel := filepath.Clean(r.PathValue("path"))
	// Reject any path that tries to escape the data directory
	if strings.HasPrefix(rel, "..") {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	abs := filepath.Join(s.dataDir, rel)
	if !strings.HasPrefix(abs+string(filepath.Separator), s.dataDir+string(filepath.Separator)) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	http.ServeFile(w, r, abs)
}
```

- [ ] **Step 6: Update `cmd/ttrpg/main.go`**

```go
package main

import (
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/digitalghost404/inkandbone/internal/ai"
	"github.com/digitalghost404/inkandbone/internal/api"
	"github.com/digitalghost404/inkandbone/internal/db"
	mcpserver "github.com/digitalghost404/inkandbone/internal/mcp"
	ttrpgweb "github.com/digitalghost404/inkandbone/web"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("home dir: %v", err)
	}
	dbPath := filepath.Join(home, ".ttrpg", "ttrpg.db")
	dataDir := filepath.Join(home, ".ttrpg")

	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer database.Close()

	var aiClient ai.Completer
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		aiClient = ai.NewClient(key)
		log.Println("AI features enabled")
	}

	httpServer := api.NewServer(database, dataDir, aiClient)

	distFS, err := fs.Sub(ttrpgweb.Static, "dist")
	if err != nil {
		log.Fatalf("embed sub: %v", err)
	}
	httpServer.RegisterStatic(http.FS(distFS))

	go func() {
		log.Println("HTTP server listening on :7432")
		if err := httpServer.ListenAndServe(":7432"); err != nil {
			log.Printf("HTTP server stopped: %v", err)
		}
	}()

	mcpSrv := mcpserver.New(database, httpServer.Bus()) // aiClient added in Task 5
	if err := mcpSrv.Start(); err != nil {
		log.Fatalf("MCP server: %v", err)
	}
}
```

- [ ] **Step 7: Run all tests to verify nothing is broken**

```bash
go test ./... -v 2>&1 | tail -30
```
Expected: PASS for all packages (including new `TestServeFile_ok`, `TestServeFile_traversal`)

- [ ] **Step 8: Commit**

```bash
git add internal/api/server.go internal/api/server_test.go internal/api/routes.go internal/api/routes_test.go cmd/ttrpg/main.go
git commit -m "feat: add dataDir + AI client to server; file-serving endpoint"
```

---

### Task 4: Map + session HTTP handlers

**Files:**
- Modify: `internal/api/routes.go`
- Modify: `internal/api/routes_test.go`

- [ ] **Step 1: Write failing tests**

In `internal/api/routes_test.go`, add:

```go
import (
	// existing
	"bytes"
	"io"
	"mime/multipart"
)

func TestListMaps_empty(t *testing.T) {
	s := newTestServer(t)
	campID, _ := seedCampaign(t, s.db)
	req := httptest.NewRequest(http.MethodGet, "/api/campaigns/"+strconv.FormatInt(campID, 10)+"/maps", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var maps []db.Map
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &maps))
	assert.Empty(t, maps)
}

func TestGetMap_found(t *testing.T) {
	s := newTestServer(t)
	campID, _ := seedCampaign(t, s.db)
	mapID, err := s.db.CreateMap(campID, "World", "maps/test.jpg")
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodGet, "/api/maps/"+strconv.FormatInt(mapID, 10), nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var m db.Map
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &m))
	assert.Equal(t, mapID, m.ID)
	assert.Equal(t, "World", m.Name)
}

func TestGetMap_notFound(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/maps/9999", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestPatchSession_ok(t *testing.T) {
	s := newTestServer(t)
	_, sessID := seedCampaign(t, s.db)
	body := `{"summary":"Session went great"}`
	req := httptest.NewRequest(http.MethodPatch,
		"/api/sessions/"+strconv.FormatInt(sessID, 10),
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
	sess, err := s.db.GetSession(sessID)
	require.NoError(t, err)
	assert.Equal(t, "Session went great", sess.Summary)
}

func TestUploadMap_ok(t *testing.T) {
	dir := t.TempDir()
	s := newTestServerWithDir(t, dir)
	campID, _ := seedCampaign(t, s.db)

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("image", "map.png")
	require.NoError(t, err)
	_, err = io.WriteString(fw, "fake-image-data")
	require.NoError(t, err)
	require.NoError(t, mw.WriteField("name", "World Map"))
	mw.Close()

	req := httptest.NewRequest(http.MethodPost,
		"/api/campaigns/"+strconv.FormatInt(campID, 10)+"/maps",
		&body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	var m db.Map
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &m))
	assert.Equal(t, campID, m.CampaignID)
	assert.Equal(t, "World Map", m.Name)
	assert.True(t, strings.HasPrefix(m.ImagePath, "maps/"))
	assert.FileExists(t, filepath.Join(dir, "maps", filepath.Base(m.ImagePath)))
}
```

The `strings` and `filepath` imports should already be present from Task 3 (for `TestServeFile`). If not, add them.

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/api/... -run "TestListMaps|TestGetMap|TestPatchSession|TestUploadMap" -v
```
Expected: FAIL — handlers undefined

- [ ] **Step 3: Add handlers to `internal/api/routes.go`**

Add these imports if not already present: `"fmt"`, `"io"`, `"os"`.

```go
import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/digitalghost404/inkandbone/internal/db"
)

func (s *Server) handleListMaps(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}
	maps, err := s.db.ListMaps(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if maps == nil {
		maps = []db.Map{}
	}
	writeJSON(w, maps)
}

func (s *Server) handleGetMap(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid map id", http.StatusBadRequest)
		return
	}
	m, err := s.db.GetMap(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if m == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, m)
}

func (s *Server) handleUploadMap(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "parse form: "+err.Error(), http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "image is required: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	filename := randomHex(16) + ext
	destDir := filepath.Join(s.dataDir, "maps")
	if err := os.MkdirAll(destDir, 0750); err != nil {
		http.Error(w, "mkdir: "+err.Error(), http.StatusInternalServerError)
		return
	}
	out, err := os.Create(filepath.Join(destDir, filename))
	if err != nil {
		http.Error(w, "create file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, "write file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	imagePath := "maps/" + filename
	mapID, err := s.db.CreateMap(id, name, imagePath)
	if err != nil {
		http.Error(w, "db: "+err.Error(), http.StatusInternalServerError)
		return
	}
	m, err := s.db.GetMap(mapID)
	if err != nil || m == nil {
		http.Error(w, "fetch created map", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(m)
}

// randomHex returns n random hex bytes.
func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b) //nolint:errcheck // crypto/rand.Read never returns an error on supported platforms
	return fmt.Sprintf("%x", b)
}

func (s *Server) handlePatchSession(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	var body struct {
		Summary string `json:"summary"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := s.db.UpdateSessionSummary(id, body.Summary); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.bus.Publish(Event{Type: EventSessionUpdated, Payload: map[string]any{
		"session_id": id,
		"summary":    body.Summary,
	}})
	w.WriteHeader(http.StatusNoContent)
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/api/... -run "TestListMaps|TestGetMap|TestPatchSession|TestUploadMap" -v
```
Expected: all PASS

- [ ] **Step 5: Run all API tests for regressions**

```bash
go test ./internal/api/... -v
```
Expected: all PASS

- [ ] **Step 6: Commit**

```bash
git add internal/api/routes.go internal/api/routes_test.go
git commit -m "feat: map list/get/upload and session PATCH HTTP handlers"
```

---

### Task 5: AI-powered HTTP handlers — draft world note + generate recap

**Files:**
- Modify: `internal/api/routes.go` — add `handleDraftWorldNote`, `handleGenerateRecap`, `buildRecap`, `parseGeneratedNote`
- Modify: `internal/api/routes_test.go` — tests using stubCompleter

- [ ] **Step 1: Write failing tests**

In `internal/api/routes_test.go`, add:

```go
func TestDraftWorldNote_ok(t *testing.T) {
	stub := &stubCompleter{response: "Title: Zara the Smith\nContent: A dwarven blacksmith known for fine steel."}
	s := newTestServerWithAI(t, stub)
	campID, _ := seedCampaign(t, s.db)

	body := `{"hint":"Dwarven blacksmith NPC"}`
	req := httptest.NewRequest(http.MethodPost,
		"/api/campaigns/"+strconv.FormatInt(campID, 10)+"/world-notes/draft",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	var note db.WorldNote
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &note))
	assert.NotZero(t, note.ID)
	assert.Equal(t, campID, note.CampaignID)
	assert.Equal(t, "Zara the Smith", note.Title)
}

func TestDraftWorldNote_noAI(t *testing.T) {
	s := newTestServer(t) // aiClient is nil
	campID, _ := seedCampaign(t, s.db)
	body := `{"hint":"test"}`
	req := httptest.NewRequest(http.MethodPost,
		"/api/campaigns/"+strconv.FormatInt(campID, 10)+"/world-notes/draft",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestGenerateRecap_ok(t *testing.T) {
	stub := &stubCompleter{response: "The party defeated the goblin horde."}
	s := newTestServerWithAI(t, stub)
	_, sessID := seedCampaign(t, s.db)

	req := httptest.NewRequest(http.MethodPost,
		"/api/sessions/"+strconv.FormatInt(sessID, 10)+"/recap",
		nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Summary string `json:"summary"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "The party defeated the goblin horde.", resp.Summary)
	// Verify DB was updated
	sess, err := s.db.GetSession(sessID)
	require.NoError(t, err)
	assert.Equal(t, "The party defeated the goblin horde.", sess.Summary)
}

func TestGenerateRecap_noAI(t *testing.T) {
	s := newTestServer(t)
	_, sessID := seedCampaign(t, s.db)
	req := httptest.NewRequest(http.MethodPost,
		"/api/sessions/"+strconv.FormatInt(sessID, 10)+"/recap",
		nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/api/... -run "TestDraftWorldNote|TestGenerateRecap" -v
```
Expected: FAIL — handlers undefined

- [ ] **Step 3: Add handlers to `internal/api/routes.go`**

```go
func (s *Server) handleDraftWorldNote(w http.ResponseWriter, r *http.Request) {
	if s.aiClient == nil {
		http.Error(w, "AI not configured — set ANTHROPIC_API_KEY", http.StatusServiceUnavailable)
		return
	}
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}
	var body struct {
		Hint string `json:"hint"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Hint == "" {
		http.Error(w, "hint is required", http.StatusBadRequest)
		return
	}

	prompt := fmt.Sprintf(
		"Create a TTRPG world note for: %s\n\nRespond with exactly two lines:\nTitle: <short name>\nContent: <2-3 sentence description>",
		body.Hint,
	)
	generated, err := s.aiClient.Generate(r.Context(), prompt)
	if err != nil {
		http.Error(w, "AI error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	title, content := parseGeneratedNote(generated)
	if title == "" {
		title = body.Hint
	}
	if content == "" {
		content = generated
	}

	noteID, err := s.db.CreateWorldNote(id, title, content, "npc")
	if err != nil {
		http.Error(w, "db: "+err.Error(), http.StatusInternalServerError)
		return
	}

	notes, err := s.db.SearchWorldNotes(id, "", "")
	if err != nil {
		http.Error(w, "fetch note: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var created *db.WorldNote
	for i := range notes {
		if notes[i].ID == noteID {
			created = &notes[i]
			break
		}
	}
	if created == nil {
		http.Error(w, "note not found after create", http.StatusInternalServerError)
		return
	}

	s.bus.Publish(Event{Type: EventWorldNoteCreated, Payload: map[string]any{"note_id": noteID, "title": title}})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// parseGeneratedNote extracts title and content from a "Title: ...\nContent: ..." response.
func parseGeneratedNote(raw string) (title, content string) {
	for _, line := range strings.Split(raw, "\n") {
		if after, found := strings.CutPrefix(line, "Title: "); found {
			title = strings.TrimSpace(after)
		}
		if after, found := strings.CutPrefix(line, "Content: "); found {
			content = strings.TrimSpace(after)
		}
	}
	return
}

func (s *Server) handleGenerateRecap(w http.ResponseWriter, r *http.Request) {
	if s.aiClient == nil {
		http.Error(w, "AI not configured — set ANTHROPIC_API_KEY", http.StatusServiceUnavailable)
		return
	}
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}

	summary, err := s.buildRecap(r.Context(), id)
	if err != nil {
		http.Error(w, "recap: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.db.UpdateSessionSummary(id, summary); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.bus.Publish(Event{Type: EventSessionUpdated, Payload: map[string]any{
		"session_id": id,
		"summary":    summary,
	}})
	writeJSON(w, map[string]string{"summary": summary})
}

// buildRecap reads messages and dice rolls, builds a prompt, and calls the AI.
func (s *Server) buildRecap(ctx context.Context, sessionID int64) (string, error) {
	msgs, err := s.db.ListMessages(sessionID)
	if err != nil {
		return "", fmt.Errorf("list messages: %w", err)
	}
	rolls, err := s.db.ListDiceRolls(sessionID)
	if err != nil {
		return "", fmt.Errorf("list rolls: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("Write a 2-3 sentence narrative recap of this TTRPG session.\n\nMessages:\n")
	for _, m := range msgs {
		fmt.Fprintf(&sb, "[%s]: %s\n", m.Role, m.Content)
	}
	sb.WriteString("\nDice rolls:\n")
	for _, r := range rolls {
		fmt.Fprintf(&sb, "%s = %d\n", r.Expression, r.Result)
	}

	return s.aiClient.Generate(ctx, sb.String())
}
```

Add `"context"` to the imports if not already present.

- [ ] **Step 4: Run tests**

```bash
go test ./internal/api/... -run "TestDraftWorldNote|TestGenerateRecap" -v
```
Expected: all PASS

- [ ] **Step 5: Run all API tests**

```bash
go test ./internal/api/... -v
```
Expected: all PASS

- [ ] **Step 6: Commit**

```bash
git add internal/api/routes.go internal/api/routes_test.go
git commit -m "feat: draft world note and generate recap HTTP endpoints"
```

---

### Task 6: `generate_session_recap` MCP tool

**Files:**
- Create: `internal/mcp/session_recap.go`
- Create: `internal/mcp/session_recap_test.go`
- Modify: `internal/mcp/server.go`
- Modify: `internal/mcp/mcp_test.go`
- Modify: `cmd/ttrpg/main.go`

- [ ] **Step 1: Write failing test**

Create `internal/mcp/session_recap_test.go`:

```go
package mcp

import (
	"context"
	"strconv"
	"testing"

	"github.com/digitalghost404/inkandbone/internal/api"
	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mcpStubCompleter struct{ response string }

func (s *mcpStubCompleter) Generate(_ context.Context, _ string) (string, error) {
	return s.response, nil
}

func TestGenerateSessionRecap(t *testing.T) {
	stub := &mcpStubCompleter{response: "The heroes fought valiantly."}
	s := newTestMCPWithAI(t, stub)

	// Seed campaign and session
	rs, err := s.db.GetRulesetByName("dnd5e")
	require.NoError(t, err)
	require.NotNil(t, rs)
	campID, err := s.db.CreateCampaign(rs.ID, "Test Campaign", "")
	require.NoError(t, err)
	sessID, err := s.db.CreateSession(campID, "S1", "2026-04-03")
	require.NoError(t, err)
	require.NoError(t, s.db.SetSetting("active_session_id", strconv.FormatInt(sessID, 10)))

	// Collect WS events
	events := []api.Event{}
	ch := s.bus.Subscribe()

	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{"session_id": float64(sessID)}
	result, err := s.handleGenerateSessionRecap(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)

	// Drain one event
	select {
	case e := <-ch:
		events = append(events, e)
	default:
	}
	require.Len(t, events, 1)
	assert.Equal(t, api.EventSessionUpdated, events[0].Type)

	// Verify DB updated
	sess, err := s.db.GetSession(sessID)
	require.NoError(t, err)
	assert.Equal(t, "The heroes fought valiantly.", sess.Summary)
}

func TestGenerateSessionRecap_noAI(t *testing.T) {
	s := newTestMCP(t) // aiClient is nil
	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{"session_id": float64(1)}
	result, err := s.handleGenerateSessionRecap(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/mcp/... -run "TestGenerateSessionRecap" -v
```
Expected: FAIL — `newTestMCPWithAI` and `handleGenerateSessionRecap` undefined

- [ ] **Step 3: Update `internal/mcp/mcp_test.go`**

Replace the file content to update `newTestMCP` and add `newTestMCPWithAI`:

```go
package mcp

import (
	"testing"

	"github.com/digitalghost404/inkandbone/internal/ai"
	"github.com/digitalghost404/inkandbone/internal/api"
	"github.com/digitalghost404/inkandbone/internal/db"
	"github.com/stretchr/testify/require"
)

func newTestDB(t *testing.T) *db.DB {
	t.Helper()
	d, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { d.Close() })
	return d
}

func newTestMCP(t *testing.T) *Server {
	t.Helper()
	d := newTestDB(t)
	bus := api.NewBus()
	return New(d, bus, nil)
}

func newTestMCPWithAI(t *testing.T, c ai.Completer) *Server {
	t.Helper()
	d := newTestDB(t)
	bus := api.NewBus()
	return New(d, bus, c)
}

func TestNewServer_doesNotPanic(t *testing.T) {
	s := newTestMCP(t)
	require.NotNil(t, s)
}
```

- [ ] **Step 4: Create `internal/mcp/session_recap.go`**

```go
package mcp

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/digitalghost404/inkandbone/internal/api"
	mcplib "github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) handleGenerateSessionRecap(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	if s.aiClient == nil {
		return mcplib.NewToolResultError("AI not configured — set ANTHROPIC_API_KEY"), nil
	}

	sessID, ok := optInt64(req, "session_id")
	if !ok {
		// Fall back to active session
		var err error
		sessID, err = s.activeSessionID()
		if err != nil {
			return mcplib.NewToolResultError("session_id required or set active session first"), nil
		}
	}

	msgs, err := s.db.ListMessages(sessID)
	if err != nil {
		return mcplib.NewToolResultError("list messages: " + err.Error()), nil
	}
	rolls, err := s.db.ListDiceRolls(sessID)
	if err != nil {
		return mcplib.NewToolResultError("list rolls: " + err.Error()), nil
	}

	var sb strings.Builder
	sb.WriteString("Write a 2-3 sentence narrative recap of this TTRPG session.\n\nMessages:\n")
	for _, m := range msgs {
		fmt.Fprintf(&sb, "[%s]: %s\n", m.Role, m.Content)
	}
	sb.WriteString("\nDice rolls:\n")
	for _, r := range rolls {
		fmt.Fprintf(&sb, "%s = %d\n", r.Expression, r.Result)
	}

	summary, err := s.aiClient.Generate(context.Background(), sb.String())
	if err != nil {
		return mcplib.NewToolResultError("AI error: " + err.Error()), nil
	}

	if err := s.db.UpdateSessionSummary(sessID, summary); err != nil {
		return mcplib.NewToolResultError("update session: " + err.Error()), nil
	}
	s.bus.Publish(api.Event{Type: api.EventSessionUpdated, Payload: map[string]any{
		"session_id": sessID,
		"summary":    summary,
	}})
	return mcplib.NewToolResultText(fmt.Sprintf("session %d recap saved: %s", sessID, summary)), nil
}
```

- [ ] **Step 5: Update `internal/mcp/server.go`**

Add `aiClient ai.Completer` to the `Server` struct, update `New`, and register the new tool.

```go
package mcp

import (
	"github.com/digitalghost404/inkandbone/internal/ai"
	"github.com/digitalghost404/inkandbone/internal/api"
	"github.com/digitalghost404/inkandbone/internal/db"
	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Server wraps the MCP server and holds shared dependencies.
type Server struct {
	db       *db.DB
	bus      *api.Bus
	srv      *server.MCPServer
	aiClient ai.Completer // nil when ANTHROPIC_API_KEY is unset
}

// New creates the MCP server and registers all tools.
func New(database *db.DB, bus *api.Bus, aiClient ai.Completer) *Server {
	s := &Server{
		db:       database,
		bus:      bus,
		srv:      server.NewMCPServer("ink & bone", "1.0.0"),
		aiClient: aiClient,
	}
	s.registerTools()
	return s
}

// Start runs the MCP stdio transport. Blocks until stdin closes.
func (s *Server) Start() error {
	return server.ServeStdio(s.srv)
}

func (s *Server) registerTools() {
	// ... all existing tool registrations unchanged ...

	// Plan 8
	s.srv.AddTool(mcplib.NewTool("generate_session_recap",
		mcplib.WithDescription("Generate a narrative recap for a session using AI. Reads messages and dice rolls, writes the summary, and fires a session_updated event."),
		mcplib.WithNumber("session_id", mcplib.Description("Session ID (defaults to active session)")),
	), s.handleGenerateSessionRecap)
}
```

**Important:** Keep all existing tool registrations in `registerTools()` — add only the new one at the end. Do not remove any existing `s.srv.AddTool(...)` calls.

- [ ] **Step 6: Update `cmd/ttrpg/main.go` to pass aiClient to mcp.New**

Change this line:
```go
mcpSrv := mcpserver.New(database, httpServer.Bus())
```
to:
```go
mcpSrv := mcpserver.New(database, httpServer.Bus(), aiClient)
```

- [ ] **Step 7: Run tests**

```bash
go test ./internal/mcp/... -run "TestGenerateSessionRecap" -v
```
Expected: both PASS

- [ ] **Step 8: Run all tests**

```bash
go test ./... -v 2>&1 | grep -E "^(ok|FAIL|---)"
```
Expected: all `ok`, no `FAIL`

- [ ] **Step 9: Commit**

```bash
git add internal/mcp/session_recap.go internal/mcp/session_recap_test.go internal/mcp/server.go internal/mcp/mcp_test.go cmd/ttrpg/main.go
git commit -m "feat: generate_session_recap MCP tool with AI recap generation"
```

---

### Task 7: TypeScript types + API functions

**Files:**
- Modify: `web/src/types.ts`
- Modify: `web/src/api.ts`
- Modify: `web/src/api.test.ts`

- [ ] **Step 1: Write failing tests**

In `web/src/api.test.ts`, add:

```typescript
describe('fetchMaps', () => {
  it('fetches maps for a campaign', async () => {
    const maps = [
      { id: 1, campaign_id: 1, name: 'World Map', image_path: 'maps/abc.jpg', created_at: '' },
    ]
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve(maps) }))
    const result = await fetchMaps(1)
    expect(result).toHaveLength(1)
    expect(result[0].name).toBe('World Map')
  })
})

describe('fetchMapPins', () => {
  it('fetches pins for a map', async () => {
    const pins = [{ id: 1, map_id: 1, x: 0.5, y: 0.3, label: 'Tavern', note: '', color: '', created_at: '' }]
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve(pins) }))
    const result = await fetchMapPins(1)
    expect(result[0].label).toBe('Tavern')
  })
})

describe('patchSessionSummary', () => {
  it('sends PATCH with summary', async () => {
    const mockFetch = vi.fn().mockResolvedValue({ ok: true })
    vi.stubGlobal('fetch', mockFetch)
    await patchSessionSummary(5, 'Great session')
    expect(mockFetch).toHaveBeenCalledWith('/api/sessions/5', expect.objectContaining({
      method: 'PATCH',
    }))
  })

  it('throws on non-ok response', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false, status: 404 }))
    await expect(patchSessionSummary(5, 'x')).rejects.toThrow('404')
  })
})

describe('generateRecap', () => {
  it('calls POST recap endpoint', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ summary: 'The party won.' }),
    })
    vi.stubGlobal('fetch', mockFetch)
    const result = await generateRecap(3)
    expect(result.summary).toBe('The party won.')
    expect(mockFetch).toHaveBeenCalledWith('/api/sessions/3/recap', expect.objectContaining({ method: 'POST' }))
  })
})

describe('draftWorldNote', () => {
  it('calls POST draft endpoint with hint', async () => {
    const note = { id: 1, campaign_id: 1, title: 'Zara', content: 'A blacksmith.', category: 'npc', tags_json: '[]', created_at: '' }
    const mockFetch = vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve(note) })
    vi.stubGlobal('fetch', mockFetch)
    const result = await draftWorldNote(1, 'blacksmith')
    expect(result.title).toBe('Zara')
    expect(mockFetch).toHaveBeenCalledWith('/api/campaigns/1/world-notes/draft', expect.objectContaining({ method: 'POST' }))
  })
})
```

Add `fetchMaps`, `fetchMapPins`, `patchSessionSummary`, `generateRecap`, `draftWorldNote` to the imports in the test file.

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd web && npx vitest run src/api.test.ts 2>&1 | tail -20
```
Expected: FAIL — named exports not found

- [ ] **Step 3: Add interfaces to `web/src/types.ts`**

Append to the end of the file:

```typescript
export interface CampaignMap {
  id: number
  campaign_id: number
  name: string
  image_path: string
  created_at: string
}

export interface MapPin {
  id: number
  map_id: number
  x: number
  y: number
  label: string
  note: string
  color: string
  created_at: string
}
```

- [ ] **Step 4: Add functions to `web/src/api.ts`**

```typescript
import type { GameContext, WorldNote, DiceRoll, CampaignMap, MapPin } from './types'

export async function fetchMaps(campaignId: number): Promise<CampaignMap[]> {
  const url = `/api/campaigns/${campaignId}/maps`
  const res = await fetch(url)
  if (!res.ok) throw new Error(`GET ${url} failed: ${res.status}`)
  return res.json()
}

export async function fetchMapPins(mapId: number): Promise<MapPin[]> {
  const url = `/api/maps/${mapId}/pins`
  const res = await fetch(url)
  if (!res.ok) throw new Error(`GET ${url} failed: ${res.status}`)
  return res.json()
}

export async function patchSessionSummary(sessionId: number, summary: string): Promise<void> {
  const url = `/api/sessions/${sessionId}`
  const res = await fetch(url, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ summary }),
  })
  if (!res.ok) throw new Error(`PATCH ${url} failed: ${res.status}`)
}

export async function generateRecap(sessionId: number): Promise<{ summary: string }> {
  const url = `/api/sessions/${sessionId}/recap`
  const res = await fetch(url, { method: 'POST' })
  if (!res.ok) throw new Error(`POST ${url} failed: ${res.status}`)
  return res.json()
}

export async function draftWorldNote(campaignId: number, hint: string): Promise<WorldNote> {
  const url = `/api/campaigns/${campaignId}/world-notes/draft`
  const res = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ hint }),
  })
  if (!res.ok) throw new Error(`POST ${url} failed: ${res.status}`)
  return res.json()
}

export async function uploadMap(campaignId: number, name: string, file: File): Promise<CampaignMap> {
  const url = `/api/campaigns/${campaignId}/maps`
  const form = new FormData()
  form.append('name', name)
  form.append('image', file)
  const res = await fetch(url, { method: 'POST', body: form })
  if (!res.ok) throw new Error(`POST ${url} failed: ${res.status}`)
  return res.json()
}
```

- [ ] **Step 5: Run API tests**

```bash
cd web && npx vitest run src/api.test.ts 2>&1 | tail -20
```
Expected: all PASS

- [ ] **Step 6: Commit**

```bash
git add web/src/types.ts web/src/api.ts web/src/api.test.ts
git commit -m "feat: TypeScript CampaignMap/MapPin types and API functions"
```

---

### Task 8: MapPanel component

**Files:**
- Create: `web/src/MapPanel.tsx`
- Create: `web/src/MapPanel.test.tsx`

- [ ] **Step 1: Write failing tests**

Create `web/src/MapPanel.test.tsx`:

```tsx
import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MapPanel } from './MapPanel'

afterEach(() => vi.restoreAllMocks())

const mockMap = { id: 1, campaign_id: 1, name: 'World Map', image_path: 'maps/world.jpg', created_at: '' }
const mockPin = { id: 1, map_id: 1, x: 0.5, y: 0.3, label: 'Tavern', note: 'A seedy place.', color: 'red', created_at: '' }

function mockFetchWith(maps: unknown[], pins: unknown[]) {
  vi.stubGlobal('fetch', vi.fn().mockImplementation((url: string) => {
    if (url.includes('/maps') && !url.includes('/pins')) {
      return Promise.resolve({ ok: true, json: () => Promise.resolve(maps) })
    }
    if (url.includes('/pins')) {
      return Promise.resolve({ ok: true, json: () => Promise.resolve(pins) })
    }
    return Promise.resolve({ ok: true, json: () => Promise.resolve([]) })
  }))
}

describe('MapPanel', () => {
  it('renders nothing when no maps exist', async () => {
    mockFetchWith([], [])
    const { container } = render(<MapPanel campaignId={1} lastEvent={null} />)
    await waitFor(() => {
      expect(container.firstChild).toBeNull()
    })
  })

  it('renders map image when maps exist', async () => {
    mockFetchWith([mockMap], [])
    render(<MapPanel campaignId={1} lastEvent={null} />)
    await waitFor(() => {
      const img = screen.getByRole('img')
      expect(img).toHaveAttribute('src', '/api/files/maps/world.jpg')
      expect(img).toHaveAttribute('alt', 'World Map')
    })
  })

  it('renders pin buttons for each pin', async () => {
    mockFetchWith([mockMap], [mockPin])
    render(<MapPanel campaignId={1} lastEvent={null} />)
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /Tavern/ })).toBeInTheDocument()
    })
  })

  it('shows popover when pin is clicked', async () => {
    mockFetchWith([mockMap], [mockPin])
    render(<MapPanel campaignId={1} lastEvent={null} />)
    await waitFor(() => screen.getByRole('button', { name: /Tavern/ }))
    fireEvent.click(screen.getByRole('button', { name: /Tavern/ }))
    expect(screen.getByText('A seedy place.')).toBeInTheDocument()
  })

  it('closes popover when same pin is clicked again', async () => {
    mockFetchWith([mockMap], [mockPin])
    render(<MapPanel campaignId={1} lastEvent={null} />)
    await waitFor(() => screen.getByRole('button', { name: /Tavern/ }))
    const btn = screen.getByRole('button', { name: /Tavern/ })
    fireEvent.click(btn)
    expect(screen.getByText('A seedy place.')).toBeInTheDocument()
    fireEvent.click(btn)
    expect(screen.queryByText('A seedy place.')).not.toBeInTheDocument()
  })
})
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd web && npx vitest run src/MapPanel.test.tsx 2>&1 | tail -20
```
Expected: FAIL — module not found

- [ ] **Step 3: Create `web/src/MapPanel.tsx`**

```tsx
import { useState, useEffect } from 'react'
import { fetchMaps, fetchMapPins } from './api'
import type { CampaignMap, MapPin } from './types'

interface WsEvent {
  type: string
  payload: Record<string, unknown>
}

interface Props {
  campaignId: number
  lastEvent: WsEvent | null
}

export function MapPanel({ campaignId, lastEvent }: Props) {
  const [maps, setMaps] = useState<CampaignMap[]>([])
  const [pins, setPins] = useState<MapPin[]>([])
  const [activePinId, setActivePinId] = useState<number | null>(null)

  const currentMap = maps[0] ?? null

  useEffect(() => {
    let ignored = false
    fetchMaps(campaignId)
      .then((data) => { if (!ignored) setMaps(data) })
      .catch(() => { if (!ignored) setMaps([]) })
    return () => { ignored = true }
  }, [campaignId])

  useEffect(() => {
    if (!currentMap) return
    let ignored = false
    fetchMapPins(currentMap.id)
      .then((data) => { if (!ignored) setPins(data) })
      .catch(() => { if (!ignored) setPins([]) })
    return () => { ignored = true }
  }, [currentMap?.id])

  useEffect(() => {
    if (!lastEvent || lastEvent.type !== 'map_pin_added') return
    if (!currentMap) return
    const payload = lastEvent.payload as { map_id?: number }
    if (payload.map_id !== currentMap.id) return
    fetchMapPins(currentMap.id).then(setPins).catch(() => {})
  }, [lastEvent, currentMap?.id])

  if (!currentMap) return null

  const imageUrl = `/api/files/${currentMap.image_path}`

  function handlePinClick(pinId: number) {
    setActivePinId((prev) => (prev === pinId ? null : pinId))
  }

  function handleUpload(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    const name = prompt('Map name:', file.name.replace(/\.[^.]+$/, '')) ?? file.name
    const form = new FormData()
    form.append('name', name)
    form.append('image', file)
    fetch(`/api/campaigns/${campaignId}/maps`, { method: 'POST', body: form })
      .then((r) => r.json())
      .then((m: CampaignMap) => setMaps((prev) => [m, ...prev]))
      .catch(() => {})
  }

  return (
    <section className="panel map-panel">
      <div className="map-header">
        <h2>{currentMap.name}</h2>
        <label className="map-upload-btn">
          Upload Map
          <input type="file" accept="image/*" onChange={handleUpload} hidden />
        </label>
      </div>
      <div className="map-container">
        <img src={imageUrl} alt={currentMap.name} className="map-image" />
        {pins.map((pin) => (
          <button
            key={pin.id}
            className="map-pin"
            style={{ left: `${pin.x * 100}%`, top: `${pin.y * 100}%` }}
            onClick={() => handlePinClick(pin.id)}
            aria-label={pin.label}
          >
            ●
            {activePinId === pin.id && (
              <div className="pin-popover">
                <strong>{pin.label}</strong>
                {pin.note && <p>{pin.note}</p>}
              </div>
            )}
          </button>
        ))}
      </div>
    </section>
  )
}
```

- [ ] **Step 4: Run tests**

```bash
cd web && npx vitest run src/MapPanel.test.tsx 2>&1 | tail -20
```
Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add web/src/MapPanel.tsx web/src/MapPanel.test.tsx
git commit -m "feat: MapPanel with map image, pin overlay, and popover"
```

---

### Task 9: JournalPanel + WorldNotesPanel draft button

**Files:**
- Create: `web/src/JournalPanel.tsx`
- Create: `web/src/JournalPanel.test.tsx`
- Modify: `web/src/WorldNotesPanel.tsx`
- Modify: `web/src/WorldNotesPanel.test.tsx`

- [ ] **Step 1: Write failing tests for JournalPanel**

Create `web/src/JournalPanel.test.tsx`:

```tsx
import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { JournalPanel } from './JournalPanel'
import type { Session } from './types'

afterEach(() => vi.restoreAllMocks())

const mockSession: Session = {
  id: 1, campaign_id: 1, title: 'S1', date: '2026-04-03', summary: 'Initial summary.', created_at: ''
}

describe('JournalPanel', () => {
  it('renders session summary in textarea', () => {
    render(<JournalPanel session={mockSession} lastEvent={null} />)
    const textarea = screen.getByRole('textbox')
    expect(textarea).toHaveValue('Initial summary.')
  })

  it('sends PATCH on blur', async () => {
    const mockFetch = vi.fn().mockResolvedValue({ ok: true })
    vi.stubGlobal('fetch', mockFetch)
    render(<JournalPanel session={mockSession} lastEvent={null} />)
    const textarea = screen.getByRole('textbox')
    fireEvent.change(textarea, { target: { value: 'Updated.' } })
    fireEvent.blur(textarea)
    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith('/api/sessions/1', expect.objectContaining({ method: 'PATCH' }))
    })
  })

  it('updates summary from session_updated WS event', () => {
    const { rerender } = render(<JournalPanel session={mockSession} lastEvent={null} />)
    const wsEvent = {
      type: 'session_updated',
      payload: { session_id: 1, summary: 'AI generated recap.' }
    }
    rerender(<JournalPanel session={mockSession} lastEvent={wsEvent} />)
    expect(screen.getByRole('textbox')).toHaveValue('AI generated recap.')
  })

  it('renders Generate Recap button', () => {
    render(<JournalPanel session={mockSession} lastEvent={null} />)
    expect(screen.getByRole('button', { name: /Generate Recap/i })).toBeInTheDocument()
  })

  it('calls recap endpoint when button is clicked', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true, json: () => Promise.resolve({ summary: 'AI recap.' })
    })
    vi.stubGlobal('fetch', mockFetch)
    render(<JournalPanel session={mockSession} lastEvent={null} />)
    fireEvent.click(screen.getByRole('button', { name: /Generate Recap/i }))
    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith('/api/sessions/1/recap', expect.objectContaining({ method: 'POST' }))
    })
  })
})
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd web && npx vitest run src/JournalPanel.test.tsx 2>&1 | tail -20
```
Expected: FAIL — module not found

- [ ] **Step 3: Create `web/src/JournalPanel.tsx`**

```tsx
import { useState, useEffect } from 'react'
import { patchSessionSummary, generateRecap } from './api'
import type { Session } from './types'

interface WsEvent {
  type: string
  payload: Record<string, unknown>
}

interface Props {
  session: Session
  lastEvent: WsEvent | null
}

export function JournalPanel({ session, lastEvent }: Props) {
  const [text, setText] = useState(session.summary)
  const [generating, setGenerating] = useState(false)

  // Update text when a session_updated WS event arrives for this session
  useEffect(() => {
    if (!lastEvent || lastEvent.type !== 'session_updated') return
    const payload = lastEvent.payload as { session_id?: number; summary?: string }
    if (payload.session_id !== session.id || payload.summary === undefined) return
    setText(payload.summary)
  }, [lastEvent, session.id])

  function handleBlur() {
    patchSessionSummary(session.id, text).catch(() => {})
  }

  async function handleGenerateRecap() {
    setGenerating(true)
    try {
      const result = await generateRecap(session.id)
      setText(result.summary)
    } catch {
      // silently fail — WS event will update if recap succeeds on server
    } finally {
      setGenerating(false)
    }
  }

  return (
    <section className="panel journal-panel">
      <div className="journal-header">
        <h2>Session Journal</h2>
        <button
          className="recap-btn"
          onClick={handleGenerateRecap}
          disabled={generating}
        >
          {generating ? 'Generating…' : 'Generate Recap'}
        </button>
      </div>
      <textarea
        className="journal-textarea"
        value={text}
        onChange={(e) => setText(e.target.value)}
        onBlur={handleBlur}
        placeholder="Session notes and recap…"
        rows={6}
      />
    </section>
  )
}
```

- [ ] **Step 4: Run JournalPanel tests**

```bash
cd web && npx vitest run src/JournalPanel.test.tsx 2>&1 | tail -20
```
Expected: all PASS

- [ ] **Step 5: Write failing test for WorldNotesPanel draft button**

In `web/src/WorldNotesPanel.test.tsx` (check current content — it may already exist from Plan 6), add or append:

```tsx
// If the file doesn't exist yet, create it. Otherwise add to the existing describe block.
import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { WorldNotesPanel } from './WorldNotesPanel'

afterEach(() => vi.restoreAllMocks())

function stubFetch(notes: unknown[] = []) {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve(notes) }))
}

describe('WorldNotesPanel draft', () => {
  it('renders Draft with Claude button', async () => {
    stubFetch([])
    render(<WorldNotesPanel campaignId={1} lastEvent={null} />)
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /Draft with Claude/i })).toBeInTheDocument()
    })
  })

  it('shows hint input when button is clicked', async () => {
    stubFetch([])
    render(<WorldNotesPanel campaignId={1} lastEvent={null} />)
    await waitFor(() => screen.getByRole('button', { name: /Draft with Claude/i }))
    fireEvent.click(screen.getByRole('button', { name: /Draft with Claude/i }))
    expect(screen.getByPlaceholderText(/hint/i)).toBeInTheDocument()
  })

  it('calls draft endpoint on submit', async () => {
    const newNote = { id: 99, campaign_id: 1, title: 'Zara', content: 'A smith.', category: 'npc', tags_json: '[]', created_at: '' }
    const mockFetch = vi.fn()
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([]) }) // initial fetch
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve(newNote) }) // draft
    vi.stubGlobal('fetch', mockFetch)
    render(<WorldNotesPanel campaignId={1} lastEvent={null} />)
    await waitFor(() => screen.getByRole('button', { name: /Draft with Claude/i }))
    fireEvent.click(screen.getByRole('button', { name: /Draft with Claude/i }))
    const input = screen.getByPlaceholderText(/hint/i)
    fireEvent.change(input, { target: { value: 'Dwarven smith' } })
    fireEvent.click(screen.getByRole('button', { name: /^Generate$/i }))
    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/campaigns/1/world-notes/draft',
        expect.objectContaining({ method: 'POST' })
      )
    })
  })
})
```

- [ ] **Step 6: Run WorldNotesPanel draft tests to verify they fail**

```bash
cd web && npx vitest run src/WorldNotesPanel.test.tsx 2>&1 | tail -20
```
Expected: FAIL — buttons not found (not yet in component)

- [ ] **Step 7: Update `web/src/WorldNotesPanel.tsx` to add draft UI**

The current component (after Plan 6 which adds `lastEvent` prop) looks roughly like this. Update it to add the draft button:

```tsx
import { useState, useEffect } from 'react'
import { fetchWorldNotes, draftWorldNote } from './api'
import type { WorldNote } from './types'

interface WsEvent {
  type: string
  payload: Record<string, unknown>
}

interface Props {
  campaignId: number
  lastEvent: WsEvent | null
}

export function WorldNotesPanel({ campaignId, lastEvent }: Props) {
  const [notes, setNotes] = useState<WorldNote[]>([])
  const [query, setQuery] = useState('')
  const [showDraft, setShowDraft] = useState(false)
  const [hint, setHint] = useState('')
  const [drafting, setDrafting] = useState(false)
  const [draftError, setDraftError] = useState<string | null>(null)

  useEffect(() => {
    let ignored = false
    fetchWorldNotes(campaignId, query || undefined)
      .then((data) => { if (!ignored) setNotes(data) })
      .catch(() => { if (!ignored) setNotes([]) })
    return () => { ignored = true }
  }, [campaignId, query])

  // Refresh notes on world_note_created or world_note_updated WS events
  useEffect(() => {
    if (!lastEvent) return
    if (lastEvent.type !== 'world_note_created' && lastEvent.type !== 'world_note_updated') return
    fetchWorldNotes(campaignId, query || undefined)
      .then(setNotes)
      .catch(() => {})
  }, [lastEvent, campaignId, query])

  async function handleDraft() {
    if (!hint.trim()) return
    setDrafting(true)
    setDraftError(null)
    try {
      await draftWorldNote(campaignId, hint.trim())
      // Notes will refresh via world_note_created WS event
      setShowDraft(false)
      setHint('')
    } catch (err) {
      setDraftError(err instanceof Error ? err.message : 'Draft failed')
    } finally {
      setDrafting(false)
    }
  }

  return (
    <section className="panel world-notes">
      <div className="world-notes-header">
        <h2>World Notes</h2>
        <button className="draft-btn" onClick={() => { setShowDraft((v) => !v); setDraftError(null) }}>
          Draft with Claude
        </button>
      </div>

      {showDraft && (
        <div className="draft-form">
          <input
            type="text"
            placeholder="Hint (e.g. dwarven blacksmith NPC)"
            value={hint}
            onChange={(e) => setHint(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleDraft()}
          />
          <button onClick={handleDraft} disabled={drafting}>
            {drafting ? 'Generating…' : 'Generate'}
          </button>
          {draftError && <p className="draft-error">{draftError}</p>}
        </div>
      )}

      <input
        className="search"
        type="search"
        placeholder="Search notes…"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
      />

      {notes.length === 0 ? (
        <p className="empty">No notes found.</p>
      ) : (
        notes.map((n) => (
          <div key={n.id} className="world-note">
            <div className="note-header">
              <span className="note-title">{n.title}</span>
              {n.category && <span className="note-category">{n.category}</span>}
            </div>
            <p className="note-content">{n.content}</p>
          </div>
        ))
      )}
    </section>
  )
}
```

**Note:** If Plan 6 already modified `WorldNotesPanel.tsx` to have a different structure, merge the draft additions into the existing component rather than replacing it wholesale. The key additions are: `showDraft`, `hint`, `drafting`, `draftError` state; `handleDraft` function; the draft button in the header; and the draft form below the header. Keep whatever Plan 6 already added.

- [ ] **Step 8: Run all frontend tests**

```bash
cd web && npx vitest run 2>&1 | tail -30
```
Expected: all PASS

- [ ] **Step 9: Commit**

```bash
git add web/src/JournalPanel.tsx web/src/JournalPanel.test.tsx web/src/WorldNotesPanel.tsx web/src/WorldNotesPanel.test.tsx
git commit -m "feat: JournalPanel with auto-save and recap; WorldNotesPanel Draft with Claude"
```

---

### Task 10: App wiring + CSS

**Files:**
- Modify: `web/src/App.tsx`
- Modify: `web/src/App.test.tsx`
- Modify: `web/src/App.css`

- [ ] **Step 1: Update `web/src/App.tsx`**

Import and add MapPanel and JournalPanel. The file after Plan 7 wiring looks like:

```tsx
import { useState, useEffect, useCallback } from 'react'
import { useWebSocket } from './useWebSocket'
import { fetchContext } from './api'
import type { GameContext, Message } from './types'
import { WorldNotesPanel } from './WorldNotesPanel'
import { DiceHistoryPanel } from './DiceHistoryPanel'
import { CombatPanel } from './CombatPanel'           // added by Plan 7
import { SessionTimeline } from './SessionTimeline'   // added by Plan 7
import { MapPanel } from './MapPanel'                  // NEW
import { JournalPanel } from './JournalPanel'          // NEW
import './App.css'

const WS_URL = `ws://${window.location.host}/ws`

export default function App() {
  const [ctx, setCtx] = useState<GameContext | null>(null)
  const [error, setError] = useState<string | null>(null)

  const loadContext = useCallback(() => {
    fetchContext()
      .then(setCtx)
      .catch(() => setError('Could not load game state'))
  }, [])

  useEffect(() => { loadContext() }, [loadContext])

  const { lastEvent } = useWebSocket(WS_URL, loadContext)

  if (error) return <div className="error">{error}</div>
  if (!ctx) return <div className="loading">Loading…</div>

  return (
    <div className="dashboard">
      <header className="state-bar">
        <span className="campaign">{ctx.campaign?.name ?? 'No campaign'}</span>
        <span className="separator">·</span>
        <span className="character">{ctx.character?.name ?? 'No character'}</span>
        <span className="separator">·</span>
        <span className="session">{ctx.session?.title ?? 'No session'}</span>
      </header>

      <main className="panels">
        <section className="panel messages">
          <h2>Session Log</h2>
          {(ctx.recent_messages ?? []).length === 0 ? (
            <p className="empty">No messages yet.</p>
          ) : (
            (ctx.recent_messages ?? []).map((m) => (
              <div key={m.id} className={`message ${m.role}`}>
                <span className="role">{m.role}</span>
                <span className="content">{m.content}</span>
              </div>
            ))
          )}
        </section>

        {ctx.active_combat && <CombatPanel combat={ctx.active_combat} />}

        {ctx.campaign && (
          <MapPanel campaignId={ctx.campaign.id} lastEvent={lastEvent} />
        )}

        {ctx.session && (
          <JournalPanel session={ctx.session} lastEvent={lastEvent} />
        )}

        {ctx.session && (
          <SessionTimeline sessionId={ctx.session.id} lastEvent={lastEvent} />
        )}

        {ctx.campaign && (
          <WorldNotesPanel campaignId={ctx.campaign.id} lastEvent={lastEvent} />
        )}

        {ctx.session && (
          <DiceHistoryPanel sessionId={ctx.session.id} lastEvent={lastEvent} />
        )}
      </main>
    </div>
  )
}
```

**Note:** If the existing `App.tsx` from Plans 6 and 7 already has a slightly different structure, merge the MapPanel and JournalPanel additions into the existing component. Do not remove anything Plans 6/7 added.

- [ ] **Step 2: Update `web/src/App.test.tsx`**

The App test mocks fetch for `/api/context`. MapPanel and JournalPanel also call fetch. Add their URLs to the mock:

```tsx
// In the existing fetch mock setup, add handlers for /maps and /recap:
vi.stubGlobal('fetch', vi.fn().mockImplementation((url: string) => {
  if (url === '/api/context') {
    return Promise.resolve({ ok: true, json: () => Promise.resolve(mockCtx) })
  }
  // MapPanel, WorldNotesPanel, DiceHistoryPanel sub-fetches
  return Promise.resolve({ ok: true, json: () => Promise.resolve([]) })
}))
```

Check the existing mock and ensure it returns `[]` for any URL that isn't `/api/context`. The MapPanel adds calls to `/api/campaigns/{id}/maps` and `/api/maps/{id}/pins`; the fetch mock returning `[]` for those is correct (no maps → MapPanel renders null).

- [ ] **Step 3: Run frontend tests**

```bash
cd web && npx vitest run 2>&1 | tail -20
```
Expected: all PASS

- [ ] **Step 4: Add CSS to `web/src/App.css`**

Append to the end of the file:

```css
/* MapPanel */
.map-panel {}

.map-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.map-upload-btn {
  cursor: pointer;
  padding: 0.25rem 0.75rem;
  background: var(--color-accent, #4a6fa5);
  color: white;
  border-radius: 4px;
  font-size: 0.85rem;
}

.map-container {
  position: relative;
  display: inline-block;
  width: 100%;
}

.map-image {
  width: 100%;
  display: block;
  border-radius: 4px;
}

.map-pin {
  position: absolute;
  transform: translate(-50%, -50%);
  background: rgba(220, 60, 60, 0.85);
  color: white;
  border: none;
  border-radius: 50%;
  width: 1.4rem;
  height: 1.4rem;
  cursor: pointer;
  font-size: 0.7rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.pin-popover {
  position: absolute;
  bottom: 130%;
  left: 50%;
  transform: translateX(-50%);
  background: var(--color-surface, #1e1e2e);
  border: 1px solid var(--color-border, #333);
  border-radius: 4px;
  padding: 0.5rem 0.75rem;
  min-width: 140px;
  z-index: 10;
  text-align: left;
  font-size: 0.85rem;
  white-space: nowrap;
}

.pin-popover p {
  margin: 0.25rem 0 0;
  white-space: normal;
}

/* JournalPanel */
.journal-panel {}

.journal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.recap-btn {
  padding: 0.25rem 0.75rem;
  background: var(--color-accent, #4a6fa5);
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.85rem;
}

.recap-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.journal-textarea {
  width: 100%;
  box-sizing: border-box;
  padding: 0.5rem;
  border: 1px solid var(--color-border, #333);
  border-radius: 4px;
  background: var(--color-bg, #13131a);
  color: inherit;
  font-family: inherit;
  font-size: 0.9rem;
  resize: vertical;
}

/* WorldNotesPanel draft UI */
.world-notes-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.draft-btn {
  padding: 0.25rem 0.75rem;
  background: var(--color-accent, #4a6fa5);
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.85rem;
}

.draft-form {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
  flex-wrap: wrap;
}

.draft-form input {
  flex: 1;
  min-width: 0;
  padding: 0.35rem 0.5rem;
  border: 1px solid var(--color-border, #333);
  border-radius: 4px;
  background: var(--color-bg, #13131a);
  color: inherit;
  font-size: 0.9rem;
}

.draft-form button {
  padding: 0.35rem 0.75rem;
  background: var(--color-accent, #4a6fa5);
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.85rem;
}

.draft-form button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.draft-error {
  width: 100%;
  color: var(--color-error, #f87171);
  font-size: 0.85rem;
  margin: 0;
}
```

- [ ] **Step 5: Build to verify no TypeScript errors**

```bash
cd web && npm run build 2>&1 | tail -20
```
Expected: build succeeds with no type errors

- [ ] **Step 6: Run all Go tests**

```bash
go test ./... 2>&1 | grep -E "^(ok|FAIL|---)"
```
Expected: all `ok`

- [ ] **Step 7: Commit**

```bash
git add web/src/App.tsx web/src/App.test.tsx web/src/App.css
git commit -m "feat: wire MapPanel and JournalPanel into App; add CSS"
```

---

## Self-Review

**Spec coverage check:**

| Spec requirement | Covered |
|---|---|
| MapPanel: renders when map exists for campaign | Task 8 |
| MapPanel: image at `/api/files/maps/{filename}` | Task 8 |
| MapPanel: pins at (x%, y%) absolute positioning | Task 8 |
| MapPanel: clicking pin opens popover | Task 8 |
| MapPanel: `map_pin_added` WS event adds pins | Task 8 |
| MapPanel: map upload via file picker | Task 8 |
| `GET /api/campaigns/{id}/maps` | Task 4 |
| `POST /api/campaigns/{id}/maps` | Task 4 |
| `GET /api/maps/{id}` | Task 4 |
| `GET /api/files/{path}` with traversal protection | Task 3 |
| JournalPanel: textarea bound to `session.summary` | Task 9 |
| JournalPanel: auto-save on blur via PATCH | Task 9 |
| JournalPanel: "Generate recap" button | Task 9 |
| `PATCH /api/sessions/{id}` | Task 4 |
| `POST /api/sessions/{id}/recap` | Task 5 |
| `session_updated` WS event fires on recap | Tasks 5, 9 |
| AI: "Draft with Claude" in WorldNotesPanel | Task 9 |
| `POST /api/campaigns/{id}/world-notes/draft` | Task 5 |
| `world_note_created` WS event fires on draft | Task 5 |
| AI disabled when no key set: 503 response | Task 5 |
| `generate_session_recap` MCP tool | Task 6 |
| `EventSessionUpdated` event type | Task 1 |
| File storage: uploaded files saved to `~/.ttrpg/maps/` | Tasks 3, 4 |

**Gaps found:** None — all spec requirements have a corresponding task.

**Placeholder scan:** No TBDs, no "implement later", no vague steps — all steps contain actual code.

**Type consistency:** 
- `CampaignMap` (TS) maps to `db.Map` (Go) — field names match JSON tags.
- `ai.Completer` used consistently in `api.Server`, `mcp.Server`, and tests.
- `lastEvent` prop type `WsEvent | null` is inline in each component — consistent shape `{ type: string, payload: Record<string, unknown> }`.
- `EventSessionUpdated` added in Task 1, used in Tasks 4, 5, 6.
