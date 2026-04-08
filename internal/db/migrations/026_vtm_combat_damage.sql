-- 026_vtm_combat_damage.sql: Add VtM V5 damage tracking to combatants
ALTER TABLE combatants ADD COLUMN damage_superficial INTEGER NOT NULL DEFAULT 0;
ALTER TABLE combatants ADD COLUMN damage_aggravated INTEGER NOT NULL DEFAULT 0;
ALTER TABLE combatants ADD COLUMN willpower_superficial INTEGER NOT NULL DEFAULT 0;
ALTER TABLE combatants ADD COLUMN willpower_aggravated INTEGER NOT NULL DEFAULT 0;
ALTER TABLE combatants ADD COLUMN hunger INTEGER NOT NULL DEFAULT 0;
