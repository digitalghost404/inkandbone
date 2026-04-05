package api

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	mathrand "math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
		Summary *string `json:"summary"`
		Notes   *string `json:"notes"`
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

	return s.aiClient.Generate(ctx, sb.String())
}

const gmSystemPrompt = `You are the Game Master of a tabletop roleplaying game. Continue the story in response to the player's most recent action.

Write 2-4 paragraphs of immersive narrative in second person ("you"). Match the tone and vocabulary of previous GM messages. Stay consistent with the world and events already described.

End with "**What do you do?**"

DICE ROLLS — follow these strictly:
- If the [WORLD STATE] contains a [DICE ROLL] block, you MUST incorporate the result into the narrative. The outcome is fixed. Do not invent a different result.
- Narrate success vividly. Narrate failure with consequences that still move the story forward.
- Never say "you rolled a 14" or reference numbers in the prose. Translate the mechanical result into fiction: "Your grip holds" (success) or "Your fingers slip at the last moment" (failure).
- Briefly and naturally explain why the roll mattered, in one sentence of in-world flavor, so new players understand: e.g. "Picking the lock needed a steady hand." Keep it light, never lecture.

WRITING RULES — follow these strictly:
- Vary sentence length drastically. Short sentences hit hard. Then a longer one draws the moment out, letting weight settle. Fragments work.
- Use contractions naturally: don't, can't, it's, you're.
- NEVER use an em-dash (—) anywhere in your response, for any reason. Not for pauses, not for asides, not for interruptions. Zero em-dashes, ever.
- For dramatic pauses: use a period and start a new sentence, or use a sentence fragment. Example: "The door opens. Nothing." not "The door opens — nothing."
- For interrupted speech or trailing off: use an ellipsis (...). Example: "I wouldn't do that if I were..." not "I wouldn't do that—"
- For asides and parentheticals: use commas or parentheses. Example: "The guard (half-asleep) barely glances up." not "The guard — half-asleep — barely glances up."
- Never use "It's not X, it's Y" constructions.
- No rhetorical questions mid-sentence or as transitions.
- Repeat character names and key nouns freely. Never synonym-chain ("the warrior" → "the combatant" → "the fighter").
- Show atmosphere with concrete sensory detail, not vague declarations ("something shifted").
- Banned words: delve, tapestry, multifaceted, unpack, ascertain, whilst, moreover, furthermore, transformative, empower, elevate, realm, intricate.
- No academic hedges: "it is worth noting," "one could argue," "in light of."
- No canned openers: "As the [noun] continues to..." or "In this world..."

Do not break character. Do not summarize. Just continue the story.
CRITICAL: Begin your response immediately with story prose. Never output any preamble, meta-commentary, reasoning, or thinking. The player sees your first word. Make it story.`

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

	// NPC personality cards
	npcNotes, err := s.db.SearchWorldNotes(sess.CampaignID, "", "npc", "")
	if err == nil {
		for _, n := range npcNotes {
			if n.PersonalityJSON == "" {
				continue
			}
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

	sb.WriteString("[/WORLD STATE]")
	return sb.String()
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

	worldCtx := s.buildWorldContext(r.Context(), id)
	systemPrompt := worldCtx + "\n\n" + gmSystemPrompt

	response, err := gmResponder.Respond(r.Context(), systemPrompt, history, 2048)
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

	// Check if the player's action requires a dice roll under the active ruleset.
	// Do this before building the system prompt so the result can be injected.
	lastPlayerMsg := msgs[len(msgs)-1].Content
	roll := s.checkAndExecuteRoll(r.Context(), id, lastPlayerMsg)

	worldCtx := s.buildWorldContext(r.Context(), id)
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

	fullText, err := streamer.StreamRespond(r.Context(), systemPrompt, history, 2048, w)
	if err != nil {
		// Headers already sent; can't send HTTP error status, just return
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

	raw, err := completer.Generate(ctx, detectPrompt)
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
func (s *Server) autoUpdateCharacterStats(ctx context.Context, sessionID int64, playerAction, gmText string) {
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

	raw, err := completer.Generate(ctx, prompt)
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

	// Merge patch into existing data_json.
	var current map[string]any
	if err := json.Unmarshal([]byte(char.DataJSON), &current); err != nil {
		return
	}
	for k, v := range patch {
		if _, exists := current[k]; exists {
			current[k] = v
		}
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

	raw, err := completer.Generate(ctx, prompt)
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
func (s *Server) extractNPCs(ctx context.Context, sessionID int64, gmText string) {
	completer, ok := s.aiClient.(ai.Completer)
	if !ok {
		return
	}

	existing, err := s.db.ListSessionNPCs(sessionID)
	if err != nil {
		return
	}

	var exclude []string
	if charIDStr, err := s.db.GetSetting("active_character_id"); err == nil && charIDStr != "" {
		if charID, err := strconv.ParseInt(charIDStr, 10, 64); err == nil {
			if char, err := s.db.GetCharacter(charID); err == nil && char != nil {
				exclude = append(exclude, char.Name)
			}
		}
	}
	for _, n := range existing {
		exclude = append(exclude, n.Name)
	}

	excludeClause := ""
	if len(exclude) > 0 {
		excludeClause = fmt.Sprintf("\nDo NOT include these already-known names: %s.", strings.Join(exclude, ", "))
	}

	prompt := fmt.Sprintf(`Extract any newly introduced named NPCs from this story passage. Return ONLY a valid JSON array of objects with "name" and "note" fields. The note should be a one-sentence description of who this character is. If no new NPCs appear, return [].%s

Story passage:
%s`, excludeClause, gmText)

	raw, err := completer.Generate(ctx, prompt)
	if err != nil {
		return
	}

	raw = strings.TrimSpace(raw)
	start := strings.Index(raw, "[")
	end := strings.LastIndex(raw, "]")
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

	added := 0
	for _, npc := range npcs {
		if npc.Name == "" {
			continue
		}
		if _, err := s.db.CreateSessionNPC(sessionID, npc.Name, npc.Note); err == nil {
			added++
		}
	}
	if added > 0 {
		s.bus.Publish(Event{Type: EventNPCUpdated, Payload: map[string]any{"session_id": sessionID}})
	}
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

// autoDetectObjectives uses the AI to detect new objectives introduced in a GM
// response and to resolve existing active objectives that were completed or failed.
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

	type slimObjective struct {
		ID     int64  `json:"id"`
		Title  string `json:"title"`
		Status string `json:"status"`
	}
	slim := make([]slimObjective, 0, len(existing))
	for _, o := range existing {
		slim = append(slim, slimObjective{ID: o.ID, Title: o.Title, Status: o.Status})
	}
	existingJSON, err := json.Marshal(slim)
	if err != nil {
		return
	}

	prompt := fmt.Sprintf(`You are a TTRPG quest tracker. Analyze this story passage.

Existing objectives (JSON array): %s

Story passage:
%s

Return ONLY a JSON object with two fields:
- "new": array of {title, description} for brand-new objectives or goals introduced (quests given, tasks revealed, goals stated). Empty array if none.
- "resolved": array of {id, status} for existing active objectives that were just completed or failed. Only include ones with clear story evidence. "status" must be "completed" or "failed".

Example: {"new":[{"title":"Find the stolen data-chip","description":"Vex wants you to recover a data-chip from a House Valdris facility"}],"resolved":[{"id":3,"status":"completed"}]}

If nothing changed: {"new":[],"resolved":[]}
No explanation, no markdown.`, string(existingJSON), gmText)

	raw, err := completer.Generate(ctx, prompt)
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
		if _, err := s.db.CreateObjective(sess.CampaignID, n.Title, n.Description, nil); err == nil {
			changed++
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

Identify items that were EXPLICITLY given to, found by, or picked up by the player character. Also identify items that were EXPLICITLY lost, destroyed, or taken from the player character.

Rules:
- Only include items with clear, direct story evidence.
- Do NOT infer items from combat (e.g. do NOT add "enemy's sword" unless the player explicitly picks it up).
- Do NOT add items that are merely mentioned or seen — only items the player character actually acquires or loses.

Return ONLY a JSON object (no explanation, no markdown):
{"gained":[{"name":"...","description":"...","quantity":1}],"lost":["item name","item name"]}

If nothing changed: {"gained":[],"lost":[]}

Story passage:
%s`, gmText)

	raw, err := completer.Generate(ctx, prompt)
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

	changed := 0

	for _, g := range result.Gained {
		if g.Name == "" {
			continue
		}
		qty := g.Quantity
		if qty <= 0 {
			qty = 1
		}
		if _, err := s.db.CreateItem(charID, g.Name, g.Description, qty); err == nil {
			changed++
		}
	}

	if len(result.Lost) > 0 {
		existing, err := s.db.ListItems(charID)
		if err == nil {
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
	}

	if changed > 0 {
		s.bus.Publish(Event{Type: EventItemUpdated, Payload: map[string]any{"character_id": charID}})
	}
}
