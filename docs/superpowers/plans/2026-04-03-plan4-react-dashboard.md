# React Game Dashboard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the Vite placeholder `App.tsx` with a real-time game dashboard that shows active campaign/session/character state, the session message log, and an active combat tracker, driven by the existing `/api/*` endpoints and WebSocket hub.

**Architecture:** Add `json` struct tags (snake_case) to all DB types so the existing HTTP handlers serialize correctly; add a `GET /api/context` endpoint that returns the same game-state snapshot as the MCP `get_context` tool; wire up a React dashboard with a `useWebSocket` hook that re-fetches context on every event.

**Tech Stack:** Go standard library, React 19, TypeScript 5, Vite 8, vitest, @testing-library/react, jsdom

---

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `internal/db/queries_core.go` | JSON tags on Ruleset, Campaign, Character |
| Modify | `internal/db/queries_session.go` | JSON tags on Session, Message, DiceRoll |
| Modify | `internal/db/queries_combat.go` | JSON tags on CombatEncounter, Combatant |
| Modify | `internal/db/queries_world.go` | JSON tags on WorldNote, Map, MapPin |
| Create | `internal/db/json_tags_test.go` | Verify snake_case JSON output |
| Modify | `internal/api/routes.go` | `handleGetContext` handler + `contextResponse` types |
| Modify | `internal/api/routes_test.go` | `TestGetContext_*` tests |
| Modify | `internal/api/server.go` | Register `GET /api/context` route |
| Modify | `web/package.json` | Add vitest + testing-library dev deps + `test` script |
| Modify | `web/tsconfig.app.json` | Add `@testing-library/jest-dom` to `types` |
| Modify | `web/vite.config.ts` | Add dev proxy + vitest test config |
| Create | `web/src/test-setup.ts` | Import `@testing-library/jest-dom/vitest` |
| Create | `web/src/types.ts` | TypeScript interfaces matching DB struct JSON shapes |
| Create | `web/src/api.ts` | `fetchContext()` function |
| Create | `web/src/api.test.ts` | Unit tests for api.ts |
| Create | `web/src/useWebSocket.ts` | WebSocket hook with auto-reconnect |
| Create | `web/src/useWebSocket.test.tsx` | Unit tests for hook |
| Replace | `web/src/App.tsx` | Game dashboard (replaces Vite placeholder) |
| Replace | `web/src/App.css` | Dark TTRPG-themed styles |
| Create | `web/src/App.test.tsx` | Smoke tests for App |

---

### Task 1: JSON Tags on DB Structs

**Files:**
- Create: `internal/db/json_tags_test.go`
- Modify: `internal/db/queries_core.go:29-34,77-84,136-143`
- Modify: `internal/db/queries_session.go:8-15,76-82,156-162`
- Modify: `internal/db/queries_combat.go:8-14,64-74`
- Modify: `internal/db/queries_world.go:10-17,80-85,112-120`

- [ ] **Step 1: Write the failing test**

Create `internal/db/json_tags_test.go`:

```go
package db

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCampaignJSONKeys(t *testing.T) {
	c := Campaign{ID: 1, RulesetID: 2, Name: "Greyhawk", Description: "desc", Active: true, CreatedAt: "2026-04-03"}
	b, err := json.Marshal(c)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(b, &m))
	assert.Contains(t, m, "id")
	assert.Contains(t, m, "ruleset_id")
	assert.Contains(t, m, "name")
	assert.Contains(t, m, "description")
	assert.Contains(t, m, "active")
	assert.Contains(t, m, "created_at")
	assert.NotContains(t, m, "ID")
	assert.NotContains(t, m, "Name")
}

func TestMessageJSONKeys(t *testing.T) {
	msg := Message{ID: 1, SessionID: 2, Role: "user", Content: "hello", CreatedAt: "2026-04-03"}
	b, err := json.Marshal(msg)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(b, &m))
	assert.Contains(t, m, "id")
	assert.Contains(t, m, "session_id")
	assert.Contains(t, m, "role")
	assert.Contains(t, m, "content")
	assert.Contains(t, m, "created_at")
	assert.NotContains(t, m, "SessionID")
}

func TestCombatantJSONKeys(t *testing.T) {
	c := Combatant{ID: 1, EncounterID: 2, Name: "Goblin", Initiative: 12, HPCurrent: 7, HPMax: 7, ConditionsJSON: "[]", IsPlayer: false}
	b, err := json.Marshal(c)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(b, &m))
	assert.Contains(t, m, "id")
	assert.Contains(t, m, "encounter_id")
	assert.Contains(t, m, "initiative")
	assert.Contains(t, m, "hp_current")
	assert.Contains(t, m, "hp_max")
	assert.Contains(t, m, "is_player")
	assert.NotContains(t, m, "HPCurrent")
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /path/to/repo && go test ./internal/db/ -run TestCampaignJSONKeys -v
```

