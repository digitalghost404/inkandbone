package db

// Objective represents a campaign goal or quest being tracked.
type Objective struct {
	ID          int64  `json:"id"`
	CampaignID  int64  `json:"campaign_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

// ListObjectives returns all objectives for a campaign, ordered by created_at DESC.
func (d *DB) ListObjectives(campaignID int64) ([]Objective, error) {
	rows, err := d.db.Query(
		"SELECT id, campaign_id, title, description, status, created_at FROM objectives WHERE campaign_id = ? ORDER BY created_at DESC",
		campaignID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Objective
	for rows.Next() {
		var o Objective
		if err := rows.Scan(&o.ID, &o.CampaignID, &o.Title, &o.Description, &o.Status, &o.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// CreateObjective inserts a new objective and returns it.
func (d *DB) CreateObjective(campaignID int64, title, description string) (*Objective, error) {
	res, err := d.db.Exec(
		"INSERT INTO objectives (campaign_id, title, description) VALUES (?, ?, ?)",
		campaignID, title, description,
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
		"SELECT id, campaign_id, title, description, status, created_at FROM objectives WHERE id = ?", id,
	).Scan(&o.ID, &o.CampaignID, &o.Title, &o.Description, &o.Status, &o.CreatedAt)
	return &o, err
}

// UpdateObjectiveStatus sets the status for an objective.
func (d *DB) UpdateObjectiveStatus(id int64, status string) error {
	_, err := d.db.Exec("UPDATE objectives SET status = ? WHERE id = ?", status, id)
	return err
}

// DeleteObjective removes an objective.
func (d *DB) DeleteObjective(id int64) error {
	_, err := d.db.Exec("DELETE FROM objectives WHERE id = ?", id)
	return err
}
