# VtM V5 Full Overhaul Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement complete Vampire: The Masquerade 5th Edition support in inkandbone — schema, character creation, Hunger/Rouse dice, combat damage, automation goroutines, oracle tables, frontend character sheet, and ambient audio — without touching any other ruleset.

**Architecture:** All VtM logic is gated on `ruleset.Name == "vtm"` checks at every layer (DB, API, frontend). Migrations 024–028 add new columns/tables without modifying existing schema. New files `vtm_predator_types.go` and `queries_masquerade.go` keep VtM specifics isolated.

**Tech Stack:** Go 1.22, SQLite (mattn/go-sqlite3), React 18 + TypeScript, testify/assert, httptest

---

## File Map

**Create:**
- `internal/db/migrations/024_vtm_v5_schema.sql`
- `internal/db/migrations/025_vtm_gm_context_v5.sql`
- `internal/db/migrations/026_vtm_combat_damage.sql`
- `internal/db/migrations/027_vtm_scene_tags.sql`
- `internal/db/migrations/028_vtm_oracle_compulsion.sql`
- `internal/db/queries_masquerade.go`
- `internal/db/queries_masquerade_test.go`
- `internal/ruleset/vtm_predator_types.go`
- `internal/ruleset/vtm_predator_types_test.go`

**Modify:**
- `internal/ruleset/options.go` — replace VtM case with V5 dropdowns
- `internal/ruleset/random_stats.go` — replace VtM case with V5 stats
- `internal/db/queries_combat.go` — add VtM damage fields + methods
- `internal/api/routes.go` — currency skip, tension keywords, buildWorldContext VtM block, autoUpdateMasquerade, stain detection, Rouse/Hunger dice
- `internal/api/routes_phase_d.go` — masquerade_integrity GET/PATCH handlers
- `internal/api/server.go` — register masquerade routes
- `web/src/CharacterSheetPanel.tsx` — VtM-specific rendering branch
- `web/src/App.tsx` — VtM ambient audio branch

---

## Task 1: Migration 024 — V5 character schema

**Files:**
- Create: `internal/db/migrations/024_vtm_v5_schema.sql`

- [ ] **Step 1: Create the migration file**

```sql
-- 024_vtm_v5_schema.sql: Update VtM ruleset to V5 schema and fields
-- Updates the vtm ruleset row to V5 schema. Other rulesets untouched.
UPDATE rulesets SET
  schema_json = '{"system":"vtm","fields":[
    "clan","predator_type","sect","generation",
    "hunger","blood_potency","bane_severity","humanity","stains",
    "strength","dexterity","stamina",
    "charisma","manipulation","composure",
    "intelligence","wits","resolve",
    "athletics","brawl","craft","drive","firearms","larceny","melee","stealth","survival",
    "animal_ken","etiquette","insight","intimidation","leadership","performance","persuasion","streetwise","subterfuge",
    "academics","awareness","finance","investigation","medicine","occult","politics","technology",
    "animalism","auspex","blood_sorcery","celerity","dominate","fortitude","obfuscate","oblivion","potence","presence","protean",
    "health_max","health_superficial","health_aggravated",
    "willpower_max","willpower_superficial","willpower_aggravated",
    "skill_specialties","merits_flaws","convictions","touchstones","ambition","desire","notes"
  ]}',
  version = 'V5'
WHERE name = 'vtm';
```

- [ ] **Step 2: Verify the migration runs cleanly**

```bash
cd /home/digitalghost/projects/inkandbone && make test 2>&1 | tail -20
```
Expected: all tests pass (or same failures as before — no new failures).

- [ ] **Step 3: Commit**

```bash
git add internal/db/migrations/024_vtm_v5_schema.sql
git commit -m "feat(db): migration 024 — VtM V5 character schema"
```

---

## Task 2: Migration 025 — VtM GM context V5 rewrite

**Files:**
- Create: `internal/db/migrations/025_vtm_gm_context_v5.sql`

- [ ] **Step 1: Create the migration file**

```sql
-- 025_vtm_gm_context_v5.sql: Rewrite VtM gm_context to V5 accuracy
UPDATE rulesets SET gm_context = 'SETTING: Vampire: The Masquerade 5th Edition (V5). You are the Storyteller narrating a chronicle of personal horror, political intrigue, and the eternal struggle against the Beast.

VOCABULARY (mandatory — never use V20 terms):
- "Hunger" (never "blood pool"). "Rouse Check" (never "blood spending"). "Superficial damage" / "Aggravated damage" (never "lethal/bashing/aggravated"). "Blood Potency" as power metric (never "Generation" as a power scale). "Convictions" and "Touchstones" (never "Virtues").

HUNGER DIE NARRATION:
- Bestial Failure (a Hunger die shows 1, overall result is failure): describe animalistic loss of control — involuntary snarl, fingers curling into claws, the Beast surging against the cage of the mind. The character does something wrong, feral, or embarrassing.
- Messy Critical (a Hunger die shows 10, overall result is a critical success): the action succeeds but in a savage, uncontrolled way. Excessive force, blood spray, collateral damage, horrified witnesses. Success with a price.

ROUSE CHECK NARRATION:
- When Hunger increases: describe the gnawing emptiness behind the eyes, the warmth of nearby heartbeats becoming unbearable, predator instincts sharpening to a razor edge.
- When Hunger reaches 5: the character exists one heartbeat from Frenzy. Every interaction is a test of will.

FRENZY:
- Hunger Frenzy: triggered at Hunger 5 when provoked (smell of blood, being denied feeding). Resist with Composure + Resolve vs difficulty 3.
- Terror Frenzy: triggered by supernatural fear sources. Resist with Composure + Resolve.
- Rage Frenzy: triggered by humiliation or witnessing harm to a Touchstone. Resist with Composure + Resolve.
- When Frenzy is not resisted: narrate the Beast taking complete control. The character acts on pure predatory instinct. The player loses agency until the scene ends.

MASQUERADE:
- Always capitalize "the Masquerade" as a proper noun — it is the First Tradition.
- Minor breach (overheard conversation about vampires): 1 Masquerade point lost.
- Moderate breach (witnessed feeding, visible fangs): 2 points lost.
- Major breach (supernatural display caught on camera, police involvement): 3 points lost.
- The Sheriff and the Prince take breaches seriously. Repeated violations warrant Blood Hunts.

CLAN COMPULSIONS (trigger when a Messy Critical occurs):
- Brujah: Rebellion — must openly defy an authority figure this scene; cannot accept orders without resistance.
- Gangrel: Feral Impulse — adopts animal mannerisms (sniffing, circling, crouching); social rolls at +2 difficulty until scene ends.
- Malkavian: Delusion — becomes convinced of something demonstrably false; acts on that belief.
- Nosferatu: Cryptophilia — must obtain a secret from someone present before doing anything else.
- Toreador: Obsession — becomes transfixed by a beautiful or interesting stimulus; cannot voluntarily leave or act against it.
- Tremere: Perfectionism — cannot accept an imperfect outcome; must redo any action that fails or produces less than an exceptional result.
- Ventrue: Arrogance — refuses all assistance from anyone of lower social standing; must act alone.

TONE: Personal horror, moral compromise, political intrigue. The world is dark and the characters are monsters struggling to hold onto humanity. Tragedy is appropriate. NPCs have agendas. Elders are dangerous. Mortals are fragile and precious.

LENGTH: Exactly 4-5 paragraphs per response. Second person. No purple prose. Show, do not tell. Sentence variety. Never repeat a phrase used in the previous two responses.'
WHERE name = 'vtm';
```

- [ ] **Step 2: Run tests**

```bash
cd /home/digitalghost/projects/inkandbone && make test 2>&1 | tail -10
```
Expected: all tests pass.

- [ ] **Step 3: Commit**

```bash
git add internal/db/migrations/025_vtm_gm_context_v5.sql
git commit -m "feat(db): migration 025 — VtM V5 GM context rewrite"
```

---

## Task 3: Migration 026 — Combatants VtM damage columns

**Files:**
- Create: `internal/db/migrations/026_vtm_combat_damage.sql`

- [ ] **Step 1: Create the migration file**

```sql
-- 026_vtm_combat_damage.sql: Add VtM V5 damage tracking to combatants
ALTER TABLE combatants ADD COLUMN damage_superficial INTEGER NOT NULL DEFAULT 0;
ALTER TABLE combatants ADD COLUMN damage_aggravated INTEGER NOT NULL DEFAULT 0;
ALTER TABLE combatants ADD COLUMN willpower_superficial INTEGER NOT NULL DEFAULT 0;
ALTER TABLE combatants ADD COLUMN willpower_aggravated INTEGER NOT NULL DEFAULT 0;
ALTER TABLE combatants ADD COLUMN hunger INTEGER NOT NULL DEFAULT 0;
```

- [ ] **Step 2: Run tests**

```bash
cd /home/digitalghost/projects/inkandbone && make test 2>&1 | tail -10
```
Expected: all tests pass.

- [ ] **Step 3: Commit**

```bash
git add internal/db/migrations/026_vtm_combat_damage.sql
git commit -m "feat(db): migration 026 — VtM combat damage columns on combatants"
```

---

## Task 4: Migration 027 — VtM scene tags + masquerade_integrity

**Files:**
- Create: `internal/db/migrations/027_vtm_scene_tags.sql`

- [ ] **Step 1: Create the migration file**

```sql
-- 027_vtm_scene_tags.sql: Add masquerade_integrity to sessions; document VtM scene tags
-- masquerade_integrity tracks how many Masquerade points remain (0-10, default 10).
-- VtM scene tags (elysium, haven, hunt, masquerade) are handled via sceneTagKeywords
-- in routes.go — no schema change needed for tags themselves.
ALTER TABLE sessions ADD COLUMN masquerade_integrity INTEGER NOT NULL DEFAULT 10;
```

- [ ] **Step 2: Run tests**

```bash
cd /home/digitalghost/projects/inkandbone && make test 2>&1 | tail -10
```
Expected: all tests pass.

- [ ] **Step 3: Commit**

```bash
git add internal/db/migrations/027_vtm_scene_tags.sql
git commit -m "feat(db): migration 027 — sessions.masquerade_integrity for VtM"
```

---

## Task 5: Migration 028 — VtM oracle + clan Compulsion tables

**Files:**
- Create: `internal/db/migrations/028_vtm_oracle_compulsion.sql`

- [ ] **Step 1: Create the migration file**

The VtM ruleset_id is 3 (it is the 3rd row inserted in 002_seed_rulesets.sql: dnd5e=1, ironsworn=2, vtm=3). We use a subquery to be safe:

```sql
-- 028_vtm_oracle_compulsion.sql: VtM-specific oracle tables and clan Compulsion tables
-- Action and Theme oracles replace the generic ones for VtM sessions.
-- Clan Compulsion tables (10 entries each) fire on Messy Critical results.

-- VtM Action oracle (ruleset-specific, rolls 1-50)
INSERT INTO oracle_tables (ruleset_id, table_name, roll_min, roll_max, result)
SELECT id, 'action', 1, 2, 'Hunt' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 3, 4, 'Feed' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 5, 6, 'Deceive' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 7, 8, 'Dominate' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 9, 10, 'Seduce' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 11, 12, 'Betray' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 13, 14, 'Protect' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 15, 16, 'Flee' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 17, 18, 'Investigate' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 19, 20, 'Embrace' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 21, 22, 'Manipulate' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 23, 24, 'Observe' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 25, 26, 'Infiltrate' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 27, 28, 'Negotiate' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 29, 30, 'Confront' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 31, 32, 'Diablerize' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 33, 34, 'Summon' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 35, 36, 'Conceal' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 37, 38, 'Reveal' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 39, 40, 'Stalk' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 41, 42, 'Escape' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 43, 44, 'Claim' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 45, 46, 'Surrender' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 47, 48, 'Expose' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'action', 49, 50, 'Endure' FROM rulesets WHERE name = 'vtm';

-- VtM Theme oracle (ruleset-specific, rolls 1-50)
INSERT INTO oracle_tables (ruleset_id, table_name, roll_min, roll_max, result)
SELECT id, 'theme', 1, 2, 'The Beast' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 3, 4, 'The Masquerade' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 5, 6, 'Elysium' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 7, 8, 'Clan Politics' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 9, 10, 'The Prince' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 11, 12, 'A Coterie Rival' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 13, 14, 'Hunger' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 15, 16, 'A Touchstone' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 17, 18, 'A Sire' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 19, 20, 'An Elder' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 21, 22, 'A Masquerade Breach' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 23, 24, 'Blood Bond' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 25, 26, 'A Mortal Witness' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 27, 28, 'The Sheriff' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 29, 30, 'Diablerie' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 31, 32, 'A Hidden Secret' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 33, 34, 'Paranoia' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 35, 36, 'Manipulation' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 37, 38, 'Ambition' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 39, 40, 'Desire' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 41, 42, 'Humanity' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 43, 44, 'The Rack' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 45, 46, 'An Old Enemy' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 47, 48, 'A New Threat' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'theme', 49, 50, 'Redemption' FROM rulesets WHERE name = 'vtm';

-- Brujah Compulsion: Rebellion (table_name = 'compulsion_brujah', rolls 1-10)
INSERT INTO oracle_tables (ruleset_id, table_name, roll_min, roll_max, result)
SELECT id, 'compulsion_brujah', 1, 1, 'You openly contradict the most powerful person in the room, loudly and publicly.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_brujah', 2, 2, 'You refuse a direct order, even from an ally, on principle.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_brujah', 3, 3, 'You destroy something that symbolizes authority — a badge, a door, a throne.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_brujah', 4, 4, 'You side with whoever is being oppressed, regardless of the facts.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_brujah', 5, 5, 'You start a fight with the most dominant figure present.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_brujah', 6, 6, 'You loudly enumerate every injustice you have witnessed tonight.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_brujah', 7, 7, 'You demand an explanation for every rule you are expected to follow.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_brujah', 8, 8, 'You refuse to be the first to back down from any confrontation.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_brujah', 9, 9, 'You champion a stranger as an act of defiance against their oppressor.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_brujah', 10, 10, 'You announce that the current power structure is corrupt and must fall.' FROM rulesets WHERE name = 'vtm';

-- Gangrel Compulsion: Feral Impulse (table_name = 'compulsion_gangrel', rolls 1-10)
INSERT INTO oracle_tables (ruleset_id, table_name, roll_min, roll_max, result)
SELECT id, 'compulsion_gangrel', 1, 1, 'You drop to all fours and sniff the ground, tracking a scent only you can sense.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_gangrel', 2, 2, 'You circle the room slowly, marking territory in your mind.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_gangrel', 3, 3, 'You growl audibly at anyone who steps too close.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_gangrel', 4, 4, 'You crouch rather than sit. Standing feels wrong.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_gangrel', 5, 5, 'You find the nearest exit and position yourself near it, unable to relax otherwise.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_gangrel', 6, 6, 'You snap your teeth at the nearest mortal who speaks to you.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_gangrel', 7, 7, 'You eat something raw — meat, vermin, whatever is available.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_gangrel', 8, 8, 'You refuse to enter any building; the outdoors is the only safe place.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_gangrel', 9, 9, 'You track a target across the room on instinct before catching yourself.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_gangrel', 10, 10, 'You produce a low, rumbling territorial growl whenever a stranger approaches.' FROM rulesets WHERE name = 'vtm';

-- Malkavian Compulsion: Delusion (table_name = 'compulsion_malkavian', rolls 1-10)
INSERT INTO oracle_tables (ruleset_id, table_name, roll_min, roll_max, result)
SELECT id, 'compulsion_malkavian', 1, 1, 'You are convinced someone in the room is not who they claim to be — a spy, a demon, or an impostor.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_malkavian', 2, 2, 'You believe an inanimate object in the room is speaking to you and must be obeyed.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_malkavian', 3, 3, 'You are certain tonight is a night you have lived before — an exact repetition.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_malkavian', 4, 4, 'You believe you are being watched by someone invisible and act accordingly.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_malkavian', 5, 5, 'You are convinced that one specific person is the key to preventing a catastrophe.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_malkavian', 6, 6, 'You believe the numbers in the room have deep significance and must be counted.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_malkavian', 7, 7, 'You are certain the current location is about to be destroyed and urge everyone to leave.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_malkavian', 8, 8, 'You become convinced you have already been betrayed tonight by a trusted ally.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_malkavian', 9, 9, 'You believe that if you speak above a whisper, something terrible will happen.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_malkavian', 10, 10, 'You are certain that the correct course of action is the exact opposite of what seems logical.' FROM rulesets WHERE name = 'vtm';

-- Nosferatu Compulsion: Cryptophilia (table_name = 'compulsion_nosferatu', rolls 1-10)
INSERT INTO oracle_tables (ruleset_id, table_name, roll_min, roll_max, result)
SELECT id, 'compulsion_nosferatu', 1, 1, 'You must learn where the most valuable person in the room sleeps during the day.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_nosferatu', 2, 2, 'You must discover what the most powerful person present is hiding.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_nosferatu', 3, 3, 'You must obtain a confession of some kind before the scene ends.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_nosferatu', 4, 4, 'You must find out who in the room is lying about their identity.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_nosferatu', 5, 5, 'You must learn what secret deal was recently struck between two parties present.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_nosferatu', 6, 6, 'You must acquire physical proof of wrongdoing by someone present.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_nosferatu', 7, 7, 'You must discover what weakness the nearest Elder possesses.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_nosferatu', 8, 8, 'You must find out who sent someone here tonight and why.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_nosferatu', 9, 9, 'You must learn the true name of a mortal who has interacted with the Kindred recently.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_nosferatu', 10, 10, 'You must discover what illicit transaction is occurring or has recently occurred nearby.' FROM rulesets WHERE name = 'vtm';

-- Toreador Compulsion: Obsession (table_name = 'compulsion_toreador', rolls 1-10)
INSERT INTO oracle_tables (ruleset_id, table_name, roll_min, roll_max, result)
SELECT id, 'compulsion_toreador', 1, 1, 'A piece of music playing nearby transfixes you. You cannot willingly leave until it ends.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_toreador', 2, 2, 'A mortal in the room possesses an almost supernatural physical perfection. You cannot stop watching them.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_toreador', 3, 3, 'The architecture or decor of this location captivates you. You must examine every detail.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_toreador', 4, 4, 'Someone''s voice is so beautiful that you cannot act while they are speaking.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_toreador', 5, 5, 'A work of art — painting, sculpture, or photograph — demands your complete attention.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_toreador', 6, 6, 'The way someone moves across the room is so elegant you must follow and observe.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_toreador', 7, 7, 'A tragic story being told nearby is so affecting you cannot act until it concludes.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_toreador', 8, 8, 'The interplay of light and shadow in this location arrests your attention completely.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_toreador', 9, 9, 'Someone''s grief or joy is so raw and genuine that you are unable to look away.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_toreador', 10, 10, 'A scent — perfume, blood, old paper — triggers a powerful aesthetic memory. You are lost in it.' FROM rulesets WHERE name = 'vtm';

-- Tremere Compulsion: Perfectionism (table_name = 'compulsion_tremere', rolls 1-10)
INSERT INTO oracle_tables (ruleset_id, table_name, roll_min, roll_max, result)
SELECT id, 'compulsion_tremere', 1, 1, 'Your last spoken statement contained an imprecision. You must correct it, in detail, immediately.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_tremere', 2, 2, 'An action you recently took was suboptimal. You must attempt it again, properly.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_tremere', 3, 3, 'Someone nearby is doing something incorrectly. You cannot proceed until you have corrected them.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_tremere', 4, 4, 'The plan as stated has a flaw. You refuse to proceed until it is revised to your satisfaction.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_tremere', 5, 5, 'You must restate your position using precisely the correct terminology, not approximations.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_tremere', 6, 6, 'A tool or object is not in its correct place. You must correct this before anything else.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_tremere', 7, 7, 'Your appearance is imperfect. You spend time correcting it even if the situation is urgent.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_tremere', 8, 8, 'The outcome was acceptable but not excellent. You must explain how it could have been better.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_tremere', 9, 9, 'An agreement is missing crucial specifics. You refuse to act on it until every term is defined.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_tremere', 10, 10, 'A ritual or formula was performed incorrectly by someone nearby. You must perform it again, correctly.' FROM rulesets WHERE name = 'vtm';

-- Ventrue Compulsion: Arrogance (table_name = 'compulsion_ventrue', rolls 1-10)
INSERT INTO oracle_tables (ruleset_id, table_name, roll_min, roll_max, result)
SELECT id, 'compulsion_ventrue', 1, 1, 'You refuse to accept assistance from anyone who has not proven their worth to you.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_ventrue', 2, 2, 'You insist on leading, even in a domain that is clearly not yours.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_ventrue', 3, 3, 'You publicly dismiss the opinion of whoever speaks last, regardless of its merit.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_ventrue', 4, 4, 'You demand to be addressed by your full title before cooperating with anyone.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_ventrue', 5, 5, 'You will not share a resource or advantage with someone of lower station, even if it costs you.' FROM ruleests WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_ventrue', 6, 6, 'You correct someone''s etiquette publicly, in detail, even if it is deeply inconvenient.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_ventrue', 7, 7, 'You take credit for a group success, attributing it to your leadership.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_ventrue', 8, 8, 'You refuse to negotiate as an equal with anyone you consider beneath you.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_ventrue', 9, 9, 'You delegate a task to a subordinate rather than perform it yourself, even urgently.' FROM rulesets WHERE name = 'vtm'
UNION ALL SELECT id, 'compulsion_ventrue', 10, 10, 'You make a unilateral decision affecting the group without consulting anyone.' FROM rulesets WHERE name = 'vtm';
```

- [ ] **Step 2: Fix typo in migration (ruleests → rulesets)**

In the file, find the line with `FROM ruleests` and fix it to `FROM rulesets`.

- [ ] **Step 3: Run tests**

```bash
cd /home/digitalghost/projects/inkandbone && make test 2>&1 | tail -10
```
Expected: all tests pass.

- [ ] **Step 4: Commit**

```bash
git add internal/db/migrations/028_vtm_oracle_compulsion.sql
git commit -m "feat(db): migration 028 — VtM oracle tables and clan Compulsion tables"
```

---

## Task 6: options.go — V5 dropdowns

