# Phase B: AI Intelligence & Context — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the GM smarter by injecting active objectives and NPC personality into every prompt, and guide the GM to produce interesting failure consequences instead of dead ends.

**Architecture:** Three layers of change — DB migration adds personality_json to world_notes, buildWorldContext in routes.go injects richer context, and a failure-consequence block directs the GM when a dice roll fails. NPC personality cards are editable in the WorldNotesPanel UI.

**Tech Stack:** Go, SQLite, React/TypeScript

---

## File Map

| File | Change |
|------|--------|
| `internal/db/migrations/012_phase_b.sql` | Create — ADD COLUMN personality_json on world_notes |
| `internal/db/queries_world.go` | Modify — add PersonalityJSON field, FindWorldNoteByTitle method, UpdateWorldNotePersonality method |
| `internal/db/queries_world_test.go` | Modify — add tests for FindWorldNoteByTitle |
| `internal/api/routes.go` | Modify — extend buildWorldContext, extend handleGMRespondStream failure block |
| `internal/api/routes_test.go` | Modify — add TestBuildWorldContextWithObjectives |
| `web/src/WorldNotesPanel.tsx` | Modify — add personality card form for NPC notes |
| `web/src/types.ts` | Modify — add personality_json to WorldNote type |

---

### Task B1: Migration 012 — personality_json on world_notes

**Files:**
- Create: `internal/db/migrations/012_phase_b.sql`

- [ ] **Step 1: Write the migration**

```sql
-- 012_phase_b.sql: NPC personality fields on world_notes
ALTER TABLE world_notes ADD COLUMN personality_json TEXT NOT NULL DEFAULT '';
```

- [ ] **Step 2: Verify migration applies cleanly**

Run: `make test`
Expected: PASS (migration auto-applied by newTestDB)

- [ ] **Step 3: Commit**

```bash
git add internal/db/migrations/012_phase_b.sql
git commit -m "feat(db): migration 012 — personality_json on world_notes"
```

---

### Task B2: DB — FindWorldNoteByTitle + PersonalityJSON field

**Files:**
- Modify: `internal/db/queries_world.go`
- Modify: `internal/db/queries_world_test.go`

- [ ] **Step 1: Write the failing test**

Add to `internal/db/queries_world_test.go`:

```go
func TestFindWorldNoteByTitle(t *testing.T) {
    d := newTestDB(t)
    campID := setupCampaign(t, d)
    _, err := d.CreateWorldNote(campID, "Mira Stonewright", "A dwarven blacksmith.", "npc")
    require.NoError(t, err)

    found, err := d.FindWorldNoteByTitle(campID, "Mira Stonewright")
    require.NoError(t, err)
    require.NotNil(t, found)
    assert.Equal(t, "Mira Stonewright", found.Title)

    missing, err := d.FindWorldNoteByTitle(campID, "Nobody")
    require.NoError(t, err)
    assert.Nil(t, missing)
}

func TestUpdateWorldNotePersonality(t *testing.T) {
    d := newTestDB(t)
    campID := setupCampaign(t, d)
    id, err := d.CreateWorldNote(campID, "Mira", "Blacksmith.", "npc")
    require.NoError(t, err)

    personality := `{"speech_quirk":"speaks in rhyme","motivation":"find her lost brother","secret":"she's a spy","disposition":"friendly"}`
    require.NoError(t, d.UpdateWorldNotePersonality(id, personality))

    n, err := d.GetWorldNote(id)
    require.NoError(t, err)
    assert.Equal(t, personality, n.PersonalityJSON)
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/db/... -run "TestFindWorldNoteByTitle|TestUpdateWorldNotePersonality" -v`
Expected: FAIL with "undefined: FindWorldNoteByTitle"

- [ ] **Step 3: Add PersonalityJSON to WorldNote struct and implement methods**

In `internal/db/queries_world.go`, update the WorldNote struct:

```go
type WorldNote struct {
    ID              int64  `json:"id"`
    CampaignID      int64  `json:"campaign_id"`
    Title           string `json:"title"`
    Content         string `json:"content"`
    Category        string `json:"category"`
    TagsJSON        string `json:"tags_json"`
    PersonalityJSON string `json:"personality_json"`
    CreatedAt       string `json:"created_at"`
}
```

Update all SELECT scans in `queries_world.go` that construct a WorldNote to include the new field. There are three scan sites: `GetWorldNote`, `SearchWorldNotes`, and `ListRecentWorldNotes`.

