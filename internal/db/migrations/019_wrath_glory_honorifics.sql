-- Append honorific/address-form guidance to the Wrath & Glory GM context.
-- Space Marines are always male (Brother) and Sisters of Battle always "Sister".
-- This prevents the model from misgendering Astartes characters due to name inference.
UPDATE rulesets
SET gm_context = gm_context || '

HONORIFICS — MANDATORY: Address characters strictly by archetype, never by name inference.
- Adeptus Astartes / Space Marines: exclusively male warriors. Always "Brother [Name]". NEVER "sister".
- Adepta Sororitas / Sisters of Battle: exclusively female. Always "Sister [Name]". NEVER "brother".
- Inquisitors: "Inquisitor [Name]".
- Commissars: "Commissar [Name]".
- Tech-Priests: "Adept [Name]" or "Magos [Name]" depending on rank.
- All others: use the character''s name or rank, not a gendered religious honorific unless it fits their archetype exactly.
Violating these honorifics breaks immersion and contradicts established 40K lore. They are not optional.'
WHERE name = 'wrath_glory';
