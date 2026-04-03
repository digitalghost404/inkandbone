package db

import (
	"encoding/json"
	"fmt"
	"sort"
)

// TimelineEntry is a single chronological item in the session feed.
// Type is one of "message" or "dice_roll".
// Data is the raw JSON of the underlying record.
type TimelineEntry struct {
	Type      string          `json:"type"`
	Timestamp string          `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

// GetSessionTimeline returns all messages and dice rolls for the given session,
// merged and sorted by created_at ascending.
func (d *DB) GetSessionTimeline(sessionID int64) ([]TimelineEntry, error) {
	msgs, err := d.ListMessages(sessionID)
	if err != nil {
		return nil, err
	}
	rolls, err := d.ListDiceRolls(sessionID)
	if err != nil {
		return nil, err
	}

	entries := make([]TimelineEntry, 0, len(msgs)+len(rolls))
	for _, m := range msgs {
		b, err := json.Marshal(m)
		if err != nil {
			return nil, fmt.Errorf("marshal message %d: %w", m.ID, err)
		}
		entries = append(entries, TimelineEntry{Type: "message", Timestamp: m.CreatedAt, Data: b})
	}
	for _, r := range rolls {
		b, err := json.Marshal(r)
		if err != nil {
			return nil, fmt.Errorf("marshal dice roll %d: %w", r.ID, err)
		}
		entries = append(entries, TimelineEntry{Type: "dice_roll", Timestamp: r.CreatedAt, Data: b})
	}

	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].Timestamp < entries[j].Timestamp
	})

	return entries, nil
}