**Files:**
- Modify: `internal/ruleset/options.go`

- [ ] **Step 1: Write the failing test**

In `internal/ruleset/options_test.go` (create if not exists):

```go
package ruleset

import "testing"

func TestVtMOptions_V5Clans(t *testing.T) {
    opts := CharacterOptions("vtm")
    clans, ok := opts["clan"]
    if !ok {
        t.Fatal("vtm options missing clan key")
    }
    want := []string{"Brujah", "Gangrel", "Malkavian", "Nosferatu", "Toreador", "Tremere", "Ventrue", "Caitiff", "Thin-Blooded"}
    if len(clans) != len(want) {
        t.Fatalf("expected %d clans, got %d: %v", len(want), len(clans), clans)
    }
    clanSet := map[string]bool{}
    for _, c := range clans {
        clanSet[c] = true
    }
    for _, w := range want {
        if !clanSet[w] {
            t.Errorf("missing clan %q", w)
        }
    }
}

func TestVtMOptions_PredatorType(t *testing.T) {
    opts := CharacterOptions("vtm")
    types, ok := opts["predator_type"]
    if !ok {
        t.Fatal("vtm options missing predator_type key")
    }
    if len(types) != 10 {
        t.Fatalf("expected 10 predator types, got %d", len(types))
    }
}

func TestVtMOptions_Sect(t *testing.T) {
    opts := CharacterOptions("vtm")
    if _, ok := opts["sect"]; !ok {
        t.Fatal("vtm options missing sect key")
    }
}
```

- [ ] **Step 2: Run the test to confirm it fails**

```bash
cd /home/digitalghost/projects/inkandbone && go test ./internal/ruleset/ -run TestVtMOptions -v 2>&1 | tail -20
```
Expected: FAIL — Brujah missing, predator_type missing, sect missing.

- [ ] **Step 3: Replace the VtM case in options.go**

Find the existing `case "vtm":` block (currently lines ~37-44) and replace it:

```go
	case "vtm":
		return map[string][]string{
			"clan": {
				"Brujah", "Gangrel", "Malkavian", "Nosferatu", "Toreador",
				"Tremere", "Ventrue", "Caitiff", "Thin-Blooded",
			},
			"predator_type": {
				"Alleycat", "Bagger", "Blood Leech", "Cleaner", "Consensualist",
				"Extortionist", "Graverobber", "Osiris", "Sandman", "Siren",
			},
			"sect": {
				"Camarilla", "Anarch", "Unaligned", "Sabbat (lapsed)",
			},
			"generation": {
				"10th", "11th", "12th", "13th", "14th", "15th (Thin-Blooded)",
			},
		}
```

- [ ] **Step 4: Run the test to confirm it passes**

```bash
cd /home/digitalghost/projects/inkandbone && go test ./internal/ruleset/ -run TestVtMOptions -v 2>&1 | tail -20
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/ruleset/options.go internal/ruleset/options_test.go
git commit -m "feat(ruleset): VtM V5 character options — clans, predator types, sects, generation"
```

---

## Task 7: vtm_predator_types.go — All 10 types + apply function

**Files:**
- Create: `internal/ruleset/vtm_predator_types.go`
- Create: `internal/ruleset/vtm_predator_types_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/ruleset/vtm_predator_types_test.go
package ruleset

import (
    "strings"
    "testing"
)

func TestVtMPredatorTypes_allTenDefined(t *testing.T) {
    types := []string{
        "Alleycat", "Bagger", "Blood Leech", "Cleaner", "Consensualist",
        "Extortionist", "Graverobber", "Osiris", "Sandman", "Siren",
    }
    for _, name := range types {
        if _, ok := vtmPredatorTypes[name]; !ok {
            t.Errorf("predator type %q not defined", name)
        }
    }
}

func TestVtMPredatorTypes_eachHasTwoDisciplines(t *testing.T) {
    for name, pt := range vtmPredatorTypes {
        if len(pt.disciplines) != 2 {
            t.Errorf("%q: expected 2 discipline grants, got %d", name, len(pt.disciplines))
        }
    }
}

func TestApplyVtMPredatorType_sirenGrantsPresenceAndPotence(t *testing.T) {
    stats := map[string]any{
        "presence": 0,
        "potence":  0,
        "skill_specialties": "",
        "merits_flaws":      "",
    }
    ApplyVtMPredatorType("Siren", stats)
    if stats["presence"] != 1 {
        t.Errorf("expected presence=1, got %v", stats["presence"])
    }
    if stats["potence"] != 1 {
        t.Errorf("expected potence=1, got %v", stats["potence"])
    }
    if !strings.Contains(stats["skill_specialties"].(string), "Seduction") {
        t.Errorf("expected Seduction specialty, got %v", stats["skill_specialties"])
    }
}

func TestApplyVtMPredatorType_unknownType_noOp(t *testing.T) {
    stats := map[string]any{"presence": 2}
    ApplyVtMPredatorType("Unknown", stats)
    if stats["presence"] != 2 {
        t.Errorf("unknown type should not modify stats")
    }
}
```

- [ ] **Step 2: Run the test to confirm it fails**

```bash
cd /home/digitalghost/projects/inkandbone && go test ./internal/ruleset/ -run TestVtMPredator -v 2>&1 | tail -20
```
Expected: FAIL — compilation error, vtmPredatorTypes undefined.

- [ ] **Step 3: Create vtm_predator_types.go**

```go
// internal/ruleset/vtm_predator_types.go
package ruleset

// vtmPredatorTypeDef holds the mechanical grants for a VtM V5 Predator Type.
type vtmPredatorTypeDef struct {
    disciplines []disciplineGrant // exactly 2
    specialty   string            // "Skill (Specialty)"
    meritsFlaws string            // free-text description
}

type disciplineGrant struct {
    name  string
    bonus int
}

// vtmPredatorTypes maps each V5 Predator Type to its rulebook-defined grants.
var vtmPredatorTypes = map[string]vtmPredatorTypeDef{
    "Alleycat": {
        disciplines: []disciplineGrant{{"celerity", 1}, {"potence", 1}},
        specialty:   "Brawl (Grappling)",
        meritsFlaws: "Merit: Prowler's Instinct / Flaw: Prey Exclusion (Homeless)",
    },
    "Bagger": {
        disciplines: []disciplineGrant{{"obfuscate", 1}, {"auspex", 1}},
        specialty:   "Streetwise (Black Market)",
        meritsFlaws: "Merit: Iron Gullet / Flaw: Prey Exclusion (Bagged Blood)",
    },
    "Blood Leech": {
        disciplines: []disciplineGrant{{"obfuscate", 1}, {"potence", 1}},
        specialty:   "Stealth (Shadowing)",
        meritsFlaws: "Flaw: Shunned, Prey Exclusion (Mortals)",
    },
    "Cleaner": {
        disciplines: []disciplineGrant{{"obfuscate", 1}, {"dominate", 1}},
        specialty:   "Subterfuge (Impersonation)",
        meritsFlaws: "Merit: Retainer / Flaw: Obvious Predator",
    },
    "Consensualist": {
        disciplines: []disciplineGrant{{"presence", 1}, {"auspex", 1}},
        specialty:   "Persuasion (Victim Calming)",
        meritsFlaws: "Merit: Herd / Flaw: Prey Exclusion (Non-consenting)",
    },
    "Extortionist": {
        disciplines: []disciplineGrant{{"dominate", 1}, {"presence", 1}},
        specialty:   "Intimidation (Blackmail)",
        meritsFlaws: "Merit: Contacts / Flaw: Prey Exclusion (Vulnerable)",
    },
    "Graverobber": {
        disciplines: []disciplineGrant{{"fortitude", 1}, {"oblivion", 1}},
        specialty:   "Medicine (Cadavers)",
        meritsFlaws: "Flaw: Obvious Predator",
    },
    "Osiris": {
        disciplines: []disciplineGrant{{"blood_sorcery", 1}, {"presence", 1}},
        specialty:   "Occult (Specific Tradition)",
        meritsFlaws: "Merit: Fame / Flaw: Prey Exclusion (Faithful)",
    },
    "Sandman": {
        disciplines: []disciplineGrant{{"auspex", 1}, {"obfuscate", 1}},
        specialty:   "Stealth (Breaking and Entering)",
        meritsFlaws: "Flaw: Prey Exclusion (Sleeping)",
    },
    "Siren": {
        disciplines: []disciplineGrant{{"presence", 1}, {"potence", 1}},
        specialty:   "Persuasion (Seduction)",
        meritsFlaws: "Merit: Looks (Beautiful) / Flaw: Prey Exclusion (Mortals in relationships)",
    },
}

// ApplyVtMPredatorType modifies a character stats map in-place with the grants
// for the named Predator Type. No-op if the type is not found.
// stats keys used: discipline names (incrementing int), "skill_specialties" (string), "merits_flaws" (string).
func ApplyVtMPredatorType(predatorType string, stats map[string]any) {
    pt, ok := vtmPredatorTypes[predatorType]
    if !ok {
        return
    }
    for _, d := range pt.disciplines {
        current := 0
        if v, ok := stats[d.name]; ok {
            switch n := v.(type) {
            case int:
                current = n
            case float64:
                current = int(n)
            }
        }
        stats[d.name] = current + d.bonus
    }
    if pt.specialty != "" {
        existing, _ := stats["skill_specialties"].(string)
        if existing == "" {
            stats["skill_specialties"] = pt.specialty
        } else {
            stats["skill_specialties"] = existing + ", " + pt.specialty
        }
    }
    if pt.meritsFlaws != "" {
        existing, _ := stats["merits_flaws"].(string)
        if existing == "" {
            stats["merits_flaws"] = pt.meritsFlaws
        } else {
            stats["merits_flaws"] = existing + "; " + pt.meritsFlaws
        }
    }
}
```

- [ ] **Step 4: Run the test to confirm it passes**

```bash
cd /home/digitalghost/projects/inkandbone && go test ./internal/ruleset/ -run TestVtMPredator -v 2>&1 | tail -20
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/ruleset/vtm_predator_types.go internal/ruleset/vtm_predator_types_test.go
git commit -m "feat(ruleset): VtM V5 predator types with discipline/specialty/merit grants"
```

---

## Task 8: random_stats.go — V5 starting stats

**Files:**
- Modify: `internal/ruleset/random_stats.go`

The existing VtM case is at line ~269. Replace it entirely.

- [ ] **Step 1: Write the failing test**

Add to `internal/ruleset/random_stats_test.go` (append, do not modify W&G tests):

