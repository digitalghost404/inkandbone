# Plan 7: Live Combat Tracker + Session Timeline

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the inline combat table with a rich CombatPanel (HP bars, condition badges, active-turn highlight) and add a SessionTimeline feed that merges messages and dice rolls from the DB and grows in real time via WebSocket events.

**Architecture:** A new `GET /api/sessions/{id}/timeline` endpoint merges messages + dice rolls from the DB sorted by created_at. CombatPanel is a pure presentational component — it receives `ctx.active_combat` as a prop and re-renders automatically whenever App.tsx refreshes context on WS events (no extra fetch or WS logic inside the component). SessionTimeline fetches the initial feed then appends WS events (dice_rolled, world_note_created, combat_started, combat_ended) client-side via the `lastEvent` prop.

**Prerequisite:** Plan 6 must be complete. App.tsx already has `const { lastEvent } = useWebSocket(...)` and passes `lastEvent` to panels.

**Tech Stack:** Go 1.22 (net/http, encoding/json, sort), SQLite, React 19 + TypeScript, Vitest + Testing Library

---

## File Map

| File | Change |
|------|--------|
| `internal/db/queries_timeline.go` | **New**: `TimelineEntry` type + `GetSessionTimeline` |
| `internal/db/queries_timeline_test.go` | **New**: merge + sort + empty tests |
| `internal/api/routes.go` | Add `handleGetTimeline` handler |
| `internal/api/server.go` | Register `GET /api/sessions/{id}/timeline` |
| `internal/api/routes_test.go` | Add timeline endpoint tests |
| `web/src/types.ts` | Add `TimelineEntry` interface |
| `web/src/api.ts` | Add `fetchTimeline` function |
| `web/src/api.test.ts` | Test `fetchTimeline` |
| `web/src/CombatPanel.tsx` | **New**: HP bars, conditions, active-turn highlight |
| `web/src/CombatPanel.test.tsx` | **New**: render, HP color, condition, active-turn tests |
| `web/src/SessionTimeline.tsx` | **New**: fetch on mount + WS append + slide-in animation |
| `web/src/SessionTimeline.test.tsx` | **New**: render, empty state, WS append tests |
| `web/src/App.tsx` | Replace inline combat `<section>`; import + render CombatPanel + SessionTimeline |
| `web/src/App.test.tsx` | Add combat panel + timeline heading tests |
| `web/src/App.css` | Add combat panel, HP bar, conditions, timeline CSS |

---

## Task 1: DB — GetSessionTimeline

**Files:**
- Create: `internal/db/queries_timeline.go`
- Create: `internal/db/queries_timeline_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/db/queries_timeline_test.go`:

```go
package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSessionTimeline_emptySession(t *testing.T) {
	d := newTestDB(t)
	rsID, err := d.CreateRuleset(t.Name(), "{}", "test")
	require.NoError(t, err)
	campID, err := d.CreateCampaign(rsID, "Camp", "")
	require.NoError(t, err)
	sessID, err := d.CreateSession(campID, "S1", "2026-04-03")
	require.NoError(t, err)

	entries, err := d.GetSessionTimeline(sessID)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestGetSessionTimeline_mergesAndSorts(t *testing.T) {
	d := newTestDB(t)
	rsID, err := d.CreateRuleset(t.Name(), "{}", "test")
	require.NoError(t, err)
	campID, err := d.CreateCampaign(rsID, "Camp", "")
	require.NoError(t, err)
	sessID, err := d.CreateSession(campID, "S1", "2026-04-03")
	require.NoError(t, err)

	_, err = d.CreateMessage(sessID, "user", "Hello")
	require.NoError(t, err)
	_, err = d.LogDiceRoll(sessID, "1d20+5", 18, "[13]")
	require.NoError(t, err)

	entries, err := d.GetSessionTimeline(sessID)
	require.NoError(t, err)
	require.Len(t, entries, 2)

	types := map[string]bool{}
	for _, e := range entries {
		types[e.Type] = true
		assert.NotEmpty(t, e.Timestamp)
		assert.NotEmpty(t, e.Data)
	}
	assert.True(t, types["message"], "expected a message entry")
	assert.True(t, types["dice_roll"], "expected a dice_roll entry")
}

func TestGetSessionTimeline_sortedByTimestamp(t *testing.T) {
	d := newTestDB(t)
	rsID, err := d.CreateRuleset(t.Name(), "{}", "test")
	require.NoError(t, err)
	campID, err := d.CreateCampaign(rsID, "Camp", "")
	require.NoError(t, err)
	sessID, err := d.CreateSession(campID, "S1", "2026-04-03")
	require.NoError(t, err)

	_, err = d.LogDiceRoll(sessID, "1d6", 4, "[4]")
	require.NoError(t, err)
	_, err = d.CreateMessage(sessID, "assistant", "The die lands on 4.")
	require.NoError(t, err)

	entries, err := d.GetSessionTimeline(sessID)
	require.NoError(t, err)
	require.Len(t, entries, 2)

	// Timestamps must be non-decreasing.
	assert.LessOrEqual(t, entries[0].Timestamp, entries[1].Timestamp)
}
```

