# ink & bone

**A private, local AI Game Master for 13 tabletop RPG systems — no cloud, no subscriptions, no limits.**

You type. Claude narrates. Your browser updates live. That's the whole loop — and nothing else on the market does it like this.

---

## Why ink & bone?

Every other AI GM tool is a SaaS product: cloud servers, token limits, monthly subscriptions, and a UI wrapper that constrains the AI so tightly it stops feeling intelligent. ink & bone is the opposite.

**It runs on your machine.** Your campaigns, characters, and session logs live in a SQLite database on your computer. Nothing leaves unless you push it. No account required. No waiting room when your tokens run out at a critical moment in the story.

**It uses the full Claude model, not a stripped-down wrapper.** Because ink & bone works through Claude Code's MCP tool layer, Claude reasons about rules, maintains context across the entire conversation history, and makes genuine GM judgment calls. Competitors constrain their AI so heavily to prevent hallucination that they also remove its capability. ink & bone grounds Claude in a structured database — so it remembers everything — without lobotomizing it.

**It supports 13 game systems out of the box — and any game you already own.** The rest of the field is almost entirely D&D 5e with a coat of paint. ink & bone ships with Ironsworn, Wrath & Glory, Blades in the Dark, Vampire: The Masquerade, Call of Cthulhu, Shadowrun, Warhammer Fantasy Roleplay, Star Wars Edge of the Empire, Legend of the Five Rings, The One Ring, Paranoia, and more. But the ruleset system is open: if you own a game that isn't on the list, you define its character sheet fields in a single JSON insert and it works immediately — correct field labels, correct input types, correct sheet layout in the browser. You can also upload the official PDF or text of any rulebook and Claude will search it during play, answering rules questions from the actual text rather than guessing.

**It costs fractions of a cent per message.** Using Claude Haiku, a full combat scene costs less than a penny. Competitors charge $10–30/month for token-gated access to a less capable model.

---

## What is ink & bone? (In Plain English)

You sit down and tell Claude a story about your character. Claude plays everyone else — the shopkeeper, the dragon, the mysterious stranger. Claude describes what happens, rolls the dice when there's uncertainty, tracks your character's health and equipment, and remembers everything that came before.

Your browser dashboard shows it all as it happens: your character's stats, the conversation transcript, dice rolls, combat, maps, NPCs, and world-building notes. Everything syncs live — no refresh, no waiting, no cloud roundtrip.

Think of it as a collaborative storytelling tool where the AI is the Game Master and you are the player. The browser is your character sheet and record keeper combined.

---

## What is a Tabletop RPG? (For First-Timers)

You've probably played video game RPGs like Baldur's Gate or Skyrim. Tabletop RPGs (TTRPGs) are the opposite direction in time — they started before computers.

Here's how they work:

**The Setup:** You play one character. Everyone else — the merchant, the guard, the villain, the monsters — is run by the Game Master (GM). In ink & bone, Claude is the GM.

**The Flow:** You describe what your character does. "I want to pick the lock on that chest." Claude narrates what happens. "You carefully insert your lockpicks... click. The lock gives way." The GM tells you what you see, hear, and feel. You respond with what you do next.

**Uncertainty and Dice:** When something's outcome is uncertain — will you successfully persuade the king, dodge that fireball, climb that cliff — you roll dice to add chance to the story. Claude rolls the dice, tells you the result, and narrates what happens. Success or failure, the story moves forward.

**The Goal:** There are no winning and losing states. The goal is to tell an interesting story together. Some campaigns are heroic quests. Others are about survival, politics, or exploration. Some are funny. Some are serious or scary. The rules define how your character works, what they can do, and how the dice work — but the story is always yours.

**Your Character:** You create a character sheet — a person with a name, abilities, skills, and equipment. Your character sheet changes as you gain experience, find treasure, or get hurt. This sheet tracks everything from your health to your inventory to your magical abilities.

---

## How ink & bone Works (The Big Picture)

```
┌──────────────────┐   tell a story    ┌─────────────────────┐
│  Claude Code     │ ◄──────────────► │  ink & bone app     │
│  (your GM)       │  save game state  │  (SQLite + tools)   │
└──────────────────┘                   └─────────────────────┘
                                               ▲
                                               │ live updates
                                               ▼
                                    ┌──────────────────────┐
                                    │ localhost:7432       │
                                    │ Character sheet      │
                                    │ Combat tracker       │
                                    │ Session log          │
                                    │ World notes / maps   │
                                    └──────────────────────┘
```

