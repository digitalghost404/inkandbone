# Auto Scene Tags & Ambient Pause/Resume Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** After each GM response, Claude automatically classifies the scene and sets the session's scene tag to drive ambient audio, while players gain a dedicated pause/resume button for ambient music.

**Architecture:** A new `autoUpdateSceneTags` goroutine fires alongside existing post-GM goroutines in `handleGMRespondStream`; it calls `completer.Generate` with a classification prompt, validates the returned tag against the allowed set, skips the write if the active tag is unchanged (stability), then calls `db.UpdateSceneTags` and publishes `EventSessionUpdated`. The frontend's existing `handleEvent → loadContext → useEffect → setAmbientTrack` chain picks up the change automatically. A new pause/resume button in `AudioControls` calls `pauseAmbient()`/`resumeAmbient()` in `ambient.ts`, which respect the paused state on track switches.

**Tech Stack:** Go (backend goroutine), ai.Completer interface, TypeScript/React (AudioControls, ambient.ts)

---

## File Map

- Modify: `internal/api/routes.go` — add `autoUpdateSceneTags` function; add goroutine call in `handleGMRespondStream`
- Modify: `internal/api/routes_test.go` — add `TestAutoUpdateSceneTags_*` tests
- Modify: `web/src/audio/ambient.ts` — add `paused` state, `pauseAmbient`, `resumeAmbient` exports; update `setAmbientTrack` and `setAmbientMuted` to respect paused state
- Modify: `web/src/AudioControls.tsx` — add ambientPaused state, ⏸/▶ button, localStorage persistence

---

### Task 1: `autoUpdateSceneTags` goroutine

**Files:**
- Modify: `internal/api/routes.go`
- Test: `internal/api/routes_test.go`

This follows the exact pattern of `autoUpdateTension` (keyword) and `autoDetectObjectives` (AI call). The function is on the `*Server` receiver and uses `s.aiClient.(ai.Completer)`.

- [ ] **Step 1: Write the failing tests**

Add to `internal/api/routes_test.go`:

```go
func TestAutoUpdateSceneTags_setsTag(t *testing.T) {
	stub := &stubCompleter{response: `{"tag":"dungeon"}`}
	s := newTestServerWithAI(t, stub)
	_, sessID := seedCampaign(t, s.db)

	s.autoUpdateSceneTags(context.Background(), sessID, "You descend into the stone corridor.")

	sess, err := s.db.GetSession(sessID)
	require.NoError(t, err)
	assert.Equal(t, "dungeon", sess.SceneTags)
}

func TestAutoUpdateSceneTags_stability_noOpWhenSameTag(t *testing.T) {
	stub := &stubCompleter{response: `{"tag":"dungeon"}`}
	s := newTestServerWithAI(t, stub)
	_, sessID := seedCampaign(t, s.db)
	require.NoError(t, s.db.UpdateSceneTags(sessID, "dungeon"))

	s.autoUpdateSceneTags(context.Background(), sessID, "The dungeon corridor stretches ahead.")

	sess, err := s.db.GetSession(sessID)
	require.NoError(t, err)
	// Tags unchanged — no DB write happened (still "dungeon", not re-written)
	assert.Equal(t, "dungeon", sess.SceneTags)
}

func TestAutoUpdateSceneTags_invalidTagIgnored(t *testing.T) {
	stub := &stubCompleter{response: `{"tag":"spaceship"}`}
	s := newTestServerWithAI(t, stub)
	_, sessID := seedCampaign(t, s.db)

	s.autoUpdateSceneTags(context.Background(), sessID, "You board the vessel.")

	sess, err := s.db.GetSession(sessID)
	require.NoError(t, err)
	assert.Equal(t, "", sess.SceneTags)
}

func TestAutoUpdateSceneTags_nilAI_noOp(t *testing.T) {
	s := newTestServer(t) // no AI client
	_, sessID := seedCampaign(t, s.db)

	// Should not panic
	s.autoUpdateSceneTags(context.Background(), sessID, "The forest rustles.")

	sess, err := s.db.GetSession(sessID)
	require.NoError(t, err)
	assert.Equal(t, "", sess.SceneTags)
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd /home/digitalghost/projects/inkandbone
go test ./internal/api/ -run TestAutoUpdateSceneTags -v
```