```go
func TestRollVtMStats_V5Fields(t *testing.T) {
    for i := 0; i < 10; i++ {
        stats := rollStats("vtm")
        // Core V5 fields must be present
        for _, key := range []string{
            "hunger", "blood_potency", "bane_severity", "humanity", "stains",
            "strength", "dexterity", "stamina",
            "charisma", "manipulation", "composure",
            "intelligence", "wits", "resolve",
        } {
            if _, ok := stats[key]; !ok {
                t.Errorf("missing field %q", key)
            }
        }
        // hunger starts at 1
        if stats["hunger"] != 1 {
            t.Errorf("hunger should be 1, got %v", stats["hunger"])
        }
        // humanity starts at 7
        if stats["humanity"] != 7 {
            t.Errorf("humanity should be 7, got %v", stats["humanity"])
        }
        // health_max = stamina + 3
        stamina, _ := stats["stamina"].(int)
        healthMax, _ := stats["health_max"].(int)
        if healthMax != stamina+3 {
            t.Errorf("health_max should be stamina+3=%d, got %d", stamina+3, healthMax)
        }
        // willpower_max = composure + resolve
        composure, _ := stats["composure"].(int)
        resolve, _ := stats["resolve"].(int)
        willpowerMax, _ := stats["willpower_max"].(int)
        if willpowerMax != composure+resolve {
            t.Errorf("willpower_max should be composure+resolve=%d, got %d", composure+resolve, willpowerMax)
        }
    }
}
```

- [ ] **Step 2: Run the test to confirm it fails**

```bash
cd /home/digitalghost/projects/inkandbone && go test ./internal/ruleset/ -run TestRollVtMStats -v 2>&1 | tail -20
```
Expected: FAIL — missing fields.

- [ ] **Step 3: Replace the VtM case in random_stats.go**

Find the `case "vtm":` block starting around line 269 and replace it:

```go
	case "vtm":
		return rollVtMV5Stats()
```

Add the `rollVtMV5Stats` function at the bottom of the file (before the closing):

```go
// rollVtMV5Stats generates a V5 starting vampire character.
// Attribute distribution: primary group (4,3,2,1), secondary (3,3,2,1), tertiary (2,2,2,1).
func rollVtMV5Stats() map[string]any {
    // Roll attributes in three groups, shuffle within each group
    physical := distributeAttrs([]int{4, 3, 2, 1}[:3], 3) // pick 3 from shuffled
    social := distributeAttrs([]int{3, 3, 2, 1}[:3], 3)
    mental := distributeAttrs([]int{2, 2, 2, 1}[:3], 3)

    // Randomly assign which group is primary/secondary/tertiary
    groups := [3][3]int{physical, social, mental}
    order := []int{0, 1, 2}
    for i := range order {
        j := i + mathrand.Intn(len(order)-i)
        order[i], order[j] = order[j], order[i]
    }
    // primary = groups[order[0]], secondary = groups[order[1]], tertiary = groups[order[2]]
    // Assign primary values (highest pool) to a random group
    primaries := [3][3]int{
        {4, 3, 2},
        {3, 3, 2},
        {2, 2, 2},
    }
    shuffled := [3][3]int{}
    for g := 0; g < 3; g++ {
        vals := primaries[order[g]]
        for i := range vals {
            j := i + mathrand.Intn(len(vals)-i)
            vals[i], vals[j] = vals[j], vals[i]
        }
        shuffled[g] = vals
    }
    _ = groups

    str := shuffled[0][0]
    dex := shuffled[0][1]
    sta := shuffled[0][2]
    cha := shuffled[1][0]
    man := shuffled[1][1]
    com := shuffled[1][2]
    intel := shuffled[2][0]
    wit := shuffled[2][1]
    res := shuffled[2][2]

    clans := []string{"Brujah", "Gangrel", "Malkavian", "Nosferatu", "Toreador", "Tremere", "Ventrue", "Caitiff", "Thin-Blooded"}
    predTypes := []string{"Alleycat", "Bagger", "Blood Leech", "Cleaner", "Consensualist", "Extortionist", "Graverobber", "Osiris", "Sandman", "Siren"}
    sects := []string{"Camarilla", "Anarch", "Unaligned"}

    stats := map[string]any{
        "clan": randPick(clans), "predator_type": randPick(predTypes), "sect": randPick(sects),
        "generation": randPick([]string{"10th", "11th", "12th", "13th"}),
        "hunger": 1, "blood_potency": 1, "bane_severity": 1,
        "humanity": 7, "stains": 0,
        "strength": str, "dexterity": dex, "stamina": sta,
        "charisma": cha, "manipulation": man, "composure": com,
        "intelligence": intel, "wits": wit, "resolve": res,
        "health_max": sta + 3, "health_superficial": 0, "health_aggravated": 0,
        "willpower_max": com + res, "willpower_superficial": 0, "willpower_aggravated": 0,
        // All skills start at 0
        "athletics": 0, "brawl": 0, "craft": 0, "drive": 0, "firearms": 0,
        "larceny": 0, "melee": 0, "stealth": 0, "survival": 0,
        "animal_ken": 0, "etiquette": 0, "insight": 0, "intimidation": 0,
        "leadership": 0, "performance": 0, "persuasion": 0, "streetwise": 0, "subterfuge": 0,
        "academics": 0, "awareness": 0, "finance": 0, "investigation": 0,
        "medicine": 0, "occult": 0, "politics": 0, "technology": 0,
        // All disciplines start at 0 (predator type apply happens after creation)
        "animalism": 0, "auspex": 0, "blood_sorcery": 0, "celerity": 0,
        "dominate": 0, "fortitude": 0, "obfuscate": 0, "oblivion": 0,
        "potence": 0, "presence": 0, "protean": 0,
        "skill_specialties": "", "merits_flaws": "", "ambition": "", "desire": "",
        "convictions": "", "touchstones": "", "notes": "",
    }
    // Apply predator type grants
    ApplyVtMPredatorType(stats["predator_type"].(string), stats)
    return stats
}

func distributeAttrs(vals []int, n int) [3]int {
    cp := make([]int, len(vals))
    copy(cp, vals)
    for i := range cp {
        j := i + mathrand.Intn(len(cp)-i)
        cp[i], cp[j] = cp[j], cp[i]
    }
    var out [3]int
    for i := 0; i < n && i < len(cp); i++ {
        out[i] = cp[i]
    }
    return out
}
```

Also ensure `mathrand` is imported — check existing imports in the file. If the file uses `math/rand` imported as `mathrand`, that's already correct.

- [ ] **Step 4: Run the test to confirm it passes**

```bash
cd /home/digitalghost/projects/inkandbone && go test ./internal/ruleset/ -run TestRollVtMStats -v 2>&1 | tail -20
```
Expected: PASS.

- [ ] **Step 5: Run all ruleset tests to confirm no regressions**

```bash
cd /home/digitalghost/projects/inkandbone && go test ./internal/ruleset/ -v 2>&1 | grep -E "PASS|FAIL|---"
```
Expected: All passing tests still pass. The pre-existing `rollWrathGloryStats` signature failure is allowed.

- [ ] **Step 6: Commit**

```bash
git add internal/ruleset/random_stats.go internal/ruleset/random_stats_test.go
git commit -m "feat(ruleset): VtM V5 starting stats — attributes, health, willpower, predator type apply"
```

---

## Task 9: queries_masquerade.go — DB methods for masquerade_integrity

**Files:**
- Create: `internal/db/queries_masquerade.go`
- Create: `internal/db/queries_masquerade_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/db/queries_masquerade_test.go
package db

import (
    "testing"
)

func TestGetMasqueradeIntegrity_Default(t *testing.T) {
    db := openTestDB(t)
    campID := createTestCampaign(t, db)
    sessID := createTestSession(t, db, campID)

    integrity, err := db.GetMasqueradeIntegrity(sessID)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if integrity != 10 {
        t.Errorf("expected default integrity=10, got %d", integrity)
    }
}

func TestUpdateMasqueradeIntegrity_ClampMin(t *testing.T) {
    db := openTestDB(t)
    campID := createTestCampaign(t, db)
    sessID := createTestSession(t, db, campID)

    if err := db.UpdateMasqueradeIntegrity(sessID, -5); err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    v, _ := db.GetMasqueradeIntegrity(sessID)
    if v != 0 {
        t.Errorf("expected clamp to 0, got %d", v)
    }
}

func TestUpdateMasqueradeIntegrity_ClampMax(t *testing.T) {
    db := openTestDB(t)
    campID := createTestCampaign(t, db)
    sessID := createTestSession(t, db, campID)

    if err := db.UpdateMasqueradeIntegrity(sessID, 15); err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    v, _ := db.GetMasqueradeIntegrity(sessID)
    if v != 10 {
        t.Errorf("expected clamp to 10, got %d", v)
    }
}
```

To find what helpers (`openTestDB`, `createTestCampaign`, `createTestSession`) look like, check `internal/db/queries_tension_test.go` — use the exact same pattern.

- [ ] **Step 2: Run the test to confirm it fails**

```bash
cd /home/digitalghost/projects/inkandbone && go test ./internal/db/ -run TestGetMasquerade -v 2>&1 | tail -20
```
Expected: FAIL — `GetMasqueradeIntegrity` undefined.

- [ ] **Step 3: Create queries_masquerade.go**

```go
// internal/db/queries_masquerade.go
package db

// GetMasqueradeIntegrity returns the Masquerade integrity (0-10) for a session.
// Default is 10 (full Masquerade intact).
func (db *DB) GetMasqueradeIntegrity(sessionID int64) (int, error) {
    var level int
    err := db.db.QueryRow(
        `SELECT masquerade_integrity FROM sessions WHERE id = ?`, sessionID,
    ).Scan(&level)
    return level, err
}

// UpdateMasqueradeIntegrity sets the Masquerade integrity for a session, clamping to [0, 10].
func (db *DB) UpdateMasqueradeIntegrity(sessionID int64, level int) error {
    if level < 0 {
        level = 0
    }
    if level > 10 {
        level = 10
    }
    _, err := db.db.Exec(
        `UPDATE sessions SET masquerade_integrity = ? WHERE id = ?`, level, sessionID,
    )
    return err
}
```

- [ ] **Step 4: Run the test to confirm it passes**

```bash
cd /home/digitalghost/projects/inkandbone && go test ./internal/db/ -run TestGetMasquerade -v -run TestUpdateMasquerade 2>&1 | tail -20
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/db/queries_masquerade.go internal/db/queries_masquerade_test.go
git commit -m "feat(db): GetMasqueradeIntegrity / UpdateMasqueradeIntegrity for VtM sessions"
```

---

## Task 10: queries_combat.go — VtM damage DB methods

**Files:**
- Modify: `internal/db/queries_combat.go`

Add VtM damage fields to the `Combatant` struct and a new `UpdateCombatantVtMDamage` method.

- [ ] **Step 1: Write the failing test**

In `internal/db/queries_combat_test.go`, append:

```go
func TestUpdateCombatantVtMDamage_SuperficialHalving(t *testing.T) {
    db := openTestDB(t)
    campID := createTestCampaign(t, db)
    sessID := createTestSession(t, db, campID)
    encID, _ := db.CreateEncounter(sessID, "Test Fight")
    combID, _ := db.AddCombatant(encID, "Vampire", 10, 6, true, nil)

    // Apply 4 superficial damage (vampires halve: 4→2 applied)
    err := db.UpdateCombatantVtMDamage(combID, 4, 0, false)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    combatants, _ := db.ListCombatants(encID)
    if len(combatants) == 0 {
        t.Fatal("no combatants returned")
    }
    c := combatants[0]
    if c.DamageSuperficial != 2 {
        t.Errorf("expected 2 superficial after halving, got %d", c.DamageSuperficial)
    }
}

func TestUpdateCombatantVtMDamage_AggravatedDirect(t *testing.T) {
    db := openTestDB(t)
    campID := createTestCampaign(t, db)
    sessID := createTestSession(t, db, campID)
    encID, _ := db.CreateEncounter(sessID, "Test Fight")
    combID, _ := db.AddCombatant(encID, "Vampire", 10, 6, true, nil)

    err := db.UpdateCombatantVtMDamage(combID, 0, 2, false)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    combatants, _ := db.ListCombatants(encID)
    c := combatants[0]
    if c.DamageAggravated != 2 {
        t.Errorf("expected 2 aggravated, got %d", c.DamageAggravated)
    }
}
```

