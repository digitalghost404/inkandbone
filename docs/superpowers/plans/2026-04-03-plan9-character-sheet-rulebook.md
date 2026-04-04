# Plan 9: Character Sheet + Rulebook Ingestion Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an editable character sheet panel driven by the ruleset's schema_json, rulebook ingestion (paste and PDF), and a `search_rulebook` MCP tool that lets Claude look up rules on demand.

**Architecture:** Three new DB methods (`GetRuleset`, `UpdateCharacterPortrait`, rulebook CRUD), four new HTTP routes (GET ruleset, PATCH character, POST portrait, POST rulebook), one MCP tool (`search_rulebook`), and a `CharacterSheetPanel` React component that debounces PATCH calls and reacts to `character_updated` WS events. Rulebook text is chunked on Markdown headings and stored in a new `rulebook_chunks` table; search is SQLite LIKE with LIMIT 3. PDF text extraction uses `github.com/ledongthuc/pdf` (pure Go, no system deps), which provides a clean `GetPlainText()` API that the spec's "pdfcpu" intent is satisfied by.

**Tech Stack:** Go 1.22+ `net/http`, SQLite via `modernc.org/sqlite`, `github.com/ledongthuc/pdf` (new dep), React 18, TypeScript, Vitest, `@testing-library/react`

**Context:** This is Plan 9 in the inkandbone series. Plans 6–8 have already been executed. The server signature is `api.NewServer(db, dataDir, aiClient)`, MCP server is `mcp.New(db, bus, aiClient)`, and `api/server.go` has a `registerRoutes()` method with Plan 8's routes already registered. The `randomHex` helper and `os`/`filepath`/`io` imports already exist in `internal/api/routes.go` from Plan 8. App.tsx uses `const { lastEvent } = useWebSocket(WS_URL)` (Plan 6 changed the hook) and passes `lastEvent` to panels. `EventCharacterUpdated` already exists in `internal/api/events.go`.

---

## File Map

| Action | File | Responsibility |
|--------|------|----------------|
| Create | `internal/db/migrations/003_rulebook_chunks.sql` | `rulebook_chunks` table |
| Modify | `internal/db/queries_core.go` | Add `GetRuleset`, `UpdateCharacterPortrait` |
| Create | `internal/db/queries_rulebook.go` | `RulebookChunk` struct + 3 DB methods |
| Create | `internal/db/queries_rulebook_test.go` | Tests for rulebook DB methods |
| Modify | `internal/db/queries_core_test.go` | Tests for GetRuleset + UpdateCharacterPortrait |
| Modify | `internal/api/routes.go` | Add `handleGetRuleset`, `handlePatchCharacter`, `handleUploadPortrait` |
| Create | `internal/api/rulebook_ingest.go` | `handleUploadRulebook`, PDF extraction, chunking |
| Modify | `internal/api/server.go` | Register 4 new routes in `registerRoutes()` |
| Modify | `internal/api/routes_test.go` | Tests for new character/ruleset routes |
| Create | `internal/api/rulebook_ingest_test.go` | Tests for rulebook paste ingestion |
| Create | `internal/mcp/rulebook.go` | `handleSearchRulebook` handler |
| Create | `internal/mcp/rulebook_test.go` | MCP tool tests |
| Modify | `internal/mcp/server.go` | Register `search_rulebook` tool |
| Modify | `web/src/types.ts` | Add `Ruleset` interface |
| Modify | `web/src/api.ts` | Add `fetchRuleset`, `patchCharacterData`, `uploadPortrait`, `uploadRulebook` |
| Modify | `web/src/api.test.ts` | Tests for new API functions |
| Create | `web/src/CharacterSheetPanel.tsx` | Character sheet with debounced PATCH + WS sync |
| Create | `web/src/CharacterSheetPanel.test.tsx` | Component render tests |
| Modify | `web/src/App.tsx` | Wire `CharacterSheetPanel` |
| Modify | `web/src/App.css` | Character sheet panel styles |

---

## Task 1: DB — migration + GetRuleset + UpdateCharacterPortrait + rulebook queries

**Files:**
- Create: `internal/db/migrations/003_rulebook_chunks.sql`
- Modify: `internal/db/queries_core.go`
- Create: `internal/db/queries_rulebook.go`
- Create: `internal/db/queries_rulebook_test.go`
- Modify: `internal/db/queries_core_test.go`

- [ ] **Step 1: Write the failing DB tests**

Create `internal/db/queries_rulebook_test.go`:

```go
package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchRulebookChunks(t *testing.T) {
	d := newTestDB(t)
	rs, err := d.GetRulesetByName("dnd5e")
	require.NoError(t, err)
	require.NoError(t, d.InsertRulebookChunk(rs.ID, "Spells", "Fireball deals 8d6 fire damage in a 20-foot radius."))
	require.NoError(t, d.InsertRulebookChunk(rs.ID, "Conditions", "Blinded: the creature cannot see."))

	results, err := d.SearchRulebookChunks(rs.ID, "fireball")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Spells", results[0].Heading)
	assert.Contains(t, results[0].Content, "Fireball")
}

func TestSearchRulebookChunks_noMatch(t *testing.T) {
	d := newTestDB(t)
	rs, _ := d.GetRulesetByName("dnd5e")
	require.NoError(t, d.InsertRulebookChunk(rs.ID, "Spells", "Fireball"))

	results, err := d.SearchRulebookChunks(rs.ID, "xyzzy")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestSearchRulebookChunks_limitThree(t *testing.T) {
	d := newTestDB(t)
	rs, _ := d.GetRulesetByName("dnd5e")
	for i := 0; i < 5; i++ {
		require.NoError(t, d.InsertRulebookChunk(rs.ID, "Spells", "Fireball variant"))
	}
	results, err := d.SearchRulebookChunks(rs.ID, "fireball")
	require.NoError(t, err)
	assert.Len(t, results, 3)
}

func TestDeleteRulebookChunks(t *testing.T) {
	d := newTestDB(t)
	rs, _ := d.GetRulesetByName("dnd5e")
	require.NoError(t, d.InsertRulebookChunk(rs.ID, "Spells", "Fireball"))
	require.NoError(t, d.DeleteRulebookChunks(rs.ID))
	results, _ := d.SearchRulebookChunks(rs.ID, "Fireball")
	assert.Empty(t, results)
}
```

