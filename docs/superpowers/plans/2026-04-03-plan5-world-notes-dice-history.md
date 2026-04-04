# World Notes & Dice History Panels Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a world notes HTTP endpoint and two new React dashboard panels — a searchable world notes list and a dice roll history — displayed alongside the existing session log and combat tracker.

**Architecture:** Wire a new `GET /api/campaigns/{id}/world-notes` endpoint using the existing `db.SearchWorldNotes` method; add two self-contained React components (`WorldNotesPanel`, `DiceHistoryPanel`) that each manage their own fetch state; integrate them into `App.tsx` conditioned on active campaign/session.

**Tech Stack:** Go standard library, React 19, TypeScript 5, Vite 8, vitest, @testing-library/react, @testing-library/user-event, jsdom

---

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `internal/api/routes.go` | Add `handleListWorldNotes` handler |
| Modify | `internal/api/server.go` | Register `GET /api/campaigns/{id}/world-notes` route |
| Modify | `internal/api/routes_test.go` | Tests for the world notes endpoint |
| Modify | `web/src/types.ts` | Add `WorldNote` and `DiceRoll` interfaces |
| Modify | `web/src/api.ts` | Add `fetchWorldNotes` and `fetchDiceRolls` functions |
| Modify | `web/src/api.test.ts` | Tests for the two new API functions |
| Create | `web/src/WorldNotesPanel.tsx` | Searchable world notes panel component |
| Create | `web/src/WorldNotesPanel.test.tsx` | Tests for `WorldNotesPanel` |
| Create | `web/src/DiceHistoryPanel.tsx` | Dice roll history panel component |
| Create | `web/src/DiceHistoryPanel.test.tsx` | Tests for `DiceHistoryPanel` |
| Modify | `web/src/App.tsx` | Import and render `WorldNotesPanel` and `DiceHistoryPanel` |
| Modify | `web/src/App.css` | Styles for world-notes and dice-history panels |
| Modify | `web/src/App.test.tsx` | Update fetch mock to handle multiple endpoint URLs |

---

### Task 1: World Notes HTTP Endpoint

**Files:**
- Modify: `internal/api/routes.go` — add `handleListWorldNotes` at the bottom of the file
- Modify: `internal/api/server.go:61` — register the route inside `registerRoutes`
- Modify: `internal/api/routes_test.go` — add three new test functions after the existing tests

- [ ] **Step 1: Write the failing tests**

Append to `internal/api/routes_test.go` (after `TestGetContext_withActiveState`):

```go
func TestListWorldNotes_empty(t *testing.T) {
	s := newTestServer(t)
	campID, _ := seedCampaign(t, s.db)
	req := httptest.NewRequest(http.MethodGet, "/api/campaigns/"+strconv.FormatInt(campID, 10)+"/world-notes", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var notes []db.WorldNote
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &notes))
	assert.Empty(t, notes)
}

func TestListWorldNotes_withData(t *testing.T) {
	s := newTestServer(t)
	campID, _ := seedCampaign(t, s.db)
	_, err := s.db.CreateWorldNote(campID, "Tavern", "A seedy place.", "location")
	require.NoError(t, err)
	_, err = s.db.CreateWorldNote(campID, "Dragon", "Ancient red dragon.", "npc")
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodGet, "/api/campaigns/"+strconv.FormatInt(campID, 10)+"/world-notes", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var notes []db.WorldNote
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &notes))
	require.Len(t, notes, 2)
}

func TestListWorldNotes_searchFilter(t *testing.T) {
	s := newTestServer(t)
	campID, _ := seedCampaign(t, s.db)
	_, err := s.db.CreateWorldNote(campID, "Tavern", "A seedy place.", "location")
	require.NoError(t, err)
	_, err = s.db.CreateWorldNote(campID, "Dragon", "Ancient red dragon.", "npc")
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodGet, "/api/campaigns/"+strconv.FormatInt(campID, 10)+"/world-notes?q=tavern", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var notes []db.WorldNote
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &notes))
	require.Len(t, notes, 1)
	assert.Equal(t, "Tavern", notes[0].Title)
}
```

