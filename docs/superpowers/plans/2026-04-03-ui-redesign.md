# UI Redesign — Worn Grimoire Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the flat dark dashboard with a three-column "Worn Grimoire" layout using a parchment palette, Cormorant Garamond headings, pip-dot attributes, segmented track bars, and Prose Journal message rendering.

**Architecture:** Presentation-only rewrite. All component logic (WS reactivity, debounce PATCH, portrait upload, AI draft) is preserved unchanged. The three-column flex layout is driven entirely by App.tsx/App.css restructuring. CharacterSheetPanel gains new visual renderers for Ironsworn attribute pips and track bars; all other components are restyle-only.

**Tech Stack:** React 19 + TypeScript, Vite, CSS custom properties, Google Fonts (Cormorant Garamond)

---

## File Map

| File | Action | Responsibility |
|------|--------|---------------|
| `web/index.html` | Modify | Add Google Fonts link for Cormorant Garamond |
| `web/src/App.css` | Full rewrite | CSS variables (grimoire palette), grimoire layout classes, header, all panel/component styles |
| `web/src/App.tsx` | Modify | Three-column layout, Prose Journal messages, map drawer state, right sidebar tab state, header breadcrumb |
| `web/src/CharacterSheetPanel.tsx` | Modify | Circular portrait + hover link; pip dots for attributes; segmented bars for tracks |
| `web/src/DiceHistoryPanel.tsx` | Modify | Cap at 5 entries (latest first), compact grimoire style |
| `web/src/WorldNotesPanel.tsx` | Modify | Remove outer `<section className="panel">` wrapper; grimoire card styles |
| `web/src/JournalPanel.tsx` | Modify | Remove outer `<section className="panel">` wrapper; grimoire textarea style |
| `web/src/CombatPanel.tsx` | Modify | Accent left-border treatment, dark-alarming background |

---

## Task 1: Add Cormorant Garamond Font

**Files:**
- Modify: `web/index.html`

- [ ] **Step 1: Add Google Fonts preconnect and stylesheet links**

Replace the contents of `web/index.html` with:

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <link rel="icon" type="image/svg+xml" href="/favicon.svg" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Ink & Bone</title>
    <link rel="preconnect" href="https://fonts.googleapis.com" />
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
    <link href="https://fonts.googleapis.com/css2?family=Cormorant+Garamond:ital,wght@0,400;0,600;1,400&display=swap" rel="stylesheet" />
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

- [ ] **Step 2: Verify TypeScript build still passes**

Run: `cd /home/digitalghost/projects/inkandbone/web && npx tsc --noEmit`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
cd /home/digitalghost/projects/inkandbone
git add web/index.html
git commit -m "feat: add Cormorant Garamond Google Font to index.html"
```

---

## Task 2: Rewrite App.css — Grimoire Palette & Layout

**Files:**
- Full rewrite: `web/src/App.css`

- [ ] **Step 1: Replace App.css with grimoire design tokens and all component styles**

Write the following to `web/src/App.css`:

```css
/* ── Design Tokens ─────────────────────────────────────── */

:root {
  --bg:       #0f0e0a;
  --surface:  #1a1710;
  --surface2: #141208;
  --gold:     #c9a84c;
  --gold-dim: #5a4a2a;
  --text:     #d4c5a0;
  --text-dim: #8b7355;
  --border:   #3a3020;
  --health:   #8fbc5a;
  --accent:   #e94560;
  --serif:    'Cormorant Garamond', Georgia, serif;
}

/* ── Reset ──────────────────────────────────────────────── */

* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

body {
  background: var(--bg);
  color: var(--text);
  font-family: system-ui, sans-serif;
  font-size: 13px;
  min-height: 100vh;
}

/* ── Root Layout ────────────────────────────────────────── */

.grimoire {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}

/* ── Header ─────────────────────────────────────────────── */

.grimoire-header {
  flex-shrink: 0;
  background: var(--surface2);
  border-bottom: 1px solid var(--border);
  padding: 0.5rem 1.25rem;
  font-family: var(--serif);
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.grimoire-header .h-campaign { color: var(--gold); }
.grimoire-header .h-sep     { color: var(--gold-dim); }
.grimoire-header .h-char,
.grimoire-header .h-session { color: var(--text-dim); }

/* ── Body ───────────────────────────────────────────────── */

.grimoire-body {
  display: flex;
  flex: 1;
  overflow: hidden;
}

/* ── Left Sidebar ───────────────────────────────────────── */

.sidebar-left {
  width: 200px;
  flex-shrink: 0;
  background: var(--surface2);
  border-right: 1px solid var(--border);
  overflow-y: auto;
  padding: 1rem 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

/* Portrait */

.portrait-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.35rem;
  position: relative;
}

.portrait-circle {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid var(--gold);
  display: block;
}

.portrait-placeholder-circle {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: var(--surface);
  border: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-dim);
  font-size: 11px;
}

