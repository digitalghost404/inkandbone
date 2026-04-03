package db

import (
	"database/sql"
	"fmt"
)

// --- Settings ---

func (d *DB) GetSetting(key string) (string, error) {
	var value string
	err := d.db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func (d *DB) SetSetting(key, value string) error {
	_, err := d.db.Exec(
		"INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value",
		key, value,
	)
	return err
}

// --- Rulesets ---

type Ruleset struct {
	ID         int64
	Name       string
	SchemaJSON string
	Version    string
}

func (d *DB) CreateRuleset(name, schemaJSON, version string) (int64, error) {
	res, err := d.db.Exec(
		"INSERT INTO rulesets (name, schema_json, version) VALUES (?, ?, ?)",
		name, schemaJSON, version,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *DB) GetRulesetByName(name string) (*Ruleset, error) {
	r := &Ruleset{}
	err := d.db.QueryRow(
		"SELECT id, name, schema_json, version FROM rulesets WHERE name = ?", name,
	).Scan(&r.ID, &r.Name, &r.SchemaJSON, &r.Version)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return r, err
}

func (d *DB) ListRulesets() ([]Ruleset, error) {
	rows, err := d.db.Query("SELECT id, name, schema_json, version FROM rulesets ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Ruleset
	for rows.Next() {
		var r Ruleset
		if err := rows.Scan(&r.ID, &r.Name, &r.SchemaJSON, &r.Version); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// --- Campaigns ---

type Campaign struct {
	ID          int64
	RulesetID   int64
	Name        string
	Description string
	Active      bool
	CreatedAt   string
}

func (d *DB) CreateCampaign(rulesetID int64, name, description string) (int64, error) {
	res, err := d.db.Exec(
		"INSERT INTO campaigns (ruleset_id, name, description) VALUES (?, ?, ?)",
		rulesetID, name, description,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *DB) GetCampaign(id int64) (*Campaign, error) {
	c := &Campaign{}
	var active int
	err := d.db.QueryRow(
		"SELECT id, ruleset_id, name, description, active, created_at FROM campaigns WHERE id = ?", id,
	).Scan(&c.ID, &c.RulesetID, &c.Name, &c.Description, &active, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	c.Active = active == 1
	return c, err
}

func (d *DB) ListCampaigns() ([]Campaign, error) {
	rows, err := d.db.Query(
		"SELECT id, ruleset_id, name, description, active, created_at FROM campaigns ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Campaign
	for rows.Next() {
		var c Campaign
		var active int
		if err := rows.Scan(&c.ID, &c.RulesetID, &c.Name, &c.Description, &active, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.Active = active == 1
		out = append(out, c)
	}
	return out, rows.Err()
}

// --- Characters ---

type Character struct {
	ID           int64
	CampaignID   int64
	Name         string
	DataJSON     string
	PortraitPath string
	CreatedAt    string
}

func (d *DB) CreateCharacter(campaignID int64, name string) (int64, error) {
	res, err := d.db.Exec(
		"INSERT INTO characters (campaign_id, name) VALUES (?, ?)",
		campaignID, name,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *DB) GetCharacter(id int64) (*Character, error) {
	c := &Character{}
	err := d.db.QueryRow(
		"SELECT id, campaign_id, name, data_json, portrait_path, created_at FROM characters WHERE id = ?", id,
	).Scan(&c.ID, &c.CampaignID, &c.Name, &c.DataJSON, &c.PortraitPath, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return c, err
}

func (d *DB) UpdateCharacterData(id int64, dataJSON string) error {
	res, err := d.db.Exec("UPDATE characters SET data_json = ? WHERE id = ?", dataJSON, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("character %d not found", id)
	}
	return nil
}

func (d *DB) ListCharacters(campaignID int64) ([]Character, error) {
	rows, err := d.db.Query(
		"SELECT id, campaign_id, name, data_json, portrait_path, created_at FROM characters WHERE campaign_id = ? ORDER BY name",
		campaignID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Character
	for rows.Next() {
		var c Character
		if err := rows.Scan(&c.ID, &c.CampaignID, &c.Name, &c.DataJSON, &c.PortraitPath, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
