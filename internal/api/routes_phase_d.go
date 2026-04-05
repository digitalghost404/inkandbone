package api

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (s *Server) handleOracleRoll(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Table     string `json:"table"`
		Roll      int    `json:"roll"`
		RulesetID *int64 `json:"ruleset_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Table == "" || body.Roll == 0 {
		http.Error(w, "table and roll required", http.StatusBadRequest)
		return
	}

	result, err := s.db.RollOracle(body.RulesetID, body.Table, body.Roll)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	s.bus.Publish(Event{Type: EventOracleRolled, Payload: map[string]any{
		"table": body.Table, "roll": body.Roll, "result": result,
	}})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
		"result": result,
		"table":  body.Table,
		"roll":   body.Roll,
	})
}

func (s *Server) handleGetTension(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	sessionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}

	level, err := s.db.GetTension(sessionID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"tension_level": level}) //nolint:errcheck
}

func (s *Server) handlePatchTension(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	sessionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}

	var body struct {
		TensionLevel int `json:"tension_level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if err := s.db.UpdateTension(sessionID, body.TensionLevel); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	s.bus.Publish(Event{Type: EventTensionUpdated, Payload: map[string]any{
		"session_id": sessionID, "tension_level": body.TensionLevel,
	}})

	w.WriteHeader(http.StatusOK)
}
