# ink & bone

A tabletop RPG companion app where you tell stories with an AI that acts as your Game Master. Everything you need — the story engine, character tracking, combat, maps, notes — runs in one app on your computer. Open your browser at localhost:7432, chat with Claude, and watch your adventure unfold live.

No cloud. No subscriptions. Just you, your stories, and an AI that keeps track of everything.

---

## What is ink & bone? (In Plain English)

You sit down and tell Claude a story about your character. Claude plays everyone else — the shopkeeper, the dragon, the mysterious stranger. Claude describes what happens, rolls the dice when there's uncertainty, tracks your character's health and equipment, and remembers everything that came before.

Your browser dashboard shows it all as it happens: your character's stats, the conversation transcript, dice rolls, combat, maps, NPCs, and world-building notes. Everything syncs live.

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

## Choose Your Game (Rulesets)

A "ruleset" is the game system — the set of rules that govern how your story works, what abilities your character has, and how dice are used. Think of rulesets like different board games: each has different pieces, different rules, and a different feel.

ink & bone comes with five built-in games. **Choose one before installing.** Each is designed for different types of stories.

### Dungeons & Dragons 5th Edition (`dnd5e`)

**What it's about:** High fantasy adventure. Elves, dwarves, wizards, clerics, rogues. Magic. Dragons. Ancient ruins. Taverns. Political intrigue. This is the most famous fantasy RPG in the world.

**Good for:** Classic adventure stories. Lots of character options (races, classes, subclasses). Well-documented online so you can look things up easily. But the rulebook is thick and complex.

**Good for beginners?** Not first choice — there's a lot of complexity. **Choose this as your second game once you understand how ink & bone works.**

**Your character sheet has:** race, class, level, hit points (HP), armor class (AC), six core abilities (strength, dexterity, constitution, intelligence, wisdom, charisma), proficiency bonus, skills, inventory, spells, and special features.

**Example:** You're a human rogue named Kess. Your stats: level 3, HP 21, AC 15, strength 10, dexterity 16 (high), constitution 12, intelligence 13, wisdom 14, charisma 11. You're sneaky and quick.

---

### Ironsworn (`ironsworn`)

**What it's about:** Dark Norse-inspired fantasy. You're a warrior who makes vows — promises you make to yourself and to NPCs — and the whole story revolves around keeping or breaking those vows. Built specifically for solo play, so you don't need other players.

**Good for:** Simpler rules. Personal, character-driven stories. Survival and struggle. The ruleset literally helps you create plot hooks.

**Good for beginners?** **Yes — this is the best choice.** Simpler mechanics, the rules help you play, and the community is welcoming to new players. Free to download online if you want to learn more.

**Your character sheet has:** five core abilities (Edge — quick and clever, Heart — empathy and bravery, Iron — raw strength, Shadow — cunning and deception, Wits — knowledge and perception), health, spirit, supply, momentum (a currency that powers your abilities), vows you've made, bonds with NPCs you've met, assets (abilities you've unlocked), and notes.

**Example:** You're a warrior named Kael. Your stats: Edge 2, Heart 1, Iron 3 (you're strong), Shadow 2, Wits 2. You have 5 health, 5 supply, 1 momentum, and the vow "Find the lost library of Valdris."

---

### Vampire: The Masquerade (`vtm`)

**What it's about:** Modern gothic horror. You're a vampire navigating a secret society of other vampires, politics, hunts for blood, and the constant struggle between your vampire nature and your remaining humanity.

**Good for:** Urban fantasy. Character-driven political intrigue. Exploring themes of power, corruption, and what you'll sacrifice to survive. Contains mature themes.

**Good for beginners?** No — assumes you're comfortable with character-driven stories and moral ambiguity.

**Your character sheet has:** clan (your vampire bloodline), generation, humanity (how much of your human morality remains), blood pool (how much blood you have), willpower, attributes, abilities, disciplines (vampire powers), virtues, backgrounds (your history and connections), and notes.

---

### Call of Cthulhu (`coc`)

**What it's about:** 1920s horror investigation. You're an ordinary person — a detective, journalist, professor — uncovering cosmic horrors inspired by H.P. Lovecraft's stories. As you learn terrible truths about the universe, your character slowly loses sanity.

**Good for:** Mystery. Horror and suspense. Stories where knowledge is as dangerous as combat. A very different tone from fantasy.

**Good for beginners?** No — the sanity mechanics are complex and the tone is heavy.

