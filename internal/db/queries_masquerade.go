// internal/db/queries_masquerade.go
package db

// GetMasqueradeIntegrity returns the Masquerade integrity (0-10) for a session.
// Default is 10 (full Masquerade intact).
func (db *DB) GetMasqueradeIntegrity(sessionID int64) (int, error) {
	var level int
	err := db.db.QueryRow(
		`SELECT masquerade_integrity FROM sessions WHERE id = ?`, sessionID,
	).Scan(&level)
	return level, err
}

// UpdateMasqueradeIntegrity sets the Masquerade integrity for a session, clamping to [0, 10].
func (db *DB) UpdateMasqueradeIntegrity(sessionID int64, level int) error {
	if level < 0 {
		level = 0
	}
	if level > 10 {
		level = 10
	}
	_, err := db.db.Exec(
		`UPDATE sessions SET masquerade_integrity = ? WHERE id = ?`, level, sessionID,
	)
	return err
}
