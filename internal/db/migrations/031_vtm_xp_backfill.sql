-- 031_vtm_xp_backfill.sql: Backfill xp=0 into existing VtM characters that
-- were created before migration 029 added xp to the schema. Without this,
-- autoUpdateCharacterStats silently drops XP awards because the field doesn't
-- exist in the character's data_json (the patch loop skips unknown keys).
UPDATE characters
SET data_json = json_set(data_json, '$.xp', 0)
WHERE campaign_id IN (
    SELECT c.id FROM campaigns c
    JOIN rulesets r ON r.id = c.ruleset_id
    WHERE r.name = 'vtm'
)
AND json_extract(data_json, '$.xp') IS NULL
AND data_json != ''
AND data_json != '{}';
