-- 011_phase_a.sql

-- Feature 4: session scratchpad
ALTER TABLE sessions ADD COLUMN notes TEXT NOT NULL DEFAULT '';

-- Feature 13: initiative tracker 2.0 — persist active turn index across refreshes
ALTER TABLE combat_encounters ADD COLUMN active_turn_index INTEGER NOT NULL DEFAULT 0;

-- Feature 1: sub-task objectives — one level of nesting only
ALTER TABLE objectives ADD COLUMN parent_id INTEGER REFERENCES objectives(id);

-- Feature 6: XP / milestone log
CREATE TABLE IF NOT EXISTS xp_log (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  session_id INTEGER NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
  note       TEXT NOT NULL,
  amount     INTEGER,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