- [ ] **Step 2: Run tests to verify they fail**

```
cd /home/digitalghost/projects/inkandbone
go test ./internal/api/... -run TestListWorldNotes -v
```

Expected: `FAIL` — `handleListWorldNotes` is not defined yet, route returns 404.

- [ ] **Step 3: Add the handler**

Append to the bottom of `internal/api/routes.go` (after `handleGetContext`):

```go
func (s *Server) handleListWorldNotes(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}
	q := r.URL.Query().Get("q")
	category := r.URL.Query().Get("category")
	notes, err := s.db.SearchWorldNotes(id, q, category)
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

- [ ] **Step 4: Register the route**

In `internal/api/server.go`, inside `registerRoutes()`, add after the existing `GET /api/maps/{id}/pins` line:

```go
s.mux.HandleFunc("GET /api/campaigns/{id}/world-notes", s.handleListWorldNotes)
```

The `registerRoutes` function should now look like:

```go
func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/ws", s.hub.ServeWS)
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("GET /api/campaigns", s.handleListCampaigns)
	s.mux.HandleFunc("GET /api/campaigns/{id}/characters", s.handleListCharacters)
	s.mux.HandleFunc("GET /api/campaigns/{id}/sessions", s.handleListSessions)
	s.mux.HandleFunc("GET /api/campaigns/{id}/world-notes", s.handleListWorldNotes)
	s.mux.HandleFunc("GET /api/sessions/{id}/messages", s.handleListMessages)
	s.mux.HandleFunc("GET /api/sessions/{id}/dice-rolls", s.handleListDiceRolls)
	s.mux.HandleFunc("GET /api/maps/{id}/pins", s.handleListMapPins)
	s.mux.HandleFunc("GET /api/context", s.handleGetContext)
}
```

- [ ] **Step 5: Run tests to verify they pass**

```
go test ./internal/api/... -run TestListWorldNotes -v
```

Expected: `PASS` — all three `TestListWorldNotes_*` tests green.

- [ ] **Step 6: Run full Go test suite**

```
go test ./...
```

Expected: all tests pass, no failures.

- [ ] **Step 7: Commit**

```bash
git add internal/api/routes.go internal/api/server.go internal/api/routes_test.go
git commit -m "feat: add GET /api/campaigns/{id}/world-notes endpoint"
```

---

### Task 2: TypeScript Types and API Functions

**Files:**
- Modify: `web/src/types.ts` — add `WorldNote` and `DiceRoll` interfaces
- Modify: `web/src/api.ts` — add `fetchWorldNotes` and `fetchDiceRolls`
- Modify: `web/src/api.test.ts` — add tests for the two new functions

- [ ] **Step 1: Write the failing tests**

Replace the entire contents of `web/src/api.test.ts` with:

```typescript
import { describe, it, expect, vi, afterEach } from 'vitest'
import { fetchContext, fetchWorldNotes, fetchDiceRolls } from './api'

afterEach(() => vi.restoreAllMocks())

describe('fetchContext', () => {
  it('returns parsed GameContext on success', async () => {
    const payload = {
      campaign: null,
      character: null,
      session: null,
      recent_messages: [],
      active_combat: null,
    }
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(payload),
    }))

    const ctx = await fetchContext()
    expect(ctx.campaign).toBeNull()
    expect(ctx.recent_messages).toEqual([])
  })

  it('throws on non-ok response', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: false,
      status: 500,
    }))

    await expect(fetchContext()).rejects.toThrow('GET /api/context failed: 500')
  })
})

