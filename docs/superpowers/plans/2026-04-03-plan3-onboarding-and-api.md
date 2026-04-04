# Plan 3: Onboarding + HTTP Read API

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make a fresh install immediately playable — seed the 5 supported rulesets, add MCP tools to create campaigns and characters, and expose HTTP read endpoints the React UI needs.

**Architecture:** A second migration seeds the rulesets table so every new DB is ready to use. Five new MCP tools cover the campaign/character lifecycle (create + list). Six HTTP GET handlers give the frontend read access to all game state. All new behaviour is tested first.

**Tech Stack:** Go stdlib `net/http` (Go 1.22+ method routing), `modernc.org/sqlite`, `github.com/mark3labs/mcp-go`, `github.com/stretchr/testify`.

---

## File Map

| Action   | Path |
|----------|------|
| Create   | `internal/db/migrations/002_seed_rulesets.sql` |
| Modify   | `internal/db/db_test.go` (+1 test) |
| Modify   | `internal/api/events.go` (+2 event constants) |
| Create   | `internal/mcp/lifecycle.go` |
| Create   | `internal/mcp/lifecycle_test.go` |
| Modify   | `internal/mcp/server.go` (+5 tool registrations in `registerTools`) |
| Create   | `internal/api/routes.go` |
| Create   | `internal/api/routes_test.go` |
| Modify   | `internal/api/server.go` (+6 route registrations in `registerRoutes`) |

---

## Task 1: Seed rulesets via migration

**Files:**
- Modify: `internal/db/db_test.go`
- Create: `internal/db/migrations/002_seed_rulesets.sql`

- [ ] **Step 1: Write the failing test**

  Append to `internal/db/db_test.go`:

  ```go
  func TestRulesets_SeededByMigration(t *testing.T) {
  	d := newTestDB(t)
  	list, err := d.ListRulesets()
  	require.NoError(t, err)
  	names := make([]string, len(list))
  	for i, r := range list {
  		names[i] = r.Name
  	}
  	assert.ElementsMatch(t, []string{"dnd5e", "ironsworn", "vtm", "coc", "cyberpunk"}, names)
  }
  ```

- [ ] **Step 2: Run test to verify it fails**

  ```bash
  go test ./internal/db/ -run TestRulesets_SeededByMigration -v
  ```

  Expected: `FAIL — got empty slice, want 5 names`

- [ ] **Step 3: Write the migration**

  Create `internal/db/migrations/002_seed_rulesets.sql`:

  ```sql
  INSERT OR IGNORE INTO rulesets (name, schema_json, version) VALUES
    ('dnd5e',     '{"system":"dnd5e","fields":["race","class","level","hp","ac","str","dex","con","int","wis","cha","proficiency_bonus","skills","inventory","spells","features"]}', '5e'),
    ('ironsworn', '{"system":"ironsworn","fields":["edge","heart","iron","shadow","wits","health","spirit","supply","momentum","vows","bonds","assets","notes"]}', '1.0'),
    ('vtm',       '{"system":"vtm","fields":["clan","generation","humanity","blood_pool","willpower","attributes","abilities","disciplines","virtues","backgrounds","notes"]}', 'V20'),
    ('coc',       '{"system":"coc","fields":["occupation","age","hp","sanity","luck","mp","str","con","siz","dex","app","int","pow","edu","skills","inventory","notes"]}', '7e'),
    ('cyberpunk', '{"system":"cyberpunk","fields":["role","int","ref","cool","tech","lk","att","ma","emp","body","humanity","eurodollars","skills","cyberware","gear","notes"]}', 'Red');
  ```

- [ ] **Step 4: Run test to verify it passes**

  ```bash
  go test ./internal/db/ -count=1 -v
  ```

  Expected: all tests PASS including `TestRulesets_SeededByMigration` and `TestOpen_IdempotentMigrations`

- [ ] **Step 5: Commit**

  ```bash
  git add internal/db/migrations/002_seed_rulesets.sql internal/db/db_test.go
  git commit -m "feat: seed 5 rulesets via migration 002"
  ```

---

## Task 2: Add event types for campaign and character creation

**Files:**
- Modify: `internal/api/events.go`

- [ ] **Step 1: Add two event constants**

  In `internal/api/events.go`, append inside the `const` block after `EventSessionEnded`:

  ```go
  EventCampaignCreated  EventType = "campaign_created"
  EventCharacterCreated EventType = "character_created"
  ```

