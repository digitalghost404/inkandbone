package db

import "database/sql"

// RollOracle looks up an oracle table result for the given roll value.
// If rulesetID is provided, it first tries ruleset-specific rows, then falls back to generic (ruleset_id IS NULL).
// Returns empty string (no error) if no matching row is found.
func (d *DB) RollOracle(rulesetID *int64, tableName string, roll int) (string, error) {
	// Try ruleset-specific first if rulesetID provided
	if rulesetID != nil {
		var result string
		err := d.db.QueryRow(
			`SELECT result FROM oracle_tables
             WHERE ruleset_id = ? AND table_name = ? AND roll_min <= ? AND roll_max >= ?
             LIMIT 1`,
			*rulesetID, tableName, roll, roll,
		).Scan(&result)
		if err == nil {
			return result, nil
		}
		if err != sql.ErrNoRows {
			return "", err
		}
	}

	// Fall back to generic (ruleset_id IS NULL)
	var result string
	err := d.db.QueryRow(
		`SELECT result FROM oracle_tables
         WHERE ruleset_id IS NULL AND table_name = ? AND roll_min <= ? AND roll_max >= ?
         LIMIT 1`,
		tableName, roll, roll,
	).Scan(&result)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return result, nil
}
