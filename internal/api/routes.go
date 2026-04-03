package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/digitalghost404/inkandbone/internal/db"
)

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func parsePathID(r *http.Request, key string) (int64, bool) {
	s := r.PathValue(key)
	id, err := strconv.ParseInt(s, 10, 64)
	return id, err == nil && id > 0
}

func (s *Server) handleListCampaigns(w http.ResponseWriter, _ *http.Request) {
	campaigns, err := s.db.ListCampaigns()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if campaigns == nil {
		campaigns = []db.Campaign{}
	}
	writeJSON(w, campaigns)
}

func (s *Server) handleListCharacters(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}
	characters, err := s.db.ListCharacters(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if characters == nil {
		characters = []db.Character{}
	}
	writeJSON(w, characters)
}

func (s *Server) handleListSessions(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}
	sessions, err := s.db.ListSessions(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if sessions == nil {
		sessions = []db.Session{}
	}
	writeJSON(w, sessions)
}

func (s *Server) handleListMessages(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	messages, err := s.db.ListMessages(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if messages == nil {
		messages = []db.Message{}
	}
	writeJSON(w, messages)
}

func (s *Server) handleListDiceRolls(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	rolls, err := s.db.ListDiceRolls(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rolls == nil {
		rolls = []db.DiceRoll{}
	}
	writeJSON(w, rolls)
}

func (s *Server) handleListMapPins(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid map id", http.StatusBadRequest)
		return
	}
	pins, err := s.db.ListMapPins(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if pins == nil {
		pins = []db.MapPin{}
	}
	writeJSON(w, pins)
}

type contextCombatSnapshot struct {
	Encounter  *db.CombatEncounter `json:"encounter"`
	Combatants []db.Combatant      `json:"combatants"`
}

type contextResponse struct {
	Campaign       *db.Campaign           `json:"campaign"`
	Character      *db.Character          `json:"character"`
	Session        *db.Session            `json:"session"`
	RecentMessages []db.Message           `json:"recent_messages"`
	ActiveCombat   *contextCombatSnapshot `json:"active_combat"`
}

func (s *Server) handleGetContext(w http.ResponseWriter, _ *http.Request) {
	resp := contextResponse{RecentMessages: []db.Message{}}

	if campIDStr, err := s.db.GetSetting("active_campaign_id"); err == nil && campIDStr != "" {
		if campID, err := strconv.ParseInt(campIDStr, 10, 64); err == nil {
			resp.Campaign, _ = s.db.GetCampaign(campID)
		}
	}
	if charIDStr, err := s.db.GetSetting("active_character_id"); err == nil && charIDStr != "" {
		if charID, err := strconv.ParseInt(charIDStr, 10, 64); err == nil {
			resp.Character, _ = s.db.GetCharacter(charID)
		}
	}
	if sessIDStr, err := s.db.GetSetting("active_session_id"); err == nil && sessIDStr != "" {
		if sessID, err := strconv.ParseInt(sessIDStr, 10, 64); err == nil {
			resp.Session, _ = s.db.GetSession(sessID)
			if msgs, err := s.db.RecentMessages(sessID, 20); err == nil {
				resp.RecentMessages = msgs
			}
			if enc, err := s.db.GetActiveEncounter(sessID); err == nil && enc != nil {
				cs := &contextCombatSnapshot{Encounter: enc, Combatants: []db.Combatant{}}
				if combatants, err := s.db.ListCombatants(enc.ID); err == nil {
					cs.Combatants = combatants
				}
				resp.ActiveCombat = cs
			}
		}
	}

	writeJSON(w, resp)
}
