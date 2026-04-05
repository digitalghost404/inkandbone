package db

// XPEntry records a story milestone or XP award for a session.
// Amount is optional — nil for rulesets without numeric XP.
type XPEntry struct {
	ID        int64  `json:"id"`
	SessionID int64  `json:"session_id"`
	Note      string `json:"note"`
	Amount    *int   `json:"amount"`
	CreatedAt string `json:"created_at"`
}

// ListXP returns all XP log entries for a session, newest first.
func (d *DB) ListXP(sessionID int64) ([]XPEntry, error) {
	rows, err := d.db.Query(
		"SELECT id, session_id, note, amount, created_at FROM xp_log WHERE session_id = ? ORDER BY created_at DESC",
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []XPEntry
	for rows.Next() {
		var x XPEntry
		if err := rows.Scan(&x.ID, &x.SessionID, &x.Note, &x.Amount, &x.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

// CreateXP inserts an XP log entry and returns it.
func (d *DB) CreateXP(sessionID int64, note string, amount *int) (*XPEntry, error) {
	res, err := d.db.Exec(
		"INSERT INTO xp_log (session_id, note, amount) VALUES (?, ?, ?)",
		sessionID, note, amount,
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	var x XPEntry
	err = d.db.QueryRow(
		"SELECT id, session_id, note, amount, created_at FROM xp_log WHERE id = ?", id,
	).Scan(&x.ID, &x.SessionID, &x.Note, &x.Amount, &x.CreatedAt)
	return &x, err
}

// DeleteXP removes an XP log entry.
func (d *DB) DeleteXP(id int64) error {
	_, err := d.db.Exec("DELETE FROM xp_log WHERE id = ?", id)
	return err
}
