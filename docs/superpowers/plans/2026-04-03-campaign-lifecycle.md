# Campaign Lifecycle Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `close_campaign`, `delete_campaign`, and auto-reopen in `set_active` to ink & bone's MCP server.

**Architecture:** DB methods live in `queries_core.go` following existing patterns. MCP handlers for close/delete go in `lifecycle.go` alongside create_campaign. `set_active` in `campaign.go` gains a reopen side-effect. Three new event constants in `events.go`. Tools registered in `server.go`.

**Tech Stack:** Go, SQLite (modernc.org/sqlite), mark3labs/mcp-go, testify

---

## File Map

| File | Change |
|---|---|
| `internal/db/queries_core.go` | Add `CloseCampaign`, `ReopenCampaign`, `GetCampaignStats`, `DeleteCampaign` |
| `internal/db/queries_core_test.go` | Add tests for all four new DB methods |
| `internal/api/events.go` | Add `EventCampaignClosed`, `EventCampaignDeleted`, `EventCampaignReopened` |
| `internal/mcp/lifecycle.go` | Add `handleCloseCampaign`, `handleDeleteCampaign` |
| `internal/mcp/lifecycle_test.go` | Add tests for both new handlers |
| `internal/mcp/campaign.go` | Update `handleSetActive` to reopen closed campaigns |
| `internal/mcp/campaign_test.go` | Add `TestSetActive_reopensClosed` |
| `internal/mcp/server.go` | Register `close_campaign` and `delete_campaign` tools |

---

## Task 1: DB methods — CloseCampaign, ReopenCampaign, GetCampaignStats, DeleteCampaign

**Files:**
- Modify: `internal/db/queries_core_test.go`
- Modify: `internal/db/queries_core.go`

- [ ] **Step 1: Write failing tests**

Append to `internal/db/queries_core_test.go`:

```go
func TestCloseCampaign(t *testing.T) {
	d := newTestDB(t)
	rs, _ := d.GetRulesetByName("dnd5e")
	id, err := d.CreateCampaign(rs.ID, "Test", "")
	require.NoError(t, err)

	require.NoError(t, d.CloseCampaign(id))

	c, err := d.GetCampaign(id)
	require.NoError(t, err)
	assert.False(t, c.Active)
}

func TestReopenCampaign(t *testing.T) {
	d := newTestDB(t)
	rs, _ := d.GetRulesetByName("dnd5e")
	id, err := d.CreateCampaign(rs.ID, "Test", "")
	require.NoError(t, err)
	require.NoError(t, d.CloseCampaign(id))

	require.NoError(t, d.ReopenCampaign(id))

	c, err := d.GetCampaign(id)
	require.NoError(t, err)
	assert.True(t, c.Active)
}

func TestGetCampaignStats(t *testing.T) {
	d := newTestDB(t)
	rs, _ := d.GetRulesetByName("dnd5e")
	campID, err := d.CreateCampaign(rs.ID, "Test", "")
	require.NoError(t, err)

	_, err = d.CreateCharacter(campID, "Hero")
	require.NoError(t, err)
	_, err = d.CreateSession(campID, "S1", "2026-04-01")
	require.NoError(t, err)

	stats, err := d.GetCampaignStats(campID)
	require.NoError(t, err)
	assert.Equal(t, 1, stats.Sessions)
	assert.Equal(t, 1, stats.Characters)
	assert.Equal(t, 0, stats.WorldNotes)
	assert.Equal(t, 0, stats.Maps)
}

func TestDeleteCampaign(t *testing.T) {
	d := newTestDB(t)
	rs, _ := d.GetRulesetByName("dnd5e")
	campID, err := d.CreateCampaign(rs.ID, "Test", "")
	require.NoError(t, err)

	charID, err := d.CreateCharacter(campID, "Hero")
	require.NoError(t, err)
	sessID, err := d.CreateSession(campID, "S1", "2026-04-01")
	require.NoError(t, err)
	_, err = d.CreateMessage(sessID, "user", "hello")
	require.NoError(t, err)

	require.NoError(t, d.DeleteCampaign(campID))

	c, err := d.GetCampaign(campID)
	require.NoError(t, err)
	assert.Nil(t, c)

	ch, err := d.GetCharacter(charID)
	require.NoError(t, err)
	assert.Nil(t, ch)

	sess, err := d.GetSession(sessID)
	require.NoError(t, err)
	assert.Nil(t, sess)
}
```

