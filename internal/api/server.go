package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/digitalghost404/inkandbone/internal/db"
)

// Server holds dependencies and registers routes.
type Server struct {
	db  *db.DB
	hub *Hub
	bus *Bus
	mux *http.ServeMux
}

func NewServer(database *db.DB) *Server {
	bus := NewBus()
	hub := NewHub(bus)
	s := &Server{
		db:  database,
		hub: hub,
		bus: bus,
		mux: http.NewServeMux(),
	}
	s.registerRoutes()
	go hub.Run()
	return s
}

// Bus returns the event bus so the MCP server (Plan 2) can publish events.
func (s *Server) Bus() *Bus { return s.bus }

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// ListenAndServe starts the HTTP server on addr (e.g. ":7432").
func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s)
}

// Shutdown is a no-op placeholder; Go's http.Server.Shutdown can be wired here later.
func (s *Server) Shutdown(_ context.Context) error { return nil }

// RegisterStatic serves the embedded React SPA for all routes not matched by /api/ or /ws.
func (s *Server) RegisterStatic(fsys http.FileSystem) {
	fileServer := http.FileServer(fsys)
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fileServer.ServeHTTP(w, r)
	})
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/ws", s.hub.ServeWS)
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("GET /api/campaigns", s.handleListCampaigns)
	s.mux.HandleFunc("GET /api/campaigns/{id}/characters", s.handleListCharacters)
	s.mux.HandleFunc("GET /api/campaigns/{id}/sessions", s.handleListSessions)
	s.mux.HandleFunc("GET /api/sessions/{id}/messages", s.handleListMessages)
	s.mux.HandleFunc("GET /api/sessions/{id}/dice-rolls", s.handleListDiceRolls)
	s.mux.HandleFunc("GET /api/maps/{id}/pins", s.handleListMapPins)
	s.mux.HandleFunc("GET /api/context", s.handleGetContext)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
