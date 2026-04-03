# Enhanced UI & Features Design

## Goal

Transform inkandbone from a passive display companion into a fully reactive, immersive TTRPG session tool — live combat tracking, clickable world maps, session timelines, editable character sheets, and rulebook ingestion — all driven by a WebSocket-first React frontend.

## Feature Scope

Eleven features organized across four implementation plans:

1. **Plan 1 — WebSocket Foundation + Quick Wins**: WebSocket reactivity, portrait display, dice breakdown, tag filtering
2. **Plan 2 — Live Play**: Combat tracker, session timeline
3. **Plan 3 — World Building**: Clickable world map, session journal, AI content generation
4. **Plan 4 — Deep Systems**: Character sheet, rulebook ingestion

## Architecture

### Shared Infrastructure Spine

```
MCP Tool fires
     │
     ▼
EventBus (internal/bus — already exists)
     │
     ▼
WebSocket Hub (internal/hub — already exists, broadcasts to all clients)
     │
     ▼
React useWebSocket hook (NEW — single WS connection per browser tab)
     │
     ├── CombatPanel        → combat_* events
     ├── DiceHistoryPanel   → dice_rolled events
     ├── WorldNotesPanel    → world_note_* events
     ├── SessionTimeline    → all session events
     └── CharacterSheetPanel → character_updated events
```

The WebSocket hub already broadcasts JSON to connected browser clients. The gap is entirely on the React side — no component currently opens a WebSocket connection. Plans 2–4 require no additional server-side WebSocket changes beyond the event emissions added in Plan 1.

### Write Endpoints

All 9 existing HTTP routes are read-only. The following `PATCH` and `POST` endpoints are added across plans:

| Plan | Method | Route | Purpose |
|------|--------|-------|---------|
| 1 | PATCH | `/api/world-notes/{id}` | Update note title/content/tags |
| 2 | GET | `/api/sessions/{id}/timeline` | Merged chronological feed |
| 3 | PATCH | `/api/sessions/{id}` | Update session summary |
| 3 | POST | `/api/campaigns/{id}/maps` | Upload map image |
| 3 | GET | `/api/maps/{id}` | Map metadata + image URL |
| 4 | PATCH | `/api/characters/{id}` | Update character data_json |
| 4 | POST | `/api/characters/{id}/portrait` | Upload portrait image |
| 4 | POST | `/api/rulesets/{id}/rulebook` | Upload/paste rulebook content |

### File Storage

Uploaded files (map images, character portraits, rulebook PDFs) are stored in a local `data/` directory relative to the binary. Paths saved in DB are relative (`maps/abc123.jpg`, `portraits/xyz.png`). Served via `GET /api/files/{path}` — a new static file endpoint with path traversal protection.

### New DB Table

