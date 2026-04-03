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
type Hub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]bool
	bus     *Bus
}

func NewHub(bus *Bus) *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]bool),
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
	for conn := range h.clients {
		if err := conn.WriteJSON(event); err != nil {
			log.Printf("ws write error: %v", err)
			conn.Close()
			delete(h.clients, conn)
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
	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, conn)
		h.mu.Unlock()
		conn.Close()
	}()

	// Read loop — keeps connection alive; client messages are ignored for now
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
