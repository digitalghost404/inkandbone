package mcp

import (
	"testing"

	"github.com/digitalghost404/inkandbone/internal/ai"
	"github.com/digitalghost404/inkandbone/internal/api"
	"github.com/digitalghost404/inkandbone/internal/db"
	"github.com/stretchr/testify/require"
)

func newTestDB(t *testing.T) *db.DB {
	t.Helper()
	d, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { d.Close() })
	return d
}

func newTestMCP(t *testing.T) *Server {
	t.Helper()
	d := newTestDB(t)
	bus := api.NewBus()
	return New(d, bus, nil)
}

func newTestMCPWithAI(t *testing.T, c ai.Completer) *Server {
	t.Helper()
	d := newTestDB(t)
	bus := api.NewBus()
	return New(d, bus, c)
}

func TestNewServer_doesNotPanic(t *testing.T) {
	s := newTestMCP(t)
	require.NotNil(t, s)
}
