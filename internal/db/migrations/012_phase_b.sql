-- 012_phase_b.sql: NPC personality fields on world_notes
ALTER TABLE world_notes ADD COLUMN personality_json TEXT NOT NULL DEFAULT '';
