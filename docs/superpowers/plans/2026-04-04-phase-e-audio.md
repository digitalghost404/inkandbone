# Phase E: Audio — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add two audio tiers: Tier 1 procedural UI sounds (dice rolls, notifications, page turns) via Web Audio API with no external files, and Tier 2 ambient scene atmosphere by playing local MP3s from `~/.ttrpg/audio/{tag}.mp3` served through the existing `/api/files/` route.

**Architecture:** Tier 1 is pure client-side Web Audio API synthesis — no backend changes, no file dependencies. Tier 2 adds a `scene_tags` column on sessions, a scene tag picker in the session header, and an ambient manager that fetches from the existing file route. A persistent mute/volume control stores preference in localStorage.

**Tech Stack:** React/TypeScript, Web Audio API, Go (DB migration + PATCH session extension)

---

## File Map

| File | Change |
|------|--------|
| `internal/db/migrations/014_phase_e.sql` | Create — sessions.scene_tags column |
| `internal/db/queries_session.go` | Modify — add SceneTags field to Session, UpdateSceneTags method |
| `internal/api/routes.go` | Modify — handlePatchSession accepts scene_tags |
| `web/src/audio/sounds.ts` | Create — Web Audio API procedural sound synthesis |
| `web/src/audio/ambient.ts` | Create — ambient audio manager |
| `web/src/AudioControls.tsx` | Create — mute/volume control component |
| `web/src/App.tsx` | Modify — wire sounds to events, add AudioControls to header |
| `web/src/types.ts` | Modify — add scene_tags to Session type |

---

### Task E1: Migration 014 — scene_tags on sessions

**Files:**
- Create: `internal/db/migrations/014_phase_e.sql`
- Modify: `internal/db/queries_session.go`

- [ ] **Step 1: Write the migration**

Create `internal/db/migrations/014_phase_e.sql`:

```sql
-- 014_phase_e.sql: Scene tags for ambient audio
ALTER TABLE sessions ADD COLUMN scene_tags TEXT NOT NULL DEFAULT '';
```

- [ ] **Step 2: Add SceneTags to Session struct and queries**

In `internal/db/queries_session.go`:

Add `SceneTags string \`json:"scene_tags"\`` to the `Session` struct after the `Notes` field.

Update all SELECT queries that scan Session rows to include `scene_tags`. There are three scan sites: `GetSession`, `ListSessions`. Update both:

For `GetSession`:
```go
err := d.db.QueryRow(
    "SELECT id, campaign_id, title, date, summary, notes, scene_tags, created_at FROM sessions WHERE id = ?", id,
).Scan(&s.ID, &s.CampaignID, &s.Title, &s.Date, &s.Summary, &s.Notes, &s.SceneTags, &s.CreatedAt)
```

For `ListSessions` rows.Scan:
```go
if err := rows.Scan(&s.ID, &s.CampaignID, &s.Title, &s.Date, &s.Summary, &s.Notes, &s.SceneTags, &s.CreatedAt); err != nil {
```
(Also update the SELECT in ListSessions to include `scene_tags`.)

Add the new update method at the end of the session functions:

```go
// UpdateSceneTags sets the scene_tags for a session.
// tags is a comma-separated list of tags (e.g. "tavern,night,rain").
func (d *DB) UpdateSceneTags(id int64, tags string) error {
    res, err := d.db.Exec("UPDATE sessions SET scene_tags = ? WHERE id = ?", tags, id)
    if err != nil {
        return err
    }
    n, err := res.RowsAffected()
    if err != nil {
        return err
    }
    if n == 0 {
        return fmt.Errorf("session %d not found", id)
    }
    return nil
}
```

- [ ] **Step 3: Write a test for UpdateSceneTags**