.portrait-change {
  font-family: var(--serif);
  font-size: 10px;
  color: var(--gold-dim);
  text-decoration: underline;
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.2s;
}

.portrait-wrap:hover .portrait-change {
  opacity: 1;
}

/* Attributes (pip dots) */

.attr-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.attr-label {
  font-family: var(--serif);
  font-size: 9px;
  color: var(--gold-dim);
  letter-spacing: 2px;
  text-transform: uppercase;
}

.attr-pips {
  display: flex;
  gap: 3px;
}

.pip {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.pip.filled  { background: var(--gold); }
.pip.empty   { background: var(--surface2); border: 1px solid var(--border); }

/* Tracks (segmented bars) */

.track-row {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.track-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.track-label {
  font-family: var(--serif);
  font-size: 9px;
  color: var(--gold-dim);
  letter-spacing: 2px;
  text-transform: uppercase;
}

.track-value {
  font-size: 10px;
  color: var(--gold);
}

.track-segments {
  display: flex;
  gap: 2px;
}

.track-seg {
  flex: 1;
  height: 5px;
  border-radius: 1px;
}

.track-seg.filled-health   { background: var(--health); }
.track-seg.filled-momentum { background: var(--gold); }
.track-seg.empty-seg       { background: var(--surface2); border: 1px solid var(--border); }

/* Sidebar rule */

.sidebar-rule {
  border: none;
  border-top: 1px solid var(--border);
  margin: 0;
}

/* Dice History (left sidebar) */

.dice-compact {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.dice-compact-label {
  font-family: var(--serif);
  font-size: 9px;
  color: var(--gold-dim);
  letter-spacing: 2px;
  text-transform: uppercase;
  margin-bottom: 0.15rem;
}

.dice-compact-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
}

.dice-compact-expr { color: var(--text-dim); font-family: monospace; }
.dice-compact-result { color: var(--gold); font-weight: 600; }

/* ── Center Column ──────────────────────────────────────── */

.story-center {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--bg);
}

.story-scroll {
  flex: 1;
  overflow-y: auto;
  padding: 1.5rem 2rem;
}

/* Session title */

.session-title {
  font-family: var(--serif);
  font-size: 18px;
  color: var(--gold);
  text-align: center;
  letter-spacing: 4px;
  text-transform: uppercase;
  margin-bottom: 0.25rem;
}

.session-date {
  text-align: center;
  font-size: 11px;
  color: var(--text-dim);
  margin-bottom: 1.5rem;
}

/* Prose Journal messages */

.prose-gm {
  color: var(--text);
  font-family: system-ui, sans-serif;
  font-size: 13px;
  line-height: 1.8;
  margin-bottom: 1rem;
}

.prose-player {
  margin-bottom: 1rem;
}

.prose-player-label {
  font-family: var(--serif);
  font-size: 9px;
  color: var(--gold-dim);
  letter-spacing: 2px;
  text-transform: uppercase;
  margin-bottom: 0.25rem;
}

.prose-player-text {
  font-style: italic;
  color: var(--gold);
  font-family: system-ui, sans-serif;
  font-size: 13px;
  line-height: 1.8;
}

.prose-divider {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 1rem 0;
  color: var(--gold-dim);
  font-size: 12px;
}

.prose-divider::before,
.prose-divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background: var(--border);
}

/* Combat Panel (center column, top) */

.combat-grimoire {
  margin: 0 2rem 0.75rem;
  border-left: 3px solid var(--accent);
  background: #120a0c;
  border: 1px solid #3a1020;
  border-left: 3px solid var(--accent);
  padding: 0.75rem 1rem;
  flex-shrink: 0;
}

.combat-grimoire h2 {
  font-family: var(--serif);
  font-size: 12px;
  color: var(--accent);
  letter-spacing: 2px;
  text-transform: uppercase;
  margin-bottom: 0.5rem;
}