describe('fetchWorldNotes', () => {
  it('returns parsed WorldNote array on success', async () => {
    const notes = [
      { id: 1, campaign_id: 1, title: 'Tavern', content: 'A seedy place.', category: 'location', tags_json: '[]', created_at: '' },
    ]
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(notes),
    }))

    const result = await fetchWorldNotes(1)
    expect(result).toHaveLength(1)
    expect(result[0].title).toBe('Tavern')
  })

  it('appends q param when query is provided', async () => {
    const mockFetch = vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve([]) })
    vi.stubGlobal('fetch', mockFetch)

    await fetchWorldNotes(1, 'dragon')
    expect(mockFetch).toHaveBeenCalledWith('/api/campaigns/1/world-notes?q=dragon')
  })

  it('omits q param when query is empty string', async () => {
    const mockFetch = vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve([]) })
    vi.stubGlobal('fetch', mockFetch)

    await fetchWorldNotes(1, '')
    expect(mockFetch).toHaveBeenCalledWith('/api/campaigns/1/world-notes')
  })

  it('throws on non-ok response', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false, status: 404 }))
    await expect(fetchWorldNotes(1)).rejects.toThrow('failed: 404')
  })
})

describe('fetchDiceRolls', () => {
  it('returns parsed DiceRoll array on success', async () => {
    const rolls = [
      { id: 1, session_id: 1, expression: '1d20+5', result: 18, breakdown_json: '[]', created_at: '' },
    ]
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(rolls),
    }))

    const result = await fetchDiceRolls(1)
    expect(result).toHaveLength(1)
    expect(result[0].expression).toBe('1d20+5')
    expect(result[0].result).toBe(18)
  })

  it('throws on non-ok response', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false, status: 500 }))
    await expect(fetchDiceRolls(1)).rejects.toThrow('failed: 500')
  })
})
```

- [ ] **Step 2: Run tests to verify they fail**

```
cd /home/digitalghost/projects/inkandbone/web
npm test -- --reporter=verbose 2>&1 | grep -E 'FAIL|PASS|fetchWorldNotes|fetchDiceRolls'
```

Expected: tests for `fetchWorldNotes` and `fetchDiceRolls` fail because those functions don't exist yet.

- [ ] **Step 3: Add types**

Replace the entire contents of `web/src/types.ts` with:

```typescript
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

export interface WorldNote {
  id: number
  campaign_id: number
  title: string
  content: string
  category: string
  tags_json: string
  created_at: string
}

export interface DiceRoll {
  id: number
  session_id: number
  expression: string
  result: number
  breakdown_json: string
  created_at: string
}
```

- [ ] **Step 4: Add API functions**

Replace the entire contents of `web/src/api.ts` with:

```typescript
import type { GameContext, WorldNote, DiceRoll } from './types'

export async function fetchContext(): Promise<GameContext> {
  const res = await fetch('/api/context')
  if (!res.ok) throw new Error(`GET /api/context failed: ${res.status}`)
  return res.json()
}

export async function fetchWorldNotes(campaignId: number, q?: string): Promise<WorldNote[]> {
  const url = q
    ? `/api/campaigns/${campaignId}/world-notes?q=${encodeURIComponent(q)}`
    : `/api/campaigns/${campaignId}/world-notes`
  const res = await fetch(url)
  if (!res.ok) throw new Error(`GET ${url} failed: ${res.status}`)
  return res.json()
}

export async function fetchDiceRolls(sessionId: number): Promise<DiceRoll[]> {
  const url = `/api/sessions/${sessionId}/dice-rolls`
  const res = await fetch(url)
  if (!res.ok) throw new Error(`GET ${url} failed: ${res.status}`)
  return res.json()
}
```

- [ ] **Step 5: Run tests to verify they pass**

```
cd /home/digitalghost/projects/inkandbone/web
npm test
```

Expected: all tests pass including the new `fetchWorldNotes` and `fetchDiceRolls` suites.

- [ ] **Step 6: Commit**

```bash
cd /home/digitalghost/projects/inkandbone
git add web/src/types.ts web/src/api.ts web/src/api.test.ts
git commit -m "feat: add WorldNote and DiceRoll types with fetch functions"
```

---

### Task 3: WorldNotesPanel Component

**Files:**
- Create: `web/src/WorldNotesPanel.tsx`
- Create: `web/src/WorldNotesPanel.test.tsx`

- [ ] **Step 1: Write the failing tests**

Create `web/src/WorldNotesPanel.test.tsx`:

```typescript
import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, cleanup, waitFor, fireEvent } from '@testing-library/react'
import { WorldNotesPanel } from './WorldNotesPanel'
import type { WorldNote } from './types'