Append to `internal/db/queries_core_test.go`:

```go
func TestGetRuleset(t *testing.T) {
	d := newTestDB(t)
	rs, err := d.GetRulesetByName("dnd5e")
	require.NoError(t, err)
	require.NotNil(t, rs)

	got, err := d.GetRuleset(rs.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "dnd5e", got.Name)

	missing, err := d.GetRuleset(99999)
	require.NoError(t, err)
	assert.Nil(t, missing)
}

func TestUpdateCharacterPortrait(t *testing.T) {
	d := newTestDB(t)
	rs, _ := d.GetRulesetByName("ironsworn")
	campID, _ := d.CreateCampaign(rs.ID, "C", "")
	charID, _ := d.CreateCharacter(campID, "Hero")

	require.NoError(t, d.UpdateCharacterPortrait(charID, "portraits/abc.jpg"))
	ch, _ := d.GetCharacter(charID)
	assert.Equal(t, "portraits/abc.jpg", ch.PortraitPath)

	err := d.UpdateCharacterPortrait(99999, "x.jpg")
	assert.ErrorContains(t, err, "not found")
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd /home/digitalghost/projects/inkandbone
go test ./internal/db/... 2>&1 | tail -20
```

Expected: FAIL — `InsertRulebookChunk`, `SearchRulebookChunks`, `DeleteRulebookChunks`, `GetRuleset`, `UpdateCharacterPortrait` undefined.

- [ ] **Step 3: Create the migration**

Create `internal/db/migrations/003_rulebook_chunks.sql`:

```sql
CREATE TABLE IF NOT EXISTS rulebook_chunks (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  ruleset_id INTEGER NOT NULL REFERENCES rulesets(id),
  heading    TEXT NOT NULL,
  content    TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

- [ ] **Step 4: Add GetRuleset and UpdateCharacterPortrait to queries_core.go**

Append to `internal/db/queries_core.go` (after `ListCharacters`):

```go
func (d *DB) GetRuleset(id int64) (*Ruleset, error) {
	r := &Ruleset{}
	err := d.db.QueryRow(
		"SELECT id, name, schema_json, version FROM rulesets WHERE id = ?", id,
	).Scan(&r.ID, &r.Name, &r.SchemaJSON, &r.Version)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return r, err
}

func (d *DB) UpdateCharacterPortrait(id int64, portraitPath string) error {
	res, err := d.db.Exec("UPDATE characters SET portrait_path = ? WHERE id = ?", portraitPath, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("character %d not found", id)
	}
	return nil
}
```

- [ ] **Step 5: Create queries_rulebook.go**

Create `internal/db/queries_rulebook.go`:

```go
package db

// RulebookChunk is one section of an ingested rulebook, split on heading lines.
type RulebookChunk struct {
	ID        int64  `json:"id"`
	RulesetID int64  `json:"ruleset_id"`
	Heading   string `json:"heading"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// DeleteRulebookChunks removes all existing chunks for a ruleset (called before re-ingestion).
func (d *DB) DeleteRulebookChunks(rulesetID int64) error {
	_, err := d.db.Exec("DELETE FROM rulebook_chunks WHERE ruleset_id = ?", rulesetID)
	return err
}

// InsertRulebookChunk saves one text chunk for a ruleset.
func (d *DB) InsertRulebookChunk(rulesetID int64, heading, content string) error {
	_, err := d.db.Exec(
		"INSERT INTO rulebook_chunks (ruleset_id, heading, content) VALUES (?, ?, ?)",
		rulesetID, heading, content,
	)
	return err
}

// SearchRulebookChunks returns up to 3 chunks whose heading or content contains query.
func (d *DB) SearchRulebookChunks(rulesetID int64, query string) ([]RulebookChunk, error) {
	like := "%" + query + "%"
	rows, err := d.db.Query(
		`SELECT id, ruleset_id, heading, content, created_at FROM rulebook_chunks
		 WHERE ruleset_id = ? AND (heading LIKE ? OR content LIKE ?) LIMIT 3`,
		rulesetID, like, like,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RulebookChunk
	for rows.Next() {
		var c RulebookChunk
		if err := rows.Scan(&c.ID, &c.RulesetID, &c.Heading, &c.Content, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
```

- [ ] **Step 6: Run tests to verify they pass**

```bash
go test ./internal/db/... -v 2>&1 | tail -30
```

Expected: all DB tests PASS including the new `TestSearchRulebookChunks*`, `TestDeleteRulebookChunks`, `TestGetRuleset`, `TestUpdateCharacterPortrait`.

- [ ] **Step 7: Commit**

```bash
git add internal/db/migrations/003_rulebook_chunks.sql \
        internal/db/queries_core.go \
        internal/db/queries_core_test.go \
        internal/db/queries_rulebook.go \
        internal/db/queries_rulebook_test.go
git commit -m "feat: rulebook_chunks migration + GetRuleset/UpdateCharacterPortrait/rulebook DB queries"
```

---

## Task 2: API — GET ruleset + PATCH character + POST portrait routes

**Files:**
- Modify: `internal/api/routes.go`
- Modify: `internal/api/server.go`
- Modify: `internal/api/routes_test.go`

- [ ] **Step 1: Write the failing route tests**

Append to `internal/api/routes_test.go`:

```go
func TestGetRuleset_found(t *testing.T) {
	s := newTestServer(t)
	rs, err := s.db.GetRulesetByName("dnd5e")
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodGet, "/api/rulesets/"+strconv.FormatInt(rs.ID, 10), nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got db.Ruleset
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, "dnd5e", got.Name)
}

func TestGetRuleset_notFound(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/rulesets/99999", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestPatchCharacter(t *testing.T) {
	s := newTestServer(t)
	campID, _ := seedCampaign(t, s.db)
	charID, err := s.db.CreateCharacter(campID, "Hero")
	require.NoError(t, err)

	body := `{"data_json":"{\"hp\":25}"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/characters/"+strconv.FormatInt(charID, 10),
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got db.Character
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, `{"hp":25}`, got.DataJSON)
	ch, _ := s.db.GetCharacter(charID)
	assert.Equal(t, `{"hp":25}`, ch.DataJSON)
}

