package api

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // local-only app
}

// Hub manages WebSocket connections and broadcasts events to all clients.
// Each client gets a dedicated send channel and write goroutine so that
// broadcast (called from Hub.Run) never shares a *websocket.Conn with the
// per-connection read goroutine in ServeWS.
type Hub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]chan Event
	bus     *Bus
}

func NewHub(bus *Bus) *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]chan Event),
		bus:     bus,
	}
}

// Run subscribes to the event bus and broadcasts all events to connected clients.
// Call in a goroutine.
func (h *Hub) Run() {
	ch := h.bus.Subscribe()
	for event := range ch {
		h.broadcast(event)
	}
}

func (h *Hub) broadcast(event Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, send := range h.clients {
		select {
		case send <- event:
		default:
			// slow client; drop rather than block the broadcast goroutine
		}
	}
}

// ClientCount returns the number of currently connected WebSocket clients.
func (h *Hub) ClientCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.clients)
}

// ServeWS upgrades an HTTP connection to WebSocket and registers it with the hub.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade: %v", err)
		return
	}

	send := make(chan Event, 64)

	h.mu.Lock()
	h.clients[conn] = send
	h.mu.Unlock()

	// Write goroutine — the only goroutine that calls WriteJSON on this conn.
	go func() {
		for event := range send {
			if err := conn.WriteJSON(event); err != nil {
				log.Printf("ws write error: %v", err)
				break
			}
		}
		conn.Close()
	}()

	defer func() {
		h.mu.Lock()
		if ch, ok := h.clients[conn]; ok {
			delete(h.clients, conn)
			close(ch)
		}
		h.mu.Unlock()
	}()

	// Read loop — keeps connection alive; client messages are ignored for now.
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
