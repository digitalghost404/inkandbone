package db

import "database/sql"

// Item represents an inventory item owned by a character.
type Item struct {
	ID          int64  `json:"id"`
	CharacterID int64  `json:"character_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
	Equipped    bool   `json:"equipped"`
	CreatedAt   string `json:"created_at"`
}

// ListItems returns all items for a character, ordered by created_at.
func (d *DB) ListItems(characterID int64) ([]Item, error) {
	rows, err := d.db.Query(
		"SELECT id, character_id, name, description, quantity, equipped, created_at FROM items WHERE character_id = ? ORDER BY created_at",
		characterID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Item
	for rows.Next() {
		var item Item
		var equipped int
		if err := rows.Scan(&item.ID, &item.CharacterID, &item.Name, &item.Description, &item.Quantity, &equipped, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.Equipped = equipped != 0
		out = append(out, item)
	}
	return out, rows.Err()
}

// CreateItem inserts a new item and returns it.
func (d *DB) CreateItem(characterID int64, name, description string, quantity int) (*Item, error) {
	res, err := d.db.Exec(
		"INSERT INTO items (character_id, name, description, quantity) VALUES (?, ?, ?, ?)",
		characterID, name, description, quantity,
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	var item Item
	var equipped int
	err = d.db.QueryRow(
		"SELECT id, character_id, name, description, quantity, equipped, created_at FROM items WHERE id = ?", id,
	).Scan(&item.ID, &item.CharacterID, &item.Name, &item.Description, &item.Quantity, &equipped, &item.CreatedAt)
	if err != nil {
		return nil, err
	}
	item.Equipped = equipped != 0
	return &item, nil
}

// UpdateItem patches mutable fields (name, description, quantity, equipped).
func (d *DB) UpdateItem(id int64, name, description string, quantity int, equipped bool) error {
	equippedInt := 0
	if equipped {
		equippedInt = 1
	}
	res, err := d.db.Exec(
		"UPDATE items SET name = ?, description = ?, quantity = ?, equipped = ? WHERE id = ?",
		name, description, quantity, equippedInt, id,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteItem removes an item by ID.
func (d *DB) DeleteItem(id int64) error {
	_, err := d.db.Exec("DELETE FROM items WHERE id = ?", id)
	return err
}

// GetItem returns a single item by ID.
func (d *DB) GetItem(id int64) (*Item, error) {
	var item Item
	var equipped int
	err := d.db.QueryRow(
		"SELECT id, character_id, name, description, quantity, equipped, created_at FROM items WHERE id = ?", id,
	).Scan(&item.ID, &item.CharacterID, &item.Name, &item.Description, &item.Quantity, &equipped, &item.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	item.Equipped = equipped != 0
	return &item, nil
}