Expected: FAIL — keys are `"ID"`, `"Name"` (PascalCase), not `"id"`, `"name"`.

- [ ] **Step 3: Add JSON tags to all DB structs**

In `internal/db/queries_core.go`, replace the three struct definitions:

```go
type Ruleset struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	SchemaJSON string `json:"schema_json"`
	Version    string `json:"version"`
}
```

```go
type Campaign struct {
	ID          int64  `json:"id"`
	RulesetID   int64  `json:"ruleset_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
	CreatedAt   string `json:"created_at"`
}
```

```go
type Character struct {
	ID           int64  `json:"id"`
	CampaignID   int64  `json:"campaign_id"`
	Name         string `json:"name"`
	DataJSON     string `json:"data_json"`
	PortraitPath string `json:"portrait_path"` // NOT NULL DEFAULT '' in schema; never nil
	CreatedAt    string `json:"created_at"`
}
```

In `internal/db/queries_session.go`, replace the three struct definitions:

```go
type Session struct {
	ID         int64  `json:"id"`
	CampaignID int64  `json:"campaign_id"`
	Title      string `json:"title"`
	Date       string `json:"date"`
	Summary    string `json:"summary"`
	CreatedAt  string `json:"created_at"`
}
```

```go
type Message struct {
	ID        int64  `json:"id"`
	SessionID int64  `json:"session_id"`
	Role      string `json:"role"` // "user" or "assistant"
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}
```

```go
type DiceRoll struct {
	ID            int64  `json:"id"`
	SessionID     int64  `json:"session_id"`
	Expression    string `json:"expression"`
	Result        int    `json:"result"`
	BreakdownJSON string `json:"breakdown_json"`
	CreatedAt     string `json:"created_at"`
}
```

In `internal/db/queries_combat.go`, replace the two struct definitions:

```go
type CombatEncounter struct {
	ID        int64  `json:"id"`
	SessionID int64  `json:"session_id"`
	Name      string `json:"name"`
	Active    bool   `json:"active"`
	CreatedAt string `json:"created_at"`
}
```

```go
type Combatant struct {
	ID             int64  `json:"id"`
	EncounterID    int64  `json:"encounter_id"`
	CharacterID    *int64 `json:"character_id"`
	Name           string `json:"name"`
	Initiative     int    `json:"initiative"`
	HPCurrent      int    `json:"hp_current"`
	HPMax          int    `json:"hp_max"`
	ConditionsJSON string `json:"conditions_json"`
	IsPlayer       bool   `json:"is_player"`
}
```

In `internal/db/queries_world.go`, replace the three struct definitions:

```go
type WorldNote struct {
	ID         int64  `json:"id"`
	CampaignID int64  `json:"campaign_id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Category   string `json:"category"`
	TagsJSON   string `json:"tags_json"`
	CreatedAt  string `json:"created_at"`
}
```

```go
type Map struct {
	ID         int64  `json:"id"`
	CampaignID int64  `json:"campaign_id"`
	Name       string `json:"name"`
	ImagePath  string `json:"image_path"`
	CreatedAt  string `json:"created_at"`
}
```

```go
type MapPin struct {
	ID        int64   `json:"id"`
	MapID     int64   `json:"map_id"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Label     string  `json:"label"`
	Note      string  `json:"note"`
	Color     string  `json:"color"`
	CreatedAt string  `json:"created_at"`
}
```

- [ ] **Step 4: Run all tests to verify they pass**

```bash
go test ./internal/db/ ./internal/api/ ./internal/mcp/ -v 2>&1 | tail -20
```

Expected: all PASS (the JSON tag tests plus all existing tests — tags are additive and don't change scan behavior).

- [ ] **Step 5: Commit**

```bash
git add internal/db/
git commit -m "feat: add snake_case json tags to all DB structs"
```

---

### Task 2: GET /api/context Endpoint

**Files:**
- Modify: `internal/api/routes.go` (add types + handler at bottom)
- Modify: `internal/api/routes_test.go` (add two tests at bottom)
- Modify: `internal/api/server.go:59` (add route registration)

- [ ] **Step 1: Write the failing tests**

Append to `internal/api/routes_test.go`:

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/api/ -run TestGetContext -v
```