- [ ] **Step 2: Verify it compiles**

  ```bash
  go build ./...
  ```

  Expected: no output, exit 0

- [ ] **Step 3: Commit**

  ```bash
  git add internal/api/events.go
  git commit -m "feat: add EventCampaignCreated and EventCharacterCreated"
  ```

---

## Task 3: MCP lifecycle tools (create_campaign, list_campaigns, create_character, list_characters, list_sessions)

**Files:**
- Create: `internal/mcp/lifecycle_test.go`
- Create: `internal/mcp/lifecycle.go`
- Modify: `internal/mcp/server.go`

- [ ] **Step 1: Write the failing tests**

  Create `internal/mcp/lifecycle_test.go`:

  ```go
  package mcp

  import (
  	"context"
  	"encoding/json"
  	"strconv"
  	"testing"

  	"github.com/digitalghost404/inkandbone/internal/db"
  	mcplib "github.com/mark3labs/mcp-go/mcp"
  	"github.com/stretchr/testify/assert"
  	"github.com/stretchr/testify/require"
  )

  func TestCreateCampaign(t *testing.T) {
  	s := newTestMCP(t)
  	// "dnd5e" is seeded by migration 002
  	req := mcplib.CallToolRequest{}
  	req.Params.Arguments = map[string]any{
  		"ruleset": "dnd5e",
  		"name":    "Dragon Campaign",
  	}
  	result, err := s.handleCreateCampaign(context.Background(), req)
  	require.NoError(t, err)
  	require.False(t, result.IsError)

  	got, _ := s.db.GetSetting("active_campaign_id")
  	assert.NotEmpty(t, got)
  }

  func TestCreateCampaign_unknownRuleset(t *testing.T) {
  	s := newTestMCP(t)
  	req := mcplib.CallToolRequest{}
  	req.Params.Arguments = map[string]any{"ruleset": "pathfinder", "name": "Test"}
  	result, err := s.handleCreateCampaign(context.Background(), req)
  	require.NoError(t, err)
  	assert.True(t, result.IsError)
  }

  func TestListCampaigns_empty(t *testing.T) {
  	s := newTestMCP(t)
  	result, err := s.handleListCampaigns(context.Background(), mcplib.CallToolRequest{})
  	require.NoError(t, err)
  	require.False(t, result.IsError)

  	tc, ok := result.Content[0].(mcplib.TextContent)
  	require.True(t, ok)
  	var campaigns []db.Campaign
  	require.NoError(t, json.Unmarshal([]byte(tc.Text), &campaigns))
  	assert.Empty(t, campaigns)
  }

  func TestCreateCharacter(t *testing.T) {
  	s := newTestMCP(t)
  	campID, _, _ := setupCampaign(t, s)
  	require.NoError(t, s.db.SetSetting("active_campaign_id", strconv.FormatInt(campID, 10)))

  	req := mcplib.CallToolRequest{}
  	req.Params.Arguments = map[string]any{"name": "Talia"}
  	result, err := s.handleCreateCharacter(context.Background(), req)
  	require.NoError(t, err)
  	require.False(t, result.IsError)

  	got, _ := s.db.GetSetting("active_character_id")
  	assert.NotEmpty(t, got)
  }

  func TestCreateCharacter_noCampaign(t *testing.T) {
  	s := newTestMCP(t)
  	req := mcplib.CallToolRequest{}
  	req.Params.Arguments = map[string]any{"name": "Talia"}
  	result, err := s.handleCreateCharacter(context.Background(), req)
  	require.NoError(t, err)
  	assert.True(t, result.IsError)
  }

  func TestListCharacters(t *testing.T) {
  	s := newTestMCP(t)
  	campID, charID, _ := setupCampaign(t, s)
  	require.NoError(t, s.db.SetSetting("active_campaign_id", strconv.FormatInt(campID, 10)))

  	result, err := s.handleListCharacters(context.Background(), mcplib.CallToolRequest{})
  	require.NoError(t, err)
  	require.False(t, result.IsError)

  	tc, ok := result.Content[0].(mcplib.TextContent)
  	require.True(t, ok)
  	var chars []db.Character
  	require.NoError(t, json.Unmarshal([]byte(tc.Text), &chars))
  	require.Len(t, chars, 1)
  	assert.Equal(t, charID, chars[0].ID)
  }

  func TestListSessions(t *testing.T) {
  	s := newTestMCP(t)
  	campID, _, sessID := setupCampaign(t, s)
  	require.NoError(t, s.db.SetSetting("active_campaign_id", strconv.FormatInt(campID, 10)))

  	result, err := s.handleListSessions(context.Background(), mcplib.CallToolRequest{})
  	require.NoError(t, err)
  	require.False(t, result.IsError)

  	tc, ok := result.Content[0].(mcplib.TextContent)
  	require.True(t, ok)
  	var sessions []db.Session
  	require.NoError(t, json.Unmarshal([]byte(tc.Text), &sessions))
  	require.Len(t, sessions, 1)
  	assert.Equal(t, sessID, sessions[0].ID)
  }
  ```

