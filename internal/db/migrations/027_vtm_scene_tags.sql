-- 027_vtm_scene_tags.sql: Add masquerade_integrity to sessions, document VtM scene tags
-- masquerade_integrity tracks how many Masquerade points remain (0-10, default 10).
-- VtM scene tags (elysium, haven, hunt, masquerade) are handled via sceneTagKeywords
-- in routes.go — no schema change needed for tags themselves.
ALTER TABLE sessions ADD COLUMN masquerade_integrity INTEGER NOT NULL DEFAULT 10;
