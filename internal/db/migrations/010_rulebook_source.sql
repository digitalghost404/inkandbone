-- Add source column to track which book each chunk came from.
-- Allows multiple books per ruleset (core + expansions) without overwriting each other.
ALTER TABLE rulebook_chunks ADD COLUMN source TEXT NOT NULL DEFAULT 'Core Rulebook';