- [ ] **Step 2: Run tests to verify they fail (compile error)**

  ```bash
  go test ./internal/mcp/ -count=1 2>&1 | head -20
  ```

  Expected: `undefined: s.handleCreateCampaign` (compile error — correct)

- [ ] **Step 3: Implement lifecycle.go**

  Create `internal/mcp/lifecycle.go`:

  ```go
  package mcp

  import (
  	"context"
  	"encoding/json"
  	"fmt"
  	"strconv"

  	"github.com/digitalghost404/inkandbone/internal/api"
  	"github.com/digitalghost404/inkandbone/internal/db"
  	mcplib "github.com/mark3labs/mcp-go/mcp"
  )

  func (s *Server) handleCreateCampaign(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
  	rulesetName, ok := reqStr(req, "ruleset")
  	if !ok {
  		return mcplib.NewToolResultError("ruleset is required"), nil
  	}
  	name, ok := reqStr(req, "name")
  	if !ok {
  		return mcplib.NewToolResultError("name is required"), nil
  	}
  	description := optStr(req, "description")

  	rs, err := s.db.GetRulesetByName(rulesetName)
  	if err != nil {
  		return mcplib.NewToolResultError("db error: " + err.Error()), nil
  	}
  	if rs == nil {
  		return mcplib.NewToolResultError(fmt.Sprintf("unknown ruleset %q — valid: dnd5e, ironsworn, vtm, coc, cyberpunk", rulesetName)), nil
  	}

  	campID, err := s.db.CreateCampaign(rs.ID, name, description)
  	if err != nil {
  		return mcplib.NewToolResultError("create campaign: " + err.Error()), nil
  	}
  	if err := s.db.SetSetting("active_campaign_id", strconv.FormatInt(campID, 10)); err != nil {
  		return mcplib.NewToolResultError("set active campaign: " + err.Error()), nil
  	}

  	s.bus.Publish(api.Event{Type: api.EventCampaignCreated, Payload: map[string]any{"campaign_id": campID, "name": name}})
  	return mcplib.NewToolResultText(fmt.Sprintf("campaign %d created and activated: %s (%s)", campID, name, rulesetName)), nil
  }

  func (s *Server) handleListCampaigns(_ context.Context, _ mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
  	campaigns, err := s.db.ListCampaigns()
  	if err != nil {
  		return mcplib.NewToolResultError("db error: " + err.Error()), nil
  	}
  	if campaigns == nil {
  		campaigns = []db.Campaign{}
  	}
  	b, err := json.Marshal(campaigns)
  	if err != nil {
  		return mcplib.NewToolResultError("marshal error: " + err.Error()), nil
  	}
  	return mcplib.NewToolResultText(string(b)), nil
  }

  func (s *Server) handleCreateCharacter(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
  	name, ok := reqStr(req, "name")
  	if !ok {
  		return mcplib.NewToolResultError("name is required"), nil
  	}

  	var campID int64
  	if id, ok := optInt64(req, "campaign_id"); ok && id > 0 {
  		campID = id
  	} else {
  		var err error
  		campID, err = s.activeCampaignID()
  		if err != nil {
  			return mcplib.NewToolResultError(err.Error()), nil
  		}
  	}

  	charID, err := s.db.CreateCharacter(campID, name)
  	if err != nil {
  		return mcplib.NewToolResultError("create character: " + err.Error()), nil
  	}
  	if err := s.db.SetSetting("active_character_id", strconv.FormatInt(charID, 10)); err != nil {
  		return mcplib.NewToolResultError("set active character: " + err.Error()), nil
  	}

  	s.bus.Publish(api.Event{Type: api.EventCharacterCreated, Payload: map[string]any{"character_id": charID, "name": name}})
  	return mcplib.NewToolResultText(fmt.Sprintf("character %d created and activated: %s", charID, name)), nil
  }

  func (s *Server) handleListCharacters(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
  	var campID int64
  	if id, ok := optInt64(req, "campaign_id"); ok && id > 0 {
  		campID = id
  	} else {
  		var err error
  		campID, err = s.activeCampaignID()
  		if err != nil {
  			return mcplib.NewToolResultError(err.Error()), nil
  		}
  	}

  	characters, err := s.db.ListCharacters(campID)
  	if err != nil {
  		return mcplib.NewToolResultError("db error: " + err.Error()), nil
  	}
  	if characters == nil {
  		characters = []db.Character{}
  	}
  	b, err := json.Marshal(characters)
  	if err != nil {
  		return mcplib.NewToolResultError("marshal error: " + err.Error()), nil
  	}
  	return mcplib.NewToolResultText(string(b)), nil
  }

  func (s *Server) handleListSessions(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
  	var campID int64
  	if id, ok := optInt64(req, "campaign_id"); ok && id > 0 {
  		campID = id
  	} else {
  		var err error
  		campID, err = s.activeCampaignID()
  		if err != nil {
  			return mcplib.NewToolResultError(err.Error()), nil
  		}
  	}

  	sessions, err := s.db.ListSessions(campID)
  	if err != nil {
  		return mcplib.NewToolResultError("db error: " + err.Error()), nil
  	}
  	if sessions == nil {
  		sessions = []db.Session{}
  	}
  	b, err := json.Marshal(sessions)
  	if err != nil {
  		return mcplib.NewToolResultError("marshal error: " + err.Error()), nil
  	}
  	return mcplib.NewToolResultText(string(b)), nil
  }
  ```