```sql
CREATE TABLE rulebook_chunks (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  ruleset_id INTEGER NOT NULL REFERENCES rulesets(id),
  heading    TEXT NOT NULL,
  content    TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### WebSocket Event Types

All events are JSON `{"type": "<event>", "payload": {...}}`:

| Event | Fired by MCP tool | Payload |
|-------|------------------|---------|
| `dice_rolled` | `roll_dice` | `{session_id, expression, result, breakdown}` |
| `world_note_created` | `create_world_note` | `{id, campaign_id, title, category}` |
| `world_note_updated` | `update_world_note` | `{id, title, content, tags}` |
| `map_pin_added` | `add_map_pin` | `{map_id, id, x, y, label}` |
| `combatant_updated` | `update_combatant` | full combatant object |
| `combat_started` | `start_combat` | `{session_id, encounter_id, name}` |
| `combat_ended` | `end_combat` | `{encounter_id}` |
| `context_changed` | `set_active`, `start_session`, `end_session` | `{campaign_id, character_id, session_id}` |
| `session_updated` | `generate_session_recap` | `{session_id, summary}` |
| `character_updated` | write endpoint or MCP | `{id, data_json, portrait_path}` |

---

## Plan 1: WebSocket Foundation + Quick Wins

### WebSocket Hook

`web/src/useWebSocket.ts` — a single `useWebSocket(url)` hook:
- Opens one persistent `WebSocket` connection per browser tab
- Reconnects automatically on disconnect: exponential backoff starting at 1s, capped at 30s
- Returns `{lastEvent}` — the most recently received parsed JSON event
- Components call `useWebSocket('/ws')` and filter `lastEvent.type`

### Portrait Display

- `ctx.character.portrait_path` rendered as `<img>` in the context panel when non-empty
- Falls back to a text placeholder when empty
- URL: `/api/files/portraits/{filename}`

### Dice Breakdown

- `breakdown_json` already stored as JSON array of individual die results (e.g. `[4,3,18]` for `3d6`)
- DiceHistoryPanel renders each die as a small badge: `[4][3][18] = 25`
- Expression label shown above badge row

### Tag Filtering

- `tags_json` on world notes already stored
- New `?tag=npc` query param on `GET /api/campaigns/{id}/world-notes`
- Tag pills rendered in WorldNotesPanel — clicking filters; clicking again deselects
- Tags edited via new `update_world_note` MCP tool (adds `tags` param)
- `world_note_updated` WS event refreshes the panel in real time

### New MCP Tool: `update_world_note`

Parameters: `note_id`, `title` (optional), `content` (optional), `tags` (optional string array).
Fires `world_note_updated` WS event on success.

---

## Plan 2: Live Combat Tracker + Session Timeline

### CombatPanel

- Renders when `ctx.active_combat` is non-null
- Combatants sorted by `initiative` descending
- Each card: name, HP bar (`hp_current / hp_max`), initiative value, condition badges from `conditions_json`
- HP bar color: green (>50%), yellow (25–50%), red (<25%)
- Active turn highlighted (first in sorted order)
- Subscribes to `combatant_updated`, `combat_started`, `combat_ended` WS events — no polling
- Read-only: all combat mutations go through MCP tools

### SessionTimeline

- Vertically scrolling feed anchored to the current session
- On load: fetches `GET /api/sessions/{id}/timeline` — merged chronological slice of messages + dice rolls
- WS events append new entries in real time with a slide-in animation
- Entry types:
  - **Dice roll**: expression, result, breakdown badges
  - **World note event**: note title, category badge
  - **Combat event**: encounter name, start/end marker
  - **Context change**: session/character label

### New Server Endpoint: `GET /api/sessions/{id}/timeline`

Returns merged, chronologically sorted array of typed entries:
```json
[
  {"type": "dice_roll", "timestamp": "...", "data": {...}},
  {"type": "message", "timestamp": "...", "data": {...}},
  {"type": "combat_event", "timestamp": "...", "data": {...}}
]
```

---

## Plan 3: World Map + Session Journal + AI Content Generation

### MapPanel

- Renders when a map exists for the active campaign
- Image displayed at full panel width via `/api/files/maps/{filename}`
- Pins overlay as `<button>` elements positioned at `(x%, y%)` using absolute positioning within a `position: relative` container
- Clicking a pin opens a popover: linked world note title + content
- New pins appear via `map_pin_added` WS event (no reload)
- Map image uploaded via drag-and-drop or file picker → `POST /api/campaigns/{id}/maps`

### Session Journal (JournalPanel)

- Textarea bound to `session.summary`
- Auto-saves on blur via `PATCH /api/sessions/{id}`
- "Generate recap" button triggers new MCP tool `generate_session_recap`
- Claude reads session messages + dice rolls, writes narrative summary, calls `PATCH /api/sessions/{id}`
- UI updates via `session_updated` WS event

### AI Content Generation

- "Draft with Claude" button in WorldNotesPanel toolbar
- Opens a small prompt input: user types a hint (e.g. "Elven blacksmith NPC")
- Calls new MCP tool `draft_world_note(campaign_id, hint)`
- Claude generates title + content, calls `create_world_note`
- `world_note_created` WS event makes it appear instantly in the panel

### New MCP Tools

- `generate_session_recap(session_id)` — reads messages/rolls, writes summary
- `draft_world_note(campaign_id, hint)` — generates and creates a world note

---

## Plan 4: Character Sheet + Rulebook Ingestion

### CharacterSheetPanel

- Renders when `ctx.character` is non-null
- Fields driven by `ctx.campaign.ruleset_id` → fetched ruleset `schema_json`
- Schema fields rendered as labeled inputs: text for strings, number for numeric fields, textarea for `notes`/`features`/`spells`
- Changes debounce 500ms → `PATCH /api/characters/{id}` (updates `data_json`)
- `character_updated` WS event syncs back to context panel
- Portrait: clickable `<img>` or placeholder → file input → `POST /api/characters/{id}/portrait`

### Rulebook Ingestion

Two ingestion modes, same storage:

**Paste mode** (settings panel):
- Textarea accepts plain text or Markdown rules content
- POST to `POST /api/rulesets/{id}/rulebook` with `Content-Type: text/plain`
- Server splits on heading lines (`#`, `##`), stores chunks in `rulebook_chunks`

**PDF mode**:
- File upload to same endpoint with `Content-Type: multipart/form-data`
- Server extracts text using `pdfcpu` (pure Go, no external deps)
- Same chunking logic applied to extracted text

**MCP Search Tool: `search_rulebook(query, ruleset_id)`**:
- Full-text LIKE search across `rulebook_chunks.content`
- Returns top 3 matching chunks (heading + content)
- Claude calls this proactively when adjudicating rules
- No vector embeddings — keyword search is sufficient for precise rules terminology

### New DB Query

`SearchRulebookChunks(rulesetID int64, query string) ([]RulebookChunk, error)` — SQLite `LIKE '%query%'` across `heading` and `content` columns, `LIMIT 3`.

---

## Testing Strategy

### Go (existing pattern)
- Route tests: `httptest.NewRequest` + `httptest.NewRecorder` pattern, seeded DB
- DB tests: in-memory SQLite, one test per query function
- WS event emission: assert event published to bus after MCP tool call

### TypeScript (existing pattern)
- `vitest` + `vi.stubGlobal('fetch', ...)` for API function tests
- `@testing-library/react` for component render tests
- WS hook: stub `WebSocket` global, emit mock events, assert component state

### Integration
- Smoketest: build binary, seed data via MCP tools, verify HTTP endpoints return expected shapes

---

## Rulebook Ingestion: System Compatibility

| Ruleset | Claude knowledge | Ingestion priority |
|---------|-----------------|-------------------|
| D&D 5e | Excellent | Low — rarely needed |
| Call of Cthulhu 7e | Excellent | Low |
| Ironsworn | Good | Low |
| Vampire the Masquerade V20 | Good | Medium — edition nuances |
| Cyberpunk Red | Moderate | High — NET architecture specifics |

Ingestion is opt-in per ruleset — users upload only what they need.