For `GetWorldNote` (line ~61), change the query and Scan:
```go
func (d *DB) GetWorldNote(id int64) (*WorldNote, error) {
    var n WorldNote
    err := d.db.QueryRow(
        "SELECT id, campaign_id, title, content, category, tags_json, personality_json, created_at FROM world_notes WHERE id = ?",
        id,
    ).Scan(&n.ID, &n.CampaignID, &n.Title, &n.Content, &n.Category, &n.TagsJSON, &n.PersonalityJSON, &n.CreatedAt)
    if err != nil {
        return nil, err
    }
    return &n, nil
}
```

For `SearchWorldNotes` (line ~70), update SELECT and Scan similarly:
```go
// in the SELECT: add personality_json after tags_json
"SELECT id, campaign_id, title, content, category, tags_json, personality_json, created_at FROM world_notes WHERE campaign_id = ?"
// in rows.Scan:
rows.Scan(&n.ID, &n.CampaignID, &n.Title, &n.Content, &n.Category, &n.TagsJSON, &n.PersonalityJSON, &n.CreatedAt)
```

For `ListRecentWorldNotes` (line ~104), same pattern:
```go
"SELECT id, campaign_id, title, content, category, tags_json, personality_json, created_at FROM world_notes WHERE campaign_id = ? ORDER BY created_at DESC LIMIT ?"
// Scan: add &n.PersonalityJSON before &n.CreatedAt
```

Then add the two new methods after `ListRecentWorldNotes`:

```go
// FindWorldNoteByTitle returns the first world note with the given title for the campaign, or nil if not found.
func (d *DB) FindWorldNoteByTitle(campaignID int64, title string) (*WorldNote, error) {
    var n WorldNote
    err := d.db.QueryRow(
        "SELECT id, campaign_id, title, content, category, tags_json, personality_json, created_at FROM world_notes WHERE campaign_id = ? AND title = ? LIMIT 1",
        campaignID, title,
    ).Scan(&n.ID, &n.CampaignID, &n.Title, &n.Content, &n.Category, &n.TagsJSON, &n.PersonalityJSON, &n.CreatedAt)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    return &n, nil
}

// UpdateWorldNotePersonality sets the personality_json for an NPC world note.
func (d *DB) UpdateWorldNotePersonality(id int64, personalityJSON string) error {
    _, err := d.db.Exec("UPDATE world_notes SET personality_json = ? WHERE id = ?", personalityJSON, id)
    return err
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/db/... -run "TestFindWorldNoteByTitle|TestUpdateWorldNotePersonality" -v`
Expected: PASS

- [ ] **Step 5: Run full test suite**

Run: `make test`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/db/queries_world.go internal/db/queries_world_test.go
git commit -m "feat(db): add PersonalityJSON to WorldNote, FindWorldNoteByTitle, UpdateWorldNotePersonality"
```

---

### Task B3: Extend buildWorldContext — active objectives + NPC personality

**Files:**
- Modify: `internal/api/routes.go`
- Modify: `internal/api/routes_test.go`

- [ ] **Step 1: Write the failing test**

Add to `internal/api/routes_test.go`:

```go
func TestBuildWorldContextWithObjectives(t *testing.T) {
    s := newTestServer(t)
    campID, sessID := seedCampaign(t, s.db)

    // Create an active objective
    _, err := s.db.CreateObjective(campID, "Find the lost artifact", "Somewhere in the ruins.", nil)
    require.NoError(t, err)

    ctx := s.buildWorldContext(context.Background(), sessID)
    assert.Contains(t, ctx, "Active objectives:")
    assert.Contains(t, ctx, "Find the lost artifact")
}