**Your character sheet has:** occupation (your job), age, hit points, sanity (how much you can witness before breaking), luck, magic points, six core abilities (strength, constitution, size, dexterity, appearance, intelligence, power, education), skills, inventory, and notes.

---

### Cyberpunk Red (`cyberpunk`)

**What it's about:** Dystopian near-future sci-fi. You're a hacker, merc, fixer, or nomad in a sprawling megacity controlled by mega-corporations. Chrome implants replace body parts. Money is tight. Violence is fast and brutal.

**Good for:** Fast-paced action. Gritty, dark settings. Tech-focused characters. Stories about resistance and survival against overwhelming odds.

**Good for beginners?** No — fast-paced action requires quick decision-making and the setting assumes sci-fi familiarity.

**Your character sheet has:** role (your archetype), nine core abilities (intelligence, reflexes, cool, technical ability, luck, attractiveness, movement allowance, empathy, body type), humanity (how much chrome you can have before losing yourself), eurodollars (money), skills, cyberware (implants), gear, and notes.

---

## Installation

### What You Need

Before installing ink & bone, make sure you have:

- **Go 1.22 or later** — the programming language the app is built in. Download from golang.org.
- **Node.js 18 or later with npm** — the runtime for building the frontend. Download from nodejs.org.
- **An Anthropic API key (optional)** — if you want the AI recap and world-note drafting features. Get one free at console.anthropic.com by signing up.

Check you have these installed:

```bash
go version      # Should print Go version 1.22 or higher
node --version  # Should print Node 18 or higher
npm --version   # Should print npm 9 or higher
```

### Download and Build

```bash
# Clone the repository
git clone https://github.com/digitalghost404/inkandbone
cd inkandbone

# Build everything (frontend + Go binary)
make build

# Install the binary to ~/bin/ttrpg (optional, but easier)
make install
```

After `make install`, you can run `ttrpg` from anywhere. Without it, use `./ttrpg` from the inkandbone directory.

---

## Configuration

The app works out of the box with no configuration. Everything is stored locally.

### Optional: Add Your API Key

If you want the AI-powered features (AI-drafted world notes and session recaps), get an Anthropic API key from console.anthropic.com. Then set it as an environment variable:

```bash
export ANTHROPIC_API_KEY="sk-ant-..."
ttrpg
```

Or add it to your shell profile (e.g., `~/.bashrc` or `~/.zshrc`):

```bash
export ANTHROPIC_API_KEY="sk-ant-..."
```

### Where Your Data Lives

- **Database:** `~/.ttrpg/ttrpg.db` — all your campaigns, characters, sessions, notes, and dice rolls
- **Portraits and Maps:** `~/.ttrpg/` — images you upload
- **Dashboard:** `localhost:7432` — served from the app itself (no cloud)

---

## First Run: Step by Step

### Step 1: Register the MCP Server

ink & bone talks to Claude Code through an "MCP server" — a bridge that lets Claude call tools in the app.

Edit `~/.claude/settings.json` and add this:

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

Replace `/path/to/ttrpg` with where the binary is. If you ran `make install`, use `~/bin/ttrpg`. If you're running from the inkandbone directory, use the full path like `/home/yourname/projects/inkandbone/ttrpg`.

The `env` block is optional — omit it if you don't have an API key or prefer to export `ANTHROPIC_API_KEY` from your shell instead.

### Step 2: Start the Server

```bash
ttrpg
```

This starts two things at once:
- The HTTP server at `localhost:7432` (your dashboard)
- The MCP stdio server (Claude Code connects to this)

You'll see output like:

```
MCP server ready on stdio
HTTP server listening on localhost:7432
```

Leave this running. Open Claude Code in a different terminal or IDE pane.

### Step 3: Open the Dashboard

In a web browser, go to `http://localhost:7432`. You'll see the dashboard — currently empty, since you have no campaigns yet.

### Step 4: Create Your First Campaign

In Claude Code, just ask:

```
What games can I play?
```

Claude will list the available rulesets (dnd5e, ironsworn, vtm, etc.). Now create a campaign:

```
Create an Ironsworn campaign called "The Ironlands" and create a character named Kael.
```

Claude will call tools to create both. The dashboard updates instantly — you'll see your campaign name and Kael's character sheet.

That's it. You're ready to play.

---

## How to Talk to Claude

This is the most important section. **You don't use commands or special syntax.** You just describe what you want in plain English.

