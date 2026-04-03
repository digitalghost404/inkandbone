package db

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpen_CreatesTablesOnFirstRun(t *testing.T) {
	d, err := Open(":memory:")
	require.NoError(t, err)
	defer d.Close()

	tables := []string{
		"settings", "rulesets", "campaigns", "characters",
		"sessions", "messages", "combat_encounters", "combatants",
		"world_notes", "maps", "map_pins", "dice_rolls",
	}
	for _, table := range tables {
		var name string
		err := d.db.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table,
		).Scan(&name)
		require.NoError(t, err, "table %s should exist", table)
	}
}

func TestOpen_IdempotentMigrations(t *testing.T) {
	d, err := Open(":memory:")
	require.NoError(t, err)
	d.Close()

	tmp := t.TempDir() + "/test.db"
	d1, err := Open(tmp)
	require.NoError(t, err)
	d1.Close()

	d2, err := Open(tmp)
	require.NoError(t, err)
	defer d2.Close()
}
