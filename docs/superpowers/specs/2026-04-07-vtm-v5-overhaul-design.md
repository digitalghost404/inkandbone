# VtM V5 Full Overhaul Design

**Date:** 2026-04-07  
**Ruleset:** Vampire: The Masquerade 5th Edition (2021)  
**Scope:** All inkandbone systems updated for VtM — schema, character creation, dice mechanics, combat, frontend, automation, audio, and oracle tables. No other rulesets are touched.

---

## 1. Character Schema (Migration 024)

A new migration updates the VtM ruleset's `schema` JSON and documents the V5 fields. No other rulesets are modified.

### Core vampire stats
| Field | Type | Default | Notes |
|-------|------|---------|-------|
| `hunger` | INT | 1 | 0–5; replaces blood_pool |
| `blood_potency` | INT | 1 | 0–10 |
| `bane_severity` | INT | 1 | Derived from BP tier (1–3) |
| `humanity` | INT | 7 | 0–10 |
| `stains` | INT | 0 | 0–10; reset after Remorse check |

### Attributes (9 fields, each INT 1–5)
- **Physical:** `strength`, `dexterity`, `stamina`
- **Social:** `charisma`, `manipulation`, `composure`
- **Mental:** `intelligence`, `wits`, `resolve`

### Skills (27 fields, each INT 0–5)
- **Physical:** `athletics`, `brawl`, `craft`, `drive`, `firearms`, `larceny`, `melee`, `stealth`, `survival`
- **Social:** `animal_ken`, `etiquette`, `insight`, `intimidation`, `leadership`, `performance`, `persuasion`, `streetwise`, `subterfuge`
- **Mental:** `academics`, `awareness`, `finance`, `investigation`, `medicine`, `occult`, `politics`, `technology`

### Disciplines (11 fields, each INT 0–5)
`animalism`, `auspex`, `blood_sorcery`, `celerity`, `dominate`, `fortitude`, `obfuscate`, `oblivion`, `potence`, `presence`, `protean`

### Derived / damage tracks
| Field | Type | Default |
|-------|------|---------|
| `health_max` | INT | stamina + 3 |
| `health_superficial` | INT | 0 |
| `health_aggravated` | INT | 0 |
| `willpower_max` | INT | composure + resolve |
| `willpower_superficial` | INT | 0 |
| `willpower_aggravated` | INT | 0 |

### Identity fields
| Field | Type |
|-------|------|
| `predator_type` | TEXT |
| `sect` | TEXT |
| `convictions` | TEXT (comma-separated) |
| `touchstones` | TEXT (comma-separated) |
| `ambition` | TEXT |
| `desire` | TEXT |

---

## 2. GM Context (Migration 025)

Rewrites VtM's `gm_context` to be fully V5-accurate. Injected into every VtM GM prompt.

### Vocabulary directives
- Use **Hunger** (not blood pool), **Rouse Check** (not blood spending)
- Use **Superficial / Aggravated** damage (not normal/lethal/aggravated)
- Use **Blood Potency** as the power metric (not Generation)
- Use **Convictions + Touchstones** (not Virtues / generic Humanity checks)

### Hunger die narration
- **Bestial Failure** (Hunger die = 1, overall failure): animalistic loss of control — snarling, snapping, involuntary frenzy edge
- **Messy Critical** (Hunger die = 10, overall critical success): success achieved in a savage, uncontrolled way — excessive force, blood spray, witnesses horrified

### Rouse Check narration
- Hunger increase: gnawing emptiness, the Beast pressing forward, predator instincts sharpening
- Hunger 5: character is one step from frenzy at all times

### Frenzy thresholds
| Type | Trigger | Resist |
|------|---------|--------|
| Hunger Frenzy | Hunger 5 + provocation (blood smell, denial) | Composure + Resolve |
| Terror Frenzy | Supernatural fear source | Composure + Resolve |
| Rage Frenzy | Humiliation or violence against loved ones | Composure + Resolve |

### Masquerade vocabulary
- Minor breach (overheard conversation): 1 point
- Moderate breach (witnessed feeding): 2 points
- Major breach (supernatural display on camera): 3 points
- Capitalize "the Masquerade" as a proper noun