- [ ] **Step 2: Run tests to verify they fail**

```
cd /home/digitalghost/projects/inkandbone
go test ./internal/db/... -run "TestCloseCampaign|TestReopenCampaign|TestGetCampaignStats|TestDeleteCampaign" -v
```

Expected: FAIL — `d.CloseCampaign undefined` (or similar)

- [ ] **Step 3: Implement the four DB methods**

Append to `internal/db/queries_core.go`:

```go
func (d *DB) CloseCampaign(id int64) error {
	res, err := d.db.Exec("UPDATE campaigns SET active = 0 WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("campaign %d not found", id)
	}
	return nil
}

func (d *DB) ReopenCampaign(id int64) error {
	res, err := d.db.Exec("UPDATE campaigns SET active = 1 WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("campaign %d not found", id)
	}
	return nil
}

// CampaignStats holds row counts for the confirmation message in delete_campaign.
type CampaignStats struct {
	Sessions   int
	Characters int
	WorldNotes int
	Maps       int
}

func (d *DB) GetCampaignStats(id int64) (CampaignStats, error) {
	var s CampaignStats
	if err := d.db.QueryRow("SELECT COUNT(*) FROM sessions WHERE campaign_id = ?", id).Scan(&s.Sessions); err != nil {
		return s, err
	}
	if err := d.db.QueryRow("SELECT COUNT(*) FROM characters WHERE campaign_id = ?", id).Scan(&s.Characters); err != nil {
		return s, err
	}
	if err := d.db.QueryRow("SELECT COUNT(*) FROM world_notes WHERE campaign_id = ?", id).Scan(&s.WorldNotes); err != nil {
		return s, err
	}
	if err := d.db.QueryRow("SELECT COUNT(*) FROM maps WHERE campaign_id = ?", id).Scan(&s.Maps); err != nil {
		return s, err
	}
	return s, nil
}

func (d *DB) DeleteCampaign(id int64) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmts := []string{
		`DELETE FROM dice_rolls WHERE session_id IN (SELECT id FROM sessions WHERE campaign_id = ?)`,
		`DELETE FROM messages WHERE session_id IN (SELECT id FROM sessions WHERE campaign_id = ?)`,
		`DELETE FROM combatants WHERE encounter_id IN (SELECT id FROM combat_encounters WHERE session_id IN (SELECT id FROM sessions WHERE campaign_id = ?))`,
		`DELETE FROM combat_encounters WHERE session_id IN (SELECT id FROM sessions WHERE campaign_id = ?)`,
		`DELETE FROM sessions WHERE campaign_id = ?`,
		`DELETE FROM world_notes WHERE campaign_id = ?`,
		`DELETE FROM map_pins WHERE map_id IN (SELECT id FROM maps WHERE campaign_id = ?)`,
		`DELETE FROM maps WHERE campaign_id = ?`,
		`DELETE FROM characters WHERE campaign_id = ?`,
		`DELETE FROM campaigns WHERE id = ?`,
	}
	for _, stmt := range stmts {
		if _, err := tx.Exec(stmt, id); err != nil {
			return fmt.Errorf("delete campaign %d: %w", id, err)
		}
	}
	return tx.Commit()
}
```

- [ ] **Step 4: Run tests to verify they pass**

```
cd /home/digitalghost/projects/inkandbone
go test ./internal/db/... -run "TestCloseCampaign|TestReopenCampaign|TestGetCampaignStats|TestDeleteCampaign" -v
```

Expected: PASS all four

- [ ] **Step 5: Run full DB test suite**

```
cd /home/digitalghost/projects/inkandbone
go test ./internal/db/... -v
```

Expected: all pass

- [ ] **Step 6: Commit**

```bash
cd /home/digitalghost/projects/inkandbone
git add internal/db/queries_core.go internal/db/queries_core_test.go
git commit -m "feat(db): add CloseCampaign, ReopenCampaign, GetCampaignStats, DeleteCampaign"
```

---

## Task 2: Event constants

**Files:**
- Modify: `internal/api/events.go`

- [ ] **Step 1: Add three constants**

In `internal/api/events.go`, add to the `const` block after `EventSessionUpdated`:

```go
	EventCampaignClosed   EventType = "campaign_closed"
	EventCampaignDeleted  EventType = "campaign_deleted"
	EventCampaignReopened EventType = "campaign_reopened"
```

- [ ] **Step 2: Verify it compiles**