.combatant-card {
  border-top: 1px solid #3a1020;
  padding: 0.4rem 0;
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.combatant-card.active-turn {
  border-left: 2px solid var(--accent);
  padding-left: 0.5rem;
}

.combatant-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.combatant-name { font-size: 13px; font-weight: 600; color: var(--text); }
.combatant-card.player .combatant-name { color: #8fbc5a; }
.combatant-card.enemy  .combatant-name { color: var(--accent); }
.combatant-init { font-size: 11px; color: var(--text-dim); }

.hp-bar-track {
  height: 4px;
  background: var(--border);
  border-radius: 2px;
  overflow: hidden;
}

.hp-bar-fill { height: 100%; border-radius: 2px; transition: width 0.3s ease; }
.hp-bar-green  { background: #4ade80; }
.hp-bar-yellow { background: #facc15; }
.hp-bar-red    { background: var(--accent); }

.hp-label  { font-size: 11px; color: var(--text-dim); }
.conditions { display: flex; flex-wrap: wrap; gap: 0.2rem; }
.condition-badge {
  font-size: 10px;
  background: rgba(233,69,96,0.15);
  color: var(--accent);
  border-radius: 2px;
  padding: 0.1rem 0.3rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

/* Map Drawer */

.map-drawer {
  flex-shrink: 0;
  border-top: 1px solid var(--border);
}

.map-drawer-handle {
  background: var(--surface2);
  padding: 0.4rem 1rem;
  text-align: center;
  font-family: var(--serif);
  font-size: 12px;
  color: var(--gold-dim);
  letter-spacing: 2px;
  cursor: pointer;
  user-select: none;
}

.map-drawer-handle:hover { color: var(--gold); }

.map-drawer-content {
  overflow: hidden;
  transition: max-height 0.3s ease;
  max-height: 0;
}

.map-drawer-content.open {
  max-height: 60vh;
}

.map-drawer-inner {
  height: 60vh;
  position: relative;
  overflow: hidden;
}

/* ── Right Sidebar ──────────────────────────────────────── */

.sidebar-right {
  width: 260px;
  flex-shrink: 0;
  background: var(--surface2);
  border-left: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.tab-bar {
  display: flex;
  gap: 1.25rem;
  padding: 0.6rem 1rem 0;
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}

.tab-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-family: var(--serif);
  font-size: 11px;
  letter-spacing: 2px;
  text-transform: uppercase;
  padding-bottom: 0.4rem;
  color: var(--text-dim);
  border-bottom: 2px solid transparent;
}

.tab-btn.active {
  color: var(--gold);
  border-bottom-color: var(--gold);
}

.tab-content {
  flex: 1;
  overflow-y: auto;
  padding: 0.75rem 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

/* Notes tab */

.notes-search {
  width: 100%;
  background: transparent;
  border: none;
  border-bottom: 1px solid var(--border);
  color: var(--text);
  font-size: 13px;
  padding: 0.25rem 0;
  outline: none;
  font-family: system-ui, sans-serif;
}

.notes-search::placeholder { color: var(--text-dim); }
.notes-search:focus { border-bottom-color: var(--gold); }

.note-card {
  border-top: 1px solid var(--border);
  padding-top: 0.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.note-title {
  font-family: var(--serif);
  font-size: 14px;
  color: var(--gold);
}

.note-content {
  font-size: 12px;
  color: var(--text);
  line-height: 1.5;
}

.note-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
  margin-top: 0.15rem;
}

.note-tag {
  font-size: 10px;
  border: 1px solid var(--gold-dim);
  border-radius: 2px;
  color: var(--gold-dim);
  padding: 0.1rem 0.35rem;
  cursor: pointer;
  background: transparent;
}

.note-tag:hover  { border-color: var(--gold); color: var(--gold); }
.note-tag.active { background: var(--gold-dim); color: var(--text); }

.ai-text-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-family: var(--serif);
  font-size: 11px;
  color: var(--gold-dim);
  padding: 0;
  text-align: left;
}

.ai-text-btn:hover  { color: var(--gold); }
.ai-text-btn:disabled { opacity: 0.4; cursor: not-allowed; }

/* Journal tab */

.journal-textarea {
  flex: 1;
  width: 100%;
  background: transparent;
  border: none;
  color: var(--text);
  font-size: 13px;
  font-family: system-ui, sans-serif;
  line-height: 1.8;
  resize: none;
  outline: none;
  min-height: 200px;
}

/* ── Shared utilities ───────────────────────────────────── */

.empty { color: var(--text-dim); font-size: 12px; }
.error { color: var(--accent); padding: 2rem; text-align: center; }
.loading { color: var(--text-dim); padding: 2rem; text-align: center; }
```

- [ ] **Step 2: Verify TypeScript build still passes**

Run: `cd /home/digitalghost/projects/inkandbone/web && npx tsc --noEmit`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
cd /home/digitalghost/projects/inkandbone
git add web/src/App.css
git commit -m "feat: rewrite App.css with Worn Grimoire design tokens and layout"
```

---

## Task 3: Restructure App.tsx — Three-Column Layout

**Files:**
- Modify: `web/src/App.tsx`

This task restructures the layout, adds map drawer state, right sidebar tab state, the Prose Journal message renderer, and the grimoire header breadcrumb. SessionTimeline is removed from render (component is kept).

- [ ] **Step 1: Replace App.tsx with the grimoire layout**

```tsx
import { useState, useEffect, useCallback } from 'react'
import { useWebSocket } from './useWebSocket'
import { fetchContext } from './api'
import type { GameContext, Message } from './types'
import { CombatPanel } from './CombatPanel'
import { WorldNotesPanel } from './WorldNotesPanel'
import { DiceHistoryPanel } from './DiceHistoryPanel'
import { MapPanel } from './MapPanel'
import { JournalPanel } from './JournalPanel'
import { CharacterSheetPanel } from './CharacterSheetPanel'
import './App.css'

const WS_URL = `ws://${window.location.host}/ws`

function ProseJournal({ messages, characterName }: { messages: Message[]; characterName: string }) {
  if (messages.length === 0) {
    return <p className="empty">The story has not yet begun.</p>
  }

  const nodes: React.ReactNode[] = []
  messages.forEach((m, i) => {
    if (m.role === 'assistant') {
      nodes.push(
        <p key={m.id} className="prose-gm">{m.content}</p>
      )
    } else {
      nodes.push(
        <div key={m.id} className="prose-player">
          <div className="prose-player-label">{characterName} speaks</div>
          <p className="prose-player-text">{m.content}</p>
        </div>
      )
      // Decorative divider after each player turn (except the last message)
      if (i < messages.length - 1) {
        nodes.push(
          <div key={`div-${m.id}`} className="prose-divider">◆</div>
        )
      }
    }
  })
  return <>{nodes}</>
}

export default function App() {
  const [ctx, setCtx] = useState<GameContext | null>(null)
  const [messages, setMessages] = useState<Message[]>([])
  const [error, setError] = useState<string | null>(null)
  const [aiEnabled, setAiEnabled] = useState(false)
  const [mapOpen, setMapOpen] = useState(false)
  const [rightTab, setRightTab] = useState<'notes' | 'journal'>('notes')

  useEffect(() => {
    fetch('/api/health')
      .then((r) => r.json())
      .then((data: { ai_enabled: boolean }) => setAiEnabled(data.ai_enabled))
      .catch(() => setAiEnabled(false))
  }, [])

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

  const handleEvent = useCallback((_data: unknown) => { loadContext() }, [loadContext])
  const { lastEvent } = useWebSocket(WS_URL, handleEvent)

  if (error) return <div className="error">{error}</div>
  if (!ctx) return <div className="loading">Loading…</div>

  const sessionTitle = ctx.session?.title?.toUpperCase() ?? ''
  const sessionDate = ctx.session?.date
    ? new Date(ctx.session.date).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })
    : ''

  return (
    <div className="grimoire">
      <header className="grimoire-header">
        <span className="h-campaign">{ctx.campaign?.name ?? 'No campaign'}</span>
        <span className="h-sep">›</span>
        <span className="h-char">{ctx.character?.name ?? 'No character'}</span>
        <span className="h-sep">›</span>
        <span className="h-session">{ctx.session?.title ?? 'No session'}</span>
      </header>

      <div className="grimoire-body">

        {/* Left Sidebar */}
        <aside className="sidebar-left">
          <CharacterSheetPanel
            character={ctx?.character ?? null}
            rulesetId={ctx?.campaign?.ruleset_id ?? null}
            lastEvent={lastEvent}
          />
          <hr className="sidebar-rule" />
          {ctx.session && (
            <DiceHistoryPanel sessionId={ctx.session.id} lastEvent={lastEvent} />
          )}
        </aside>

        {/* Center Column */}
        <main className="story-center">
          <div className="story-scroll">
            {sessionTitle && (
              <>
                <div className="session-title">✦ {sessionTitle} ✦</div>
                {sessionDate && <div className="session-date">{sessionDate}</div>}
              </>
            )}
            {ctx.active_combat && <CombatPanel combat={ctx.active_combat} />}
            <ProseJournal messages={messages} characterName={ctx.character?.name ?? 'Player'} />
          </div>

          <div className="map-drawer">
            <div
              className="map-drawer-handle"
              onClick={() => setMapOpen((o) => !o)}
            >
              {mapOpen
                ? '[ ▴ COLLAPSE ]'
                : `[ ${ctx.campaign?.name?.toUpperCase() ?? 'THE IRONLANDS'} ▾ ]`}
            </div>
            <div className={`map-drawer-content${mapOpen ? ' open' : ''}`}>
              <div className="map-drawer-inner">
                <MapPanel campaignId={ctx?.campaign?.id ?? null} lastEvent={lastEvent} />
              </div>
            </div>
          </div>
        </main>

        {/* Right Sidebar */}
        <aside className="sidebar-right">
          <div className="tab-bar">
            <button
              className={`tab-btn${rightTab === 'notes' ? ' active' : ''}`}
              onClick={() => setRightTab('notes')}
            >
              Notes
            </button>
            <button
              className={`tab-btn${rightTab === 'journal' ? ' active' : ''}`}
              onClick={() => setRightTab('journal')}
            >
              Journal
            </button>
          </div>
          <div className="tab-content">
            {rightTab === 'notes' && ctx.campaign && (
              <WorldNotesPanel
                campaignId={ctx.campaign.id}
                lastEvent={lastEvent}
                aiEnabled={aiEnabled}
              />
            )}
            {rightTab === 'journal' && (
              <JournalPanel
                session={ctx?.session ?? null}
                lastEvent={lastEvent}
                aiEnabled={aiEnabled}
              />
            )}
          </div>
        </aside>

      </div>
    </div>
  )
}
```

- [ ] **Step 2: Verify TypeScript build passes**

Run: `cd /home/digitalghost/projects/inkandbone/web && npx tsc --noEmit`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
cd /home/digitalghost/projects/inkandbone
git add web/src/App.tsx
git commit -m "feat: restructure App.tsx to three-column Worn Grimoire layout"
```

---

## Task 4: Restyle CharacterSheetPanel — Pips & Segmented Bars

**Files:**
- Modify: `web/src/CharacterSheetPanel.tsx`

Ironsworn attributes (edge, heart, iron, shadow, wits) render as 3 pip dots. Ironsworn tracks (health, spirit, supply, momentum) render as segmented bars. The portrait becomes a circular 80px crop with a gold border; "Change portrait" appears only on hover. All schema-driven patch logic is preserved unchanged.

- [ ] **Step 1: Replace CharacterSheetPanel.tsx**

```tsx
import { useState, useEffect, useRef } from 'react'
import { fetchRuleset, patchCharacter, uploadPortrait } from './api'
import type { Ruleset } from './api'
import type { Character } from './types'

interface SchemaField {
  key: string
  label: string
  type: 'text' | 'number' | 'textarea'
}

interface CharacterSheetPanelProps {
  character: Character | null
  rulesetId: number | null
  lastEvent: unknown
}

interface CharacterUpdatedPayload {
  id: number
  data_json?: string
  portrait_path?: string
}

interface CharacterUpdatedEvent {
  type: 'character_updated'
  payload: CharacterUpdatedPayload
}

function isCharacterUpdatedEvent(ev: unknown): ev is CharacterUpdatedEvent {
  if (typeof ev !== 'object' || ev === null) return false
  const e = ev as Record<string, unknown>
  if (e['type'] !== 'character_updated') return false
  const p = e['payload']
  if (typeof p !== 'object' || p === null) return false
  return typeof (p as Record<string, unknown>)['id'] === 'number'
}

const ATTRIBUTE_KEYS = new Set(['edge', 'heart', 'iron', 'shadow', 'wits'])
const TRACK_KEYS     = new Set(['health', 'spirit', 'supply', 'momentum'])

function AttributePips({ value }: { value: number }) {
  return (
    <div className="attr-pips">
      {[1, 2, 3].map((i) => (
        <div key={i} className={`pip ${i <= value ? 'filled' : 'empty'}`} />
      ))}
    </div>
  )
}

function TrackBar({ fieldKey, value }: { fieldKey: string; value: number }) {
  const isMomentum = fieldKey === 'momentum'
  const max = isMomentum ? 10 : 5
  const filled = Math.max(0, Math.min(max, value))
  const colorClass = isMomentum ? 'filled-momentum' : 'filled-health'
  const displayValue = isMomentum ? value : `${value}/${max}`

  return (
    <div className="track-row">
      <div className="track-header">
        <span className="track-label">{fieldKey}</span>
        <span className="track-value">{displayValue}</span>
      </div>
      <div className="track-segments">
        {Array.from({ length: max }, (_, i) => (
          <div
            key={i}
            className={`track-seg ${i < filled ? colorClass : 'empty-seg'}`}
          />
        ))}
      </div>
    </div>
  )
}

export function CharacterSheetPanel({ character, rulesetId, lastEvent }: CharacterSheetPanelProps) {
  const [ruleset, setRuleset] = useState<Ruleset | null>(null)
  const [fields, setFields] = useState<Record<string, string>>({})
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (rulesetId === null) return
    fetchRuleset(rulesetId)
      .then(setRuleset)
      .catch(console.error)
  }, [rulesetId])

  useEffect(() => {
    if (!character) return
    try {
      const data = JSON.parse(character.data_json || '{}') as Record<string, unknown>
      setFields(Object.fromEntries(Object.entries(data).map(([k, v]) => [k, String(v ?? '')])))
    } catch {
      setFields({})
    }
  }, [character?.id])

  useEffect(() => {
    if (!isCharacterUpdatedEvent(lastEvent)) return
    if (lastEvent.payload.id !== character?.id) return
    if (lastEvent.payload.data_json) {
      try {
        const data = JSON.parse(lastEvent.payload.data_json) as Record<string, unknown>
        setFields(Object.fromEntries(Object.entries(data).map(([k, v]) => [k, String(v ?? '')])))
      } catch { /* ignore */ }
    }
  }, [lastEvent, character?.id])

  if (!character) return null

  const schema: SchemaField[] = (() => {
    try {
      const parsed = JSON.parse(ruleset?.schema_json ?? '[]') as unknown
      if (!Array.isArray(parsed) && typeof parsed === 'object' && parsed !== null) {
        const legacy = parsed as Record<string, unknown>
        if (Array.isArray(legacy['fields'])) {
          return (legacy['fields'] as string[]).map((key) => ({
            key,
            label: key.charAt(0).toUpperCase() + key.slice(1).replace(/_/g, ' '),
            type: 'text' as const,
          }))
        }
      }
      return parsed as SchemaField[]
    } catch {
      return []
    }
  })()

  function handleChange(key: string, value: string) {
    const next = { ...fields, [key]: value }
    setFields(next)
    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => {
      const updates: Record<string, unknown> = {}
      schema.forEach((f) => {
        const v = next[f.key]
        updates[f.key] = f.type === 'number' ? (v === '' ? null : Number(v)) : v
      })
      patchCharacter(character!.id, updates).catch(console.error)
    }, 500)
  }

  function handlePortraitChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    uploadPortrait(character!.id, file).catch(console.error)
  }

  const attributeFields = schema.filter((f) => ATTRIBUTE_KEYS.has(f.key))
  const trackFields     = schema.filter((f) => TRACK_KEYS.has(f.key))
  const otherFields     = schema.filter((f) => !ATTRIBUTE_KEYS.has(f.key) && !TRACK_KEYS.has(f.key))

  return (
    <>
      {/* Portrait */}
      <div className="portrait-wrap">
        {character.portrait_path ? (
          <img
            className="portrait-circle"
            src={`/api/files/${character.portrait_path}`}
            alt={character.name}
          />
        ) : (
          <div className="portrait-placeholder-circle">{character.name[0]}</div>
        )}
        <label className="portrait-change">
          <input
            ref={fileInputRef}
            type="file"
            accept="image/*"
            onChange={handlePortraitChange}
            style={{ display: 'none' }}
          />
          Change portrait
        </label>
      </div>

      {/* Attributes — pip dots */}
      {attributeFields.length > 0 && (
        <div>
          {attributeFields.map((f) => (
            <div key={f.key} className="attr-row">
              <span className="attr-label">{f.key}</span>
              <AttributePips value={Number(fields[f.key] ?? 0)} />
            </div>
          ))}
        </div>
      )}

      {/* Tracks — segmented bars */}
      {trackFields.length > 0 && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
          {trackFields.map((f) => (
            <TrackBar
              key={f.key}
              fieldKey={f.key}
              value={Number(fields[f.key] ?? 0)}
            />
          ))}
        </div>
      )}

      {/* Other fields — plain inputs (non-Ironsworn rulesets) */}
      {otherFields.map((field) => (
        <div key={field.key} style={{ display: 'flex', flexDirection: 'column', gap: '0.2rem' }}>
          <label style={{ fontSize: '9px', textTransform: 'uppercase', letterSpacing: '2px', color: 'var(--gold-dim)', fontFamily: 'var(--serif)' }}>
            {field.label}
          </label>
          {field.type === 'textarea' ? (
            <textarea
              value={fields[field.key] ?? ''}
              onChange={(e) => handleChange(field.key, e.target.value)}
              style={{ background: 'var(--surface)', border: '1px solid var(--border)', color: 'var(--text)', fontSize: '12px', padding: '0.25rem', fontFamily: 'inherit', resize: 'vertical', minHeight: '3rem' }}
            />
          ) : (
            <input
              type={field.type}
              value={fields[field.key] ?? ''}
              onChange={(e) => handleChange(field.key, e.target.value)}
              style={{ background: 'var(--surface)', border: '1px solid var(--border)', color: 'var(--text)', fontSize: '12px', padding: '0.25rem 0.4rem' }}
            />
          )}
        </div>
      ))}
    </>
  )
}
```

- [ ] **Step 2: Verify TypeScript build passes**

Run: `cd /home/digitalghost/projects/inkandbone/web && npx tsc --noEmit`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
cd /home/digitalghost/projects/inkandbone
git add web/src/CharacterSheetPanel.tsx
git commit -m "feat: restyle CharacterSheetPanel with pip dots and segmented track bars"
```

---

## Task 5: Restyle DiceHistoryPanel — Compact Grimoire List

**Files:**
- Modify: `web/src/DiceHistoryPanel.tsx`

Cap display at 5 entries (latest first — API already returns descending order). Remove the `.panel` wrapper since DiceHistoryPanel now lives inside `.sidebar-left`.

- [ ] **Step 1: Replace DiceHistoryPanel.tsx**

```tsx
import { useState, useEffect } from 'react'
import { fetchDiceRolls } from './api'
import type { DiceRoll } from './types'

interface Props {
  sessionId: number
  lastEvent: unknown
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
      fetchDiceRolls(sessionId).then(setRolls).catch(() => {})
    }
  }, [lastEvent, sessionId])

  const recent = rolls.slice(0, 5)

  if (recent.length === 0) return null

  return (
    <div className="dice-compact">
      <div className="dice-compact-label">Dice</div>
      {recent.map((r) => (
        <div key={r.id} className="dice-compact-row">
          <span className="dice-compact-expr">{r.expression}</span>
          <span className="dice-compact-result">{r.result}</span>
        </div>
      ))}
    </div>
  )
}
```

- [ ] **Step 2: Verify TypeScript build passes**

Run: `cd /home/digitalghost/projects/inkandbone/web && npx tsc --noEmit`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
cd /home/digitalghost/projects/inkandbone
git add web/src/DiceHistoryPanel.tsx
git commit -m "feat: restyle DiceHistoryPanel to compact 5-entry grimoire list"
```

