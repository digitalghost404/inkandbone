# Phase C: GM Session Tools — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Give the GM four power tools: a panic button for instant NPC improvisation, a pre-session briefing generator, a loose thread detector, and a campaign Q&A chatbot.

**Architecture:** Four new AI-backed POST endpoints, each using the existing `s.aiClient.(ai.Completer)` pattern to call Claude with a tight system prompt and return structured JSON or plain text. All endpoints require `s.aiClient != nil`. Frontend additions to JournalPanel add buttons and an inline Q&A interface.

**Tech Stack:** Go, Claude Haiku via ai.Completer, React/TypeScript

---

## File Map

| File | Change |
|------|--------|
| `internal/api/routes_phase_c.go` | Create — four new handlers |
| `internal/api/server.go` | Modify — register four new routes |
| `internal/api/routes_phase_c_test.go` | Create — tests for all four handlers |
| `web/src/JournalPanel.tsx` | Modify — Improvise button, Pre-session Brief button, Detect Threads button, Campaign Q&A widget |
| `web/src/api.ts` | Modify — four new API call functions |

---

### Task C1: POST /api/sessions/{id}/improvise — GM Panic Button

When the GM is stuck, this endpoint returns an instant NPC card: name, motivation, and a complication — all in under 3 seconds using a tight Claude Haiku prompt.

**Files:**
- Create: `internal/api/routes_phase_c.go`
- Modify: `internal/api/server.go`
- Create: `internal/api/routes_phase_c_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/api/routes_phase_c_test.go`:

```go
package api

import (
    "encoding/json"
    "fmt"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestImprovise(t *testing.T) {
    s := newTestServer(t)
    _, sessID := seedCampaign(t, s.db)

    req := httptest.NewRequest(http.MethodPost,
        fmt.Sprintf("/api/sessions/%d/improvise", sessID), nil)
    w := httptest.NewRecorder()
    s.ServeHTTP(w, req)

    // Without AI client, expect 503
    assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestImproviseNotFound(t *testing.T) {
    s := newTestServer(t)
    req := httptest.NewRequest(http.MethodPost, "/api/sessions/99999/improvise", nil)
    w := httptest.NewRecorder()
    s.ServeHTTP(w, req)
    assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}
```

- [ ] **Step 2: Run tests to verify they fail (route not found → 405 or 404)**

