# XP Advancement System Design

## Overview

After every GM response, a goroutine (`autoSuggestXPSpend`) checks whether the active character earned any XP (or equivalent currency) this turn. If so, it calls Claude to generate 2–3 ranked advancement suggestions, then pushes them to the frontend as a `xp_spend_suggestions` WebSocket event. The player sees a dismissible overlay showing each option with a **Spend** button. Clicking Spend calls `POST /api/characters/{id}/advance` — the server validates, recalculates derived stats, and broadcasts `character_updated`.

---

## Section 1 — Architecture

Three new pieces:

1. **`autoSuggestXPSpend` goroutine** — fires from inside `autoUpdateCharacterStats`, after the XP patch is applied, only when the XP field increased.
2. **`POST /api/characters/{id}/advance` endpoint** — validates affordability, applies the advance, recalculates W&G derived stats.
3. **`xp_spend_suggestions` WebSocket event** — pushes suggestion payload to the client; `XPSuggestionsPanel` renders the overlay.

No polling. No cron. The goroutine only runs when XP actually changes.

---

## Section 2 — Per-Ruleset Advancement Data

Lives in `internal/ruleset/advancement.go`. Exported function:

```go
func CanAffordAdvancement(system, field string, currentVal int, currentXP int) bool
func XPCostFor(system, field string, newVal int) int
func ValidFields(system string) []string
```

### Wrath & Glory (`wrath_glory`)

XP key: `xp`

**Attribute advance:** cost = `new_rating × 4`  
Attributes: `str`, `agi`, `tgh`, `int`, `wil`, `fel`, `ini` (index order matches `[7]int` in `random_stats.go`: 0=STR 1=AGI 2=TGH 3=INT 4=WIL 5=FEL 6=INI)

**Skill advance:** cost = `new_rating × 4`  
Skills: `ws`, `bs`, `athletics`, `awareness`, `cunning`, `deception`, `fortitude`, `insight`, `intimidation`, `investigation`, `leadership`, `medicae`, `persuasion`, `pilot`, `psychic_mastery`, `scholar`, `stealth`, `survival`, `tech`

**Talents:** cost = per-talent XP value (hardcoded table in `wg_talents.go`, range 10–60, default 20)  
Talent advance field: `talent:<TalentName>` — server checks character doesn't already own it and has no archetype ability with that name.

**Derived stats recalculated after attribute advance:**
- `tgh` → `wounds = (tier×2) + tgh`, `resilience = tgh + 1`, `determination = tgh`
- `wil` → `shock = wil + tier`, `resolve = max(1, wil-1)`, `conviction = wil`
- `fel` → `influence = fel - 1`
- `ini` → `defence = ini - 1`

### D&D 5e (`dnd5e`)

XP key: `xp`

Simplified level-up only. No individual stat purchases.

**Level advance:** triggers when `xp ≥ threshold[level+1]`.  
XP thresholds (standard 5e): 0, 300, 900, 2700, 6500, 14000, 23000, 34000, 48000, 64000, 85000, 100000, 120000, 140000, 165000, 195000, 225000, 265000, 305000, 355000.  
On level-up: `level += 1`, `hp += 5` (average flat increase), `proficiency_bonus = floor((level-1)/4)+2`.  
Field: `level`.

### Vampire: The Masquerade (`vtm`)

XP key: `xp`

Cost table (new dots × multiplier, minimum 1 XP per dot):

| Trait | Cost per new dot |
|-------|-----------------|
| Attribute | new_dots × 4 |
| Skill | new_dots × 3 |
| Discipline (in-clan) | new_dots × 5 |
| Discipline (out-of-clan) | new_dots × 7 |
| Blood Potency | new_dots × 10 |