```
cd /home/digitalghost/projects/inkandbone
go build ./internal/api/...
```

Expected: no output (success)

- [ ] **Step 3: Commit**

```bash
cd /home/digitalghost/projects/inkandbone
git add internal/api/events.go
git commit -m "feat(api): add campaign_closed, campaign_deleted, campaign_reopened events"
```

---

## Task 3: handleCloseCampaign and handleDeleteCampaign

**Files:**
- Modify: `internal/mcp/lifecycle_test.go`
- Modify: `internal/mcp/lifecycle.go`
- Modify: `internal/mcp/server.go`

- [ ] **Step 1: Write failing tests**

Append to `internal/mcp/lifecycle_test.go`:

```go
func TestCloseCampaign(t *testing.T) {
	s := newTestMCP(t)
	campID, _, _ := setupCampaign(t, s)
	require.NoError(t, s.db.SetSetting("active_campaign_id", strconv.FormatInt(campID, 10)))

	req := mcplib.CallToolRequest{}
	result, err := s.handleCloseCampaign(context.Background(), req)
	require.NoError(t, err)
	require.False(t, result.IsError)

	// active_campaign_id cleared
	got, _ := s.db.GetSetting("active_campaign_id")
	assert.Empty(t, got)

	// campaign marked inactive
	c, err := s.db.GetCampaign(campID)
	require.NoError(t, err)
	assert.False(t, c.Active)
}

func TestCloseCampaign_withOpenSession(t *testing.T) {
	s := newTestMCP(t)
	campID, _, sessID := setupCampaign(t, s)
	require.NoError(t, s.db.SetSetting("active_campaign_id", strconv.FormatInt(campID, 10)))
	require.NoError(t, s.db.SetSetting("active_session_id", strconv.FormatInt(sessID, 10)))

	req := mcplib.CallToolRequest{}
	result, err := s.handleCloseCampaign(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, result.IsError)

	tc, ok := result.Content[0].(mcplib.TextContent)
	require.True(t, ok)
	assert.Contains(t, tc.Text, "end your current session")
}

func TestCloseCampaign_explicitID(t *testing.T) {
	s := newTestMCP(t)
	campID, _, _ := setupCampaign(t, s)

	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{"campaign_id": float64(campID)}
	result, err := s.handleCloseCampaign(context.Background(), req)
	require.NoError(t, err)
	require.False(t, result.IsError)

	c, err := s.db.GetCampaign(campID)
	require.NoError(t, err)
	assert.False(t, c.Active)
}

func TestDeleteCampaign_missingID(t *testing.T) {
	s := newTestMCP(t)
	req := mcplib.CallToolRequest{}
	result, err := s.handleDeleteCampaign(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestDeleteCampaign_noConfirm(t *testing.T) {
	s := newTestMCP(t)
	campID, _, _ := setupCampaign(t, s)

	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{"campaign_id": float64(campID)}
	result, err := s.handleDeleteCampaign(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, result.IsError)

	tc, ok := result.Content[0].(mcplib.TextContent)
	require.True(t, ok)
	assert.Contains(t, tc.Text, "confirm: true")

	// campaign still exists
	c, err := s.db.GetCampaign(campID)
	require.NoError(t, err)
	assert.NotNil(t, c)
}

func TestDeleteCampaign_confirm(t *testing.T) {
	s := newTestMCP(t)
	campID, _, _ := setupCampaign(t, s)
	require.NoError(t, s.db.SetSetting("active_campaign_id", strconv.FormatInt(campID, 10)))

	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"campaign_id": float64(campID),
		"confirm":     true,
	}
	result, err := s.handleDeleteCampaign(context.Background(), req)
	require.NoError(t, err)
	require.False(t, result.IsError)

	// campaign gone
	c, err := s.db.GetCampaign(campID)
	require.NoError(t, err)
	assert.Nil(t, c)

	// active_campaign_id cleared
	got, _ := s.db.GetSetting("active_campaign_id")
	assert.Empty(t, got)
}
```

- [ ] **Step 2: Run tests to verify they fail**

```
cd /home/digitalghost/projects/inkandbone
go test ./internal/mcp/... -run "TestCloseCampaign|TestDeleteCampaign" -v
```

Expected: FAIL — `s.handleCloseCampaign undefined`

- [ ] **Step 3: Implement handlers in lifecycle.go**

Append to `internal/mcp/lifecycle.go`:

