# D&D 5th Edition — First Adventure Walkthrough

## System Overview

Dungeons & Dragons 5th Edition is the world's most popular tabletop RPG. The AI acts as your Dungeon Master, narrating the world, voicing NPCs, adjudicating rules, and running combat. You describe your character's actions; the AI tells you what happens.

## Setting Up Your Campaign

1. **Create a campaign.** Use the `create_campaign` tool with ruleset `dnd5e`. Name it something evocative — "The Sunken Coast", "Shattered Crown", "Embers of the Empire".
2. **Create your character.** Use the `create_character` tool with just your character's name — ink & bone automatically rolls your six ability scores (4d6 drop lowest), sets level to 1, HP to 10, and AC to 10. Tell the AI your race, class, and background so it can narrate appropriately. Adjust any values with `update_character`.
3. **Start a session.** Name your first session something like "The Road to Thornwall" or "A Stranger in Neverwinter".

## Suggested Opening Prompt

Tell the AI your character's background and how they arrived at the adventure's starting point:

> "I'm playing Kael, a human Fighter (soldier background) arriving at the town of Thornwall after a week on the road. I've heard there's work — something about missing villagers and lights in the old keep. I walk into the first tavern I see."

The AI will set the scene, introduce NPCs, and let the adventure develop naturally from there.

## Key Mechanics to Establish Early

- **Ability checks:** Tell the AI which skill you want to use ("I try to intimidate the guard — rolling Intimidation"). The AI will call for a DC and interpret your roll.
- **Combat:** When a fight starts, tell the AI you're initiating combat. Describe your action each round (attack, spell, dash, hide, etc.) and roll your dice.
- **Advantage/Disadvantage:** If a situation calls for it, the AI will tell you to roll with advantage (roll twice, take higher) or disadvantage (take lower).
- **Short and long rests:** Ask the AI when it makes sense to rest. Short rests recover some HP via Hit Dice; long rests fully restore HP and spell slots.
- **Death saving throws:** If you drop to 0 HP, roll a d20 each turn — three successes stabilizes you, three failures means death.

## A Classic First Session Arc

**Act 1 — The Tavern Hook:** The innkeeper or a worried farmer approaches you. People have gone missing near the old keep on the hill. A reward is offered.

**Act 2 — Investigation:** Talk to locals, gather clues, maybe discover the keep is occupied by goblin scouts — or something worse.

**Act 3 — The Keep:** Clear the first floor. Find a prisoner. Discover a note hinting at a larger threat. End the session on a cliffhanger.

## Tips

- Ask the AI to describe the environment before acting — "What do I see in this room?"
- If you forget a rule, ask the AI: "How does the Shove action work?"
- Track your spell slots, hit dice, and any limited-use abilities in the journal tab.
- Let the AI surprise you. Don't plan too far ahead.
