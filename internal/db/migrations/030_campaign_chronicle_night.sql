-- 030_campaign_chronicle_night.sql: Track in-game nights for VtM campaigns
-- chronicle_night is the current in-game night of the chronicle (1-based).
-- Important in VtM for feeding intervals, blood bonds, Masquerade timelines.
ALTER TABLE campaigns ADD COLUMN chronicle_night INTEGER NOT NULL DEFAULT 1;