**Step 1: You type a message in Claude Code.** Your message is just plain English: "I want to sneak past the guards" or "Roll for initiative, combat starts now."

**Step 2: Claude (as GM) responds.** Claude reads your character's sheet, the current game state, and the ruleset. Claude then narrates what happens, rolls dice if needed, and calls tools to update the game.

**Step 3: The app saves everything.** All dice rolls, character updates, NPCs, combat turns, and notes are saved to a local SQLite database.

**Step 4: Your browser dashboard updates.** In real time, your browser shows your character sheet, the session transcript, combat status, and anything else Claude changes. No refresh needed — it's all live.

Repeat. That's it. You play from Claude Code. The browser is your always-open reference.

---

## Install & Run

### Prerequisites

- **An API key for Claude** — Set `ANTHROPIC_API_KEY` in your shell environment. Get one free at [console.anthropic.com](https://console.anthropic.com).
- **Go 1.22+** — Download from [golang.org](https://golang.org/dl).
- **Node.js 18+ and npm 9+** — Download from [nodejs.org](https://nodejs.org).
- **Claude Code** — Install from [claude.com/claude-code](https://claude.com/claude-code) and configure the `~/bin/ttrpg` wrapper.

### Quick Start

Clone the repo and build the binary:

```bash
git clone https://github.com/digitalghost/inkandbone
cd inkandbone
make install
```

This builds the Go server with embedded React frontend and installs the binary to `~/bin/ttrpg-bin`.

You must also have a wrapper script at `~/bin/ttrpg` that passes your API key to the binary:

```bash
#!/bin/bash
export ANTHROPIC_API_KEY="your-key-here"
~/bin/ttrpg-bin "$@"
```

Make it executable:

```bash
chmod +x ~/bin/ttrpg
```

Now start a game from Claude Code:

```
/ttrpg new "My Campaign" ironsworn
```

This starts the server on `localhost:7432` and opens your browser to the dashboard. You can now play.

### Development

For live reloading during development:

```bash
make dev
```

This runs the Go server with hot reload (via `air`) and the React Vite dev server concurrently.

---

## Playing Your First Game

### 1. Create a Campaign

In Claude Code, create a new campaign:

```
/ttrpg new "My Campaign Name" ironsworn
```

Replace `ironsworn` with any supported ruleset (see list below). The server starts on `:7432` and opens your browser.

### 2. Create a Character

Click "+ New Character" in the dashboard. You'll fill out a character sheet for your chosen ruleset. All numeric fields are auto-rolled; all choice fields are auto-selected from canonical options. Nothing is left blank.

Example for Ironsworn: Edge 2, Heart 1, Iron 3, Shadow 2, Wits 2. Health 5, Spirit 5, Supply 5, Momentum 1.

### 3. Start a Session

Click "+ New Session" and give it a title (e.g., "The Lost Library"). This opens a blank session where you and Claude can play together.

### 4. Tell Your Story

In Claude Code, type what your character does. Claude reads your message and responds as the GM, narrating what happens next. Your browser updates in real time.

```
In Claude Code:
> I want to search the ancient library for clues about the lost city.

Claude responds:
> You push through the heavy oak doors of the archive. The smell of leather and aged parchment fills your lungs. Shafts of afternoon light cut through dust motes as you scan the towering shelves. On a low table near the entrance, a leather-bound journal lies open...
```

**Everything is saved automatically** — your character sheet, the conversation, dice rolls, items, NPCs, and notes all persist in the local SQLite database.

### 5. Explore the Features as You Play

As you tell your story, you'll discover features unlocking automatically:

- **Dice rolls:** Type a 1d20 or 2d6 roll into Claude's turn, and it auto-executes. History appears in the left sidebar.
- **Combat:** When a fight starts, the turn order strip appears at the top. Claude manages initiative, HP, and conditions. You see the combat tracker in real time.
- **Maps:** When a location is described, Claude can generate an SVG map. Click a message to pin it to the map.
- **NPCs:** When Claude mentions a character's name, it's auto-added to your NPC roster in the right sidebar.
- **World Notes:** Click "Draft with AI" to auto-generate lore entries (locations, factions, items, etc.) with tags for easy filtering.
- **Objectives:** Claude detects story goals (quests, mysteries, personal goals) and adds them to a tracker. Mark them complete or failed.
- **Items & Inventory:** When you gain or lose gear, it's auto-tracked. View your full inventory in the left sidebar under items.
- **Session Recap:** Every 4 GM messages, Claude regenerates a summary of the session in the Journal tab.

---

## The Dashboard (UI Reference)

The interface is called the "Worn Grimoire" — a parchment-dark theme with warm gold accents, serif typography, and ornamental separators.

### Header (Breadcrumb & Controls)

At the top:

- **Breadcrumb:** `Campaign Name › Character Name › Session Title` (campaign in gold, others in dim text)
- **Theme toggle (☀/🌙):** Top right. Switch between dark and light themes. Preference is saved to localStorage.
- **Actions button (⚔ Actions):** Opens a panel showing only your player messages in chronological order, excluding GM narration. Click × to close.
- **Export button (↓ Export):** Downloads the full session as a `.md` file with all narration and player actions (whispers excluded).

### Left Sidebar — Character Sheet

Your character's live stats and tracker:

- **Portrait area:** Circular 80px image. Click to upload a JPG, PNG, GIF, or WebP (up to 5 MB). Shows character initial if no portrait.
- **Attributes & Tracks (system-specific):** For Ironsworn, attribute pips and track bars. For D&D, six ability scores. All fields are live-editable.
- **All ruleset fields:** Every field defined by the campaign ruleset appears here as an editable input. Changes save instantly without refresh.
- **Dice roller buttons:** Six buttons (d4, d6, d8, d10, d12, d20) for quick in-browser rolls. See results and history below.
- **Dice history:** Last 5 rolls in this session, showing expression and result.

### Center Column — Session Transcript

The main story log and interaction panel:

- **Session header:** Centered ornamental title with session name in all caps (e.g., "✦ WHISPERS IN THE MIST ✦") and date in small dim text below.
- **Turn order strip (combat only):** When combat is active, a horizontal strip shows all combatants as chips with their initiative. Active combatant is highlighted. Dead combatants (0 HP) appear dimmed and struck through.
- **Story search bar:** Filter messages by content. Matching text is highlighted in gold. Click × to clear.
- **Combat panel (combat only):** One card per combatant showing name, HP bar (color-coded by health %), initiative, and clickable condition badges (poisoned, paralyzed, etc.).
- **GM narration:** Larger prose blocks (unmarked) describing what happens. Markdown is rendered (bold, italics, lists). Text streams character-by-character with a blinking cursor when the GM is responding.
- **Player actions:** Your messages appear in italic gold text, labeled with your character's name (e.g., "KAEL SPEAKS"). Whispered messages have a 🔒 lock icon and appear dimmed (not included in GM memory or session exports).
- **Separator diamonds (◆):** Mark boundaries between turns for readability.
- **Message persistence:** All messages are saved permanently and persist between sessions.

Below the story scroll:

- **Player input bar:** Textarea where you type your character's action or dialogue.
  - **Whisper toggle (🔒):** Click to mark your message as private. When enabled, the GM won't read it in the conversation context, and it won't be included in exports. The button highlights.
  - **Send button (↵):** Press Enter or click to send. Disabled if the session is inactive or the input is empty.

At the bottom:

- **Map drawer:** Expandable/collapsible section showing campaign maps. Collapsed state shows a thin bar (`[ CAMPAIGN NAME ▾ ]`). Expanded state fills ~60% of column height and displays the full map image with clickable pins (hover to see label and notes). The 📍 button on GM messages opens a modal to place that message as a map pin.
- **Generate Map button (AI enabled only):** Click to generate a new map using Claude based on recent session context. Auto-selects the new map in the drawer.

### Right Sidebar — Notes, Journal, NPCs, and Objectives

Four tabs showing different campaign data:

**Notes tab (World Notes):**
- Search bar to filter by keyword or category (NPC, location, faction, item, other).
- Note cards with title in serif gold, content in body text, and tags as small dim pills.
- Click any note to read the full text in a modal.
- Click "Draft with AI" (AI enabled only) to generate a new note from a hint prompt (e.g., "a mysterious hooded figure" → auto-generates lore, name, and tags).

**Journal tab (Session Recap):**
- Textarea for freeform notes about the session.
- Write your own summary, or click "Generate recap" to have Claude read the entire session and draft a narrative summary.
- Changes auto-save on blur via PATCH request.
- AI recap captures the narrative arc, key decisions, character growth, and major events.

**NPCs tab (NPC Roster):**
- A roster of named NPCs for the current session.
- Each NPC shows a name label, editable note field (auto-saves on blur), and a delete button.
- "+ Add NPC" button at the bottom to create new entries.
- NPCs auto-add when Claude mentions a character's name during play.
- Live WebSocket updates when NPCs are added or modified.

**Objectives tab (Quest Tracker):**
- List of active, completed, and failed objectives for the campaign.
- Click to view full description or edit status.
- Click "+ New Objective" to add a quest or goal manually.
- Objectives auto-add when Claude detects a new story goal.

**Items tab (Inventory):**
- All items owned by your character.
- Shows name, description, quantity, and equipped status.
- Click to edit or delete items.
- Items auto-add when Claude mentions you gaining or losing gear.

---

## Supported Rulesets

A "ruleset" is the game system you're playing. Each has different character sheet fields, abilities, and mechanics. ink & bone ships with 13 built-in systems.

### Built-in Systems

**Ironsworn** (`ironsworn`) — Dark Norse fantasy, solo-friendly, vow-driven narratives. Good for beginners.

**Wrath & Glory** (`wrathglory`) — Grimdark Warhammer 40K. Over-the-top sci-fi combat and corruption.

**Blades in the Dark** (`bitd`) — Victorian ghosts, heists, and crew mechanics. Position and effect system.

**Vampire: The Masquerade** (`vtm`) — Gothic horror, vampire politics, and the struggle for humanity.

**Call of Cthulhu** (`cthulhu`) — 1920s cosmic horror. Sanity and investigators.

**Shadowrun** (`shadowrun`) — Cyberpunk + fantasy. Hacking, magic, and street samurai.

**Warhammer Fantasy Roleplay** (`whfrp`) — Gritty medieval fantasy with career progression.

**Star Wars: Edge of the Empire** (`sweoote`) — Star Wars expanded universe, scoundrels and rogues.

**Legend of the Five Rings** (`l5r`) — Japanese-inspired fantasy with honor mechanics.

**The One Ring** (`lotr`) — Middle-earth journeys and fellowships.

**Paranoia** (`paranoia`) — Cold War sci-fi satire. Backstabbing and high lethality.

**Dungeons & Dragons 5th Edition** (`dnd5e`) — High fantasy with classes, levels, and classic monsters.

**Custom Ruleset** — Define your own fields in JSON and add them directly to the database.

### Using a Ruleset

When you create a campaign, pick one:

```
/ttrpg new "My Campaign" ironsworn
```

When you create a character, all fields from that ruleset's schema are presented as form inputs. Numeric fields are auto-rolled; choice fields are auto-selected from canonical options.

### Adding Your Own Ruleset

If your favorite game isn't on the list, you can add it:

1. Define the ruleset schema as JSON (field name, type, default value, options for enums).
2. Insert it directly into the `rulesets` table.
3. Create a campaign using your custom ruleset.
4. When characters are created, they'll have exactly the fields you defined.

### Rulebooks and Supplemental Material

#### What You Can Do Without a Rulebook

ink & bone ships with built-in character sheet schemas for all 13 supported systems. Without a rulebook uploaded, you can still:

- Create characters, run sessions, and have Claude narrate the story
- Track stats, health, inventory, combat, maps, and NPCs automatically
- Roll dice with proper expressions (d4 through d100, pools, modifiers)
- Get a competent AI GM that applies common-sense interpretations of the rules

**What's missing without the rulebook:**

- Claude cannot cite specific page references or exact rule text
- Edge-case rules (unusual combat conditions, specific spell interactions, advanced abilities) rely on Claude's general knowledge rather than the actual book, and may be wrong or outdated
- Class-specific abilities, feats, spells, and equipment tables are not indexed
- Rules clarifications, errata, and system-specific edge cases are not available

For casual play this is fine. For rules-precise, competitive, or rulebook-faithful play, uploading the PDF gives Claude the full text to search.

---

#### What Types of Books You Can Upload

Every TTRPG line produces several categories of material. ink & bone treats them all the same — they're indexed and searched together — but labeling them keeps things organized.

**Core Rules (label: `Core Rulebook`)**
The foundation. Covers character creation, core resolution mechanics, combat rules, advancement, equipment, and basic world-building. This is the most important upload. Every table needs this.

Examples: *Player's Handbook* (D&D), *Wrath & Glory Core Rulebook*, *Ironsworn*, *Blades in the Dark*, *Call of Cthulhu Keeper Rulebook*, *VtM 5th Edition Core*.

**Game Master's Guide / Keeper's Guide (label: `GM Guide`)**
Rules and advice for running the game: encounter design, NPC creation, loot tables, campaign structure, secret rules, and setting up adventures. Useful for Claude when adjudicating GM-side mechanics.

Examples: *Dungeon Master's Guide*, *Call of Cthulhu Keeper's Guide*, *Blades in the Dark* (the book itself doubles as both).

**Bestiary / Monster Manual (label: `Bestiary`)**
Creature stat blocks, abilities, tactics, lore, and encounter suggestions. Upload this if you want Claude to reference accurate creature stats rather than improvising them.

Examples: *Monster Manual* (D&D), *Wrath & Glory Threat Assessment: Xenos*, *VtM Anarch Cookbook*, *WFRP Bestiary*.

**Player Options / Sourcebook (label: `Sourcebook: [name]`)**
Expands character options: new classes, subclasses, archetypes, races, backgrounds, spells, feats, prestige paths, and equipment. Upload these when your character uses options from outside the core book.

Examples: *Tasha's Cauldron of Everything*, *Xanathar's Guide*, *VtM Chicago By Night* (clans + coterie options), *Wrath & Glory Forsaken System Guide* (new archetypes), *Shadowrun Street Grimoire* (magic rules).

**Campaign Setting (label: `Setting: [name]`)**
World-building lore: geography, factions, history, politics, religions, and plot hooks specific to a setting. Gives Claude deep context for narrating the world accurately.

Examples: *Forgotten Realms Campaign Setting*, *VtM Chicago By Night*, *Wrath & Glory Gilead System*, *Eberron: Rising from the Last War*.

**Adventure Module / Campaign Book (label: `Adventure: [name]`)**
Pre-written scenarios, dungeons, encounters, NPCs, and story beats. Upload the adventure you're running so Claude can reference the actual plot, maps, and encounter stats.

Examples: *Curse of Strahd*, *Death on the Reik* (WFRP), *The Long Night* (IoS), *The Fall of Delta Green*.

**Faction / Clan Book (label: `Faction: [name]`)**
Deep lore and mechanical options for a specific faction, clan, or organization. Upload the ones your character belongs to.

Examples: VtM Clanbooks (Brujah, Malkavian, etc.), *Wrath & Glory Talents of Chaos* (Chaos faction), Shadowrun runner archetype books.

**Rules Supplement / Errata (label: `Supplement: [name]`)**
Additional rules, optional subsystems, or official corrections. Combat rules expansions, social encounter frameworks, crafting systems, vehicle rules, etc.

Examples: *WFRP Winds of Magic*, *Ironsworn Starforged*, *D&D Dungeon Master's Screen* (quick reference), official errata PDFs.

---

#### How to Upload a Rulebook

**The first time — uploading the core rulebook:**

1. Find your ruleset ID:
   ```
   GET /api/campaigns  → note the ruleset_id field
   GET /api/rulesets/{id}  → confirms the system name
   ```

2. Upload a PDF (up to 50 MB):
   ```bash
   curl -X POST http://localhost:7432/api/rulesets/{id}/rulebook \
     -F "rulebook=@/path/to/corebook.pdf" \
     -F "source=Core Rulebook"
   ```

3. Upload plain text (up to 2 MB) — useful for SRDs and free rules:
   ```bash
   curl -X POST http://localhost:7432/api/rulesets/{id}/rulebook \
     -H "Content-Type: text/plain" \
     --data-binary @/path/to/rules.txt \
     "?source=Core+Rulebook"
   ```

**Adding an expansion without overwriting the core book:**

Each `source` label is stored independently. Uploading a new source only replaces chunks with that same label — all other books remain intact.

```bash
# Upload a bestiary — core book is untouched
curl -X POST http://localhost:7432/api/rulesets/{id}/rulebook \
  -F "rulebook=@/path/to/bestiary.pdf" \
  -F "source=Bestiary"

# Upload a campaign setting
curl -X POST http://localhost:7432/api/rulesets/{id}/rulebook \
  -F "rulebook=@/path/to/setting.pdf" \
  -F "source=Setting: Gilead System"

# Upload the adventure you're running
curl -X POST http://localhost:7432/api/rulesets/{id}/rulebook \
  -F "rulebook=@/path/to/adventure.pdf" \
  -F "source=Adventure: The Long Night"
```

**Re-uploading an updated edition:**

Re-uploading with the same `source` name replaces only that source's chunks. Use this to update errata, fix a bad extraction, or swap editions.

**Check what's been uploaded:**

```bash
GET /api/rulesets/{id}/rulebook
# Returns: [{"source":"Core Rulebook","chunks":312},{"source":"Bestiary","chunks":180}]
```

---

#### Tips for Better PDF Extraction

- **Digital PDFs extract better than scanned books.** Scanned PDFs are images — pdfcpu can't read them. Use digitally-typeset PDFs (buy from DriveThruRPG or publisher websites for best results).
- **If extraction returns 0 chunks**, the PDF may be image-only. Try a plain-text version instead (many publishers sell both).
- **Large books may exceed 50 MB** — particularly art-heavy hardcovers. Try uploading just the rules chapters as separate files, or use a text export.
- **Pre-format text files with `#` headings** to improve chunk quality. Each `#` line becomes a searchable section heading.
- **Markdown and plain text SRDs** work perfectly. Systems like Ironsworn, Blades in the Dark, and many OSR games publish free SRDs as plain text.

---

## Features & Automation

### AI GM (Claude Haiku)

When you send a message, Claude reads:

- Your character sheet (all fields and current values).
- The full conversation history.
- The ruleset schema and rules context.
- Your campaign description and world notes.

Claude then narrates what happens, applying rules as needed, and responds via Server-Sent Events (SSE). The text streams character-by-character to your browser. A system prompt injects your character's name so Claude consistently uses the correct name when referring to you.

**Stream cleanup:** Em-dashes in Claude's output are automatically stripped programmatically before display.

### Automation Goroutines

After every GM response, eight background tasks fire automatically (no player action required):

**autoExtractNPCs** — Claude's text is analyzed for proper names. New named characters are automatically added to the session NPC roster. Names appear in the NPCs tab on the right sidebar.

**autoGenerateMap** — When Claude describes a new location with a proper name, an SVG map is generated and added to the campaign map gallery. You can view it and place pins on locations mentioned in the story.

**autoUpdateCharacterStats** — Detects story events that affect your character (taking damage, gaining XP, level-up, acquiring abilities, etc.). Applies rule-based stat updates automatically per ruleset. Changes appear instantly in the character sheet.

**autoUpdateRecap** — Every 4 GM responses, the session journal entry is regenerated with a fresh summary capturing narrative arc, key decisions, and character growth.

**autoDetectObjectives** — Claude's response is analyzed for newly introduced story goals (quests, mysteries, personal objectives). New objectives are added to the Objectives tab automatically.

**autoExtractItems** — When Claude describes items you gain or lose, they're automatically added to or removed from your inventory (Items tab).

**checkAndExecuteRoll** — Before Claude responds, the player's action is analyzed. If the ruleset requires a dice roll for that action (e.g., attack roll in combat, climb check, persuasion roll), the roll is enforced first. Claude sees the result and narrates accordingly. Keeps story moving without manual dice rolling.

**autoUpdateTension** — After every GM response, the session's tension level automatically increments if the text contains crisis keywords (ambush, betrayal, catastrophe, danger, doom, enemy, escape, failure, fear, fight, flee, loss, peril, threat, trapped, wounded, etc.) or when a dice roll critically fails. The tension tracker is visible in the session UI and influences narrative pacing. You can manually adjust tension via the UI at any time.

All automation runs in the background without interrupting your gameplay. Updates appear in real time via WebSocket.

### NPC Personality System

World notes with the `npc` category can store a personality profile as structured JSON. This allows you to define NPC traits that Claude incorporates into every turn:

**To set an NPC personality:**

```bash
PATCH /api/world-notes/{id}/personality
Content-Type: application/json

{
  "personality_json": "{\"traits\": [\"cunning\", \"honorable\"], \"motivations\": \"power and legacy\", \"quirks\": \"speaks in riddles\"}"
}
```

The personality JSON is injected into Claude's world context block before every GM turn. Any valid JSON object is accepted — use whatever fields describe your NPC best. Claude will reference personality traits when the NPC appears in the story.

### World Context Injection

When you send a player action, Claude receives an enriched world context block that includes:

- `[ACTIVE OBJECTIVES]` — All active quests and story goals for the campaign, so Claude tracks narrative threads without you having to remind them.
- `[NPC: Name]` personality cards — For every world note tagged as `npc` with a non-empty personality JSON, Claude receives the personality definition and incorporates it into NPC dialogue and actions.

This ensures NPCs stay consistent and plot threads remain visible throughout the session.

### GM Session Tools

Four AI-powered endpoints let you get Claude's analysis of the campaign and session. All require `ANTHROPIC_API_KEY` to be set:

**POST /api/sessions/{id}/improvise** — Generate an improvised scene suggestion, NPC complication, or plot twist from the last 5 messages. Returns `{"result":"..."}` with a 2-3 sentence suggestion.

**POST /api/campaigns/{id}/pre-session-brief** — Generate a concise GM prep brief (3-5 bullets) summarizing world notes and active objectives. Use this before your next session to remember what happened and what's at stake. Returns `{"result":"..."}`.

**POST /api/sessions/{id}/detect-threads** — Analyze the full session transcript and identify unresolved narrative threads, loose ends, and plot hooks for future sessions. Returns `{"result":"..."}` with a list of thread recommendations.

**POST /api/campaigns/{id}/ask** — Ask Claude a freeform question about your campaign. Body: `{"question":"..."}`. Claude uses world notes as context to answer. Returns `{"result":"..."}` with the answer.

All four endpoints are handy for GM prep, campaign planning, and breaking writer's block mid-session.

### Oracle Tables & Narrative Systems

**Oracle Tables** — Roll dice against seeded action and theme tables (50 rows each, numbered 1-50). Use oracles for random inspiration when you're stuck:

```bash
POST /api/oracle/roll
Content-Type: application/json

{"table": "action", "roll": 23, "ruleset_id": null}
```

Response: `{"result":"Betray","table":"action","roll":23}`

Replace `"action"` with `"theme"` to roll the theme table. Custom rulesets can provide their own oracle tables.

**Tension Tracker** — Each session has a tension level (1-10, default 5) that influences narrative pacing and danger:

- `GET /api/sessions/{id}/tension` — View current tension level
- `PATCH /api/sessions/{id}/tension` — Manually set tension (body: `{"tension_level":N}`)

Tension auto-increments when Claude's responses contain crisis keywords (ambush, betrayal, catastrophe, danger, doom, escape, failure, fear, fight, loss, peril, threat, trapped, wounded) or on critical dice failures. You can override it anytime via the UI.

**Relationship Web** — Track named relationships between characters and factions to drive roleplaying and plot complications:

```bash
POST /api/campaigns/{id}/relationships
{"from_name": "Kael", "to_name": "The Warlord", "relationship_type": "enemy", "description": "Killed Kael's mentor five years ago"}

GET /api/campaigns/{id}/relationships
→ [{id: 1, from_name: "Kael", to_name: "The Warlord", relationship_type: "enemy", description: "...", created_at: "2026-04-04T..."}]

PATCH /api/relationships/{id}
{"relationship_type": "rival", "description": "Now seeks redemption through direct challenge"}

DELETE /api/relationships/{id}
```

Relationships are campaign-wide and persist. Use them to track feuds, alliances, mentorships, and rivalries that shape your story.

### Audio & Ambience

**Procedural Sound Effects** — Web Audio API synthesis provides automatic sound effects during play:

- Dice rolls trigger a percussive rattle sound when `dice_rolled` events occur.
- New messages trigger an ascending two-tone chime notification.
- Combat start triggers a low sawtooth pulse when `combat_started` is broadcast.

All sounds respect the mute toggle and volume slider in the grimoire header. No audio files required — synthesis happens in your browser.

**Ambient Audio Loops** — Set the mood with ambient music tied to scene tags. Place your own MP3 files in `~/.ttrpg/audio/` with names matching scene tags (e.g., `tavern.mp3`, `dungeon.mp3`, `forest.mp3`):

```
~/.ttrpg/audio/
├── tavern.mp3
├── dungeon.mp3
├── forest.mp3
├── city.mp3
├── ocean.mp3
├── cave.mp3
├── castle.mp3
├── rain.mp3
├── night.mp3
├── battle.mp3
├── market.mp3
├── temple.mp3
└── ruins.mp3
```

Supported scene tags (13 total): `tavern`, `dungeon`, `forest`, `city`, `ocean`, `cave`, `castle`, `rain`, `night`, `battle`, `market`, `temple`, `ruins`.

**How Ambient Audio Works:**

1. Click the scene tag buttons in the session header to toggle tags on/off.
2. The first active tag is used to select the ambient audio track (e.g., if `dungeon` and `rain` are both active, `dungeon.mp3` plays).
3. The ambient track fades in smoothly and loops continuously.
4. Audio respects the master mute toggle and volume slider in the header.
5. When you switch scene tags, the ambient track fades out and a new one fades in.

**AudioControls Component** — The grimoire header displays a mute toggle (🔔/🔕) and volume slider (0–100%). Settings persist in browser localStorage, so your audio preferences are remembered between sessions.

---

## Architecture

### Tech Stack

- **Go 1.22+:** HTTP server, SQLite database layer, MCP integration.
- **SQLite:** Persistent session, character, and campaign data in a single local file (`~/.ttrpg`).
- **React 18 + TypeScript:** Vite-bundled frontend, embedded in the binary.
- **WebSocket:** Live dashboard updates from server to browser.
- **SSE (Server-Sent Events):** Streaming GM responses for character-by-character prose display.
- **Claude Haiku (MCP tools):** AI GM reasoning and context injection.
- **Dice library:** Dice roll expression parsing and evaluation.

### Project Layout

```
inkandbone/
├── cmd/ttrpg/              # Go binary entrypoint
├── internal/
│   ├── api/                # HTTP handlers, WebSocket hub, event bus, automation goroutines
│   ├── db/                 # SQLite schema, migrations, query layer
│   ├── ai/                 # Claude API integration, system prompts, SSE streaming
│   └── utils/              # Dice roller, ruleset validation, rulebook parsing
├── web/                    # React/TypeScript frontend (Vite)
│   ├── src/
│   │   ├── components/     # UI panels, character sheet, combat tracker, map viewer
│   │   ├── pages/          # Campaign list, session view, settings
│   │   └── hooks/          # WebSocket, API requests, state management
│   └── package.json        # npm dependencies
├── Makefile                # Build, install, dev commands
└── README.md               # This file
```

### Database Schema

SQLite stores campaigns, characters, sessions, messages, NPCs, world notes, maps, pins, dice rolls, objectives, items, combat encounters, oracle tables, relationships, and narrative tension. All data persists locally. Migrations are applied on startup.

Key tables:

- `campaigns` — Campaign metadata and ruleset reference.
- `characters` — Player characters with stats (JSON) and portrait path.
- `sessions` — Play sessions with title, date, summary, and `tension_level` (1-10).
- `messages` — Full conversation history (role: 'user' or 'assistant').
- `session_npcs` — Named characters for each session.
- `world_notes` — Lore entries (locations, NPCs, factions, items) with optional `personality_json` for NPC profiles.
- `maps` — Campaign maps (uploaded or AI-generated SVG).
- `map_pins` — Pins placed on maps with labels and notes.
- `objectives` — Story goals (active, completed, failed).
- `items` — Character inventory (name, description, quantity, equipped status).
- `combat_encounters` — Combat tracks (one per encounter).
- `combatants` — Combatants in an encounter (initiative, HP, conditions).
- `dice_rolls` — Roll history with expression and result breakdown.
- `oracle_tables` — Seeded oracle tables (action and theme) with 50 rows each. Custom rulesets can provide their own.
- `relationships` — Named relationships between characters/factions (from_name, to_name, relationship_type, description, campaign_id).
- `scene_tags` — Session scene tags (tavern, dungeon, forest, city, ocean, cave, castle, rain, night, battle, market, temple, ruins) linked to sessions for ambient audio selection.

---

## Troubleshooting

### The browser shows a blank page

Ensure the server is running and listening on `localhost:7432`. Check that `ANTHROPIC_API_KEY` is set in your environment.

```bash
echo $ANTHROPIC_API_KEY
```

If empty, set it before running `ttrpg`:

```bash
export ANTHROPIC_API_KEY="sk-..."
```

### Dice rolls don't execute

Check that the ruleset has dice roll rules defined. Some systems don't enforce automatic dice rolls. You can always roll manually using the dice buttons in the left sidebar.

### Maps aren't generating

Maps require AI to be enabled (ANTHROPIC_API_KEY must be set). If AI is enabled, Claude analyzes the last few messages for location names and generates an SVG map. This can take a few seconds.

### Character stats aren't updating

The automation goroutine `autoUpdateCharacterStats` runs after every GM response. It analyzes Claude's text for events that affect your character (damage, healing, level-up, etc.). If updates don't appear, check that the ruleset schema has the correct field names.

### WebSocket connection drops

The client automatically reconnects every 5 seconds. If the connection persists in dropping, check that your firewall allows localhost connections and that the server is still running.

### Sessions are empty after restart

Sessions and characters are saved in the local SQLite database. If the database file is lost or corrupted, data is gone. Keep backups of `~/.ttrpg`.

---

## Contributing

Contributions are welcome. If you have a new ruleset, a feature idea, or a bug fix, please open an issue or pull request on GitHub.

### Adding a New Ruleset

1. Create a JSON schema defining all character fields (name, type, default, options).
2. Add the schema to the seed migrations file (`internal/db/migrations/002_seed_rulesets.sql`).
3. Test by creating a campaign and character with the new ruleset.

### Adding a Feature

Discuss larger features in an issue first. For bug fixes, submit a PR with a clear description of the problem and solution.

### Testing

Run all tests:

```bash
make test
```

Tests cover the database layer, API handlers, automation goroutines, and AI integration.

---

## License

MIT. See LICENSE file for details.

---

## Support & Feedback

If you have questions, found a bug, or want to share what you're building, open an issue on GitHub or reach out directly.

Enjoy your game.
