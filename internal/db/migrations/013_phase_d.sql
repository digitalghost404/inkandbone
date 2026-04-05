-- 013_phase_d.sql: Oracle tables, tension tracking, character relationships

-- Oracle tables: random prompts for narrative inspiration
CREATE TABLE IF NOT EXISTS oracle_tables (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ruleset_id INTEGER REFERENCES rulesets(id) ON DELETE SET NULL,
    table_name TEXT NOT NULL,
    roll_min INTEGER NOT NULL,
    roll_max INTEGER NOT NULL,
    result TEXT NOT NULL
);

-- Seed generic Action oracle (1-50)
INSERT INTO oracle_tables (ruleset_id, table_name, roll_min, roll_max, result) VALUES
(NULL, 'action', 1, 2, 'Attack'),
(NULL, 'action', 3, 4, 'Defend'),
(NULL, 'action', 5, 6, 'Negotiate'),
(NULL, 'action', 7, 8, 'Flee'),
(NULL, 'action', 9, 10, 'Investigate'),
(NULL, 'action', 11, 12, 'Deceive'),
(NULL, 'action', 13, 14, 'Persuade'),
(NULL, 'action', 15, 16, 'Intimidate'),
(NULL, 'action', 17, 18, 'Aid'),
(NULL, 'action', 19, 20, 'Hinder'),
(NULL, 'action', 21, 22, 'Reveal'),
(NULL, 'action', 23, 24, 'Conceal'),
(NULL, 'action', 25, 26, 'Create'),
(NULL, 'action', 27, 28, 'Destroy'),
(NULL, 'action', 29, 30, 'Transform'),
(NULL, 'action', 31, 32, 'Explore'),
(NULL, 'action', 33, 34, 'Guard'),
(NULL, 'action', 35, 36, 'Betray'),
(NULL, 'action', 37, 38, 'Unite'),
(NULL, 'action', 39, 40, 'Divide'),
(NULL, 'action', 41, 42, 'Seek'),
(NULL, 'action', 43, 44, 'Lose'),
(NULL, 'action', 45, 46, 'Gain'),
(NULL, 'action', 47, 48, 'Challenge'),
(NULL, 'action', 49, 50, 'Endure');

-- Seed generic Theme oracle (1-50)
INSERT INTO oracle_tables (ruleset_id, table_name, roll_min, roll_max, result) VALUES
(NULL, 'theme', 1, 2, 'Power'),
(NULL, 'theme', 3, 4, 'Betrayal'),
(NULL, 'theme', 5, 6, 'Redemption'),
(NULL, 'theme', 7, 8, 'Survival'),
(NULL, 'theme', 9, 10, 'Love'),
(NULL, 'theme', 11, 12, 'Revenge'),
(NULL, 'theme', 13, 14, 'Identity'),
(NULL, 'theme', 15, 16, 'Freedom'),
(NULL, 'theme', 17, 18, 'Sacrifice'),
(NULL, 'theme', 19, 20, 'Corruption'),
(NULL, 'theme', 21, 22, 'Loyalty'),
(NULL, 'theme', 23, 24, 'Mystery'),
(NULL, 'theme', 25, 26, 'War'),
(NULL, 'theme', 27, 28, 'Peace'),
(NULL, 'theme', 29, 30, 'Knowledge'),
(NULL, 'theme', 31, 32, 'Ignorance'),
(NULL, 'theme', 33, 34, 'Fate'),
(NULL, 'theme', 35, 36, 'Chaos'),
(NULL, 'theme', 37, 38, 'Order'),
(NULL, 'theme', 39, 40, 'Legacy'),
(NULL, 'theme', 41, 42, 'Loss'),
(NULL, 'theme', 43, 44, 'Discovery'),
(NULL, 'theme', 45, 46, 'Justice'),
(NULL, 'theme', 47, 48, 'Deception'),
(NULL, 'theme', 49, 50, 'Hope');

-- Tension level on sessions (1-10 scale, default 5)
ALTER TABLE sessions ADD COLUMN tension_level INTEGER NOT NULL DEFAULT 5;

-- Character relationships
CREATE TABLE IF NOT EXISTS relationships (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id INTEGER NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    from_name TEXT NOT NULL,
    to_name TEXT NOT NULL,
    relationship_type TEXT NOT NULL DEFAULT 'neutral',
    description TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
