package api

import (
	"context"
	"testing"

	"github.com/digitalghost404/inkandbone/internal/ai"
	"github.com/digitalghost404/inkandbone/internal/db"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	d, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { d.Close() })
	return NewServer(d, t.TempDir(), nil)
}

func newTestServerWithDir(t *testing.T, dir string) *Server {
	t.Helper()
	d, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { d.Close() })
	return NewServer(d, dir, nil)
}

type stubCompleter struct{ response string }

func (s *stubCompleter) Generate(_ context.Context, _ string, _ int) (string, error) {
	return s.response, nil
}

func newTestServerWithAI(t *testing.T, c ai.Completer) *Server {
	t.Helper()
	d, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { d.Close() })
	return NewServer(d, t.TempDir(), c)
}
