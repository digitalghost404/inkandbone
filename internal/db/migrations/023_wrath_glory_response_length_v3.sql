-- Increase W&G response length from 3-4 to 4-5 paragraphs.
UPDATE rulesets
SET gm_context = REPLACE(
    gm_context,
    '- LENGTH: Every narrative response must be 3-4 paragraphs. No more. Stop writing after the fourth paragraph under any circumstances. Do not pad beyond 4 paragraphs.',
    '- LENGTH: Every narrative response must be 4-5 paragraphs. No more. Stop writing after the fifth paragraph under any circumstances. Do not pad beyond 5 paragraphs.'
)
WHERE name = 'wrath_glory';