### Basic Examples

Here are things you can say and what happens:

```
I want to explore the forest to the east.
```
Claude narrates what you find, describes the scenery, introduces NPCs you meet.

```
I try to persuade the innkeeper to give me a discount. I roll to persuade.
```
Claude rolls the persuasion check, tells you if you succeed, and narrates the result.

```
My character attacks the bandit with my sword.
```
Claude rolls an attack check, rolls damage if you hit, and narrates the combat.

```
What's on my character sheet right now?
```
Claude reads and describes your character — stats, HP, inventory, everything.

```
Update my HP to 18. I took some damage.
```
Claude updates your health on the character sheet.

```
I want to create a world note for the tavern owner we just met. He's a grumpy dwarf named Bort who makes incredible ale.
```
Claude saves this NPC to your world notes.

```
Start a combat encounter. We're facing two bandits and a bandit captain.
```
Claude sets up combat, shows the combat panel in your dashboard, and asks what turn order you want.

### What Claude Understands

- Claude knows what campaign and character are active (you don't have to say their names each time)
- Claude reads your character sheet automatically
- Claude knows the ruleset you're playing (Ironsworn, D&D, etc.)
- Claude understands context from the entire conversation — if you mentioned a person three messages ago, Claude still remembers them

### If Something Goes Wrong

Just tell Claude:

```
Actually, I wanted to roll 2d6, not 1d6. Let me redo that.
```

Claude will reroll.

```
I don't think that's how this ability works. Can you explain what my Edge stat does in Ironsworn?
```

Claude will explain.

**There's no penalty for asking Claude to redo something or clarify the rules.** The story is collaborative.

---

## Understanding the Dashboard

Your browser dashboard at localhost:7432 is always open and always updating. Here's what each part does:

### Header (Top)

Shows your active campaign name, your active character (with portrait if you've uploaded one), and the current session title. Updates instantly when you change which campaign or character is active.

### Session Log

The transcript of your adventure in reading order. Shows what you said (marked as "user") and what Claude narrated (marked as "assistant"). This is your story in text form.

### Session Timeline

A unified chronological view of everything that happened: narrative messages, dice rolls, world note creations, NPCs defeated, and combat events. Useful if you want to see what happened in order without parsing the chat.

### Combat Panel

Only visible when a combat encounter is active. Shows a card for each combatant (you and enemies) with:
- Name and initiative (who goes first)
- HP bar (green when healthy, yellow when hurt, red when critical)
- Current HP and max HP
- Any conditions (poisoned, stunned, unconscious, etc.)
- A marker showing whose turn it is

### Dice History

Every dice roll from this session with the total result and the individual dice breakdown. If you rolled 2d6+3, you see both the individual die results and the final sum.

### Journal Panel

A text area for the session summary. You can write it yourself, or click "Generate recap" to have Claude AI write a summary of everything that happened.

### World Notes

Your campaign encyclopedia. Every NPC, location, faction, monster, and item gets an entry. Features:
- Search by name or content
- Filter by category (NPC, location, faction, item, other)
- Create new notes manually or draft them with AI
- Click any note to read or edit it

### Map Panel

Upload a campaign map image (JPG, PNG, GIF, WebP). Drop pins on locations. Click a pin to see what that location is (the label and notes you wrote about it).

### Character Sheet

Your character's full sheet based on the ruleset. For Ironsworn, you see Edge, Heart, Iron, Shadow, Wits, health, spirit, supply, momentum, vows, bonds, and assets. For D&D you see attributes, HP, AC, skills, inventory, and spells. All fields are editable directly in the browser — change a number and it auto-saves.

Click the portrait area to upload a character portrait image.

---

## Playing Your First Session: A Walkthrough

Let's walk through a complete Ironsworn session from creation to recap.

### Create Campaign and Character

You've already done this, but here's what it looks like:

```
User: Create an Ironsworn campaign called "The Ironlands" and create a character named Kael.

Claude: I've created your campaign and character. Kael is an Ironsworn warrior ready for adventure — stats rolled for ironsworn.
```

Check the dashboard. You see:
- Campaign name: "The Ironlands"
- Character name: "Kael"
- Character sheet with Kael's stats already filled in (randomly rolled for the ruleset)

### Start a Session

```
User: Start a session called "Whispers in the Mist" for today.

Claude: Session started. We're ready to begin your adventure.
```

The dashboard now shows your session title.

### Narrate the Opening

```
User: Set the scene for us. Kael is traveling through the forest looking for a rumored shrine.

Claude: The forest stretches endlessly in all directions. Mist coils between ancient oaks, muffling your footsteps. You've been searching for three days now, following only rumors and the cryptic map an old sage gave you. As dusk falls, you spot something ahead—carved stone poking through moss and vines. This might be it.

What do you do?
```

This text appears in your Session Log in the dashboard. Everything is being saved.

### Attempt Something and Roll

```
User: I approach carefully and look for any signs of danger. I'll roll for perception.

Claude: A tense moment. You creep forward, examining the clearing around the shrine. Let me roll for your perception check.

I roll: 2d6 = 1, 5 = 6

Not great. You notice the shrine is old and covered in moss, but you miss the tripwire stretched across the entrance at ankle height. As you step forward, it snaps taut.
```

The Dice History panel now shows your 2d6 roll with the individual die results (1 and 5) and the sum (6).

### Get Into Combat

```
User: Oh no. What happens?

Claude: A carved stone door creaks open. Two stone guardians step out, eyes glowing red. Combat begins!

Kael, both guards. Initiative—you roll 2d6+Wits. You have Wits 2, so that's 2d6+2.

Claude rolls: 2d6 = 6, 4 = 10 + 2 = 12
Guard 1: 8
Guard 2: 6

You go first.

User: I draw my sword and slash at the nearest guard. Rolling to attack.

Claude: You strike at Guard 1. Rolling 2d6+Iron (you have Iron 3).

I roll: 2d6 = 5, 3 = 8 + 3 = 11

That hits! The sword bites deep. Guard 1 takes 5 damage (1d6 roll was 5). Guard 1 is hurt.

Guard 1 attacks you back...
```

The Combat Panel appears in the dashboard showing:
- Kael (your character)
- Guard 1 (HP 5/10)
- Guard 2 (HP 10/10)
- Turn indicator: Kael

As each update happens, the HP bars update in real time. When Guard 2 attacks you and deals damage, your HP drops in the dashboard.

### End Combat

```
User: After three more rounds, I finally fell both guards. They crumble to stone dust.

Claude: With a final slash, Guard 2 shatters into fragments. Your sword is dusty but intact. Kael, you stand alone in front of the shrine door.

You're at HP 3 (took 7 damage total). Catch your breath?

User: End combat. I need to rest before entering.

Claude: Combat ended. The doors to the shrine are now open.
```

The Combat Panel disappears from the dashboard.

### End the Session and Recap

```
User: I'll stop here and get a recap of this session.

Claude: I'll generate a session recap for you.
```

Claude reads all your messages, all the dice rolls, and the entire narrative from this session. It writes:

```
Session Recap: "Whispers in the Mist"

Kael ventured into the mist-shrouded forest seeking a legendary shrine. After three days of travel, you found it hidden among ancient oaks. Your perception check revealed only the shrine's outline, missing a hidden tripwire in your haste. Two stone guardians emerged from a sealed door and combat erupted. Through superior swordsmanship (aided by your high Iron attribute), you defeated both guardians despite taking significant damage. The shrine door now stands open, its secrets waiting within.
```

This appears in the Journal Panel of your dashboard. You can edit it or leave it as-is.

That's a full session: creation, play, combat, and recap. Everything was saved. You can start a new session tomorrow and the world notes, your character's stats, and all memories of this adventure stay.

---

## Things You Can Do (Complete Feature List)

### Managing Characters

**Create a new character:**

```
Create a character named Lyra for this campaign.
```

**Update character stats:**

```
My HP is now 35. Add "Healing Potion" to my inventory. Remove "Broken Sword".
```

**Upload a portrait:**

Open the dashboard, go to the character sheet panel, and click the portrait area. Select a JPG, PNG, GIF, or WebP file (up to 5 MB). Your character's face appears in the header and sheet.

**Switch between characters:**

```
List my characters. Set character 2 as active.
```

You can have multiple characters in one campaign and play different ones in different sessions.

### Sessions and Play

**Start a session:**

```
Start a session called "Expedition to the Lost City" for today.
```

**End a session:**

```
End the session. Write a summary of what happened.
```

**Generate an AI recap:**

```
Generate a session recap.
```

Claude reads the entire session transcript and writes a summary. Requires an API key.

### Combat

**Start combat:**

```
Combat begins. We face three goblin archers and a goblin leader. All three archers have 7 HP. The leader has 20 HP.
```

**Update combatant health and conditions:**

```
Archer 1 takes 8 damage. The leader is now poisoned and stunned.
```

**End combat:**

```
We defeat the goblins. End combat.
```

### Dice Rolling

Roll any expression Claude understands:

```
Roll d20 for the save.
Roll 4d6 drop lowest for stats.
Roll 2d6+3 for the check.
Roll d100 for percentile.
Roll 3d8 for damage.
```

Every roll is saved. The Dice History shows all results.

### World Notes (Your Campaign Wiki)

**Create a note:**

```
Create a world note for the Red Dragon we just fought. It's guarding a hoard of gold. Category: Monster.
```

**Draft a note with AI:**

In the dashboard, go to World Notes, click "Draft with AI", enter a hint like "mysterious sage who speaks in riddles", and click Generate. Claude drafts a complete note.

**Update a note:**

```
Update the Red Dragon note. We learned it's actually the ancient guardian of the kingdom.
```

**Search notes:**

```
Search world notes for "dragon".
Search world notes in the NPC category.
```

### Maps and Pins

**Upload a map:**

In the dashboard's Map Panel, click "Upload Map" and choose a JPG, PNG, GIF, or WebP file (up to 10 MB).

**Add pins to the map:**

```
Add a pin at coordinates 0.25, 0.5 (left-center) labeled "The Tavern" with the note "Mira's base". Use color #e67e22.
```

Pins use fractional coordinates where (0.0, 0.0) is the top-left corner and (1.0, 1.0) is the bottom-right. Click any pin in the dashboard to see its label and notes.

### Rules Lookups

**Upload a rulebook:**

If you have the official rulebook as a PDF or plain text, you can upload it so Claude can search it during play:

```bash
curl -X POST http://localhost:7432/api/rulesets/1/rulebook \
  -H "Content-Type: text/plain" \
  --data-binary @my-rulebook.txt
```

Or for a PDF:

```bash
curl -X POST http://localhost:7432/api/rulesets/1/rulebook \
  -F "rulebook=@my-rulebook.pdf"
```

**Look up a rule during play:**

```
Look up the grappling rules for me.
What does the Dodge ability do in Ironsworn?
```

Claude searches your uploaded rulebook and answers based on the official rules.

### Managing Campaigns

**Create multiple campaigns:**

```
Create a Cyberpunk Red campaign called "Night City Shadows".
```

**List and switch campaigns:**

```
List my campaigns.
Set campaign 3 as active.
```

You can run multiple campaigns and switch between them. Each has its own characters, sessions, and notes.

---

## For Developers

This section is for people who want to extend ink & bone or integrate it with other tools. If you're just playing, skip this section.

### MCP Tool Reference

All tools are available to Claude Code once the MCP server is registered.

#### Context
| Tool | Parameters | Returns |
|---|---|---|
| `get_context` | — | Full game state snapshot: active campaign, character, session, last 20 messages, active combat |

#### Campaigns & Sessions
| Tool | Required | Optional | Returns |
|---|---|---|---|
| `create_campaign` | `ruleset` (string), `name` (string) | `description` | Campaign created, activated |
| `list_campaigns` | — | — | JSON array of campaigns |
| `set_active` | At least one of `campaign_id`, `session_id`, `character_id` | — | Confirmation |
| `start_session` | `title` (string), `date` (YYYY-MM-DD) | `narrative` | Session created, activated |
| `end_session` | `summary` (string) | `narrative` | Session closed |
| `list_sessions` | — | `campaign_id` | JSON array of sessions |

#### Characters
| Tool | Required | Optional | Returns |
|---|---|---|---|
| `create_character` | `name` (string) | `campaign_id` | Character created, activated, stats auto-rolled for the campaign's ruleset |
| `list_characters` | — | `campaign_id` | JSON array of characters |
| `get_character_sheet` | — | `character_id` | Full character JSON |
| `update_character` | `updates` (JSON object as string, e.g. `{"hp":15}`) | `character_id`, `narrative` | Confirmation |
| `add_item` | `item_name` (string) | `character_id`, `narrative` | Item appended to inventory |
| `remove_item` | `item_name` (string) | `character_id`, `narrative` | Item removed from inventory |

The `updates` parameter to `update_character` is a JSON object string with any keys from the ruleset's schema. Only the specified keys are updated; all other field values are preserved.

#### Combat
| Tool | Required | Optional | Returns |
|---|---|---|---|
| `start_combat` | `name` (string), `combatants` (JSON array string) | `narrative` | Encounter created |
| `update_combatant` | `combatant_id` (number), `hp_current` (number) | `conditions` (JSON array string), `narrative` | Combatant updated |
| `end_combat` | — | `narrative` | Encounter closed |

**Combatants format:**

```json
[
  {"name": "Kael",   "initiative": 18, "hp_max": 52, "is_player": true},
  {"name": "Bandit",  "initiative": 12, "hp_max": 11, "is_player": false},
  {"name": "Captain", "initiative": 15, "hp_max": 65, "is_player": false}
]
```

**Conditions format:**

```json
["poisoned", "prone"]
```

#### World Notes
| Tool | Required | Optional | Returns |
|---|---|---|---|
| `create_world_note` | `title`, `content`, `category` (npc/location/faction/item) | `narrative` | Note created |
| `update_world_note` | `note_id` (number), `title`, `content` | `tags` (JSON array string), `narrative` | Note updated |
| `search_world_notes` | — | `query` (text), `category` | Matching notes |

#### Dice
| Tool | Required | Optional | Returns |
|---|---|---|---|
| `roll_dice` | `expression` (string, e.g. `"2d6+3"`) | `narrative` | Result string with total and breakdown |

Requires an active session.

#### Maps
| Tool | Required | Optional | Returns |
|---|---|---|---|
| `add_map_pin` | `map_id` (number), `x` (float), `y` (float), `label` (string) | `note`, `color` (hex) | Pin created |

`x` and `y` are fractional coordinates from 0.0 (top-left) to 1.0 (bottom-right).

#### AI Tools
| Tool | Required | Optional | Notes |
|---|---|---|---|
| `generate_session_recap` | — | `session_id` | Requires `ANTHROPIC_API_KEY`. Reads session messages and dice rolls, generates summary, saves it. |
| `search_rulebook` | `query` (string) | `ruleset_id` | Returns up to 3 matching rulebook chunks by heading or content. |

### HTTP API Reference

The API is available at `http://localhost:7432`. All JSON request bodies use `Content-Type: application/json`.

#### Health
```
GET /api/health
→ { "status": "ok", "ai_enabled": true }
```

#### Campaigns
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

POST /api/campaigns/{id}/world-notes/draft
Content-Type: application/json
Body: { "hint": "mysterious fence in the docks district" }
→ 201 { "id": N, "title": "...", "content": "..." }
```

#### Sessions
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

POST /api/sessions/{id}/recap
→ { "summary": "..." }
```

#### Maps & Pins
```
GET /api/maps/{id}
→ CampaignMap

GET /api/maps/{id}/pins
→ []MapPin
```

#### World Notes
```
PATCH /api/world-notes/{id}
Body: { "title": "...", "content": "...", "tags_json": "[\"tag1\"]" }
→ 204 No Content
```

#### Characters
```
PATCH /api/characters/{id}
Body: { "data_json": "{\"hp\":35,\"level\":4}" }
→ 204 No Content

POST /api/characters/{id}/portrait
Content-Type: multipart/form-data
Fields: portrait (file, max 5 MB, jpg/png/gif/webp)
→ { "portrait_path": "portraits/4_filename.jpg" }
```

#### Rulesets & Rulebook
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

#### Game Context
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

#### Static Files
```
GET /api/files/{path}
→ File contents from ~/.ttrpg/{path}
```

Used to serve portrait images and map images. Path traversal outside `~/.ttrpg/` is blocked.

### WebSocket Events

Connect to `ws://localhost:7432/ws`. The server broadcasts JSON events on every state change.

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

### File Storage

All uploaded files are stored under `~/.ttrpg/`:

| Type | Directory | Naming | Max size |
|---|---|---|---|
| Portraits | `~/.ttrpg/portraits/` | `{character_id}_{original_filename}` | 5 MB |
| Maps | `~/.ttrpg/maps/` | `{32-char hex}{extension}` | 10 MB |

Accepted formats for both: `.jpg`, `.jpeg`, `.png`, `.gif`, `.webp`

Files are served at `GET /api/files/portraits/{filename}` and `GET /api/files/maps/{filename}`.

The database file lives at `~/.ttrpg/ttrpg.db`. Back this up to preserve all campaign data.

### Adding a Custom Ruleset

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

### Development

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