---

## Task 6: Restyle WorldNotesPanel — Remove Wrapper, Grimoire Cards

**Files:**
- Modify: `web/src/WorldNotesPanel.tsx`

Remove the outer `<section className="panel world-notes">` so the panel renders directly inside the right sidebar's `.tab-content`. Apply grimoire note card styles.

- [ ] **Step 1: Replace WorldNotesPanel.tsx**

```tsx
import { useState, useEffect, useCallback } from 'react'
import { fetchWorldNotes, draftWorldNote } from './api'
import type { WorldNote } from './types'

interface Props {
  campaignId: number
  lastEvent: unknown
  aiEnabled: boolean
}

function parseTags(json: string): string[] {
  try { return JSON.parse(json) as string[] }
  catch { return [] }
}

export function WorldNotesPanel({ campaignId, lastEvent, aiEnabled }: Props) {
  const [notes, setNotes] = useState<WorldNote[]>([])
  const [query, setQuery] = useState('')
  const [activeTag, setActiveTag] = useState<string | null>(null)
  const [drafting, setDrafting] = useState(false)

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

  async function handleDraftWithAI() {
    const hint = window.prompt('Describe the note:')
    if (!hint) return
    setDrafting(true)
    try {
      await draftWorldNote(campaignId, hint)
    } catch (err) {
      console.error(err)
    } finally {
      setDrafting(false)
    }
  }

  return (
    <>
      <input
        className="notes-search"
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
            <div key={n.id} className="note-card">
              <div className="note-title">{n.title}</div>
              {tags.length > 0 && (
                <div className="note-tags">
                  {tags.map((tag) => (
                    <button
                      key={tag}
                      className={`note-tag${activeTag === tag ? ' active' : ''}`}
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
      {aiEnabled && (
        <button className="ai-text-btn" disabled={drafting} onClick={handleDraftWithAI}>
          {drafting ? 'Drafting…' : 'Draft with AI'}
        </button>
      )}
    </>
  )
}
```