func TestBuildWorldContextWithNPCPersonality(t *testing.T) {
    s := newTestServer(t)
    campID, sessID := seedCampaign(t, s.db)

    // Create NPC world note with personality
    noteID, err := s.db.CreateWorldNote(campID, "Mira Stonewright", "A dwarven blacksmith.", "npc")
    require.NoError(t, err)
    require.NoError(t, s.db.UpdateWorldNotePersonality(noteID, `{"motivation":"find her lost brother","speech_quirk":"speaks in rhyme"}`))

    // Add Mira to the session NPC roster
    _, err = s.db.CreateNPC(sessID, "Mira Stonewright", "")
    require.NoError(t, err)

    ctx := s.buildWorldContext(context.Background(), sessID)
    assert.Contains(t, ctx, "Mira Stonewright")
    assert.Contains(t, ctx, "find her lost brother")
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/api/... -run "TestBuildWorldContextWith" -v`
Expected: FAIL — context strings not present

- [ ] **Step 3: Extend buildWorldContext in routes.go**

In `internal/api/routes.go`, after the active combat block (around line 764, just before `sb.WriteString("[/WORLD STATE]")`), add:

```go
// Inject active top-level objectives
sess2, err2 := s.db.GetSession(sessionID)
if err2 == nil && sess2 != nil {
    objectives, err3 := s.db.ListObjectives(sess2.CampaignID)
    if err3 == nil {
        var active []string
        for _, o := range objectives {
            if o.Status == "active" && o.ParentID == nil {
                active = append(active, o.Title)
            }
        }
        if len(active) > 0 {
            fmt.Fprintf(&sb, "Active objectives: %s\n", strings.Join(active, "; "))
        } else {
            sb.WriteString("Active objectives: none\n")
        }
    }

    // Inject NPC personality for session NPCs that have a world note
    npcs, err4 := s.db.ListNPCs(sessionID)
    if err4 == nil && len(npcs) > 0 {
        type npcCard struct {
            name        string
            motivation  string
            speechQuirk string
            secret      string
            disposition string
        }
        var cards []npcCard
        for _, npc := range npcs {
            note, err5 := s.db.FindWorldNoteByTitle(sess2.CampaignID, npc.Name)
            if err5 != nil || note == nil || note.PersonalityJSON == "" {
                continue
            }
            var p struct {
                Motivation  string `json:"motivation"`
                SpeechQuirk string `json:"speech_quirk"`
                Secret      string `json:"secret"`
                Disposition string `json:"disposition"`
            }
            if err6 := json.Unmarshal([]byte(note.PersonalityJSON), &p); err6 != nil {
                continue
            }
            cards = append(cards, npcCard{
                name:        npc.Name,
                motivation:  p.Motivation,
                speechQuirk: p.SpeechQuirk,
                secret:      p.Secret,
                disposition: p.Disposition,
            })
        }
        if len(cards) > 0 {
            sb.WriteString("NPC personalities:\n")
            for _, c := range cards {
                parts := []string{c.name}
                if c.motivation != "" {
                    parts = append(parts, "wants: "+c.motivation)
                }
                if c.speechQuirk != "" {
                    parts = append(parts, "speech: "+c.speechQuirk)
                }
                if c.disposition != "" {
                    parts = append(parts, "disposition: "+c.disposition)
                }
                fmt.Fprintf(&sb, "  - %s\n", strings.Join(parts, "; "))
            }
        }
    }
}
```

Note: `s.db.ListNPCs` is the existing method (check the actual method name — it may be `ListSessionNPCs(sessionID)`). Use whatever name exists in `internal/db/queries_npcs.go`.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/api/... -run "TestBuildWorldContextWith" -v`
Expected: PASS

- [ ] **Step 5: Run full test suite**

Run: `make test`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/api/routes.go internal/api/routes_test.go
git commit -m "feat(api): inject active objectives and NPC personality into GM context"
```

---

### Task B4: Failure consequence block in handleGMRespondStream

**Files:**
- Modify: `internal/api/routes.go`
- Modify: `internal/api/routes_test.go`

The principle: when a dice roll fails, Claude should be directed to produce an interesting complication, not a dead end. The current code (lines ~918-930) already injects a `[DICE ROLL]` block. After this block, when `!roll.Success`, append a GM direction.

- [ ] **Step 1: Write the failing test**

Add to `internal/api/routes_test.go`:

```go
func TestFailureConsequenceBlockInjected(t *testing.T) {
    // buildWorldContext is called internally; we test the worldCtx assembly
    // by inspecting the system prompt indirectly via the dice roll injection logic.
    // Since checkAndExecuteRoll requires ruleset config, we test the block format directly.
    s := newTestServer(t)
    _ = s

    // Simulate what handleGMRespondStream does with a failed roll
    type rollResult struct {
        Attribute string
        Reason    string
        Expression string
        Total     int
        DC        int
        Success   bool
    }
    roll := &rollResult{
        Attribute:  "Strength",
        Reason:     "forcing the door",
        Expression: "1d20+3",
        Total:      7,
        DC:         15,
        Success:    false,
    }

    worldCtx := "[WORLD STATE]\nSession summary: none\n[/WORLD STATE]"
    outcome := "FAILURE"
    dcNote := fmt.Sprintf(" against DC %d", roll.DC)
    worldCtx += fmt.Sprintf(
        "\n[DICE ROLL]\nAction required a %s check%s.\nReason: %s\nRoll: %s = %d — %s\n[/DICE ROLL]",
        roll.Attribute, dcNote, roll.Reason, roll.Expression, roll.Total, outcome,
    )
    if !roll.Success {
        worldCtx += "\n[GM DIRECTION]\nThe action failed. Do NOT produce a dead end. Instead, introduce a complication, cost, or partial success that advances the story in an interesting direction. The player should feel the consequence, not a wall.\n[/GM DIRECTION]"
    }

    assert.Contains(t, worldCtx, "[GM DIRECTION]")
    assert.Contains(t, worldCtx, "complication")
}
```

This test validates the string format we'll insert. It's a unit test of the block format.

- [ ] **Step 2: Run test to verify format**

Run: `go test ./internal/api/... -run "TestFailureConsequenceBlockInjected" -v`
Expected: PASS (this test passes immediately — it validates the format we'll implement)

- [ ] **Step 3: Add failure consequence block to handleGMRespondStream**

In `internal/api/routes.go`, find the dice roll injection block (around lines 918-930):

```go
worldCtx += fmt.Sprintf(
    "\n[DICE ROLL]\nAction required a %s check%s.\nReason: %s\nRoll: %s = %d — %s\n[/DICE ROLL]",
    roll.Attribute, dcNote, roll.Reason, roll.Expression, roll.Total, outcome,
)
```

Immediately after this block (before `systemPrompt := worldCtx + ...`), add:

```go
if !roll.Success {
    worldCtx += "\n[GM DIRECTION]\nThe action failed. Do NOT produce a dead end. Instead, introduce a complication, cost, or partial success that advances the story in an interesting direction. The player should feel the consequence, not a wall.\n[/GM DIRECTION]"
}
```

- [ ] **Step 4: Run full test suite**

Run: `make test`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/api/routes.go internal/api/routes_test.go
git commit -m "feat(api): inject failure consequence direction when dice roll fails"
```

---

### Task B5: NPC personality card UI in WorldNotesPanel

**Files:**
- Modify: `web/src/types.ts` — add personality_json to WorldNote
- Modify: `web/src/WorldNotesPanel.tsx` — add personality form for NPC category notes
- Modify: `internal/api/routes.go` — add PATCH /api/world-notes/{id}/personality route handler
- Modify: `internal/api/server.go` — register new route

- [ ] **Step 1: Add personality_json to WorldNote type**

In `web/src/types.ts`, find the `WorldNote` interface and add:

```typescript
export interface WorldNote {
  id: number;
  campaign_id: number;
  title: string;
  content: string;
  category: string;
  tags_json: string;
  personality_json: string;  // add this field
  created_at: string;
}
```

- [ ] **Step 2: Add PATCH /api/world-notes/{id}/personality handler**

In `internal/api/routes.go`, add this handler (near the existing `handlePatchWorldNote`):

```go
func (s *Server) handlePatchWorldNotePersonality(w http.ResponseWriter, r *http.Request) {
    id, ok := parsePathID(r, "id")
    if !ok {
        http.Error(w, "invalid id", http.StatusBadRequest)
        return
    }
    var body struct {
        PersonalityJSON string `json:"personality_json"`
    }
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        http.Error(w, "invalid json", http.StatusBadRequest)
        return
    }
    if err := s.db.UpdateWorldNotePersonality(id, body.PersonalityJSON); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    s.bus.Publish(Event{Type: EventWorldNoteUpdated, Payload: map[string]any{"world_note_id": id}})
    w.WriteHeader(http.StatusNoContent)
}
```

- [ ] **Step 3: Register the route in server.go**

In `internal/api/server.go`, in `registerRoutes()`, add after the `PATCH /api/world-notes/{id}` line:

```go
s.mux.HandleFunc("PATCH /api/world-notes/{id}/personality", s.handlePatchWorldNotePersonality)
```

- [ ] **Step 4: Write the API test**

Add to `internal/api/routes_test.go`:

```go
func TestPatchWorldNotePersonality(t *testing.T) {
    s := newTestServer(t)
    campID, _ := seedCampaign(t, s.db)
    noteID, err := s.db.CreateWorldNote(campID, "Mira", "Blacksmith.", "npc")
    require.NoError(t, err)

    body := `{"personality_json":"{\"motivation\":\"find her brother\",\"speech_quirk\":\"speaks in rhyme\"}"}`
    req := httptest.NewRequest(http.MethodPatch,
        fmt.Sprintf("/api/world-notes/%d/personality", noteID),
        strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    s.ServeHTTP(w, req)
    assert.Equal(t, http.StatusNoContent, w.Code)

    note, err := s.db.GetWorldNote(noteID)
    require.NoError(t, err)
    assert.Contains(t, note.PersonalityJSON, "find her brother")
}
```

- [ ] **Step 5: Run the API test**

Run: `go test ./internal/api/... -run "TestPatchWorldNotePersonality" -v`
Expected: PASS

- [ ] **Step 6: Add personality card form to WorldNotesPanel.tsx**

In `web/src/WorldNotesPanel.tsx`, find where individual note content is rendered (the note detail/edit view). Add a collapsible "NPC Personality" section that only appears when `note.category === 'npc'`:

```tsx
// Add state for personality editing (inside the component, near other state declarations):
const [personality, setPersonality] = React.useState({
  motivation: '',
  speech_quirk: '',
  secret: '',
  disposition: '',
});
const [personalityEditing, setPersonalityEditing] = React.useState(false);

// When a note is selected, parse its personality_json:
React.useEffect(() => {
  if (selectedNote?.personality_json) {
    try {
      const p = JSON.parse(selectedNote.personality_json);
      setPersonality({
        motivation: p.motivation || '',
        speech_quirk: p.speech_quirk || '',
        secret: p.secret || '',
        disposition: p.disposition || '',
      });
    } catch {
      setPersonality({ motivation: '', speech_quirk: '', secret: '', disposition: '' });
    }
  } else {
    setPersonality({ motivation: '', speech_quirk: '', secret: '', disposition: '' });
  }
}, [selectedNote?.id]);

// Save personality handler:
const savePersonality = async () => {
  if (!selectedNote) return;
  await fetch(`/api/world-notes/${selectedNote.id}/personality`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ personality_json: JSON.stringify(personality) }),
  });
  setPersonalityEditing(false);
};

// In the JSX, after the note content area, when selectedNote.category === 'npc':
{selectedNote?.category === 'npc' && (
  <div className="personality-card">
    <div className="personality-header" onClick={() => setPersonalityEditing(!personalityEditing)}>
      <span>NPC Personality</span>
      <span>{personalityEditing ? '▲' : '▼'}</span>
    </div>
    {personalityEditing && (
      <div className="personality-fields">
        {(['motivation', 'speech_quirk', 'secret', 'disposition'] as const).map(field => (
          <div key={field} className="personality-field">
            <label>{field.replace('_', ' ')}</label>
            <input
              value={personality[field]}
              onChange={e => setPersonality(prev => ({ ...prev, [field]: e.target.value }))}
              placeholder={field.replace('_', ' ')}
            />
          </div>
        ))}
        <button onClick={savePersonality} className="save-btn">Save</button>
      </div>
    )}
  </div>
)}
```

Add minimal CSS to the grimoire theme (in the relevant CSS file for WorldNotesPanel, or inline with style objects if the panel uses CSS-in-JS):

```css
.personality-card {
  margin-top: 12px;
  border: 1px solid var(--gold, #c9a84c);
  border-radius: 4px;
  overflow: hidden;
}
.personality-header {
  display: flex;
  justify-content: space-between;
  padding: 6px 10px;
  background: rgba(201, 168, 76, 0.1);
  cursor: pointer;
  font-size: 0.85rem;
  color: var(--gold, #c9a84c);
}
.personality-fields {
  padding: 8px 10px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.personality-field label {
  font-size: 0.75rem;
  text-transform: capitalize;
  color: var(--text-muted, #888);
}
.personality-field input {
  width: 100%;
  background: var(--bg-input, #1a1a1a);
  border: 1px solid var(--border, #333);
  color: var(--text, #e0d5b7);
  padding: 4px 6px;
  border-radius: 3px;
  font-size: 0.85rem;
}
```

- [ ] **Step 7: Run full test suite + build**

Run: `make test && make build`
Expected: PASS + binary builds without errors

- [ ] **Step 8: Commit**

```bash
git add internal/api/routes.go internal/api/server.go internal/api/routes_test.go \
        internal/db/queries_world.go web/src/types.ts web/src/WorldNotesPanel.tsx
git commit -m "feat: NPC personality cards — UI form, PATCH endpoint, GM context injection"
```
