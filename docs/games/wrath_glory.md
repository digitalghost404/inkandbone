# Warhammer 40,000: Wrath & Glory — First Adventure Walkthrough

## System Overview

Wrath & Glory is a tabletop RPG set in the grim darkness of the 41st millennium — a universe of endless war, xenos threats, heresy, and the undying faith of the Imperium of Man. You play agents of the Imperium (or, rarely, allies at the fringes of it), undertaking missions that range from exterminating a genestealer cult to investigating a planetary governor's possible corruption. The AI acts as your Game Master (Warden), narrating hive cities and void stations, voicing Inquisitors and corrupted cultists, and managing the brutal dice system of Wrath & Glory.

## Setting Up Your Campaign

1. **Create a campaign.** Use the `create_campaign` tool with ruleset `wrath_glory`. Name it after your mission or cell — "The Gilded Knife", "Purge at Hive Tertius", "The Warden's Reach".
2. **Create your character.** Use the `create_character` tool with just your character's name — ink & bone automatically rolls your seven Attributes (1d3+3 each), derives Initiative from Agility, Resilience from Toughness, Determination from Willpower, and sets starting Wounds and Shock. Tell the AI your Tier, Archetype, and Keywords so it can frame your role in the 41st Millennium. Adjust any values with `update_character`.
3. **Start a session.** Name it after the briefing — "Throne Warrant 114-Sigma", "The Silence on Ferrus IV", "What the Pict-Feeds Missed".

## Suggested Opening Prompt

> "I'm playing Sister Vael, a Sister of Battle (Tier 2) with Strength 4, Agility 4, Toughness 5, Intellect 3, Willpower 5, Fellowship 3. Wounds 10, Shock 5. I've been sent by my Inquisitor to investigate rumors of heretical preaching in a hive city's under-district. I arrive in plain clothes — difficult for a Sister, but necessary. The address is a disused promethium refinery."

The AI will describe the refinery and its approaches, and let you begin the investigation.

## Key Mechanics to Establish Early

- **Dice pool:** Roll a number of d6s equal to your Attribute + Skill. Count dice showing 4, 5, or 6 as successes (Icons). Most tests have a Difficulty (number of successes needed). Each extra Icon beyond the threshold is a Shift — use Shifts to enhance results.
- **Wrath and Glory:** Before rolling, take one die from your pool and make it the Wrath Die. If the Wrath Die shows a 6, it's an Exalted Icon — it counts double. If it shows a 1, add a Complication. The Warden has a Glory pool (starts at 0, grows through your actions and failures) and can spend Glory to add complications or empower enemies.
- **Combat:** Initiative is rolled (Initiative attribute + d6 pool). Attacks use Ballistic Skill or Weapon Skill + Agility vs. a target's Defence. Damage minus Resilience = Wounds taken. At 0 Wounds, roll on the Critical Wound table.
- **Corruption and Ruin:** Exposure to the warp generates Corruption. The Warden tracks Ruin — a campaign-level threat meter that rises as things go wrong. High Ruin summons escalating horrors.
- **Keywords:** Your character has Keywords (Imperium, Adeptus Sororitas, etc.) that unlock faction-specific abilities, determine allegiances, and affect how NPCs react to you.

## A Classic First Session Arc

**Act 1 — The Brief:** Your Inquisitor (or cell leader) issues orders. Establish your mission, your cover if any, and what resources you have. Gather initial intelligence with Awareness and Scholar rolls.

**Act 2 — The Investigation:** Infiltrate the location. Question witnesses, examine evidence. One tense encounter — a guard you must avoid, a cultist who spots your rosette, a locked door hiding something worse.

**Act 3 — The Confrontation:** Find the heresy at its source. One combat or dramatic resolution. The immediate threat is contained. A thread remains — a name, a symbol, a supply crate with an off-world origin. The Inquisition's work is never done.

## Tips

- State your Keywords and Archetype at session start — the Warden (AI) will use them to tailor the world's reaction.
- Glory spent by the Warden is not punishment — it's dramatic escalation. Embrace the chaos.
- Wrath & Glory supports mixed parties (humans and Space Marines together) by making high-Tier characters face proportionally larger threats. Let the AI calibrate.
- The Imperium is corrupt, bureaucratic, and brutal. Roleplay the moral weight of serving it.
- Faith Points and Wrath abilities are powerful and thematic. Ask the AI to remind you when they apply.