const notes: WorldNote[] = [
  { id: 1, campaign_id: 1, title: 'Tavern', content: 'A seedy place.', category: 'location', tags_json: '[]', created_at: '' },
  { id: 2, campaign_id: 1, title: 'Dragon', content: 'Ancient red dragon.', category: 'npc', tags_json: '[]', created_at: '' },
]

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

describe('WorldNotesPanel', () => {
  it('renders notes returned by fetch', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve(notes) }))
    render(<WorldNotesPanel campaignId={1} />)
    expect(await screen.findByText('Tavern')).toBeInTheDocument()
    expect(screen.getByText('Dragon')).toBeInTheDocument()
  })

  it('shows empty state when no notes', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve([]) }))
    render(<WorldNotesPanel campaignId={1} />)
    expect(await screen.findByText('No notes found.')).toBeInTheDocument()
  })

  it('calls fetch with q param when search input changes', async () => {
    const mockFetch = vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve([]) })
    vi.stubGlobal('fetch', mockFetch)
    render(<WorldNotesPanel campaignId={1} />)
    await screen.findByText('No notes found.')
    fireEvent.change(screen.getByRole('searchbox'), { target: { value: 'tavern' } })
    await waitFor(() => {
      expect(mockFetch).toHaveBeenLastCalledWith('/api/campaigns/1/world-notes?q=tavern')
    })
  })
})
```

- [ ] **Step 2: Run tests to verify they fail**

```
cd /home/digitalghost/projects/inkandbone/web
npm test -- --reporter=verbose 2>&1 | grep -E 'WorldNotesPanel|FAIL|Cannot find module'
```

Expected: `FAIL` — `WorldNotesPanel` module not found.

- [ ] **Step 3: Implement the component**

Create `web/src/WorldNotesPanel.tsx`:

```typescript
import { useState, useEffect, useCallback } from 'react'
import { fetchWorldNotes } from './api'
import type { WorldNote } from './types'

interface Props {
  campaignId: number
}