Expected: FAIL — `s.autoUpdateSceneTags undefined`

- [ ] **Step 3: Implement `autoUpdateSceneTags` in `routes.go`**

Add this function at the end of `internal/api/routes.go` (after `autoUpdateTension`):

```go
// autoUpdateSceneTags uses the AI to classify the scene from each GM response
// and updates the session's scene_tags to drive ambient audio. Skips the write
// if the active (first) tag is unchanged (stability — avoids restarting the track
// mid-scene). Runs in a background goroutine.
func (s *Server) autoUpdateSceneTags(ctx context.Context, sessionID int64, gmText string) {
	completer, ok := s.aiClient.(ai.Completer)
	if !ok {
		return
	}
	if gmText == "" {
		return
	}

	sess, err := s.db.GetSession(sessionID)
	if err != nil || sess == nil {
		return
	}

	prompt := fmt.Sprintf(`You are a scene classifier for a tabletop RPG. Based on the GM's narrative below, select the single most fitting scene tag.

Valid tags: tavern, dungeon, forest, city, ocean, cave, castle, rain, night, battle, market, temple, ruins

Rules:
- Return exactly one tag from the list above
- Choose based on the dominant environment or mood
- If no tag fits well, return the closest match
- Return only JSON: {"tag":"<chosen_tag>"}
- No explanation, no markdown

GM text:
%s`, gmText)

	raw, err := completer.Generate(ctx, prompt)
	if err != nil {
		return
	}

	raw = strings.TrimSpace(raw)
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return
	}

	var result struct {
		Tag string `json:"tag"`
	}
	if err := json.Unmarshal([]byte(raw[start:end+1]), &result); err != nil {
		return
	}

	validTags := map[string]bool{
		"tavern": true, "dungeon": true, "forest": true, "city": true,
		"ocean": true, "cave": true, "castle": true, "rain": true,
		"night": true, "battle": true, "market": true, "temple": true,
		"ruins": true,
	}
	if !validTags[result.Tag] {
		return
	}

	// Tag stability: skip if the active (first) tag is unchanged
	currentFirst := ""
	if sess.SceneTags != "" {
		currentFirst = strings.SplitN(sess.SceneTags, ",", 2)[0]
	}
	if result.Tag == currentFirst {
		return
	}

	if err := s.db.UpdateSceneTags(sessionID, result.Tag); err != nil {
		return
	}
	s.bus.Publish(Event{Type: EventSessionUpdated, Payload: map[string]any{
		"session_id": sessionID,
		"scene_tags": result.Tag,
	}})
}
```

- [ ] **Step 4: Wire the goroutine into `handleGMRespondStream`**

In `internal/api/routes.go`, find the goroutine launch block (around line 1058):

```go
	go s.extractNPCs(context.Background(), id, fullText)
	go s.autoGenerateMap(context.Background(), id, fullText)
	go s.autoUpdateCharacterStats(context.Background(), id, lastPlayerMsg, fullText)
	go s.autoUpdateRecap(context.Background(), id)
	go s.autoDetectObjectives(context.Background(), id, fullText)
	go s.autoExtractItems(context.Background(), id, fullText)
	tensionText := fullText
	if roll != nil && !roll.Success {
		tensionText = "critical failure " + fullText
	}
	go s.autoUpdateTension(id, tensionText)
```

Add one line after `autoUpdateTension`:

```go
	go s.autoUpdateSceneTags(context.Background(), id, fullText)
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
cd /home/digitalghost/projects/inkandbone
go test ./internal/api/ -run TestAutoUpdateSceneTags -v
```

Expected: 4 tests PASS

- [ ] **Step 6: Run full test suite**

```bash
cd /home/digitalghost/projects/inkandbone
go test ./...
```

Expected: all tests pass, 0 failures