- [ ] **Step 2: Run tests to verify they fail**

```
cd /path/to/inkandbone
go test ./internal/db/ -run TestGetSessionTimeline -v
```

Expected: FAIL — `d.GetSessionTimeline undefined`

- [ ] **Step 3: Implement GetSessionTimeline**

Create `internal/db/queries_timeline.go`:

```go
package db

import (
	"encoding/json"
	"sort"
)

// TimelineEntry is a single chronological item in the session feed.
// Type is one of "message" or "dice_roll".
// Data is the raw JSON of the underlying record.
type TimelineEntry struct {
	Type      string          `json:"type"`
	Timestamp string          `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

// GetSessionTimeline returns all messages and dice rolls for the given session,
// merged and sorted by created_at ascending.
func (d *DB) GetSessionTimeline(sessionID int64) ([]TimelineEntry, error) {
	msgs, err := d.ListMessages(sessionID)
	if err != nil {
		return nil, err
	}
	rolls, err := d.ListDiceRolls(sessionID)
	if err != nil {
		return nil, err
	}

	entries := make([]TimelineEntry, 0, len(msgs)+len(rolls))
	for _, m := range msgs {
		b, _ := json.Marshal(m)
		entries = append(entries, TimelineEntry{Type: "message", Timestamp: m.CreatedAt, Data: b})
	}
	for _, r := range rolls {
		b, _ := json.Marshal(r)
		entries = append(entries, TimelineEntry{Type: "dice_roll", Timestamp: r.CreatedAt, Data: b})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp < entries[j].Timestamp
	})

	return entries, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

```
go test ./internal/db/ -run TestGetSessionTimeline -v
```

Expected: PASS (3 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/db/queries_timeline.go internal/db/queries_timeline_test.go
git commit -m "feat: DB GetSessionTimeline merges messages + dice rolls"
```

---

## Task 2: HTTP Timeline Endpoint

**Files:**
- Modify: `internal/api/routes.go`
- Modify: `internal/api/server.go`
- Modify: `internal/api/routes_test.go`

- [ ] **Step 1: Write failing tests**

Add to `internal/api/routes_test.go`:

```go
func TestGetTimeline_empty(t *testing.T) {
	s := newTestServer(t)
	_, sessID := seedCampaign(t, s.db)
	req := httptest.NewRequest(http.MethodGet, "/api/sessions/"+strconv.FormatInt(sessID, 10)+"/timeline", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var entries []db.TimelineEntry
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &entries))
	assert.Empty(t, entries)
}

func TestGetTimeline_withData(t *testing.T) {
	s := newTestServer(t)
	_, sessID := seedCampaign(t, s.db)
	_, err := s.db.CreateMessage(sessID, "user", "A brave move.")
	require.NoError(t, err)
	_, err = s.db.LogDiceRoll(sessID, "2d6", 9, "[4,5]")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/sessions/"+strconv.FormatInt(sessID, 10)+"/timeline", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var entries []db.TimelineEntry
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &entries))
	assert.Len(t, entries, 2)
}

