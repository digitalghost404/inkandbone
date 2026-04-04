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

Do not break character. Do not summarize. Just continue the story.`

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

	worldCtx := s.buildWorldContext(r.Context(), id)
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
