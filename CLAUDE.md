# inkandbone

A private, local AI Game Master for 13 tabletop RPG systems. Single-binary Go server (HTTP, WebSocket, SQLite) with embedded React/TypeScript frontend. Streams GM responses via SSE. Automation goroutines handle NPC extraction, map generation, stat updates, recap regeneration, objective detection, and item tracking.

## Tech Stack

**Backend:**
- Go 1.22+
- SQLite (persisted to `~/.ttrpg`)
- HTTP + WebSocket + SSE
- AI clients: Claude Haiku (Anthropic), Ollama (local), HybridClient (Ollama GM + Claude automation), DualOllamaClient (two local models)

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
  db/                 - SQLite layer, migrations (20 files: initial + 19 features)
  ai/                 - AI client implementations (Claude Haiku, Ollama, Hybrid, DualOllama), system prompt injection, SSE streaming
  utils/              - Dice roller, validation, rulebook parsing
web/                  - React frontend (src/components, src/pages, src/hooks)
Makefile              - build, install, dev, test, clean
```

## Key Database Tables

- `campaigns` (ruleset reference, `chronicle_night` INTEGER DEFAULT 1 for VtM in-game night tracking)
- `rulesets` (name, schema as JSON, `gm_context` TEXT for per-system narrative guidance — 13 rulesets seeded with tone, vocabulary, honorifics, and mechanical language)
- `characters` (stats as JSON, portrait path, `currency_balance` INTEGER DEFAULT 0, `currency_label` TEXT DEFAULT 'Gold')
- `sessions` (title, date, summary, `tension_level` 1-10, `masquerade_integrity` INTEGER DEFAULT 10 for VtM Masquerade tracking)
- `messages` (full conversation history)
- `session_npcs` (named NPCs per session)
- `world_notes` (tagged lore: NPC, location, faction, item, other; `personality_json` for NPC traits)
- `maps` (uploaded or generated SVG)
- `map_pins` (coordinate + label + notes)
- `objectives` (active/completed/failed)
- `items` (character inventory)
- `combat_encounters` + `combatants` (turn tracking; VtM V5 combatants also have `damage_superficial`, `damage_aggravated`, `willpower_superficial`, `willpower_aggravated`, `hunger`)
- `dice_rolls` (expression + result breakdown)
- `oracle_tables` (action/theme table entries, 50 rows each, seeded by ruleset; VtM also has compulsion tables per clan: compulsion_brujah, compulsion_gangrel, compulsion_malkavian, compulsion_nosferatu, compulsion_toreador, compulsion_tremere, compulsion_ventrue)
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

1. **extractNPCs** — AI-powered NPC roster management: adds new named NPCs and removes NPCs confirmed dead/gone. Sends `{"add":[{name,note}],"remove":[ids]}` to AI; validates returned IDs against known roster before deleting.
2. **autoGenerateMap** — Detect new location names, generate SVG via Claude. Post-processes output via `extractSVG()` to ensure `xmlns="http://www.w3.org/2000/svg"` is present (required for browser rendering via `<img>`).
3. **autoUpdateCharacterStats** — Analyze story events, apply ruleset-based stat changes. When XP increases, fires `autoSuggestXPSpend` to generate advancement recommendations.
4. **autoUpdateRecap** — Every 4 GM messages, regenerate session journal.
5. **autoDetectObjectives** — Detect story goals, add to Objectives tab.
6. **autoExtractItems** — Parse items gained/lost, update inventory.
7. **checkAndExecuteRoll** — Before GM responds, enforce dice rolls if action requires them per ruleset.
8. **autoUpdateTension** — After GM response, check for crisis keywords (ambush, betrayal, catastrophe, danger, doom, enemy, escape, failure, fear, fight, flee, loss, peril, threat, trapped, wounded, etc.) or critical dice failures; auto-increment tension_level if found. Capped at 10.
9. **autoUpdateCurrency** — Analyze GM text for explicit currency transactions (number + currency word). Apply `MAX(0, balance + delta)` and persist. Broadcast `character_updated` with `currency_delta` for frontend undo toast. No-op when delta is 0 or parse fails.
10. **autoUpdateSceneTags** — Keyword-match GM text for environment terms; update `scene_tags` on the session to drive ambient audio. Zero AI cost (no Claude call).
11. **autoUpdateMasquerade** — VtM-only: Scan GM text for Masquerade breach keywords (camera, caught on film, mortal witnesses, police, etc.). Decrement `masquerade_integrity` by breach severity (1-3 points). No-op for non-VtM campaigns.
12. **autoUpdateChronicleNight** — VtM-only: Scan GM text for night-transition phrases (dusk falls, nightfall, darkness descends, fall of night, night has reclaimed, etc.). Increment `chronicle_night` on campaign when matched. The `[VtM MECHANICS]` world context block injects the current night number and explicit trigger phrases to ensure GM narration uses detectable language. No-op for non-VtM.
13. **autoSuggestXPSpend** — Fires when `xp` field increases within `autoUpdateCharacterStats`. Uses AI to generate 2-3 ranked advancement suggestions with `field`, `current_value`, `new_value`, `cost`, `reason`. Always reads `current_value` from actual `stats[field]` (never trusts AI-reported value) to prevent "new_value must be current value + 1" errors. Balanced bracket extraction prevents markdown fence contamination. Sends `xp_spend_suggestions` WebSocket event to frontend.

## Supported Rulesets

13 built-in systems seeded in `migrations/002_seed_rulesets.sql`:
- ironsworn, wrathglory, bitd, vtm, cthulhu, shadowrun, whfrp, sweoote, l5r, lotr, paranoia, dnd5e
- Custom rulesets can be added via JSON schema insert into `rulesets` table

Character creation options (race/class/archetype/faction dropdowns) are defined in `internal/ruleset/options.go`. Fully expanded for: dnd5e (race ×17, class ×14, background ×13, alignment ×9), wrath_glory (archetype ×23, faction ×13), shadowrun (metatype ×5, archetype ×10), theonering (culture ×8, calling ×8), blades (playbook, heritage, background, vice).

wrath_glory character schema includes all 19 skills (ws, bs, athletics, awareness, cunning, deception, fortitude, insight, intimidation, investigation, leadership, medicae, persuasion, pilot, psychic_mastery, scholar, stealth, survival, tech), talents (textarea), powers (textarea), corruption, speed, wealth tier, plus all core attributes, derived values, and Wrath/Glory/Ruin/XP. Migration: 016_wrath_glory_skills.sql.

### Database Migrations

30 total migrations in `internal/db/migrations/`:
- **001–017:** Core schema, features (rulesets, rulebooks, NPCs, objectives, items, combat, phases A–E, currency, W&G skills)
- **018:** `018_ruleset_gm_context.sql` — Adds `gm_context` TEXT column to rulesets table. Seeded with narrative guidance (tone, vocabulary, NPC conventions, mechanical language) for all 13 built-in rulesets. GM context is injected into the system prompt to ensure consistent ruleset flavor.
- **019:** `019_wrath_glory_honorifics.sql` — Appends honorific rules to W&G gm_context (Space Marines = "Brother", Sisters of Battle = "Sister"). Prevents misgendering and breaks immersion.
- **020:** `020_wrath_glory_prose_directives.sql` — Appends prose quality directives to W&G gm_context (length, second person, no purple prose, specificity, no repeated phrases, sentence variety, show-don't-tell).
- **021:** `021_wrath_glory_response_length.sql` — Fixes conflicting LENGTH directive; changes "at least 3 substantial paragraphs" to "exactly 4-5 paragraphs" to align with base system prompt.
- **022:** `022_wrath_glory_response_length_v2.sql` — Intermediate length adjustment (3-4 paragraphs); superseded by 023.
- **023:** `023_wrath_glory_response_length_v3.sql` — Final response length: 4-5 paragraphs. Matches base gmSystemPrompt FORMAT rule.
- **024:** `024_vtm_v5_schema.sql` — Rewrites VtM character schema to V5: hunger, blood_potency, bane_severity, humanity, stains, all 7 attribute pools, 40 skills, 11 Discipline fields, health/willpower tracks (max/superficial/aggravated), and free-text fields.
- **025:** `025_vtm_gm_context_v5.sql` — Rewrites VtM `gm_context` to V5 accuracy: correct V5 vocabulary (Hunger not blood pool), Hunger Die/Bestial Failure/Messy Critical narration, Rouse Check phrasing, Frenzy types, Masquerade breach severity table, all 7 clan Compulsion descriptions.
- **026:** `026_vtm_combat_damage.sql` — Adds V5 damage columns to `combatants`: `damage_superficial`, `damage_aggravated`, `willpower_superficial`, `willpower_aggravated`, `hunger`.
- **027:** `027_vtm_scene_tags.sql` — Adds `masquerade_integrity` INTEGER DEFAULT 10 to `sessions` for Masquerade breach tracking.
- **028:** `028_vtm_oracle_compulsion.sql` — Seeds VtM-specific Action and Theme oracle tables (50 rows each) and all 7 clan Compulsion tables (10 rows each: compulsion_brujah, compulsion_gangrel, compulsion_malkavian, compulsion_nosferatu, compulsion_toreador, compulsion_tremere, compulsion_ventrue).
- **029:** `029_vtm_xp.sql` — Adds `xp` field to VtM character schema for Beat/XP advancement tracking.
- **030:** `030_campaign_chronicle_night.sql` — Adds `chronicle_night` INTEGER DEFAULT 1 to `campaigns` for VtM in-game night tracking.

## AI Client Configuration

The binary in `cmd/ttrpg/main.go` detects and configures the AI client based on environment variables (checked in this order):

| Config | Env Vars | Behavior | Use Case |
|--------|----------|----------|----------|
| **Hybrid** | `OLLAMA_GM_MODEL` + `ANTHROPIC_API_KEY` | Ollama for GM narration (no API cost), Claude Haiku for automation (NPC extraction, maps, etc.) | Best of both: free local prose + reliable structured tasks |
| **Claude** | `ANTHROPIC_API_KEY` only | All operations via Claude Haiku | Simple, low-latency, low-cost production |
| **Dual Ollama** | `OLLAMA_GM_MODEL` + `OLLAMA_AI_MODEL` | Two local models: one for narrative, one for structured tasks | Full local, different models for different jobs |
| **Single Ollama** | `OLLAMA_MODEL` only | Same local model for both GM and automation | Simplest local setup, single model does all work |
| **Disabled** | None set | AI client is nil; all endpoints that call AI fail gracefully | Testing/dev without API costs |

**NewOllamaGMClient** (internal/ai/ollama.go) is tuned for prose quality:
- `num_ctx`: 16384 (full session history available)
- `temperature`: 0.85 (creative but not random)
- `repeat_penalty`: 1.15 (prevents loopy prose)
- `repeat_last_n`: 128
- `top_p`: 0.92, `top_k`: 60

**Qwen3 Think Mode:** Ollama clients can enable `/think` reasoning mode for models like Qwen3 that support it. The streaming path buffers and strips `<think>…</think>` blocks so only visible output reaches the client. Non-streaming path uses `stripThinkBlock()`.

**SVG Map Generation:** Always uses the structured AI client (Claude Haiku or fallback automation model) — Ollama GM narration is not used for maps because map generation requires precise SVG syntax. The `extractSVG()` function in routes.go post-processes AI output to ensure `xmlns="http://www.w3.org/2000/svg"` is present so browsers render SVG via `<img>` tags.

## Key Implementation Details

**System Prompt Injection:** Character name injected per turn. World context block includes:
- `[ACTIVE OBJECTIVES]` — All active quests/goals for the campaign
- `[NPC: Name]` personality cards — For every NPC world note with non-empty `personality_json`, the personality definition is injected
- `[RULEBOOK REFERENCES]` — Up to 5 rulebook chunks searched from player message keywords (with mechanic term expansion); injected before GM responds. Governed by a `RULEBOOK ADHERENCE` directive in the system prompt that binds Claude to apply referenced rules exactly.
- Ensures NPCs stay consistent, plot threads remain visible, and GM rulings follow the actual rulebook.

**GM Response Length:** Hard-enforced at 4-5 paragraphs via three mechanisms: (1) `gmSystemPrompt` FORMAT rule, (2) per-ruleset `gm_context` LENGTH directive (migrations 021-023), (3) `[REMINDER]` block appended as the final line of every system prompt. The FORMAT rule explicitly tells the model not to match the length of prior responses — necessary because the conversation history may contain longer responses from before the limit was tuned.

**Em-dash Strip:** All output em-dashes (`—`) programmatically stripped before display.

**Dice Roll Enforcement:** `checkAndExecuteRoll` analyzes player action before GM responds; rolls are pre-executed if ruleset requires them.

**SSE Streaming:** GM responses stream character-by-character via SSE to create live prose effect.

**WebSocket Hub:** Event bus broadcasts all state changes (character updates, NPCs, objectives, items, tension, relationships, oracle rolls) to connected clients in real time.

**Rulebook Chunks:** Supports PDF/text upload for rules lookup during play. Sources are labeled (Core Rulebook, Bestiary, Adventure, etc.) and stored independently; re-uploading the same source overwrites only that source's chunks.

**Whisper Mode:** Player messages marked private (🔒 lock icon) are excluded from GM context and session exports.

**NPC Personality JSON (Phase B):** World notes tagged as `npc` can store a `personality_json` field with any valid JSON object defining NPC traits. This is injected into Claude's context on every turn via the world context block.

**NPC Disambiguation:** `appendNPCDisambiguation` runs before every GM turn. Scans player message words against the session NPC roster using Levenshtein distance (threshold: 1 for short words, 2 for words 6+ chars). Injects `[NPC DISAMBIGUATION]` hint block into world context when a likely typo/misspelling is detected. Exact matches and substring matches skip disambiguation.

**Oracle System (Phase D):** 50-row Action and Theme tables seeded per ruleset. Roll endpoint accepts 1-50 roll value and returns matching oracle result. VtM also has 7 clan Compulsion tables (10 rows each). Custom rulesets can provide their own oracle tables.

**Tension Auto-Update (Phase D):** `autoUpdateTension` goroutine scans GM response text for crisis keywords; increments `tension_level` (max 10) on match. Also increments on critical dice failures. Manual override via PATCH endpoint.

**Relationships (Phase D):** Campaign-wide relationship tracking with from_name, to_name, relationship_type (neutral, ally, enemy, rival, mentor, etc.), and description. Used to drive roleplay and plot complications.

**Procedural Audio (Phase E):** Web Audio API synthesis for dice rolls, notifications, and combat start sounds. No audio files required. All sounds respect mute toggle and volume slider.

**Ambient Audio (Phase E):** Local MP3 loop manager with fade in/out. Loads from `/api/files/audio/{tag}.mp3`. Scene tags toggle ambient audio track selection. Supports 13 scene tags: tavern, dungeon, forest, city, ocean, cave, castle, rain, night, battle, market, temple, ruins. User can place custom MP3 files in `~/.ttrpg/audio/` directory.

**Audio Controls (Phase E):** AudioControls component in grimoire header with mute toggle (🔔/🔕) and volume slider (0-100). Settings persisted to localStorage.

**VtM V5 Chronicle Night:** `autoUpdateChronicleNight` goroutine matches a broad regex (`vtmNewNightRE`) of ~40 night-transition phrases against GM text (dusk falls, nightfall, fall of night, night has reclaimed, darkness descends, etc.). On match, increments `chronicle_night` on the campaign. The `[VtM MECHANICS]` world context block injects the current night number and a list of exact trigger phrases the GM must use when narrating a new night — two-layer reliability. Chronicle Night tracker in the UI is **display-only** (no +/− buttons); the AI GM is the sole source of truth. No-op for non-VtM campaigns.

**VtM V5 Masquerade Integrity:** `autoUpdateMasquerade` goroutine scans GM text for Masquerade breach keywords (camera, caught on film, mortal witness, police, goes viral, etc.). Decrements `masquerade_integrity` on the session based on breach severity. Displayed in the session header. No-op for non-VtM campaigns.

**VtM V5 Dice Display:** VtM pool rolls store success counts (not pip sums) in `dice_rolls.result`. `DiceHistoryPanel` detects VtM pool rolls by checking for the `(xN+yH)` suffix in the expression and displays "N successes" instead of a raw number.

**VtM V5 XP Suggestions:** `autoSuggestXPSpend` fires when `xp` increases within `autoUpdateCharacterStats`. Uses balanced bracket extraction (depth-counting, not `strings.LastIndex`) to find the JSON array in AI output — prevents markdown fence contamination. Always reads `current_value` from actual `stats[field]` (never trusts AI-reported value) to prevent off-by-one errors in the ADVANCE handler's `new_value == current + 1` check. Sends `xp_spend_suggestions` WebSocket event.

## Build & Deploy

```bash
make build   # React (Vite) → Go binary (embed dist/)
make install # Binary to ~/bin/ttrpg-bin
make dev     # Hot reload: air (Go) + Vite (React) concurrently
make test    # Run all tests
```

Wrapper script at `~/bin/ttrpg` passes AI configuration to the binary. See README.md **AI Configuration** for examples: Claude Haiku, Ollama single-model, dual-model, or hybrid modes.

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
