-- Fix conflicting length directive in W&G gm_context.
-- Migration 020 set "at least 3 substantial paragraphs" which overrides the base
-- system prompt's "exactly 2-3 paragraphs" hard limit. Replace the LENGTH line only.
UPDATE rulesets
SET gm_context = REPLACE(
    gm_context,
    '- LENGTH: Every narrative response must be at least 3 substantial paragraphs. Short player inputs still deserve full scene descriptions. Never truncate a scene.',
    '- LENGTH: Every narrative response must be 3-5 paragraphs. No more. Stop writing after the fifth paragraph under any circumstances. Do not pad beyond 5 paragraphs.'
)
WHERE name = 'wrath_glory';