### Clan Compulsions
| Clan | Compulsion |
|------|-----------|
| Brujah | Rebellion — must defy authority this scene |
| Gangrel | Feral Impulse — animal behavior overtakes social norms |
| Malkavian | Delusion — believes something demonstrably false |
| Nosferatu | Cryptophilia — must learn or hoard a secret |
| Toreador | Obsession — fixated on nearby beauty, cannot focus elsewhere |
| Tremere | Perfectionism — must redo any imperfect action |
| Ventrue | Arrogance — refuses to accept help from social inferiors |

---

## 3. Character Creation

### options.go — VtM dropdowns
- **clan:** Brujah, Gangrel, Malkavian, Nosferatu, Toreador, Tremere, Ventrue, Caitiff, Thin-Blooded
- **predator_type:** Alleycat, Bagger, Blood Leech, Cleaner, Consensualist, Extortionist, Graverobber, Osiris, Sandman, Siren
- **sect:** Camarilla, Anarch, Unaligned, Sabbat (lapsed)
- **generation:** 10th, 11th, 12th, 13th, 14th, 15th (Thin-Blooded)

### random_stats.go — V5 starting stats
- One attribute group primary (4+3+2+1), one secondary (3+3+2+1), one tertiary (2+2+2+1) — distributed randomly across the 9 stats
- `health_max` = stamina + 3, `willpower_max` = composure + resolve
- `hunger` = 1, `blood_potency` = 1, `bane_severity` = 1, `humanity` = 7, `stains` = 0
- All disciplines = 0 (Predator Type logic applies them)

### Predator Type auto-population (`internal/ruleset/vtm_predator_types.go`)
Each of the 10 types encodes exact V5 rulebook grants:
- 2 Discipline dots (+1 each to two disciplines)
- 1 skill specialty (written to a `skill_specialties` text field)
- Merits/Flaws (written to a `merits_flaws` text field)

Apply function fires when `predator_type` is set during character creation.

| Predator Type | Disciplines | Skill Specialty | Notable Merit/Flaw |
|---------------|-------------|-----------------|-------------------|
| Alleycat | Celerity +1, Potence +1 | Brawl (Grappling) | Prowler's Instinct / Prey Exclusion (Homeless) |
| Bagger | Obfuscate +1, Auspex +1 | Streetwise (Black Market) | Iron Gullet / Prey Exclusion (Bagged) |
| Blood Leech | Obfuscate +1, Potence +1 | Stealth (Shadowing) | No Merit / Shunned, Prey Exclusion (Mortals) |
| Cleaner | Obfuscate +1, Dominate +1 | Subterfuge (Impersonation) | Retainer / Obvious Predator |
| Consensualist | Presence +1, Auspex +1 | Persuasion (Victim Calming) | Herd / Prey Exclusion (Non-consenting) |
| Extortionist | Dominate +1, Presence +1 | Intimidation (Blackmail) | Contacts / Prey Exclusion (Vulnerable) |
| Graverobber | Fortitude +1, Oblivion +1 | Medicine (Cadavers) | No Merit / Obvious Predator |
| Osiris | Blood Sorcery +1, Presence +1 | Occult (specific tradition) | Fame / Prey Exclusion (Faithful) |
| Sandman | Auspex +1, Obfuscate +1 | Stealth (Breaking and Entering) | — / Prey Exclusion (Sleeping) |
| Siren | Presence +1, Potence +1 | Persuasion (Seduction) | Looks (Beautiful) / Prey Exclusion (Mortals in relationships) |

---

## 4. Hunger & Rouse Dice System

### Rouse Check command (`/rouse`)
- `checkAndExecuteRoll` intercepts `/rouse` before GM responds
- Rolls 1d10: 6–10 = no Hunger change; 1–5 = Hunger +1 (persisted, `character_updated` broadcast)
- At Hunger 5: Rouse Check instead triggers a Hunger Frenzy resist roll (Composure + Resolve, difficulty 3)

### Hunger dice in pools
For any declared dice roll:
- `hunger_dice` = min(hunger, pool)
- `regular_dice` = pool − hunger_dice
- **Bestial Failure**: any Hunger die = 1 AND overall failure → flag in GM context
- **Messy Critical**: any Hunger die = 10 AND overall critical (two 10s) → flag in GM context; auto-roll clan Compulsion table

