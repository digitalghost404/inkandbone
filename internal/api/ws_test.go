package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHubBroadcastsEvents(t *testing.T) {
	bus := NewBus()
	hub := NewHub(bus)
	go hub.Run()

	srv := httptest.NewServer(http.HandlerFunc(hub.ServeWS))
	defer srv.Close()

	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	defer conn.Close()

	// Give hub time to register the client
	time.Sleep(10 * time.Millisecond)

	bus.Publish(Event{Type: EventDiceRolled, Payload: map[string]any{"result": 18}})

	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	var received Event
	err = conn.ReadJSON(&received)
	require.NoError(t, err)
	assert.Equal(t, EventDiceRolled, received.Type)
}
