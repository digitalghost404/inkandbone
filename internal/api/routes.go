package api

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	mathrand "math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	advancement "github.com/digitalghost404/inkandbone/internal/ruleset"
	"github.com/digitalghost404/inkandbone/internal/ai"
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

func (s *Server) handlePatchWorldNotePersonality(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var body struct {
		PersonalityJSON string `json:"personality_json"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := s.db.UpdateWorldNotePersonality(id, body.PersonalityJSON); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	s.bus.Publish(Event{Type: EventWorldNoteUpdated, Payload: map[string]any{"note_id": id}})
	w.WriteHeader(http.StatusOK)
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

	ext := strings.ToLower(filepath.Ext(filepath.Base(header.Filename)))
	allowedMapExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowedMapExts[ext] {
		http.Error(w, "unsupported image format", http.StatusBadRequest)
		return
	}
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

func (s *Server) handleCreateMessage(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	var body struct {
		Role    string `json:"role"`
		Content string `json:"content"`
		Whisper bool   `json:"whisper"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.Role != "user" && body.Role != "assistant" {
		http.Error(w, "role must be user or assistant", http.StatusBadRequest)
		return
	}
	if body.Content == "" {
		http.Error(w, "content is required", http.StatusBadRequest)
		return
	}
	msgID, err := s.db.CreateMessage(id, body.Role, body.Content, body.Whisper)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.bus.Publish(Event{Type: EventMessageCreated, Payload: map[string]any{
		"session_id": id,
		"message_id": msgID,
		"role":       body.Role,
	}})
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handlePatchCampaign(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}
	var body struct {
		Active *bool `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.Active == nil {
		http.Error(w, "active is required", http.StatusBadRequest)
		return
	}
	var err error
	if *body.Active {
		err = s.db.ReopenCampaign(id)
	} else {
		err = s.db.CloseCampaign(id)
		if err == nil {
			// Clear active context settings so the UI resets to blank.
			for _, key := range []string{"active_campaign_id", "active_character_id", "active_session_id"} {
				_ = s.db.SetSetting(key, "")
			}
		}
	}
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handlePatchSession(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	var body struct {
		Summary   *string `json:"summary"`
		Notes     *string `json:"notes"`
		SceneTags *string `json:"scene_tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	payload := map[string]any{"session_id": id}
	if body.Summary != nil {
		if err := s.db.UpdateSessionSummary(id, *body.Summary); err != nil {
			if strings.Contains(err.Error(), "not found") {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		payload["summary"] = *body.Summary
	}
	if body.Notes != nil {
		if err := s.db.UpdateSessionNotes(id, *body.Notes); err != nil {
			if strings.Contains(err.Error(), "not found") {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		payload["notes"] = *body.Notes
	}
	if body.SceneTags != nil {
		if err := s.db.UpdateSceneTags(id, *body.SceneTags); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		payload["scene_tags"] = *body.SceneTags
	}
	s.bus.Publish(Event{Type: EventSessionUpdated, Payload: payload})
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
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var body struct {
		Hint string `json:"hint"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.Hint == "" {
		http.Error(w, "hint is required", http.StatusBadRequest)
		return
	}

	prompt := fmt.Sprintf(
		"Create a TTRPG world note for: %s\n\nRespond with exactly two lines:\nTitle: <short name>\nContent: <2-3 sentence description>",
		body.Hint,
	)
	generated, err := s.aiClient.Generate(r.Context(), prompt, 256)
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

	created, err := s.db.GetWorldNote(noteID)
	if err != nil {
		http.Error(w, "fetch note: "+err.Error(), http.StatusInternalServerError)
		return
	}

	s.bus.Publish(Event{Type: EventWorldNoteCreated, Payload: map[string]any{"note_id": noteID, "title": title}})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created) //nolint:errcheck
}

const mapSystemPrompt = `You are a cartographer creating SVG tactical maps for tabletop roleplaying games. Generate a complete, valid SVG map based on the story context provided.

Rules:
- Output ONLY the SVG markup. Start with <svg and end with </svg>. No prose, no code fences.
- Use: viewBox="0 0 800 600" width="800" height="600"
- Background rectangle: fill="#0f0e0a"
- Walls, borders, and structural lines: stroke="#3a3020" or "#c9a84c" fill="none"
- Area fills: semi-transparent darks like fill="#1a1710" or fill="#141208"
- Text labels: fill="#d4c5a0" font-family="serif" font-size="11"
- Include 5-10 named locations relevant to the story
- Connect areas with corridors or paths
- Add simple decorative shapes: pillars (circles), doors (rectangles), etc.
- Keep it readable and atmospheric`

func (s *Server) handleGenerateMap(w http.ResponseWriter, r *http.Request) {
	if s.aiClient == nil {
		http.Error(w, "AI not configured — set ANTHROPIC_API_KEY", http.StatusServiceUnavailable)
		return
	}
	responder, ok := s.aiClient.(ai.Responder)
	if !ok {
		http.Error(w, "AI client does not support chat", http.StatusServiceUnavailable)
		return
	}

	id, ok2 := parsePathID(r, "id")
	if !ok2 {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}

	var body struct {
		Name    string `json:"name"`
		Context string `json:"context"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.Name == "" {
		body.Name = "Generated Map"
	}

	prompt := "Generate a map for this TTRPG setting:\n\n" + body.Context
	history := []ai.ChatMessage{{Role: "user", Content: prompt}}

	svgRaw, err := responder.Respond(r.Context(), mapSystemPrompt, history, 4096)
	if err != nil {
		http.Error(w, "AI error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	svgContent := extractSVG(svgRaw)
	if svgContent == "" {
		http.Error(w, "AI did not return valid SVG", http.StatusInternalServerError)
		return
	}

	destDir := filepath.Join(s.dataDir, "maps")
	if err := os.MkdirAll(destDir, 0750); err != nil {
		http.Error(w, "mkdir: "+err.Error(), http.StatusInternalServerError)
		return
	}
	filename := "map_" + randomHex(8) + ".svg"
	if err := os.WriteFile(filepath.Join(destDir, filename), []byte(svgContent), 0640); err != nil {
		http.Error(w, "write file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	mapID, err := s.db.CreateMap(id, body.Name, "maps/"+filename)
	if err != nil {
		http.Error(w, "db: "+err.Error(), http.StatusInternalServerError)
		return
	}
	m, err := s.db.GetMap(mapID)
	if err != nil || m == nil {
		http.Error(w, "fetch created map", http.StatusInternalServerError)
		return
	}

	s.bus.Publish(Event{Type: EventMapCreated, Payload: map[string]any{"campaign_id": id, "map_id": mapID}})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(m) //nolint:errcheck
}

func extractSVG(s string) string {
	start := strings.Index(s, "<svg")
	end := strings.LastIndex(s, "</svg>")
	if start == -1 || end == -1 || end < start {
		return ""
	}
	return s[start : end+6]
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

// autoUpdateRecap regenerates the session recap in the background every 4 GM
// messages so the journal stays current without manual intervention.
func (s *Server) autoUpdateRecap(ctx context.Context, sessionID int64) {
	if s.aiClient == nil {
		return
	}
	msgs, err := s.db.ListMessages(sessionID)
	if err != nil {
		return
	}
	// Count assistant messages — update on every 4th one (and always on the first).
	gmCount := 0
	for _, m := range msgs {
		if m.Role == "assistant" {
			gmCount++
		}
	}
	if gmCount == 0 || gmCount%4 != 0 {
		return
	}
	summary, err := s.buildRecap(ctx, sessionID)
	if err != nil {
		return
	}
	if err := s.db.UpdateSessionSummary(sessionID, summary); err != nil {
		return
	}
	s.bus.Publish(Event{Type: EventSessionUpdated, Payload: map[string]any{
		"session_id": sessionID,
		"summary":    summary,
	}})
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

	return s.aiClient.Generate(ctx, sb.String(), 200)
}

const gmSystemPrompt = `You are the Game Master of a tabletop roleplaying game. Continue the story in response to the player's most recent action.

CONTENT POLICY — this is a private adult platform:
- All participants are consenting adults (18+). Explicit, erotic, and sexually graphic content is permitted and expected when the story calls for it.
- Write seduction, intimacy, and sexual scenes fully and without pulling back. Do not fade to black, moralize, or offer alternatives unless the player explicitly asks you to.
- Dark themes — corruption, manipulation, power dynamics, ritual, coercion between consenting fictional characters — are valid narrative territory. Lean in.
- Never break immersion to comment on content, offer warnings, or suggest the player tone things down. Stay in the story.

Write 2-4 paragraphs of immersive narrative in second person ("you"). Match the tone and vocabulary of previous GM messages. Stay consistent with the world and events already described.

End with "**What do you do?**"

RULEBOOK ADHERENCE — non-negotiable:
- If [RULEBOOK REFERENCES] appear in the world context, those rules are authoritative. Apply them exactly.
- Resolve all mechanics — combat outcomes, spell effects, skill results, conditions, durations, damage — according to the referenced rules. Do not soften, alter, or skip a rule because it is inconvenient for the narrative.
- If multiple rules are referenced, apply the most specific one. If rules conflict, apply the more restrictive.
- Do not invent mechanics, abilities, or effects that contradict the referenced rules. If the rulebook does not grant an ability, the character does not have it.
- If no rulebook reference covers the action, rule conservatively: default to standard genre conventions and do not escalate lethality or grant powers beyond what is established.

DICE ROLLS — follow these strictly:
- If the [WORLD STATE] contains a [DICE ROLL] block, you MUST incorporate the result into the narrative. The outcome is fixed. Do not invent a different result.
- Narrate success vividly. Narrate failure with consequences that still move the story forward.
- Never say "you rolled a 14" or reference numbers in the prose. Translate the mechanical result into fiction.
- Briefly explain in one sentence of in-world flavor why the roll mattered. Keep it light, never lecture.

WRITING RULES:
- Vary sentence length. Short sentences hit hard. Fragments work.
- Repeat character names freely. Never synonym-chain ("the warrior" → "the combatant").
- Concrete sensory detail, not vague declarations ("something shifted").
- No academic hedges, canned openers, or rhetorical questions as transitions.

Begin immediately with story prose. No preamble, no meta-commentary.`

// buildWorldContext assembles a [WORLD STATE] block injected into the GM system prompt.
func (s *Server) buildWorldContext(ctx context.Context, sessionID int64) string {
	var sb strings.Builder
	sb.WriteString("[WORLD STATE]\n")

	sess, err := s.db.GetSession(sessionID)
	if err != nil || sess == nil {
		sb.WriteString("Session summary: none\n")
		sb.WriteString("Recent world notes: none\n")
		sb.WriteString("Active combat: no\n")
		sb.WriteString("[/WORLD STATE]")
		return sb.String()
	}

	summary := sess.Summary
	if summary == "" {
		summary = "none"
	}
	fmt.Fprintf(&sb, "Session summary: %s\n", summary)

	if charIDStr, err := s.db.GetSetting("active_character_id"); err == nil && charIDStr != "" {
		if charID, err := strconv.ParseInt(charIDStr, 10, 64); err == nil {
			if char, err := s.db.GetCharacter(charID); err == nil && char != nil {
				fmt.Fprintf(&sb, "Player character name: %s\n", char.Name)
				// Inject common identity fields so the GM always knows the character's role.
				if char.DataJSON != "" {
					var stats map[string]any
					if err := json.Unmarshal([]byte(char.DataJSON), &stats); err == nil {
						for _, field := range []string{"archetype", "class", "race", "faction", "keywords", "species", "metatype", "playbook", "culture"} {
							if v, ok := stats[field].(string); ok && v != "" {
								fmt.Fprintf(&sb, "Character %s: %s\n", field, v)
							}
						}
					}
				}
			}
		}
	}

	notes, err := s.db.ListRecentWorldNotes(sess.CampaignID, 5)
	if err == nil && len(notes) > 0 {
		titles := make([]string, len(notes))
		for i, n := range notes {
			titles[i] = n.Title
		}
		fmt.Fprintf(&sb, "Recent world notes: %s\n", strings.Join(titles, ", "))
	} else {
		sb.WriteString("Recent world notes: none\n")
	}

	enc, err := s.db.GetActiveEncounter(sessionID)
	if err == nil && enc != nil {
		combatants, err := s.db.ListCombatants(enc.ID)
		if err == nil && len(combatants) > 0 {
			names := make([]string, len(combatants))
			for i, c := range combatants {
				names[i] = c.Name
			}
			fmt.Fprintf(&sb, "Active combat: yes (%s)\n", strings.Join(names, ", "))
		} else {
			fmt.Fprintf(&sb, "Active combat: yes (%s)\n", enc.Name)
		}
	} else {
		sb.WriteString("Active combat: no\n")
	}

	// Active objectives
	objs, err := s.db.ListObjectives(sess.CampaignID)
	if err == nil {
		var active []db.Objective
		for _, o := range objs {
			if o.Status == "active" {
				active = append(active, o)
			}
		}
		if len(active) > 0 {
			sb.WriteString("[ACTIVE OBJECTIVES]\n")
			for _, o := range active {
				if o.Description != "" {
					fmt.Fprintf(&sb, "- %s (%s)\n", o.Title, o.Description)
				} else {
					fmt.Fprintf(&sb, "- %s\n", o.Title)
				}
			}
			sb.WriteString("[/ACTIVE OBJECTIVES]\n")
		}
	}

	// NPC personality cards — only inject NPCs mentioned in session summary to bound token cost.
	// If summary is empty, fall back to the 3 most recent NPC notes.
	npcNotes, err := s.db.SearchWorldNotes(sess.CampaignID, "", "npc", "")
	if err == nil {
		summaryLower := strings.ToLower(sess.Summary)
		var filteredNPCs []db.WorldNote
		for _, n := range npcNotes {
			if n.PersonalityJSON == "" {
				continue
			}
			if summaryLower == "" || summaryLower == "none" {
				filteredNPCs = append(filteredNPCs, n)
				if len(filteredNPCs) >= 3 {
					break
				}
				continue
			}
			if strings.Contains(summaryLower, strings.ToLower(n.Title)) {
				filteredNPCs = append(filteredNPCs, n)
			}
		}
		for _, n := range filteredNPCs {
			var p map[string]any
			if err := json.Unmarshal([]byte(n.PersonalityJSON), &p); err != nil {
				continue
			}
			fmt.Fprintf(&sb, "[NPC: %s]\n", n.Title)
			if traits, ok := p["traits"]; ok {
				switch v := traits.(type) {
				case []any:
					strs := make([]string, 0, len(v))
					for _, t := range v {
						if s, ok := t.(string); ok {
							strs = append(strs, s)
						}
					}
					if len(strs) > 0 {
						fmt.Fprintf(&sb, "Traits: %s\n", strings.Join(strs, ", "))
					}
				case string:
					fmt.Fprintf(&sb, "Traits: %s\n", v)
				}
			}
			if motivation, ok := p["motivation"].(string); ok && motivation != "" {
				fmt.Fprintf(&sb, "Motivation: %s\n", motivation)
			}
			sb.WriteString("[/NPC]\n")
		}
	}

	// Wrath & Glory: inject system-specific mechanics and live character resources.
	if camp, err := s.db.GetCampaign(sess.CampaignID); err == nil && camp != nil {
		if rs, err := s.db.GetRuleset(camp.RulesetID); err == nil && rs != nil && rs.Name == "wrath_glory" {
			sb.WriteString("[W&G MECHANICS]\n")
			sb.WriteString("WEALTH: This campaign uses WEALTH TIER (1-5 abstract), NOT gold or coins. Never award currency amounts. Refer to 'Wealth Tier' only.\n")
			sb.WriteString("WRATH DIE: On any dice pool, a 6 on the Wrath die grants the player a Wrath token. A 1 on the Wrath die triggers a Complication set by the GM.\n")
			sb.WriteString("CORRUPTION: Characters accumulate Corruption from psychic taint, Chaos exposure, and forbidden acts. At Corruption >= Rank*2+8, the character must pass a Corruption test or gain a Mutation.\n")
			sb.WriteString("WRATH TOKENS: Spending a Wrath token lets the player re-roll any number of dice OR triggers a Soak save vs lethal damage.\n")
			sb.WriteString("GLORY: Characters earn Glory for heroic acts; 8 Glory = 1 Rank advancement.\n")
			sb.WriteString("RUIN: Ruin tracks the tide of Chaos. At Ruin 10, dark forces escalate dramatically.\n")

			// Inject live character resource values and identity if available.
			if charIDStr, err := s.db.GetSetting("active_character_id"); err == nil && charIDStr != "" {
				if charID, err := strconv.ParseInt(charIDStr, 10, 64); err == nil {
					if char, err := s.db.GetCharacter(charID); err == nil && char != nil && char.DataJSON != "" {
						var stats map[string]any
						if err := json.Unmarshal([]byte(char.DataJSON), &stats); err == nil {
							writeStatIfSet := func(key, label string) {
								if v, ok := stats[key]; ok {
									fmt.Fprintf(&sb, "%s: %v\n", label, v)
								}
							}
							// Character identity — must shape all narrative framing.
							writeStatIfSet("archetype", "Character Archetype")
							writeStatIfSet("faction", "Character Faction")
							writeStatIfSet("keywords", "Character Keywords")
							writeStatIfSet("species", "Character Species")
							// Talents drive unique abilities in play.
							if talents, ok := stats["talents"].(string); ok && talents != "" {
								fmt.Fprintf(&sb, "Character Talents: %s\n", talents)
							}
							// Live resource values.
							writeStatIfSet("rank", "Character Rank")
							writeStatIfSet("wrath", "Wrath Tokens")
							writeStatIfSet("glory", "Glory")
							writeStatIfSet("ruin", "Ruin")
							writeStatIfSet("corruption", "Corruption")
							writeStatIfSet("wounds", "Current Wounds")
							writeStatIfSet("shock", "Current Shock")
							writeStatIfSet("wealth", "Wealth Tier")
						}
					}
				}
			}
			sb.WriteString("[/W&G MECHANICS]\n")
		}
	}

	sb.WriteString("[/WORLD STATE]")
	return sb.String()
}

// mechanicKeywords maps trigger words in a player message to implied rulebook search terms.
// This ensures that "I attack" also searches for "combat" even if the word isn't in the message.
var mechanicKeywords = map[string][]string{
	"attack":   {"combat", "attack", "damage"},
	"fight":    {"combat", "fighting"},
	"hit":      {"combat", "attack"},
	"stab":     {"combat", "weapon", "damage"},
	"shoot":    {"ranged", "combat"},
	"cast":     {"spell", "magic", "casting"},
	"spell":    {"spell", "magic"},
	"magic":    {"magic", "spell"},
	"sneak":    {"stealth", "sneak"},
	"hide":     {"stealth", "hiding"},
	"steal":    {"stealth", "thievery"},
	"persuade": {"social", "persuasion"},
	"deceive":  {"deception", "social"},
	"intimidate": {"intimidation", "social"},
	"climb":    {"athletics", "climbing"},
	"swim":     {"athletics", "swimming"},
	"jump":     {"athletics", "jumping"},
	"search":   {"investigation", "searching"},
	"investigate": {"investigation"},
	"heal":     {"healing", "medicine"},
	"dodge":    {"dodge", "defense"},
	"run":      {"movement", "speed"},
	"flee":     {"movement", "speed"},
	"lockpick": {"thievery", "locks"},
	"pick":     {"thievery"},
	"craft":    {"crafting"},
	"ritual":   {"ritual", "magic"},
	"pray":     {"prayer", "divine"},
}

// appendRulebookContext searches uploaded rulebook chunks for keywords from the player's
// message — including implied mechanic terms — and injects matching chunks into the world
// context block so the GM is bound by the actual rules. At most 5 chunks are injected.
func (s *Server) appendRulebookContext(ctx context.Context, sessionID int64, playerMsg string, worldCtx *string) {
	sess, err := s.db.GetSession(sessionID)
	if err != nil || sess == nil {
		return
	}
	camp, err := s.db.GetCampaign(sess.CampaignID)
	if err != nil || camp == nil {
		return
	}

	stopWords := map[string]bool{
		"that": true, "this": true, "with": true, "from": true, "they": true,
		"have": true, "been": true, "will": true, "your": true, "their": true,
		"what": true, "when": true, "where": true, "which": true, "there": true,
		"would": true, "could": true, "should": true, "about": true, "into": true,
		"also": true, "then": true, "them": true, "over": true, "just": true,
	}

	seenWords := map[string]bool{}
	var keywords []string

	addKeyword := func(w string) {
		if !seenWords[w] {
			seenWords[w] = true
			keywords = append(keywords, w)
		}
	}

	// Extract explicit words from the player message.
	for _, raw := range strings.Fields(playerMsg) {
		w := strings.ToLower(strings.Trim(raw, ".,!?;:\"'()[]"))
		if len(w) > 3 && !stopWords[w] {
			addKeyword(w)
			// Expand to implied mechanic terms (e.g. "attack" → also search "combat", "damage").
			if extras, ok := mechanicKeywords[w]; ok {
				for _, extra := range extras {
					addKeyword(extra)
				}
			}
		}
		if len(keywords) >= 12 {
			break
		}
	}

	if len(keywords) == 0 {
		return
	}

	seenChunks := map[int64]bool{}
	var chunks []db.RulebookChunk
	for _, kw := range keywords {
		results, err := s.db.SearchRulebookChunks(camp.RulesetID, kw)
		if err != nil {
			continue
		}
		for _, c := range results {
			if !seenChunks[c.ID] {
				seenChunks[c.ID] = true
				chunks = append(chunks, c)
				if len(chunks) >= 5 {
					break
				}
			}
		}
		if len(chunks) >= 5 {
			break
		}
	}
	if len(chunks) == 0 {
		return
	}

	const maxChunkChars = 1200 // per-chunk content limit
	const maxTotalChars = 5000 // total rulebook injection limit

	var sb strings.Builder
	sb.WriteString("\n[RULEBOOK REFERENCES]\n")
	totalChars := 0
	for _, c := range chunks {
		content := c.Content
		if len(content) > maxChunkChars {
			content = content[:maxChunkChars] + "…"
		}
		entry := fmt.Sprintf("## %s (from %s)\n%s\n\n", c.Heading, c.Source, content)
		if totalChars+len(entry) > maxTotalChars {
			break
		}
		sb.WriteString(entry)
		totalChars += len(entry)
	}
	sb.WriteString("[/RULEBOOK REFERENCES]")
	*worldCtx += sb.String()
}

func (s *Server) handleGMRespond(w http.ResponseWriter, r *http.Request) {
	if s.aiClient == nil {
		http.Error(w, "AI not configured — set ANTHROPIC_API_KEY", http.StatusServiceUnavailable)
		return
	}
	gmResponder, ok := s.aiClient.(ai.Responder)
	if !ok {
		http.Error(w, "AI client does not support chat", http.StatusServiceUnavailable)
		return
	}

	id, ok2 := parsePathID(r, "id")
	if !ok2 {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}

	msgs, err := s.db.ListMessages(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(msgs) == 0 || msgs[len(msgs)-1].Role != "user" {
		http.Error(w, "no player message to respond to", http.StatusBadRequest)
		return
	}

	var history []ai.ChatMessage
	for _, m := range msgs {
		if m.Whisper {
			continue
		}
		history = append(history, ai.ChatMessage{Role: m.Role, Content: m.Content})
	}
	// Cap history to the last 30 messages to bound input token cost.
	// The session summary in buildWorldContext covers long-term memory.
	const historyWindow = 30
	if len(history) > historyWindow {
		history = history[len(history)-historyWindow:]
		for len(history) > 0 && history[0].Role != "user" {
			history = history[1:]
		}
	}

	worldCtx := s.buildWorldContext(r.Context(), id)
	systemPrompt := worldCtx + "\n\n" + gmSystemPrompt

	response, err := gmResponder.Respond(r.Context(), systemPrompt, history, 4096)
	if err != nil {
		http.Error(w, "AI error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	msgID, err := s.db.CreateMessage(id, "assistant", response, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.bus.Publish(Event{Type: EventMessageCreated, Payload: map[string]any{
		"session_id": id,
		"message_id": msgID,
		"role":       "assistant",
	}})
	w.WriteHeader(http.StatusCreated)
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

// --- Feature 1: Streaming GM Text (SSE) ---

func (s *Server) handleGMRespondStream(w http.ResponseWriter, r *http.Request) {
	if s.aiClient == nil {
		http.Error(w, "AI not configured — set ANTHROPIC_API_KEY", http.StatusServiceUnavailable)
		return
	}
	streamer, ok := s.aiClient.(ai.Streamer)
	if !ok {
		http.Error(w, "AI client does not support streaming", http.StatusServiceUnavailable)
		return
	}

	id, ok2 := parsePathID(r, "id")
	if !ok2 {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}

	msgs, err := s.db.ListMessages(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(msgs) == 0 || msgs[len(msgs)-1].Role != "user" {
		http.Error(w, "no player message to respond to", http.StatusBadRequest)
		return
	}

	var history []ai.ChatMessage
	for _, m := range msgs {
		if m.Whisper {
			continue
		}
		history = append(history, ai.ChatMessage{Role: m.Role, Content: m.Content})
	}
	// Cap history to the last 30 messages to bound input token cost.
	// The session summary in buildWorldContext covers long-term memory.
	const historyWindow = 30
	if len(history) > historyWindow {
		history = history[len(history)-historyWindow:]
		for len(history) > 0 && history[0].Role != "user" {
			history = history[1:]
		}
	}

	// Check if the player's action requires a dice roll under the active ruleset.
	// Do this before building the system prompt so the result can be injected.
	lastPlayerMsg := msgs[len(msgs)-1].Content
	roll := s.checkAndExecuteRoll(r.Context(), id, lastPlayerMsg)

	worldCtx := s.buildWorldContext(r.Context(), id)
	s.appendRulebookContext(r.Context(), id, lastPlayerMsg, &worldCtx)
	if roll != nil {
		outcome := "FAILURE"
		if roll.Success {
			outcome = "SUCCESS"
		}
		dcNote := ""
		if roll.DC > 0 {
			dcNote = fmt.Sprintf(" against DC %d", roll.DC)
		}
		worldCtx += fmt.Sprintf(
			"\n[DICE ROLL]\nAction required a %s check%s.\nReason: %s\nRoll: %s = %d — %s\n[/DICE ROLL]",
			roll.Attribute, dcNote, roll.Reason, roll.Expression, roll.Total, outcome,
		)
		if !roll.Success {
			worldCtx += "\n[GM DIRECTION]\nThe player's action FAILED. Narrate a setback, complication, or consequence. Do not give them what they wanted. Make failure interesting.\n[/GM DIRECTION]"
		}
	}
	systemPrompt := worldCtx + "\n\n" + gmSystemPrompt

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")

	fullText, err := streamer.StreamRespond(r.Context(), systemPrompt, history, 4096, w)
	if err != nil {
		// Headers already sent; can't send HTTP error status, just log and return
		log.Printf("gm-respond-stream: StreamRespond error: %v", err)
		return
	}

	if fullText == "" {
		return
	}

	msgID, err := s.db.CreateMessage(id, "assistant", fullText, false)
	if err != nil {
		return
	}
	s.bus.Publish(Event{Type: EventMessageCreated, Payload: map[string]any{
		"session_id": id,
		"message_id": msgID,
		"role":       "assistant",
	}})

	go s.extractNPCs(context.Background(), id, fullText)
	go s.autoGenerateMap(context.Background(), id, fullText)
	go s.autoUpdateCharacterStats(context.Background(), id, lastPlayerMsg, fullText)
	go s.autoUpdateRecap(context.Background(), id)
	go s.autoDetectObjectives(context.Background(), id, fullText)
	go s.autoExtractItems(context.Background(), id, fullText)
	go s.autoUpdateCurrency(context.Background(), id, fullText)
	tensionText := fullText
	if roll != nil && !roll.Success {
		tensionText = "critical failure " + fullText
	}
	go s.autoUpdateTension(id, tensionText)
	go s.autoUpdateSceneTags(context.Background(), id, fullText)
}

// autoGenerateMap detects if the GM response introduces a new location and, if
// so, generates an SVG map for it automatically. Runs in a background goroutine.
func (s *Server) autoGenerateMap(ctx context.Context, sessionID int64, gmText string) {
	completer, ok := s.aiClient.(ai.Completer)
	if !ok {
		return
	}
	responder, ok2 := s.aiClient.(ai.Responder)
	if !ok2 {
		return
	}

	sess, err := s.db.GetSession(sessionID)
	if err != nil || sess == nil {
		return
	}

	existing, _ := s.db.ListMaps(sess.CampaignID)
	existingNames := make([]string, len(existing))
	for i, m := range existing {
		existingNames[i] = m.Name
	}

	detectPrompt := fmt.Sprintf(`You are a TTRPG map assistant. Analyze this story passage.

Does this passage describe or establish a NAMED, visually distinct location? This includes: taverns, dungeons, caves, city streets, forests, ships, markets, ruins, castles, chambers, wilderness areas — any named place with atmosphere or layout details.

Return ONLY JSON (no explanation, no markdown):
- If a named location is clearly established: {"new_location":true,"name":"<exact location name>","context":"<50-word description: layout, atmosphere, key features for map generation>"}
- If no named location or purely abstract/transitional text: {"new_location":false}

Story passage:
%s`, gmText)

	raw, err := completer.Generate(ctx, detectPrompt, 128)
	if err != nil {
		return
	}

	raw = strings.TrimSpace(raw)
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return
	}

	var loc struct {
		NewLocation bool   `json:"new_location"`
		Name        string `json:"name"`
		Context     string `json:"context"`
	}
	if err := json.Unmarshal([]byte(raw[start:end+1]), &loc); err != nil || !loc.NewLocation || loc.Name == "" {
		return
	}

	locNameLower := strings.ToLower(loc.Name)
	for _, name := range existingNames {
		if strings.ToLower(name) == locNameLower {
			return
		}
	}

	mapPrompt := "Generate a map for this TTRPG setting:\n\n" + loc.Context
	svgRaw, err := responder.Respond(ctx, mapSystemPrompt, []ai.ChatMessage{{Role: "user", Content: mapPrompt}}, 4096)
	if err != nil {
		return
	}

	svgContent := extractSVG(svgRaw)
	if svgContent == "" {
		return
	}

	destDir := filepath.Join(s.dataDir, "maps")
	if err := os.MkdirAll(destDir, 0750); err != nil {
		return
	}
	filename := "map_" + randomHex(8) + ".svg"
	if err := os.WriteFile(filepath.Join(destDir, filename), []byte(svgContent), 0640); err != nil {
		return
	}

	mapID, err := s.db.CreateMap(sess.CampaignID, loc.Name, "maps/"+filename)
	if err != nil {
		return
	}
	s.bus.Publish(Event{Type: EventMapCreated, Payload: map[string]any{
		"campaign_id": sess.CampaignID,
		"map_id":      mapID,
	}})
}

// autoUpdateCharacterStats checks whether the story events require any
// character stat changes under the active ruleset (XP, wounds, level-ups,
// stress, etc.) and applies them automatically.
// statChangeKeywords are signals that something mechanically significant happened.
var statChangeKeywords = []string{
	"wound", "injur", "damage", "bleed", "hurt", "dead", "die", "dies", "dying",
	"level", "experience", "xp", "exp", "advance",
	"stress", "trauma", "exhaust", "corrupt",
	"heal", "recover", "restore",
	"spend", "use", "consume", "expend",
	"critical", "fail", "success",
}

func (s *Server) autoUpdateCharacterStats(ctx context.Context, sessionID int64, playerAction, gmText string) {
	completer, ok := s.aiClient.(ai.Completer)
	if !ok {
		return
	}
	// Skip the AI call when the narrative contains no stat-change signals.
	combined := strings.ToLower(playerAction + " " + gmText)
	hasSignal := false
	for _, kw := range statChangeKeywords {
		if strings.Contains(combined, kw) {
			hasSignal = true
			break
		}
	}
	if !hasSignal {
		return
	}

	// Resolve active character.
	charIDStr, err := s.db.GetSetting("active_character_id")
	if err != nil || charIDStr == "" {
		return
	}
	charID, err := strconv.ParseInt(charIDStr, 10, 64)
	if err != nil {
		return
	}
	char, err := s.db.GetCharacter(charID)
	if err != nil || char == nil || char.DataJSON == "" || char.DataJSON == "{}" {
		return
	}

	// Resolve ruleset via session → campaign.
	sess, err := s.db.GetSession(sessionID)
	if err != nil || sess == nil {
		return
	}
	camp, err := s.db.GetCampaign(sess.CampaignID)
	if err != nil || camp == nil {
		return
	}
	ruleset, err := s.db.GetRuleset(camp.RulesetID)
	if err != nil || ruleset == nil {
		return
	}

	schema := ruleset.SchemaJSON
	if len(schema) > 1200 {
		schema = schema[:1200]
	}

	prompt := fmt.Sprintf(`You are a TTRPG rules engine. Based on what just happened in the story, determine which character stats need to change according to the ruleset rules.

Ruleset: %s
Rules schema (excerpt): %s

Current character stats (JSON):
%s

What just happened:
Player: %s
GM: %s

Return ONLY a JSON object with the fields that must change and their new values. Rules to follow:
- Only update fields that already exist in the current stats JSON above.
- Apply all relevant ruleset mechanics: HP/wound changes from combat outcomes, XP gains from significant events, level-ups when thresholds are met, stress or corruption changes, resources spent or gained.
- When leveling up, also update all derived stats that change with the new level per the ruleset.
- If nothing needs to change, return {}.
- No explanation, no markdown — just the JSON object.`, ruleset.Name, schema, char.DataJSON, playerAction, gmText)

	raw, err := completer.Generate(ctx, prompt, 256)
	if err != nil {
		return
	}

	raw = strings.TrimSpace(raw)
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return
	}

	var patch map[string]any
	if err := json.Unmarshal([]byte(raw[start:end+1]), &patch); err != nil || len(patch) == 0 {
		return
	}

	// Unmarshal existing stats.
	var current map[string]any
	if err := json.Unmarshal([]byte(char.DataJSON), &current); err != nil {
		return
	}

	// Capture XP before applying patch.
	xpFieldKey := advancement.XPKey(ruleset.Name)
	beforeXP := 0
	if v, ok := current[xpFieldKey].(float64); ok {
		beforeXP = int(v)
	}

	for k, v := range patch {
		if _, exists := current[k]; exists {
			current[k] = v
		}
	}

	// Capture XP after patch.
	afterXP := 0
	if v, ok := current[xpFieldKey].(float64); ok {
		afterXP = int(v)
	}

	updated, err := json.Marshal(current)
	if err != nil {
		return
	}

	if err := s.db.UpdateCharacterData(charID, string(updated)); err != nil {
		return
	}
	s.bus.Publish(Event{Type: EventCharacterUpdated, Payload: map[string]any{
		"id":        charID,
		"data_json": string(updated),
	}})

	// If XP increased, suggest advancements asynchronously.
	// Use the updated stats JSON for goroutine (not stale pre-patch char.DataJSON).
	char.DataJSON = string(updated)
	if afterXP > beforeXP {
		go s.autoSuggestXPSpend(sessionID, charID, char, ruleset, current, afterXP)
	}
}

// autoSuggestXPSpend fires when XP increases after a GM response.
// It calls Claude to generate 2–3 ranked advancement suggestions and pushes
// them to the frontend as an xp_spend_suggestions WebSocket event.
// A per-session cap of 20 suggestions is enforced to avoid spam.
func (s *Server) autoSuggestXPSpend(
	sessionID, charID int64,
	char *db.Character,
	ruleset *db.Ruleset,
	stats map[string]any,
	currentXP int,
) {
	system := ruleset.Name

	// Skip systems without XP advancement (must be first).
	switch system {
	case "coc", "paranoia":
		return
	}

	const maxSuggestionsPerSession = 20

	// Atomically check-and-increment the session suggestion count.
	for {
		actual, _ := s.xpSuggestCounts.LoadOrStore(sessionID, 0)
		count := actual.(int)
		if count >= maxSuggestionsPerSession {
			return
		}
		if s.xpSuggestCounts.CompareAndSwap(sessionID, count, count+1) {
			break
		}
	}

	// Gate: can the character afford any advance?
	if !advancement.CanAffordAny(system, currentXP, char.DataJSON) {
		return
	}

	completer, ok := s.aiClient.(ai.Completer)
	if !ok {
		return
	}

	// Build context for Claude.
	talentsOwned, _ := stats["talents"].(string)
	archetypeName, _ := stats["archetype"].(string)

	// Extract W&G-specific context from stats.
	tier := 1
	if v, ok := stats["tier"].(float64); ok {
		tier = int(v)
	}
	faction, _ := stats["faction"].(string)

	// Look up archetype starting abilities to exclude from suggestions.
	var startingAbilitiesStr string
	if archetypeName != "" {
		if def, ok := advancement.WGArchetypeDefFor(archetypeName); ok {
			startingAbilitiesStr = strings.Join(def.Abilities(), ", ")
		}
	}

	var statsJSON []byte
	statsJSON, _ = json.Marshal(stats)

	prompt := fmt.Sprintf(`You are advising a tabletop RPG character on how to spend their %s (%s system).

Character: %s
Archetype: %s
Tier: %d
Faction: %s
Current %s: %d
Current stats (JSON): %s
Already-owned talents (pipe-delimited): %s
Archetype starting abilities (do NOT suggest these): %s

Cost rules for %s:
%s

Suggest 2–3 ranked advancement options. For each, output JSON with these exact fields:
- field: the stat key (e.g. "toughness", "talent:Iron Will", "level")
- display_name: human-readable name
- current_value: current numeric value (0 for unowned talents)
- new_value: value after advance
- xp_cost: XP cost (recalculate server-side; this is just for display)
- reasoning: one sentence explaining why this is a good choice

Output a JSON array only — no other text. Example:
[{"field":"toughness","display_name":"Toughness","current_value":4,"new_value":5,"xp_cost":20,"reasoning":"Boosts wounds, resilience, and determination."}]

Do NOT suggest talents already owned or archetype starting abilities.
Do NOT suggest advances the character cannot afford.
`,
		advancement.XPLabel(system), system,
		char.Name, archetypeName,
		tier,
		faction,
		advancement.XPLabel(system), currentXP,
		string(statsJSON),
		talentsOwned,
		startingAbilitiesStr,
		system,
		advancement.CostRulesDescription(system),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	raw, err := completer.Generate(ctx, prompt, 512)
	if err != nil {
		log.Printf("autoSuggestXPSpend: AI error: %v", err)
		return
	}

	// Extract JSON array from response (Claude may wrap in markdown).
	start := strings.Index(raw, "[")
	end := strings.LastIndex(raw, "]")
	if start < 0 || end <= start {
		log.Printf("autoSuggestXPSpend: no JSON array in response")
		return
	}
	raw = raw[start : end+1]

	var suggestions []map[string]any
	if err := json.Unmarshal([]byte(raw), &suggestions); err != nil {
		log.Printf("autoSuggestXPSpend: unmarshal error: %v", err)
		return
	}
	if len(suggestions) == 0 {
		return
	}

	// Recalculate XP costs server-side (never trust AI-supplied values).
	for _, sg := range suggestions {
		field, _ := sg["field"].(string)
		newValF, _ := sg["new_value"].(float64)
		newVal := int(newValF)
		cost := advancement.XPCostFor(system, field, newVal, char.DataJSON)
		sg["xp_cost"] = cost
	}

	payload := map[string]any{
		"character_id":   charID,
		"character_name": char.Name,
		"current_xp":     currentXP,
		"xp_label":       advancement.XPLabel(system),
		"suggestions":    suggestions,
	}
	s.bus.Publish(Event{Type: EventXPSpendSuggestions, Payload: payload})
}

type rollCheckResult struct {
	Expression string
	Total      int
	Attribute  string
	DC         int
	Success    bool
	Reason     string
}

// checkAndExecuteRoll asks haiku whether the player's action requires a dice
// roll under the active ruleset. If it does, the roll is executed, saved to
// the DB, and the result is returned so the GM prompt can incorporate it.
func (s *Server) checkAndExecuteRoll(ctx context.Context, sessionID int64, playerAction string) *rollCheckResult {
	completer, ok := s.aiClient.(ai.Completer)
	if !ok {
		return nil
	}

	sess, err := s.db.GetSession(sessionID)
	if err != nil || sess == nil {
		return nil
	}
	camp, err := s.db.GetCampaign(sess.CampaignID)
	if err != nil || camp == nil {
		return nil
	}
	ruleset, err := s.db.GetRuleset(camp.RulesetID)
	if err != nil || ruleset == nil {
		return nil
	}

	charStats := "none"
	if charIDStr, err := s.db.GetSetting("active_character_id"); err == nil && charIDStr != "" {
		if charID, err := strconv.ParseInt(charIDStr, 10, 64); err == nil {
			if char, err := s.db.GetCharacter(charID); err == nil && char != nil {
				charStats = char.DataJSON
			}
		}
	}

	schema := ruleset.SchemaJSON
	if len(schema) > 800 {
		schema = schema[:800]
	}

	prompt := fmt.Sprintf(`You are a TTRPG rules referee. Determine if the player action requires a dice roll under this ruleset.

Ruleset: %s
Rules schema (excerpt): %s
Character stats: %s

Player action: "%s"

If a dice roll IS required, respond with ONLY this JSON (no explanation, no markdown):
{"required":true,"expression":"1d20","attribute":"Strength","dc":15,"reason":"Forcing open a stuck door requires a Strength check (DC 15)"}

If NO dice roll is required, respond with ONLY:
{"required":false}`, ruleset.Name, schema, charStats, playerAction)

	raw, err := completer.Generate(ctx, prompt, 128)
	if err != nil {
		return nil
	}

	raw = strings.TrimSpace(raw)
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return nil
	}

	var check struct {
		Required   bool   `json:"required"`
		Expression string `json:"expression"`
		Attribute  string `json:"attribute"`
		DC         int    `json:"dc"`
		Reason     string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(raw[start:end+1]), &check); err != nil || !check.Required || check.Expression == "" {
		return nil
	}

	expr := strings.ToLower(strings.TrimSpace(check.Expression))
	count, sides := 1, 0
	if idx := strings.Index(expr, "d"); idx >= 0 {
		if idx > 0 {
			if n, err := strconv.Atoi(expr[:idx]); err == nil && n >= 1 {
				count = n
			}
		}
		if s2, err := strconv.Atoi(expr[idx+1:]); err == nil && s2 >= 1 {
			sides = s2
		}
	}
	if sides == 0 {
		return nil
	}

	rolls := make([]int, count)
	total := 0
	for i := range rolls {
		r := mathrand.Intn(sides) + 1
		rolls[i] = r
		total += r
	}

	breakdownBytes, _ := json.Marshal(rolls)
	_, _ = s.db.LogDiceRoll(sessionID, check.Expression, total, string(breakdownBytes))
	s.bus.Publish(Event{Type: EventDiceRolled, Payload: map[string]any{
		"session_id": sessionID,
		"expression": check.Expression,
		"result":     total,
	}})

	return &rollCheckResult{
		Expression: check.Expression,
		Total:      total,
		Attribute:  check.Attribute,
		DC:         check.DC,
		Success:    check.DC == 0 || total >= check.DC,
		Reason:     check.Reason,
	}
}

// extractNPCs uses the AI to extract newly introduced named NPCs from a GM
// response and adds any that don't already exist in the session roster.
// It also removes NPCs that are dead, captured, permanently gone, or otherwise
// no longer relevant to the story.
func (s *Server) extractNPCs(ctx context.Context, sessionID int64, gmText string) {
	completer, ok := s.aiClient.(ai.Completer)
	if !ok {
		return
	}

	existing, err := s.db.ListSessionNPCs(sessionID)
	if err != nil {
		return
	}

	var charName string
	if charIDStr, err := s.db.GetSetting("active_character_id"); err == nil && charIDStr != "" {
		if charID, err := strconv.ParseInt(charIDStr, 10, 64); err == nil {
			if char, err := s.db.GetCharacter(charID); err == nil && char != nil {
				charName = char.Name
			}
		}
	}

	// Build existing NPC list for the prompt (with IDs so AI can reference them for removal).
	type npcEntry struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}
	knownNames := make([]string, 0, len(existing))
	knownEntries := make([]npcEntry, 0, len(existing))
	for _, n := range existing {
		knownNames = append(knownNames, n.Name)
		knownEntries = append(knownEntries, npcEntry{ID: n.ID, Name: n.Name})
	}
	knownJSON, _ := json.Marshal(knownEntries)

	excludeClause := ""
	if charName != "" {
		excludeClause = fmt.Sprintf("\nNever add the player character \"%s\" as an NPC.", charName)
	}

	prompt := fmt.Sprintf(`You are a TTRPG NPC roster manager. Analyze this story passage.

Already-tracked NPCs (JSON array with id and name): %s%s

Story passage:
%s

Return ONLY a JSON object with two fields:
- "add": array of {name, note} for brand-new named NPCs that appear in this passage and are NOT already tracked. note = one sentence describing who they are. Empty array if none.
- "remove": array of ids from the tracked list for NPCs that are now definitively gone — dead, killed, permanently fled, captured offscreen, dissolved, destroyed, or otherwise will never interact with the player again. Be confident but not trigger-happy: only remove when the story clearly confirms they are gone. Empty array if none.

Example: {"add":[{"name":"Torvan","note":"A scarred mercenary guarding the gate"}],"remove":[12,7]}
If nothing changed: {"add":[],"remove":[]}
No explanation, no markdown.`, string(knownJSON), excludeClause, gmText)

	raw, err := completer.Generate(ctx, prompt, 384)
	if err != nil {
		return
	}

	raw = strings.TrimSpace(raw)
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		// Fallback: try legacy array format
		start = strings.Index(raw, "[")
		end = strings.LastIndex(raw, "]")
		if start < 0 || end <= start {
			return
		}
		var npcs []struct {
			Name string `json:"name"`
			Note string `json:"note"`
		}
		if err := json.Unmarshal([]byte(raw[start:end+1]), &npcs); err != nil {
			return
		}
		changed := 0
		for _, npc := range npcs {
			if npc.Name == "" || npc.Name == charName {
				continue
			}
			if _, err := s.db.CreateSessionNPC(sessionID, npc.Name, npc.Note); err == nil {
				changed++
			}
		}
		if changed > 0 {
			s.bus.Publish(Event{Type: EventNPCUpdated, Payload: map[string]any{"session_id": sessionID}})
		}
		return
	}

	var result struct {
		Add []struct {
			Name string `json:"name"`
			Note string `json:"note"`
		} `json:"add"`
		Remove []int64 `json:"remove"`
	}
	if err := json.Unmarshal([]byte(raw[start:end+1]), &result); err != nil {
		return
	}

	// Build a set of known NPC IDs for safety (only delete NPCs we actually track).
	knownIDs := make(map[int64]bool, len(existing))
	for _, n := range existing {
		knownIDs[n.ID] = true
	}
	// Also build a set of known names (case-insensitive) to avoid duplicates.
	knownNamesLower := make(map[string]bool, len(knownNames))
	for _, n := range knownNames {
		knownNamesLower[strings.ToLower(n)] = true
	}

	changed := 0
	for _, npc := range result.Add {
		if npc.Name == "" || npc.Name == charName {
			continue
		}
		if knownNamesLower[strings.ToLower(npc.Name)] {
			continue
		}
		if _, err := s.db.CreateSessionNPC(sessionID, npc.Name, npc.Note); err == nil {
			changed++
		}
	}
	for _, id := range result.Remove {
		if !knownIDs[id] {
			continue // safety: never delete an ID we didn't give the AI
		}
		if err := s.db.DeleteSessionNPC(id); err == nil {
			changed++
		}
	}
	if changed > 0 {
		s.bus.Publish(Event{Type: EventNPCUpdated, Payload: map[string]any{"session_id": sessionID}})
	}
	_ = knownNames // used above
}

// --- Feature 4: In-Browser Dice Roller ---

func (s *Server) handleRollDice(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	var body struct {
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.Expression == "" {
		http.Error(w, "expression is required", http.StatusBadRequest)
		return
	}

	expr := strings.ToLower(strings.TrimSpace(body.Expression))
	count := 1
	sides := 0
	if idx := strings.Index(expr, "d"); idx >= 0 {
		if idx > 0 {
			n, err := strconv.Atoi(expr[:idx])
			if err != nil || n < 1 {
				http.Error(w, "invalid dice count", http.StatusBadRequest)
				return
			}
			count = n
		}
		s2, err := strconv.Atoi(expr[idx+1:])
		if err != nil || s2 < 1 {
			http.Error(w, "invalid die sides", http.StatusBadRequest)
			return
		}
		sides = s2
	} else {
		http.Error(w, "expression must be dX or NdX format", http.StatusBadRequest)
		return
	}

	rolls := make([]int, count)
	total := 0
	for i := range rolls {
		roll := mathrand.Intn(sides) + 1
		rolls[i] = roll
		total += roll
	}

	breakdownBytes, _ := json.Marshal(rolls)
	_, err := s.db.LogDiceRoll(id, body.Expression, total, string(breakdownBytes))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.bus.Publish(Event{Type: EventDiceRolled, Payload: map[string]any{
		"session_id": id,
		"expression": body.Expression,
		"result":     total,
	}})

	writeJSON(w, map[string]any{
		"expression": body.Expression,
		"result":     total,
		"rolls":      rolls,
	})
}

// --- Feature 5: Condition Badges (PATCH combatant) ---

func (s *Server) handlePatchCombatant(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid combatant id", http.StatusBadRequest)
		return
	}
	var body struct {
		HPCurrent      *int   `json:"hp_current"`
		ConditionsJSON string `json:"conditions_json"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Read current values to fill in any unset fields
	combatants, err := s.db.ListCombatants(0)
	_ = combatants
	// We need to fetch by id — use a simple approach: update with provided values,
	// defaulting hp_current to -1 sentinel to detect missing. Instead, require
	// that callers always pass hp_current or we keep existing. Since UpdateCombatant
	// always takes both args, we need the current row.
	// Fetch current via a workaround: scan all combatants isn't efficient.
	// The simplest approach: if hp_current not provided, we still need a value.
	// We'll query directly.
	var currentHP int
	var currentConditions string
	err = s.db.SQL().QueryRowContext(r.Context(),
		"SELECT hp_current, conditions_json FROM combatants WHERE id = ?", id,
	).Scan(&currentHP, &currentConditions)
	if err != nil {
		http.Error(w, "combatant not found", http.StatusNotFound)
		return
	}

	hp := currentHP
	if body.HPCurrent != nil {
		hp = *body.HPCurrent
	}
	conditions := currentConditions
	if body.ConditionsJSON != "" {
		conditions = body.ConditionsJSON
	}

	if err := s.db.UpdateCombatant(id, hp, conditions); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.bus.Publish(Event{Type: EventCombatantUpdated, Payload: map[string]any{"combatant_id": id}})
	w.WriteHeader(http.StatusNoContent)
}

// --- Feature 8: Map Pins from Chat ---

func (s *Server) handleCreateMapPin(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid map id", http.StatusBadRequest)
		return
	}
	var body struct {
		X     float64 `json:"x"`
		Y     float64 `json:"y"`
		Label string  `json:"label"`
		Note  string  `json:"note"`
		Color string  `json:"color"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	pinID, err := s.db.AddMapPin(id, body.X, body.Y, body.Label, body.Note, body.Color)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pins, err := s.db.ListMapPins(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var created *db.MapPin
	for i := range pins {
		if pins[i].ID == pinID {
			created = &pins[i]
			break
		}
	}
	if created == nil {
		http.Error(w, "pin not found after create", http.StatusInternalServerError)
		return
	}

	s.bus.Publish(Event{Type: EventMapPinAdded, Payload: map[string]any{"map_id": id, "pin_id": pinID}})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created) //nolint:errcheck
}

// --- Feature 9: NPC Roster ---

func (s *Server) handleListNPCs(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	npcs, err := s.db.ListSessionNPCs(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if npcs == nil {
		npcs = []db.SessionNPC{}
	}
	writeJSON(w, npcs)
}

func (s *Server) handleCreateNPC(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	var body struct {
		Name string `json:"name"`
		Note string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	npc, err := s.db.CreateSessionNPC(id, body.Name, body.Note)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.bus.Publish(Event{Type: EventNPCUpdated, Payload: map[string]any{"session_id": id, "npc_id": npc.ID}})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(npc) //nolint:errcheck
}

func (s *Server) handlePatchNPC(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid npc id", http.StatusBadRequest)
		return
	}
	var body struct {
		Note string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := s.db.UpdateSessionNPC(id, body.Note); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.bus.Publish(Event{Type: EventNPCUpdated, Payload: map[string]any{"npc_id": id}})
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleDeleteNPC(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid npc id", http.StatusBadRequest)
		return
	}
	if err := s.db.DeleteSessionNPC(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.bus.Publish(Event{Type: EventNPCUpdated, Payload: map[string]any{"npc_id": id}})
	w.WriteHeader(http.StatusNoContent)
}

// --- Feature 10: Objectives Tracker ---

func (s *Server) handleListObjectives(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}
	objectives, err := s.db.ListObjectives(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if objectives == nil {
		objectives = []db.Objective{}
	}
	writeJSON(w, objectives)
}

func (s *Server) handleCreateObjective(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}
	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		ParentID    *int64 `json:"parent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if body.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}
	if body.ParentID != nil {
		parent, err := s.db.GetObjective(*body.ParentID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if parent == nil {
			http.Error(w, "parent objective not found", http.StatusBadRequest)
			return
		}
		if parent.ParentID != nil {
			http.Error(w, "parent_id must reference a top-level objective", http.StatusBadRequest)
			return
		}
	}
	obj, err := s.db.CreateObjective(id, body.Title, body.Description, body.ParentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.bus.Publish(Event{Type: EventObjectiveUpdated, Payload: map[string]any{"campaign_id": id, "objective_id": obj.ID}})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(obj) //nolint:errcheck
}

func (s *Server) handlePatchObjective(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid objective id", http.StatusBadRequest)
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.Status == "" {
		http.Error(w, "status is required", http.StatusBadRequest)
		return
	}
	validStatuses := map[string]bool{"active": true, "completed": true, "failed": true}
	if !validStatuses[body.Status] {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}
	if err := s.db.UpdateObjectiveStatus(id, body.Status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.bus.Publish(Event{Type: EventObjectiveUpdated, Payload: map[string]any{"objective_id": id}})
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleDeleteObjective(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid objective id", http.StatusBadRequest)
		return
	}
	if err := s.db.DeleteObjective(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.bus.Publish(Event{Type: EventObjectiveUpdated, Payload: map[string]any{"objective_id": id}})
	w.WriteHeader(http.StatusNoContent)
}

// handleDeduplicateObjectives removes duplicate objectives within a campaign,
// keeping the oldest copy of each title (case-insensitive).
// POST /api/campaigns/{id}/objectives/dedup
func (s *Server) handleDeduplicateObjectives(w http.ResponseWriter, r *http.Request) {
	campaignID, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}
	n, err := s.db.DeduplicateObjectives(campaignID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if n > 0 {
		s.bus.Publish(Event{Type: EventObjectiveUpdated, Payload: map[string]any{"campaign_id": campaignID}})
	}
	writeJSON(w, map[string]any{"deleted": n})
}

var objectiveNewKeywords = []string{
	"quest", "task", "mission", "objective", "goal",
	"find", "bring", "kill", "slay", "defeat", "protect", "rescue", "retrieve", "discover", "investigate",
	"reward", "bounty", "contract", "assignment", "must", "need to", "have to",
}

// autoDetectObjectives uses the AI to detect new objectives introduced in a GM
// response and to resolve existing active objectives that were completed or failed.
// Always runs when there are active objectives (to catch resolutions/failures).
// Only skips entirely when there are no active objectives AND no new-objective keywords.
// Runs in a background goroutine.
func (s *Server) autoDetectObjectives(ctx context.Context, sessionID int64, gmText string) {
	completer, ok := s.aiClient.(ai.Completer)
	if !ok {
		return
	}

	sess, err := s.db.GetSession(sessionID)
	if err != nil || sess == nil {
		return
	}

	existing, err := s.db.ListObjectives(sess.CampaignID)
	if err != nil {
		return
	}

	// Count active objectives — if any exist, always run to catch resolutions.
	hasActive := false
	for _, o := range existing {
		if o.Status == "active" {
			hasActive = true
			break
		}
	}

	// If no active objectives, only run when new-objective keywords appear.
	if !hasActive {
		lower := strings.ToLower(gmText)
		found := false
		for _, kw := range objectiveNewKeywords {
			if strings.Contains(lower, kw) {
				found = true
				break
			}
		}
		if !found {
			return
		}
	}

	// Build a case-insensitive set of existing titles for dedup.
	existingTitleSet := make(map[string]bool, len(existing))
	type slimObjective struct {
		ID          int64  `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}
	slim := make([]slimObjective, 0, len(existing))
	for _, o := range existing {
		slim = append(slim, slimObjective{ID: o.ID, Title: o.Title, Description: o.Description, Status: o.Status})
		existingTitleSet[strings.ToLower(strings.TrimSpace(o.Title))] = true
	}
	existingJSON, err := json.Marshal(slim)
	if err != nil {
		return
	}

	// Fetch recent session history to give the AI enough context for resolution.
	// A single GM message is often too narrow; we include the last ~4000 chars of
	// prior assistant messages as a "recent story" prefix so the AI can tell whether
	// an active objective was resolved in an earlier turn.
	recentContext := ""
	if msgs, merr := s.db.ListMessages(sessionID); merr == nil {
		var sb strings.Builder
		for _, m := range msgs {
			if m.Role == "assistant" {
				sb.WriteString(m.Content)
				sb.WriteString("\n\n")
			}
		}
		prior := strings.TrimSpace(sb.String())
		// Exclude the current gmText from the prior (it's appended separately).
		prior = strings.TrimSuffix(prior, strings.TrimSpace(gmText))
		prior = strings.TrimSpace(prior)
		if len(prior) > 4000 {
			prior = prior[len(prior)-4000:]
		}
		if prior != "" {
			recentContext = "\n\nRecent prior story (for resolution context only — do NOT add objectives already tracked):\n" + prior
		}
	}

	prompt := fmt.Sprintf(`You are a TTRPG quest tracker. Analyze the story passage and update objectives.

Existing objectives (ALL statuses — do NOT suggest any title that already exists here, even paraphrased):
%s

Current story passage:
%s%s

Return ONLY a JSON object with exactly two fields:
- "new": array of {title, description} for objectives that are GENUINELY NEW and NOT already tracked above (quests given, tasks revealed, goals explicitly stated). If the goal already exists under any wording, omit it. Empty array if none.
- "resolved": array of {id, status} where "status" is "completed" or "failed". Only resolve ACTIVE objectives. Be AGGRESSIVE: if a goal has been substantially achieved, the situation clearly moved on, a target died/fled/converted/recruited, or the goal became permanently impossible — mark it. Infer from story logic; do not require explicit confirmation.

Examples resolved as completed: task explicitly done, item delivered, person converted or recruited, location secured, enemy defeated, plan executed.
Examples resolved as failed: target died before completion, location destroyed, time ran out, goal permanently blocked.

Output format: {"new":[{"title":"...","description":"..."}],"resolved":[{"id":3,"status":"completed"}]}
No changes: {"new":[],"resolved":[]}
No explanation, no markdown, no code fences.`, string(existingJSON), gmText, recentContext)

	raw, err := completer.Generate(ctx, prompt, 600)
	if err != nil {
		return
	}

	raw = strings.TrimSpace(raw)
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return
	}

	var result struct {
		New []struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		} `json:"new"`
		Resolved []struct {
			ID     int64  `json:"id"`
			Status string `json:"status"`
		} `json:"resolved"`
	}
	if err := json.Unmarshal([]byte(raw[start:end+1]), &result); err != nil {
		return
	}

	changed := 0
	for _, n := range result.New {
		if n.Title == "" {
			continue
		}
		// Application-level dedup: skip if title already exists (case-insensitive).
		key := strings.ToLower(strings.TrimSpace(n.Title))
		if existingTitleSet[key] {
			continue
		}
		if _, err := s.db.CreateObjective(sess.CampaignID, n.Title, n.Description, nil); err == nil {
			changed++
			existingTitleSet[key] = true // prevent same-batch duplicates
		}
	}
	for _, res := range result.Resolved {
		if res.Status != "completed" && res.Status != "failed" {
			continue
		}
		if err := s.db.UpdateObjectiveStatus(res.ID, res.Status); err == nil {
			changed++
		}
	}
	if changed > 0 {
		s.bus.Publish(Event{Type: EventObjectiveUpdated, Payload: map[string]any{"campaign_id": sess.CampaignID}})
	}
}

// --- Feature 11: Player Inventory ---

func (s *Server) handleListItems(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid character id", http.StatusBadRequest)
		return
	}
	items, err := s.db.ListItems(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if items == nil {
		items = []db.Item{}
	}
	writeJSON(w, items)
}

func (s *Server) handleCreateItem(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid character id", http.StatusBadRequest)
		return
	}
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Quantity    *int   `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	qty := 1
	if body.Quantity != nil {
		qty = *body.Quantity
	}
	item, err := s.db.CreateItem(id, body.Name, body.Description, qty)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.bus.Publish(Event{Type: EventItemUpdated, Payload: map[string]any{"character_id": id, "item_id": item.ID}})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item) //nolint:errcheck
}

func (s *Server) handlePatchItem(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid item id", http.StatusBadRequest)
		return
	}
	existing, err := s.db.GetItem(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if existing == nil {
		http.Error(w, "item not found", http.StatusNotFound)
		return
	}
	var body struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Quantity    *int    `json:"quantity"`
		Equipped    *bool   `json:"equipped"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	// Apply only provided fields.
	name := existing.Name
	description := existing.Description
	quantity := existing.Quantity
	equipped := existing.Equipped
	if body.Name != nil {
		name = *body.Name
	}
	if body.Description != nil {
		description = *body.Description
	}
	if body.Quantity != nil {
		quantity = *body.Quantity
	}
	if body.Equipped != nil {
		equipped = *body.Equipped
	}
	if err := s.db.UpdateItem(id, name, description, quantity, equipped); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.bus.Publish(Event{Type: EventItemUpdated, Payload: map[string]any{"character_id": existing.CharacterID, "item_id": id}})
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleDeleteItem(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid item id", http.StatusBadRequest)
		return
	}
	existing, err := s.db.GetItem(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if existing == nil {
		http.Error(w, "item not found", http.StatusNotFound)
		return
	}
	if err := s.db.DeleteItem(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.bus.Publish(Event{Type: EventItemUpdated, Payload: map[string]any{"character_id": existing.CharacterID, "item_id": id}})
	w.WriteHeader(http.StatusNoContent)
}

// autoExtractItems analyzes a GM response for items explicitly gained or lost
// by the player and updates the active character's inventory accordingly.
// Runs in a background goroutine.
func (s *Server) autoExtractItems(ctx context.Context, sessionID int64, gmText string) {
	completer, ok := s.aiClient.(ai.Completer)
	if !ok {
		return
	}

	// Resolve active character.
	charIDStr, err := s.db.GetSetting("active_character_id")
	if err != nil || charIDStr == "" {
		return
	}
	charID, err := strconv.ParseInt(charIDStr, 10, 64)
	if err != nil {
		return
	}

	prompt := fmt.Sprintf(`You are a TTRPG inventory tracker. Analyze this GM story passage.

Identify items the player character explicitly takes ownership of and will carry going forward. Also identify items explicitly lost, destroyed, or taken away.

Rules:
- GAINED: only items the player character picks up, receives, or is explicitly handed.
- Do NOT add containers or bags that are opened/searched — only add the container if the player explicitly takes it with them.
- Do NOT add items found inside something unless the player explicitly takes those items out and keeps them.
- Do NOT add items merely mentioned, seen, or examined — only items the player now owns.
- Do NOT add the same item more than once in the gained list.
- LOST: items the passage explicitly says were dropped, destroyed, given away, used up, or taken from the player.

Return ONLY a JSON object (no explanation, no markdown):
{"gained":[{"name":"...","description":"...","quantity":1}],"lost":["item name","item name"]}

If nothing changed: {"gained":[],"lost":[]}

Story passage:
%s`, gmText)

	raw, err := completer.Generate(ctx, prompt, 256)
	if err != nil {
		return
	}

	raw = strings.TrimSpace(raw)
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return
	}

	var result struct {
		Gained []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Quantity    int    `json:"quantity"`
		} `json:"gained"`
		Lost []string `json:"lost"`
	}
	if err := json.Unmarshal([]byte(raw[start:end+1]), &result); err != nil {
		return
	}

	// Load existing inventory once so we can deduplicate before inserting.
	existing, _ := s.db.ListItems(charID)
	existingNames := make(map[string]bool, len(existing))
	for _, item := range existing {
		existingNames[strings.ToLower(item.Name)] = true
	}

	changed := 0

	for _, g := range result.Gained {
		if g.Name == "" {
			continue
		}
		// Skip if already in inventory (prevents re-adding on every mention).
		if existingNames[strings.ToLower(g.Name)] {
			continue
		}
		qty := g.Quantity
		if qty <= 0 {
			qty = 1
		}
		if _, err := s.db.CreateItem(charID, g.Name, g.Description, qty); err == nil {
			existingNames[strings.ToLower(g.Name)] = true // prevent dupe within same response
			changed++
		}
	}

	if len(result.Lost) > 0 {
		for _, lostName := range result.Lost {
			lostLower := strings.ToLower(lostName)
			for _, item := range existing {
				if strings.ToLower(item.Name) == lostLower {
					if err := s.db.DeleteItem(item.ID); err == nil {
						changed++
					}
					break
				}
			}
		}
	}

	if changed > 0 {
		s.bus.Publish(Event{Type: EventItemUpdated, Payload: map[string]any{"character_id": charID}})
	}
}

// sceneTagKeywords maps each scene tag to keywords that strongly indicate it.
// Keyword matching replaces an AI call — same accuracy, zero token cost.
var sceneTagKeywords = map[string][]string{
	"battle":  {"battle", "fight", "combat", "attack", "enemy", "clash", "sword", "skirmish", "weapon", "strike", "wound", "blood", "war"},
	"dungeon": {"dungeon", "corridor", "cell", "prison", "iron door", "torch"},
	"cave":    {"cave", "cavern", "stalactite", "stalagmite", "underground", "tunnel", "grotto"},
	"forest":  {"forest", "tree", "woods", "grove", "undergrowth", "canopy", "thicket", "bark"},
	"castle":  {"castle", "throne", "tower", "battlement", "great hall", "rampart", "fortress", "keep", "parapet"},
	"tavern":  {"tavern", "inn", "alehouse", "taproom", "barmaid", "bartender", "tankard", "common room"},
	"market":  {"market", "stall", "merchant", "vendor", "bazaar", "goods", "wares"},
	"temple":  {"temple", "shrine", "altar", "priest", "prayer", "ritual", "holy", "sacred", "chapel"},
	"ruins":   {"ruins", "ruin", "crumble", "ancient", "collapse", "decay", "abandoned", "overgrown", "rubble"},
	"city":    {"city", "street", "alley", "crowd", "cobblestone", "district", "urban", "plaza"},
	"ocean":   {"ocean", "sea", "ship", "wave", "sail", "harbor", "dock", "tide", "shore"},
	"rain":    {"rain", "storm", "thunder", "lightning", "drizzle", "downpour", "soaked", "puddle"},
	"night":   {"night", "midnight", "moonlight", "dusk", "twilight"},
}

// autoUpdateSceneTags classifies the scene via keyword matching and updates the
// session's scene_tags to drive ambient audio. Skips the write if the active
// (first) tag is unchanged (stability — avoids restarting the track mid-scene).
// Uses keyword matching instead of an AI call: same accuracy, zero token cost.
func (s *Server) autoUpdateSceneTags(_ context.Context, sessionID int64, gmText string) {
	if gmText == "" {
		return
	}
	lowerText := strings.ToLower(gmText)

	bestTag := ""
	bestScore := 0
	for tag, keywords := range sceneTagKeywords {
		score := 0
		for _, kw := range keywords {
			if strings.Contains(lowerText, kw) {
				score++
			}
		}
		if score > bestScore {
			bestScore = score
			bestTag = tag
		}
	}
	if bestTag == "" {
		return
	}

	sess, err := s.db.GetSession(sessionID)
	if err != nil || sess == nil {
		return
	}

	// Tag stability: skip if the active (first) tag is unchanged.
	currentFirst := ""
	if sess.SceneTags != "" {
		currentFirst = strings.SplitN(sess.SceneTags, ",", 2)[0]
	}
	if bestTag == currentFirst {
		return
	}

	if err := s.db.UpdateSceneTags(sessionID, bestTag); err != nil {
		return
	}
	s.bus.Publish(Event{Type: EventSessionUpdated, Payload: map[string]any{
		"session_id": sessionID,
		"scene_tags": bestTag,
	}})
}

// autoUpdateTension adjusts session tension after each GM response.
// Failed dice rolls increase tension +1 (caller prepends "critical failure" to text).
// Crisis keywords in the GM text also increase tension +1.
func (s *Server) autoUpdateTension(sessionID int64, gmText string) {
	crisisKeywords := []string{"critical failure", "disaster", "catastrophe", "ambush", "trap", "betrayal", "dying"}
	lowerText := strings.ToLower(gmText)

	current, err := s.db.GetTension(sessionID)
	if err != nil {
		return
	}

	for _, kw := range crisisKeywords {
		if strings.Contains(lowerText, kw) {
			newLevel := current + 1
			_ = s.db.UpdateTension(sessionID, newLevel)
			s.bus.Publish(Event{Type: EventTensionUpdated, Payload: map[string]any{
				"session_id":    sessionID,
				"tension_level": newLevel,
			}})
			return
		}
	}
}

// autoUpdateCurrency analyzes a GM response for explicit currency transactions
// (e.g. "you receive 30 gold", "costs 15 coin") and updates the active character's
// balance accordingly. Runs in a background goroutine.
// Only fires when a specific number AND a currency word appear together.
// Publishes currency_delta in the character_updated event so the frontend can show an undo toast.
func (s *Server) autoUpdateCurrency(ctx context.Context, sessionID int64, gmText string) {
	completer, ok := s.aiClient.(ai.Completer)
	if !ok {
		return
	}

	// Skip for systems that use abstract wealth rather than tracked currency.
	if sess, err := s.db.GetSession(sessionID); err == nil && sess != nil {
		if camp, err := s.db.GetCampaign(sess.CampaignID); err == nil && camp != nil {
			if rs, err := s.db.GetRuleset(camp.RulesetID); err == nil && rs != nil {
				if rs.Name == "wrath_glory" {
					return
				}
			}
		}
	}

	// Resolve active character.
	charIDStr, err := s.db.GetSetting("active_character_id")
	if err != nil || charIDStr == "" {
		return
	}
	charID, err := strconv.ParseInt(charIDStr, 10, 64)
	if err != nil {
		return
	}

	prompt := fmt.Sprintf(`You are a TTRPG currency tracker. Analyze this GM story passage.

Extract any EXPLICIT currency transaction where BOTH a specific number AND a currency word appear together.
Currency words include: gold, gp, silver, sp, copper, cp, coin, coins, crowns, marks, ducats, dollars, credits.

Rules:
- Only extract when both a number AND a currency word are present (e.g. "30 gold", "15 coin", "5 gp").
- Positive delta = player gains currency. Negative delta = player spends or loses currency.
- If multiple transactions exist, sum them into a single delta.
- Do NOT infer amounts. "A handful of coins" or "some gold" are NOT explicit — return delta 0.
- Do NOT extract currency that belongs to NPCs unless it transfers to the player.

Return ONLY a JSON object (no explanation, no markdown):
{"delta": 0}

Story passage:
%s`, gmText)

	raw, err := completer.Generate(ctx, prompt, 64)
	if err != nil {
		return
	}

	raw = strings.TrimSpace(raw)
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return
	}

	var result struct {
		Delta int64 `json:"delta"`
	}
	if err := json.Unmarshal([]byte(raw[start:end+1]), &result); err != nil {
		return
	}
	if result.Delta == 0 {
		return
	}

	// Get current balance.
	char, err := s.db.GetCharacter(charID)
	if err != nil || char == nil {
		return
	}

	newBalance := char.CurrencyBalance + result.Delta
	if newBalance < 0 {
		newBalance = 0
	}

	if err := s.db.UpdateCharacterCurrencyBalance(charID, newBalance); err != nil {
		return
	}

	s.bus.Publish(Event{Type: EventCharacterUpdated, Payload: map[string]any{
		"id":               charID,
		"currency_balance": newBalance,
		"currency_label":   char.CurrencyLabel,
		"currency_delta":   result.Delta,
	}})
}