### Blood Surge (`/surge`)
- Player types `/surge` before a roll → one Rouse Check fires
- On success (Hunger unchanged): add Blood Potency bonus dice to the pool for that turn
- BP bonus die table: BP 1–3 → +1 die; BP 4–6 → +2 dice; BP 7–9 → +3 dice; BP 10 → +4 dice

### Stains & Remorse
- `autoUpdateCharacterStats` detects Humanity-violating acts (feeding, killing, diablerie, compulsion, breaking conviction) → `stains` +1 (capped at 10)
- At session end: if `stains > 0`, trigger Remorse Check (roll 10 − humanity dice, need one 6+); failure = `humanity` −1; then `stains` = 0

---

## 5. Combat & Damage System

### Combatants table additions (new migration)
```sql
ALTER TABLE combatants ADD COLUMN damage_superficial INT DEFAULT 0;
ALTER TABLE combatants ADD COLUMN damage_aggravated INT DEFAULT 0;
ALTER TABLE combatants ADD COLUMN willpower_superficial INT DEFAULT 0;
ALTER TABLE combatants ADD COLUMN willpower_aggravated INT DEFAULT 0;
ALTER TABLE combatants ADD COLUMN hunger INT DEFAULT 0;
```

### Damage resolution (VtM rules)
- All incoming damage starts as Superficial
- Aggravated sources (fire, sunlight, high-Potence vampires, supernatural powers): deal Aggravated directly
- Vampires halve Superficial damage (round up) before applying
- When Superficial fills health track: overflow converts to Aggravated
- When Aggravated fills health track: torpor

### PATCH `/api/combatants/{id}` new action types
- `superficial_damage`: applies halving for vampire combatants, handles overflow
- `aggravated_damage`: applies directly, handles overflow
- `spend_willpower`: 1 Superficial willpower damage; grants +2 dice for the turn

### Torpor detection
When `damage_aggravated >= health_max`: inject `[TORPOR]` flag into GM context for that combatant.

---

## 6. Frontend — CharacterSheetPanel

VtM rendering gated behind `ruleset.name === "vtm"`. No other rulesets affected.

### Layout components
- **Hunger track**: 5 red squares at the top; filled = current Hunger; click sets Hunger + PATCH; at Hunger 5 all squares pulse (CSS animation)
- **Attribute block**: 3 groups (Physical / Social / Mental), 3 attributes each, 5-dot pip rows, inline editable
- **Health track**: `health_max` boxes, each cycles empty → `/` (Superficial, grey) → `X` (Aggravated, red); Aggravated fills from right
- **Willpower track**: same pattern as Health, separate row
- **Humanity track**: 10-dot horizontal row, read-only; Stains shown below as small red marks (0–10)
- **Skills block**: 3 columns (Physical / Social / Mental), 5-dot rows, compact layout
- **Disciplines block**: 11 disciplines as 5-dot rows; disciplines with 0 dots collapsed by default
- **Identity fields**: Predator Type, Sect, Generation, Ambition, Desire as labeled text fields; Convictions and Touchstones as comma-separated text areas

---

## 7. Automation Goroutines & World Context

### autoUpdateCurrency — skip VtM
VtM uses Resources background (1–5 dots), not currency. Add `"vtm"` to the currency skip condition (same pattern as W&G).

### autoUpdateTension — VtM keywords
For VtM sessions only, add crisis keywords: `frenzy`, `beast`, `hunger`, `torpor`, `masquerade`, `breach`, `blood hunt`, `diablerie`, `farenheit`, `daybreak`.

### buildWorldContext — `[VtM MECHANICS]` block
Injected for VtM sessions after the character block:
```
[VtM MECHANICS]
Hunger: 2/5 | Humanity: 7 | Blood Potency: 1
Predator Type: Siren | Clan: Toreador
Health: ●●●●○ (1 Superficial) | Willpower: ●●●○○
Stains: 0
```