- [ ] **Step 2: Verify TypeScript build passes**

Run: `cd /home/digitalghost/projects/inkandbone/web && npx tsc --noEmit`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
cd /home/digitalghost/projects/inkandbone
git add web/src/WorldNotesPanel.tsx
git commit -m "feat: restyle WorldNotesPanel with grimoire card styles, remove panel wrapper"
```

---

## Task 7: Restyle JournalPanel — Remove Wrapper, Grimoire Textarea

**Files:**
- Modify: `web/src/JournalPanel.tsx`

Remove the outer `<section className="panel journal-panel">` wrapper. Apply grimoire textarea and button styles.

- [ ] **Step 1: Replace JournalPanel.tsx**

```tsx
import { useState, useEffect } from 'react'
import { patchSessionSummary, generateRecap } from './api'

interface JournalPanelProps {
  session: { id: number; summary: string } | null
  lastEvent: unknown
  aiEnabled: boolean
}

interface SessionUpdatedPayload {
  session_id: number
  summary: string
}

interface SessionUpdatedEvent {
  type: 'session_updated'
  payload: SessionUpdatedPayload
}

function isSessionUpdatedEvent(ev: unknown): ev is SessionUpdatedEvent {
  if (typeof ev !== 'object' || ev === null) return false
  const e = ev as Record<string, unknown>
  if (e['type'] !== 'session_updated') return false
  const payload = e['payload']
  if (typeof payload !== 'object' || payload === null) return false
  const p = payload as Record<string, unknown>
  return typeof p['session_id'] === 'number' && typeof p['summary'] === 'string'
}