- [ ] **Step 4: Register the 5 new tools in server.go**

  In `internal/mcp/server.go`, inside `registerTools()` before the closing `}`, add:

  ```go
  	// Lifecycle — campaign & character creation
  	s.srv.AddTool(mcplib.NewTool("create_campaign",
  		mcplib.WithDescription("Create a new campaign under a ruleset and make it active. Call once at the very start of a new game."),
  		mcplib.WithString("ruleset", mcplib.Required(), mcplib.Description("Ruleset name: dnd5e, ironsworn, vtm, coc, or cyberpunk")),
  		mcplib.WithString("name", mcplib.Required(), mcplib.Description("Campaign name")),
  		mcplib.WithString("description", mcplib.Description("Optional campaign description")),
  	), s.handleCreateCampaign)

  	s.srv.AddTool(mcplib.NewTool("list_campaigns",
  		mcplib.WithDescription("List all campaigns."),
  	), s.handleListCampaigns)

  	s.srv.AddTool(mcplib.NewTool("create_character",
  		mcplib.WithDescription("Create a new player character in the active campaign and make them active."),
  		mcplib.WithString("name", mcplib.Required(), mcplib.Description("Character name")),
  		mcplib.WithNumber("campaign_id", mcplib.Description("Campaign ID (defaults to active campaign)")),
  	), s.handleCreateCharacter)

  	s.srv.AddTool(mcplib.NewTool("list_characters",
  		mcplib.WithDescription("List all characters in the active (or specified) campaign."),
  		mcplib.WithNumber("campaign_id", mcplib.Description("Campaign ID (defaults to active campaign)")),
  	), s.handleListCharacters)

  	s.srv.AddTool(mcplib.NewTool("list_sessions",
  		mcplib.WithDescription("List all sessions for the active (or specified) campaign, newest first."),
  		mcplib.WithNumber("campaign_id", mcplib.Description("Campaign ID (defaults to active campaign)")),
  	), s.handleListSessions)
  ```

- [ ] **Step 5: Run tests to verify they pass**

  ```bash
  go test ./internal/mcp/ -count=1 -v 2>&1 | tail -20
  ```

  Expected: all tests PASS (was 23, now 30)

