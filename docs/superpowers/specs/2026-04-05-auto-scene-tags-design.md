# Auto Scene Tags & Ambient Audio Controls Design

## Goal

After every GM response, Claude automatically classifies the scene and sets the session's scene tags, driving ambient audio track selection without any player interaction. Players retain manual override via the existing SceneTagPicker buttons, and gain dedicated pause/resume and volume controls for ambient music.

## Architecture

### Backend: `autoUpdateSceneTags` goroutine

A new goroutine added to the post-GM-response automation chain in `handleGMRespondStream` (alongside `autoUpdateTension`, `autoExtractItems`, etc.).

**Trigger:** Fires after every complete GM response, same as all other auto goroutines.

**Early exits:**
- `s.aiClient == nil` (AI not configured)
- `gmText == ""` (empty response)

**Tag stability:** Before updating, fetch the session's current `scene_tags`. Compute the new tag from Claude's response. If the new tag matches the first entry of the current tags (i.e., the active track tag is unchanged), skip the DB write and WebSocket publish entirely. This prevents the ambient track from restarting mid-scene just because Claude picked the same tag again.

**Classification prompt:** A minimal, low-token prompt sent to Claude Haiku:

```
You are a scene classifier for a tabletop RPG. Based on the GM's narrative below, select the single most fitting scene tag from this list:
tavern, dungeon, forest, city, ocean, cave, castle, rain, night, battle, market, temple, ruins

Rules:
- Return exactly one tag from the list above
- Choose based on the dominant environment/mood
- If no tag fits well, return the closest match
- Return only JSON: {"tag":"<chosen_tag>"}

GM text:
<gmText>
```

**DB update:** `db.UpdateSceneTags(sessionID, newTag)` — sets `scene_tags` to the single AI-chosen tag (overwriting any previous value, including player-set multi-tag strings).

**WebSocket event:** Publishes `EventSessionUpdated` with `{"session_id": id, "scene_tags": newTag}`. The frontend's existing `handleEvent → loadContext() → useEffect([ctx.session.scene_tags]) → setAmbientTrack(firstTag)` chain handles the rest automatically.

**Error handling:** Any AI call failure or DB error is logged and silently dropped (same pattern as all other auto goroutines).

### Frontend: Ambient Pause/Resume

**`ambient.ts` changes:**
- Add module-level `paused = false` state variable
- `pauseAmbient()`: sets `paused = true`, calls `currentTrack.audio.pause()`
- `resumeAmbient()`: sets `paused = false`, calls `currentTrack.audio.play()`
- `setAmbientTrack()`: after loading the new track, check `paused` — if true, load the audio element but do not call `fadeIn`/`play()`. The track is ready but silent until the player resumes.
- `setAmbientMuted()`: when unmuting, also respect `paused` (don't auto-play if the player has paused)

**`AudioControls.tsx` changes:**
- Add `ambientPaused` state (boolean, persisted to `localStorage` key `inkandbone_ambient_paused`)
- On mount, initialize `ambient.ts` pause state from localStorage
- Render a ⏸/▶ button next to the existing mute/volume controls
- Clicking toggles `paused` state, calls `pauseAmbient()` or `resumeAmbient()`
- Button is disabled when muted (ambient is already silent)

## Data Flow

```
GM responds → handleGMRespondStream → go autoUpdateSceneTags()
  → Claude: "which tag?" → {"tag":"dungeon"}
  → fetch current session.scene_tags
  → if tag unchanged: return (no-op)
  → db.UpdateSceneTags(sessionID, "dungeon")
  → bus.Publish(EventSessionUpdated{scene_tags:"dungeon"})
    → WebSocket → frontend handleEvent → loadContext()
      → ctx.session.scene_tags = "dungeon"
        → useEffect → setAmbientTrack("dungeon")
          → if paused: load audio, don't play
          → else: fadeOut old, fadeIn new
```

## Player Controls Interaction

- **Manual SceneTagPicker:** Player can still click tags to override. Their choice updates `scene_tags` in the DB the same way; the next GM response will let the AI overwrite it again.
- **Pause/Resume:** Independent of mute. Pausing stops the audio element. The AI can still switch tags while paused — the new track loads silently.
- **Mute:** Mutes all audio (dice, notifications, ambient). Pause/resume button disabled while muted.
- **Volume slider:** Controls ambient volume (already wired; no changes needed).

## Files Changed

- `internal/api/routes.go` — add `autoUpdateSceneTags` function, add `go s.autoUpdateSceneTags(...)` call in `handleGMRespondStream`
- `web/src/audio/ambient.ts` — add `pauseAmbient`, `resumeAmbient` exports; update `setAmbientTrack` and `setAmbientMuted` to respect paused state
- `web/src/AudioControls.tsx` — add pause/resume button and localStorage persistence

## Out of Scope

- Multi-tag selection by AI (single tag keeps audio deterministic)
- Separate ambient volume slider (existing master slider is sufficient)
- Player ability to permanently lock a tag against AI override