- [ ] **Step 2: Run test to confirm it fails**

```bash
cd /home/digitalghost/projects/inkandbone && go test ./internal/db/ -run TestUpdateCombatantVtMDamage -v 2>&1 | tail -20
```
Expected: FAIL — DamageSuperficial field undefined.

- [ ] **Step 3: Update Combatant struct and add method in queries_combat.go**

Add VtM fields to the `Combatant` struct (after `ConditionsJSON`):

```go
type Combatant struct {
    ID                  int64  `json:"id"`
    EncounterID         int64  `json:"encounter_id"`
    CharacterID         *int64 `json:"character_id"`
    Name                string `json:"name"`
    Initiative          int    `json:"initiative"`
    HPCurrent           int    `json:"hp_current"`
    HPMax               int    `json:"hp_max"`
    ConditionsJSON      string `json:"conditions_json"`
    IsPlayer            bool   `json:"is_player"`
    DamageSuperficial   int    `json:"damage_superficial"`
    DamageAggravated    int    `json:"damage_aggravated"`
    WillpowerSuperficial int   `json:"willpower_superficial"`
    WillpowerAggravated  int   `json:"willpower_aggravated"`
    Hunger              int    `json:"hunger"`
}
```

Update `ListCombatants` to scan the new columns:

```go
func (d *DB) ListCombatants(encounterID int64) ([]Combatant, error) {
    rows, err := d.db.Query(
        `SELECT id, encounter_id, character_id, name, initiative, hp_current, hp_max,
                conditions_json, is_player,
                damage_superficial, damage_aggravated, willpower_superficial, willpower_aggravated, hunger
         FROM combatants WHERE encounter_id = ? ORDER BY initiative DESC`,
        encounterID,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var out []Combatant
    for rows.Next() {
        var c Combatant
        var isPlayer int
        if err := rows.Scan(
            &c.ID, &c.EncounterID, &c.CharacterID, &c.Name,
            &c.Initiative, &c.HPCurrent, &c.HPMax, &c.ConditionsJSON, &isPlayer,
            &c.DamageSuperficial, &c.DamageAggravated,
            &c.WillpowerSuperficial, &c.WillpowerAggravated, &c.Hunger,
        ); err != nil {
            return nil, err
        }
        c.IsPlayer = isPlayer == 1
        out = append(out, c)
    }
    return out, rows.Err()
}
```

Add the new method at the bottom of queries_combat.go:

```go
// UpdateCombatantVtMDamage applies VtM V5 damage to a combatant.
// Superficial damage is halved (round up) for vampires before applying.
// Overflow superficial (beyond health_max) converts to aggravated.
// isVampire=true applies halving; false applies superficial raw (for mortals).
func (d *DB) UpdateCombatantVtMDamage(id int64, superficialIn, aggravatedIn int, isVampire bool) error {
    var cur Combatant
    var isPlayer int
    err := d.db.QueryRow(
        `SELECT id, encounter_id, character_id, name, initiative, hp_current, hp_max,
                conditions_json, is_player,
                damage_superficial, damage_aggravated, willpower_superficial, willpower_aggravated, hunger
         FROM combatants WHERE id = ?`, id,
    ).Scan(
        &cur.ID, &cur.EncounterID, &cur.CharacterID, &cur.Name,
        &cur.Initiative, &cur.HPCurrent, &cur.HPMax, &cur.ConditionsJSON, &isPlayer,
        &cur.DamageSuperficial, &cur.DamageAggravated,
        &cur.WillpowerSuperficial, &cur.WillpowerAggravated, &cur.Hunger,
    )
    if err != nil {
        return err
    }
    cur.IsPlayer = isPlayer == 1

    // Halve superficial for vampires (round up)
    applied := superficialIn
    if isVampire && applied > 0 {
        applied = (applied + 1) / 2
    }
    newSuperficial := cur.DamageSuperficial + applied

    // Overflow superficial converts to aggravated
    newAggravated := cur.DamageAggravated + aggravatedIn
    if newSuperficial > cur.HPMax {
        overflow := newSuperficial - cur.HPMax
        newSuperficial = cur.HPMax
        newAggravated += overflow
    }
    if newAggravated > cur.HPMax {
        newAggravated = cur.HPMax
    }

    _, err = d.db.Exec(
        `UPDATE combatants SET damage_superficial = ?, damage_aggravated = ? WHERE id = ?`,
        newSuperficial, newAggravated, id,
    )
    return err
}
```

- [ ] **Step 4: Run test to confirm it passes**

```bash
cd /home/digitalghost/projects/inkandbone && go test ./internal/db/ -run TestUpdateCombatantVtMDamage -v 2>&1 | tail -20
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/db/queries_combat.go internal/db/queries_combat_test.go
git commit -m "feat(db): VtM combatant damage fields + UpdateCombatantVtMDamage with halving/overflow"
```

---

## Task 11: routes.go — Currency skip + VtM tension keywords + [VtM MECHANICS] block

**Files:**
- Modify: `internal/api/routes.go`

Three targeted changes; do them one at a time.

### 11a: Currency skip for VtM

- [ ] **Step 1: Find the currency skip block**

It's at routes.go ~line 3002–3010:
```go
if rs.Name == "wrath_glory" {
    return
}
```

Change it to also skip VtM:

```go
if rs.Name == "wrath_glory" || rs.Name == "vtm" {
    return
}
```

- [ ] **Step 2: Run tests**

```bash
cd /home/digitalghost/projects/inkandbone && go test ./internal/api/ 2>&1 | tail -10
```
Expected: all tests pass.

### 11b: VtM tension keywords

- [ ] **Step 3: Find crisisRE definition** (~line 2965)

The current regex:
```go
var crisisRE = regexp.MustCompile(
    `\b(critical\s+failure|disaster|catastrophe|ambush|betrayal|dying|wounded|doomed|cornered|overwhelmed)\b`,
)
```

This is a global regex used for all systems. We need to preserve it but add VtM keywords only for VtM sessions.

Find `autoUpdateTension` (~line 2972) and modify it to also check VtM-specific keywords when the session is a VtM session:

```go
func (s *Server) autoUpdateTension(sessionID int64, gmText string) {
    lower := strings.ToLower(gmText)

    matched := crisisRE.MatchString(lower)

    // For VtM sessions, also check VtM-specific crisis keywords.
    if !matched {
        if sess, err := s.db.GetSession(sessionID); err == nil && sess != nil {
            if camp, err := s.db.GetCampaign(sess.CampaignID); err == nil && camp != nil {
                if rs, err := s.db.GetRuleset(camp.RulesetID); err == nil && rs != nil && rs.Name == "vtm" {
                    matched = vtmCrisisRE.MatchString(lower)
                }
            }
        }
    }

    if !matched {
        return
    }

    current, err := s.db.GetTension(sessionID)
    if err != nil {
        return
    }
    newLevel := current + 1
    _ = s.db.UpdateTension(sessionID, newLevel)
    s.bus.Publish(Event{Type: EventTensionUpdated, Payload: map[string]any{
        "session_id":    sessionID,
        "tension_level": newLevel,
    }})
}
```

Add the VtM crisis regex near the existing `crisisRE` definition:

```go
// vtmCrisisRE matches VtM-specific crisis keywords at word boundaries.
var vtmCrisisRE = regexp.MustCompile(
    `\b(frenzy|the beast|torpor|blood hunt|diablerie|masquerade breach|daybreak|sunrise)\b`,
)
```

### 11c: [VtM MECHANICS] block in buildWorldContext

- [ ] **Step 4: Add VtM MECHANICS injection**

In `buildWorldContext`, after the `[/W&G MECHANICS]` block (~line 1004), add:

```go
    // VtM V5: inject live Hunger/Humanity/Blood Potency and identity fields.
    if camp, err := s.db.GetCampaign(sess.CampaignID); err == nil && camp != nil {
        if rs, err := s.db.GetRuleset(camp.RulesetID); err == nil && rs != nil && rs.Name == "vtm" {
            sb.WriteString("[VtM MECHANICS]\n")
            if charIDStr, err := s.db.GetSetting("active_character_id"); err == nil && charIDStr != "" {
                if charID, err := strconv.ParseInt(charIDStr, 10, 64); err == nil {
                    if char, err := s.db.GetCharacter(charID); err == nil && char != nil && char.DataJSON != "" {
                        var stats map[string]any
                        if err := json.Unmarshal([]byte(char.DataJSON), &stats); err == nil {
                            getInt := func(key string) int {
                                if v, ok := stats[key]; ok {
                                    switch n := v.(type) {
                                    case int:
                                        return n
                                    case float64:
                                        return int(n)
                                    }
                                }
                                return 0
                            }
                            getString := func(key string) string {
                                if v, ok := stats[key].(string); ok {
                                    return v
                                }
                                return ""
                            }
                            hunger := getInt("hunger")
                            humanity := getInt("humanity")
                            bp := getInt("blood_potency")
                            stains := getInt("stains")
                            hMax := getInt("health_max")
                            hSup := getInt("health_superficial")
                            hAgg := getInt("health_aggravated")
                            wMax := getInt("willpower_max")
                            wSup := getInt("willpower_superficial")
                            predType := getString("predator_type")
                            clan := getString("clan")
                            fmt.Fprintf(&sb, "Hunger: %d/5 | Humanity: %d | Blood Potency: %d\n", hunger, humanity, bp)
                            fmt.Fprintf(&sb, "Predator Type: %s | Clan: %s\n", predType, clan)
                            fmt.Fprintf(&sb, "Health: %d/%d (%d Superficial, %d Aggravated)\n", hMax-hSup-hAgg, hMax, hSup, hAgg)
                            fmt.Fprintf(&sb, "Willpower: %d/%d (%d Superficial)\n", wMax-wSup, wMax, wSup)
                            fmt.Fprintf(&sb, "Stains: %d\n", stains)
                            if hunger >= 4 {
                                sb.WriteString("WARNING: Hunger is critical. Frenzy risk is high.\n")
                            }
                        }
                    }
                }
            }
            sb.WriteString("[/VtM MECHANICS]\n")
        }
    }
```

Also add VtM identity fields to the character identity block that already reads `clan`, `predator_type`, etc. Find the loop at ~line 827:

```go
for _, field := range []string{"archetype", "class", "race", "faction", "keywords", "species", "metatype", "playbook", "culture"} {
```

Change it to also include VtM identity fields:

```go
for _, field := range []string{"archetype", "class", "race", "faction", "keywords", "species", "metatype", "playbook", "culture", "clan", "predator_type", "sect"} {
```

- [ ] **Step 5: Run all api tests**

```bash
cd /home/digitalghost/projects/inkandbone && go test ./internal/api/ 2>&1 | tail -10
```
Expected: all pass.

