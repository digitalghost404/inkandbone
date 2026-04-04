# ink & bone

A local TTRPG companion app where Claude Code acts as your Game Master. One binary runs everything: an MCP stdio server (Claude's GM interface), an HTTP/WebSocket API, and an embedded React dashboard — all backed by a local SQLite database. No cloud accounts, no subscriptions, no external services beyond an optional Anthropic API key for AI features.

```
┌──────────────┐   MCP stdio   ┌─────────────────┐   WS / HTTP   ┌────────────────┐
│  Claude Code │ ◄──────────► │  ./ttrpg binary  │ ────────────► │ localhost:7432 │
│  (your GM)   │               │  SQLite + API    │               │  React UI      │
└──────────────┘               └─────────────────┘               └────────────────┘
```

---

## Table of Contents

1. [Requirements](#requirements)
2. [Installation](#installation)
3. [Configuration](#configuration)
4. [First Run](#first-run)
5. [User Guide](#user-guide)
   - [Core Concepts](#core-concepts)
   - [Starting a Campaign](#starting-a-campaign)
   - [Managing Characters](#managing-characters)
   - [Playing a Session](#playing-a-session)
   - [Combat Tracking](#combat-tracking)
   - [World Notes](#world-notes)
   - [Campaign Maps](#campaign-maps)
   - [Dice Rolling](#dice-rolling)
   - [Session Recap (AI)](#session-recap-ai)
   - [Rulebook Ingestion](#rulebook-ingestion)
   - [Character Sheet Panel](#character-sheet-panel)
6. [Supported Rulesets](#supported-rulesets)
7. [Adding a Custom Ruleset](#adding-a-custom-ruleset)
8. [Dashboard Reference](#dashboard-reference)
9. [MCP Tool Reference](#mcp-tool-reference)
10. [HTTP API Reference](#http-api-reference)
11. [WebSocket Events](#websocket-events)
12. [File Storage](#file-storage)
13. [Development](#development)

---

## Requirements

- Go 1.22+
- Node 18+ with npm
- [air](https://github.com/air-verse/air) (for live reload during development only)

```bash
go install github.com/air-verse/air@latest
```

An Anthropic API key is **optional**. Without one the app works fully — you just won't have the AI-powered recap and world-note draft features.

---

## Installation

```bash
# Clone and build
git clone https://github.com/digitalghost404/inkandbone
cd inkandbone
make build          # compiles frontend + Go binary → ./ttrpg

# Install to ~/bin (adds to PATH if ~/bin is in your PATH)
make install        # runs make build, then copies binary to ~/bin/ttrpg
```

The resulting binary is fully self-contained. The React frontend is embedded at compile time — no separate web server needed.

---

## Configuration

The app has one environment variable:

| Variable | Required | Description |
|---|---|---|
| `ANTHROPIC_API_KEY` | No | Enables AI features: session recap generation and AI world-note drafting. |

Everything else is hardcoded:

| Setting | Value |
|---|---|
| HTTP port | `:7432` |
| Database path | `~/.ttrpg/ttrpg.db` (auto-created on first run) |
| Data directory | `~/.ttrpg/` (portraits, maps stored here) |

---

## First Run

### 1. Register the MCP server

Add ink & bone to Claude Code's MCP server list. Edit `~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "inkandbone": {
      "command": "/path/to/ttrpg",
      "env": {
        "ANTHROPIC_API_KEY": "sk-ant-..."
      }
    }
  }
}
```

Replace `/path/to/ttrpg` with the actual binary path (e.g. `~/bin/ttrpg` after `make install`). The `env` block is optional — omit it if you don't have an API key or prefer to export the variable from your shell.

### 2. Start the server

```bash
ttrpg
```

The binary starts two things simultaneously:
- HTTP server on `localhost:7432` (the dashboard and API)
- MCP stdio server (Claude Code connects to this automatically)

Open `localhost:7432` in a browser. You'll see the dashboard. Open Claude Code in a terminal or IDE pane alongside it.

### 3. Create your first campaign

In Claude Code, tell Claude to set things up:

```
Create a D&D 5e campaign called "The Iron Vale Chronicles" and create a character named Sable.
```

Claude will call the `create_campaign` and `create_character` MCP tools. The dashboard refreshes live.

---

## User Guide

### Core Concepts

ink & bone has five main entities that build on each other:

```
Ruleset ──► Campaign ──► Characters
                │
                └──► Sessions ──► Messages, Dice Rolls, Combat
                │
                └──► World Notes
                │
                └──► Maps ──► Pins
```

- **Ruleset** — defines which game system you're playing and what fields a character sheet has (D&D 5e, Ironsworn, etc.)
- **Campaign** — the overarching story. Belongs to one ruleset. Has characters, sessions, world notes, and maps.
- **Character** — a player character in a campaign. Has a freeform JSON data blob (the sheet) and an optional portrait image.
- **Session** — one play session. Has a title, date, narrative messages, dice rolls, and combat encounters.
- **Active context** — one campaign, character, and session can be "active" at a time. MCP tools default to the active context when no IDs are specified.

### Starting a Campaign

Tell Claude which ruleset and name you want:

```
Create a Cyberpunk Red campaign called "Night City Shadows".
```

Behind the scenes Claude calls `create_campaign`. The campaign is immediately set as active. You can have multiple campaigns and switch between them:

```
List my campaigns.
Set campaign 2 as active.
```

### Managing Characters

**Create a character:**

```
Create a character named Vesper for the active campaign.
```

Claude calls `create_character`. The new character is activated automatically.

**Update character fields:**

```
Set Vesper's hp to 35, her level to 4, and add "Cloak of Elvenkind" to her inventory.
```

Claude translates this into `update_character` and `add_item` calls. The character sheet panel in the dashboard updates in real time via WebSocket.

**Upload a portrait:**

Open the dashboard, navigate to the character sheet panel, and click the portrait area. Select any `.jpg`, `.jpeg`, `.png`, `.gif`, or `.webp` file up to 5 MB. The portrait appears immediately in the header and sheet panel.

**Switch active character:**

```
List characters. Set character 3 as active.
```

### Playing a Session

**Start a session:**

```
Start a session called "The Ambush at Millhaven Bridge" for today.
```

Claude calls `start_session`. The session is set active and all subsequent dice rolls, messages, and combat will be logged to it.

**Narrate events:**

Any tool call Claude makes can include a `narrative` string. When Claude narrates, that text is saved as an assistant message in the session log and appears in the dashboard's session log panel.

```
Roll for initiative and describe the ambush scene.
```

**End a session:**

```
End the session. Write a one-paragraph summary of what happened.
```

Claude calls `end_session` with the summary you provided. The summary is saved and appears in the journal panel.

**View timeline:**

The session timeline panel in the dashboard shows all events in chronological order: narrative messages, dice rolls, world note creations, and combat events.

### Combat Tracking

**Start a combat encounter:**

```
Start combat called "Ambush". The party faces two bandits (HP 11 each) and a bandit captain (HP 65).
```

Claude calls `start_combat` with a name and a list of combatants. Each combatant needs a name, initiative, max HP, and whether they're a player character. The combat panel appears in the dashboard.

**Update combatant HP and conditions:**

```
The bandit captain takes 18 damage. Vesper is now poisoned.
```

Claude calls `update_combatant` for each affected combatant. The HP bars in the dashboard update in real time. HP bars turn yellow below 50% and red below 25%.

**Conditions:**

Conditions are free-form strings. Common values: `"poisoned"`, `"prone"`, `"unconscious"`, `"stunned"`, `"blinded"`. They appear as badges on combatant cards.

**End combat:**

```
The party defeats the bandits. End combat.
```

### World Notes

World notes are your campaign wiki. Every NPC, location, faction, item, and lore entry lives here.

**Create a note via Claude:**

```
Create a world note for Mira Ashvale, a half-elf fence who operates out of the Tattered Lantern inn. Category: NPC.
```

Claude calls `create_world_note`.

**Draft a note with AI:**

In the dashboard's World Notes panel, click "Draft with AI", enter a hint like `"mysterious masked courier who delivers jobs to the party"`, and click Generate. The AI drafts a full note with title and content for you to review and save.

**Search notes:**

In the dashboard, type in the search box to filter notes by title and content. Click any tag pill to filter by that tag.

Via Claude:

```
Search world notes for "Mira".
Search world notes in the location category.
```

**Update a note:**

```
Update the Mira Ashvale note — she was revealed to be an agent of the Obsidian Court.
```

Claude calls `update_world_note` with the new content.

### Campaign Maps

**Upload a map:**

In the dashboard's Map Panel, click "Upload Map" and select an image file (`.jpg`, `.jpeg`, `.png`, `.gif`, `.webp`, up to 10 MB). The map name defaults to the filename without extension.

**Add pins via Claude:**

Map pins use fractional coordinates where `(0.0, 0.0)` is the top-left corner and `(1.0, 1.0)` is the bottom-right. Tell Claude where to pin things:

```
Add a pin on the map at roughly the center-left (0.2, 0.5) labeled "Tattered Lantern" with the note "Mira's base of operations". Use color #e67e22.
```

Claude calls `add_map_pin`. The pin appears on the map immediately.

**Interact with pins:**

Click any pin on the map to open a popup showing its label and note text.

### Dice Rolling

**Via Claude:**

```
Roll 2d6+3 for the attack.
Roll a d20 for perception.
Roll 4d6 drop lowest for strength.
```

Claude calls `roll_dice` with the expression. Results appear in the dice history panel with per-die breakdowns. A narrative can be attached: `"Vesper reaches out into the darkness — roll perception"`.

**Dice expression syntax:**

| Expression | Meaning |
|---|---|
| `d20` | Roll one d20 |
| `2d6` | Roll two d6, sum them |
| `1d8+3` | Roll one d8, add 3 |
| `2d6-1` | Roll two d6, subtract 1 |
| `d100` | Roll percentile dice |

Dice rolls are only recorded when there is an active session.

### Session Recap (AI)

At the end of a session, or any time during play:

```
Generate a session recap.
```

Claude calls `generate_session_recap`. The tool reads all narrative messages and dice rolls from the active session, sends them to the AI, and saves the generated summary. The journal panel updates automatically.

You can also click "Generate recap" directly in the journal panel in the dashboard.

Requires `ANTHROPIC_API_KEY`. Session transcripts over 32,000 characters are rejected.

### Rulebook Ingestion

You can upload rulebook text so Claude can search it during play.

**Upload plain text:**

```bash
curl -X POST http://localhost:7432/api/rulesets/1/rulebook \
  -H "Content-Type: text/plain" \
  --data-binary @my-rulebook.txt
```

The text is split into chunks at `#` headings. Each heading and its following content become one chunk. Text without any headings becomes a single chunk.

**Upload a PDF:**

```bash
curl -X POST http://localhost:7432/api/rulesets/1/rulebook \
  -F "rulebook=@my-rulebook.pdf"
```

The PDF is validated and text is extracted from content streams. Re-uploading replaces all previous chunks for that ruleset.

**Response:**

```json
{ "chunks_created": 42 }
```

**Search during play:**

```
Look up the grappling rules.
```

Claude calls `search_rulebook` with the query. Up to 3 matching chunks are returned and used to answer your question.

To find your ruleset ID, run `GET /api/rulesets/{id}` or check the seeded rulesets (IDs 1–5 for the built-in systems).

### Character Sheet Panel

The character sheet panel renders based on the active ruleset's `schema_json`. Fields are editable directly in the browser — changes are debounced by 500 ms and auto-saved to the server.

For built-in rulesets, fields are listed on the [Supported Rulesets](#supported-rulesets) page. For custom rulesets using the extended schema format, each field can be `text`, `number`, or `textarea`.

Portrait images can be updated from the sheet panel directly — click the portrait area to upload a new image.

---

## Supported Rulesets

Five rulesets are seeded into the database on first run:

### D&D 5e (`dnd5e`, version `5e`)

Fields: `race`, `class`, `level`, `hp`, `ac`, `str`, `dex`, `con`, `int`, `wis`, `cha`, `proficiency_bonus`, `skills`, `inventory`, `spells`, `features`

### Ironsworn (`ironsworn`, version `1.0`)

Fields: `edge`, `heart`, `iron`, `shadow`, `wits`, `health`, `spirit`, `supply`, `momentum`, `vows`, `bonds`, `assets`, `notes`

### Vampire: the Masquerade (`vtm`, version `V20`)

Fields: `clan`, `generation`, `humanity`, `blood_pool`, `willpower`, `attributes`, `abilities`, `disciplines`, `virtues`, `backgrounds`, `notes`

### Call of Cthulhu (`coc`, version `7e`)

Fields: `occupation`, `age`, `hp`, `sanity`, `luck`, `mp`, `str`, `con`, `siz`, `dex`, `app`, `int`, `pow`, `edu`, `skills`, `inventory`, `notes`

### Cyberpunk Red (`cyberpunk`, version `Red`)

Fields: `role`, `int`, `ref`, `cool`, `tech`, `lk`, `att`, `ma`, `emp`, `body`, `humanity`, `eurodollars`, `skills`, `cyberware`, `gear`, `notes`

---

## Adding a Custom Ruleset

Custom rulesets are not yet creatable through the UI or MCP tools. To add one, insert a row directly into the database:

```bash
sqlite3 ~/.ttrpg/ttrpg.db
```

```sql
INSERT INTO rulesets (name, schema_json, version) VALUES (
  'shadowrun',
  '[
    {"key":"metatype","label":"Metatype","type":"text"},
    {"key":"essence","label":"Essence","type":"number"},
    {"key":"nuyen","label":"Nuyen","type":"number"},
    {"key":"skills","label":"Skills","type":"textarea"},
    {"key":"cyberware","label":"Cyberware","type":"textarea"},
    {"key":"gear","label":"Gear","type":"textarea"},
    {"key":"contacts","label":"Contacts","type":"textarea"},
    {"key":"notes","label":"Notes","type":"textarea"}
  ]',
  '6e'
);
```

The `schema_json` field accepts an array of field descriptors:

```json
[
  { "key": "field_key", "label": "Display Label", "type": "text" },
  { "key": "hp", "label": "Hit Points", "type": "number" },
  { "key": "backstory", "label": "Backstory", "type": "textarea" }
]
```

| Field | Values | Effect |
|---|---|---|
| `key` | any string | Used as the JSON key in `data_json`; must be unique within the schema |
| `label` | any string | Displayed in the character sheet panel |
| `type` | `text`, `number`, `textarea` | Renders as a text input, number input, or multi-line textarea |

After inserting, create a campaign that uses the new ruleset name:

```
Create a Shadowrun campaign called "Emerald City Runs".
```

---

## Dashboard Reference

The dashboard at `localhost:7432` is a live-updating React SPA. It connects over WebSocket and re-renders on every server-side event without a full page reload.

### Header / State Bar

Shows the active campaign name, the active character's portrait thumbnail and name, and the active session title. Updates live when context changes.

### Session Log

Lists all narrative messages for the active session in order, labelled by role (`user` for player actions, `assistant` for GM narration).

### Session Timeline

A unified chronological view of everything that happened in the session: narrative messages, dice rolls, world note creations, and combat events. New entries animate in with a brief highlight.

### Combat Panel

Visible only during an active combat encounter. Shows a card for each combatant with:
- Name and initiative
- HP bar (green above 50%, yellow above 25%, red at 25% and below) and HP fraction
- Condition badges (poisoned, prone, etc.)
- The first combatant in the list is marked as having the active turn

### Dice History Panel

Shows all dice rolls for the active session with the expression, total result, and individual die values.

### Journal Panel

Editable textarea for the session summary. Saves automatically on blur. Includes a "Generate recap" button when AI is enabled.

### World Notes Panel

Displays all world notes for the active campaign. Features:
- Text search filtering by title and content
- Tag pill filtering (click any tag to filter by it)
- "Draft with AI" button (requires `ANTHROPIC_API_KEY`)
- Notes grouped by their most recent state

### Map Panel

Shows the first map uploaded for the active campaign. Features:
- Pins rendered as clickable markers positioned at fractional coordinates
- Click a pin to view its label and note text

### Character Sheet Panel

Renders the active character's fields based on the ruleset schema. Features:
- All fields editable inline with 500 ms auto-save debounce
- Portrait upload via click on the portrait area

---

## MCP Tool Reference

All tools are available to Claude Code once the MCP server is registered. Tools marked **AI required** need `ANTHROPIC_API_KEY`.

### Context

| Tool | Parameters | Returns |
|---|---|---|
| `get_context` | — | Full game state snapshot: active campaign, character, session, last 20 messages, active combat |

### Campaigns & Sessions

| Tool | Required | Optional | Returns |
|---|---|---|---|
| `create_campaign` | `ruleset` (string), `name` (string) | `description` | Campaign created, activated |
| `list_campaigns` | — | — | JSON array of campaigns |
| `set_active` | At least one of `campaign_id`, `session_id`, `character_id` | — | Confirmation |
| `start_session` | `title` (string), `date` (YYYY-MM-DD) | `narrative` | Session created, activated |
| `end_session` | `summary` (string) | `narrative` | Session closed |
| `list_sessions` | — | `campaign_id` | JSON array of sessions |

### Characters

| Tool | Required | Optional | Returns |
|---|---|---|---|
| `create_character` | `name` (string) | `campaign_id` | Character created, activated |
| `list_characters` | — | `campaign_id` | JSON array of characters |
| `get_character_sheet` | — | `character_id` | Full character JSON |
| `update_character` | `updates` (JSON object as string, e.g. `{"hp":15}`) | `character_id`, `narrative` | Confirmation |
| `add_item` | `item_name` (string) | `character_id`, `narrative` | Item appended to inventory |
| `remove_item` | `item_name` (string) | `character_id`, `narrative` | Item removed from inventory |

The `updates` parameter to `update_character` is a JSON object string with any keys from the ruleset's schema. Only the specified keys are updated; all other field values are preserved.

### Combat

| Tool | Required | Optional | Returns |
|---|---|---|---|
| `start_combat` | `name` (string), `combatants` (JSON array string) | `narrative` | Encounter created |
| `update_combatant` | `combatant_id` (number), `hp_current` (number) | `conditions` (JSON array string), `narrative` | Combatant updated |
| `end_combat` | — | `narrative` | Encounter closed |

**Combatants format:**

```json
[
  {"name": "Sable",   "initiative": 18, "hp_max": 52, "is_player": true},
  {"name": "Bandit",  "initiative": 12, "hp_max": 11, "is_player": false},
  {"name": "Captain", "initiative": 15, "hp_max": 65, "is_player": false}
]
```

**Conditions format:**

```json
["poisoned", "prone"]
```

### World Notes

| Tool | Required | Optional | Returns |
|---|---|---|---|
| `create_world_note` | `title`, `content`, `category` (npc/location/faction/item) | `narrative` | Note created |
| `update_world_note` | `note_id` (number), `title`, `content` | `tags` (JSON array string), `narrative` | Note updated |
| `search_world_notes` | — | `query` (text), `category` | Matching notes |

### Dice

| Tool | Required | Optional | Returns |
|---|---|---|---|
| `roll_dice` | `expression` (string, e.g. `"2d6+3"`) | `narrative` | Result string with total and breakdown |

Requires an active session.

### Maps

| Tool | Required | Optional | Returns |
|---|---|---|---|
| `add_map_pin` | `map_id` (number), `x` (float), `y` (float), `label` (string) | `note`, `color` (hex) | Pin created |

`x` and `y` are fractional coordinates from 0.0 (top-left) to 1.0 (bottom-right).

### AI Tools

| Tool | Required | Optional | Notes |
|---|---|---|---|
| `generate_session_recap` | — | `session_id` | **AI required.** Reads session messages and dice rolls, generates summary, saves it. |
| `search_rulebook` | `query` (string) | `ruleset_id` | Returns up to 3 matching rulebook chunks by heading or content. |

---

## HTTP API Reference

The API is available at `http://localhost:7432`. All JSON request bodies use `Content-Type: application/json`.

### Health

```
GET /api/health
→ { "status": "ok", "ai_enabled": true }
```

### Campaigns

```
GET /api/campaigns
→ []Campaign

GET /api/campaigns/{id}/characters
→ []Character

GET /api/campaigns/{id}/sessions
→ []Session

GET /api/campaigns/{id}/world-notes?q=text&category=npc&tag=mytag
→ []WorldNote

GET /api/campaigns/{id}/maps
→ []CampaignMap

POST /api/campaigns/{id}/maps
Content-Type: multipart/form-data
Fields: image (file), name (string)
→ 201 CampaignMap

POST /api/campaigns/{id}/world-notes/draft       (AI required)
Content-Type: application/json
Body: { "hint": "mysterious fence in the docks district" }
→ 201 { "id": N, "title": "...", "content": "..." }
```

### Sessions

```
GET /api/sessions/{id}/messages
→ []Message

GET /api/sessions/{id}/dice-rolls
→ []DiceRoll

GET /api/sessions/{id}/timeline
→ []TimelineEntry

PATCH /api/sessions/{id}
Body: { "summary": "..." }
→ 204 No Content

POST /api/sessions/{id}/recap                    (AI required)
→ { "summary": "..." }
```

### Maps & Pins

```
GET /api/maps/{id}
→ CampaignMap

GET /api/maps/{id}/pins
→ []MapPin
```

### World Notes

```
PATCH /api/world-notes/{id}
Body: { "title": "...", "content": "...", "tags_json": "[\"tag1\"]" }
→ 204 No Content
```

### Characters

```
PATCH /api/characters/{id}
Body: { "data_json": "{\"hp\":35,\"level\":4}" }
→ 204 No Content

POST /api/characters/{id}/portrait
Content-Type: multipart/form-data
Fields: portrait (file, max 5 MB, jpg/png/gif/webp)
→ { "portrait_path": "portraits/4_filename.jpg" }
```

### Rulesets & Rulebook

```
GET /api/rulesets/{id}
→ Ruleset

POST /api/rulesets/{id}/rulebook
Content-Type: text/plain
Body: plain text (markdown headings split into chunks, max 1 MB)
→ { "chunks_created": N }

POST /api/rulesets/{id}/rulebook
Content-Type: multipart/form-data
Fields: rulebook (PDF file)
→ { "chunks_created": N }
```

### Game Context

```
GET /api/context
→ {
    "campaign": Campaign,
    "character": Character,
    "session": Session,
    "recent_messages": []Message,   // last 20
    "active_combat": CombatSnapshot | null
  }
```

### Static Files

```
GET /api/files/{path}
→ File contents from ~/.ttrpg/{path}
```

Used to serve portrait images and map images. Path traversal outside `~/.ttrpg/` is blocked.

---

## WebSocket Events

Connect to `ws://localhost:7432/ws`. The server broadcasts JSON events on every state change. The dashboard connects automatically and uses these to update panels without polling.

**Event shape:**

```json
{ "type": "event_type", "payload": { ... } }
```

| `type` | Trigger | Payload keys |
|---|---|---|
| `campaign_created` | Campaign created | `campaign_id`, `name` |
| `character_created` | Character created | `character_id`, `name` |
| `character_updated` | Character data or portrait changed | `character_id`, optionally `portrait_path` |
| `session_started` | Session created | `session_id`, `title` |
| `session_ended` | Session closed | `session_id` |
| `session_updated` | Session summary changed | `session_id`, `summary` |
| `message_created` | Narrative logged to session | `session_id`, `content` |
| `dice_rolled` | Dice roll logged | `expression`, `total`, `breakdown` |
| `combat_started` | Encounter created | `encounter_id`, `name` |
| `combatant_updated` | Combatant HP or conditions changed | `combatant_id` |
| `combat_ended` | Encounter closed | `encounter_id` |
| `world_note_created` | World note created | `note_id`, `title` |
| `world_note_updated` | World note edited | `note_id` |
| `map_pin_added` | Pin added to map | `pin_id`, `map_id`, `label` |

The frontend reconnects automatically after a 2-second delay on disconnect.

---

## File Storage

All uploaded files are stored under `~/.ttrpg/`:

| Type | Directory | Naming | Max size |
|---|---|---|---|
| Portraits | `~/.ttrpg/portraits/` | `{character_id}_{original_filename}` | 5 MB |
| Maps | `~/.ttrpg/maps/` | `{32-char hex}{extension}` | 10 MB |

Accepted formats for both: `.jpg`, `.jpeg`, `.png`, `.gif`, `.webp`

Files are served at `GET /api/files/portraits/{filename}` and `GET /api/files/maps/{filename}`.

The database file lives at `~/.ttrpg/ttrpg.db`. Back this up to preserve all campaign data.

---

## Development

```bash
# Run with hot reload (starts Vite dev server + air for Go)
make dev

# Run all Go tests
make test

# Run frontend tests
cd web && npm test

# Production build
make build

# Clean artifacts
make clean
```

**Vite dev proxy:** When running `make dev`, the frontend dev server proxies `/api` and `/ws` requests to `localhost:7432`, so you can open the Vite port directly and all API calls go to the running Go server.

**Database inspection:**

```bash
sqlite3 ~/.ttrpg/ttrpg.db
.tables
SELECT * FROM rulesets;
SELECT * FROM campaigns;
SELECT * FROM characters;
```

**Running the binary without blocking on stdin** (useful for manual API testing while the MCP server is also running):

```bash
sleep infinity | ttrpg &
curl http://localhost:7432/api/health
```
