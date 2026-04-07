-- Reduce W&G response length from 3-5 to 3-4 paragraphs.
-- Migration 021 set "3-5 paragraphs" but model was still producing 5-6.
UPDATE rulesets
SET gm_context = REPLACE(
    gm_context,
    '- LENGTH: Every narrative response must be 3-5 paragraphs. No more. Stop writing after the fifth paragraph under any circumstances. Do not pad beyond 5 paragraphs.',
    '- LENGTH: Every narrative response must be 3-4 paragraphs. No more. Stop writing after the fourth paragraph under any circumstances. Do not pad beyond 4 paragraphs.'
)
WHERE name = 'wrath_glory';