func TestGetTimeline_invalidID(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/sessions/abc/timeline", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
```

- [ ] **Step 2: Run tests to verify they fail**

```
go test ./internal/api/ -run TestGetTimeline -v
```

Expected: FAIL — route not registered

- [ ] **Step 3: Add handler to routes.go**

Add to `internal/api/routes.go` (after the existing handlers):

```go
func (s *Server) handleGetTimeline(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	entries, err := s.db.GetSessionTimeline(id)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if entries == nil {
		entries = []db.TimelineEntry{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}
```

- [ ] **Step 4: Register route in server.go**

In `internal/api/server.go`, inside `registerRoutes()`, add after the existing session routes:

```go
s.mux.HandleFunc("GET /api/sessions/{id}/timeline", s.handleGetTimeline)
```

- [ ] **Step 5: Run tests to verify they pass**

```
go test ./internal/api/ -run TestGetTimeline -v
```

Expected: PASS (3 tests)

- [ ] **Step 6: Run the full Go test suite**

```
go test ./...
```

Expected: all tests pass

- [ ] **Step 7: Commit**

```bash
git add internal/api/routes.go internal/api/server.go internal/api/routes_test.go
git commit -m "feat: GET /api/sessions/{id}/timeline endpoint"
```

---

## Task 3: TypeScript Types + fetchTimeline

**Files:**
- Modify: `web/src/types.ts`
- Modify: `web/src/api.ts`
- Modify: `web/src/api.test.ts`

- [ ] **Step 1: Write failing test**

Add to `web/src/api.test.ts`:

```typescript
describe('fetchTimeline', () => {
  it('returns parsed TimelineEntry array on success', async () => {
    const entries = [
      { type: 'message', timestamp: '2026-04-03T10:00:00Z', data: { id: 1, role: 'user', content: 'Hi' } },
      { type: 'dice_roll', timestamp: '2026-04-03T10:01:00Z', data: { id: 1, expression: '1d20', result: 15 } },
    ]
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve(entries) }))
    const { fetchTimeline } = await import('./api')
    const result = await fetchTimeline(1)
    expect(result).toHaveLength(2)
    expect(result[0].type).toBe('message')
    expect(result[1].type).toBe('dice_roll')
  })

  it('throws on non-ok response', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false, status: 404 }))
    const { fetchTimeline } = await import('./api')
    await expect(fetchTimeline(1)).rejects.toThrow('failed: 404')
  })
})
```

- [ ] **Step 2: Run tests to verify they fail**

```
cd web && npx vitest run src/api.test.ts
```

Expected: FAIL — `fetchTimeline` not exported

- [ ] **Step 3: Add TimelineEntry to types.ts**

Add to the end of `web/src/types.ts`:

```typescript
export interface TimelineEntry {
  type: 'message' | 'dice_roll' | 'world_note_event' | 'combat_event'
  timestamp: string
  data: Record<string, unknown>
}
```

- [ ] **Step 4: Add fetchTimeline to api.ts**

Add to `web/src/api.ts`:

First, update the import line at the top:

```typescript
import type { GameContext, WorldNote, DiceRoll, TimelineEntry } from './types'
```

Then add the function:

```typescript
export async function fetchTimeline(sessionId: number): Promise<TimelineEntry[]> {
  const url = `/api/sessions/${sessionId}/timeline`
  const res = await fetch(url)
  if (!res.ok) throw new Error(`GET ${url} failed: ${res.status}`)
  return res.json()
}
```

- [ ] **Step 5: Run tests to verify they pass**

```
npx vitest run src/api.test.ts
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add web/src/types.ts web/src/api.ts web/src/api.test.ts
git commit -m "feat: TimelineEntry type and fetchTimeline API function"
```

---

## Task 4: CombatPanel Component

**Files:**
- Create: `web/src/CombatPanel.tsx`
- Create: `web/src/CombatPanel.test.tsx`

- [ ] **Step 1: Write failing tests**

Create `web/src/CombatPanel.test.tsx`:

```typescript
import { describe, it, expect, afterEach } from 'vitest'
import { render, screen, cleanup } from '@testing-library/react'
import { CombatPanel } from './CombatPanel'
import type { CombatSnapshot } from './types'

