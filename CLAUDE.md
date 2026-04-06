# inkandbone

A private, local AI Game Master for 13 tabletop RPG systems. Single-binary Go server (HTTP, WebSocket, SQLite) with embedded React/TypeScript frontend. Streams GM responses via SSE. Automation goroutines handle NPC extraction, map generation, stat updates, recap regeneration, objective detection, and item tracking.

## Tech Stack

**Backend:**
- Go 1.22+
- SQLite (persisted to `~/.ttrpg`)
- HTTP + WebSocket + SSE
- Claude Haiku (MCP tools for AI GM)

**Frontend:**
- React 18 + TypeScript
- Vite (development and bundled into binary)
- WebSocket client for live updates
- "Worn Grimoire" dark theme (parchment + gold) + light theme toggle

**Tools & Libraries:**
- Dice roller (expression parsing: `1d20`, `2d6+3`, etc.)
- SVG map generator (AI-generated location maps)
- Ruleset schema validator (character sheet generation)
- Rulebook indexing (PDF/text upload for rules lookup)

## Project Structure

```
cmd/ttrpg/            - Binary entrypoint
internal/
  api/                - HTTP handlers, WebSocket hub, event bus, automation goroutines
  db/                 - SQLite layer, migrations (15 files: initial + 14 features)
  ai/                 - Claude API client, system prompt injection, SSE streaming
  utils/              - Dice roller, validation, rulebook parsing
web/                  - React frontend (src/components, src/pages, src/hooks)
Makefile              - build, install, dev, test, clean
```

## Key Database Tables

- `campaigns` (ruleset reference)
- `characters` (stats as JSON, portrait path, `currency_balance` INTEGER DEFAULT 0, `currency_label` TEXT DEFAULT 'Gold')
- `sessions` (title, date, summary, `tension_level` 1-10)
- `messages` (full conversation history)
- `session_npcs` (named NPCs per session)
- `world_notes` (tagged lore: NPC, location, faction, item, other; `personality_json` for NPC traits)
- `maps` (uploaded or generated SVG)
- `map_pins` (coordinate + label + notes)
- `objectives` (active/completed/failed)
- `items` (character inventory)
- `combat_encounters` + `combatants` (turn tracking)
- `dice_rolls` (expression + result breakdown)
- `oracle_tables` (action/theme table entries, 50 rows each, seeded by ruleset)
- `relationships` (from_name, to_name, relationship_type, description, campaign_id)
- `scene_tags` (session scene tags: tavern, dungeon, forest, city, ocean, cave, castle, rain, night, battle, market, temple, ruins)

## Routes (HTTP + WebSocket)

**Read:**
- `GET /api/campaigns` — List campaigns
- `GET /api/campaigns/{id}/characters` — Characters for campaign
- `GET /api/campaigns/{id}/sessions` — Sessions for campaign
- `GET /api/sessions/{id}/messages` — Message history
- `GET /api/sessions/{id}/timeline` — Session timeline (auto-recap)
- `GET /api/campaigns/{id}/world-notes` — World notes with tags
- `GET /api/sessions/{id}/dice-rolls` — Dice roll history
- `GET /api/maps/{id}/pins` — Pins on a map
- `GET /api/campaigns/{id}/maps` — Campaign maps
- `GET /api/sessions/{id}/npcs` — NPC roster
- `GET /api/campaigns/{id}/objectives` — Quest tracker
- `GET /api/characters/{id}/items` — Inventory
- `GET /api/sessions/{id}/tension` — Current tension level
- `GET /api/campaigns/{id}/relationships` — All relationships for campaign

**Write:**
- `POST /api/sessions/{id}/messages` — Send player action
- `POST /api/sessions/{id}/gm-respond-stream` — Stream GM response (SSE)
- `POST /api/sessions/{id}/dice-rolls` — Record dice roll
- `POST /api/campaigns/{id}/world-notes/draft` — AI-generate note
- `POST /api/campaigns/{id}/maps/generate` — AI-generate map
- `POST /api/campaigns/{id}/objectives` — Create objective
- `POST /api/characters/{id}/items` — Add item to inventory
- `POST /api/sessions/{id}/improvise` — Generate improvised scene/complication (Phase C)
- `POST /api/campaigns/{id}/pre-session-brief` — Generate GM prep brief (Phase C)
- `POST /api/sessions/{id}/detect-threads` — Identify unresolved narrative threads (Phase C)
- `POST /api/campaigns/{id}/ask` — Ask freeform question about campaign (Phase C)
- `POST /api/oracle/roll` — Roll oracle table (Phase D)
- `POST /api/campaigns/{id}/relationships` — Create relationship (Phase D)