export function JournalPanel({ session, lastEvent, aiEnabled }: JournalPanelProps) {
  const [draft, setDraft] = useState(session?.summary ?? '')

  useEffect(() => {
    if (session) {
      setDraft(session.summary)
    }
  // Intentionally omit session.summary: reset draft only when session changes
  }, [session?.id]) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (!isSessionUpdatedEvent(lastEvent)) return
    if (lastEvent.payload.session_id !== session?.id) return
    setDraft(lastEvent.payload.summary)
  }, [lastEvent, session?.id])

  if (session === null) return null

  function handleBlur() {
    patchSessionSummary(session!.id, draft).catch(console.error)
  }

  async function handleGenerateRecap() {
    const result = await generateRecap(session!.id)
    setDraft(result.summary)
  }

  return (
    <>
      <textarea
        className="journal-textarea"
        value={draft}
        onChange={(e) => setDraft(e.target.value)}
        onBlur={handleBlur}
        placeholder="Your session journal…"
      />
      {aiEnabled && (
        <button className="ai-text-btn" onClick={handleGenerateRecap}>
          Generate recap
        </button>
      )}
    </>
  )
}
```

- [ ] **Step 2: Verify TypeScript build passes**

Run: `cd /home/digitalghost/projects/inkandbone/web && npx tsc --noEmit`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
cd /home/digitalghost/projects/inkandbone
git add web/src/JournalPanel.tsx
git commit -m "feat: restyle JournalPanel with grimoire textarea, remove panel wrapper"
```

