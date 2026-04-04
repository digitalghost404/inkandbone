# Plan 6: WebSocket Foundation + Quick Wins

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Wire up the existing WebSocket infrastructure so React panels react to live events, and ship four quick UI wins: dice breakdown badges, world note tag filtering, portrait display, and a PATCH/file-serving HTTP layer.

**Architecture:** The Go WebSocket hub and event bus already work end-to-end. The `useWebSocket` hook and App.tsx already open a connection. This plan fills the gaps: DB tag filtering, two new HTTP routes (PATCH world-note, GET file), MCP tags param, and React panels that subscribe to specific WS event types via a `lastEvent` prop passed from App.

**Tech Stack:** Go 1.22 (net/http mux, path, path/filepath), SQLite LIKE filter, React 19 + TypeScript, Vitest + Testing Library, mcp-go

---

## File Map

| File | Change |
|------|--------|
| `internal/db/queries_world.go` | Add `tag` param to `SearchWorldNotes`; add `tagsJSON` param to `UpdateWorldNote` |
| `internal/db/queries_world_test.go` | Add tag filter test; update UpdateWorldNote test |
| `internal/api/routes.go` | Pass `tag` in `handleListWorldNotes`; add `handlePatchWorldNote`; add `handleServeFile` |
| `internal/api/server.go` | Register two new routes |
| `internal/api/routes_test.go` | Tests for tag filter, PATCH route, file route |
| `internal/mcp/world.go` | Read optional `tags` param; pass `tagsJSON` to DB; update `SearchWorldNotes` call |
| `internal/mcp/server.go` | Add `tags` param to `update_world_note` tool registration |
| `internal/mcp/world_test.go` | Update `SearchWorldNotes` call sites; add tags test |
| `web/src/useWebSocket.ts` | Return `{ lastEvent }` |
| `web/src/useWebSocket.test.tsx` | Add `lastEvent` return test |
| `web/src/api.ts` | Add `tag?` param to `fetchWorldNotes` |
| `web/src/api.test.ts` | Test `tag` param URL construction |
| `web/src/DiceHistoryPanel.tsx` | Parse `breakdown_json`; render die badges; refetch on `dice_rolled` event |
| `web/src/DiceHistoryPanel.test.tsx` | Add breakdown and WS reactivity tests |
| `web/src/WorldNotesPanel.tsx` | Add `lastEvent` prop; tag pills; active-tag filter; refetch on note events |
| `web/src/WorldNotesPanel.test.tsx` | Add tag pill and WS reactivity tests |
| `web/src/App.tsx` | Destructure `lastEvent`; render portrait; pass `lastEvent` to panels |
| `web/src/App.test.tsx` | Add portrait test |
| `web/src/App.css` | Add portrait, tag-pill, die-badge styles |

---

## Task 1: DB — Tag Filter + UpdateWorldNote Tags

**Files:**
- Modify: `internal/db/queries_world.go`
- Modify: `internal/db/queries_world_test.go`

- [ ] **Step 1: Write failing tests for tag filter and UpdateWorldNote tags**

Add to the bottom of `internal/db/queries_world_test.go`:

```go
func TestSearchWorldNotes_tagFilter(t *testing.T) {
	d := newTestDB(t)
	campID := setupCampaign(t, d)

	id1, err := d.CreateWorldNote(campID, "Goblin Den", "Dark cave", "location")
	require.NoError(t, err)
	id2, err := d.CreateWorldNote(campID, "Orc Warlord", "Fierce enemy", "npc")
	require.NoError(t, err)

	require.NoError(t, d.UpdateWorldNote(id1, "Goblin Den", "Dark cave", `["dungeon","encounter"]`))
	require.NoError(t, d.UpdateWorldNote(id2, "Orc Warlord", "Fierce enemy", `["encounter","boss"]`))

	results, err := d.SearchWorldNotes(campID, "", "", "dungeon")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, id1, results[0].ID)

	results, err = d.SearchWorldNotes(campID, "", "", "encounter")
	require.NoError(t, err)
	assert.Len(t, results, 2)

	results, err = d.SearchWorldNotes(campID, "", "", "boss")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, id2, results[0].ID)

	results, err = d.SearchWorldNotes(campID, "", "", "")
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestUpdateWorldNote_setsTagsJSON(t *testing.T) {
	d := newTestDB(t)
	campID := setupCampaign(t, d)

	id, err := d.CreateWorldNote(campID, "Mira", "A merchant.", "npc")
	require.NoError(t, err)

	require.NoError(t, d.UpdateWorldNote(id, "Mira", "A merchant.", `["npc","ally"]`))
	results, err := d.SearchWorldNotes(campID, "", "", "ally")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Contains(t, results[0].TagsJSON, "ally")
}
```

Also update the existing `TestWorldNotes` call sites to pass the new `tag` param (fourth arg `""`):

```go
// line: results, err := d.SearchWorldNotes(campID, "Gareth", "")
// becomes:
results, err := d.SearchWorldNotes(campID, "Gareth", "", "")
// ... and all other SearchWorldNotes calls in this test
results, err = d.SearchWorldNotes(campID, "", "npc", "")
results, err = d.SearchWorldNotes(campID, "", "location", "")
results, err = d.SearchWorldNotes(campID, "kind", "", "")
// and the UpdateWorldNote call:
require.NoError(t, d.UpdateWorldNote(id, "Gareth the Guard", "A surly but kind dwarf", ""))
```