- [ ] **Step 6: Commit**

  ```bash
  git add internal/mcp/lifecycle.go internal/mcp/lifecycle_test.go internal/mcp/server.go
  git commit -m "feat: add lifecycle MCP tools (create/list campaign, character, session)"
  ```

---

## Task 4: HTTP read API (6 endpoints)

**Files:**
- Create: `internal/api/routes_test.go`
- Create: `internal/api/routes.go`
- Modify: `internal/api/server.go`

- [ ] **Step 1: Write the failing tests**

  Create `internal/api/routes_test.go`:

  ```go
  package api

  import (
  	"encoding/json"
  	"net/http"
  	"net/http/httptest"
  	"strconv"
  	"testing"

  	"github.com/digitalghost404/inkandbone/internal/db"
  	"github.com/stretchr/testify/assert"
  	"github.com/stretchr/testify/require"
  )

  // seedCampaign creates a ruleset, campaign, and session for use in route tests.
  func seedCampaign(t *testing.T, d *db.DB) (campID, sessID int64) {
  	t.Helper()
  	rsID, err := d.CreateRuleset("dnd5e", `{}`, "5e")
  	require.NoError(t, err)
  	campID, err = d.CreateCampaign(rsID, "Test Campaign", "")
  	require.NoError(t, err)
  	sessID, err = d.CreateSession(campID, "S1", "2026-04-03")
  	require.NoError(t, err)
  	return
  }

  func TestListCampaigns_empty(t *testing.T) {
  	s := newTestServer(t)
  	req := httptest.NewRequest(http.MethodGet, "/api/campaigns", nil)
  	w := httptest.NewRecorder()
  	s.ServeHTTP(w, req)
  	assert.Equal(t, http.StatusOK, w.Code)
  	var campaigns []db.Campaign
  	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &campaigns))
  	assert.Empty(t, campaigns)
  }

  func TestListCampaigns_withData(t *testing.T) {
  	s := newTestServer(t)
  	campID, _ := seedCampaign(t, s.db)
  	req := httptest.NewRequest(http.MethodGet, "/api/campaigns", nil)
  	w := httptest.NewRecorder()
  	s.ServeHTTP(w, req)
  	assert.Equal(t, http.StatusOK, w.Code)
  	var campaigns []db.Campaign
  	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &campaigns))
  	require.Len(t, campaigns, 1)
  	assert.Equal(t, campID, campaigns[0].ID)
  }

  func TestListCharacters_empty(t *testing.T) {
  	s := newTestServer(t)
  	campID, _ := seedCampaign(t, s.db)
  	req := httptest.NewRequest(http.MethodGet, "/api/campaigns/"+strconv.FormatInt(campID, 10)+"/characters", nil)
  	w := httptest.NewRecorder()
  	s.ServeHTTP(w, req)
  	assert.Equal(t, http.StatusOK, w.Code)
  	var chars []db.Character
  	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &chars))
  	assert.Empty(t, chars)
  }

  func TestListCharacters_withData(t *testing.T) {
  	s := newTestServer(t)
  	campID, _ := seedCampaign(t, s.db)
  	charID, err := s.db.CreateCharacter(campID, "Kael")
  	require.NoError(t, err)
  	req := httptest.NewRequest(http.MethodGet, "/api/campaigns/"+strconv.FormatInt(campID, 10)+"/characters", nil)
  	w := httptest.NewRecorder()
  	s.ServeHTTP(w, req)
  	assert.Equal(t, http.StatusOK, w.Code)
  	var chars []db.Character
  	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &chars))
  	require.Len(t, chars, 1)
  	assert.Equal(t, charID, chars[0].ID)
  }

  func TestListSessions_withData(t *testing.T) {
  	s := newTestServer(t)
  	campID, sessID := seedCampaign(t, s.db)
  	req := httptest.NewRequest(http.MethodGet, "/api/campaigns/"+strconv.FormatInt(campID, 10)+"/sessions", nil)
  	w := httptest.NewRecorder()
  	s.ServeHTTP(w, req)
  	assert.Equal(t, http.StatusOK, w.Code)
  	var sessions []db.Session
  	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &sessions))
  	require.Len(t, sessions, 1)
  	assert.Equal(t, sessID, sessions[0].ID)
  }

  func TestListMessages_empty(t *testing.T) {
  	s := newTestServer(t)
  	_, sessID := seedCampaign(t, s.db)
  	req := httptest.NewRequest(http.MethodGet, "/api/sessions/"+strconv.FormatInt(sessID, 10)+"/messages", nil)
  	w := httptest.NewRecorder()
  	s.ServeHTTP(w, req)
  	assert.Equal(t, http.StatusOK, w.Code)
  	var msgs []db.Message
  	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &msgs))
  	assert.Empty(t, msgs)
  }

  func TestListDiceRolls_empty(t *testing.T) {
  	s := newTestServer(t)
  	_, sessID := seedCampaign(t, s.db)
  	req := httptest.NewRequest(http.MethodGet, "/api/sessions/"+strconv.FormatInt(sessID, 10)+"/dice-rolls", nil)
  	w := httptest.NewRecorder()
  	s.ServeHTTP(w, req)
  	assert.Equal(t, http.StatusOK, w.Code)
  	var rolls []db.DiceRoll
  	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &rolls))
  	assert.Empty(t, rolls)
  }

  func TestListMapPins_empty(t *testing.T) {
  	s := newTestServer(t)
  	campID, _ := seedCampaign(t, s.db)
  	mapID, err := s.db.CreateMap(campID, "World Map", "")
  	require.NoError(t, err)
  	req := httptest.NewRequest(http.MethodGet, "/api/maps/"+strconv.FormatInt(mapID, 10)+"/pins", nil)
  	w := httptest.NewRecorder()
  	s.ServeHTTP(w, req)
  	assert.Equal(t, http.StatusOK, w.Code)
  	var pins []db.MapPin
  	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &pins))
  	assert.Empty(t, pins)
  }
  ```