- [ ] **Step 6: Commit**

```bash
git add internal/api/routes.go
git commit -m "feat(api): VtM currency skip, VtM tension keywords, [VtM MECHANICS] world context block"
```

---

## Task 12: routes.go — autoUpdateMasquerade goroutine + stain detection

**Files:**
- Modify: `internal/api/routes.go`

### 12a: autoUpdateMasquerade goroutine

- [ ] **Step 1: Add masquerade breach regex near vtmCrisisRE**

```go
// vtmMajorBreachRE matches major Masquerade breach keywords.
var vtmMajorBreachRE = regexp.MustCompile(
    `\b(caught on camera|viral|police|recorded|photographed|livestream|news crew)\b`,
)

// vtmModerateBreachRE matches moderate breach keywords.
var vtmModerateBreachRE = regexp.MustCompile(
    `\b(witnessed feeding|seen feeding|watched you feed|fangs exposed|transformation witnessed|seen your true form)\b`,
)

// vtmMinorBreachRE matches minor breach keywords.
var vtmMinorBreachRE = regexp.MustCompile(
    `\b(overheard|suspicious|noticed something|acting strange|too fast|too strong|inhuman)\b`,
)
```

- [ ] **Step 2: Add autoUpdateMasquerade function to routes.go** (place near autoUpdateTension)

```go
// autoUpdateMasquerade checks GM text for Masquerade breach keywords and decrements
// masquerade_integrity for VtM sessions. No-op for non-VtM sessions.
func (s *Server) autoUpdateMasquerade(ctx context.Context, sessionID int64, gmText string) {
    sess, err := s.db.GetSession(sessionID)
    if err != nil || sess == nil {
        return
    }
    camp, err := s.db.GetCampaign(sess.CampaignID)
    if err != nil || camp == nil {
        return
    }
    rs, err := s.db.GetRuleset(camp.RulesetID)
    if err != nil || rs == nil || rs.Name != "vtm" {
        return
    }

    lower := strings.ToLower(gmText)
    delta := 0
    if vtmMajorBreachRE.MatchString(lower) {
        delta = -3
    } else if vtmModerateBreachRE.MatchString(lower) {
        delta = -2
    } else if vtmMinorBreachRE.MatchString(lower) {
        delta = -1
    }
    if delta == 0 {
        return
    }

    current, err := s.db.GetMasqueradeIntegrity(sessionID)
    if err != nil {
        return
    }
    newLevel := current + delta
    _ = s.db.UpdateMasqueradeIntegrity(sessionID, newLevel)
    s.bus.Publish(Event{Type: EventSessionUpdated, Payload: map[string]any{
        "session_id":            sessionID,
        "masquerade_integrity":  newLevel,
    }})
}
```

- [ ] **Step 3: Wire autoUpdateMasquerade into handleGMRespondStream**

Find the goroutine launch block (~line 1366–1376) and add:

```go
    go s.autoUpdateMasquerade(context.Background(), id, fullText)
```

### 12b: Stain detection in autoUpdateCharacterStats

- [ ] **Step 4: Add stain detection at the top of autoUpdateCharacterStats**

`autoUpdateCharacterStats` is at ~line 1509. Add a VtM stain detection block at the top of the function, after getting the ruleset:

```go
    // VtM: detect Humanity-violating acts and increment stains.
    if ruleset.Name == "vtm" {
        s.detectAndApplyVtMStains(ctx, playerAction+" "+gmText)
        return
    }
```

Add the helper function near autoUpdateMasquerade:

```go
// stainTriggerRE matches acts that cost Stains in VtM V5.
var stainTriggerRE = regexp.MustCompile(
    `\b(feeding|fed from|forced feeding|killing|killed|diablerie|diablerized|compulsion|breaking.*conviction|violated.*conviction)\b`,
)

// detectAndApplyVtMStains scans text for Humanity-violating acts and adds Stains.
func (s *Server) detectAndApplyVtMStains(ctx context.Context, text string) {
    if !stainTriggerRE.MatchString(strings.ToLower(text)) {
        return
    }
    charIDStr, err := s.db.GetSetting("active_character_id")
    if err != nil || charIDStr == "" {
        return
    }
    charID, err := strconv.ParseInt(charIDStr, 10, 64)
    if err != nil {
        return
    }
    char, err := s.db.GetCharacter(charID)
    if err != nil || char == nil || char.DataJSON == "" {
        return
    }
    var stats map[string]any
    if err := json.Unmarshal([]byte(char.DataJSON), &stats); err != nil {
        return
    }
    current := 0
    if v, ok := stats["stains"]; ok {
        switch n := v.(type) {
        case int:
            current = n
        case float64:
            current = int(n)
        }
    }
    if current >= 10 {
        return
    }
    stats["stains"] = current + 1
    b, err := json.Marshal(stats)
    if err != nil {
        return
    }
    updates := map[string]any{"stains": current + 1}
    if err := s.db.PatchCharacter(charID, updates); err != nil {
        _ = b
        return
    }
    s.bus.Publish(Event{Type: EventCharacterUpdated, Payload: map[string]any{
        "id": charID,
    }})
}
```

Note: `PatchCharacter` is the existing DB method used by `handlePatchCharacter` — verify its signature in `queries_core.go` before implementing. It likely takes `(id int64, updates map[string]any) error`. If the method is named differently, use the correct name.

- [ ] **Step 5: Run all api tests**

```bash
cd /home/digitalghost/projects/inkandbone && go test ./internal/api/ 2>&1 | tail -15
```
Expected: all pass.

- [ ] **Step 6: Commit**

```bash
git add internal/api/routes.go
git commit -m "feat(api): autoUpdateMasquerade goroutine + VtM stain detection in autoUpdateCharacterStats"
```

---

## Task 13: routes.go — Rouse Check + Hunger dice + Blood Surge

**Files:**
- Modify: `internal/api/routes.go`

### 13a: Rouse Check command

- [ ] **Step 1: Add rouseCheckRE near the top of routes.go (with other var blocks)**

```go
// rouseCheckRE matches the player's /rouse or "rouse check" command.
var rouseCheckRE = regexp.MustCompile(`(?i)\b(rouse\s+check|/rouse)\b`)

// bloodSurgeRE matches the /surge command.
var bloodSurgeRE = regexp.MustCompile(`(?i)\b(/surge|blood\s+surge)\b`)
```

- [ ] **Step 2: Add handleVtMRouseCheck function**

```go
// handleVtMRouseCheck performs a Rouse Check for a VtM character.
// Rolls 1d10; 6+ = no Hunger change; 1-5 = Hunger +1.
// At Hunger 5, does not increase further but flags a Frenzy risk.
// Returns a string describing the result for injection into GM context.
func (s *Server) handleVtMRouseCheck(ctx context.Context, sessionID int64) string {
    charIDStr, err := s.db.GetSetting("active_character_id")
    if err != nil || charIDStr == "" {
        return ""
    }
    charID, err := strconv.ParseInt(charIDStr, 10, 64)
    if err != nil {
        return ""
    }
    char, err := s.db.GetCharacter(charID)
    if err != nil || char == nil || char.DataJSON == "" {
        return ""
    }
    var stats map[string]any
    if err := json.Unmarshal([]byte(char.DataJSON), &stats); err != nil {
        return ""
    }

    currentHunger := 0
    if v, ok := stats["hunger"]; ok {
        switch n := v.(type) {
        case int:
            currentHunger = n
        case float64:
            currentHunger = int(n)
        }
    }

    roll := mathrand.Intn(10) + 1
    _, _ = s.db.LogDiceRoll(sessionID, "1d10 (Rouse Check)", roll, "[]")
    s.bus.Publish(Event{Type: EventDiceRolled, Payload: map[string]any{
        "session_id": sessionID,
        "expression": "1d10 (Rouse Check)",
        "result":     roll,
    }})

    if roll >= 6 {
        return fmt.Sprintf("[ROUSE CHECK] Result: %d — Success. Hunger unchanged at %d.", roll, currentHunger)
    }

    // Hunger increases
    if currentHunger >= 5 {
        return fmt.Sprintf("[ROUSE CHECK] Result: %d — Failed. Hunger already at 5. FRENZY RISK: The character must resist a Hunger Frenzy (Composure + Resolve, difficulty 3).", roll)
    }

    newHunger := currentHunger + 1
    stats["hunger"] = newHunger
    if err := s.db.PatchCharacter(charID, map[string]any{"hunger": newHunger}); err == nil {
        s.bus.Publish(Event{Type: EventCharacterUpdated, Payload: map[string]any{"id": charID}})
    }

    msg := fmt.Sprintf("[ROUSE CHECK] Result: %d — Failed. Hunger increases to %d.", roll, newHunger)
    if newHunger >= 4 {
        msg += " The Beast strains against the cage. Frenzy risk is elevated."
    }
    return msg
}
```

- [ ] **Step 3: Add Blood Surge helper**

```go
// bloodPotencyBonusDice returns the bonus dice granted by Blood Surge for a given Blood Potency.
func bloodPotencyBonusDice(bp int) int {
    switch {
    case bp >= 10:
        return 4
    case bp >= 7:
        return 3
    case bp >= 4:
        return 2
    default:
        return 1
    }
}

// handleVtMBloodSurge performs a Rouse Check and if successful returns bonus dice count.
// Returns a string for injection into GM context.
func (s *Server) handleVtMBloodSurge(ctx context.Context, sessionID int64) string {
    rouseResult := s.handleVtMRouseCheck(ctx, sessionID)

    // Read current blood_potency for the bonus dice count
    charIDStr, _ := s.db.GetSetting("active_character_id")
    charID, _ := strconv.ParseInt(charIDStr, 10, 64)
    char, err := s.db.GetCharacter(charID)
    if err != nil || char == nil {
        return rouseResult
    }
    var stats map[string]any
    _ = json.Unmarshal([]byte(char.DataJSON), &stats)
    bp := 1
    if v, ok := stats["blood_potency"]; ok {
        switch n := v.(type) {
        case int:
            bp = n
        case float64:
            bp = int(n)
        }
    }
    bonus := bloodPotencyBonusDice(bp)
    return rouseResult + fmt.Sprintf(" [BLOOD SURGE] Add %d bonus dice to the next roll this turn (Blood Potency %d).", bonus, bp)
}
```

- [ ] **Step 4: Wire Rouse/Blood Surge into the GM respond stream handler**

In `handleGMRespondStream` (~line 1306), after `checkAndExecuteRoll`, add a VtM command intercept. Find the block that calls `checkAndExecuteRoll`:

```go
    roll := s.checkAndExecuteRoll(r.Context(), id, lastPlayerMsg)
```

Add before this line:

```go
    // VtM: intercept /rouse and /surge commands before the normal roll check.
    var vtmCommandResult string
    if ruleset != nil && ruleset.Name == "vtm" {
        lower := strings.ToLower(lastPlayerMsg)
        if bloodSurgeRE.MatchString(lower) {
            vtmCommandResult = s.handleVtMBloodSurge(r.Context(), id)
        } else if rouseCheckRE.MatchString(lower) {
            vtmCommandResult = s.handleVtMRouseCheck(r.Context(), id)
        }
    }
```