func TestPatchCharacter_invalidJSON(t *testing.T) {
	s := newTestServer(t)
	campID, _ := seedCampaign(t, s.db)
	charID, _ := s.db.CreateCharacter(campID, "Hero")
	req := httptest.NewRequest(http.MethodPatch, "/api/characters/"+strconv.FormatInt(charID, 10),
		strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
```

`routes_test.go` imports already include `"strings"` — if not, add it.

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/api/... -run 'TestGetRuleset|TestPatchCharacter' -v 2>&1 | tail -20
```

Expected: FAIL — routes not registered.

- [ ] **Step 3: Add handlers to routes.go**

Append to `internal/api/routes.go`:

```go
func (s *Server) handleGetRuleset(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid ruleset id", http.StatusBadRequest)
		return
	}
	rs, err := s.db.GetRuleset(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rs == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, rs)
}

func (s *Server) handlePatchCharacter(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid character id", http.StatusBadRequest)
		return
	}
	var body struct {
		DataJSON string `json:"data_json"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := s.db.UpdateCharacterData(id, body.DataJSON); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ch, err := s.db.GetCharacter(id)
	if err != nil || ch == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	s.bus.Publish(Event{Type: EventCharacterUpdated, Payload: ch})
	writeJSON(w, ch)
}

func (s *Server) handleUploadPortrait(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid character id", http.StatusBadRequest)
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "multipart parse error", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("portrait")
	if err != nil {
		http.Error(w, "missing portrait file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	filename := randomHex(16) + ext
	destDir := filepath.Join(s.dataDir, "portraits")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}
	dest, err := os.Create(filepath.Join(destDir, filename))
	if err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}
	defer dest.Close()
	if _, err := io.Copy(dest, file); err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}

	portraitPath := "portraits/" + filename
	if err := s.db.UpdateCharacterPortrait(id, portraitPath); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ch, _ := s.db.GetCharacter(id)
	s.bus.Publish(Event{Type: EventCharacterUpdated, Payload: ch})
	writeJSON(w, ch)
}
```

Note: `randomHex`, `os`, `filepath`, `io` are already present from Plan 8. If they aren't for any reason, add `"io"`, `"os"`, `"path/filepath"` to the import block.

- [ ] **Step 4: Register routes in server.go**

In `internal/api/server.go`, find `registerRoutes()` and append these 3 lines at the end of the method body (before the closing `}`):

```go
	s.mux.HandleFunc("GET /api/rulesets/{id}", s.handleGetRuleset)
	s.mux.HandleFunc("PATCH /api/characters/{id}", s.handlePatchCharacter)
	s.mux.HandleFunc("POST /api/characters/{id}/portrait", s.handleUploadPortrait)
```

Note: the rulebook ingestion route (`POST /api/rulesets/{id}/rulebook`) is added in Task 3 when that handler is ready.

- [ ] **Step 5: Run tests to verify they pass**

```bash
go test ./internal/api/... -run 'TestGetRuleset|TestPatchCharacter' -v 2>&1 | tail -20
```

Expected: PASS.

- [ ] **Step 6: Run full test suite**

```bash
go test ./... 2>&1 | tail -20
```

Expected: all PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/api/routes.go internal/api/server.go internal/api/routes_test.go
git commit -m "feat: GET /api/rulesets/{id}, PATCH /api/characters/{id}, POST /api/characters/{id}/portrait"
```

---

## Task 3: API — rulebook ingestion (paste + PDF)

**Files:**
- Create: `internal/api/rulebook_ingest.go`
- Create: `internal/api/rulebook_ingest_test.go`
- Modify: `internal/api/server.go`
- Modify: `go.mod`, `go.sum` (via `go get`)

- [ ] **Step 1: Write the failing tests**

Create `internal/api/rulebook_ingest_test.go`:

```go
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

func TestUploadRulebook_paste_chunksText(t *testing.T) {
	s := newTestServer(t)
	rs, err := s.db.GetRulesetByName("dnd5e")
	require.NoError(t, err)

	body := "# Spells\nFireball deals 8d6 fire damage.\n## Combat\nRoll initiative at the start of combat."
	req := httptest.NewRequest(http.MethodPost, "/api/rulesets/"+strconv.FormatInt(rs.ID, 10)+"/rulebook",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]int
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 2, resp["chunks_inserted"])

	chunks, err := s.db.SearchRulebookChunks(rs.ID, "fireball")
	require.NoError(t, err)
	require.Len(t, chunks, 1)
	assert.Equal(t, "Spells", chunks[0].Heading)
}

func TestUploadRulebook_paste_replacesExisting(t *testing.T) {
	s := newTestServer(t)
	rs, _ := s.db.GetRulesetByName("dnd5e")

	// Ingest once
	for _, txt := range []string{"# Old\nOld content."} {
		req := httptest.NewRequest(http.MethodPost, "/api/rulesets/"+strconv.FormatInt(rs.ID, 10)+"/rulebook",
			strings.NewReader(txt))
		req.Header.Set("Content-Type", "text/plain")
		httptest.NewRecorder() // discard
		s.ServeHTTP(httptest.NewRecorder(), req)
	}

	// Ingest again — should replace
	req := httptest.NewRequest(http.MethodPost, "/api/rulesets/"+strconv.FormatInt(rs.ID, 10)+"/rulebook",
		strings.NewReader("# New\nNew content."))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	oldChunks, _ := s.db.SearchRulebookChunks(rs.ID, "Old content")
	assert.Empty(t, oldChunks)
	newChunks, _ := s.db.SearchRulebookChunks(rs.ID, "New content")
	assert.Len(t, newChunks, 1)
}

func TestUploadRulebook_notFound(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/api/rulesets/99999/rulebook",
		strings.NewReader("# Test\nContent"))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestChunkByHeadings(t *testing.T) {
	text := "# Spells\nFireball\n## Weapons\nSword\n### Bows\nLongbow"
	chunks := chunkByHeadings(text)
	require.Len(t, chunks, 3)
	assert.Equal(t, "Spells", chunks[0].heading)
	assert.Contains(t, chunks[0].content, "Fireball")
	assert.Equal(t, "Weapons", chunks[1].heading)
	assert.Equal(t, "Bows", chunks[2].heading)
}

func TestChunkByHeadings_textBeforeFirstHeading(t *testing.T) {
	text := "Preamble text.\n# Chapter\nContent."
	chunks := chunkByHeadings(text)
	require.Len(t, chunks, 2)
	assert.Equal(t, "Introduction", chunks[0].heading)
	assert.Equal(t, "Chapter", chunks[1].heading)
}

func TestChunkByHeadings_skipsEmptyContent(t *testing.T) {
	text := "# Heading\n\n# Non-empty\nContent."
	chunks := chunkByHeadings(text)
	// "Heading" has no content, should be skipped
	require.Len(t, chunks, 1)
	assert.Equal(t, "Non-empty", chunks[0].heading)
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/api/... -run 'TestUploadRulebook|TestChunkByHeadings' -v 2>&1 | tail -20
```

Expected: FAIL — `handleUploadRulebook` and `chunkByHeadings` undefined.

- [ ] **Step 3: Add the ledongthuc/pdf dependency**

```bash
cd /home/digitalghost/projects/inkandbone
go get github.com/ledongthuc/pdf
go mod tidy
```

Expected: `go.mod` and `go.sum` updated with `github.com/ledongthuc/pdf`.

- [ ] **Step 4: Create rulebook_ingest.go**

Create `internal/api/rulebook_ingest.go`:

```go
package api

import (
	"bytes"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/ledongthuc/pdf"
)

func (s *Server) handleUploadRulebook(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid ruleset id", http.StatusBadRequest)
		return
	}
	rs, err := s.db.GetRuleset(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rs == nil {
		http.Error(w, "ruleset not found", http.StatusNotFound)
		return
	}

	var text string
	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "text/plain") {
		body, err := io.ReadAll(io.LimitReader(r.Body, 10<<20))
		if err != nil {
			http.Error(w, "read error", http.StatusBadRequest)
			return
		}
		text = string(body)
	} else {
		if err := r.ParseMultipartForm(50 << 20); err != nil {
			http.Error(w, "multipart parse error", http.StatusBadRequest)
			return
		}
		file, _, err := r.FormFile("rulebook")
		if err != nil {
			http.Error(w, "missing rulebook file", http.StatusBadRequest)
			return
		}
		defer file.Close()
		data, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "read error", http.StatusBadRequest)
			return
		}
		text, err = extractPDFText(data)
		if err != nil {
			http.Error(w, "pdf extraction failed: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	chunks := chunkByHeadings(text)
	if err := s.db.DeleteRulebookChunks(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	inserted := 0
	for _, c := range chunks {
		if err := s.db.InsertRulebookChunk(id, c.heading, c.content); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		inserted++
	}
	writeJSON(w, map[string]int{"chunks_inserted": inserted})
}

func extractPDFText(data []byte) (string, error) {
	r, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	for i := 1; i <= r.NumPage(); i++ {
		page := r.Page(i)
		if page.V.IsNull() {
			continue
		}
		content, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		sb.WriteString(content)
		sb.WriteString("\n")
	}
	return sb.String(), nil
}

type textChunk struct {
	heading string
	content string
}

func chunkByHeadings(text string) []textChunk {
	headingRe := regexp.MustCompile(`^#{1,3}\s+(.+)`)
	var chunks []textChunk
	var currentHeading string
	var currentContent strings.Builder

	flush := func() {
		content := strings.TrimSpace(currentContent.String())
		if content != "" {
			heading := currentHeading
			if heading == "" {
				heading = "Introduction"
			}
			chunks = append(chunks, textChunk{heading: heading, content: content})
		}
		currentContent.Reset()
	}

	for _, line := range strings.Split(text, "\n") {
		if m := headingRe.FindStringSubmatch(line); m != nil {
			flush()
			currentHeading = m[1]
		} else {
			currentContent.WriteString(line)
			currentContent.WriteString("\n")
		}
	}
	flush()
	return chunks
}
```

- [ ] **Step 5: Register route in server.go**

In `internal/api/server.go`, append to `registerRoutes()`:

```go
	s.mux.HandleFunc("POST /api/rulesets/{id}/rulebook", s.handleUploadRulebook)