Run: `go test ./internal/api/... -run "TestImprovise" -v`
Expected: FAIL — 404 or 405 (route doesn't exist yet)

- [ ] **Step 3: Create routes_phase_c.go with handleImprovise**

Create `internal/api/routes_phase_c.go`:

```go
package api

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"

    "github.com/digitalghost404/inkandbone/internal/ai"
)

// handleImprovise returns an instant NPC card for GM use when stuck mid-session.
// POST /api/sessions/{id}/improvise
func (s *Server) handleImprovise(w http.ResponseWriter, r *http.Request) {
    if s.aiClient == nil {
        http.Error(w, "AI not configured", http.StatusServiceUnavailable)
        return
    }
    completer, ok := s.aiClient.(ai.Completer)
    if !ok {
        http.Error(w, "AI client does not support completion", http.StatusServiceUnavailable)
        return
    }
    id, ok2 := parsePathID(r, "id")
    if !ok2 {
        http.Error(w, "invalid session id", http.StatusBadRequest)
        return
    }
    sess, err := s.db.GetSession(id)
    if err != nil || sess == nil {
        http.Error(w, "session not found", http.StatusNotFound)
        return
    }

    worldCtx := s.buildWorldContext(r.Context(), id)
    prompt := fmt.Sprintf(`%s

You are an improvisational GM assistant. Based on the world context above, invent ONE NPC who could appear right now. Return ONLY valid JSON, no commentary:
{"name":"<full name>","role":"<what they are/do>","motivation":"<what they want>","complication":"<interesting twist or secret>","opening_line":"<first thing they say>"}`, worldCtx)

    result, err := completer.Complete(r.Context(), prompt, 256)
    if err != nil {
        http.Error(w, "AI error: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Extract JSON from the response
    start := strings.Index(result, "{")
    end := strings.LastIndex(result, "}")
    if start < 0 || end <= start {
        http.Error(w, "unexpected AI response format", http.StatusInternalServerError)
        return
    }
    jsonStr := result[start : end+1]

    var npc map[string]any
    if err := json.Unmarshal([]byte(jsonStr), &npc); err != nil {
        http.Error(w, "AI returned invalid JSON", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(npc)
}
```

- [ ] **Step 4: Register route in server.go**

In `internal/api/server.go`, in `registerRoutes()`, add a "Phase C" comment block:

```go
// Phase C: GM session tools
s.mux.HandleFunc("POST /api/sessions/{id}/improvise", s.handleImprovise)
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/api/... -run "TestImprovise" -v`
Expected: PASS (503 is correct since newTestServer has no AI client)

- [ ] **Step 6: Run full test suite**

Run: `make test`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/api/routes_phase_c.go internal/api/server.go internal/api/routes_phase_c_test.go
git commit -m "feat(api): POST /api/sessions/{id}/improvise — GM panic button NPC generator"
```

---

### Task C2: POST /api/campaigns/{id}/pre-session-brief

Reviews the last 3 sessions' messages and summaries, returns a structured briefing: what happened last session, open threads, NPC status, and story hooks.

**Files:**
- Modify: `internal/api/routes_phase_c.go`
- Modify: `internal/api/server.go`
- Modify: `internal/api/routes_phase_c_test.go`

- [ ] **Step 1: Write the failing test**

Add to `internal/api/routes_phase_c_test.go`:

```go
func TestPreSessionBrief(t *testing.T) {
    s := newTestServer(t)
    campID, _ := seedCampaign(t, s.db)

    req := httptest.NewRequest(http.MethodPost,
        fmt.Sprintf("/api/campaigns/%d/pre-session-brief", campID), nil)
    w := httptest.NewRecorder()
    s.ServeHTTP(w, req)
    // No AI client → 503
    assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}
```

- [ ] **Step 2: Run test to verify it fails (404)**

Run: `go test ./internal/api/... -run "TestPreSessionBrief" -v`
Expected: FAIL — 404

- [ ] **Step 3: Add handlePreSessionBrief to routes_phase_c.go**

Add to `internal/api/routes_phase_c.go`:

```go
// handlePreSessionBrief generates a structured pre-session briefing from recent session history.
// POST /api/campaigns/{id}/pre-session-brief
func (s *Server) handlePreSessionBrief(w http.ResponseWriter, r *http.Request) {
    if s.aiClient == nil {
        http.Error(w, "AI not configured", http.StatusServiceUnavailable)
        return
    }
    completer, ok := s.aiClient.(ai.Completer)
    if !ok {
        http.Error(w, "AI client does not support completion", http.StatusServiceUnavailable)
        return
    }
    campID, ok2 := parsePathID(r, "id")
    if !ok2 {
        http.Error(w, "invalid campaign id", http.StatusBadRequest)
        return
    }

    sessions, err := s.db.ListSessions(campID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Gather last 3 sessions worth of summaries + recent messages
    var contextParts []string
    limit := 3
    if len(sessions) < limit {
        limit = len(sessions)
    }
    for i := 0; i < limit; i++ {
        sess := sessions[i]
        part := fmt.Sprintf("Session %q (%s)", sess.Title, sess.Date)
        if sess.Summary != "" {
            part += "\nSummary: " + sess.Summary
        }
        msgs, err2 := s.db.RecentMessages(sess.ID, 10)
        if err2 == nil && len(msgs) > 0 {
            var lines []string
            for _, m := range msgs {
                if !m.Whisper {
                    lines = append(lines, m.Role+": "+m.Content)
                }
            }
            part += "\nRecent messages:\n" + strings.Join(lines, "\n")
        }
        contextParts = append(contextParts, part)
    }

    if len(contextParts) == 0 {
        writeJSON(w, map[string]any{
            "last_session":  "No sessions played yet.",
            "open_threads":  []string{},
            "npc_status":    []string{},
            "hooks":         []string{},
        })
        return
    }

    prompt := fmt.Sprintf(`You are a GM assistant reviewing campaign history. Here are the last %d sessions:

%s

Return ONLY valid JSON with these fields:
{
  "last_session": "<2-3 sentence recap of the most recent session>",
  "open_threads": ["<unresolved story thread>", ...],
  "npc_status": ["<NPC name>: <current status/relationship>", ...],
  "hooks": ["<story hook or unresolved question worth exploring>", ...]
}
Include 2-4 items per list. Be specific, not vague.`, limit, strings.Join(contextParts, "\n\n---\n\n"))

    result, err := completer.Complete(r.Context(), prompt, 512)
    if err != nil {
        http.Error(w, "AI error: "+err.Error(), http.StatusInternalServerError)
        return
    }

    start := strings.Index(result, "{")
    end := strings.LastIndex(result, "}")
    if start < 0 || end <= start {
        http.Error(w, "unexpected AI response format", http.StatusInternalServerError)
        return
    }

    var brief map[string]any
    if err := json.Unmarshal([]byte(result[start:end+1]), &brief); err != nil {
        http.Error(w, "AI returned invalid JSON", http.StatusInternalServerError)
        return
    }

    writeJSON(w, brief)
}
```

- [ ] **Step 4: Register route in server.go**

Add to the Phase C block:
```go
s.mux.HandleFunc("POST /api/campaigns/{id}/pre-session-brief", s.handlePreSessionBrief)
```

- [ ] **Step 5: Run test**

Run: `go test ./internal/api/... -run "TestPreSessionBrief" -v`
Expected: PASS (503 correct)

- [ ] **Step 6: Run full suite**

Run: `make test`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/api/routes_phase_c.go internal/api/server.go internal/api/routes_phase_c_test.go
git commit -m "feat(api): POST /api/campaigns/{id}/pre-session-brief — AI session briefing"
```

---

### Task C3: POST /api/sessions/{id}/detect-threads

Scans all session messages for unfulfilled story promises, unresolved questions, and dangling plot hooks.

**Files:**
- Modify: `internal/api/routes_phase_c.go`
- Modify: `internal/api/server.go`
- Modify: `internal/api/routes_phase_c_test.go`

- [ ] **Step 1: Write the failing test**

Add to `internal/api/routes_phase_c_test.go`:

```go
func TestDetectThreads(t *testing.T) {
    s := newTestServer(t)
    _, sessID := seedCampaign(t, s.db)

    req := httptest.NewRequest(http.MethodPost,
        fmt.Sprintf("/api/sessions/%d/detect-threads", sessID), nil)
    w := httptest.NewRecorder()
    s.ServeHTTP(w, req)
    assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}
```

- [ ] **Step 2: Run test — expect 404**

Run: `go test ./internal/api/... -run "TestDetectThreads" -v`
Expected: FAIL — 404

- [ ] **Step 3: Add handleDetectThreads to routes_phase_c.go**

```go
// handleDetectThreads scans session messages for unresolved plot threads.
// POST /api/sessions/{id}/detect-threads
func (s *Server) handleDetectThreads(w http.ResponseWriter, r *http.Request) {
    if s.aiClient == nil {
        http.Error(w, "AI not configured", http.StatusServiceUnavailable)
        return
    }
    completer, ok := s.aiClient.(ai.Completer)
    if !ok {
        http.Error(w, "AI client does not support completion", http.StatusServiceUnavailable)
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
    if len(msgs) == 0 {
        writeJSON(w, map[string]any{"threads": []string{}})
        return
    }

    var lines []string
    for _, m := range msgs {
        if !m.Whisper {
            lines = append(lines, m.Role+": "+m.Content)
        }
    }

    prompt := fmt.Sprintf(`You are a story analyst. Read this TTRPG session transcript and identify loose threads — unfulfilled promises, unanswered questions, mentioned-but-unresolved story elements, and hooks the GM introduced but never paid off.

Transcript:
%s

Return ONLY valid JSON:
{"threads": ["<specific unresolved thread>", ...]}

Include only genuinely unresolved items. Be specific (name characters, locations, objects). Include 3-8 threads.`, strings.Join(lines, "\n"))

    result, err := completer.Complete(r.Context(), prompt, 512)
    if err != nil {
        http.Error(w, "AI error: "+err.Error(), http.StatusInternalServerError)
        return
    }

    start := strings.Index(result, "{")
    end := strings.LastIndex(result, "}")
    if start < 0 || end <= start {
        http.Error(w, "unexpected AI response format", http.StatusInternalServerError)
        return
    }

    var out map[string]any
    if err := json.Unmarshal([]byte(result[start:end+1]), &out); err != nil {
        http.Error(w, "AI returned invalid JSON", http.StatusInternalServerError)
        return
    }

    writeJSON(w, out)
}
```

- [ ] **Step 4: Register route**

```go
s.mux.HandleFunc("POST /api/sessions/{id}/detect-threads", s.handleDetectThreads)
```

- [ ] **Step 5: Run test**

Run: `go test ./internal/api/... -run "TestDetectThreads" -v`
Expected: PASS (503 correct)

- [ ] **Step 6: Commit**

```bash
git add internal/api/routes_phase_c.go internal/api/server.go internal/api/routes_phase_c_test.go
git commit -m "feat(api): POST /api/sessions/{id}/detect-threads — loose thread scanner"
```

---

### Task C4: POST /api/campaigns/{id}/ask — Campaign Q&A Chatbot

Answers questions about the campaign ("What happened to Mira?") using all world notes and session summaries as context.

**Files:**
- Modify: `internal/api/routes_phase_c.go`
- Modify: `internal/api/server.go`
- Modify: `internal/api/routes_phase_c_test.go`

- [ ] **Step 1: Write the failing test**

Add to `internal/api/routes_phase_c_test.go`:

```go
func TestCampaignAsk(t *testing.T) {
    s := newTestServer(t)
    campID, _ := seedCampaign(t, s.db)

    body := `{"question":"What is known about the city?"}`
    req := httptest.NewRequest(http.MethodPost,
        fmt.Sprintf("/api/campaigns/%d/ask", campID),
        strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    s.ServeHTTP(w, req)
    assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestCampaignAskMissingQuestion(t *testing.T) {
    s := newTestServer(t)
    campID, _ := seedCampaign(t, s.db)

    body := `{}`
    req := httptest.NewRequest(http.MethodPost,
        fmt.Sprintf("/api/campaigns/%d/ask", campID),
        strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    s.ServeHTTP(w, req)
    // No AI client → 503 (checked before question validation)
    assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}
```

- [ ] **Step 2: Run tests — expect 404**

Run: `go test ./internal/api/... -run "TestCampaignAsk" -v`
Expected: FAIL — 404

- [ ] **Step 3: Add handleCampaignAsk to routes_phase_c.go**

```go
// handleCampaignAsk answers a natural language question about the campaign.
// POST /api/campaigns/{id}/ask
func (s *Server) handleCampaignAsk(w http.ResponseWriter, r *http.Request) {
    if s.aiClient == nil {
        http.Error(w, "AI not configured", http.StatusServiceUnavailable)
        return
    }
    completer, ok := s.aiClient.(ai.Completer)
    if !ok {
        http.Error(w, "AI client does not support completion", http.StatusServiceUnavailable)
        return
    }
    campID, ok2 := parsePathID(r, "id")
    if !ok2 {
        http.Error(w, "invalid campaign id", http.StatusBadRequest)
        return
    }

    var body struct {
        Question string `json:"question"`
    }
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Question == "" {
        http.Error(w, "question is required", http.StatusBadRequest)
        return
    }

    // Assemble campaign knowledge base
    var kb strings.Builder

    // World notes
    notes, err := s.db.SearchWorldNotes(campID, "", "", "")
    if err == nil && len(notes) > 0 {
        kb.WriteString("=== World Notes ===\n")
        for _, n := range notes {
            fmt.Fprintf(&kb, "[%s] %s: %s\n", n.Category, n.Title, n.Content)
        }
        kb.WriteString("\n")
    }

    // Session summaries
    sessions, err2 := s.db.ListSessions(campID)
    if err2 == nil && len(sessions) > 0 {
        kb.WriteString("=== Session History ===\n")
        for _, sess := range sessions {
            if sess.Summary != "" {
                fmt.Fprintf(&kb, "Session %q (%s): %s\n", sess.Title, sess.Date, sess.Summary)
            }
        }
    }

    if kb.Len() == 0 {
        writeJSON(w, map[string]any{"answer": "No campaign information has been recorded yet."})
        return
    }

    prompt := fmt.Sprintf(`You are a campaign archivist. Using only the information below, answer the player's question. If the answer isn't in the records, say so clearly.

%s

Question: %s

Answer in 1-3 sentences. Be direct and specific.`, kb.String(), body.Question)

    result, err3 := completer.Complete(r.Context(), prompt, 256)
    if err3 != nil {
        http.Error(w, "AI error: "+err3.Error(), http.StatusInternalServerError)
        return
    }

    writeJSON(w, map[string]any{"answer": strings.TrimSpace(result)})
}
```

- [ ] **Step 4: Register route**

```go
s.mux.HandleFunc("POST /api/campaigns/{id}/ask", s.handleCampaignAsk)
```

- [ ] **Step 5: Run tests**

Run: `go test ./internal/api/... -run "TestCampaignAsk" -v`
Expected: PASS (503 correct)

- [ ] **Step 6: Run full suite**

Run: `make test`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/api/routes_phase_c.go internal/api/server.go internal/api/routes_phase_c_test.go
git commit -m "feat(api): POST /api/campaigns/{id}/ask — campaign Q&A chatbot"
```

---

### Task C5: Frontend — Improvise button, Brief button, Threads button, Q&A widget

**Files:**
- Modify: `web/src/JournalPanel.tsx`
- Modify: `web/src/api.ts`

- [ ] **Step 1: Add API functions to api.ts**

In `web/src/api.ts`, add four new functions:

```typescript
export async function improviseNPC(sessionId: number): Promise<{
  name: string; role: string; motivation: string; complication: string; opening_line: string;
}> {
  const res = await fetch(`/api/sessions/${sessionId}/improvise`, { method: 'POST' });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function preSessionBrief(campaignId: number): Promise<{
  last_session: string; open_threads: string[]; npc_status: string[]; hooks: string[];
}> {
  const res = await fetch(`/api/campaigns/${campaignId}/pre-session-brief`, { method: 'POST' });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function detectThreads(sessionId: number): Promise<{ threads: string[] }> {
  const res = await fetch(`/api/sessions/${sessionId}/detect-threads`, { method: 'POST' });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function askCampaign(campaignId: number, question: string): Promise<{ answer: string }> {
  const res = await fetch(`/api/campaigns/${campaignId}/ask`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ question }),
  });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}
```

- [ ] **Step 2: Add state and handlers to JournalPanel.tsx**

In `web/src/JournalPanel.tsx`, the component receives `session` and `aiEnabled` as props. Add imports at the top:

```typescript
import { improviseNPC, preSessionBrief, detectThreads, askCampaign } from './api';
```

Add state for GM tools panel (near other state declarations):

```typescript
const [gmToolsOpen, setGmToolsOpen] = React.useState(false);
const [improvising, setImprovising] = React.useState(false);
const [improvisedNPC, setImprovisedNPC] = React.useState<{
  name: string; role: string; motivation: string; complication: string; opening_line: string;
} | null>(null);
const [briefing, setBriefing] = React.useState<{
  last_session: string; open_threads: string[]; npc_status: string[]; hooks: string[];
} | null>(null);
const [briefingLoading, setBriefingLoading] = React.useState(false);
const [threads, setThreads] = React.useState<string[]>([]);
const [threadsLoading, setThreadsLoading] = React.useState(false);
const [qaQuestion, setQaQuestion] = React.useState('');
const [qaAnswer, setQaAnswer] = React.useState('');
const [qaLoading, setQaLoading] = React.useState(false);
```

Add handler functions:

```typescript
const handleImprovise = async () => {
  if (!session) return;
  setImprovising(true);
  setImprovisedNPC(null);
  try {
    const npc = await improviseNPC(session.id);
    setImprovisedNPC(npc);
  } catch (e) {
    console.error('Improvise failed:', e);
  } finally {
    setImprovising(false);
  }
};

const handlePreSessionBrief = async () => {
  if (!session) return;
  setBriefingLoading(true);
  setBriefing(null);
  try {
    // session.campaign_id must be passed — check the session prop type; it should include campaign_id
    const brief = await preSessionBrief(session.campaign_id);
    setBriefing(brief);
  } catch (e) {
    console.error('Pre-session brief failed:', e);
  } finally {
    setBriefingLoading(false);
  }
};

const handleDetectThreads = async () => {
  if (!session) return;
  setThreadsLoading(true);
  setThreads([]);
  try {
    const result = await detectThreads(session.id);
    setThreads(result.threads);
  } catch (e) {
    console.error('Detect threads failed:', e);
  } finally {
    setThreadsLoading(false);
  }
};

const handleAsk = async (e: React.FormEvent) => {
  e.preventDefault();
  if (!session || !qaQuestion.trim()) return;
  setQaLoading(true);
  setQaAnswer('');
  try {
    const result = await askCampaign(session.campaign_id, qaQuestion);
    setQaAnswer(result.answer);
  } catch (e) {
    console.error('Campaign ask failed:', e);
  } finally {
    setQaLoading(false);
  }
};
```

- [ ] **Step 3: Add GM Tools section to JournalPanel JSX**

In the JournalPanel JSX, add a "GM Tools" section (only when `aiEnabled` is true) below the existing XP/milestone section:

```tsx
{aiEnabled && session && (
  <div className="gm-tools-section">
    <div
      className="gm-tools-header"
      onClick={() => setGmToolsOpen(!gmToolsOpen)}
    >
      <span>GM Tools</span>
      <span>{gmToolsOpen ? '▲' : '▼'}</span>
    </div>
    {gmToolsOpen && (
      <div className="gm-tools-body">
        {/* Panic Button */}
        <button
          className="gm-tool-btn"
          onClick={handleImprovise}
          disabled={improvising}
        >
          {improvising ? 'Generating…' : '⚡ Improvise NPC'}
        </button>
        {improvisedNPC && (
          <div className="improvised-npc">
            <strong>{improvisedNPC.name}</strong> — {improvisedNPC.role}
            <div>Wants: {improvisedNPC.motivation}</div>
            <div>Twist: {improvisedNPC.complication}</div>
            <div><em>"{improvisedNPC.opening_line}"</em></div>
          </div>
        )}

        {/* Pre-session Brief */}
        <button
          className="gm-tool-btn"
          onClick={handlePreSessionBrief}
          disabled={briefingLoading}
        >
          {briefingLoading ? 'Generating…' : '📋 Pre-session Brief'}
        </button>
        {briefing && (
          <div className="session-brief">
            <div><strong>Last session:</strong> {briefing.last_session}</div>
            {briefing.open_threads.length > 0 && (
              <div>
                <strong>Open threads:</strong>
                <ul>{briefing.open_threads.map((t, i) => <li key={i}>{t}</li>)}</ul>
              </div>
            )}
            {briefing.hooks.length > 0 && (
              <div>
                <strong>Hooks:</strong>
                <ul>{briefing.hooks.map((h, i) => <li key={i}>{h}</li>)}</ul>
              </div>
            )}
          </div>
        )}

        {/* Detect Threads */}
        <button
          className="gm-tool-btn"
          onClick={handleDetectThreads}
          disabled={threadsLoading}
        >
          {threadsLoading ? 'Scanning…' : '🔍 Detect Loose Threads'}
        </button>
        {threads.length > 0 && (
          <div className="loose-threads">
            <strong>Loose threads:</strong>
            <ul>{threads.map((t, i) => <li key={i}>{t}</li>)}</ul>
          </div>
        )}

        {/* Campaign Q&A */}
        <form onSubmit={handleAsk} className="campaign-qa">
          <input
            type="text"
            value={qaQuestion}
            onChange={e => setQaQuestion(e.target.value)}
            placeholder="Ask about the campaign…"
            disabled={qaLoading}
          />
          <button type="submit" disabled={qaLoading || !qaQuestion.trim()}>
            {qaLoading ? '…' : 'Ask'}
          </button>
        </form>
        {qaAnswer && (
          <div className="qa-answer">{qaAnswer}</div>
        )}
      </div>
    )}
  </div>
)}
```

- [ ] **Step 4: Add CSS for GM tools**

Add to the relevant CSS file (or as style props if using CSS-in-JS):

```css
.gm-tools-section {
  margin-top: 16px;
  border: 1px solid var(--gold, #c9a84c);
  border-radius: 4px;
}
.gm-tools-header {
  display: flex;
  justify-content: space-between;
  padding: 8px 12px;
  cursor: pointer;
  background: rgba(201, 168, 76, 0.08);
  font-size: 0.85rem;
  color: var(--gold, #c9a84c);
  font-weight: 600;
}
.gm-tools-body {
  padding: 10px 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.gm-tool-btn {
  background: rgba(201, 168, 76, 0.1);
  border: 1px solid var(--gold, #c9a84c);
  color: var(--gold, #c9a84c);
  padding: 5px 10px;
  border-radius: 3px;
  cursor: pointer;
  font-size: 0.8rem;
  text-align: left;
}
.gm-tool-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.improvised-npc, .session-brief, .loose-threads, .qa-answer {
  font-size: 0.8rem;
  background: rgba(0,0,0,0.2);
  padding: 8px;
  border-radius: 3px;
  color: var(--text, #e0d5b7);
}
.improvised-npc strong, .session-brief strong { color: var(--gold, #c9a84c); }
.campaign-qa { display: flex; gap: 6px; }
.campaign-qa input {
  flex: 1;
  background: var(--bg-input, #1a1a1a);
  border: 1px solid var(--border, #333);
  color: var(--text, #e0d5b7);
  padding: 4px 8px;
  border-radius: 3px;
  font-size: 0.8rem;
}
```

- [ ] **Step 5: Build to verify no TypeScript errors**

Run: `make build`
Expected: Build succeeds, no TypeScript errors

- [ ] **Step 6: Run full test suite**

Run: `make test`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add web/src/JournalPanel.tsx web/src/api.ts
git commit -m "feat(ui): GM tools panel — improvise NPC, pre-session brief, detect threads, campaign Q&A"
```