```go
func (s *Server) handleCloseCampaign(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	var campID int64
	if id, ok := optInt64(req, "campaign_id"); ok && id > 0 {
		campID = id
	} else {
		idStr, err := s.db.GetSetting("active_campaign_id")
		if err != nil || idStr == "" {
			return mcplib.NewToolResultError("no active campaign — provide campaign_id or set one active"), nil
		}
		campID, err = strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return mcplib.NewToolResultError("invalid active_campaign_id in settings"), nil
		}
	}

	// Error if there is an open session belonging to this campaign.
	sessIDStr, _ := s.db.GetSetting("active_session_id")
	if sessIDStr != "" {
		if sessID, err := strconv.ParseInt(sessIDStr, 10, 64); err == nil {
			if sess, err := s.db.GetSession(sessID); err == nil && sess != nil && sess.CampaignID == campID {
				return mcplib.NewToolResultError("end your current session before closing the campaign"), nil
			}
		}
	}

	camp, err := s.db.GetCampaign(campID)
	if err != nil {
		return mcplib.NewToolResultError("db error: " + err.Error()), nil
	}
	if camp == nil {
		return mcplib.NewToolResultError(fmt.Sprintf("campaign %d not found", campID)), nil
	}

	if err := s.db.CloseCampaign(campID); err != nil {
		return mcplib.NewToolResultError("close campaign: " + err.Error()), nil
	}

	// Clear settings if this was the active campaign.
	activeIDStr, _ := s.db.GetSetting("active_campaign_id")
	if activeID, err := strconv.ParseInt(activeIDStr, 10, 64); err == nil && activeID == campID {
		_ = s.db.SetSetting("active_campaign_id", "")
		_ = s.db.SetSetting("active_session_id", "")
		_ = s.db.SetSetting("active_character_id", "")
	}

	s.bus.Publish(api.Event{Type: api.EventCampaignClosed, Payload: map[string]any{"campaign_id": campID}})
	return mcplib.NewToolResultText(fmt.Sprintf("campaign %d closed: %s", campID, camp.Name)), nil
}

func (s *Server) handleDeleteCampaign(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	campID, ok := optInt64(req, "campaign_id")
	if !ok || campID <= 0 {
		return mcplib.NewToolResultError("campaign_id is required"), nil
	}

	camp, err := s.db.GetCampaign(campID)
	if err != nil {
		return mcplib.NewToolResultError("db error: " + err.Error()), nil
	}
	if camp == nil {
		return mcplib.NewToolResultError(fmt.Sprintf("campaign %d not found", campID)), nil
	}

	confirm, _ := req.GetArguments()["confirm"].(bool)
	if !confirm {
		stats, err := s.db.GetCampaignStats(campID)
		if err != nil {
			return mcplib.NewToolResultError("stats error: " + err.Error()), nil
		}
		return mcplib.NewToolResultError(fmt.Sprintf(
			"campaign %d %q and all its data will be permanently deleted:\n  - %d sessions, %d characters, %d world notes, %d maps\ncall delete_campaign again with confirm: true to proceed",
			campID, camp.Name, stats.Sessions, stats.Characters, stats.WorldNotes, stats.Maps,
		)), nil
	}

	if err := s.db.DeleteCampaign(campID); err != nil {
		return mcplib.NewToolResultError("delete campaign: " + err.Error()), nil
	}

	// Clear settings if this was the active campaign.
	activeIDStr, _ := s.db.GetSetting("active_campaign_id")
	if activeID, err := strconv.ParseInt(activeIDStr, 10, 64); err == nil && activeID == campID {
		_ = s.db.SetSetting("active_campaign_id", "")
		_ = s.db.SetSetting("active_session_id", "")
		_ = s.db.SetSetting("active_character_id", "")
	}

	s.bus.Publish(api.Event{Type: api.EventCampaignDeleted, Payload: map[string]any{"campaign_id": campID}})
	return mcplib.NewToolResultText(fmt.Sprintf("campaign %d deleted", campID)), nil
}
```

- [ ] **Step 4: Register the two new tools in server.go**

In `internal/mcp/server.go`, in `registerTools()`, append after the `list_sessions` tool block (before the `// Maps` comment):