**Patches:**
- `PATCH /api/campaigns/{id}` — Update campaign (name, description, active status)
- `PATCH /api/sessions/{id}` — Update session (title, summary, scene tags)
- `PATCH /api/characters/{id}` — Update character sheet (includes `currency_balance` and `currency_label`)
- `PATCH /api/world-notes/{id}` — Update world note
- `PATCH /api/world-notes/{id}/personality` — Set NPC personality JSON (Phase B)
- `PATCH /api/objectives/{id}` — Mark objective complete/failed
- `PATCH /api/items/{id}` — Update item (quantity, equipped)
- `PATCH /api/combatants/{id}` — Add/remove conditions (poisoned, paralyzed, etc.)
- `PATCH /api/sessions/{id}/tension` — Set tension level (Phase D)
- `PATCH /api/relationships/{id}` — Update relationship type/description (Phase D)

**Delete:**
- `DELETE /api/relationships/{id}` — Remove relationship (Phase D)

**WebSocket:**
- `GET /ws` — Upgrade to WebSocket for real-time dashboard updates

## Automation Goroutines

All fire after every GM response via `handleGMRespondStream`. Goroutines organized in separate files:

1. **autoExtractNPCs** — Parse proper names from GM text, add to session NPC roster.
2. **autoGenerateMap** — Detect new location names, generate SVG via Claude.
3. **autoUpdateCharacterStats** — Analyze story events, apply ruleset-based stat changes.
4. **autoUpdateRecap** — Every 4 GM messages, regenerate session journal.
5. **autoDetectObjectives** — Detect story goals, add to Objectives tab.
6. **autoExtractItems** — Parse items gained/lost, update inventory.
7. **checkAndExecuteRoll** — Before GM responds, enforce dice rolls if action requires them per ruleset.
8. **autoUpdateTension** — After GM response, check for crisis keywords (ambush, betrayal, catastrophe, danger, doom, enemy, escape, failure, fear, fight, flee, loss, peril, threat, trapped, wounded, etc.) or critical dice failures; auto-increment tension_level if found. Capped at 10.
9. **autoUpdateCurrency** — Analyze GM text for explicit currency transactions (number + currency word). Apply `MAX(0, balance + delta)` and persist. Broadcast `character_updated` with `currency_delta` for frontend undo toast. No-op when delta is 0 or parse fails.
10. **autoUpdateSceneTags** — Keyword-match GM text for environment terms; update `scene_tags` on the session to drive ambient audio. Zero AI cost (no Claude call).

## Supported Rulesets

13 built-in systems seeded in `migrations/002_seed_rulesets.sql`:
- ironsworn, wrathglory, bitd, vtm, cthulhu, shadowrun, whfrp, sweoote, l5r, lotr, paranoia, dnd5e
- Custom rulesets can be added via JSON schema insert into `rulesets` table

Character creation options (race/class/archetype/faction dropdowns) are defined in `internal/ruleset/options.go`. Fully expanded for: dnd5e (race ×17, class ×14, background ×13, alignment ×9), wrath_glory (archetype ×23, faction ×13), shadowrun (metatype ×5, archetype ×10), theonering (culture ×8, calling ×8), blades (playbook, heritage, background, vice).

wrath_glory character schema includes all 19 skills (ws, bs, athletics, awareness, cunning, deception, fortitude, insight, intimidation, investigation, leadership, medicae, persuasion, pilot, psychic_mastery, scholar, stealth, survival, tech), talents (textarea), powers (textarea), corruption, speed, wealth tier, plus all core attributes, derived values, and Wrath/Glory/Ruin/XP. Migration: 016_wrath_glory_skills.sql.

## Key Implementation Details