- [ ] **Step 2: Run tests — verify compile error on signature mismatch**

```bash
go test ./internal/db/... -run TestWorldNotes -v
```

Expected: FAIL — compile error, wrong number of arguments.

- [ ] **Step 3: Update SearchWorldNotes to accept tag param**

In `internal/db/queries_world.go`, replace the `SearchWorldNotes` function:

```go
func (d *DB) SearchWorldNotes(campaignID int64, query, category, tag string) ([]WorldNote, error) {
	q := "SELECT id, campaign_id, title, content, category, tags_json, created_at FROM world_notes WHERE campaign_id = ?"
	args := []any{campaignID}
	if query != "" {
		q += " AND (title LIKE ? OR content LIKE ?)"
		like := "%" + query + "%"
		args = append(args, like, like)
	}
	if category != "" {
		q += " AND category = ?"
		args = append(args, category)
	}
	if tag != "" {
		q += " AND tags_json LIKE ?"
		args = append(args, `%"`+tag+`"%`)
	}
	q += " ORDER BY title"
	rows, err := d.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []WorldNote
	for rows.Next() {
		var n WorldNote
		if err := rows.Scan(&n.ID, &n.CampaignID, &n.Title, &n.Content, &n.Category, &n.TagsJSON, &n.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}
```

- [ ] **Step 4: Update UpdateWorldNote to accept tagsJSON param**

Replace the `UpdateWorldNote` function in `internal/db/queries_world.go`:

```go
func (d *DB) UpdateWorldNote(id int64, title, content, tagsJSON string) error {
	var res sql.Result
	var err error
	if tagsJSON != "" {
		res, err = d.db.Exec(
			"UPDATE world_notes SET title = ?, content = ?, tags_json = ? WHERE id = ?",
			title, content, tagsJSON, id,
		)
	} else {
		res, err = d.db.Exec(
			"UPDATE world_notes SET title = ?, content = ? WHERE id = ?",
			title, content, id,
		)
	}
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("world note %d not found", id)
	}
	return nil
}
```

Add `"database/sql"` to the import block if not already present (it already is — `sql.ErrNoRows` is used in `GetMap`).

- [ ] **Step 5: Fix all call sites that break from the signature changes**

In `internal/api/routes.go`, update `handleListWorldNotes` to extract and pass the `tag` query param:

```go
func (s *Server) handleListWorldNotes(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}
	q := r.URL.Query().Get("q")
	category := r.URL.Query().Get("category")
	tag := r.URL.Query().Get("tag")
	notes, err := s.db.SearchWorldNotes(id, q, category, tag)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if notes == nil {
		notes = []db.WorldNote{}
	}
	writeJSON(w, notes)
}
```

In `internal/mcp/world.go`, update `handleSearchWorldNotes` to pass `""` for tag:

```go
notes, err := s.db.SearchWorldNotes(campID, query, category, "")
```

In `internal/mcp/world_test.go`, update all `SearchWorldNotes` call sites:

```go
// TestCreateWorldNote: change
notes, err := s.db.SearchWorldNotes(campID, "Mira", "")
// to:
notes, err := s.db.SearchWorldNotes(campID, "Mira", "", "")

// TestUpdateWorldNote: change
notes, err := s.db.SearchWorldNotes(campID, "New Title", "")
// to:
notes, err := s.db.SearchWorldNotes(campID, "New Title", "", "")

// TestSearchWorldNotes: no direct SearchWorldNotes call on db — no change needed
```

- [ ] **Step 6: Run all Go tests to verify green**

```bash
go test ./internal/db/... ./internal/api/... ./internal/mcp/... -v
```

Expected: ALL PASS. The two new tag tests pass, all existing tests still pass.

- [ ] **Step 7: Commit**

```bash
git add internal/db/queries_world.go internal/db/queries_world_test.go \
        internal/api/routes.go \
        internal/mcp/world.go internal/mcp/world_test.go
git commit -m "feat: add tag filter to SearchWorldNotes and tagsJSON to UpdateWorldNote"
```

---

## Task 2: HTTP — PATCH World-Note + File Serving

**Files:**
- Modify: `internal/api/routes.go`
- Modify: `internal/api/server.go`
- Modify: `internal/api/routes_test.go`

- [ ] **Step 1: Write failing tests for PATCH world-note and file serving**

Add to `internal/api/routes_test.go`:

```go
func TestListWorldNotes_tagFilter(t *testing.T) {
	s := newTestServer(t)
	campID, _ := seedCampaign(t, s.db)
	noteID, err := s.db.CreateWorldNote(campID, "Shrine", "Ancient shrine.", "location")
	require.NoError(t, err)
	require.NoError(t, s.db.UpdateWorldNote(noteID, "Shrine", "Ancient shrine.", `["dungeon"]`))
	_, err = s.db.CreateWorldNote(campID, "Merchant", "Sells goods.", "npc")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/campaigns/"+strconv.FormatInt(campID, 10)+"/world-notes?tag=dungeon", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var notes []db.WorldNote
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &notes))
	require.Len(t, notes, 1)
	assert.Equal(t, "Shrine", notes[0].Title)
}

func TestPatchWorldNote_updatesNote(t *testing.T) {
	s := newTestServer(t)
	campID, _ := seedCampaign(t, s.db)
	noteID, err := s.db.CreateWorldNote(campID, "Old Title", "Old content", "npc")
	require.NoError(t, err)

	body := `{"title":"New Title","content":"New content","tags_json":"[\"ally\"]"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/world-notes/"+strconv.FormatInt(noteID, 10), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)

	notes, err := s.db.SearchWorldNotes(campID, "New Title", "", "")
	require.NoError(t, err)
	require.Len(t, notes, 1)
	assert.Equal(t, "New content", notes[0].Content)
	assert.Contains(t, notes[0].TagsJSON, "ally")
}