```go
	s.srv.AddTool(mcplib.NewTool("close_campaign",
		mcplib.WithDescription("Close the active (or specified) campaign. Errors if a session is still open. Closed campaigns remain visible in list_campaigns and can be reopened via set_active."),
		mcplib.WithNumber("campaign_id", mcplib.Description("Campaign ID to close (defaults to active campaign)")),
	), s.handleCloseCampaign)

	s.srv.AddTool(mcplib.NewTool("delete_campaign",
		mcplib.WithDescription("Permanently delete a campaign and all its data (sessions, characters, world notes, maps). Without confirm:true, returns a summary of what will be deleted. This cannot be undone."),
		mcplib.WithNumber("campaign_id", mcplib.Required(), mcplib.Description("Campaign ID to delete")),
		mcplib.WithBoolean("confirm", mcplib.Description("Must be true to confirm permanent deletion")),
	), s.handleDeleteCampaign)
```

- [ ] **Step 5: Run tests to verify they pass**

```
cd /home/digitalghost/projects/inkandbone
go test ./internal/mcp/... -run "TestCloseCampaign|TestDeleteCampaign" -v
```

Expected: all pass

- [ ] **Step 6: Commit**

```bash
cd /home/digitalghost/projects/inkandbone
git add internal/mcp/lifecycle.go internal/mcp/lifecycle_test.go internal/mcp/server.go
git commit -m "feat(mcp): add close_campaign and delete_campaign tools"
```

---

## Task 4: set_active auto-reopen

**Files:**
- Modify: `internal/mcp/campaign_test.go`
- Modify: `internal/mcp/campaign.go`

- [ ] **Step 1: Write failing test**

Append to `internal/mcp/campaign_test.go`:

```go
func TestSetActive_reopensClosed(t *testing.T) {
	s := newTestMCP(t)
	campID, _, _ := setupCampaign(t, s)

	// Close the campaign first
	require.NoError(t, s.db.CloseCampaign(campID))
	c, err := s.db.GetCampaign(campID)
	require.NoError(t, err)
	require.False(t, c.Active)

	// set_active should reopen it
	req := mcplib.CallToolRequest{}
	req.Params.Arguments = map[string]any{"campaign_id": float64(campID)}
	result, err := s.handleSetActive(context.Background(), req)
	require.NoError(t, err)
	require.False(t, result.IsError)

	c, err = s.db.GetCampaign(campID)
	require.NoError(t, err)
	assert.True(t, c.Active)

	got, _ := s.db.GetSetting("active_campaign_id")
	assert.Equal(t, strconv.FormatInt(campID, 10), got)
}
```

- [ ] **Step 2: Run test to verify it fails**

```
cd /home/digitalghost/projects/inkandbone
go test ./internal/mcp/... -run "TestSetActive_reopensClosed" -v
```

Expected: FAIL — campaign remains closed after set_active

- [ ] **Step 3: Update handleSetActive in campaign.go**

Replace the existing campaign_id block in `handleSetActive`:

```go
// Before (lines ~13-17):
	if id, ok := optInt64(req, "campaign_id"); ok && id > 0 {
		if err := s.db.SetSetting("active_campaign_id", strconv.FormatInt(id, 10)); err != nil {
			return mcplib.NewToolResultError("set campaign: " + err.Error()), nil
		}
	}
```

With:

```go
	if id, ok := optInt64(req, "campaign_id"); ok && id > 0 {
		camp, err := s.db.GetCampaign(id)
		if err != nil {
			return mcplib.NewToolResultError("db error: " + err.Error()), nil
		}
		if camp != nil && !camp.Active {
			if err := s.db.ReopenCampaign(id); err != nil {
				return mcplib.NewToolResultError("reopen campaign: " + err.Error()), nil
			}
			s.bus.Publish(api.Event{Type: api.EventCampaignReopened, Payload: map[string]any{"campaign_id": id}})
		}
		if err := s.db.SetSetting("active_campaign_id", strconv.FormatInt(id, 10)); err != nil {
			return mcplib.NewToolResultError("set campaign: " + err.Error()), nil
		}
	}
```

- [ ] **Step 4: Run test to verify it passes**

```
cd /home/digitalghost/projects/inkandbone
go test ./internal/mcp/... -run "TestSetActive_reopensClosed" -v
```

Expected: PASS

- [ ] **Step 5: Run full test suite**

```
cd /home/digitalghost/projects/inkandbone
go test ./... -v
```

Expected: all pass

- [ ] **Step 6: Commit**

```bash
cd /home/digitalghost/projects/inkandbone
git add internal/mcp/campaign.go internal/mcp/campaign_test.go
git commit -m "feat(mcp): set_active auto-reopens closed campaigns"
```
