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
  db/                 - SQLite layer, migrations (8 files: initial + 7 features)
  ai/                 - Claude API client, system prompt injection, SSE streaming
  utils/              - Dice roller, validation, rulebook parsing
web/                  - React frontend (src/components, src/pages, src/hooks)
Makefile              - build, install, dev, test, clean
```

## Key Database Tables

- `campaigns` (ruleset reference)
- `characters` (stats as JSON, portrait path)
- `sessions` (title, date, summary)
- `messages` (full conversation history)
- `session_npcs` (named NPCs per session)
- `world_notes` (tagged lore: NPC, location, faction, item, other)
- `maps` (uploaded or generated SVG)
- `map_pins` (coordinate + label + notes)
- `objectives` (active/completed/failed)
- `items` (character inventory)
- `combat_encounters` + `combatants` (turn tracking)
- `dice_rolls` (expression + result breakdown)

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

**Write:**
- `POST /api/sessions/{id}/messages` — Send player action
- `POST /api/sessions/{id}/gm-respond-stream` — Stream GM response (SSE)
- `POST /api/sessions/{id}/dice-rolls` — Record dice roll
- `POST /api/campaigns/{id}/world-notes/draft` — AI-generate note
- `POST /api/campaigns/{id}/maps/generate` — AI-generate map
- `POST /api/campaigns/{id}/objectives` — Create objective
- `POST /api/characters/{id}/items` — Add item to inventory

**Patches:**
- `PATCH /api/campaigns/{id}` — Update campaign (name, description, active status)
- `PATCH /api/sessions/{id}` — Update session (title, summary)
- `PATCH /api/characters/{id}` — Update character sheet
- `PATCH /api/world-notes/{id}` — Update world note
- `PATCH /api/objectives/{id}` — Mark objective complete/failed
- `PATCH /api/items/{id}` — Update item (quantity, equipped)
- `PATCH /api/combatants/{id}` — Add/remove conditions (poisoned, paralyzed, etc.)

**WebSocket:**
- `GET /ws` — Upgrade to WebSocket for real-time dashboard updates

## Automation Goroutines

All fire after every GM response via `handleGMRespondStream`. Access goroutines in `/internal/api/routes.go`:

1. **autoExtractNPCs** (line ~700) — Parse proper names from GM text, add to session NPC roster.
2. **autoGenerateMap** (line ~953) — Detect new location names, generate SVG via Claude.
3. **autoUpdateCharacterStats** (line ~1048) — Analyze story events, apply ruleset-based stat changes.
4. **autoUpdateRecap** (line ~612) — Every 4 GM messages, regenerate session journal.
5. **autoDetectObjectives** (line ~1688) — Detect story goals, add to Objectives tab.
6. **autoExtractItems** (line ~1911) — Parse items gained/lost, update inventory.
7. **checkAndExecuteRoll** (line ~1160) — Before GM responds, enforce dice rolls if action requires them per ruleset.

## Supported Rulesets

13 built-in systems seeded in `migrations/002_seed_rulesets.sql`:
- ironsworn, wrathglory, bitd, vtm, cthulhu, shadowrun, whfrp, sweoote, l5r, lotr, paranoia, dnd5e
- Custom rulesets can be added via JSON schema insert into `rulesets` table

## Key Implementation Details

**System Prompt Injection:** Character name injected per turn so Claude uses correct pronouns/names.

**Em-dash Strip:** All output em-dashes (`—`) programmatically stripped before display.

**Dice Roll Enforcement:** `checkAndExecuteRoll` analyzes player action before GM responds; rolls are pre-executed if ruleset requires them.

**SSE Streaming:** GM responses stream character-by-character via SSE to create live prose effect.

**WebSocket Hub:** Event bus broadcasts all state changes (character updates, NPCs, objectives, items, etc.) to connected clients in real time.

**Rulebook Chunks:** `migrations/003_rulebook_chunks.sql` supports PDF/text upload for rules lookup during play.

**Whisper Mode:** Player messages marked private (🔒 lock icon) are excluded from GM context and session exports.

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