Add to `internal/db/queries_session_test.go` (create this file if it doesn't exist, following the `newTestDB(t)` pattern):

```go
func TestUpdateSceneTags(t *testing.T) {
    d := newTestDB(t)
    campID := setupCampaign(t, d)
    sessID, err := d.CreateSession(campID, "Session 1", "2026-04-03")
    require.NoError(t, err)

    require.NoError(t, d.UpdateSceneTags(sessID, "tavern,night"))

    sess, err := d.GetSession(sessID)
    require.NoError(t, err)
    assert.Equal(t, "tavern,night", sess.SceneTags)
}
```

- [ ] **Step 4: Run test**

Run: `go test ./internal/db/... -run "TestUpdateSceneTags" -v`
Expected: PASS

- [ ] **Step 5: Run full suite**

Run: `make test`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/db/migrations/014_phase_e.sql internal/db/queries_session.go
git commit -m "feat(db): migration 014 — scene_tags on sessions, UpdateSceneTags method"
```

---

### Task E2: Extend handlePatchSession to accept scene_tags

**Files:**
- Modify: `internal/api/routes.go`
- Modify: `internal/api/routes_test.go`

- [ ] **Step 1: Write the failing test**

Add to `internal/api/routes_test.go`:

```go
func TestPatchSessionSceneTags(t *testing.T) {
    s := newTestServer(t)
    _, sessID := seedCampaign(t, s.db)

    body := `{"scene_tags":"dungeon,dark"}`
    req := httptest.NewRequest(http.MethodPatch,
        fmt.Sprintf("/api/sessions/%d", sessID),
        strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    s.ServeHTTP(w, req)
    assert.Equal(t, http.StatusNoContent, w.Code)

    sess, err := s.db.GetSession(sessID)
    require.NoError(t, err)
    assert.Equal(t, "dungeon,dark", sess.SceneTags)
}
```

- [ ] **Step 2: Run test — expect fail**

Run: `go test ./internal/api/... -run "TestPatchSessionSceneTags" -v`
Expected: FAIL — scene_tags not persisted (field not handled)

- [ ] **Step 3: Extend handlePatchSession**

In `internal/api/routes.go`, find `handlePatchSession` (around line 384). Add `SceneTags *string \`json:"scene_tags"\`` to the body struct:

```go
var body struct {
    Summary   *string `json:"summary"`
    Notes     *string `json:"notes"`
    SceneTags *string `json:"scene_tags"`
}
```

After the existing `body.Notes != nil` block, add:

```go
if body.SceneTags != nil {
    if err := s.db.UpdateSceneTags(id, *body.SceneTags); err != nil {
        if strings.Contains(err.Error(), "not found") {
            http.Error(w, err.Error(), http.StatusNotFound)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    payload["scene_tags"] = *body.SceneTags
}
```

- [ ] **Step 4: Run test**

Run: `go test ./internal/api/... -run "TestPatchSessionSceneTags" -v`
Expected: PASS

- [ ] **Step 5: Run full suite**

Run: `make test`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/api/routes.go internal/api/routes_test.go
git commit -m "feat(api): handlePatchSession accepts scene_tags for ambient audio"
```

---

### Task E3: Web Audio API procedural sounds (Tier 1)

**Files:**
- Create: `web/src/audio/sounds.ts`

No backend changes. All synthesis is in-browser via Web Audio API.

- [ ] **Step 1: Create the sounds module**

Create `web/src/audio/sounds.ts`:

```typescript
// Procedural UI sounds using Web Audio API — no audio files required.

let ctx: AudioContext | null = null;
let masterGain: GainNode | null = null;
let muted = false;

function getCtx(): AudioContext {
  if (!ctx) {
    ctx = new AudioContext();
    masterGain = ctx.createGain();
    masterGain.connect(ctx.destination);
    masterGain.gain.value = getVolume();
  }
  return ctx;
}

export function setMuted(m: boolean): void {
  muted = m;
  localStorage.setItem('audio_muted', m ? '1' : '0');
}

export function isMuted(): boolean {
  return muted;
}

export function setVolume(v: number): void {
  localStorage.setItem('audio_volume', String(v));
  if (masterGain) masterGain.gain.value = v;
}

export function getVolume(): number {
  return parseFloat(localStorage.getItem('audio_volume') ?? '0.5');
}

// Initialize muted state from localStorage
muted = localStorage.getItem('audio_muted') === '1';

function play(fn: (ctx: AudioContext, gain: GainNode) => void): void {
  if (muted) return;
  try {
    const c = getCtx();
    if (c.state === 'suspended') c.resume();
    fn(c, masterGain!);
  } catch {
    // Web Audio not available (e.g. test environment)
  }
}

// Dice roll: short percussive rattle (3 quick clicks)
export function playDiceRoll(): void {
  play((c, g) => {
    [0, 0.06, 0.12].forEach(delay => {
      const buf = c.createBuffer(1, c.sampleRate * 0.04, c.sampleRate);
      const data = buf.getChannelData(0);
      for (let i = 0; i < data.length; i++) {
        data[i] = (Math.random() * 2 - 1) * Math.exp(-i / (data.length * 0.3));
      }
      const src = c.createBufferSource();
      src.buffer = buf;
      const gainNode = c.createGain();
      gainNode.gain.value = 0.4;
      src.connect(gainNode);
      gainNode.connect(g);
      src.start(c.currentTime + delay);
    });
  });
}

// Notification chime: ascending two-tone
export function playNotification(): void {
  play((c, g) => {
    [440, 660].forEach((freq, i) => {
      const osc = c.createOscillator();
      const gainNode = c.createGain();
      osc.type = 'sine';
      osc.frequency.value = freq;
      gainNode.gain.setValueAtTime(0, c.currentTime + i * 0.12);
      gainNode.gain.linearRampToValueAtTime(0.25, c.currentTime + i * 0.12 + 0.02);
      gainNode.gain.exponentialRampToValueAtTime(0.001, c.currentTime + i * 0.12 + 0.3);
      osc.connect(gainNode);
      gainNode.connect(g);
      osc.start(c.currentTime + i * 0.12);
      osc.stop(c.currentTime + i * 0.12 + 0.3);
    });
  });
}

// Page turn: soft paper swoosh
export function playPageTurn(): void {
  play((c, g) => {
    const buf = c.createBuffer(1, c.sampleRate * 0.15, c.sampleRate);
    const data = buf.getChannelData(0);
    for (let i = 0; i < data.length; i++) {
      const t = i / data.length;
      const envelope = Math.sin(Math.PI * t);
      data[i] = (Math.random() * 2 - 1) * envelope * 0.15;
    }
    const src = c.createBufferSource();
    src.buffer = buf;
    // Band-pass to make it sound papery
    const bpf = c.createBiquadFilter();
    bpf.type = 'bandpass';
    bpf.frequency.value = 3000;
    bpf.Q.value = 0.5;
    src.connect(bpf);
    bpf.connect(g);
    src.start();
  });
}

// Combat start: ominous low pulse
export function playCombatStart(): void {
  play((c, g) => {
    const osc = c.createOscillator();
    const gainNode = c.createGain();
    osc.type = 'sawtooth';
    osc.frequency.setValueAtTime(80, c.currentTime);
    osc.frequency.exponentialRampToValueAtTime(40, c.currentTime + 0.5);
    gainNode.gain.setValueAtTime(0.3, c.currentTime);
    gainNode.gain.exponentialRampToValueAtTime(0.001, c.currentTime + 0.6);
    osc.connect(gainNode);
    gainNode.connect(g);
    osc.start();
    osc.stop(c.currentTime + 0.6);
  });
}
```

- [ ] **Step 2: Verify TypeScript compiles**

Run: `make build`
Expected: Build succeeds (TypeScript compilation passes)

- [ ] **Step 3: Commit**

```bash
git add web/src/audio/sounds.ts
git commit -m "feat(audio): Tier 1 Web Audio API procedural sounds — dice, notification, page turn, combat"
```

---

### Task E4: Ambient audio manager (Tier 2)

**Files:**
- Create: `web/src/audio/ambient.ts`

The ambient manager plays looping audio from `~/.ttrpg/audio/{tag}.mp3` via the existing `/api/files/` route. The `dataDir` is `~/.ttrpg`, so a file at `~/.ttrpg/audio/tavern.mp3` is served at `/api/files/audio/tavern.mp3`.

- [ ] **Step 1: Create the ambient module**

Create `web/src/audio/ambient.ts`:

```typescript
// Ambient audio manager — loops local MP3 files served via /api/files/audio/{tag}.mp3

import { isMuted, getVolume } from './sounds';

let currentTag: string | null = null;
let audioEl: HTMLAudioElement | null = null;
let fadeInterval: ReturnType<typeof setInterval> | null = null;

const FADE_STEPS = 20;
const FADE_INTERVAL_MS = 50;

function stopFade(): void {
  if (fadeInterval !== null) {
    clearInterval(fadeInterval);
    fadeInterval = null;
  }
}

function fadeOut(el: HTMLAudioElement, onDone: () => void): void {
  stopFade();
  const startVol = el.volume;
  let step = 0;
  fadeInterval = setInterval(() => {
    step++;
    el.volume = Math.max(0, startVol * (1 - step / FADE_STEPS));
    if (step >= FADE_STEPS) {
      stopFade();
      el.pause();
      el.src = '';
      onDone();
    }
  }, FADE_INTERVAL_MS);
}

function fadeIn(el: HTMLAudioElement): void {
  stopFade();
  el.volume = 0;
  const targetVol = isMuted() ? 0 : getVolume() * 0.5; // ambient at half of master
  let step = 0;
  fadeInterval = setInterval(() => {
    step++;
    el.volume = Math.min(targetVol, targetVol * (step / FADE_STEPS));
    if (step >= FADE_STEPS) stopFade();
  }, FADE_INTERVAL_MS);
}

export function setAmbientTag(tag: string | null): void {
  if (tag === currentTag) return;

  if (audioEl && currentTag) {
    const oldEl = audioEl;
    fadeOut(oldEl, () => {
      oldEl.remove();
    });
    audioEl = null;
  }

  currentTag = tag;

  if (!tag) return;

  const el = new Audio(`/api/files/audio/${encodeURIComponent(tag)}.mp3`);
  el.loop = true;
  el.volume = 0;
  audioEl = el;

  el.onerror = () => {
    // File doesn't exist — silently stop
    currentTag = null;
    audioEl = null;
  };

  el.play()
    .then(() => fadeIn(el))
    .catch(() => {
      // Autoplay blocked — will retry on user interaction
    });
}

export function stopAmbient(): void {
  setAmbientTag(null);
}

export function syncAmbientVolume(): void {
  if (!audioEl) return;
  const target = isMuted() ? 0 : getVolume() * 0.5;
  audioEl.volume = target;
}
```

- [ ] **Step 2: Build**

Run: `make build`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add web/src/audio/ambient.ts
git commit -m "feat(audio): Tier 2 ambient manager — fade in/out loop from /api/files/audio/"
```

---

### Task E5: AudioControls component — mute + volume in header

**Files:**
- Create: `web/src/AudioControls.tsx`
- Modify: `web/src/App.tsx`

- [ ] **Step 1: Create AudioControls.tsx**

Create `web/src/AudioControls.tsx`:

```tsx
import React from 'react';
import { setMuted, isMuted, setVolume, getVolume } from './audio/sounds';
import { syncAmbientVolume } from './audio/ambient';

export default function AudioControls() {
  const [muted, setMutedState] = React.useState(isMuted);
  const [volume, setVolumeState] = React.useState(getVolume);

  const handleMuteToggle = () => {
    const newMuted = !muted;
    setMuted(newMuted);
    setMutedState(newMuted);
    syncAmbientVolume();
  };

  const handleVolumeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const v = parseFloat(e.target.value);
    setVolume(v);
    setVolumeState(v);
    syncAmbientVolume();
  };

  return (
    <div className="audio-controls">
      <button
        className="mute-btn"
        onClick={handleMuteToggle}
        title={muted ? 'Unmute' : 'Mute'}
        aria-label={muted ? 'Unmute audio' : 'Mute audio'}
      >
        {muted ? '🔇' : '🔊'}
      </button>
      {!muted && (
        <input
          type="range"
          min="0"
          max="1"
          step="0.05"
          value={volume}
          onChange={handleVolumeChange}
          className="volume-slider"
          aria-label="Volume"
          title={`Volume: ${Math.round(volume * 100)}%`}
        />
      )}
    </div>
  );
}
```

Add CSS:

```css
.audio-controls {
  display: flex;
  align-items: center;
  gap: 6px;
}
.mute-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 1rem;
  padding: 2px 4px;
  border-radius: 3px;
  line-height: 1;
}
.mute-btn:hover { background: rgba(255,255,255,0.08); }
.volume-slider {
  width: 70px;
  height: 4px;
  accent-color: var(--gold, #c9a84c);
  cursor: pointer;
}
```

- [ ] **Step 2: Add AudioControls to App.tsx header**

In `web/src/App.tsx`, import AudioControls:

```typescript
import AudioControls from './AudioControls';
```

Find the header/toolbar area and add `<AudioControls />` next to the existing theme toggle or at the end of the header controls row.

- [ ] **Step 3: Build**

Run: `make build`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add web/src/AudioControls.tsx web/src/App.tsx
git commit -m "feat(ui): AudioControls component — mute toggle + volume slider in header"
```

---

### Task E6: Wire sounds to WebSocket events in App.tsx

**Files:**
- Modify: `web/src/App.tsx`
- Modify: `web/src/types.ts`

- [ ] **Step 1: Add scene_tags to Session type**

In `web/src/types.ts`, add `scene_tags: string` to the `Session` interface.

- [ ] **Step 2: Wire sound triggers in App.tsx**

In `web/src/App.tsx`, import sound functions:

```typescript
import { playDiceRoll, playNotification, playCombatStart, playPageTurn } from './audio/sounds';
import { setAmbientTag } from './audio/ambient';
```

In the WebSocket message handler (or wherever `lastEvent` is set after receiving a WS message), add sound triggers based on event type. Find the section where event.type is checked (likely a switch or if-chain) and add:

```typescript
// Add to wherever WS events are processed:
switch (event.type) {
  case 'dice_rolled':
    playDiceRoll();
    break;
  case 'message_created':
    if (event.payload?.role === 'assistant') {
      playNotification();
    }
    break;
  case 'combat_started':
    playCombatStart();
    break;
  // ... existing cases ...
}
```

- [ ] **Step 3: Update ambient when session scene_tags change**

In App.tsx, find where the active session is loaded/updated. After loading or patching a session, update the ambient:

```typescript
// After loading active session:
if (session?.scene_tags) {
  const tags = session.scene_tags.split(',').map(t => t.trim()).filter(Boolean);
  setAmbientTag(tags[0] ?? null); // play first tag's audio
} else {
  setAmbientTag(null);
}
```

Also hook into the `session_updated` event:

```typescript
case 'session_updated':
  if (event.payload?.scene_tags !== undefined) {
    const tags = (event.payload.scene_tags as string).split(',').map((t: string) => t.trim()).filter(Boolean);
    setAmbientTag(tags[0] ?? null);
  }
  break;
```

- [ ] **Step 4: Build**

Run: `make build`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/src/App.tsx web/src/types.ts
git commit -m "feat(audio): wire sound effects to WS events + ambient to session scene_tags"
```

---

### Task E7: Scene tag picker in session header

**Files:**
- Modify: `web/src/SessionView.tsx` (or wherever the active session header/controls are rendered — the component that shows session title/date and contains the scratchpad button)

The scene tag picker is a compact multi-select or tag chips UI that patches the session's `scene_tags` when changed.

- [ ] **Step 1: Add scene tag picker to session header**

In the session header component, add the tag picker alongside other session controls:

```tsx
import React from 'react';

const SCENE_TAGS = [
  'tavern', 'dungeon', 'forest', 'city', 'ocean', 'cave', 'castle',
  'rain', 'night', 'battle', 'market', 'temple', 'ruins',
];

// Inside the component:
const [sceneTags, setSceneTags] = React.useState<string[]>(
  session?.scene_tags ? session.scene_tags.split(',').filter(Boolean) : []
);

const toggleTag = async (tag: string) => {
  const newTags = sceneTags.includes(tag)
    ? sceneTags.filter(t => t !== tag)
    : [...sceneTags, tag];
  setSceneTags(newTags);
  // Patch the session
  await fetch(`/api/sessions/${session.id}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ scene_tags: newTags.join(',') }),
  });
};