export function WorldNotesPanel({ campaignId }: Props) {
  const [notes, setNotes] = useState<WorldNote[]>([])
  const [query, setQuery] = useState('')

  const load = useCallback(() => {
    fetchWorldNotes(campaignId, query || undefined)
      .then(setNotes)
      .catch(() => setNotes([]))
  }, [campaignId, query])

  useEffect(() => {
    load()
  }, [load])

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

- [ ] **Step 4: Run tests to verify they pass**

```
cd /home/digitalghost/projects/inkandbone/web
npm test -- --reporter=verbose 2>&1 | grep -E 'WorldNotesPanel|✓|✗|FAIL|PASS'
```

Expected: all three `WorldNotesPanel` tests pass.

- [ ] **Step 5: Commit**

```bash
cd /home/digitalghost/projects/inkandbone
git add web/src/WorldNotesPanel.tsx web/src/WorldNotesPanel.test.tsx
git commit -m "feat: add WorldNotesPanel component with search"
```

---

### Task 4: DiceHistoryPanel Component

**Files:**
- Create: `web/src/DiceHistoryPanel.tsx`
- Create: `web/src/DiceHistoryPanel.test.tsx`

- [ ] **Step 1: Write the failing tests**

Create `web/src/DiceHistoryPanel.test.tsx`:

```typescript
import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, cleanup } from '@testing-library/react'
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
    render(<DiceHistoryPanel sessionId={1} />)
    expect(await screen.findByText('1d20+5')).toBeInTheDocument()
    expect(screen.getByText('18')).toBeInTheDocument()
    expect(screen.getByText('2d6')).toBeInTheDocument()
    expect(screen.getByText('7')).toBeInTheDocument()
  })

  it('shows empty state when no rolls', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve([]) }))
    render(<DiceHistoryPanel sessionId={1} />)
    expect(await screen.findByText('No rolls yet.')).toBeInTheDocument()
  })
})
```

- [ ] **Step 2: Run tests to verify they fail**

```
cd /home/digitalghost/projects/inkandbone/web
npm test -- --reporter=verbose 2>&1 | grep -E 'DiceHistoryPanel|FAIL|Cannot find module'
```

Expected: `FAIL` — `DiceHistoryPanel` module not found.

- [ ] **Step 3: Implement the component**

Create `web/src/DiceHistoryPanel.tsx`:

```typescript
import { useState, useEffect } from 'react'
import { fetchDiceRolls } from './api'
import type { DiceRoll } from './types'

interface Props {
  sessionId: number
}

export function DiceHistoryPanel({ sessionId }: Props) {
  const [rolls, setRolls] = useState<DiceRoll[]>([])

  useEffect(() => {
    fetchDiceRolls(sessionId)
      .then(setRolls)
      .catch(() => setRolls([]))
  }, [sessionId])

  return (
    <section className="panel dice-history">
      <h2>Dice History</h2>
      {rolls.length === 0 ? (
        <p className="empty">No rolls yet.</p>
      ) : (
        rolls.map((r) => (
          <div key={r.id} className="dice-roll">
            <span className="expression">{r.expression}</span>
            <span className="result">{r.result}</span>
          </div>
        ))
      )}
    </section>
  )
}
```

- [ ] **Step 4: Run tests to verify they pass**

```
cd /home/digitalghost/projects/inkandbone/web
npm test -- --reporter=verbose 2>&1 | grep -E 'DiceHistoryPanel|✓|✗|FAIL|PASS'
```

Expected: both `DiceHistoryPanel` tests pass.

- [ ] **Step 5: Commit**

```bash
cd /home/digitalghost/projects/inkandbone
git add web/src/DiceHistoryPanel.tsx web/src/DiceHistoryPanel.test.tsx
git commit -m "feat: add DiceHistoryPanel component"
```

---

### Task 5: Integrate Panels into App + Styles

**Files:**
- Modify: `web/src/App.tsx` — import and render `WorldNotesPanel` and `DiceHistoryPanel`
- Modify: `web/src/App.css` — add styles for world-notes and dice-history panels
- Modify: `web/src/App.test.tsx` — update fetch mock to handle multiple endpoint URLs

- [ ] **Step 1: Update the App tests to handle multi-URL fetch**

Replace the entire contents of `web/src/App.test.tsx` with:

```typescript
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, cleanup } from '@testing-library/react'
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
    vi.stubGlobal('fetch', vi.fn().mockImplementation((url: string) => {
      if (url === '/api/context') {
        return Promise.resolve({ ok: true, json: () => Promise.resolve(mockCtx) })
      }
      // WorldNotesPanel and DiceHistoryPanel sub-fetches return empty arrays
      return Promise.resolve({ ok: true, json: () => Promise.resolve([]) })
    }))
  })

  afterEach(() => {
    cleanup()
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

  it('renders world notes panel heading', async () => {
    render(<App />)
    // Wait for context to load, then the panel heading should appear
    await screen.findByText('Greyhawk')
    expect(screen.getByText('World Notes')).toBeInTheDocument()
  })

  it('renders dice history panel heading', async () => {
    render(<App />)
    await screen.findByText('Greyhawk')
    expect(screen.getByText('Dice History')).toBeInTheDocument()
  })
})
```

- [ ] **Step 2: Run tests to verify the App tests fail on the new assertions**

```
cd /home/digitalghost/projects/inkandbone/web
npm test -- --reporter=verbose 2>&1 | grep -E 'App|renders world|renders dice|FAIL|PASS'
```

Expected: existing App tests pass (mock is still compatible), new `renders world notes panel heading` and `renders dice history panel heading` tests fail because the panels aren't in App yet.

- [ ] **Step 3: Update App.tsx to include the two panels**

Replace the entire contents of `web/src/App.tsx` with:

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

        {ctx.campaign && <WorldNotesPanel campaignId={ctx.campaign.id} />}

        {ctx.session && <DiceHistoryPanel sessionId={ctx.session.id} />}
      </main>
    </div>
  )
}
```

- [ ] **Step 4: Add styles for the new panels**

Replace the entire contents of `web/src/App.css` with:

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
  --note-category: #a78bfa;
  --dice-result: #fbbf24;
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

/* Session Log */

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

/* Combat Tracker */

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

/* World Notes Panel */

.world-notes {
  max-width: 320px;
  flex: none;
}

.search {
  width: 100%;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 4px;
  color: var(--text);
  font-size: 0.82rem;
  padding: 0.35rem 0.6rem;
  outline: none;
}

.search:focus {
  border-color: var(--accent);
}

.world-note {
  border-top: 1px solid var(--border);
  padding-top: 0.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.note-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.note-title {
  font-size: 0.88rem;
  font-weight: 600;
  color: var(--text);
}

.note-category {
  font-size: 0.7rem;
  color: var(--note-category);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.note-content {
  font-size: 0.82rem;
  color: var(--muted);
  line-height: 1.4;
}

/* Dice History Panel */

.dice-history {
  max-width: 220px;
  flex: none;
}

.dice-roll {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.88rem;
  border-top: 1px solid var(--border);
  padding-top: 0.4rem;
}

.dice-roll .expression {
  color: var(--text);
  font-family: monospace;
}

.dice-roll .result {
  font-weight: 700;
  color: var(--dice-result);
  font-size: 1rem;
}

/* Shared */

.empty {
  color: var(--muted);
  font-size: 0.85rem;
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

- [ ] **Step 5: Run the full frontend test suite**

```
cd /home/digitalghost/projects/inkandbone/web
npm test
```

Expected: all tests pass — App, WorldNotesPanel, DiceHistoryPanel, api, useWebSocket.

- [ ] **Step 6: TypeScript compilation check**

```
cd /home/digitalghost/projects/inkandbone/web
npx tsc -b --noEmit
```

Expected: no errors.

- [ ] **Step 7: Run full Go test suite**

```
cd /home/digitalghost/projects/inkandbone
go test ./...
```

Expected: all tests pass.

- [ ] **Step 8: Commit**

```bash
cd /home/digitalghost/projects/inkandbone
git add web/src/App.tsx web/src/App.css web/src/App.test.tsx
git commit -m "feat: integrate WorldNotesPanel and DiceHistoryPanel into dashboard"
```

---

## Self-Review

### Spec Coverage

All features from this plan's goal are covered:
- `GET /api/campaigns/{id}/world-notes` endpoint with `?q=` search → Task 1
- `WorldNote` and `DiceRoll` TypeScript types → Task 2
- `fetchWorldNotes` and `fetchDiceRolls` API functions → Task 2
- `WorldNotesPanel` with search input → Task 3
- `DiceHistoryPanel` → Task 4
- Both panels integrated into `App.tsx` with conditional rendering → Task 5

### Placeholder Scan

No TBDs, no "implement later", no "similar to Task N" — every step has full code.

### Type Consistency

- `WorldNote` defined in Task 2 `types.ts` — used in Task 3 `WorldNotesPanel.tsx` and `WorldNotesPanel.test.tsx` ✓
- `DiceRoll` defined in Task 2 `types.ts` — used in Task 4 `DiceHistoryPanel.tsx` and `DiceHistoryPanel.test.tsx` ✓
- `fetchWorldNotes(campaignId: number, q?: string)` defined in Task 2 `api.ts` — used in Task 3 `WorldNotesPanel.tsx` as `fetchWorldNotes(campaignId, query || undefined)` ✓
- `fetchDiceRolls(sessionId: number)` defined in Task 2 `api.ts` — used in Task 4 `DiceHistoryPanel.tsx` ✓
- `WorldNotesPanel` exported as named export in Task 3, imported in Task 5 as `{ WorldNotesPanel }` ✓
- `DiceHistoryPanel` exported as named export in Task 4, imported in Task 5 as `{ DiceHistoryPanel }` ✓
- `ctx.campaign.id` (type `number`) → `campaignId: number` in `WorldNotesPanel` Props ✓
- `ctx.session.id` (type `number`) → `sessionId: number` in `DiceHistoryPanel` Props ✓