---

## Task 8: Restyle CombatPanel — Accent Left-Border Treatment

**Files:**
- Modify: `web/src/CombatPanel.tsx`

Change the wrapper class from `panel combat-panel` to `combat-grimoire` so it picks up the alarming dark-red accent styling defined in App.css. All HP bar and combatant logic is unchanged.

- [ ] **Step 1: Replace CombatPanel.tsx**

```tsx
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
    <div className="combat-grimoire">
      <h2>⚔ {encounter.name}</h2>
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
    </div>
  )
}
```

- [ ] **Step 2: Verify TypeScript build passes**

Run: `cd /home/digitalghost/projects/inkandbone/web && npx tsc --noEmit`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
cd /home/digitalghost/projects/inkandbone
git add web/src/CombatPanel.tsx
git commit -m "feat: restyle CombatPanel with accent left-border grimoire treatment"
```

---

## Task 9: Full Build Verification

- [ ] **Step 1: Run the full production build**

Run: `cd /home/digitalghost/projects/inkandbone && make build`
Expected: `web/dist/` produced, `ttrpg` binary built with no errors

- [ ] **Step 2: Run Go tests**

Run: `cd /home/digitalghost/projects/inkandbone && go test ./... -v`
Expected: all tests pass (no backend changes were made)

- [ ] **Step 3: Start the server and visually verify**

Run: `ttrpg`
Open browser at `http://localhost:8080`

Verify:
- Three-column layout renders (left sidebar 200px, center flex, right sidebar 260px)
- Header shows "The Ironlands › Xavier › Whispers in the Mist"
- Left sidebar: portrait circle, pip dots for edge/heart/iron/shadow/wits, segmented bars for health/spirit/supply/momentum
- Center: "✦ WHISPERS IN THE MIST ✦" title, GM messages as plain prose, player messages as italic gold with label above
- Center bottom: map drawer handle visible, expands/collapses on click
- Right sidebar: NOTES / JOURNAL tabs toggle correctly
- Google Fonts Cormorant Garamond loads for headings/labels

---

## Spec Self-Review

All requirements from the spec are covered. `ProseJournal` uses `characterName` prop (dynamic, reads from `ctx.character?.name`). No placeholders. Type signatures are consistent across all tasks.
