package api

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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

func (s *Server) handleListWorldNotes(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}
	q := r.URL.Query().Get("q")
	category := r.URL.Query().Get("category")
	tag := r.URL.Query().Get("tag")
	notes, err := s.db.SearchWorldNotes(id, q, category, tag)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if notes == nil {
		notes = []db.WorldNote{}
	}
	writeJSON(w, notes)
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

func (s *Server) handlePatchWorldNote(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid world note id", http.StatusBadRequest)
		return
	}
	var body struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		TagsJSON string `json:"tags_json"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if body.Title == "" || body.Content == "" {
		http.Error(w, "title and content are required", http.StatusBadRequest)
		return
	}
	if err := s.db.UpdateWorldNote(id, body.Title, body.Content, body.TagsJSON); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.bus.Publish(Event{Type: EventWorldNoteUpdated, Payload: map[string]any{"note_id": id}})
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleServeFile(w http.ResponseWriter, r *http.Request) {
	rel := filepath.Clean(r.PathValue("path"))
	// Reject any path that tries to escape the data directory
	if strings.HasPrefix(rel, "..") {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	abs := filepath.Join(s.dataDir, rel)
	if !strings.HasPrefix(abs+string(filepath.Separator), s.dataDir+string(filepath.Separator)) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	http.ServeFile(w, r, abs)
}

func (s *Server) handleListMaps(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}
	maps, err := s.db.ListMaps(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if maps == nil {
		maps = []db.Map{}
	}
	writeJSON(w, maps)
}

func (s *Server) handleGetMap(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid map id", http.StatusBadRequest)
		return
	}
	m, err := s.db.GetMap(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if m == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, m)
}

func (s *Server) handleUploadMap(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "parse form: "+err.Error(), http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "image is required: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	filename := randomHex(16) + ext
	destDir := filepath.Join(s.dataDir, "maps")
	if err := os.MkdirAll(destDir, 0750); err != nil {
		http.Error(w, "mkdir: "+err.Error(), http.StatusInternalServerError)
		return
	}
	out, err := os.Create(filepath.Join(destDir, filename))
	if err != nil {
		http.Error(w, "create file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		out.Close()
		os.Remove(filepath.Join(destDir, filename))
		http.Error(w, "write file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	imagePath := "maps/" + filename
	mapID, err := s.db.CreateMap(id, name, imagePath)
	if err != nil {
		http.Error(w, "db: "+err.Error(), http.StatusInternalServerError)
		return
	}
	m, err := s.db.GetMap(mapID)
	if err != nil || m == nil {
		http.Error(w, "fetch created map", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(m) //nolint:errcheck
}

// randomHex returns n random hex bytes as a hex string.
func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b) //nolint:errcheck // crypto/rand.Read never returns an error on supported platforms
	return fmt.Sprintf("%x", b)
}

func (s *Server) handlePatchSession(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	var body struct {
		Summary string `json:"summary"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := s.db.UpdateSessionSummary(id, body.Summary); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.bus.Publish(Event{Type: EventSessionUpdated, Payload: map[string]any{
		"session_id": id,
		"summary":    body.Summary,
	}})
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleDraftWorldNote(w http.ResponseWriter, r *http.Request) {
	if s.aiClient == nil {
		http.Error(w, "AI not configured — set ANTHROPIC_API_KEY", http.StatusServiceUnavailable)
		return
	}
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}
	var body struct {
		Hint string `json:"hint"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Hint == "" {
		http.Error(w, "hint is required", http.StatusBadRequest)
		return
	}

	prompt := fmt.Sprintf(
		"Create a TTRPG world note for: %s\n\nRespond with exactly two lines:\nTitle: <short name>\nContent: <2-3 sentence description>",
		body.Hint,
	)
	generated, err := s.aiClient.Generate(r.Context(), prompt)
	if err != nil {
		http.Error(w, "AI error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	title, content := parseGeneratedNote(generated)
	if title == "" {
		title = body.Hint
	}
	if content == "" {
		content = generated
	}

	noteID, err := s.db.CreateWorldNote(id, title, content, "npc")
	if err != nil {
		http.Error(w, "db: "+err.Error(), http.StatusInternalServerError)
		return
	}

	notes, err := s.db.SearchWorldNotes(id, "", "", "")
	if err != nil {
		http.Error(w, "fetch note: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var created *db.WorldNote
	for i := range notes {
		if notes[i].ID == noteID {
			created = &notes[i]
			break
		}
	}
	if created == nil {
		http.Error(w, "note not found after create", http.StatusInternalServerError)
		return
	}

	s.bus.Publish(Event{Type: EventWorldNoteCreated, Payload: map[string]any{"note_id": noteID, "title": title}})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created) //nolint:errcheck
}

// parseGeneratedNote extracts title and content from a "Title: ...\nContent: ..." response.
func parseGeneratedNote(raw string) (title, content string) {
	for _, line := range strings.Split(raw, "\n") {
		if after, found := strings.CutPrefix(line, "Title: "); found {
			title = strings.TrimSpace(after)
		}
		if after, found := strings.CutPrefix(line, "Content: "); found {
			content = strings.TrimSpace(after)
		}
	}
	return
}

func (s *Server) handleGenerateRecap(w http.ResponseWriter, r *http.Request) {
	if s.aiClient == nil {
		http.Error(w, "AI not configured — set ANTHROPIC_API_KEY", http.StatusServiceUnavailable)
		return
	}
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}

	summary, err := s.buildRecap(r.Context(), id)
	if err != nil {
		http.Error(w, "recap: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.db.UpdateSessionSummary(id, summary); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.bus.Publish(Event{Type: EventSessionUpdated, Payload: map[string]any{
		"session_id": id,
		"summary":    summary,
	}})
	writeJSON(w, map[string]string{"summary": summary})
}

// buildRecap reads messages and dice rolls, builds a prompt, and calls the AI.
func (s *Server) buildRecap(ctx context.Context, sessionID int64) (string, error) {
	msgs, err := s.db.ListMessages(sessionID)
	if err != nil {
		return "", fmt.Errorf("list messages: %w", err)
	}
	rolls, err := s.db.ListDiceRolls(sessionID)
	if err != nil {
		return "", fmt.Errorf("list rolls: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("Write a 2-3 sentence narrative recap of this TTRPG session.\n\nMessages:\n")
	for _, m := range msgs {
		fmt.Fprintf(&sb, "[%s]: %s\n", m.Role, m.Content)
	}
	sb.WriteString("\nDice rolls:\n")
	for _, r := range rolls {
		fmt.Fprintf(&sb, "%s = %d\n", r.Expression, r.Result)
	}

	return s.aiClient.Generate(ctx, sb.String())
}

func (s *Server) handleGetTimeline(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	entries, err := s.db.GetSessionTimeline(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if entries == nil {
		entries = []db.TimelineEntry{}
	}
	writeJSON(w, entries)
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