const combat: CombatSnapshot = {
  encounter: { id: 1, session_id: 1, name: 'Bandit Ambush', active: true, created_at: '' },
  combatants: [
    {
      id: 1, encounter_id: 1, character_id: null,
      name: 'Kael', initiative: 18, hp_current: 30, hp_max: 40,
      conditions_json: '[]', is_player: true,
    },
    {
      id: 2, encounter_id: 1, character_id: null,
      name: 'Bandit', initiative: 12, hp_current: 5, hp_max: 20,
      conditions_json: '["frightened"]', is_player: false,
    },
  ],
}

afterEach(cleanup)

describe('CombatPanel', () => {
  it('renders encounter name and all combatants', () => {
    render(<CombatPanel combat={combat} />)
    expect(screen.getByText('Combat: Bandit Ambush')).toBeInTheDocument()
    expect(screen.getByText('Kael')).toBeInTheDocument()
    expect(screen.getByText('Bandit')).toBeInTheDocument()
  })

  it('marks first combatant (highest initiative) as active turn', () => {
    render(<CombatPanel combat={combat} />)
    const cards = document.querySelectorAll('.combatant-card')
    expect(cards[0]).toHaveClass('active-turn')
    expect(cards[1]).not.toHaveClass('active-turn')
  })

  it('applies player class to player combatant', () => {
    render(<CombatPanel combat={combat} />)
    const cards = document.querySelectorAll('.combatant-card')
    expect(cards[0]).toHaveClass('player')
    expect(cards[1]).toHaveClass('enemy')
  })

  it('renders condition badges', () => {
    render(<CombatPanel combat={combat} />)
    expect(screen.getByText('frightened')).toBeInTheDocument()
  })

  it('applies red hp bar class when HP is at or below 25%', () => {
    // Bandit: 5/20 = 25% → red
    render(<CombatPanel combat={combat} />)
    const fills = document.querySelectorAll('.hp-bar-fill')
    expect(fills[1]).toHaveClass('hp-bar-red')
  })

  it('applies green hp bar class when HP is above 50%', () => {
    // Kael: 30/40 = 75% → green
    render(<CombatPanel combat={combat} />)
    const fills = document.querySelectorAll('.hp-bar-fill')
    expect(fills[0]).toHaveClass('hp-bar-green')
  })

  it('renders hp text label', () => {
    render(<CombatPanel combat={combat} />)
    expect(screen.getByText('30 / 40 HP')).toBeInTheDocument()
  })
})
```

- [ ] **Step 2: Run tests to verify they fail**

```
npx vitest run src/CombatPanel.test.tsx
```

Expected: FAIL — `CombatPanel` not found

- [ ] **Step 3: Implement CombatPanel**

Create `web/src/CombatPanel.tsx`:

```typescript
import type { CombatSnapshot } from './types'

interface Props {
  combat: CombatSnapshot
}

function hpBarClass(current: number, max: number): string {
  if (max === 0) return 'hp-bar-green'
  const ratio = current / max
  if (ratio > 0.5) return 'hp-bar-green'
  if (ratio > 0.25) return 'hp-bar-yellow'
  return 'hp-bar-red'
}

function parseConditions(json: string): string[] {
  try {
    return JSON.parse(json) as string[]
  } catch {
    return []
  }
}