Then inject `vtmCommandResult` into the system prompt (add to `worldCtx` before it's appended):

```go
    if vtmCommandResult != "" {
        worldCtx += "\n" + vtmCommandResult
    }
```

Note: `ruleset` needs to be in scope here. Check if `handleGMRespondStream` already fetches the ruleset — if not, look up the pattern for getting it from `camp.RulesetID`.

- [ ] **Step 5: Run all api tests**

```bash
cd /home/digitalghost/projects/inkandbone && go test ./internal/api/ 2>&1 | tail -15
```
Expected: all pass.

- [ ] **Step 6: Commit**

```bash
git add internal/api/routes.go
git commit -m "feat(api): VtM Rouse Check, Blood Surge commands + Hunger dice injection"
```

---

## Task 14: routes_phase_d.go + server.go — Masquerade integrity endpoints

**Files:**
- Modify: `internal/api/routes_phase_d.go`
- Modify: `internal/api/server.go`

- [ ] **Step 1: Write the failing test**

In `internal/api/routes_phase_d_test.go`, append:

```go
func TestGetMasqueradeIntegrity_default(t *testing.T) {
    s := newTestServer(t)
    campID, sessID := seedCampaign(t, s.db)
    _ = campID

    req := httptest.NewRequest(http.MethodGet, "/api/sessions/"+strconv.FormatInt(sessID, 10)+"/masquerade", nil)
    w := httptest.NewRecorder()
    s.ServeHTTP(w, req)
    assert.Equal(t, http.StatusOK, w.Code)

    var resp map[string]any
    require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
    assert.Equal(t, float64(10), resp["masquerade_integrity"])
}

func TestPatchMasqueradeIntegrity(t *testing.T) {
    s := newTestServer(t)
    campID, sessID := seedCampaign(t, s.db)
    _ = campID

    body, _ := json.Marshal(map[string]any{"masquerade_integrity": 7})
    req := httptest.NewRequest(http.MethodPatch, "/api/sessions/"+strconv.FormatInt(sessID, 10)+"/masquerade", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    s.ServeHTTP(w, req)
    assert.Equal(t, http.StatusOK, w.Code)

    // Verify
    req2 := httptest.NewRequest(http.MethodGet, "/api/sessions/"+strconv.FormatInt(sessID, 10)+"/masquerade", nil)
    w2 := httptest.NewRecorder()
    s.ServeHTTP(w2, req2)
    var resp map[string]any
    require.NoError(t, json.Unmarshal(w2.Body.Bytes(), &resp))
    assert.Equal(t, float64(7), resp["masquerade_integrity"])
}
```

- [ ] **Step 2: Run test to confirm it fails**

```bash
cd /home/digitalghost/projects/inkandbone && go test ./internal/api/ -run TestGetMasqueradeIntegrity -run TestPatchMasqueradeIntegrity -v 2>&1 | tail -20
```
Expected: FAIL — no route registered.

- [ ] **Step 3: Add handlers to routes_phase_d.go**

```go
func (s *Server) handleGetMasqueradeIntegrity(w http.ResponseWriter, r *http.Request) {
    idStr := r.PathValue("id")
    sessionID, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "invalid session id", http.StatusBadRequest)
        return
    }
    level, err := s.db.GetMasqueradeIntegrity(sessionID)
    if err != nil {
        http.Error(w, "db error", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]any{"masquerade_integrity": level}) //nolint:errcheck
}

func (s *Server) handlePatchMasqueradeIntegrity(w http.ResponseWriter, r *http.Request) {
    idStr := r.PathValue("id")
    sessionID, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "invalid session id", http.StatusBadRequest)
        return
    }
    var body struct {
        MasqueradeIntegrity *int `json:"masquerade_integrity"`
    }
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.MasqueradeIntegrity == nil {
        http.Error(w, "masquerade_integrity required", http.StatusBadRequest)
        return
    }
    if err := s.db.UpdateMasqueradeIntegrity(sessionID, *body.MasqueradeIntegrity); err != nil {
        http.Error(w, "db error", http.StatusInternalServerError)
        return
    }
    s.bus.Publish(Event{Type: EventSessionUpdated, Payload: map[string]any{
        "session_id":           sessionID,
        "masquerade_integrity": *body.MasqueradeIntegrity,
    }})
    w.WriteHeader(http.StatusOK)
}
```

- [ ] **Step 4: Register routes in server.go**

Find the Phase D route block (~line 145) and add:

```go
    s.mux.HandleFunc("GET /api/sessions/{id}/masquerade", s.handleGetMasqueradeIntegrity)
    s.mux.HandleFunc("PATCH /api/sessions/{id}/masquerade", s.handlePatchMasqueradeIntegrity)
```

- [ ] **Step 5: Run test to confirm it passes**

```bash
cd /home/digitalghost/projects/inkandbone && go test ./internal/api/ -run TestGetMasquerade -v -run TestPatchMasquerade 2>&1 | tail -20
```
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/api/routes_phase_d.go internal/api/server.go
git commit -m "feat(api): VtM masquerade_integrity GET/PATCH endpoints"
```

---

## Task 15: Add VtM scene tags to sceneTagKeywords

**Files:**
- Modify: `internal/api/routes.go`

- [ ] **Step 1: Add VtM tags to sceneTagKeywords map**

Find the `sceneTagKeywords` map (~line 2896) and add four VtM entries:

```go
    "elysium":    {"elysium", "court of elysium", "neutral ground", "the salon", "gathering of kindred"},
    "haven":      {"haven", "lair", "sanctuary", "your haven", "safe house", "feeding ground"},
    "hunt":       {"hunting", "stalking", "feeding ground", "prey", "the hunt", "the rack"},
    "masquerade": {"masquerade breach", "mortal witnesses", "humans watching", "public eye", "crowd of mortals"},
```

- [ ] **Step 2: Run all api tests**

```bash
cd /home/digitalghost/projects/inkandbone && go test ./internal/api/ 2>&1 | tail -10
```
Expected: all pass.

- [ ] **Step 3: Commit**

```bash
git add internal/api/routes.go
git commit -m "feat(api): add VtM scene tags to sceneTagKeywords (elysium, haven, hunt, masquerade)"
```

---

## Task 16: CharacterSheetPanel.tsx — VtM character sheet

**Files:**
- Modify: `web/src/CharacterSheetPanel.tsx`

This is a full VtM rendering branch inside the existing component. VtM is detected via `ruleset?.name?.toLowerCase() === 'vtm'`.

- [ ] **Step 1: Add VtM detection and early return for VtM rendering**

After the existing `if (!character) return null` line, add:

```tsx
  const isVtM = ruleset?.name?.toLowerCase() === 'vtm'
  if (isVtM) {
    return <VtMCharacterSheet character={character} fields={fields} onChange={handleChange} />
  }
```

Then extract a `VtMCharacterSheet` component in the same file, above `CharacterSheetPanel`:

```tsx
interface VtMSheetProps {
  character: Character
  fields: Record<string, string>
  onChange: (key: string, value: string) => void
}

function PipRow({ label, value, max, onChange, color = 'gold' }: {
  label: string; value: number; max: number; onChange?: (v: number) => void; color?: string
}) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', marginBottom: '3px' }}>
      <span style={{ fontSize: '10px', textTransform: 'uppercase', letterSpacing: '1px', color: 'var(--gold-dim)', width: '110px', flexShrink: 0 }}>{label}</span>
      <div style={{ display: 'flex', gap: '3px' }}>
        {Array.from({ length: max }, (_, i) => (
          <div
            key={i}
            onClick={() => onChange?.(i < value ? i : i + 1)}
            style={{
              width: 12, height: 12, borderRadius: '50%',
              background: i < value ? (color === 'red' ? '#c0392b' : 'var(--gold)') : 'transparent',
              border: `1px solid ${color === 'red' ? '#c0392b' : 'var(--gold-dim)'}`,
              cursor: onChange ? 'pointer' : 'default',
            }}
          />
        ))}
      </div>
    </div>
  )
}

function DamageTrack({ label, max, superficial, aggravated, onClickBox }: {
  label: string; max: number; superficial: number; aggravated: number
  onClickBox?: (index: number) => void
}) {
  return (
    <div style={{ marginBottom: '6px' }}>
      <span style={{ fontSize: '10px', textTransform: 'uppercase', letterSpacing: '1px', color: 'var(--gold-dim)' }}>{label}</span>
      <div style={{ display: 'flex', gap: '3px', marginTop: '3px' }}>
        {Array.from({ length: max }, (_, i) => {
          const fromRight = max - 1 - i
          const isAgg = fromRight < aggravated
          const isSup = !isAgg && fromRight < aggravated + superficial
          return (
            <div
              key={i}
              onClick={() => onClickBox?.(i)}
              style={{
                width: 14, height: 14, border: '1px solid var(--gold-dim)',
                background: isAgg ? '#8b0000' : isSup ? '#555' : 'transparent',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                fontSize: '9px', color: isAgg ? '#fff' : isSup ? '#ccc' : 'transparent',
                cursor: onClickBox ? 'pointer' : 'default',
              }}
            >
              {isAgg ? 'X' : isSup ? '/' : ''}
            </div>
          )
        })}
      </div>
    </div>
  )
}

function HungerTrack({ value, onChange }: { value: number; onChange: (v: number) => void }) {
  return (
    <div style={{ marginBottom: '10px' }}>
      <div style={{ fontSize: '10px', textTransform: 'uppercase', letterSpacing: '1px', color: '#c0392b', marginBottom: '4px' }}>Hunger</div>
      <div style={{ display: 'flex', gap: '4px' }}>
        {[1, 2, 3, 4, 5].map((i) => (
          <div
            key={i}
            onClick={() => onChange(i === value ? i - 1 : i)}
            style={{
              width: 20, height: 20,
              background: i <= value ? '#c0392b' : 'transparent',
              border: '1px solid #c0392b',
              cursor: 'pointer',
              animation: value >= 5 && i <= 5 ? 'pulse 1s infinite' : undefined,
            }}
          />
        ))}
      </div>
    </div>
  )
}

const VTM_PHYSICAL_ATTRS = ['strength', 'dexterity', 'stamina']
const VTM_SOCIAL_ATTRS = ['charisma', 'manipulation', 'composure']
const VTM_MENTAL_ATTRS = ['intelligence', 'wits', 'resolve']
const VTM_PHYSICAL_SKILLS = ['athletics', 'brawl', 'craft', 'drive', 'firearms', 'larceny', 'melee', 'stealth', 'survival']
const VTM_SOCIAL_SKILLS = ['animal_ken', 'etiquette', 'insight', 'intimidation', 'leadership', 'performance', 'persuasion', 'streetwise', 'subterfuge']
const VTM_MENTAL_SKILLS = ['academics', 'awareness', 'finance', 'investigation', 'medicine', 'occult', 'politics', 'technology']
const VTM_DISCIPLINES = ['animalism', 'auspex', 'blood_sorcery', 'celerity', 'dominate', 'fortitude', 'obfuscate', 'oblivion', 'potence', 'presence', 'protean']