```

- [ ] **Step 6: Run tests to verify they pass**

```bash
go test ./internal/api/... -run 'TestUploadRulebook|TestChunkByHeadings' -v 2>&1 | tail -20
```

Expected: PASS.

- [ ] **Step 7: Run full test suite**

```bash
go test ./... 2>&1 | tail -20
```

Expected: all PASS.

- [ ] **Step 8: Commit**

```bash
git add internal/api/rulebook_ingest.go internal/api/rulebook_ingest_test.go \
        internal/api/server.go go.mod go.sum
git commit -m "feat: POST /api/rulesets/{id}/rulebook — paste and PDF ingestion, heading-based chunking"
```

---

## Task 4: MCP — search_rulebook tool

**Files:**
- Create: `internal/mcp/rulebook.go`
- Create: `internal/mcp/rulebook_test.go`
- Modify: `internal/mcp/server.go`

- [ ] **Step 1: Write the failing MCP tests**

Create `internal/mcp/rulebook_test.go`:

```go
package mcp

import (
	"context"
	"testing"

	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchRulebook_noResults(t *testing.T) {
	s := newTestMCP(t)
	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"ruleset_id": float64(1),
		"query":      "xyzzy",
	}
	result, err := s.handleSearchRulebook(context.Background(), req)
	require.NoError(t, err)
	require.False(t, result.IsError)
	tc, ok := result.Content[0].(mcplib.TextContent)
	require.True(t, ok)
	assert.Contains(t, tc.Text, "No matching")
}

func TestSearchRulebook_withResults(t *testing.T) {
	s := newTestMCP(t)
	rs, err := s.db.GetRulesetByName("dnd5e")
	require.NoError(t, err)
	require.NoError(t, s.db.InsertRulebookChunk(rs.ID, "Spells", "Fireball deals 8d6 fire damage in a 20-foot radius."))

	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"ruleset_id": float64(rs.ID),
		"query":      "fireball",
	}
	result, err := s.handleSearchRulebook(context.Background(), req)
	require.NoError(t, err)
	require.False(t, result.IsError)
	tc, ok := result.Content[0].(mcplib.TextContent)
	require.True(t, ok)
	assert.Contains(t, tc.Text, "Spells")
	assert.Contains(t, tc.Text, "Fireball")
}