export function CombatPanel({ combat }: Props) {
  const { encounter, combatants } = combat
  return (
    <section className="panel combat-panel">
      <h2>Combat: {encounter.name}</h2>
      {combatants.map((c, idx) => {
        const pct = c.hp_max > 0 ? Math.max(0, Math.round((c.hp_current / c.hp_max) * 100)) : 0
        const colorClass = hpBarClass(c.hp_current, c.hp_max)
        const conditions = parseConditions(c.conditions_json)
        const isActive = idx === 0
        return (
          <div
            key={c.id}
            className={`combatant-card ${c.is_player ? 'player' : 'enemy'} ${isActive ? 'active-turn' : ''}`}
          >
            <div className="combatant-header">
              <span className="combatant-name">{c.name}</span>
              <span className="combatant-init">Init {c.initiative}</span>
            </div>
            <div className="hp-bar-track">
              <div className={`hp-bar-fill ${colorClass}`} style={{ width: `${pct}%` }} />
            </div>
            <div className="hp-label">
              {c.hp_current} / {c.hp_max} HP
            </div>
            {conditions.length > 0 && (
              <div className="conditions">
                {conditions.map((cond) => (
                  <span key={cond} className="condition-badge">
                    {cond}
                  </span>
                ))}
              </div>
            )}
          </div>
        )
      })}
    </section>
  )
}
```

- [ ] **Step 4: Run tests to verify they pass**

```
npx vitest run src/CombatPanel.test.tsx
```

Expected: PASS (7 tests)

- [ ] **Step 5: Commit**

```bash
git add web/src/CombatPanel.tsx web/src/CombatPanel.test.tsx
git commit -m "feat: CombatPanel with HP bars, conditions, and active-turn highlight"
```

---

## Task 5: SessionTimeline Component

**Files:**
- Create: `web/src/SessionTimeline.tsx`
- Create: `web/src/SessionTimeline.test.tsx`

- [ ] **Step 1: Write failing tests**

Create `web/src/SessionTimeline.test.tsx`:

```typescript
import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, cleanup, waitFor } from '@testing-library/react'
import { SessionTimeline } from './SessionTimeline'
import type { TimelineEntry } from './types'

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

const msgEntry: TimelineEntry = {
  type: 'message',
  timestamp: '2026-04-03T10:00:00Z',
  data: { id: 1, session_id: 1, role: 'user', content: 'We enter the dungeon.', created_at: '' },
}

const diceEntry: TimelineEntry = {
  type: 'dice_roll',
  timestamp: '2026-04-03T10:01:00Z',
  data: { id: 2, session_id: 1, expression: '1d20+3', result: 17, breakdown_json: '[14]', created_at: '' },
}

describe('SessionTimeline', () => {
  it('renders timeline entries fetched from API', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve([msgEntry, diceEntry]),
    }))
    render(<SessionTimeline sessionId={1} lastEvent={null} />)
    expect(await screen.findByText('We enter the dungeon.')).toBeInTheDocument()
    expect(screen.getByText('1d20+3')).toBeInTheDocument()
  })

  it('shows empty state when timeline is empty', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve([]),
    }))
    render(<SessionTimeline sessionId={1} lastEvent={null} />)
    expect(await screen.findByText('No events yet.')).toBeInTheDocument()
  })

  it('calls fetch with correct session URL', async () => {
    const mockFetch = vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve([]) })
    vi.stubGlobal('fetch', mockFetch)
    render(<SessionTimeline sessionId={42} lastEvent={null} />)
    await screen.findByText('No events yet.')
    expect(mockFetch).toHaveBeenCalledWith('/api/sessions/42/timeline')
  })

  it('appends a dice_rolled WS event as a new timeline entry', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve([]) }))
    const { rerender } = render(<SessionTimeline sessionId={1} lastEvent={null} />)
    await screen.findByText('No events yet.')

    rerender(
      <SessionTimeline
        sessionId={1}
        lastEvent={{ type: 'dice_rolled', payload: { expression: '2d6', total: 8, breakdown: [3, 5] } }}
      />,
    )

    await waitFor(() => {
      expect(screen.getByText('2d6')).toBeInTheDocument()
    })
  })

  it('appends a combat_started WS event as a new timeline entry', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve([]) }))
    const { rerender } = render(<SessionTimeline sessionId={1} lastEvent={null} />)
    await screen.findByText('No events yet.')

    rerender(
      <SessionTimeline
        sessionId={1}
        lastEvent={{ type: 'combat_started', payload: { encounter_id: 1, name: 'Goblin Raid' } }}
      />,
    )

    await waitFor(() => {
      expect(screen.getByText('combat started')).toBeInTheDocument()
      expect(screen.getByText('Goblin Raid')).toBeInTheDocument()
    })
  })
})
```

- [ ] **Step 2: Run tests to verify they fail**

```
npx vitest run src/SessionTimeline.test.tsx
```

Expected: FAIL — `SessionTimeline` not found

- [ ] **Step 3: Implement SessionTimeline**

Create `web/src/SessionTimeline.tsx`:

```typescript
import { useState, useEffect, useCallback, useRef } from 'react'
import { fetchTimeline } from './api'
import type { TimelineEntry } from './types'