Expected: FAIL — `contextResponse` undefined, route 404.

- [ ] **Step 3: Add types and handler to routes.go**

Append to `internal/api/routes.go`:

```go
type contextCombatSnapshot struct {
	Encounter  *db.CombatEncounter `json:"encounter"`
	Combatants []db.Combatant      `json:"combatants"`
}

type contextResponse struct {
	Campaign       *db.Campaign           `json:"campaign"`
	Character      *db.Character          `json:"character"`
	Session        *db.Session            `json:"session"`
	RecentMessages []db.Message           `json:"recent_messages"`
	ActiveCombat   *contextCombatSnapshot `json:"active_combat"`
}

func (s *Server) handleGetContext(w http.ResponseWriter, _ *http.Request) {
	resp := contextResponse{RecentMessages: []db.Message{}}

	if campIDStr, err := s.db.GetSetting("active_campaign_id"); err == nil && campIDStr != "" {
		if campID, err := strconv.ParseInt(campIDStr, 10, 64); err == nil {
			resp.Campaign, _ = s.db.GetCampaign(campID)
		}
	}
	if charIDStr, err := s.db.GetSetting("active_character_id"); err == nil && charIDStr != "" {
		if charID, err := strconv.ParseInt(charIDStr, 10, 64); err == nil {
			resp.Character, _ = s.db.GetCharacter(charID)
		}
	}
	if sessIDStr, err := s.db.GetSetting("active_session_id"); err == nil && sessIDStr != "" {
		if sessID, err := strconv.ParseInt(sessIDStr, 10, 64); err == nil {
			resp.Session, _ = s.db.GetSession(sessID)
			if msgs, err := s.db.RecentMessages(sessID, 20); err == nil {
				resp.RecentMessages = msgs
			}
			if enc, err := s.db.GetActiveEncounter(sessID); err == nil && enc != nil {
				cs := &contextCombatSnapshot{Encounter: enc}
				if combatants, err := s.db.ListCombatants(enc.ID); err == nil {
					cs.Combatants = combatants
				}
				resp.ActiveCombat = cs
			}
		}
	}

	writeJSON(w, resp)
}
```

- [ ] **Step 4: Register the route in server.go**

In `internal/api/server.go`, inside `registerRoutes()`, add after the existing routes:

```go
s.mux.HandleFunc("GET /api/context", s.handleGetContext)
```

The full `registerRoutes` should look like:

```go
func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/ws", s.hub.ServeWS)
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("GET /api/campaigns", s.handleListCampaigns)
	s.mux.HandleFunc("GET /api/campaigns/{id}/characters", s.handleListCharacters)
	s.mux.HandleFunc("GET /api/campaigns/{id}/sessions", s.handleListSessions)
	s.mux.HandleFunc("GET /api/sessions/{id}/messages", s.handleListMessages)
	s.mux.HandleFunc("GET /api/sessions/{id}/dice-rolls", s.handleListDiceRolls)
	s.mux.HandleFunc("GET /api/maps/{id}/pins", s.handleListMapPins)
	s.mux.HandleFunc("GET /api/context", s.handleGetContext)
}
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
go test ./internal/api/ -run TestGetContext -v
```