func TestSearchRulebook_missingRulesetID(t *testing.T) {
	s := newTestMCP(t)
	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{"query": "fireball"}
	result, err := s.handleSearchRulebook(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestSearchRulebook_missingQuery(t *testing.T) {
	s := newTestMCP(t)
	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{"ruleset_id": float64(1)}
	result, err := s.handleSearchRulebook(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/mcp/... -run 'TestSearchRulebook' -v 2>&1 | tail -20
```

Expected: FAIL — `handleSearchRulebook` undefined.

- [ ] **Step 3: Create rulebook.go**

Create `internal/mcp/rulebook.go`:

```go
package mcp

import (
	"context"
	"fmt"
	"strings"

	mcplib "github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) handleSearchRulebook(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	rulesetID, ok := optInt64(req, "ruleset_id")
	if !ok {
		return mcplib.NewToolResultError("ruleset_id is required"), nil
	}
	query, ok := reqStr(req, "query")
	if !ok || query == "" {
		return mcplib.NewToolResultError("query is required"), nil
	}

	chunks, err := s.db.SearchRulebookChunks(rulesetID, query)
	if err != nil {
		return mcplib.NewToolResultError("search failed: " + err.Error()), nil
	}
	if len(chunks) == 0 {
		return mcplib.NewToolResultText("No matching rulebook entries found."), nil
	}

	var sb strings.Builder
	for _, c := range chunks {
		fmt.Fprintf(&sb, "## %s\n%s\n\n", c.Heading, c.Content)
	}
	return mcplib.NewToolResultText(sb.String()), nil
}
```

- [ ] **Step 4: Register the tool in server.go**

In `internal/mcp/server.go`, append to the end of `registerTools()`:

```go
	s.srv.AddTool(mcplib.NewTool("search_rulebook",
		mcplib.WithDescription("Search the ingested rulebook for rules matching a keyword. Returns up to 3 matching sections. Call proactively when adjudicating rules."),
		mcplib.WithNumber("ruleset_id", mcplib.Required(), mcplib.Description("Ruleset ID to search")),
		mcplib.WithString("query", mcplib.Required(), mcplib.Description("Keyword or phrase to search for")),
	), s.handleSearchRulebook)
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
go test ./internal/mcp/... -run 'TestSearchRulebook' -v 2>&1 | tail -20
```

Expected: PASS.

- [ ] **Step 6: Run full test suite**

```bash
go test ./... 2>&1 | tail -20
```

Expected: all PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/mcp/rulebook.go internal/mcp/rulebook_test.go internal/mcp/server.go
git commit -m "feat: search_rulebook MCP tool — LIKE search across ingested rulebook_chunks"
```

---

## Task 5: TypeScript — Ruleset type + new API functions

**Files:**
- Modify: `web/src/types.ts`
- Modify: `web/src/api.ts`
- Modify: `web/src/api.test.ts`

- [ ] **Step 1: Write the failing API tests**

Append to `web/src/api.test.ts` (also update the import at the top):

Update the import line:
```typescript
import { fetchContext, fetchWorldNotes, fetchDiceRolls, fetchRuleset, patchCharacterData } from './api'
```

Add these describe blocks:
```typescript
describe('fetchRuleset', () => {
  it('returns parsed Ruleset on success', async () => {
    const ruleset = { id: 1, name: 'dnd5e', schema_json: '{}', version: '5e' }
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(ruleset),
    }))
    const result = await fetchRuleset(1)
    expect(result.name).toBe('dnd5e')
    expect(result.schema_json).toBe('{}')
  })

  it('throws on non-ok response', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false, status: 404 }))
    await expect(fetchRuleset(99999)).rejects.toThrow('failed: 404')
  })
})

describe('patchCharacterData', () => {
  it('sends PATCH with data_json body', async () => {
    const char = {
      id: 1, campaign_id: 1, name: 'Hero',
      data_json: '{"hp":20}', portrait_path: '', created_at: '',
    }
    const mockFetch = vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve(char) })
    vi.stubGlobal('fetch', mockFetch)
    const result = await patchCharacterData(1, '{"hp":20}')
    expect(result.data_json).toBe('{"hp":20}')
    const [url, opts] = mockFetch.mock.calls[0]
    expect(url).toBe('/api/characters/1')
    expect(opts.method).toBe('PATCH')
    expect(JSON.parse(opts.body)).toEqual({ data_json: '{"hp":20}' })
  })

  it('throws on non-ok response', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false, status: 500 }))
    await expect(patchCharacterData(1, '{}')).rejects.toThrow('failed: 500')
  })
})
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd /home/digitalghost/projects/inkandbone
npm --prefix web test -- --run 2>&1 | tail -20
```

Expected: FAIL — `fetchRuleset` and `patchCharacterData` not exported from `./api`.

- [ ] **Step 3: Add Ruleset type to types.ts**

Append to `web/src/types.ts`:

```typescript
export interface Ruleset {
  id: number
  name: string
  schema_json: string
  version: string
}
```

- [ ] **Step 4: Add API functions to api.ts**

Update the import at the top of `web/src/api.ts`:
```typescript
import type { GameContext, WorldNote, DiceRoll, Ruleset, Character } from './types'
```

Append to `web/src/api.ts`:

```typescript
export async function fetchRuleset(rulesetId: number): Promise<Ruleset> {
  const url = `/api/rulesets/${rulesetId}`
  const res = await fetch(url)
  if (!res.ok) throw new Error(`GET ${url} failed: ${res.status}`)
  return res.json()
}

export async function patchCharacterData(characterId: number, dataJson: string): Promise<Character> {
  const url = `/api/characters/${characterId}`
  const res = await fetch(url, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ data_json: dataJson }),
  })
  if (!res.ok) throw new Error(`PATCH ${url} failed: ${res.status}`)
  return res.json()
}

export async function uploadPortrait(characterId: number, file: File): Promise<Character> {
  const url = `/api/characters/${characterId}/portrait`
  const form = new FormData()
  form.append('portrait', file)
  const res = await fetch(url, { method: 'POST', body: form })
  if (!res.ok) throw new Error(`POST ${url} failed: ${res.status}`)
  return res.json()
}

export async function uploadRulebook(rulesetId: number, text: string): Promise<{ chunks_inserted: number }> {
  const url = `/api/rulesets/${rulesetId}/rulebook`
  const res = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'text/plain' },
    body: text,
  })
  if (!res.ok) throw new Error(`POST ${url} failed: ${res.status}`)
  return res.json()
}
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
npm --prefix web test -- --run 2>&1 | tail -20
```

Expected: all PASS.

- [ ] **Step 6: Commit**

```bash
git add web/src/types.ts web/src/api.ts web/src/api.test.ts
git commit -m "feat: Ruleset type + fetchRuleset/patchCharacterData/uploadPortrait/uploadRulebook API functions"
```

---

## Task 6: CharacterSheetPanel component

**Files:**
- Create: `web/src/CharacterSheetPanel.tsx`
- Create: `web/src/CharacterSheetPanel.test.tsx`

- [ ] **Step 1: Write the failing component tests**

Create `web/src/CharacterSheetPanel.test.tsx`:

```typescript
import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { CharacterSheetPanel } from './CharacterSheetPanel'

afterEach(() => vi.restoreAllMocks())

const baseChar = {
  id: 1, campaign_id: 1, name: 'Kael',
  data_json: '{"hp":"20","class":"fighter"}',
  portrait_path: '', created_at: '',
}

const baseRuleset = {
  id: 1, name: 'dnd5e',
  schema_json: '{"system":"dnd5e","fields":["hp","class"]}',
  version: '5e',
}

function setup(lastEvent: { type: string; payload: unknown } | null = null) {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
    ok: true,
    json: () => Promise.resolve(baseRuleset),
  }))
  return render(
    <CharacterSheetPanel character={baseChar} rulesetId={1} lastEvent={lastEvent} />,
  )
}

describe('CharacterSheetPanel', () => {
  it('renders character name as heading', () => {
    setup()
    expect(screen.getByRole('heading', { name: 'Kael' })).toBeInTheDocument()
  })

  it('renders field labels from schema after fetch', async () => {
    setup()
    await waitFor(() => {
      expect(screen.getByText('hp')).toBeInTheDocument()
      expect(screen.getByText('class')).toBeInTheDocument()
    })
  })

  it('shows current data_json values in inputs', async () => {
    setup()
    await waitFor(() => {
      expect(screen.getByDisplayValue('20')).toBeInTheDocument()
      expect(screen.getByDisplayValue('fighter')).toBeInTheDocument()
    })
  })

  it('updates field values on character_updated WS event for same character', async () => {
    const { rerender } = setup()
    await waitFor(() => screen.getByDisplayValue('20'))

    const event = {
      type: 'character_updated',
      payload: { id: 1, data_json: '{"hp":"99","class":"fighter"}' },
    }
    rerender(
      <CharacterSheetPanel character={baseChar} rulesetId={1} lastEvent={event} />,
    )
    await waitFor(() => {
      expect(screen.getByDisplayValue('99')).toBeInTheDocument()
    })
  })

  it('ignores character_updated event for a different character', async () => {
    const { rerender } = setup()
    await waitFor(() => screen.getByDisplayValue('20'))

    const event = {
      type: 'character_updated',
      payload: { id: 999, data_json: '{"hp":"0"}' },
    }
    rerender(
      <CharacterSheetPanel character={baseChar} rulesetId={1} lastEvent={event} />,
    )
    // hp should still be 20 — event was for a different character
    await waitFor(() => {
      expect(screen.getByDisplayValue('20')).toBeInTheDocument()
    })
  })

  it('shows portrait placeholder when portrait_path is empty', () => {
    setup()
    expect(screen.getByText('No portrait')).toBeInTheDocument()
  })
})
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
npm --prefix web test -- --run 2>&1 | tail -20
```

Expected: FAIL — `CharacterSheetPanel` not found.

- [ ] **Step 3: Create CharacterSheetPanel.tsx**

Create `web/src/CharacterSheetPanel.tsx`:

```typescript
import { useState, useEffect, useRef } from 'react'
import { fetchRuleset, patchCharacterData, uploadPortrait } from './api'
import type { Character, Ruleset } from './types'

// Fields that should render as <input type="number">
const NUMERIC_FIELDS = new Set([
  'hp', 'ac', 'str', 'dex', 'con', 'int', 'wis', 'cha',
  'level', 'proficiency_bonus', 'sanity', 'luck', 'mp', 'humanity',
  'eurodollars', 'age', 'health', 'spirit', 'supply', 'momentum',
  'blood_pool', 'willpower', 'generation', 'edge', 'heart', 'iron',
  'shadow', 'wits', 'initiative', 'speed', 'body', 'ref', 'cool',
  'tech', 'lk', 'att', 'ma', 'emp', 'siz', 'app', 'pow', 'edu',
])

// Fields that should render as <textarea>
const TEXTAREA_FIELDS = new Set([
  'spells', 'features', 'skills', 'inventory', 'notes', 'vows',
  'bonds', 'assets', 'disciplines', 'backgrounds', 'abilities',
  'virtues', 'attributes', 'cyberware', 'gear',
])

interface WSEvent {
  type: string
  payload: unknown
}

interface Props {
  character: Character
  rulesetId: number
  lastEvent: WSEvent | null
}

export function CharacterSheetPanel({ character, rulesetId, lastEvent }: Props) {
  const [ruleset, setRuleset] = useState<Ruleset | null>(null)
  const [data, setData] = useState<Record<string, string>>({})
  const [portraitError, setPortraitError] = useState<string | null>(null)
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Fetch ruleset schema on mount / rulesetId change
  useEffect(() => {
    fetchRuleset(rulesetId).then(setRuleset).catch(() => {})
  }, [rulesetId])

  // Sync data_json into local state when character prop changes
  useEffect(() => {
    try {
      setData(JSON.parse(character.data_json) as Record<string, string>)
    } catch {
      setData({})
    }
  }, [character.data_json])

  // React to character_updated WS events
  useEffect(() => {
    if (!lastEvent || lastEvent.type !== 'character_updated') return
    const payload = lastEvent.payload as { id: number; data_json: string }
    if (payload.id !== character.id) return
    try {
      setData(JSON.parse(payload.data_json) as Record<string, string>)
    } catch {
      // ignore malformed payload
    }
  }, [lastEvent, character.id])

  function handleFieldChange(field: string, value: string) {
    const next = { ...data, [field]: value }
    setData(next)
    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => {
      patchCharacterData(character.id, JSON.stringify(next)).catch(console.error)
    }, 500)
  }

  async function handlePortraitChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    setPortraitError(null)
    try {
      await uploadPortrait(character.id, file)
    } catch {
      setPortraitError('Portrait upload failed.')
    }
  }

  const fields: string[] = ruleset
    ? (JSON.parse(ruleset.schema_json) as { fields: string[] }).fields
    : []

  const portraitUrl = character.portrait_path
    ? `/api/files/${character.portrait_path}`
    : null

  return (
    <section className="panel character-sheet">
      <h2>{character.name}</h2>

      <div className="portrait-wrapper">
        {portraitUrl ? (
          <img className="portrait" src={portraitUrl} alt={`${character.name} portrait`} />
        ) : (
          <div className="portrait-placeholder">No portrait</div>
        )}
        <label className="portrait-upload-label">
          Change portrait
          <input type="file" accept="image/*" onChange={handlePortraitChange} hidden />
        </label>
        {portraitError && <p className="error-text">{portraitError}</p>}
      </div>

      <div className="sheet-fields">
        {fields.map((field) => {
          const value = String(data[field] ?? '')
          if (TEXTAREA_FIELDS.has(field)) {
            return (
              <label key={field} className="sheet-field">
                <span className="field-label">{field}</span>
                <textarea
                  value={value}
                  onChange={(e) => handleFieldChange(field, e.target.value)}
                />
              </label>
            )
          }
          return (
            <label key={field} className="sheet-field">
              <span className="field-label">{field}</span>
              <input
                type={NUMERIC_FIELDS.has(field) ? 'number' : 'text'}
                value={value}
                onChange={(e) => handleFieldChange(field, e.target.value)}
              />
            </label>
          )
        })}
      </div>
    </section>
  )
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
npm --prefix web test -- --run 2>&1 | tail -20
```

Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add web/src/CharacterSheetPanel.tsx web/src/CharacterSheetPanel.test.tsx
git commit -m "feat: CharacterSheetPanel — schema-driven fields, debounced PATCH, WS character_updated sync"
```

---

## Task 7: App wiring + CSS

**Files:**
- Modify: `web/src/App.tsx`
- Modify: `web/src/App.css`

- [ ] **Step 1: Add CharacterSheetPanel to App.tsx**

In `web/src/App.tsx`, add the import at the top with other panel imports:

```typescript
import { CharacterSheetPanel } from './CharacterSheetPanel'
```

Inside the `<main className="panels">` element, add after the last panel (or after the DiceHistoryPanel line). The component renders only when `ctx.character` is non-null:

```tsx
{ctx.character && ctx.campaign && (
  <CharacterSheetPanel
    character={ctx.character}
    rulesetId={ctx.campaign.ruleset_id}
    lastEvent={lastEvent}
  />
)}
```

Note: `lastEvent` is the variable returned by `useWebSocket` (added by Plan 6). If after Plans 6–8 the variable has a different name, match it. In the current (pre-Plan 6) codebase it does not exist — this step applies after Plan 6 has been executed.

- [ ] **Step 2: Verify the app compiles**

```bash
npm --prefix web run build 2>&1 | tail -20
```

Expected: build succeeds, no TypeScript errors.

- [ ] **Step 3: Add character sheet styles to App.css**

Append to `web/src/App.css`:

```css
/* --- Character Sheet Panel --- */
.character-sheet {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.portrait-wrapper {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 6px;
}

.portrait {
  width: 120px;
  height: 120px;
  object-fit: cover;
  border-radius: 4px;
  border: 1px solid #444;
}

.portrait-placeholder {
  width: 120px;
  height: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #2a2a2a;
  border: 1px dashed #555;
  border-radius: 4px;
  font-size: 12px;
  color: #777;
}

.portrait-upload-label {
  font-size: 12px;
  color: #8ab4f8;
  cursor: pointer;
  text-decoration: underline;
}

.sheet-fields {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 8px;
}

.sheet-field {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.field-label {
  font-size: 11px;
  color: #aaa;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.sheet-field input[type="text"],
.sheet-field input[type="number"] {
  background: #1e1e1e;
  border: 1px solid #444;
  border-radius: 3px;
  color: #e0e0e0;
  padding: 4px 6px;
  font-size: 13px;
  width: 100%;
}

.sheet-field textarea {
  background: #1e1e1e;
  border: 1px solid #444;
  border-radius: 3px;
  color: #e0e0e0;
  padding: 4px 6px;
  font-size: 12px;
  resize: vertical;
  min-height: 60px;
  width: 100%;
  grid-column: 1 / -1;
}

.error-text {
  color: #f28b82;
  font-size: 12px;
}
```

Note: textarea fields should span the full width of the grid. Add `grid-column: span 2` or `grid-column: 1 / -1` to textarea's parent `.sheet-field` to make multi-line fields full width. This can be done in the component by adding a `wide` class to textarea fields:

In `CharacterSheetPanel.tsx`, update the textarea branch to add a `wide` class to the label:

```tsx
return (
  <label key={field} className="sheet-field wide">
    <span className="field-label">{field}</span>
    <textarea
      value={value}
      onChange={(e) => handleFieldChange(field, e.target.value)}
    />
  </label>
)
```

And in `App.css` add:

```css
.sheet-field.wide {
  grid-column: 1 / -1;
}
```

- [ ] **Step 4: Run all tests**

```bash
go test ./... && npm --prefix web test -- --run
```

Expected: all Go tests PASS, all TS tests PASS.

- [ ] **Step 5: Commit**

```bash
git add web/src/App.tsx web/src/App.css web/src/CharacterSheetPanel.tsx
git commit -m "feat: wire CharacterSheetPanel into App + character sheet CSS grid layout"
```

---

## Self-Review Checklist

After writing this plan, the following spec requirements from Plan 4 (Plan 9 in our numbering) were verified:

| Spec requirement | Covered in task |
|-----------------|----------------|
| CharacterSheetPanel renders when `ctx.character` non-null | Task 7 |
| Fields driven by `ctx.campaign.ruleset_id` → fetched `schema_json` | Task 6 |
| Schema fields rendered as labeled inputs (text/number/textarea) | Task 6 |
| Debounce 500ms → `PATCH /api/characters/{id}` | Task 6 |
| `character_updated` WS event syncs data_json back | Task 6 |
| Portrait: clickable image or placeholder → file input → POST portrait | Task 6 |
| `PATCH /api/characters/{id}` endpoint | Task 2 |
| `POST /api/characters/{id}/portrait` endpoint | Task 2 |
| `POST /api/rulesets/{id}/rulebook` — paste mode | Task 3 |
| `POST /api/rulesets/{id}/rulebook` — PDF mode | Task 3 |
| Rulebook chunks stored in `rulebook_chunks` table | Task 1 |
| `search_rulebook(query, ruleset_id)` MCP tool | Task 4 |
| Top 3 matching chunks returned | Task 1 (`LIMIT 3`) |
| No vector embeddings — LIKE search | Task 1 |

All requirements covered. No placeholders found.
