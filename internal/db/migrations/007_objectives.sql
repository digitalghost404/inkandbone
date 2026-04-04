CREATE TABLE IF NOT EXISTS objectives (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  campaign_id INTEGER NOT NULL REFERENCES campaigns(id),
  title       TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  status      TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active','completed','failed')),
  created_at  TEXT NOT NULL DEFAULT (datetime('now'))
);