- [ ] **Step 7: Commit**

```bash
cd /home/digitalghost/projects/inkandbone
git add internal/api/routes.go internal/api/routes_test.go
git commit -m "feat: auto-classify scene tags via AI after each GM response"
```

---

### Task 2: `ambient.ts` pause/resume

**Files:**
- Modify: `web/src/audio/ambient.ts`

Add module-level `paused` state and two new exports. Update `setAmbientTrack` so it loads the audio element but does not play when paused. Update `setAmbientMuted` so unmuting respects paused state.

- [ ] **Step 1: Replace `web/src/audio/ambient.ts` with the updated version**

```ts
// ambient.ts — Ambient audio loop manager with fade in/out

const FADE_STEPS = 20;
const FADE_INTERVAL_MS = 50;
const MAX_VOLUME = 0.6;

interface AmbientTrack {
  audio: HTMLAudioElement;
  tag: string;
}

let currentTrack: AmbientTrack | null = null;
let masterVolume = 1.0;
let muted = false;
let paused = false;

export function setAmbientVolume(volume: number): void {
  masterVolume = Math.max(0, Math.min(1, volume));
  if (currentTrack && !muted && !paused) {
    currentTrack.audio.volume = masterVolume * MAX_VOLUME;
  }
}

export function setAmbientMuted(isMuted: boolean): void {
  muted = isMuted;
  if (currentTrack) {
    if (muted) {
      currentTrack.audio.volume = 0;
    } else if (!paused) {
      currentTrack.audio.volume = masterVolume * MAX_VOLUME;
    }
  }
}

export function pauseAmbient(): void {
  paused = true;
  if (currentTrack) {
    currentTrack.audio.pause();
  }
}

export function resumeAmbient(): void {
  paused = false;
  if (currentTrack && !muted) {
    currentTrack.audio.play().catch(() => {});
  }
}

function fadeOut(audio: HTMLAudioElement): Promise<void> {
  return new Promise(resolve => {
    const startVol = audio.volume;
    const step = startVol / FADE_STEPS;
    let count = 0;
    const interval = setInterval(() => {
      count++;
      audio.volume = Math.max(0, startVol - step * count);
      if (count >= FADE_STEPS) {
        clearInterval(interval);
        audio.pause();
        audio.currentTime = 0;
        resolve();
      }
    }, FADE_INTERVAL_MS);
  });
}

function fadeIn(audio: HTMLAudioElement, targetVolume: number): void {
  audio.volume = 0;
  audio.play().catch(() => {});
  const step = targetVolume / FADE_STEPS;
  let count = 0;
  const interval = setInterval(() => {
    count++;
    audio.volume = Math.min(targetVolume, step * count);
    if (count >= FADE_STEPS) {
      clearInterval(interval);
    }
  }, FADE_INTERVAL_MS);
}

export async function setAmbientTrack(tag: string | null): Promise<void> {
  // If same tag, do nothing
  if (currentTrack && tag === currentTrack.tag) return;

  // Fade out current track
  if (currentTrack) {
    await fadeOut(currentTrack.audio);
    currentTrack = null;
  }

  if (!tag) return;

  // Load new track
  const audio = new Audio(`/api/files/audio/${tag}.mp3`);
  audio.loop = true;
  currentTrack = { audio, tag };

  // Respect paused and muted state — load but don't play if paused
  if (!paused && !muted) {
    fadeIn(audio, masterVolume * MAX_VOLUME);
  }
}

export function stopAmbient(): void {
  if (currentTrack) {
    currentTrack.audio.pause();
    currentTrack.audio.currentTime = 0;
    currentTrack = null;
  }
}
```

- [ ] **Step 2: Verify TypeScript compiles**

```bash
cd /home/digitalghost/projects/inkandbone/web
npm run build 2>&1 | tail -10
```

Expected: build succeeds, no TypeScript errors

- [ ] **Step 3: Commit**

```bash
cd /home/digitalghost/projects/inkandbone
git add web/src/audio/ambient.ts
git commit -m "feat: add pauseAmbient/resumeAmbient; respect paused state on track switch"
```

