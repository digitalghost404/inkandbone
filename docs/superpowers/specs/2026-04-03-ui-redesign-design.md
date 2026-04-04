# UI Redesign: Worn Grimoire

**Date:** 2026-04-03  
**Status:** Implemented  
**Scope:** Frontend presentation layer only — no API, DB, or WebSocket logic changes

**Implementation Notes (2026-04-03):**
The Worn Grimoire aesthetic has been fully implemented and is now live in production. The three-column layout, parchment color palette, Cormorant Garamond serif headings, pip-dot attributes (for Ironsworn), segmented track bars, and Prose Journal message rendering are all in place. The "Parchment" light theme is also available via the theme toggle (☀/🌙) in the header. All existing component logic (WebSocket reactivity, debounce PATCH, portrait upload, AI features) has been preserved. Component-level CSS rewrites are complete (App.css, CharacterSheetPanel.css, DiceHistoryPanel.css, CombatPanel.css, WorldNotesPanel.css, JournalPanel.css).

---

## Summary

Replace the current flat, generic dark dashboard with a Worn Grimoire aesthetic using a Story Scroll layout. All existing panel component logic is preserved; only presentation and layout structure change.

---

## Design Decisions

| Dimension | Choice |
|-----------|--------|
| Aesthetic | Worn Grimoire — parchment darks, candlelight gold, serif accents |
| Layout | Story Scroll — wide center column, slim left sidebar, slim right sidebar |
| Stats display | Candlelit Pips — dot pips for attributes, segmented bars for tracks |
| Message style | Prose Journal — flowing narration, player actions italicized in gold |
| Typography | Cormorant Garamond for headings/labels; system-ui for body text |
| Map | Collapsible drawer at bottom of center column |
| Right sidebar | NOTES / JOURNAL tabs (one visible at a time) |

---

## Color Palette

```css
--bg:       #0f0e0a   /* near-black parchment background */
--surface:  #1a1710   /* panel fill                      */
--surface2: #141208   /* recessed / sidebar fill         */
--gold:     #c9a84c   /* primary accent                  */
--gold-dim: #5a4a2a   /* muted labels                    */
--text:     #d4c5a0   /* body text                       */
--text-dim: #8b7355   /* secondary / timestamps          */
--border:   #3a3020   /* dividers and borders            */
--health:   #8fbc5a   /* health / spirit track fill      */
--accent:   #e94560   /* reserved for combat / danger    */
```

---

## Typography

- **Headings, labels, decorative elements:** `'Cormorant Garamond', Georgia, serif` — imported from Google Fonts
- **Body text (prose log, notes, journal):** `system-ui, sans-serif`
- **Minimum font size:** 13px everywhere
- **Body line-height:** 1.8
- **Label style:** uppercase, letter-spacing: 2–3px, `--gold-dim` color

---

## Layout Structure

```
┌──────────────────────────────────────────────────────────────┐
│  HEADER                                                       │
│  The Ironlands › Xavier › Whispers in the Mist               │
├──────────────┬───────────────────────────────┬───────────────┤
│  LEFT        │  CENTER                       │  RIGHT        │
│  200px fixed │  flex: 1                      │  260px fixed  │
│              │                               │               │
│  Portrait    │  Session title (ornamental)   │  [NOTES|JOUR] │
│  Attributes  │                               │               │
│  (pip dots)  │  Session Log                  │  World Notes  │
│              │  (Prose Journal)              │  — or —       │
│  ── rule ──  │                               │  Journal      │
│              │  CombatPanel (when active)    │               │
│  Tracks      │                               │               │
│  (seg bars)  │  ── [ THE IRONLANDS ▾ ] ──   │               │
│              │  Map drawer (expandable)      │               │
│  ── rule ──  │                               │               │
│  Dice History│                               │               │
└──────────────┴───────────────────────────────┴───────────────┘
```

### CSS layout

```css
.grimoire { display: flex; flex-direction: column; height: 100vh; }
.grimoire-header { flex-shrink: 0; }
.grimoire-body { display: flex; flex: 1; overflow: hidden; }
.sidebar-left { width: 200px; flex-shrink: 0; overflow-y: auto; }
.story-center { flex: 1; display: flex; flex-direction: column; overflow: hidden; }
.story-scroll { flex: 1; overflow-y: auto; }
.map-drawer { flex-shrink: 0; }
.sidebar-right { width: 260px; flex-shrink: 0; display: flex; flex-direction: column; }
```

---

## Header

- Single line: `The Ironlands › Xavier › Whispers in the Mist`
- Campaign name in `--gold`, separators `›` in `--gold-dim`, character and session in `--text-dim`
- Font: Cormorant Garamond, ~14px
- No buttons or chrome — purely informational