interface Props {
  sessionId: number
  lastEvent: unknown
}

type WsPayload = Record<string, unknown>
type WsEvent = { type?: string; payload?: WsPayload }

// Build a TimelineEntry from a WS event payload. Returns null if the event
// type is not one the timeline cares about.
function wsToEntry(ev: WsEvent): TimelineEntry | null {
  const now = new Date().toISOString()
  const p = ev.payload ?? {}

  switch (ev.type) {
    case 'dice_rolled':
      return {
        type: 'dice_roll',
        timestamp: now,
        data: {
          expression: p.expression as string,
          result: p.total as number,
          breakdown_json: JSON.stringify(p.breakdown ?? []),
        },
      }
    case 'world_note_created':
      return {
        type: 'world_note_event',
        timestamp: now,
        data: { note_id: p.note_id as number, title: p.title as string, action: 'created' },
      }
    case 'combat_started':
      return {
        type: 'combat_event',
        timestamp: now,
        data: { encounter_id: p.encounter_id as number, name: p.name as string, ended: false },
      }
    case 'combat_ended':
      return {
        type: 'combat_event',
        timestamp: now,
        data: { encounter_id: p.encounter_id as number, ended: true },
      }
    default:
      return null
  }
}

export function SessionTimeline({ sessionId, lastEvent }: Props) {
  const [entries, setEntries] = useState<TimelineEntry[]>([])
  const [newCount, setNewCount] = useState(0)
  const [error, setError] = useState<string | null>(null)
  const newCountTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const loadTimeline = useCallback(() => {
    let ignored = false
    fetchTimeline(sessionId)
      .then((data) => { if (!ignored) setEntries(data) })
      .catch(() => { if (!ignored) setError('Could not load timeline') })
    return () => { ignored = true }
  }, [sessionId])

  useEffect(() => loadTimeline(), [loadTimeline])

  useEffect(() => {
    const ev = lastEvent as WsEvent | null
    if (!ev?.type) return
    const entry = wsToEntry(ev)
    if (!entry) return

    setEntries((prev) => [...prev, entry])
    setNewCount((c) => c + 1)

    if (newCountTimerRef.current) clearTimeout(newCountTimerRef.current)
    newCountTimerRef.current = setTimeout(() => {
      setNewCount((c) => Math.max(0, c - 1))
    }, 600)
  }, [lastEvent])

  if (error) return <p className="error">{error}</p>

  return (
    <section className="panel timeline-panel">
      <h2>Session Timeline</h2>
      {entries.length === 0 ? (
        <p className="empty">No events yet.</p>
      ) : (
        <div className="timeline-feed">
          {entries.map((e, idx) => {
            const isNew = idx >= entries.length - newCount
            return (
              <div
                key={`${e.type}-${idx}`}
                className={`timeline-entry timeline-${e.type}${isNew ? ' entry-new' : ''}`}
              >
                {renderEntry(e)}
              </div>
            )
          })}
        </div>
      )}
    </section>
  )
}