Expected: PASS for both `TestGetContext_empty` and `TestGetContext_withActiveState`.

- [ ] **Step 6: Run full test suite to verify no regressions**

```bash
go test ./internal/... -v 2>&1 | grep -E "^(ok|FAIL|---)"
```

Expected: all `ok`, no `FAIL`.

- [ ] **Step 7: Commit**

```bash
git add internal/api/routes.go internal/api/routes_test.go internal/api/server.go
git commit -m "feat: add GET /api/context HTTP endpoint"
```

---

### Task 3: Frontend Testing Setup + TypeScript Types + API Client

**Files:**
- Modify: `web/package.json`
- Modify: `web/tsconfig.app.json`
- Modify: `web/vite.config.ts`
- Create: `web/src/test-setup.ts`
- Create: `web/src/types.ts`
- Create: `web/src/api.ts`
- Create: `web/src/api.test.ts`

- [ ] **Step 1: Install test dependencies**

```bash
cd web && npm install -D vitest @testing-library/react @testing-library/jest-dom @testing-library/user-event jsdom
```

Expected: `package-lock.json` updated; `node_modules` has `vitest/`, `@testing-library/`.

- [ ] **Step 2: Add `test` script to package.json**

In `web/package.json`, replace the `"scripts"` block:

```json
"scripts": {
  "dev": "vite",
  "build": "tsc -b && vite build",
  "lint": "eslint .",
  "preview": "vite preview",
  "test": "vitest run"
},
```

- [ ] **Step 3: Update vite.config.ts with proxy and test config**

Replace `web/vite.config.ts` entirely:

```ts
/// <reference types="vitest" />
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': 'http://localhost:7432',
      '/ws': { target: 'ws://localhost:7432', ws: true },
    },
  },
  test: {
    environment: 'jsdom',
    setupFiles: ['./src/test-setup.ts'],
  },
})
```

The `/// <reference types="vitest" />` directive tells TypeScript that the `test` property is valid in `defineConfig`.

- [ ] **Step 4: Add @testing-library/jest-dom to tsconfig types**

In `web/tsconfig.app.json`, replace the `"types"` line:

```json
"types": ["vite/client", "@testing-library/jest-dom"],
```

- [ ] **Step 5: Create test-setup.ts**

Create `web/src/test-setup.ts`:

```ts
import '@testing-library/jest-dom/vitest'
```

- [ ] **Step 6: Write failing tests for api.ts**

Create `web/src/api.test.ts`:

```ts
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { fetchContext } from './api'

describe('fetchContext', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
  })

  it('returns parsed GameContext on success', async () => {
    const mockCtx = {
      campaign: null,
      character: null,
      session: null,
      recent_messages: [],
      active_combat: null,
    }
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(mockCtx),
    }))
    const result = await fetchContext()
    expect(result).toEqual(mockCtx)
  })

  it('throws on non-ok response', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: false,
      status: 500,
    }))
    await expect(fetchContext()).rejects.toThrow('HTTP 500')
  })
})
```

- [ ] **Step 7: Run test to verify it fails**

```bash
cd web && npm test -- --reporter=verbose 2>&1 | head -30
```

Expected: FAIL — `Cannot find module './api'`.

- [ ] **Step 8: Create types.ts**

Create `web/src/types.ts`:

```ts
export interface Campaign {
  id: number
  ruleset_id: number
  name: string
  description: string
  active: boolean
  created_at: string
}

export interface Character {
  id: number
  campaign_id: number
  name: string
  data_json: string
  portrait_path: string
  created_at: string
}

export interface Session {
  id: number
  campaign_id: number
  title: string
  date: string
  summary: string
  created_at: string
}

export interface Message {
  id: number
  session_id: number
  role: string
  content: string
  created_at: string
}

export interface CombatEncounter {
  id: number
  session_id: number
  name: string
  active: boolean
  created_at: string
}

export interface Combatant {
  id: number
  encounter_id: number
  character_id: number | null
  name: string
  initiative: number
  hp_current: number
  hp_max: number
  conditions_json: string
  is_player: boolean
}

export interface CombatSnapshot {
  encounter: CombatEncounter
  combatants: Combatant[]
}

export interface GameContext {
  campaign: Campaign | null
  character: Character | null
  session: Session | null
  recent_messages: Message[]
  active_combat: CombatSnapshot | null
}
```

- [ ] **Step 9: Create api.ts**

Create `web/src/api.ts`:

```ts
import type { GameContext } from './types'

export async function fetchContext(): Promise<GameContext> {
  const res = await fetch('/api/context')
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  return res.json() as Promise<GameContext>
}
```

- [ ] **Step 10: Run tests to verify they pass**

```bash
cd web && npm test -- --reporter=verbose
```

Expected: `fetchContext > returns parsed GameContext on success` PASS, `fetchContext > throws on non-ok response` PASS.

- [ ] **Step 11: Commit**

```bash
cd web && git add package.json tsconfig.app.json vite.config.ts src/test-setup.ts src/types.ts src/api.ts src/api.test.ts package-lock.json
git commit -m "feat: frontend testing setup, types, and api client"
```

---

### Task 4: useWebSocket Hook

**Files:**
- Create: `web/src/useWebSocket.ts`
- Create: `web/src/useWebSocket.test.tsx`

- [ ] **Step 1: Write the failing tests**

Create `web/src/useWebSocket.test.tsx`:

```tsx
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useWebSocket } from './useWebSocket'
import type { WSEvent } from './useWebSocket'

class MockWebSocket {
  static instances: MockWebSocket[] = []
  url: string
  onmessage: ((e: MessageEvent) => void) | null = null
  onclose: (() => void) | null = null
  onerror: (() => void) | null = null

  constructor(url: string) {
    this.url = url
    MockWebSocket.instances.push(this)
  }

  close = vi.fn(() => {
    if (this.onclose) this.onclose()
  })

  simulateMessage(data: string) {
    this.onmessage?.({ data } as MessageEvent)
  }

  simulateError() {
    this.onerror?.()
  }
}

describe('useWebSocket', () => {
  beforeEach(() => {
    MockWebSocket.instances = []
    vi.stubGlobal('WebSocket', MockWebSocket)
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.unstubAllGlobals()
  })

  it('calls onEvent when a valid JSON message is received', () => {
    const handler = vi.fn<[WSEvent], void>()
    renderHook(() => useWebSocket('ws://localhost/ws', handler))
    const ws = MockWebSocket.instances[0]
    act(() => {
      ws.simulateMessage(JSON.stringify({ type: 'session_started', payload: { session_id: 1 } }))
    })
    expect(handler).toHaveBeenCalledWith({ type: 'session_started', payload: { session_id: 1 } })
  })

  it('ignores malformed JSON messages without throwing', () => {
    const handler = vi.fn<[WSEvent], void>()
    renderHook(() => useWebSocket('ws://localhost/ws', handler))
    const ws = MockWebSocket.instances[0]
    act(() => {
      ws.simulateMessage('not valid json')
    })
    expect(handler).not.toHaveBeenCalled()
  })

  it('closes WebSocket on unmount without reconnecting', () => {
    const handler = vi.fn<[WSEvent], void>()
    const { unmount } = renderHook(() => useWebSocket('ws://localhost/ws', handler))
    const ws = MockWebSocket.instances[0]
    unmount()
    // close was called
    expect(ws.close).toHaveBeenCalled()
    // no reconnect timer fires
    vi.runAllTimers()
    expect(MockWebSocket.instances).toHaveLength(1)
  })
})
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd web && npm test -- --reporter=verbose 2>&1 | head -30
```

Expected: FAIL — `Cannot find module './useWebSocket'`.

- [ ] **Step 3: Create useWebSocket.ts**

Create `web/src/useWebSocket.ts`:

```ts
import { useEffect, useRef, useCallback } from 'react'

export type WSEvent = {
  type: string
  payload: unknown
}

export type WSHandler = (event: WSEvent) => void

export function useWebSocket(url: string, onEvent: WSHandler): void {
  const onEventRef = useRef(onEvent)
  onEventRef.current = onEvent
  const unmountedRef = useRef(false)

  const connect = useCallback(() => {
    const ws = new WebSocket(url)

    ws.onmessage = (e) => {
      try {
        const event = JSON.parse(e.data as string) as WSEvent
        onEventRef.current(event)
      } catch {
        // ignore malformed messages
      }
    }

    ws.onclose = () => {
      if (!unmountedRef.current) {
        setTimeout(connect, 2000)
      }
    }

    ws.onerror = () => {
      ws.close()
    }

    return ws
  }, [url])

  useEffect(() => {
    unmountedRef.current = false
    const ws = connect()
    return () => {
      unmountedRef.current = true
      ws.onclose = null
      ws.close()
    }
  }, [connect])
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd web && npm test -- --reporter=verbose
```

Expected: all 3 `useWebSocket` tests PASS plus the 2 `fetchContext` tests PASS (5 total).

- [ ] **Step 5: Commit**

```bash
git add web/src/useWebSocket.ts web/src/useWebSocket.test.tsx
git commit -m "feat: useWebSocket hook with auto-reconnect"
```

---

### Task 5: App.tsx Game Dashboard

**Files:**
- Create: `web/src/App.test.tsx`
- Replace: `web/src/App.tsx`
- Replace: `web/src/App.css`

- [ ] **Step 1: Write the failing tests**

Create `web/src/App.test.tsx`:

```tsx
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import App from './App'
import type { GameContext } from './types'

const mockCtx: GameContext = {
  campaign: { id: 1, ruleset_id: 1, name: 'Greyhawk', description: '', active: true, created_at: '' },
  character: { id: 1, campaign_id: 1, name: 'Zara', data_json: '{}', portrait_path: '', created_at: '' },
  session: { id: 1, campaign_id: 1, title: 'Session 1', date: '2026-04-03', summary: '', created_at: '' },
  recent_messages: [
    { id: 1, session_id: 1, role: 'assistant', content: 'You enter the tavern.', created_at: '' },
    { id: 2, session_id: 1, role: 'user', content: 'I look for a table.', created_at: '' },
  ],
  active_combat: null,
}

class MockWebSocket {
  onmessage = null; onclose: (() => void) | null = null; onerror = null
  close = vi.fn()
}

describe('App', () => {
  beforeEach(() => {
    vi.stubGlobal('WebSocket', MockWebSocket)
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(mockCtx),
    }))
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('renders campaign name in state bar', async () => {
    render(<App />)
    expect(await screen.findByText('Greyhawk')).toBeInTheDocument()
  })

  it('renders character name in state bar', async () => {
    render(<App />)
    expect(await screen.findByText('Zara')).toBeInTheDocument()
  })

  it('renders session title in state bar', async () => {
    render(<App />)
    expect(await screen.findByText('Session 1')).toBeInTheDocument()
  })

  it('renders session log messages', async () => {
    render(<App />)
    expect(await screen.findByText('You enter the tavern.')).toBeInTheDocument()
    expect(await screen.findByText('I look for a table.')).toBeInTheDocument()
  })
})
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd web && npm test -- --reporter=verbose 2>&1 | grep -E "(FAIL|PASS|Error)"
```

Expected: FAIL on all 4 App tests — App still renders Vite placeholder content.

- [ ] **Step 3: Replace App.tsx with the dashboard**

Replace `web/src/App.tsx` entirely:

```tsx
import { useState, useEffect, useCallback } from 'react'
import { useWebSocket } from './useWebSocket'
import type { WSEvent } from './useWebSocket'
import { fetchContext } from './api'
import type { GameContext, Message } from './types'
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
    (_event: WSEvent) => {
      loadContext()
    },
    [loadContext],
  )

  useWebSocket(WS_URL, handleEvent)

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
      </main>
    </div>
  )
}
```

