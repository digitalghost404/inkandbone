package db

import (
	"database/sql"
	"fmt"
)

type CombatEncounter struct {
	ID               int64  `json:"id"`
	SessionID        int64  `json:"session_id"`
	Name             string `json:"name"`
	Active           bool   `json:"active"`
	ActiveTurnIndex  int    `json:"active_turn_index"`
	CreatedAt        string `json:"created_at"`
}

func (d *DB) CreateEncounter(sessionID int64, name string) (int64, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.Exec("UPDATE combat_encounters SET active = 0 WHERE session_id = ? AND active = 1", sessionID); err != nil {
		return 0, fmt.Errorf("deactivate existing encounter: %w", err)
	}
	res, err := tx.Exec(
		"INSERT INTO combat_encounters (session_id, name) VALUES (?, ?)",
		sessionID, name,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

func (d *DB) GetActiveEncounter(sessionID int64) (*CombatEncounter, error) {
	e := &CombatEncounter{}
	var active int
	err := d.db.QueryRow(
		"SELECT id, session_id, name, active, active_turn_index, created_at FROM combat_encounters WHERE session_id = ? AND active = 1",
		sessionID,
	).Scan(&e.ID, &e.SessionID, &e.Name, &active, &e.ActiveTurnIndex, &e.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	e.Active = active == 1
	return e, nil
}

func (d *DB) EndEncounter(id int64) error {
	_, err := d.db.Exec("UPDATE combat_encounters SET active = 0 WHERE id = ?", id)
	return err
}

// AdvanceTurn increments active_turn_index modulo the number of combatants.
// Returns the new index. Returns 0 if the encounter has no combatants.
// Returns an error containing "not found" if the encounter does not exist.
func (d *DB) AdvanceTurn(encounterID int64) (int, error) {
	var current int
	err := d.db.QueryRow("SELECT active_turn_index FROM combat_encounters WHERE id = ?", encounterID).Scan(&current)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("combat encounter %d not found", encounterID)
	}
	if err != nil {
		return 0, err
	}
	var count int
	if err := d.db.QueryRow("SELECT COUNT(*) FROM combatants WHERE encounter_id = ?", encounterID).Scan(&count); err != nil {
		return 0, err
	}
	if count == 0 {
		return 0, nil
	}
	next := (current + 1) % count
	_, err = d.db.Exec("UPDATE combat_encounters SET active_turn_index = ? WHERE id = ?", next, encounterID)
	return next, err
}

// --- Combatants ---

type Combatant struct {
	ID             int64  `json:"id"`
	EncounterID    int64  `json:"encounter_id"`
	CharacterID    *int64 `json:"character_id"`
	Name           string `json:"name"`
	Initiative     int    `json:"initiative"`
	HPCurrent      int    `json:"hp_current"`
	HPMax          int    `json:"hp_max"`
	ConditionsJSON string `json:"conditions_json"`
	IsPlayer       bool   `json:"is_player"`
}

func (d *DB) AddCombatant(encounterID int64, name string, initiative, hpMax int, isPlayer bool, characterID *int64) (int64, error) {
	res, err := d.db.Exec(
		`INSERT INTO combatants (encounter_id, character_id, name, initiative, hp_current, hp_max, is_player)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		encounterID, characterID, name, initiative, hpMax, hpMax, boolToInt(isPlayer),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *DB) UpdateCombatant(id int64, hpCurrent int, conditionsJSON string) error {
	_, err := d.db.Exec(
		"UPDATE combatants SET hp_current = ?, conditions_json = ? WHERE id = ?",
		hpCurrent, conditionsJSON, id,
	)
	return err
}

func (d *DB) ListCombatants(encounterID int64) ([]Combatant, error) {
	rows, err := d.db.Query(
		`SELECT id, encounter_id, character_id, name, initiative, hp_current, hp_max, conditions_json, is_player
		 FROM combatants WHERE encounter_id = ? ORDER BY initiative DESC`,
		encounterID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Combatant
	for rows.Next() {
		var c Combatant
		var isPlayer int
		if err := rows.Scan(&c.ID, &c.EncounterID, &c.CharacterID, &c.Name,
			&c.Initiative, &c.HPCurrent, &c.HPMax, &c.ConditionsJSON, &isPlayer); err != nil {
			return nil, err
		}
		c.IsPlayer = isPlayer == 1
		out = append(out, c)
	}
	return out, rows.Err()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