// In JSX:
<div className="scene-tags-picker">
  <span className="scene-tags-label">🎵 Scene:</span>
  <div className="scene-tags-chips">
    {SCENE_TAGS.map(tag => (
      <button
        key={tag}
        className={`scene-tag-chip ${sceneTags.includes(tag) ? 'active' : ''}`}
        onClick={() => toggleTag(tag)}
      >
        {tag}
      </button>
    ))}
  </div>
</div>
```

Add CSS:

```css
.scene-tags-picker {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  padding: 4px 0;
}
.scene-tags-label {
  font-size: 0.75rem;
  color: var(--text-muted, #888);
  white-space: nowrap;
}
.scene-tags-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}
.scene-tag-chip {
  font-size: 0.7rem;
  padding: 2px 7px;
  border-radius: 10px;
  border: 1px solid var(--border, #444);
  background: transparent;
  color: var(--text-muted, #888);
  cursor: pointer;
}
.scene-tag-chip.active {
  background: rgba(201, 168, 76, 0.15);
  border-color: var(--gold, #c9a84c);
  color: var(--gold, #c9a84c);
}
```

- [ ] **Step 2: Build + test**

Run: `make test && make build`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add web/src/SessionView.tsx
git commit -m "feat(ui): scene tag picker in session header for ambient audio selection"
```

---

## Manual Smoke Test

After all tasks are complete, run a quick manual check:

1. Start the server: `ttrpg` (or `make dev`)
2. Open the app, start a session
3. Roll dice → hear dice rattle sound
4. GM responds → hear notification chime
5. Toggle mute button in header → sounds stop
6. Adjust volume slider → volume changes
7. Click scene tags (e.g., "tavern") → if `~/.ttrpg/audio/tavern.mp3` exists, ambient plays; if not, no error
8. Clear scene tags → ambient stops with fade-out