**System Prompt Injection:** Character name injected per turn. World context block includes:
- `[ACTIVE OBJECTIVES]` — All active quests/goals for the campaign
- `[NPC: Name]` personality cards — For every NPC world note with non-empty `personality_json`, the personality definition is injected
- `[RULEBOOK REFERENCES]` — Up to 5 rulebook chunks searched from player message keywords (with mechanic term expansion); injected before GM responds. Governed by a `RULEBOOK ADHERENCE` directive in the system prompt that binds Claude to apply referenced rules exactly.
- Ensures NPCs stay consistent, plot threads remain visible, and GM rulings follow the actual rulebook.

**Em-dash Strip:** All output em-dashes (`—`) programmatically stripped before display.

**Dice Roll Enforcement:** `checkAndExecuteRoll` analyzes player action before GM responds; rolls are pre-executed if ruleset requires them.

**SSE Streaming:** GM responses stream character-by-character via SSE to create live prose effect.

**WebSocket Hub:** Event bus broadcasts all state changes (character updates, NPCs, objectives, items, tension, relationships, oracle rolls) to connected clients in real time.

**Rulebook Chunks:** Supports PDF/text upload for rules lookup during play. Sources are labeled (Core Rulebook, Bestiary, Adventure, etc.) and stored independently; re-uploading the same source overwrites only that source's chunks.

**Whisper Mode:** Player messages marked private (🔒 lock icon) are excluded from GM context and session exports.

**NPC Personality JSON (Phase B):** World notes tagged as `npc` can store a `personality_json` field with any valid JSON object defining NPC traits. This is injected into Claude's context on every turn via the world context block.

**Oracle System (Phase D):** 50-row Action and Theme tables seeded per ruleset. Roll endpoint accepts 1-50 roll value and returns matching oracle result. Custom rulesets can provide their own oracle tables.

**Tension Auto-Update (Phase D):** `autoUpdateTension` goroutine scans GM response text for crisis keywords; increments `tension_level` (max 10) on match. Also increments on critical dice failures. Manual override via PATCH endpoint.

**Relationships (Phase D):** Campaign-wide relationship tracking with from_name, to_name, relationship_type (neutral, ally, enemy, rival, mentor, etc.), and description. Used to drive roleplay and plot complications.

**Procedural Audio (Phase E):** Web Audio API synthesis for dice rolls, notifications, and combat start sounds. No audio files required. All sounds respect mute toggle and volume slider.

**Ambient Audio (Phase E):** Local MP3 loop manager with fade in/out. Loads from `/api/files/audio/{tag}.mp3`. Scene tags toggle ambient audio track selection. Supports 13 scene tags: tavern, dungeon, forest, city, ocean, cave, castle, rain, night, battle, market, temple, ruins. User can place custom MP3 files in `~/.ttrpg/audio/` directory.

**Audio Controls (Phase E):** AudioControls component in grimoire header with mute toggle (🔔/🔕) and volume slider (0-100). Settings persisted to localStorage.

## Build & Deploy

```bash
make build   # React (Vite) → Go binary (embed dist/)
make install # Binary to ~/bin/ttrpg-bin
make dev     # Hot reload: air (Go) + Vite (React) concurrently
make test    # Run all tests
```

Wrapper script at `~/bin/ttrpg` passes `ANTHROPIC_API_KEY` to the binary.

## Notes for Contributors

1. **All routes tested?** Check that both request/response and database mutations are tested.
2. **Automation goroutine?** Ensure it gracefully handles nil pointers and logs errors.
3. **New feature = new migration.** Don't modify existing migrations; add incremental .sql files.
4. **Frontend updates?** Hot reload via Vite during `make dev`; rebuild with `make build`.
5. **Ruleset additions?** Add JSON schema to `rulesets` table; test with character creation.
6. **Documentation?** Update README.md and inline comments before committing.

## Related Files

- README.md — User-facing feature guide and tutorials
- docs/games/ — Game-specific lore and rules (DO NOT EDIT—content author owns)
- docs/superpowers/ — Internal planning docs (DO NOT EDIT—for MCP/automation)
- internal/db/migrations/ — Database schema (8 files, one per feature)
- web/src/components/ — React UI panels
- Makefile — Build, install, dev commands