---

## Left Sidebar — CharacterSheetPanel

### Portrait
- Circular crop, 80px diameter
- 1px gold border
- "Change portrait" link below, visible on hover only

### Attributes (Edge, Heart, Iron, Shadow, Wits)
- Label: Cormorant, 9px, `--gold-dim`, letter-spacing 2px, uppercase
- 3 pip dots per attribute: 8px circles, filled `--gold` / empty `--surface2` with `--border` border
- Arranged as label left, dots right

### Separator
- `1px solid --border` horizontal rule

### Tracks (Health, Spirit, Supply, Momentum)
- Label: same as attributes
- Current/max value in `--gold`, 10px
- Segmented bar: divided into segments matching max value, filled segments use `--health` (green) for Health/Spirit/Supply, `--gold` for Momentum
- Track updates saved via existing 500ms debounce PATCH

### Separator

### Dice History
- Compact list: expression in `--text-dim` + result in `--gold`
- Max 5 entries, no scrolling — latest at top

---

## Center Column — Session Log

### Session Title
```
✦ WHISPERS IN THE MIST ✦
```
- Cormorant Garamond, ~18px, `--gold`, centered, letter-spacing 4px
- Below: date in `--text-dim`, 11px

### Message Rendering

**GM/assistant messages:**
- Plain prose, `--text`, system-ui, 13px, line-height 1.8
- No label or badge

**Player/user messages:**
- Small label above: `"Xavier speaks"` in Cormorant, 9px, `--gold-dim`, letter-spacing 2px
- Message text: italic, `--gold`, system-ui

**Between exchanges:**
```
── ◆ ──
```
- Thin decorative divider, `--gold-dim`, centered, after each exchange pair

### Combat Panel
- Appears at top of center column when `ctx.active_combat` is set
- Styled with `--accent` left border, dark background — visually alarming, distinct from narrative

---

## Map Drawer

- Collapsed state: a thin bar at the bottom of the center column
  - Text: `[ THE IRONLANDS ▾ ]` in Cormorant, `--gold-dim`, centered
  - 1px `--border` top
- Expanded state: grows upward to fill ~60% of center column height
  - Map image fills the area
  - Pins rendered as before
  - `[ ▴ COLLAPSE ]` label to close
- Transition: CSS `max-height` animation, 300ms ease

---

## Right Sidebar — WorldNotesPanel + JournalPanel

### Tab Toggle
- Two text links: `NOTES` and `JOURNAL`
- Active tab: `--gold`, underline
- Inactive: `--text-dim`, no underline
- No button chrome — plain text toggle

### Notes Tab
- Search input: bottom-border only, no box, placeholder "Search notes..." in `--text-dim`
- Note cards: title in Cormorant `--gold`, content in system-ui `--text`, tags as small pill spans in `--gold-dim` border
- "Draft with AI" button: text-only, `--gold-dim`, appears only when `aiEnabled`

### Journal Tab
- Editable textarea: no border, `--text` on `--surface`, system-ui, full height
- Saves on blur via existing `patchSessionSummary()`
- "Generate recap" button: text-only, `--gold-dim`, only when `aiEnabled`

---

## Component Changes

| Component | Change |
|-----------|--------|
| `App.tsx` | New three-column layout; redistribute panels into sidebars and center; add map drawer state; add right sidebar tab state |
| `App.css` | Replace all CSS variables; add grimoire layout classes; rewrite state-bar |
| `CharacterSheetPanel.tsx` | Restyle only — pip dots and segmented bars replace plain inputs for stats/tracks; portrait hover behavior |
| `CharacterSheetPanel.css` | Full rewrite |
| `WorldNotesPanel.tsx` | Restyle only; remove outer panel wrapper (sidebar provides container) |
| `WorldNotesPanel.css` | Full rewrite |
| `JournalPanel.tsx` | Restyle only; remove outer panel wrapper |
| `JournalPanel.css` | Full rewrite |
| `DiceHistoryPanel.tsx` | Restyle to compact list; cap at 5 entries |
| `DiceHistoryPanel.css` | Full rewrite |
| `CombatPanel.tsx` | Restyle only; add accent left-border treatment |
| `CombatPanel.css` | Full rewrite |
| `SessionTimeline.tsx` | **Removed from layout** — component kept but not rendered |
| `index.html` | Add Google Fonts link for Cormorant Garamond |

---

## Out of Scope

- No changes to `api.ts`, `useWebSocket.ts`, `types.ts`
- No changes to any backend Go code
- No new API endpoints
- SessionTimeline component is hidden, not deleted
- No mobile/responsive layout (desktop only)
