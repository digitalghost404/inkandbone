package db

import "database/sql"

// Objective represents a campaign goal or quest being tracked.
type Objective struct {
	ID          int64  `json:"id"`
	CampaignID  int64  `json:"campaign_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	ParentID    *int64 `json:"parent_id"`
	CreatedAt   string `json:"created_at"`
}

// ListObjectives returns all objectives for a campaign, ordered by created_at DESC.
func (d *DB) ListObjectives(campaignID int64) ([]Objective, error) {
	rows, err := d.db.Query(
		"SELECT id, campaign_id, title, description, status, parent_id, created_at FROM objectives WHERE campaign_id = ? ORDER BY created_at DESC",
		campaignID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Objective
	for rows.Next() {
		var o Objective
		if err := rows.Scan(&o.ID, &o.CampaignID, &o.Title, &o.Description, &o.Status, &o.ParentID, &o.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// GetObjective returns a single objective by ID, or nil if not found.
func (d *DB) GetObjective(id int64) (*Objective, error) {
	var o Objective
	err := d.db.QueryRow(
		"SELECT id, campaign_id, title, description, status, parent_id, created_at FROM objectives WHERE id = ?", id,
	).Scan(&o.ID, &o.CampaignID, &o.Title, &o.Description, &o.Status, &o.ParentID, &o.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}

// CreateObjective inserts a new objective and returns it.
// parentID may be nil for top-level objectives.
func (d *DB) CreateObjective(campaignID int64, title, description string, parentID *int64) (*Objective, error) {
	res, err := d.db.Exec(
		"INSERT INTO objectives (campaign_id, title, description, parent_id) VALUES (?, ?, ?, ?)",
		campaignID, title, description, parentID,
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	var o Objective
	err = d.db.QueryRow(
		"SELECT id, campaign_id, title, description, status, parent_id, created_at FROM objectives WHERE id = ?", id,
	).Scan(&o.ID, &o.CampaignID, &o.Title, &o.Description, &o.Status, &o.ParentID, &o.CreatedAt)
	return &o, err
}

// UpdateObjectiveStatus sets the status for an objective.
func (d *DB) UpdateObjectiveStatus(id int64, status string) error {
	_, err := d.db.Exec("UPDATE objectives SET status = ? WHERE id = ?", status, id)
	return err
}

// DeleteObjective removes an objective and all its sub-tasks.
func (d *DB) DeleteObjective(id int64) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck
	// Delete sub-tasks first
	if _, err := tx.Exec("DELETE FROM objectives WHERE parent_id = ?", id); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM objectives WHERE id = ?", id); err != nil {
		return err
	}
	return tx.Commit()
}