func TestPatchWorldNote_invalidID(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodPatch, "/api/world-notes/abc", strings.NewReader(`{"title":"x","content":"y"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPatchWorldNote_missingTitle(t *testing.T) {
	s := newTestServer(t)
	campID, _ := seedCampaign(t, s.db)
	noteID, err := s.db.CreateWorldNote(campID, "A Note", "Content.", "npc")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPatch, "/api/world-notes/"+strconv.FormatInt(noteID, 10), strings.NewReader(`{"content":"y"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServeFile_notFound(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/files/portraits/nonexistent.jpg", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
```

Add `"strings"` to the import block in routes_test.go.

- [ ] **Step 2: Run tests — verify they fail**

```bash
go test ./internal/api/... -run "TestPatchWorldNote|TestServeFile|TestListWorldNotes_tagFilter" -v
```

Expected: FAIL — routes not registered yet.

- [ ] **Step 3: Add handlePatchWorldNote and handleServeFile to routes.go**

Add `"path"` and `"path/filepath"` to the import block. Add to the bottom of `internal/api/routes.go`:

```go
func (s *Server) handlePatchWorldNote(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid world note id", http.StatusBadRequest)
		return
	}
	var body struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		TagsJSON string `json:"tags_json"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if body.Title == "" || body.Content == "" {
		http.Error(w, "title and content are required", http.StatusBadRequest)
		return
	}
	if err := s.db.UpdateWorldNote(id, body.Title, body.Content, body.TagsJSON); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.bus.Publish(Event{Type: EventWorldNoteUpdated, Payload: map[string]any{"note_id": id}})
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleServeFile(w http.ResponseWriter, r *http.Request) {
	rawPath := r.PathValue("path")
	// path.Clean prevents traversal: "/../.." resolves to "/" which stays under data/
	safe := path.Clean("/" + rawPath)
	http.ServeFile(w, r, filepath.Join("data", safe))
}
```

- [ ] **Step 4: Register the new routes in server.go**

In `internal/api/server.go`, inside `registerRoutes()`, add after the existing routes:

```go
s.mux.HandleFunc("PATCH /api/world-notes/{id}", s.handlePatchWorldNote)
s.mux.HandleFunc("GET /api/files/{path...}", s.handleServeFile)
```

- [ ] **Step 5: Run tests — verify green**

```bash
go test ./internal/api/... -v
```

Expected: ALL PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/api/routes.go internal/api/server.go internal/api/routes_test.go
git commit -m "feat: add PATCH /api/world-notes/{id} and GET /api/files/{path} routes"
```

---

## Task 3: MCP — Tags Param for update_world_note

**Files:**
- Modify: `internal/mcp/world.go`
- Modify: `internal/mcp/server.go`
- Modify: `internal/mcp/world_test.go`

- [ ] **Step 1: Write failing test for tags param**

Add to `internal/mcp/world_test.go`:

```go
func TestUpdateWorldNote_withTags(t *testing.T) {
	s := newTestMCP(t)
	campID := setupActiveCampaign(t, s)

	noteID, err := s.db.CreateWorldNote(campID, "Old", "Old content", "npc")
	require.NoError(t, err)

	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"note_id": float64(noteID),
		"title":   "New",
		"content": "New content",
		"tags":    `["boss","undead"]`,
	}
	result, err := s.handleUpdateWorldNote(context.Background(), req)
	require.NoError(t, err)
	require.False(t, result.IsError)

	notes, err := s.db.SearchWorldNotes(campID, "New", "", "boss")
	require.NoError(t, err)
	require.Len(t, notes, 1)
	assert.Contains(t, notes[0].TagsJSON, "boss")
}
```

- [ ] **Step 2: Run test — verify it fails**

```bash
go test ./internal/mcp/... -run TestUpdateWorldNote_withTags -v
```

Expected: FAIL — tags not parsed, `SearchWorldNotes` with tag returns 0 results.

- [ ] **Step 3: Update handleUpdateWorldNote to read and pass tags**

Replace `handleUpdateWorldNote` in `internal/mcp/world.go`:

```go
func (s *Server) handleUpdateWorldNote(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	noteID, ok := optInt64(req, "note_id")
	if !ok {
		return mcplib.NewToolResultError("note_id is required"), nil
	}
	title, ok := reqStr(req, "title")
	if !ok {
		return mcplib.NewToolResultError("title is required"), nil
	}
	content, ok := reqStr(req, "content")
	if !ok {
		return mcplib.NewToolResultError("content is required"), nil
	}

	tagsJSON := ""
	if tagsRaw := optStr(req, "tags"); tagsRaw != "" {
		var tags []string
		if err := json.Unmarshal([]byte(tagsRaw), &tags); err != nil {
			return mcplib.NewToolResultError("tags must be a JSON array of strings"), nil
		}
		b, _ := json.Marshal(tags)
		tagsJSON = string(b)
	}

	if err := s.db.UpdateWorldNote(noteID, title, content, tagsJSON); err != nil {
		return mcplib.NewToolResultError("update note: " + err.Error()), nil
	}

	sessID, _ := s.activeSessionID()
	s.logNarrative(req, sessID)
	s.bus.Publish(api.Event{Type: api.EventWorldNoteUpdated, Payload: map[string]any{"note_id": noteID}})
	return mcplib.NewToolResultText(fmt.Sprintf("world note %d updated", noteID)), nil
}
```

The `json` package is already imported in `world.go`. If not, add `"encoding/json"` to the import block.

- [ ] **Step 4: Add tags param to tool registration in server.go**

In `internal/mcp/server.go`, update the `update_world_note` tool registration to add the `tags` option before the closing `)`:

```go
s.srv.AddTool(mcplib.NewTool("update_world_note",
    mcplib.WithDescription("Edit an existing world note."),
    mcplib.WithNumber("note_id", mcplib.Required(), mcplib.Description("World note ID")),
    mcplib.WithString("title", mcplib.Required(), mcplib.Description("New title")),
    mcplib.WithString("content", mcplib.Required(), mcplib.Description("New content")),
    mcplib.WithString("tags", mcplib.Description(`JSON array of tag strings, e.g. ["npc","villain"]`)),
    mcplib.WithString("narrative", mcplib.Description("Optional narrative to log")),
), s.handleUpdateWorldNote)
```

- [ ] **Step 5: Run all MCP tests — verify green**

```bash
go test ./internal/mcp/... -v
```

Expected: ALL PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/mcp/world.go internal/mcp/server.go internal/mcp/world_test.go
git commit -m "feat: add optional tags param to update_world_note MCP tool"
```

---

## Task 4: useWebSocket — Return lastEvent

**Files:**
- Modify: `web/src/useWebSocket.ts`
- Modify: `web/src/useWebSocket.test.tsx`

- [ ] **Step 1: Write failing test for lastEvent return value**

Add inside the `describe('useWebSocket', ...)` block in `web/src/useWebSocket.test.tsx`:

```typescript
it('returns lastEvent after receiving a message', () => {
  const { result } = renderHook(() => useWebSocket('/ws', vi.fn()))
  act(() => instances[0].open())
  act(() => instances[0].receive({ type: 'dice_rolled', payload: { total: 15 } }))
  expect(result.current.lastEvent).toEqual({ type: 'dice_rolled', payload: { total: 15 } })
})
```

- [ ] **Step 2: Run test — verify it fails**

```bash
cd web && npx vitest run src/useWebSocket.test.tsx
```

Expected: FAIL — `result.current.lastEvent` is undefined (hook returns void).

- [ ] **Step 3: Update useWebSocket to return lastEvent**

Replace `web/src/useWebSocket.ts` entirely:

```typescript
import { useEffect, useRef, useState } from 'react'

export function useWebSocket(url: string, onMessage: (data: unknown) => void): { lastEvent: unknown } {
  const [lastEvent, setLastEvent] = useState<unknown>(null)
  const onMessageRef = useRef(onMessage)
  onMessageRef.current = onMessage

  useEffect(() => {
    let ws: WebSocket
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null
    let cancelled = false

    function connect() {
      ws = new WebSocket(url)

      ws.onmessage = (e) => {
        try {
          const parsed = JSON.parse(e.data as string)
          setLastEvent(parsed)
          onMessageRef.current(parsed)
        } catch {
          // ignore malformed messages
        }
      }

      ws.onclose = () => {
        if (!cancelled) {
          reconnectTimer = setTimeout(connect, 2000)
        }
      }
    }

    connect()

    return () => {
      cancelled = true
      if (reconnectTimer !== null) clearTimeout(reconnectTimer)
      ws.close()
    }
  }, [url])

  return { lastEvent }
}
```

- [ ] **Step 4: Run all web tests — verify green**

```bash
cd web && npm test
```

Expected: ALL PASS. The three existing tests still pass because they don't use the return value.

- [ ] **Step 5: Commit**

```bash
git add web/src/useWebSocket.ts web/src/useWebSocket.test.tsx
git commit -m "feat: useWebSocket returns lastEvent for component-level WS reactivity"
```

---

## Task 5: DiceHistoryPanel — Breakdown Badges + WS Reactivity

**Files:**
- Modify: `web/src/DiceHistoryPanel.tsx`
- Modify: `web/src/DiceHistoryPanel.test.tsx`

- [ ] **Step 1: Write failing tests**

Replace `web/src/DiceHistoryPanel.test.tsx`:

```typescript
import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, cleanup, waitFor } from '@testing-library/react'
import { DiceHistoryPanel } from './DiceHistoryPanel'
import type { DiceRoll } from './types'

const rolls: DiceRoll[] = [
  { id: 1, session_id: 1, expression: '1d20+5', result: 18, breakdown_json: '[]', created_at: '' },
  { id: 2, session_id: 1, expression: '2d6', result: 7, breakdown_json: '[]', created_at: '' },
]

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

describe('DiceHistoryPanel', () => {
  it('renders dice rolls', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve(rolls) }))
    render(<DiceHistoryPanel sessionId={1} lastEvent={null} />)
    expect(await screen.findByText('1d20+5')).toBeInTheDocument()
    expect(screen.getByText('18')).toBeInTheDocument()
    expect(screen.getByText('2d6')).toBeInTheDocument()
    expect(screen.getByText('7')).toBeInTheDocument()
  })

  it('shows empty state when no rolls', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve([]) }))
    render(<DiceHistoryPanel sessionId={1} lastEvent={null} />)
    expect(await screen.findByText('No rolls yet.')).toBeInTheDocument()
  })

  it('renders breakdown badges from breakdown_json', async () => {
    const withBreakdown: DiceRoll[] = [
      { id: 1, session_id: 1, expression: '3d6', result: 12, breakdown_json: '[4,3,5]', created_at: '' },
    ]
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve(withBreakdown) }))
    render(<DiceHistoryPanel sessionId={1} lastEvent={null} />)
    expect(await screen.findByText('[4]')).toBeInTheDocument()
    expect(screen.getByText('[3]')).toBeInTheDocument()
    expect(screen.getByText('[5]')).toBeInTheDocument()
  })

  it('does not render badges when breakdown_json is empty array', async () => {
    const noBreakdown: DiceRoll[] = [
      { id: 1, session_id: 1, expression: '1d20', result: 15, breakdown_json: '[]', created_at: '' },
    ]
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve(noBreakdown) }))
    render(<DiceHistoryPanel sessionId={1} lastEvent={null} />)
    await screen.findByText('1d20')
    expect(screen.queryByText(/^\[/)).toBeNull()
  })

  it('refetches rolls on dice_rolled event', async () => {
    const mockFetch = vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve([]) })
    vi.stubGlobal('fetch', mockFetch)
    const { rerender } = render(<DiceHistoryPanel sessionId={1} lastEvent={null} />)
    await screen.findByText('No rolls yet.')
    const callsBefore = mockFetch.mock.calls.length
    rerender(<DiceHistoryPanel sessionId={1} lastEvent={{ type: 'dice_rolled', payload: { total: 15 } }} />)
    await waitFor(() => {
      expect(mockFetch.mock.calls.length).toBeGreaterThan(callsBefore)
    })
  })
})
```

- [ ] **Step 2: Run tests — verify they fail**

```bash
cd web && npx vitest run src/DiceHistoryPanel.test.tsx
```

Expected: FAIL — `lastEvent` prop missing, no breakdown rendering.

- [ ] **Step 3: Implement breakdown badges and WS reactivity**

Replace `web/src/DiceHistoryPanel.tsx`:

```typescript
import { useState, useEffect } from 'react'
import { fetchDiceRolls } from './api'
import type { DiceRoll } from './types'

interface Props {
  sessionId: number
  lastEvent: unknown
}

function parseBreakdown(json: string): number[] {
  try { return JSON.parse(json) as number[] }
  catch { return [] }
}

export function DiceHistoryPanel({ sessionId, lastEvent }: Props) {
  const [rolls, setRolls] = useState<DiceRoll[]>([])

  useEffect(() => {
    let ignored = false
    fetchDiceRolls(sessionId)
      .then((data) => { if (!ignored) setRolls(data) })
      .catch(() => { if (!ignored) setRolls([]) })
    return () => { ignored = true }
  }, [sessionId])

  useEffect(() => {
    const ev = lastEvent as { type?: string } | null
    if (ev?.type === 'dice_rolled') {
      fetchDiceRolls(sessionId)
        .then(setRolls)
        .catch(() => {})
    }
  }, [lastEvent, sessionId])

  return (
    <section className="panel dice-history">
      <h2>Dice History</h2>
      {rolls.length === 0 ? (
        <p className="empty">No rolls yet.</p>
      ) : (
        rolls.map((r) => {
          const breakdown = parseBreakdown(r.breakdown_json)
          return (
            <div key={r.id} className="dice-roll">
              <div className="roll-top">
                <span className="expression">{r.expression}</span>
                <span className="result">{r.result}</span>
              </div>
              {breakdown.length > 0 && (
                <div className="breakdown">
                  {breakdown.map((d, i) => (
                    <span key={i} className="die-badge">[{d}]</span>
                  ))}
                </div>
              )}
            </div>
          )
        })
      )}
    </section>
  )
}
```

- [ ] **Step 4: Run tests — verify green**

```bash
cd web && npx vitest run src/DiceHistoryPanel.test.tsx
```

Expected: ALL PASS.

- [ ] **Step 5: Commit**

```bash
git add web/src/DiceHistoryPanel.tsx web/src/DiceHistoryPanel.test.tsx
git commit -m "feat: DiceHistoryPanel renders breakdown badges and reacts to dice_rolled WS event"
```

---

## Task 6: WorldNotesPanel — Tag Pills + Tag Filter + WS Reactivity

**Files:**
- Modify: `web/src/api.ts`
- Modify: `web/src/api.test.ts`
- Modify: `web/src/WorldNotesPanel.tsx`
- Modify: `web/src/WorldNotesPanel.test.tsx`

- [ ] **Step 1: Write failing tests for fetchWorldNotes tag param**

Add to the `describe('fetchWorldNotes', ...)` block in `web/src/api.test.ts`:

```typescript
it('appends tag param when tag is provided', async () => {
  const mockFetch = vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve([]) })
  vi.stubGlobal('fetch', mockFetch)
  await fetchWorldNotes(1, undefined, 'npc')
  expect(mockFetch).toHaveBeenCalledWith('/api/campaigns/1/world-notes?tag=npc')
})

it('appends both q and tag when both provided', async () => {
  const mockFetch = vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve([]) })
  vi.stubGlobal('fetch', mockFetch)
  await fetchWorldNotes(1, 'tavern', 'location')
  expect(mockFetch).toHaveBeenCalledWith('/api/campaigns/1/world-notes?q=tavern&tag=location')
})
```

- [ ] **Step 2: Run tests — verify they fail**

```bash
cd web && npx vitest run src/api.test.ts
```

Expected: FAIL — `fetchWorldNotes` doesn't accept third arg, URL doesn't include `tag`.

- [ ] **Step 3: Update fetchWorldNotes to accept tag param**

Replace `fetchWorldNotes` in `web/src/api.ts`:

```typescript
export async function fetchWorldNotes(campaignId: number, q?: string, tag?: string): Promise<WorldNote[]> {
  const params = new URLSearchParams()
  if (q) params.set('q', q)
  if (tag) params.set('tag', tag)
  const qs = params.toString()
  const url = qs
    ? `/api/campaigns/${campaignId}/world-notes?${qs}`
    : `/api/campaigns/${campaignId}/world-notes`
  const res = await fetch(url)
  if (!res.ok) throw new Error(`GET ${url} failed: ${res.status}`)
  return res.json()
}
```

- [ ] **Step 4: Run api tests — verify green**

```bash
cd web && npx vitest run src/api.test.ts
```

Expected: ALL PASS. The existing `'appends q param'` and `'omits q param'` tests still pass.

- [ ] **Step 5: Write failing tests for WorldNotesPanel tag pills and WS reactivity**

Replace `web/src/WorldNotesPanel.test.tsx`:

```typescript
import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, cleanup, waitFor, fireEvent } from '@testing-library/react'
import { WorldNotesPanel } from './WorldNotesPanel'
import type { WorldNote } from './types'

const notes: WorldNote[] = [
  { id: 1, campaign_id: 1, title: 'Tavern', content: 'A seedy place.', category: 'location', tags_json: '["inn"]', created_at: '' },
  { id: 2, campaign_id: 1, title: 'Dragon', content: 'Ancient red dragon.', category: 'npc', tags_json: '[]', created_at: '' },
]

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

describe('WorldNotesPanel', () => {
  it('renders notes returned by fetch', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve(notes) }))
    render(<WorldNotesPanel campaignId={1} lastEvent={null} />)
    expect(await screen.findByText('Tavern')).toBeInTheDocument()
    expect(screen.getByText('Dragon')).toBeInTheDocument()
  })

  it('shows empty state when no notes', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve([]) }))
    render(<WorldNotesPanel campaignId={1} lastEvent={null} />)
    expect(await screen.findByText('No notes found.')).toBeInTheDocument()
  })

  it('calls fetch with q param when search input changes', async () => {
    const mockFetch = vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve([]) })
    vi.stubGlobal('fetch', mockFetch)
    render(<WorldNotesPanel campaignId={1} lastEvent={null} />)
    await screen.findByText('No notes found.')
    fireEvent.change(screen.getByRole('searchbox'), { target: { value: 'tavern' } })
    await waitFor(() => {
      expect(mockFetch).toHaveBeenLastCalledWith('/api/campaigns/1/world-notes?q=tavern')
    })
  })

  it('renders tag pills from tags_json', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve(notes) }))
    render(<WorldNotesPanel campaignId={1} lastEvent={null} />)
    expect(await screen.findByText('inn')).toBeInTheDocument()
  })

  it('filters by tag when a tag pill is clicked', async () => {
    const mockFetch = vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve(notes) })
    vi.stubGlobal('fetch', mockFetch)
    render(<WorldNotesPanel campaignId={1} lastEvent={null} />)
    fireEvent.click(await screen.findByText('inn'))
    await waitFor(() => {
      expect(mockFetch).toHaveBeenLastCalledWith('/api/campaigns/1/world-notes?tag=inn')
    })
  })

  it('deselects tag when same pill is clicked again', async () => {
    const mockFetch = vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve(notes) })
    vi.stubGlobal('fetch', mockFetch)
    render(<WorldNotesPanel campaignId={1} lastEvent={null} />)
    const pill = await screen.findByText('inn')
    fireEvent.click(pill)
    await waitFor(() => expect(mockFetch).toHaveBeenLastCalledWith('/api/campaigns/1/world-notes?tag=inn'))
    fireEvent.click(pill)
    await waitFor(() => expect(mockFetch).toHaveBeenLastCalledWith('/api/campaigns/1/world-notes'))
  })

  it('refetches on world_note_updated event', async () => {
    const mockFetch = vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve([]) })
    vi.stubGlobal('fetch', mockFetch)
    const { rerender } = render(<WorldNotesPanel campaignId={1} lastEvent={null} />)
    await screen.findByText('No notes found.')
    const callsBefore = mockFetch.mock.calls.length
    rerender(<WorldNotesPanel campaignId={1} lastEvent={{ type: 'world_note_updated', payload: { note_id: 1 } }} />)
    await waitFor(() => {
      expect(mockFetch.mock.calls.length).toBeGreaterThan(callsBefore)
    })
  })

  it('refetches on world_note_created event', async () => {
    const mockFetch = vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve([]) })
    vi.stubGlobal('fetch', mockFetch)
    const { rerender } = render(<WorldNotesPanel campaignId={1} lastEvent={null} />)
    await screen.findByText('No notes found.')
    const callsBefore = mockFetch.mock.calls.length
    rerender(<WorldNotesPanel campaignId={1} lastEvent={{ type: 'world_note_created', payload: { note_id: 2 } }} />)
    await waitFor(() => {
      expect(mockFetch.mock.calls.length).toBeGreaterThan(callsBefore)
    })
  })
})
```

- [ ] **Step 6: Run tests — verify they fail**

```bash
cd web && npx vitest run src/WorldNotesPanel.test.tsx
```

Expected: FAIL — `lastEvent` prop missing, no tag pills.

- [ ] **Step 7: Implement tag pills and WS reactivity**

Replace `web/src/WorldNotesPanel.tsx`:

```typescript
import { useState, useEffect, useCallback } from 'react'
import { fetchWorldNotes } from './api'
import type { WorldNote } from './types'

interface Props {
  campaignId: number
  lastEvent: unknown
}

function parseTags(json: string): string[] {
  try { return JSON.parse(json) as string[] }
  catch { return [] }
}

export function WorldNotesPanel({ campaignId, lastEvent }: Props) {
  const [notes, setNotes] = useState<WorldNote[]>([])
  const [query, setQuery] = useState('')
  const [activeTag, setActiveTag] = useState<string | null>(null)

  const loadNotes = useCallback(() => {
    fetchWorldNotes(campaignId, query || undefined, activeTag || undefined)
      .then(setNotes)
      .catch(() => setNotes([]))
  }, [campaignId, query, activeTag])

  useEffect(() => {
    loadNotes()
  }, [loadNotes])

  useEffect(() => {
    const ev = lastEvent as { type?: string } | null
    if (ev?.type === 'world_note_updated' || ev?.type === 'world_note_created') {
      loadNotes()
    }
  }, [lastEvent, loadNotes])

  return (
    <section className="panel world-notes">
      <h2>World Notes</h2>
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
        notes.map((n) => {
          const tags = parseTags(n.tags_json)
          return (
            <div key={n.id} className="world-note">
              <div className="note-header">
                <span className="note-title">{n.title}</span>
                {n.category && <span className="note-category">{n.category}</span>}
              </div>
              {tags.length > 0 && (
                <div className="tag-pills">
                  {tags.map((tag) => (
                    <button
                      key={tag}
                      className={`tag-pill${activeTag === tag ? ' active' : ''}`}
                      onClick={() => setActiveTag((t) => (t === tag ? null : tag))}
                    >
                      {tag}
                    </button>
                  ))}
                </div>
              )}
              <p className="note-content">{n.content}</p>
            </div>
          )
        })
      )}
    </section>
  )
}
```

- [ ] **Step 8: Run all web tests — verify green**

```bash
cd web && npm test
```

Expected: ALL PASS.

- [ ] **Step 9: Commit**

```bash
git add web/src/api.ts web/src/api.test.ts \
        web/src/WorldNotesPanel.tsx web/src/WorldNotesPanel.test.tsx
git commit -m "feat: WorldNotesPanel tag pills, tag filtering, and WS reactivity"
```

---

## Task 7: App — Portrait Display + Wire lastEvent + Styles

**Files:**
- Modify: `web/src/App.tsx`
- Modify: `web/src/App.test.tsx`
- Modify: `web/src/App.css`

- [ ] **Step 1: Write failing portrait test**

Add to `web/src/App.test.tsx` inside the `describe('App', ...)` block:

```typescript
it('renders portrait img when portrait_path is set', async () => {
  const ctxWithPortrait = {
    ...mockCtx,
    character: { ...mockCtx.character!, portrait_path: 'portraits/zara.jpg' },
  }
  vi.stubGlobal('fetch', vi.fn().mockImplementation((url: string) => {
    if (url === '/api/context') {
      return Promise.resolve({ ok: true, json: () => Promise.resolve(ctxWithPortrait) })
    }
    return Promise.resolve({ ok: true, json: () => Promise.resolve([]) })
  }))
  render(<App />)
  const img = await screen.findByRole('img', { name: 'Zara' })
  expect(img).toHaveAttribute('src', '/api/files/portraits/zara.jpg')
})

it('does not render portrait img when portrait_path is empty', async () => {
  vi.stubGlobal('fetch', vi.fn().mockImplementation((url: string) => {
    if (url === '/api/context') {
      return Promise.resolve({ ok: true, json: () => Promise.resolve(mockCtx) })
    }
    return Promise.resolve({ ok: true, json: () => Promise.resolve([]) })
  }))
  render(<App />)
  await screen.findByText('Greyhawk')
  expect(screen.queryByRole('img')).toBeNull()
})
```

- [ ] **Step 2: Run tests — verify they fail**

```bash
cd web && npx vitest run src/App.test.tsx
```

Expected: FAIL — portrait not rendered, `lastEvent` not passed to panels (type error).

- [ ] **Step 3: Update App.tsx**

Replace `web/src/App.tsx`:

```typescript
import { useState, useEffect, useCallback } from 'react'
import { useWebSocket } from './useWebSocket'
import { fetchContext } from './api'
import type { GameContext, Message } from './types'
import { WorldNotesPanel } from './WorldNotesPanel'
import { DiceHistoryPanel } from './DiceHistoryPanel'
import './App.css'

const WS_URL = `ws://${window.location.host}/ws`

export default function App() {
  const [ctx, setCtx] = useState<GameContext | null>(null)
  const [messages, setMessages] = useState<Message[]>([])
  const [error, setError] = useState<string | null>(null)

  const loadContext = useCallback(() => {
    fetchContext()
      .then((data) => {
        setCtx(data)
        setMessages(data.recent_messages ?? [])
      })
      .catch(() => setError('Could not load game state'))
  }, [])

  useEffect(() => {
    loadContext()
  }, [loadContext])

  const handleEvent = useCallback(
    (_data: unknown) => {
      loadContext()
    },
    [loadContext],
  )

  const { lastEvent } = useWebSocket(WS_URL, handleEvent)

  if (error) return <div className="error">{error}</div>
  if (!ctx) return <div className="loading">Loading…</div>

  return (
    <div className="dashboard">
      <header className="state-bar">
        <span className="campaign">{ctx.campaign?.name ?? 'No campaign'}</span>
        <span className="separator">·</span>
        <span className="character-info">
          {ctx.character?.portrait_path && (
            <img
              className="portrait"
              src={`/api/files/${ctx.character.portrait_path}`}
              alt={ctx.character.name}
            />
          )}
          <span className="character">{ctx.character?.name ?? 'No character'}</span>
        </span>
        <span className="separator">·</span>
        <span className="session">{ctx.session?.title ?? 'No session'}</span>
      </header>

      <main className="panels">
        <section className="panel messages">
          <h2>Session Log</h2>
          {messages.length === 0 ? (
            <p className="empty">No messages yet.</p>
          ) : (
            messages.map((m) => (
              <div key={m.id} className={`message ${m.role}`}>
                <span className="role">{m.role}</span>
                <span className="content">{m.content}</span>
              </div>
            ))
          )}
        </section>

        {ctx.active_combat && (
          <section className="panel combat">
            <h2>Combat: {ctx.active_combat.encounter.name}</h2>
            <table>
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Init</th>
                  <th>HP</th>
                </tr>
              </thead>
              <tbody>
                {ctx.active_combat.combatants.map((c) => (
                  <tr key={c.id} className={c.is_player ? 'player' : 'enemy'}>
                    <td>{c.name}</td>
                    <td>{c.initiative}</td>
                    <td>
                      {c.hp_current}/{c.hp_max}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </section>
        )}

        {ctx.campaign && <WorldNotesPanel campaignId={ctx.campaign.id} lastEvent={lastEvent} />}

        {ctx.session && <DiceHistoryPanel sessionId={ctx.session.id} lastEvent={lastEvent} />}
      </main>
    </div>
  )
}
```

- [ ] **Step 4: Run App tests — verify green**

```bash
cd web && npx vitest run src/App.test.tsx
```

Expected: ALL PASS.

- [ ] **Step 5: Add portrait, tag-pill, and die-badge CSS**

Add to the bottom of `web/src/App.css`:

```css
/* Portrait */

.character-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.portrait {
  width: 1.75rem;
  height: 1.75rem;
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid var(--border);
  flex-shrink: 0;
}

/* Tag Pills */

.tag-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
  margin-top: 0.2rem;
}

.tag-pill {
  background: transparent;
  border: 1px solid var(--border);
  border-radius: 999px;
  color: var(--muted);
  cursor: pointer;
  font-size: 0.7rem;
  padding: 0.1rem 0.5rem;
}

.tag-pill:hover {
  border-color: var(--accent);
  color: var(--text);
}

.tag-pill.active {
  background: var(--accent);
  border-color: var(--accent);
  color: #fff;
}

/* Die Badges */

.breakdown {
  display: flex;
  flex-wrap: wrap;
  gap: 0.2rem;
  margin-top: 0.15rem;
}

.die-badge {
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 3px;
  color: var(--muted);
  font-family: monospace;
  font-size: 0.75rem;
  padding: 0.05rem 0.25rem;
}

.roll-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
```

- [ ] **Step 6: Run all web tests — verify everything green**

```bash
cd web && npm test
```

Expected: ALL PASS.

- [ ] **Step 7: Run all Go tests — verify still green**

```bash
go test ./...
```

Expected: ALL PASS.

- [ ] **Step 8: Commit**

```bash
git add web/src/App.tsx web/src/App.test.tsx web/src/App.css
git commit -m "feat: portrait display, lastEvent wiring to panels, and tag/die badge styles"
```

---

## Self-Review

**Spec coverage check:**

| Spec requirement | Task |
|-----------------|------|
| WebSocket hook returns `{lastEvent}` | Task 4 |
| Portrait display via `/api/files/...` | Task 7 |
| `GET /api/files/{path}` endpoint | Task 2 |
| Dice breakdown badge rendering | Task 5 |
| `breakdown_json` parsed as `[d1][d2]` display | Task 5 |
| `?tag=npc` query param on world-notes route | Tasks 1 & 2 |
| Tag pills in WorldNotesPanel, click to filter | Task 6 |
| Click again to deselect | Task 6 |
| `update_world_note` MCP tool adds `tags` param | Task 3 |
| `world_note_updated` WS event refreshes panel | Task 6 |
| `dice_rolled` WS event refreshes dice panel | Task 5 |
| `PATCH /api/world-notes/{id}` route | Task 2 |
| File serving with path traversal protection | Task 2 |

**Placeholder scan:** No TBDs, all steps contain full code.

**Type consistency check:**
- `SearchWorldNotes(campaignID int64, query, category, tag string)` used consistently in Tasks 1, 2, 3
- `UpdateWorldNote(id int64, title, content, tagsJSON string)` used consistently in Tasks 1, 2, 3
- `DiceHistoryPanel` Props `{ sessionId: number, lastEvent: unknown }` defined in Task 5, consumed in Task 7
- `WorldNotesPanel` Props `{ campaignId: number, lastEvent: unknown }` defined in Task 6, consumed in Task 7
- `fetchWorldNotes(campaignId, q?, tag?)` defined in Task 6 Step 3, called in WorldNotesPanel Task 6 Step 7
