package db

// GetTension returns the tension level (1-10) for a session.
func (db *DB) GetTension(sessionID int64) (int, error) {
	var level int
	err := db.db.QueryRow(
		`SELECT tension_level FROM sessions WHERE id = ?`, sessionID,
	).Scan(&level)
	return level, err
}

// UpdateTension sets the tension level for a session, clamping to [1, 10].
func (db *DB) UpdateTension(sessionID int64, level int) error {
	if level < 1 {
		level = 1
	}
	if level > 10 {
		level = 10
	}
	_, err := db.db.Exec(
		`UPDATE sessions SET tension_level = ? WHERE id = ?`, level, sessionID,
	)
	return err
}