function renderEntry(e: TimelineEntry) {
  const d = e.data

  if (e.type === 'message') {
    return (
      <>
        <span className="tl-role">{String(d.role ?? '')}</span>
        <span className="tl-content">{String(d.content ?? '')}</span>
      </>
    )
  }

  if (e.type === 'dice_roll') {
    const breakdown = (() => {
      try {
        return JSON.parse(String(d.breakdown_json ?? '[]')) as number[]
      } catch {
        return []
      }
    })()
    return (
      <>
        <span className="tl-expr">{String(d.expression ?? '')}</span>
        <span className="tl-result">{String(d.result ?? '')}</span>
        {breakdown.length > 0 && (
          <span className="tl-breakdown">
            {breakdown.map((v, i) => (
              <span key={i} className="die-badge">
                [{v}]
              </span>
            ))}
          </span>
        )}
      </>
    )
  }

  if (e.type === 'world_note_event') {
    return (
      <>
        <span className="tl-badge note">note</span>
        <span className="tl-content">{String(d.title ?? '')}</span>
      </>
    )
  }

  if (e.type === 'combat_event') {
    const ended = Boolean(d.ended)
    return (
      <>
        <span className="tl-badge combat">{ended ? 'combat ended' : 'combat started'}</span>
        {!ended && <span className="tl-content">{String(d.name ?? '')}</span>}
      </>
    )
  }

  return null
}
```

- [ ] **Step 4: Run tests to verify they pass**

```
npx vitest run src/SessionTimeline.test.tsx
```

Expected: PASS (5 tests)

- [ ] **Step 5: Commit**

```bash
git add web/src/SessionTimeline.tsx web/src/SessionTimeline.test.tsx
git commit -m "feat: SessionTimeline with DB feed + WS append + slide-in animation"
```

---

## Task 6: App Wiring + CSS

**Files:**
- Modify: `web/src/App.tsx`
- Modify: `web/src/App.test.tsx`
- Modify: `web/src/App.css`

**Note:** At this point Plan 6 is already merged. App.tsx has `const { lastEvent } = useWebSocket(...)` and passes `lastEvent` to WorldNotesPanel and DiceHistoryPanel. The inline `{ctx.active_combat && <section className="panel combat">...</section>}` block still exists and needs replacing.

- [ ] **Step 1: Write failing tests**

Add to `web/src/App.test.tsx`:

In the `beforeEach`, update the fetch mock to handle the timeline route:

```typescript
// Replace the existing beforeEach with this:
beforeEach(() => {
  vi.stubGlobal('WebSocket', MockWebSocket)
  vi.stubGlobal('fetch', vi.fn().mockImplementation((url: string) => {
    if (url === '/api/context') {
      return Promise.resolve({ ok: true, json: () => Promise.resolve(mockCtx) })
    }
    // Sub-panel fetches return empty arrays
    return Promise.resolve({ ok: true, json: () => Promise.resolve([]) })
  }))
})
```

Add these two tests:

```typescript
it('renders combat panel when active_combat is set', async () => {
  const ctxWithCombat: GameContext = {
    ...mockCtx,
    active_combat: {
      encounter: { id: 1, session_id: 1, name: 'Dragon Fight', active: true, created_at: '' },
      combatants: [
        { id: 1, encounter_id: 1, character_id: null, name: 'Zara', initiative: 20, hp_current: 40, hp_max: 40, conditions_json: '[]', is_player: true },
      ],
    },
  }
  vi.stubGlobal('fetch', vi.fn().mockImplementation((url: string) => {
    if (url === '/api/context') {
      return Promise.resolve({ ok: true, json: () => Promise.resolve(ctxWithCombat) })
    }
    return Promise.resolve({ ok: true, json: () => Promise.resolve([]) })
  }))
  render(<App />)
  expect(await screen.findByText('Combat: Dragon Fight')).toBeInTheDocument()
  expect(screen.getByText('Zara')).toBeInTheDocument()
})

