package db

import (
	"database/sql"
	"time"
)

// RulebookChunk represents a parsed chunk of rulebook text for a given ruleset.
type RulebookChunk struct {
	ID        int64     `json:"id"`
	RulesetID int64     `json:"ruleset_id"`
	Heading   string    `json:"heading"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// GetRuleset returns a single ruleset by ID.
func (d *DB) GetRuleset(id int64) (*Ruleset, error) {
	var r Ruleset
	err := d.db.QueryRow(
		"SELECT id, name, schema_json, version FROM rulesets WHERE id = ?", id,
	).Scan(&r.ID, &r.Name, &r.SchemaJSON, &r.Version)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// CreateRulebookChunks inserts multiple chunks for a ruleset in a single transaction.
func (d *DB) CreateRulebookChunks(rulesetID int64, chunks []RulebookChunk) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(
		"INSERT INTO rulebook_chunks (ruleset_id, heading, content) VALUES (?, ?, ?)",
	)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, c := range chunks {
		if _, err := stmt.Exec(rulesetID, c.Heading, c.Content); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// SearchRulebookChunks returns up to 3 chunks matching query in heading or content (LIKE search).
func (d *DB) SearchRulebookChunks(rulesetID int64, query string) ([]RulebookChunk, error) {
	like := "%" + query + "%"
	rows, err := d.db.Query(
		`SELECT id, ruleset_id, heading, content, created_at
		 FROM rulebook_chunks
		 WHERE ruleset_id = ? AND (heading LIKE ? OR content LIKE ?)
		 LIMIT 3`,
		rulesetID, like, like,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var chunks []RulebookChunk
	for rows.Next() {
		var c RulebookChunk
		if err := rows.Scan(&c.ID, &c.RulesetID, &c.Heading, &c.Content, &c.CreatedAt); err != nil {
			return nil, err
		}
		chunks = append(chunks, c)
	}
	return chunks, rows.Err()
}
