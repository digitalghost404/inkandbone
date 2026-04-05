-- 014_phase_e.sql: Scene tags for audio atmosphere
ALTER TABLE sessions ADD COLUMN scene_tags TEXT NOT NULL DEFAULT '';
