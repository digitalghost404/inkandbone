package db

import (
	"database/sql"
	"fmt"
)

type Session struct {
	ID         int64  `json:"id"`
	CampaignID int64  `json:"campaign_id"`
	Title      string `json:"title"`
	Date       string `json:"date"`
	Summary    string `json:"summary"`
	CreatedAt  string `json:"created_at"`
}

func (d *DB) CreateSession(campaignID int64, title, date string) (int64, error) {
	res, err := d.db.Exec(
		"INSERT INTO sessions (campaign_id, title, date) VALUES (?, ?, ?)",
		campaignID, title, date,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *DB) GetSession(id int64) (*Session, error) {
	s := &Session{}
	err := d.db.QueryRow(
		"SELECT id, campaign_id, title, date, summary, created_at FROM sessions WHERE id = ?", id,
	).Scan(&s.ID, &s.CampaignID, &s.Title, &s.Date, &s.Summary, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return s, err
}

func (d *DB) UpdateSessionSummary(id int64, summary string) error {
	res, err := d.db.Exec("UPDATE sessions SET summary = ? WHERE id = ?", summary, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("session %d not found", id)
	}
	return nil
}

func (d *DB) ListSessions(campaignID int64) ([]Session, error) {
	rows, err := d.db.Query(
		"SELECT id, campaign_id, title, date, summary, created_at FROM sessions WHERE campaign_id = ? ORDER BY date DESC",
		campaignID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Session
	for rows.Next() {
		var s Session
		if err := rows.Scan(&s.ID, &s.CampaignID, &s.Title, &s.Date, &s.Summary, &s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// --- Messages ---

type Message struct {
	ID        int64  `json:"id"`
	SessionID int64  `json:"session_id"`
	Role      string `json:"role"` // "user" or "assistant"
	Content   string `json:"content"`
	Whisper   bool   `json:"whisper"`
	CreatedAt string `json:"created_at"`
}

func (d *DB) CreateMessage(sessionID int64, role, content string, whisper bool) (int64, error) {
	res, err := d.db.Exec(
		"INSERT INTO messages (session_id, role, content, whisper) VALUES (?, ?, ?, ?)",
		sessionID, role, content, boolToIntMsg(whisper),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func boolToIntMsg(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (d *DB) ListMessages(sessionID int64) ([]Message, error) {
	rows, err := d.db.Query(
		"SELECT id, session_id, role, content, whisper, created_at FROM messages WHERE session_id = ? ORDER BY created_at, id",
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Message
	for rows.Next() {
		var m Message
		var whisper int
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &whisper, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.Whisper = whisper == 1
		out = append(out, m)
	}
	return out, rows.Err()
}

func (d *DB) RecentMessages(sessionID int64, limit int) ([]Message, error) {
	rows, err := d.db.Query(
		`SELECT id, session_id, role, content, whisper, created_at FROM messages
		 WHERE session_id = ? ORDER BY created_at DESC, id DESC LIMIT ?`,
		sessionID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Message
	for rows.Next() {
		var m Message
		var whisper int
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &whisper, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.Whisper = whisper == 1
		out = append(out, m)
	}
	// reverse to chronological order
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, rows.Err()
}
