# Phase A: Quick Wins — Design Spec

**Date:** 2026-04-04  
**Features:** 4 (session scratchpad, initiative tracker 2.0, sub-task objectives, session XP log)  
**Scope:** DB migrations, backend routes, frontend component changes — no AI involvement.

---

## Overview

Four small, self-contained improvements that extend existing tables and panels without introducing new architectural patterns.

---

## 1. Session Scratchpad

**Purpose:** Freeform per-session notes for the GM — quick jottings during play, separate from the permanent world notes.

### DB

```sql
ALTER TABLE sessions ADD COLUMN notes TEXT NOT NULL DEFAULT '';
```

### Backend

`PATCH /api/sessions/{id}` already exists. Add `notes` to the patchable body fields. No new route needed.

### Frontend

In **JournalPanel**, add a "Session Notes" textarea below the recap block. Autosave via debounced `PATCH /api/sessions/{id}` (300ms delay). No explicit save button — saving is invisible. Label: "Notes" with a small italic placeholder: *Quick notes for this session…*

---

## 2. Initiative Tracker 2.0

**Purpose:** Sort combatants by initiative and give the GM a manual "Next Turn" button to track whose action it is.

### DB

```sql
ALTER TABLE combat_encounters ADD COLUMN active_turn_index INTEGER NOT NULL DEFAULT 0;
```

`combatants.initiative` already exists (INTEGER, default 0).

### Backend

**New route:** `POST /api/combat-encounters/{id}/next-turn`

- Loads encounter, counts combatants sorted by `initiative DESC`.
- Increments `active_turn_index` modulo combatant count.
- Persists to DB.
- Publishes `turn_advanced` WS event: `{ encounter_id, active_turn_index }`.
- Returns `204 No Content`.

**Existing `GET /api/sessions/{id}` combat snapshot** — ensure `active_turn_index` is included in the `CombatSnapshot` response (add to `CombatEncounter` struct).

### Frontend

**CombatPanel:**
- Sort combatants by `initiative` descending before rendering.
- Highlight the row at `active_turn_index` with a distinct left-border accent (gold/amber).
- Add "Next Turn →" button (bottom of panel). On click: `POST .../next-turn`, then refetch combat state.
- Subscribe to `turn_advanced` WS event to update highlight without full refetch.

---

## 3. Sub-task Objectives

**Purpose:** Allow objectives to have nested sub-tasks (one level deep only — no recursive trees).

### DB

```sql
ALTER TABLE objectives ADD COLUMN parent_id INTEGER REFERENCES objectives(id);
```

`parent_id NULL` = top-level objective. `parent_id = N` = sub-task of objective N. No deeper nesting.

### Backend

`GET /api/campaigns/{id}/objectives` — no query change needed; `parent_id` is included in the row struct automatically. Frontend handles grouping.

`POST /api/campaigns/{id}/objectives` — body already accepts arbitrary fields; add optional `parent_id` field.

`DELETE /api/objectives/{id}` — when deleting a parent, cascade-delete its sub-tasks (add `ON DELETE CASCADE` to the FK, or handle in the query).

### Frontend

**ObjectivesPanel:**
- Group response: separate top-level objectives from sub-tasks by `parent_id`.
- Render top-level objective; indent sub-tasks beneath it with a left-border indent style.
- Each top-level objective row gets a small "＋" button that opens an inline "Add sub-task" input (name only, status defaults to `active`).
- Sub-tasks show their own status badge and delete button, same as parent objectives.
- Sub-tasks do NOT have their own "＋" button (one level only).

---

## 4. Session XP / Milestone Log

**Purpose:** Record what was achieved in a session — story milestones, XP earned, discoveries. System-agnostic: `note` is always required, `amount` is optional (nullable) for systems with numeric XP.

### DB

```sql
CREATE TABLE xp_log (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  session_id INTEGER NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
  note       TEXT NOT NULL,
  amount     INTEGER,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
```

### Backend

| Method | Route | Notes |
|--------|-------|-------|
| `GET /api/sessions/{id}/xp` | new | returns list, newest-first |
| `POST /api/sessions/{id}/xp` | new | body: `{ note, amount? }` — publishes `xp_added` WS event |
| `DELETE /api/xp/{id}` | new | 204; no WS event needed |

`xp_added` WS event payload: `{ session_id, id, note, amount }`.

### Frontend

**JournalPanel** (not a new panel) — below the scratchpad textarea, add a "Milestones" section:
- List of XP log entries (note + optional amount badge).
- "＋ Add milestone" inline form: note text input + optional number input labeled "XP".
- Subscribe to `xp_added` WS event to append new entries in real time.
- Delete button (×) on each entry.

---

## WS Events Summary

| Event | Payload | Trigger |
|-------|---------|---------|
| `turn_advanced` | `{ encounter_id, active_turn_index }` | `POST /api/combat-encounters/{id}/next-turn` |
| `xp_added` | `{ session_id, id, note, amount }` | `POST /api/sessions/{id}/xp` |

---

## What This Does NOT Include

- Drag-to-reorder objectives (sort_order deferred to later)
- Auto-detection of XP milestones from AI narration (deferred to Phase E)
- Roll-initiative button (combatants already have initiative set when combat starts via MCP tools; this feature only improves display and turn tracking)
- Player-visible turn indicator (future multi-player feature)