Fields: `str`, `dex`, `sta`, `cha`, `man`, `com`, `int`, `wits`, `res` (attributes); `athletics`, `brawl`, `craft`, `drive`, `firearms`, `larceny`, `melee`, `stealth`, `survival`, `animal_ken`, `etiquette`, `insight`, `intimidation`, `leadership`, `performance`, `persuasion`, `streetwise`, `subterfuge`, `academics`, `awareness`, `finance`, `investigation`, `medicine`, `occult`, `politics`, `science`, `technology` (skills); `blood_potency`; `discipline:<name>` (Oblivion, Dominate, Presence, etc. — in-clan vs out-of-clan determined by character's clan stored in stats JSON).

### Call of Cthulhu (`coc`)

No XP advancement system. Skill improvement is post-session dice-roll based, not XP. **Skip** — `autoSuggestXPSpend` no-ops for `coc`.

### Cyberpunk Red (`cyberpunk_red`)

XP key: `xp`

Improvement Points (IP) system. Cost = 10 × new_rating for skills, flat 30 for role ability rank.

Fields: all skills in stats JSON; `role_ability` (single rank advance = 30 IP).

### Shadowrun (`shadowrun`)

XP key: `karma`

Cost table:

| Trait | Karma cost |
|-------|-----------|
| Attribute | new_rating × 5 |
| Active skill | new_rating × 5 |
| Knowledge skill | new_rating × 2 |
| Specialization | 5 |
| Quality | varies (hardcode common ones at 10 each) |

Fields: attributes, active skills, knowledge skills.

### WFRP (`wfrp`)

XP key: `xp`

Advances cost 10 XP each (characteristic, skill, or talent — flat).  
Fields: all WFRP characteristics and skills in stats JSON; talent advances add 1 talent purchase.

### Star Wars EotE (`starwars`)

XP key: `xp`

Skills only (talent trees are hierarchical purchase sequences — too complex for auto-suggest; omit).

Cost: skill advance = new_rating × 5 (Force-sensitive career skills × 5, non-career × 5 — treat all as career for simplicity).

Fields: all skills in stats JSON.

### Legend of the Five Rings (`l5r`)

XP key: `xp`

Cost table (rings and skills):

| Trait | Cost |
|-------|------|
| Ring advance | new_rank × 3 |
| Skill advance | new_rank × 2 |

Fields: `air`, `earth`, `fire`, `water`, `void` (rings); all skills in stats JSON.

### The One Ring (`theonering`)

XP key: `xp`

Advancement Points (AP) — skill advances cost new_rank × 1 (each rank is 1 AP). Attribute advances not available via XP.

Fields: all skills in stats JSON.

### Blades in the Dark (`blades`)

XP key: `xp`

XP tracks fill per playbook (typically 8 boxes). On spend, track resets to 0 (not subtracted by 8) and character gains 1 advancement (new action dot or special ability). Server sets `xp = 0` on advance.

Fields: `action:<ActionName>` for action rating advances; `ability:<AbilityName>` for special abilities.

Threshold: `xp >= 8`.

### Ironsworn (`ironsworn`)

XP key: `xp`

Assets cost 2 XP each, asset upgrades cost 1 XP. No attribute stats to advance.

Fields: `asset:<AssetName>`.

### Paranoia (`paranoia`)

No persistent XP advancement. **Skip** — `autoSuggestXPSpend` no-ops for `paranoia`.

---

## Section 3 — `autoSuggestXPSpend` Goroutine

**Trigger:** Called from inside `autoUpdateCharacterStats` after the XP patch, only if the XP field value increased.

**Gate:** `CanAffordAdvancement()` pure-Go check before any AI call. If the character cannot afford any advancement (checked across all valid fields for the system), goroutine exits silently.

**AI call:** Provide Claude with:
- Character name, archetype, tier, faction (W&G), current stats
- Current XP balance
- System cost table (as prose or structured)
- Already-owned talents (pipe-delimited string for exact match)
- Archetype starting abilities (to exclude from suggestions)

Ask for 2–3 ranked suggestions. Each suggestion: `field` (e.g. `"tgh"`, `"talent:Iron Will"`, `"level"`), `display_name`, `current_value`, `new_value`, `xp_cost`, `reasoning` (1 sentence).

**Cap:** Maximum 20 suggestions per character per session (tracked in memory by session, reset on new session).

**Output:** Push `xp_spend_suggestions` WebSocket event.

---

## Section 4 — `POST /api/characters/{id}/advance` Endpoint

**Request body:**
```json
{
  "field": "tgh",
  "new_value": 5
}
```
For talent advances: `"field": "talent:Iron Will"`, `"new_value": 1`.

**Server validation:**
1. Look up character and campaign (to get system).
2. Fetch current stats JSON.
3. For attributes/skills: assert `new_value == current_value + 1` — no skipping ranks.
4. For talents: assert character does not already own talent (pipe-delimited exact match), assert talent exists in `wg_talents.go` table (W&G only).
5. Calculate `cost = XPCostFor(system, field, new_value)`.
6. Assert `current_xp >= cost`.
7. Apply: set `field = new_value`, set `xp -= cost`.
8. W&G: recalculate affected derived stats (see Section 2).
9. Blades: set `xp = 0` (full reset).
10. Persist character stats JSON.
11. Broadcast `character_updated` WebSocket event.

**Response:** `200 OK` with updated character JSON, or `400` with error message.

---

## Section 5 — WebSocket Event

Event type: `xp_spend_suggestions`

Payload:
```json
{
  "character_id": 7,
  "character_name": "Brother Cato",
  "current_xp": 14,
  "xp_label": "XP",
  "suggestions": [
    {
      "field": "tgh",
      "display_name": "Toughness",
      "current_value": 4,
      "new_value": 5,
      "xp_cost": 20,
      "reasoning": "Your survivability is your greatest asset — boosting TGH also raises wounds, resilience, and determination."
    },
    {
      "field": "talent:Iron Will",
      "display_name": "Iron Will (Talent)",
      "current_value": 0,
      "new_value": 1,
      "xp_cost": 20,
      "reasoning": "Directly reinforces your role as a frontline fighter."
    }
  ]
}
```

`xp_label` is system-specific: `"XP"`, `"Karma"`, `"IP"`, `"AP"` as appropriate.

---

## Section 6 — `XPSuggestionsPanel` Frontend Component

Fixed-position overlay, bottom-right corner, above audio controls. Appears when a `xp_spend_suggestions` event arrives, dismissed by clicking X or after any Spend action.

Layout per suggestion card:
- Bold display name + current → new value arrow
- XP cost badge (display only, never editable)
- Reasoning text in muted style
- **Spend** button

On Spend click: POST to `/api/characters/{id}/advance`, show brief green flash on success, dismiss panel. On error: show red toast.

If the panel is already visible when a new suggestion event arrives, replace the suggestions (don't stack).

---

## Section 7 — W&G Archetype Starting Abilities

`wgArchetypeDef` gains an `abilities []string` field. At character creation, `rollWrathGloryStats()` pre-populates the `talents` stat field with the pipe-delimited ability names. The advance endpoint and suggestion goroutine exclude these from talent suggestions.

Complete mapping (core rulebook, sourced from doctors-of-doom.com API):

| Archetype | Tier | Starting Abilities |
|-----------|------|--------------------|
| Ministorum Priest | 1 | Fiery Invective |
| Sister Hospitaller | 1 | Loyal Compassion |
| Imperial Guardsman | 1 | Look Out, Sir! |
| Inquisitorial Acolyte | 1 | Inquisitorial Decree |
| Inquisitorial Sage | 1 | Administratum Records |
| Ganger | 1 | Scrounger |
| Cultist | 1 | Enemy Within, Corruption |
| Corsair | 1 | Dancing the Blade's Edge |
| Boy | 1 | Get Stuck In |
| Death Cult Assassin | 2 | Glancing Blow |
| Sister of Battle | 2 | Purity of Faith |
| Tempestus Scion | 2 | Elite Soldier |
| Rogue Trader | 2 | Warrant of Trade |
| Sanctioned Psyker | 2 | Psyker, Unlock Disciplines |
| Scavvy | 2 | Mutant |
| Rogue Psyker | 2 | Psyker, Unlock Disciplines, Corruption |
| Ranger | 2 | From the Shadows |
| Kommando | 2 | Kunnin' Plan |
| Space Marine Scout | 2 | Use the Terrain |
| Skitarius | 2 | Heavily Augmented |
| Crusader | 3 | Armour of Faith |
| Imperial Commissar | 3 | Fearsome Respect |
| Tech-Priest | 3 | Rite of Repair |
| Desperado | 3 | Valuable Prey |
| Heretek | 3 | Rite of Repair, Corruption |
| Tactical Space Marine | 3 | Tactical Versatility |
| Chaos Space Marine | 3 | Tactical Versatility, Corruption |
| Warlock | 3 | Runes of Battle, Unlock Disciplines |
| Nob | 3 | The Green Tide |
| Inquisitor | 4 | Unchecked Authority |
| Primaris Intercessor | 4 | Intercessor Focus |

`abilities` slice values must exactly match the talent name strings used elsewhere (same capitalization, same spelling). These are excluded from both suggestions and from the advance endpoint's "already owned" check — they are free starting gifts, not purchased advances.

---

## New Files

- `internal/ruleset/advancement.go` — `CanAffordAdvancement`, `XPCostFor`, `ValidFields`, `XPLabel`
- `internal/ruleset/wg_talents.go` — hardcoded W&G talent table (236 talents with XP costs; already planned)
- `web/src/components/XPSuggestionsPanel.tsx` — overlay component

## Modified Files

- `internal/ruleset/random_stats.go` — add `abilities []string` to `wgArchetypeDef`, populate `wgArchetypes`, set talents field at creation
- `internal/api/automation.go` — add `autoSuggestXPSpend`, called from `autoUpdateCharacterStats`
- `internal/api/routes.go` — register `POST /api/characters/{id}/advance`
- `internal/api/handlers.go` (or new `advance.go`) — `handleAdvanceCharacter`
- `web/src/App.tsx` (or WebSocket hook) — handle `xp_spend_suggestions` event, pass to panel
- `internal/db/db.go` — no new migration needed (uses existing `characters.stats` JSON column)
