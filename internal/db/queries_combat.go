package db

import "database/sql"

type CombatEncounter struct {
	ID        int64
	SessionID int64
	Name      string
	Active    bool
	CreatedAt string
}

func (d *DB) CreateEncounter(sessionID int64, name string) (int64, error) {
	// Deactivate any existing active encounter in this session first
	d.db.Exec("UPDATE combat_encounters SET active = 0 WHERE session_id = ? AND active = 1", sessionID)
	res, err := d.db.Exec(
		"INSERT INTO combat_encounters (session_id, name) VALUES (?, ?)",
		sessionID, name,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *DB) GetActiveEncounter(sessionID int64) (*CombatEncounter, error) {
	e := &CombatEncounter{}
	var active int
	err := d.db.QueryRow(
		"SELECT id, session_id, name, active, created_at FROM combat_encounters WHERE session_id = ? AND active = 1",
		sessionID,
	).Scan(&e.ID, &e.SessionID, &e.Name, &active, &e.CreatedAt)
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

// --- Combatants ---

type Combatant struct {
	ID             int64
	EncounterID    int64
	CharacterID    *int64
	Name           string
	Initiative     int
	HPCurrent      int
	HPMax          int
	ConditionsJSON string
	IsPlayer       bool
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