---

### Task 3: AudioControls pause/resume button

**Files:**
- Modify: `web/src/AudioControls.tsx`

Add `ambientPaused` state persisted to `localStorage`. Render a ⏸/▶ button. Initialize pause state from localStorage on mount. Button is disabled while muted.

- [ ] **Step 1: Replace `web/src/AudioControls.tsx` with the updated version**

```tsx
import { useState, useEffect } from 'react';
import { setAmbientVolume, setAmbientMuted, pauseAmbient, resumeAmbient } from './audio/ambient';

const STORAGE_KEY_MUTED = 'inkandbone_audio_muted';
const STORAGE_KEY_VOLUME = 'inkandbone_audio_volume';
const STORAGE_KEY_PAUSED = 'inkandbone_ambient_paused';

// Exported state accessor for sounds.ts and other consumers
export function getAudioMuted(): boolean {
  return localStorage.getItem(STORAGE_KEY_MUTED) === 'true';
}

export function getAudioVolume(): number {
  const v = localStorage.getItem(STORAGE_KEY_VOLUME);
  return v ? parseFloat(v) : 0.7;
}

export default function AudioControls() {
  const [muted, setMuted] = useState<boolean>(() => getAudioMuted());
  const [volume, setVolume] = useState<number>(() => getAudioVolume());
  const [ambientPaused, setAmbientPaused] = useState<boolean>(
    () => localStorage.getItem(STORAGE_KEY_PAUSED) === 'true'
  );

  // Sync muted state to ambient module
  useEffect(() => {
    setAmbientMuted(muted);
    localStorage.setItem(STORAGE_KEY_MUTED, String(muted));
  }, [muted]);

  // Sync volume to ambient module
  useEffect(() => {
    setAmbientVolume(volume);
    localStorage.setItem(STORAGE_KEY_VOLUME, String(volume));
  }, [volume]);

  // Initialize ambient pause state on mount from localStorage
  useEffect(() => {
    if (localStorage.getItem(STORAGE_KEY_PAUSED) === 'true') {
      pauseAmbient();
    }
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  function toggleAmbientPause() {
    const next = !ambientPaused;
    setAmbientPaused(next);
    localStorage.setItem(STORAGE_KEY_PAUSED, String(next));
    if (next) {
      pauseAmbient();
    } else {
      resumeAmbient();
    }
  }

  return (
    <div className="audio-controls">
      <button
        onClick={() => setMuted(m => !m)}
        title={muted ? 'Unmute audio' : 'Mute audio'}
        className="audio-mute-btn"
      >
        {muted ? '🔇' : '🔊'}
      </button>
      <input
        type="range"
        min={0}
        max={1}
        step={0.05}
        value={volume}
        onChange={e => setVolume(parseFloat(e.target.value))}
        disabled={muted}
        title="Volume"
        className="audio-volume-slider"
      />
      <button
        onClick={toggleAmbientPause}
        disabled={muted}
        title={ambientPaused ? 'Resume ambient music' : 'Pause ambient music'}
        className="audio-pause-btn"
      >
        {ambientPaused ? '▶' : '⏸'}
      </button>
    </div>
  );
}
```

- [ ] **Step 2: Verify TypeScript compiles and builds**

```bash
cd /home/digitalghost/projects/inkandbone/web
npm run build 2>&1 | tail -10
```

Expected: build succeeds, `dist/assets/index-*.js` generated with no errors

- [ ] **Step 3: Run Go tests to ensure nothing broken**

```bash
cd /home/digitalghost/projects/inkandbone
go test ./...
```

Expected: all pass

- [ ] **Step 4: Build full binary**

```bash
cd /home/digitalghost/projects/inkandbone
make build
```

Expected: `ttrpg` binary produced

- [ ] **Step 5: Commit**

```bash
cd /home/digitalghost/projects/inkandbone
git add web/src/AudioControls.tsx
git commit -m "feat: add ambient pause/resume button to AudioControls"
```