- [ ] **Step 2: Run tests to verify they fail**

  ```bash
  go test ./internal/api/ -run "TestList" -v 2>&1 | head -20
  ```

  Expected: 404 responses (routes not registered yet)

- [ ] **Step 3: Implement routes.go**

  Create `internal/api/routes.go`:

  ```go
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
  ```

- [ ] **Step 4: Register routes in server.go**

  Replace the `registerRoutes` method body in `internal/api/server.go`:

  ```go
  func (s *Server) registerRoutes() {
  	s.mux.HandleFunc("/ws", s.hub.ServeWS)
  	s.mux.HandleFunc("/api/health", s.handleHealth)
  	s.mux.HandleFunc("GET /api/campaigns", s.handleListCampaigns)
  	s.mux.HandleFunc("GET /api/campaigns/{id}/characters", s.handleListCharacters)
  	s.mux.HandleFunc("GET /api/campaigns/{id}/sessions", s.handleListSessions)
  	s.mux.HandleFunc("GET /api/sessions/{id}/messages", s.handleListMessages)
  	s.mux.HandleFunc("GET /api/sessions/{id}/dice-rolls", s.handleListDiceRolls)
  	s.mux.HandleFunc("GET /api/maps/{id}/pins", s.handleListMapPins)
  }
  ```

- [ ] **Step 5: Run all tests to verify they pass**

  ```bash
  go test ./internal/... -count=1 -v 2>&1 | tail -30
  ```

  Expected: all tests PASS across all 4 packages

- [ ] **Step 6: Commit**

  ```bash
  git add internal/api/routes.go internal/api/routes_test.go internal/api/server.go
  git commit -m "feat: HTTP read API — campaigns, characters, sessions, messages, dice rolls, map pins"
  ```

---

## Self-Review

**Spec coverage check:**
- Ruleset seeding → Task 1 ✓
- Campaign lifecycle MCP tools → Task 3 ✓
- Character lifecycle MCP tools → Task 3 ✓
- Session listing MCP tool → Task 3 ✓
- HTTP read endpoints for the UI → Task 4 ✓
- Event types for new tools → Task 2 ✓

**No placeholders:** All tasks contain complete code — no TBD or "similar to above."

**Type consistency:**
- `db.Campaign`, `db.Character`, `db.Session`, `db.Message`, `db.DiceRoll`, `db.MapPin`, `db.WorldNote` — all defined in `internal/db/queries_*.go` and used consistently.
- `s.activeCampaignID()` — defined in `internal/mcp/world.go:14`, used in lifecycle.go ✓
- `s.db.CreateMap()` — defined in `internal/db/queries_world.go:88`, used in routes_test.go ✓
- `mcplib.TextContent` — the type-assert pattern matches `context_test.go:19` ✓
