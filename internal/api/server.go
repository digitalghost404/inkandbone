package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/digitalghost404/inkandbone/internal/ai"
	"github.com/digitalghost404/inkandbone/internal/db"
)

// Server holds dependencies and registers routes.
type Server struct {
	db       *db.DB
	hub      *Hub
	bus      *Bus
	mux      *http.ServeMux
	dataDir  string
	aiClient ai.Completer // nil when ANTHROPIC_API_KEY is unset
}

// NewServer creates the HTTP server. dataDir is the base path for uploaded files
// (e.g. ~/.ttrpg). aiClient may be nil if AI features are disabled.
func NewServer(database *db.DB, dataDir string, aiClient ai.Completer) *Server {
	bus := NewBus()
	hub := NewHub(bus)
	s := &Server{
		db:       database,
		hub:      hub,
		bus:      bus,
		mux:      http.NewServeMux(),
		dataDir:  dataDir,
		aiClient: aiClient,
	}
	s.registerRoutes()
	go hub.Run()
	return s
}

// Bus returns the event bus so the MCP server can publish events.
func (s *Server) Bus() *Bus { return s.bus }

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Block path traversal attempts on file-serving routes before the mux
	// redirects them (Go's mux normalises .. segments via 307).
	if strings.HasPrefix(r.URL.Path, "/api/files/") && strings.Contains(r.URL.Path, "..") {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	s.mux.ServeHTTP(w, r)
}

// ListenAndServe starts the HTTP server on addr (e.g. ":7432").
func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s)
}

// Shutdown is a no-op placeholder.
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
	// Existing read routes
	s.mux.HandleFunc("GET /api/campaigns", s.handleListCampaigns)
	s.mux.HandleFunc("GET /api/campaigns/{id}/characters", s.handleListCharacters)
	s.mux.HandleFunc("GET /api/campaigns/{id}/sessions", s.handleListSessions)
	s.mux.HandleFunc("GET /api/campaigns/{id}/world-notes", s.handleListWorldNotes)
	s.mux.HandleFunc("GET /api/sessions/{id}/messages", s.handleListMessages)
	s.mux.HandleFunc("GET /api/sessions/{id}/dice-rolls", s.handleListDiceRolls)
	s.mux.HandleFunc("GET /api/maps/{id}/pins", s.handleListMapPins)
	s.mux.HandleFunc("GET /api/context", s.handleGetContext)
	// Plan 7
	s.mux.HandleFunc("GET /api/sessions/{id}/timeline", s.handleGetTimeline)
	// Plan 8
	s.mux.HandleFunc("GET /api/files/{path...}", s.handleServeFile)
	s.mux.HandleFunc("GET /api/campaigns/{id}/maps", s.handleListMaps)
	s.mux.HandleFunc("POST /api/campaigns/{id}/maps", s.handleUploadMap)
	s.mux.HandleFunc("GET /api/maps/{id}", s.handleGetMap)
	s.mux.HandleFunc("PATCH /api/sessions/{id}", s.handlePatchSession)
	s.mux.HandleFunc("POST /api/sessions/{id}/recap", s.handleGenerateRecap)
	s.mux.HandleFunc("POST /api/campaigns/{id}/world-notes/draft", s.handleDraftWorldNote)
	s.mux.HandleFunc("PATCH /api/world-notes/{id}", s.handlePatchWorldNote)
	// Plan 9
	s.mux.HandleFunc("GET /api/rulesets/{id}", s.handleGetRuleset)
	s.mux.HandleFunc("POST /api/rulesets/{id}/rulebook", s.handleIngestRulebook)
	s.mux.HandleFunc("PATCH /api/characters/{id}", s.handlePatchCharacter)
	s.mux.HandleFunc("POST /api/characters/{id}/portrait", s.handleUploadPortrait)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":     "ok",
		"ai_enabled": s.aiClient != nil,
	})
}