- [ ] **Step 4: Replace App.css with dark theme styles**

Replace `web/src/App.css` entirely:

```css
:root {
  --bg: #1a1a2e;
  --surface: #16213e;
  --accent: #e94560;
  --text: #e0e0e0;
  --muted: #888;
  --border: #2a2a4a;
  --player: #6bffb8;
  --enemy: #ff6b6b;
}

* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

body {
  background: var(--bg);
  color: var(--text);
  font-family: system-ui, sans-serif;
  min-height: 100vh;
}

.dashboard {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

.state-bar {
  background: var(--surface);
  border-bottom: 1px solid var(--border);
  padding: 0.75rem 1.5rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-size: 0.9rem;
}

.state-bar .campaign {
  font-weight: 600;
  color: var(--accent);
}

.state-bar .separator {
  color: var(--muted);
}

.panels {
  display: flex;
  gap: 1rem;
  padding: 1rem;
  flex: 1;
  overflow: hidden;
}

.panel {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 1rem;
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.panel h2 {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--muted);
  margin-bottom: 0.25rem;
}

.message {
  display: flex;
  gap: 0.75rem;
  font-size: 0.88rem;
  line-height: 1.5;
}

.message .role {
  min-width: 4.5rem;
  font-size: 0.7rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding-top: 0.2rem;
  color: var(--accent);
  flex-shrink: 0;
}

.message.user .role {
  color: #4a9eff;
}

.empty {
  color: var(--muted);
  font-size: 0.85rem;
}

.combat {
  max-width: 340px;
  flex: none;
}

.combat table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.85rem;
}

.combat th {
  text-align: left;
  color: var(--muted);
  border-bottom: 1px solid var(--border);
  padding: 0.3rem 0.5rem;
  font-size: 0.75rem;
  font-weight: 600;
}

.combat td {
  padding: 0.4rem 0.5rem;
}

.combat tr.player td {
  color: var(--player);
}

.combat tr.enemy td {
  color: var(--enemy);
}

.error {
  color: var(--enemy);
  padding: 2rem;
  text-align: center;
}

.loading {
  color: var(--muted);
  padding: 2rem;
  text-align: center;
}
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
cd web && npm test -- --reporter=verbose
```

Expected: all tests PASS — 2 `fetchContext`, 3 `useWebSocket`, 4 `App`. Output should show `9 passed`.

- [ ] **Step 6: Run TypeScript type-check**

```bash
cd web && npx tsc -b --noEmit 2>&1
```

Expected: no errors.

- [ ] **Step 7: Commit**

```bash
git add web/src/App.tsx web/src/App.css web/src/App.test.tsx
git commit -m "feat: replace Vite placeholder with game dashboard"
```

---

## Self-Review

### Spec Coverage

| Requirement | Task |
|-------------|------|
| JSON tags snake_case on DB structs | Task 1 |
| GET /api/context endpoint | Task 2 |
| Frontend test infrastructure | Task 3 |
| TypeScript types matching JSON shapes | Task 3 |
| fetchContext API client | Task 3 |
| useWebSocket hook | Task 4 |
| Dashboard: state bar (campaign/character/session) | Task 5 |
| Dashboard: session message log | Task 5 |
| Dashboard: active combat tracker | Task 5 |
| Real-time updates via WebSocket | Task 5 (re-fetch on event) |

### Type Consistency

- `GameContext` defined in `types.ts` (Task 3), used in `api.ts` (Task 3) and `App.tsx` (Task 5)
- `WSEvent` defined in `useWebSocket.ts` (Task 4), re-exported and used in `App.tsx` (Task 5)
- `Message` from `types.ts` used in `App.tsx` for message list state
- `contextResponse` in Go (`routes.go`) matches `GameContext` in TypeScript — both have `campaign`, `character`, `session`, `recent_messages`, `active_combat`
- `CombatSnapshot` in TypeScript matches `contextCombatSnapshot` in Go