### autoUpdateCharacterStats — stain detection
Scan GM text for Humanity-violating keywords: `feeding`, `killing`, `diablerie`, `compulsion`, `breaking conviction`. On match: `stains` +1 (capped at 10), PATCH character, broadcast `character_updated`.

### autoUpdateRecap — Remorse reminder
After recap generation, if `stains > 0`: append Remorse Check reminder to session summary.

---

## 8. Ambient Audio & Scene Tags

### VtM single-track audio
- Frontend detects `ruleset.name === "vtm"` → loads `~/.ttrpg/audio/vtm/ambient.mp3` directly
- Track loops for the full session; scene_tags updates ignored for audio purposes
- Existing ambient audio manager gets a VtM branch; other rulesets unchanged

### New VtM scene tags (new migration)
Four tags added to allowed set, keyword-matched by `autoUpdateSceneTags` (zero AI cost):
| Tag | Keywords | Context |
|-----|----------|---------|
| `elysium` | elysium, court, gathering | Neutral vampire political ground |
| `haven` | haven, haven, lair, sanctuary | Personal safe space |
| `hunt` | hunt, feeding, stalking, prey | Active predation scene |
| `masquerade` | masquerade, crowd, public, humans watching | High-stakes human-presence scene |

### Masquerade integrity tracker
- New session-level field `masquerade_integrity` INT DEFAULT 10
- New goroutine `autoUpdateMasquerade` scans GM text:
  - Minor breach keywords (overheard, suspicious): −1
  - Moderate breach keywords (witnessed feeding, seen transforming): −2
  - Major breach keywords (on camera, viral, police): −3
  - Capped at 0
- GET/PATCH endpoints mirror tension endpoints
- Frontend: displayed in session header alongside Tension for VtM campaigns

---

## 9. Oracle & Compulsion Tables

### VtM oracle tables (new migration)
Replace generic Action/Theme oracle for VtM (ruleset_id = 8) with V5-thematic entries.

**Action table (50 entries):** VtM-flavored verbs — Hunt, Feed, Deceive, Dominate, Seduce, Betray, Protect, Flee, Investigate, Embrace, Manipulate, Observe, Infiltrate, Negotiate, Confront, Flee, Diablerize, Summon, Conceal, Reveal, etc.

**Theme table (50 entries):** VtM-flavored nouns — The Beast, The Masquerade, Elysium, Clan Politics, the Prince, a Coterie rival, Hunger, a Touchstone, a Sire, an Elder, a Threat to the Masquerade, Blood Bond, a Mortal, the Sheriff, Diablerie, a Secret, Paranoia, Manipulation, Ambition, Desire, etc.

### Clan Compulsion tables (7 new oracle tables)
One per core clan, 10 entries each. Rolls fire automatically on Messy Critical and inject result into GM context.

Thin-Blooded and Caitiff characters roll the generic Action/Theme oracle (no fixed Compulsion).

---

## Implementation Order

1. **Migration 024** — V5 character schema
2. **Migration 025** — VtM gm_context V5 rewrite
3. **Migration (combat)** — combatants damage columns
4. **Migration (scene tags)** — elysium, haven, hunt, masquerade tags + masquerade_integrity field
5. **Migration (oracle)** — VtM oracle + clan Compulsion tables
6. **options.go** — V5 clan/predator_type/sect/generation dropdowns
7. **vtm_predator_types.go** — All 10 types with exact grants
8. **random_stats.go** — V5 starting stat distribution
9. **routes.go / automation** — Currency skip, tension keywords, world context block, stain detection, Masquerade integrity goroutine
10. **Dice system** — Rouse check, Hunger dice, Blood Surge, Remorse check
11. **CharacterSheetPanel.tsx** — Full VtM rendering
12. **Ambient audio frontend** — VtM single-track branch

---

## Constraints

- **Per-ruleset isolation**: Every change is gated on `ruleset.name == "vtm"`. W&G, Ironsworn, and all other rulesets are completely untouched.
- **No global fallbacks**: `chunkRulebook` dispatch pattern — each ruleset is an explicit case.
- **Pre-existing test failures**: `rollWrathGloryStats()` signature mismatch in `internal/ruleset` is pre-existing; do not fix without confirming W&G gameplay is broken.