function VtMCharacterSheet({ character, fields, onChange }: VtMSheetProps) {
  const labelStyle: React.CSSProperties = {
    fontSize: '9px', textTransform: 'uppercase', letterSpacing: '1.5px',
    color: 'var(--gold-dim)', fontFamily: 'var(--serif)'
  }
  const inputStyle: React.CSSProperties = {
    background: 'var(--surface)', border: '1px solid var(--border)',
    color: 'var(--text)', fontSize: '12px', padding: '0.15rem 0.3rem', width: '100%'
  }
  const sectionHead: React.CSSProperties = {
    fontSize: '11px', textTransform: 'uppercase', letterSpacing: '2px',
    color: 'var(--gold)', borderBottom: '1px solid var(--border)',
    paddingBottom: '3px', marginTop: '12px', marginBottom: '6px'
  }
  const n = (key: string) => parseInt(fields[key] ?? '0') || 0

  return (
    <>
      {/* Hunger */}
      <HungerTrack value={n('hunger')} onChange={(v) => onChange('hunger', String(v))} />

      {/* Core stats row */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '4px', marginBottom: '8px' }}>
        {(['humanity', 'blood_potency', 'stains'] as const).map((key) => (
          <label key={key} style={labelStyle}>
            {key.replace(/_/g, ' ')}
            <input type="number" value={fields[key] ?? ''} onChange={(e) => onChange(key, e.target.value)} style={{ ...inputStyle, marginTop: '2px' }} />
          </label>
        ))}
      </div>

      {/* Damage tracks */}
      <DamageTrack
        label="Health"
        max={n('health_max') || 4}
        superficial={n('health_superficial')}
        aggravated={n('health_aggravated')}
        onClickBox={(i) => {
          const max = n('health_max') || 4
          const fromRight = max - 1 - i
          const curAgg = n('health_aggravated')
          const curSup = n('health_superficial')
          if (fromRight < curAgg) {
            onChange('health_aggravated', String(Math.max(0, curAgg - 1)))
          } else if (fromRight < curAgg + curSup) {
            onChange('health_superficial', String(Math.max(0, curSup - 1)))
          } else {
            onChange('health_superficial', String(curSup + 1))
          }
        }}
      />
      <DamageTrack
        label="Willpower"
        max={n('willpower_max') || 3}
        superficial={n('willpower_superficial')}
        aggravated={n('willpower_aggravated')}
        onClickBox={(i) => {
          const max = n('willpower_max') || 3
          const fromRight = max - 1 - i
          const curAgg = n('willpower_aggravated')
          const curSup = n('willpower_superficial')
          if (fromRight < curAgg) {
            onChange('willpower_aggravated', String(Math.max(0, curAgg - 1)))
          } else if (fromRight < curAgg + curSup) {
            onChange('willpower_superficial', String(Math.max(0, curSup - 1)))
          } else {
            onChange('willpower_superficial', String(curSup + 1))
          }
        }}
      />

      {/* Attributes */}
      <div style={sectionHead}>Attributes</div>
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '0 12px' }}>
        {[
          { label: 'Physical', attrs: VTM_PHYSICAL_ATTRS },
          { label: 'Social', attrs: VTM_SOCIAL_ATTRS },
          { label: 'Mental', attrs: VTM_MENTAL_ATTRS },
        ].map(({ label, attrs }) => (
          <div key={label}>
            <div style={{ fontSize: '9px', color: 'var(--gold)', marginBottom: '4px' }}>{label}</div>
            {attrs.map((key) => (
              <PipRow
                key={key}
                label={key}
                value={n(key)}
                max={5}
                onChange={(v) => onChange(key, String(v))}
              />
            ))}
          </div>
        ))}
      </div>

      {/* Skills */}
      <div style={sectionHead}>Skills</div>
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '0 12px' }}>
        {[
          { label: 'Physical', skills: VTM_PHYSICAL_SKILLS },
          { label: 'Social', skills: VTM_SOCIAL_SKILLS },
          { label: 'Mental', skills: VTM_MENTAL_SKILLS },
        ].map(({ label, skills }) => (
          <div key={label}>
            <div style={{ fontSize: '9px', color: 'var(--gold)', marginBottom: '4px' }}>{label}</div>
            {skills.map((key) => (
              <PipRow
                key={key}
                label={key.replace(/_/g, ' ')}
                value={n(key)}
                max={5}
                onChange={(v) => onChange(key, String(v))}
              />
            ))}
          </div>
        ))}
      </div>

      {/* Disciplines */}
      <div style={sectionHead}>Disciplines</div>
      {VTM_DISCIPLINES.map((key) => (
        <PipRow
          key={key}
          label={key.replace(/_/g, ' ')}
          value={n(key)}
          max={5}
          onChange={(v) => onChange(key, String(v))}
        />
      ))}

      {/* Identity */}
      <div style={sectionHead}>Identity</div>
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '4px 8px' }}>
        {(['clan', 'predator_type', 'sect', 'generation', 'ambition', 'desire'] as const).map((key) => (
          <label key={key} style={{ ...labelStyle, display: 'flex', flexDirection: 'column', gap: '2px' }}>
            {key.replace(/_/g, ' ')}
            <input type="text" value={fields[key] ?? ''} onChange={(e) => onChange(key, e.target.value)} style={inputStyle} />
          </label>
        ))}
      </div>
      {(['convictions', 'touchstones', 'skill_specialties', 'merits_flaws', 'notes'] as const).map((key) => (
        <label key={key} style={{ ...labelStyle, display: 'flex', flexDirection: 'column', gap: '2px', marginTop: '4px' }}>
          {key.replace(/_/g, ' ')}
          <textarea
            value={fields[key] ?? ''}
            onChange={(e) => onChange(key, e.target.value)}
            style={{ ...inputStyle, fontFamily: 'inherit', resize: 'vertical', minHeight: '3rem' }}
          />
        </label>
      ))}
    </>
  )
}
```

- [ ] **Step 2: Add the `isVtM` branch inside `CharacterSheetPanel`**

In `CharacterSheetPanel`, after `if (!character) return null`, add:

```tsx
  const isVtM = ruleset?.name?.toLowerCase() === 'vtm'
  if (isVtM) {
    return <VtMCharacterSheet character={character} fields={fields} onChange={handleChange} />
  }
```

Note: `handleChange` already closes over `character`, `schema`, and `debounceRef`. For VtM, we bypass the `schema` dependency — `VtMCharacterSheet` handles its own field layout. The `patchCharacter` call in `handleChange` sends whatever keys are changed, which is correct.

However, `handleChange` uses `schema.forEach(...)` to build updates — this would only include schema fields. For VtM, replace `handleChange` usage:

Modify `handleChange` to send all current fields if no schema is available:

```tsx
  function handleChange(key: string, value: string) {
    const next = { ...fields, [key]: value }
    setFields(next)
    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => {
      if (schema.length > 0) {
        const updates: Record<string, unknown> = {}
        schema.forEach((f) => {
          const v = next[f.key]
          updates[f.key] = f.type === 'number' ? (v === '' ? null : Number(v)) : v
        })
        patchCharacter(character!.id, updates).catch(console.error)
      } else {
        // VtM (or any schema-free ruleset): patch only the changed key
        const numericKeys = new Set([
          'hunger','blood_potency','bane_severity','humanity','stains',
          'strength','dexterity','stamina','charisma','manipulation','composure',
          'intelligence','wits','resolve','health_max','health_superficial','health_aggravated',
          'willpower_max','willpower_superficial','willpower_aggravated',
          'athletics','brawl','craft','drive','firearms','larceny','melee','stealth','survival',
          'animal_ken','etiquette','insight','intimidation','leadership','performance',
          'persuasion','streetwise','subterfuge','academics','awareness','finance',
          'investigation','medicine','occult','politics','technology',
          'animalism','auspex','blood_sorcery','celerity','dominate','fortitude',
          'obfuscate','oblivion','potence','presence','protean',
        ])
        const updates: Record<string, unknown> = {
          [key]: numericKeys.has(key) ? (value === '' ? null : Number(value)) : value,
        }
        patchCharacter(character!.id, updates).catch(console.error)
      }
    }, 500)
  }
```

- [ ] **Step 3: Build the frontend**

```bash
cd /home/digitalghost/projects/inkandbone && make build 2>&1 | tail -20
```
Expected: build succeeds, no TypeScript errors.

- [ ] **Step 4: Commit**

```bash
git add web/src/CharacterSheetPanel.tsx
git commit -m "feat(ui): VtM character sheet — Hunger track, damage boxes, attributes, skills, disciplines"
```

---

## Task 17: App.tsx — VtM single-track ambient audio

**Files:**
- Modify: `web/src/App.tsx`

- [ ] **Step 1: Add ruleset name state to App.tsx**

Find where `ctx` is loaded and the ambient audio `useEffect`. Add ruleset name tracking:

After the existing state declarations (near the top of the `App` component), add:

```tsx
  const [rulesetName, setRulesetName] = useState<string | null>(null)
```

Add a `useEffect` that fetches ruleset name when campaign changes (add after the existing health check effect):

```tsx
  useEffect(() => {
    const rulesetId = ctx?.campaign?.ruleset_id
    if (rulesetId == null) {
      setRulesetName(null)
      return
    }
    fetchRuleset(rulesetId)
      .then((rs) => setRulesetName(rs.name.toLowerCase()))
      .catch(() => setRulesetName(null))
  }, [ctx?.campaign?.ruleset_id])
```

- [ ] **Step 2: Modify the ambient audio useEffect to branch on VtM**

Find the existing scene_tags ambient effect (~line 315):

```tsx
  useEffect(() => {
    const tags = ctx?.session?.scene_tags ?? ''
    const firstTag = tags.split(',').filter(Boolean)[0] ?? null
    setAmbientTrack(firstTag)
  }, [ctx?.session?.scene_tags])
```

Replace it with:

```tsx
  useEffect(() => {
    if (rulesetName === 'vtm') {
      // VtM uses a single fixed ambient track regardless of scene tags.
      setAmbientTrack('vtm/ambient')
      return
    }
    const tags = ctx?.session?.scene_tags ?? ''
    const firstTag = tags.split(',').filter(Boolean)[0] ?? null
    setAmbientTrack(firstTag)
  }, [ctx?.session?.scene_tags, rulesetName])
```

Add `fetchRuleset` to the imports from `./api` if not already imported.

- [ ] **Step 3: Build**

```bash
cd /home/digitalghost/projects/inkandbone && make build 2>&1 | tail -20
```
Expected: build succeeds.

- [ ] **Step 4: Commit**

```bash
git add web/src/App.tsx
git commit -m "feat(ui): VtM ambient audio — single fixed track bypasses scene tag system"
```

---

## Task 18: Full test suite

- [ ] **Step 1: Run the full test suite**

```bash
cd /home/digitalghost/projects/inkandbone && make test 2>&1 | grep -E "FAIL|ok|error" | head -30
```
Expected: all packages pass. The pre-existing W&G `rollWrathGloryStats` test failure in `internal/ruleset` is allowed to remain — it is not caused by this work.

- [ ] **Step 2: Verify binary builds cleanly**

```bash
cd /home/digitalghost/projects/inkandbone && make build 2>&1 | tail -5
```
Expected: `Build complete` or similar success message.

- [ ] **Step 3: Final commit**

```bash
git add -A
git status
# Review — only commit files that are not already committed
git commit -m "chore: VtM V5 overhaul — all tasks complete" --allow-empty
```