it('renders session timeline heading', async () => {
  render(<App />)
  await screen.findByText('Greyhawk')
  expect(screen.getByText('Session Timeline')).toBeInTheDocument()
})
```

- [ ] **Step 2: Run tests to verify they fail**

```
npx vitest run src/App.test.tsx
```

Expected: FAIL — `Combat: Dragon Fight` and `Session Timeline` not in document

- [ ] **Step 3: Update App.tsx**

Replace the content of `web/src/App.tsx`:

```typescript
import { useState, useEffect, useCallback } from 'react'
import { useWebSocket } from './useWebSocket'
import { fetchContext } from './api'
import type { GameContext, Message } from './types'
import { CombatPanel } from './CombatPanel'
import { SessionTimeline } from './SessionTimeline'
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
        {ctx.character?.portrait_path && (
          <img
            className="portrait"
            src={`/api/files/${ctx.character.portrait_path}`}
            alt={ctx.character.name}
          />
        )}
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

        {ctx.active_combat && <CombatPanel combat={ctx.active_combat} />}

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

- [ ] **Step 4: Add CSS**

Append to `web/src/App.css`:

```css
/* Combat Panel */

.combat-panel {
  max-width: 300px;
  flex: none;
}

.combatant-card {
  border-top: 1px solid var(--border);
  padding: 0.5rem 0;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.combatant-card.active-turn {
  border-left: 3px solid var(--accent);
  padding-left: 0.5rem;
}

.combatant-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.combatant-name {
  font-size: 0.88rem;
  font-weight: 600;
}

.combatant-card.player .combatant-name {
  color: var(--player);
}

.combatant-card.enemy .combatant-name {
  color: var(--enemy);
}

.combatant-init {
  font-size: 0.75rem;
  color: var(--muted);
}

.hp-bar-track {
  height: 6px;
  background: var(--border);
  border-radius: 3px;
  overflow: hidden;
}

.hp-bar-fill {
  height: 100%;
  border-radius: 3px;
  transition: width 0.3s ease;
}

.hp-bar-green { background: #4ade80; }
.hp-bar-yellow { background: #facc15; }
.hp-bar-red { background: var(--enemy); }

.hp-label {
  font-size: 0.75rem;
  color: var(--muted);
}

.conditions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
}

.condition-badge {
  font-size: 0.65rem;
  background: var(--border);
  color: var(--note-category);
  border-radius: 3px;
  padding: 0.1rem 0.35rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

/* Session Timeline */

.timeline-panel {
  min-width: 260px;
  flex: 1.5;
}

.timeline-feed {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.timeline-entry {
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
  font-size: 0.82rem;
  border-top: 1px solid var(--border);
  padding-top: 0.35rem;
}

.timeline-entry.entry-new {
  animation: slide-in 0.25s ease-out;
}

@keyframes slide-in {
  from { opacity: 0; transform: translateY(-6px); }
  to   { opacity: 1; transform: translateY(0); }
}

.tl-role {
  font-size: 0.7rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--accent);
  flex-shrink: 0;
}

.timeline-message .tl-role {
  color: #4a9eff;
}

.tl-content {
  color: var(--muted);
  line-height: 1.4;
}

.tl-expr {
  font-family: monospace;
  color: var(--text);
}

.tl-result {
  font-weight: 700;
  color: var(--dice-result);
}

.tl-breakdown {
  display: flex;
  gap: 0.15rem;
}

.tl-badge {
  font-size: 0.65rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-radius: 3px;
  padding: 0.1rem 0.35rem;
  flex-shrink: 0;
}

.tl-badge.note {
  background: rgba(167, 139, 250, 0.2);
  color: var(--note-category);
}

.tl-badge.combat {
  background: rgba(233, 69, 96, 0.2);
  color: var(--accent);
}
```

- [ ] **Step 5: Run App tests**

```
npx vitest run src/App.test.tsx
```

Expected: PASS

- [ ] **Step 6: Run the full frontend test suite**

```
npx vitest run
```

Expected: all tests pass

- [ ] **Step 7: Build and smoke-test**

```
cd .. && make build
```

Open the browser. Verify: Session Timeline heading appears, combat panel shows HP bars when combat is active, dice rolls appear in the timeline feed.

- [ ] **Step 8: Commit**

```bash
git add web/src/App.tsx web/src/App.test.tsx web/src/App.css
git commit -m "feat: wire CombatPanel and SessionTimeline into App"
```
