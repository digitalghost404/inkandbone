# Phase D: Narrative Systems — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add three interlocking narrative systems: an Ironsworn-style oracle table for creative prompts, a chaos/tension meter that tracks story stakes, and a faction/relationship tracker for political dynamics.

**Architecture:** Migration 013 adds three tables/columns. Each system has its own DB query file and route file, staying consistent with the existing pattern. The tension meter updates rule-based (no AI calls) via a new `autoUpdateTension` goroutine in routes.go. The oracle table is seeded generically and rulesets can override.

**Tech Stack:** Go, SQLite, React/TypeScript

---

## File Map

| File | Change |
|------|--------|
| `internal/db/migrations/013_phase_d.sql` | Create — oracle_tables, sessions.tension_level, relationships |
| `internal/db/queries_oracle.go` | Create — oracle table seeding + GetOracleRoll |
| `internal/db/queries_oracle_test.go` | Create — oracle DB tests |
| `internal/db/queries_tension.go` | Create — GetTension, UpdateTension |
| `internal/db/queries_tension_test.go` | Create — tension DB tests |
| `internal/db/queries_relationships.go` | Create — Relationship CRUD |
| `internal/db/queries_relationships_test.go` | Create — relationship DB tests |
| `internal/api/routes_phase_d.go` | Create — oracle roll handler, tension endpoint, relationship CRUD handlers |
| `internal/api/server.go` | Modify — register Phase D routes |
| `internal/api/routes_phase_d_test.go` | Create — API tests |
| `internal/api/routes.go` | Modify — add autoUpdateTension goroutine call in handleGMRespondStream |
| `internal/api/events.go` | Modify — add EventOracleRolled, EventTensionUpdated, EventRelationshipUpdated |
| `web/src/types.ts` | Modify — add Oracle, Tension, Relationship types |
| `web/src/OraclePanel.tsx` | Create — Oracle roll UI + tension gauge |
| `web/src/RelationshipsPanel.tsx` | Create — Faction/relationship tracker UI |

---

### Task D1: Migration 013 — oracle_tables, tension_level, relationships

**Files:**
- Create: `internal/db/migrations/013_phase_d.sql`

- [ ] **Step 1: Write the migration**

```sql
-- 013_phase_d.sql: Oracle tables, tension meter, faction relationships

-- Generic oracle prompt tables (action + theme)
CREATE TABLE IF NOT EXISTS oracle_tables (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ruleset_id INTEGER,                   -- NULL = generic (all campaigns), non-NULL = ruleset-specific
    table_name TEXT NOT NULL,             -- 'action' or 'theme'
    roll_min INTEGER NOT NULL,            -- 1-100 range start
    roll_max INTEGER NOT NULL,            -- 1-100 range end
    result TEXT NOT NULL,
    FOREIGN KEY (ruleset_id) REFERENCES rulesets(id) ON DELETE CASCADE
);

-- Tension level on sessions (1-10 scale, starts at 5)
ALTER TABLE sessions ADD COLUMN tension_level INTEGER NOT NULL DEFAULT 5;

-- Faction/relationship tracker (campaign-scoped)
CREATE TABLE IF NOT EXISTS relationships (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id INTEGER NOT NULL,
    entity_a TEXT NOT NULL,              -- character or faction name
    entity_b TEXT NOT NULL,              -- character or faction name
    score INTEGER NOT NULL DEFAULT 0,   -- -10 (hostile) to +10 (allied)
    disposition TEXT NOT NULL DEFAULT 'neutral', -- 'hostile','cold','neutral','warm','allied'
    notes TEXT NOT NULL DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE CASCADE
);

-- Seed generic oracle tables (Action d100 + Theme d100)
-- Action table (Ironsworn-inspired, generic enough for any system)
INSERT INTO oracle_tables (ruleset_id, table_name, roll_min, roll_max, result) VALUES
(NULL,'action',1,2,'Abandon'),(NULL,'action',3,4,'Advance'),(NULL,'action',5,6,'Affect'),
(NULL,'action',7,8,'Aid'),(NULL,'action',9,10,'Arrive'),(NULL,'action',11,12,'Assault'),
(NULL,'action',13,14,'Attack'),(NULL,'action',15,16,'Betray'),(NULL,'action',17,18,'Bolster'),
(NULL,'action',19,20,'Breach'),(NULL,'action',21,22,'Capture'),(NULL,'action',23,24,'Challenge'),
(NULL,'action',25,26,'Change'),(NULL,'action',27,28,'Charge'),(NULL,'action',29,30,'Command'),
(NULL,'action',31,32,'Conceal'),(NULL,'action',33,34,'Create'),(NULL,'action',35,36,'Defy'),
(NULL,'action',37,38,'Deliver'),(NULL,'action',39,40,'Demand'),(NULL,'action',41,42,'Destroy'),
(NULL,'action',43,44,'Discover'),(NULL,'action',45,46,'Evade'),(NULL,'action',47,48,'Explore'),
(NULL,'action',49,50,'Expose'),(NULL,'action',51,52,'Fail'),(NULL,'action',53,54,'Find'),
(NULL,'action',55,56,'Flee'),(NULL,'action',57,58,'Follow'),(NULL,'action',59,60,'Guard'),
(NULL,'action',61,62,'Hide'),(NULL,'action',63,64,'Hinder'),(NULL,'action',65,66,'Inspect'),
(NULL,'action',67,68,'Lead'),(NULL,'action',69,70,'Learn'),(NULL,'action',71,72,'Leave'),
(NULL,'action',73,74,'Locate'),(NULL,'action',75,76,'Lose'),(NULL,'action',77,78,'Manipulate'),
(NULL,'action',79,80,'Oppose'),(NULL,'action',81,82,'Overwhelm'),(NULL,'action',83,84,'Persevere'),
(NULL,'action',85,86,'Protect'),(NULL,'action',87,88,'Reveal'),(NULL,'action',89,90,'Seek'),
(NULL,'action',91,92,'Seize'),(NULL,'action',93,94,'Strike'),(NULL,'action',95,96,'Transform'),
(NULL,'action',97,98,'Trap'),(NULL,'action',99,100,'Uncover');

-- Theme table
INSERT INTO oracle_tables (ruleset_id, table_name, roll_min, roll_max, result) VALUES
(NULL,'theme',1,2,'Ability'),(NULL,'theme',3,4,'Alliance'),(NULL,'theme',5,6,'Ambition'),
(NULL,'theme',7,8,'Ancient'),(NULL,'theme',9,10,'Arcane'),(NULL,'theme',11,12,'Belief'),
(NULL,'theme',13,14,'Blood'),(NULL,'theme',15,16,'Burden'),(NULL,'theme',17,18,'Commerce'),
(NULL,'theme',19,20,'Corruption'),(NULL,'theme',21,22,'Creation'),(NULL,'theme',23,24,'Danger'),
(NULL,'theme',25,26,'Death'),(NULL,'theme',27,28,'Debt'),(NULL,'theme',29,30,'Deception'),
(NULL,'theme',31,32,'Defense'),(NULL,'theme',33,34,'Destiny'),(NULL,'theme',35,36,'Disaster'),
(NULL,'theme',37,38,'Discovery'),(NULL,'theme',39,40,'Enemy'),(NULL,'theme',41,42,'Escape'),
(NULL,'theme',43,44,'Exploration'),(NULL,'theme',45,46,'Faith'),(NULL,'theme',47,48,'Fame'),
(NULL,'theme',49,50,'Fear'),(NULL,'theme',51,52,'Freedom'),(NULL,'theme',53,54,'Greed'),
(NULL,'theme',55,56,'Grief'),(NULL,'theme',57,58,'History'),(NULL,'theme',59,60,'Honor'),
(NULL,'theme',61,62,'Hope'),(NULL,'theme',63,64,'Isolation'),(NULL,'theme',65,66,'Justice'),
(NULL,'theme',67,68,'Knowledge'),(NULL,'theme',69,70,'Leadership'),(NULL,'theme',71,72,'Legacy'),
(NULL,'theme',73,74,'Loss'),(NULL,'theme',75,76,'Love'),(NULL,'theme',77,78,'Loyalty'),
(NULL,'theme',79,80,'Mystery'),(NULL,'theme',81,82,'Nature'),(NULL,'theme',83,84,'Pain'),
(NULL,'theme',85,86,'Power'),(NULL,'theme',87,88,'Protection'),(NULL,'theme',89,90,'Revenge'),
(NULL,'theme',91,92,'Rivalry'),(NULL,'theme',93,94,'Survival'),(NULL,'theme',95,96,'Trade'),
(NULL,'theme',97,98,'Truth'),(NULL,'theme',99,100,'War');
```

- [ ] **Step 2: Verify migration applies**

Run: `make test`
Expected: PASS (migrations auto-applied in newTestDB)

- [ ] **Step 3: Commit**

```bash
git add internal/db/migrations/013_phase_d.sql
git commit -m "feat(db): migration 013 — oracle tables, tension meter, relationships"
```

---

### Task D2: DB — Oracle roll query

**Files:**
- Create: `internal/db/queries_oracle.go`
- Create: `internal/db/queries_oracle_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/db/queries_oracle_test.go`:

```go
package db

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestRollOracle(t *testing.T) {
    d := newTestDB(t)

    // Roll action oracle (generic, no ruleset)
    result, err := d.RollOracle(nil, "action", 1)
    require.NoError(t, err)
    assert.Equal(t, "Abandon", result)

    result, err = d.RollOracle(nil, "action", 50)
    require.NoError(t, err)
    assert.Equal(t, "Expose", result)

    result, err = d.RollOracle(nil, "action", 100)
    require.NoError(t, err)
    assert.Equal(t, "Uncover", result)
}

func TestRollOracleTheme(t *testing.T) {
    d := newTestDB(t)

    result, err := d.RollOracle(nil, "theme", 25)
    require.NoError(t, err)
    assert.Equal(t, "Danger", result)
}

func TestRollOracleNotFound(t *testing.T) {
    d := newTestDB(t)
    _, err := d.RollOracle(nil, "unknown_table", 50)
    require.Error(t, err)
}
```

- [ ] **Step 2: Run tests — expect fail**

Run: `go test ./internal/db/... -run "TestRollOracle" -v`
Expected: FAIL — undefined RollOracle

- [ ] **Step 3: Create queries_oracle.go**

Create `internal/db/queries_oracle.go`:

```go
package db

import (
    "database/sql"
    "fmt"
)

// RollOracle returns the oracle result for a given roll (1-100) and table name.
// rulesetID may be nil for generic oracle tables.
func (d *DB) RollOracle(rulesetID *int64, tableName string, roll int) (string, error) {
    var result string
    var err error
    if rulesetID != nil {
        // Try ruleset-specific first, fall back to generic
        err = d.db.QueryRow(
            "SELECT result FROM oracle_tables WHERE table_name = ? AND roll_min <= ? AND roll_max >= ? AND ruleset_id = ? LIMIT 1",
            tableName, roll, roll, *rulesetID,
        ).Scan(&result)
        if err == sql.ErrNoRows {
            err = d.db.QueryRow(
                "SELECT result FROM oracle_tables WHERE table_name = ? AND roll_min <= ? AND roll_max >= ? AND ruleset_id IS NULL LIMIT 1",
                tableName, roll, roll,
            ).Scan(&result)
        }
    } else {
        err = d.db.QueryRow(
            "SELECT result FROM oracle_tables WHERE table_name = ? AND roll_min <= ? AND roll_max >= ? AND ruleset_id IS NULL LIMIT 1",
            tableName, roll, roll,
        ).Scan(&result)
    }
    if err == sql.ErrNoRows {
        return "", fmt.Errorf("no oracle result for table %q roll %d", tableName, roll)
    }
    return result, err
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/db/... -run "TestRollOracle" -v`
Expected: PASS

- [ ] **Step 5: Run full suite**

Run: `make test`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/db/queries_oracle.go internal/db/queries_oracle_test.go
git commit -m "feat(db): oracle table queries — RollOracle with ruleset fallback to generic"
```

---

### Task D3: DB — Tension level queries

**Files:**
- Create: `internal/db/queries_tension.go`
- Create: `internal/db/queries_tension_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/db/queries_tension_test.go`:

```go
package db

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestGetSetTension(t *testing.T) {
    d := newTestDB(t)
    campID := setupCampaign(t, d)
    sessID, err := d.CreateSession(campID, "S1", "2026-04-03")
    require.NoError(t, err)

    level, err := d.GetTension(sessID)
    require.NoError(t, err)
    assert.Equal(t, 5, level) // default

    require.NoError(t, d.UpdateTension(sessID, 7))
    level, err = d.GetTension(sessID)
    require.NoError(t, err)
    assert.Equal(t, 7, level)

    // Clamp to 1-10
    require.NoError(t, d.UpdateTension(sessID, 15))
    level, err = d.GetTension(sessID)
    require.NoError(t, err)
    assert.Equal(t, 10, level)

    require.NoError(t, d.UpdateTension(sessID, -3))
    level, err = d.GetTension(sessID)
    require.NoError(t, err)
    assert.Equal(t, 1, level)
}
```

- [ ] **Step 2: Run test — expect fail**

Run: `go test ./internal/db/... -run "TestGetSetTension" -v`
Expected: FAIL — undefined GetTension

- [ ] **Step 3: Create queries_tension.go**

Create `internal/db/queries_tension.go`:

```go
package db

import "database/sql"

// GetTension returns the tension level (1-10) for a session. Default is 5.
func (d *DB) GetTension(sessionID int64) (int, error) {
    var level int
    err := d.db.QueryRow("SELECT tension_level FROM sessions WHERE id = ?", sessionID).Scan(&level)
    if err == sql.ErrNoRows {
        return 5, nil
    }
    return level, err
}

// UpdateTension sets the tension level for a session, clamped to 1-10.
func (d *DB) UpdateTension(sessionID int64, level int) error {
    if level < 1 {
        level = 1
    }
    if level > 10 {
        level = 10
    }
    _, err := d.db.Exec("UPDATE sessions SET tension_level = ? WHERE id = ?", level, sessionID)
    return err
}
```

- [ ] **Step 4: Run test**

Run: `go test ./internal/db/... -run "TestGetSetTension" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/db/queries_tension.go internal/db/queries_tension_test.go
git commit -m "feat(db): tension level queries — GetTension, UpdateTension with 1-10 clamp"
```

---

### Task D4: DB — Relationships CRUD

**Files:**
- Create: `internal/db/queries_relationships.go`
- Create: `internal/db/queries_relationships_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/db/queries_relationships_test.go`:

```go
package db

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestRelationshipCRUD(t *testing.T) {
    d := newTestDB(t)
    campID := setupCampaign(t, d)

    // Create
    rel, err := d.CreateRelationship(campID, "Thorin", "The Iron Guild", 3, "warm", "Thorin owes them a debt")
    require.NoError(t, err)
    assert.Equal(t, "Thorin", rel.EntityA)
    assert.Equal(t, "The Iron Guild", rel.EntityB)
    assert.Equal(t, 3, rel.Score)

    // List
    list, err := d.ListRelationships(campID)
    require.NoError(t, err)
    assert.Len(t, list, 1)

    // Update
    require.NoError(t, d.UpdateRelationship(rel.ID, -2, "cold", "betrayal"))
    list, err = d.ListRelationships(campID)
    require.NoError(t, err)
    assert.Equal(t, -2, list[0].Score)
    assert.Equal(t, "cold", list[0].Disposition)

    // Delete
    require.NoError(t, d.DeleteRelationship(rel.ID))
    list, err = d.ListRelationships(campID)
    require.NoError(t, err)
    assert.Empty(t, list)
}
```

- [ ] **Step 2: Run test — expect fail**

Run: `go test ./internal/db/... -run "TestRelationshipCRUD" -v`
Expected: FAIL — undefined CreateRelationship

- [ ] **Step 3: Create queries_relationships.go**

Create `internal/db/queries_relationships.go`:

```go
package db

// Relationship tracks a political or personal connection between two entities.
type Relationship struct {
    ID          int64  `json:"id"`
    CampaignID  int64  `json:"campaign_id"`
    EntityA     string `json:"entity_a"`
    EntityB     string `json:"entity_b"`
    Score       int    `json:"score"`        // -10 to +10
    Disposition string `json:"disposition"`  // hostile, cold, neutral, warm, allied
    Notes       string `json:"notes"`
    CreatedAt   string `json:"created_at"`
}

func (d *DB) CreateRelationship(campaignID int64, entityA, entityB string, score int, disposition, notes string) (*Relationship, error) {
    res, err := d.db.Exec(
        "INSERT INTO relationships (campaign_id, entity_a, entity_b, score, disposition, notes) VALUES (?, ?, ?, ?, ?, ?)",
        campaignID, entityA, entityB, score, disposition, notes,
    )
    if err != nil {
        return nil, err
    }
    id, err := res.LastInsertId()
    if err != nil {
        return nil, err
    }
    return &Relationship{
        ID: id, CampaignID: campaignID,
        EntityA: entityA, EntityB: entityB,
        Score: score, Disposition: disposition, Notes: notes,
    }, nil
}

func (d *DB) ListRelationships(campaignID int64) ([]Relationship, error) {
    rows, err := d.db.Query(
        "SELECT id, campaign_id, entity_a, entity_b, score, disposition, notes, created_at FROM relationships WHERE campaign_id = ? ORDER BY created_at DESC",
        campaignID,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var out []Relationship
    for rows.Next() {
        var r Relationship
        if err := rows.Scan(&r.ID, &r.CampaignID, &r.EntityA, &r.EntityB, &r.Score, &r.Disposition, &r.Notes, &r.CreatedAt); err != nil {
            return nil, err
        }
        out = append(out, r)
    }
    return out, rows.Err()
}

func (d *DB) UpdateRelationship(id int64, score int, disposition, notes string) error {
    _, err := d.db.Exec(
        "UPDATE relationships SET score = ?, disposition = ?, notes = ? WHERE id = ?",
        score, disposition, notes, id,
    )
    return err
}

func (d *DB) DeleteRelationship(id int64) error {
    _, err := d.db.Exec("DELETE FROM relationships WHERE id = ?", id)
    return err
}
```

- [ ] **Step 4: Run test**

Run: `go test ./internal/db/... -run "TestRelationshipCRUD" -v`
Expected: PASS

- [ ] **Step 5: Run full suite**

Run: `make test`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/db/queries_relationships.go internal/db/queries_relationships_test.go
git commit -m "feat(db): relationships CRUD — CreateRelationship, ListRelationships, UpdateRelationship, DeleteRelationship"
```

---

### Task D5: API — Oracle roll endpoint + tension endpoint

**Files:**
- Create: `internal/api/routes_phase_d.go`
- Modify: `internal/api/server.go`
- Create: `internal/api/routes_phase_d_test.go`
- Modify: `internal/api/events.go`

- [ ] **Step 1: Add new event types**

In `internal/api/events.go`, add to the const block:

```go
EventOracleRolled      EventType = "oracle_rolled"
EventTensionUpdated    EventType = "tension_updated"
EventRelationshipUpdated EventType = "relationship_updated"
```

- [ ] **Step 2: Write the failing tests**

Create `internal/api/routes_phase_d_test.go`:

```go
package api

import (
    "encoding/json"
    "fmt"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestOracleRoll(t *testing.T) {
    s := newTestServer(t)
    _, sessID := seedCampaign(t, s.db)

    req := httptest.NewRequest(http.MethodPost,
        fmt.Sprintf("/api/sessions/%d/oracle-roll", sessID), nil)
    w := httptest.NewRecorder()
    s.ServeHTTP(w, req)
    assert.Equal(t, http.StatusOK, w.Code)

    var result map[string]any
    require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
    assert.NotEmpty(t, result["action"])
    assert.NotEmpty(t, result["theme"])
    assert.NotZero(t, result["action_roll"])
    assert.NotZero(t, result["theme_roll"])
}

func TestGetTension(t *testing.T) {
    s := newTestServer(t)
    _, sessID := seedCampaign(t, s.db)

    req := httptest.NewRequest(http.MethodGet,
        fmt.Sprintf("/api/sessions/%d/tension", sessID), nil)
    w := httptest.NewRecorder()
    s.ServeHTTP(w, req)
    assert.Equal(t, http.StatusOK, w.Code)

    var result map[string]any
    require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
    assert.EqualValues(t, 5, result["tension_level"])
}

func TestPatchTension(t *testing.T) {
    s := newTestServer(t)
    _, sessID := seedCampaign(t, s.db)

    body := `{"tension_level":8}`
    req := httptest.NewRequest(http.MethodPatch,
        fmt.Sprintf("/api/sessions/%d/tension", sessID),
        strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    s.ServeHTTP(w, req)
    assert.Equal(t, http.StatusNoContent, w.Code)

    level, err := s.db.GetTension(sessID)
    require.NoError(t, err)
    assert.Equal(t, 8, level)
}
```

- [ ] **Step 3: Run tests — expect 404**

Run: `go test ./internal/api/... -run "TestOracleRoll|TestGetTension|TestPatchTension" -v`
Expected: FAIL — 404

- [ ] **Step 4: Create routes_phase_d.go with oracle and tension handlers**

Create `internal/api/routes_phase_d.go`:

```go
package api

import (
    "encoding/json"
    "math/rand/v2"
    "net/http"
)

// handleOracleRoll rolls both the action and theme oracle tables.
// POST /api/sessions/{id}/oracle-roll
func (s *Server) handleOracleRoll(w http.ResponseWriter, r *http.Request) {
    id, ok := parsePathID(r, "id")
    if !ok {
        http.Error(w, "invalid session id", http.StatusBadRequest)
        return
    }
    // Verify session exists
    sess, err := s.db.GetSession(id)
    if err != nil || sess == nil {
        http.Error(w, "session not found", http.StatusNotFound)
        return
    }

    actionRoll := rand.IntN(100) + 1
    themeRoll := rand.IntN(100) + 1

    action, err := s.db.RollOracle(nil, "action", actionRoll)
    if err != nil {
        http.Error(w, "oracle error: "+err.Error(), http.StatusInternalServerError)
        return
    }
    theme, err := s.db.RollOracle(nil, "theme", themeRoll)
    if err != nil {
        http.Error(w, "oracle error: "+err.Error(), http.StatusInternalServerError)
        return
    }

    result := map[string]any{
        "action":      action,
        "theme":       theme,
        "action_roll": actionRoll,
        "theme_roll":  themeRoll,
        "session_id":  id,
    }
    s.bus.Publish(Event{Type: EventOracleRolled, Payload: result})
    writeJSON(w, result)
}

// handleGetTension returns the current tension level for a session.
// GET /api/sessions/{id}/tension
func (s *Server) handleGetTension(w http.ResponseWriter, r *http.Request) {
    id, ok := parsePathID(r, "id")
    if !ok {
        http.Error(w, "invalid session id", http.StatusBadRequest)
        return
    }
    level, err := s.db.GetTension(id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    writeJSON(w, map[string]any{"session_id": id, "tension_level": level})
}

// handlePatchTension sets the tension level for a session.
// PATCH /api/sessions/{id}/tension
func (s *Server) handlePatchTension(w http.ResponseWriter, r *http.Request) {
    id, ok := parsePathID(r, "id")
    if !ok {
        http.Error(w, "invalid session id", http.StatusBadRequest)
        return
    }
    var body struct {
        TensionLevel int `json:"tension_level"`
    }
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        http.Error(w, "invalid json", http.StatusBadRequest)
        return
    }
    if err := s.db.UpdateTension(id, body.TensionLevel); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    level, _ := s.db.GetTension(id)
    s.bus.Publish(Event{Type: EventTensionUpdated, Payload: map[string]any{
        "session_id":    id,
        "tension_level": level,
    }})
    w.WriteHeader(http.StatusNoContent)
}
```

- [ ] **Step 5: Register oracle and tension routes in server.go**

In `registerRoutes()`, add a Phase D block:

```go
// Phase D: Narrative systems
s.mux.HandleFunc("POST /api/sessions/{id}/oracle-roll", s.handleOracleRoll)
s.mux.HandleFunc("GET /api/sessions/{id}/tension", s.handleGetTension)
s.mux.HandleFunc("PATCH /api/sessions/{id}/tension", s.handlePatchTension)
```

- [ ] **Step 6: Run tests**

Run: `go test ./internal/api/... -run "TestOracleRoll|TestGetTension|TestPatchTension" -v`
Expected: PASS

- [ ] **Step 7: Run full suite**

Run: `make test`
Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add internal/api/routes_phase_d.go internal/api/server.go internal/api/routes_phase_d_test.go internal/api/events.go
git commit -m "feat(api): oracle roll endpoint, tension GET/PATCH, new event types"
```

---

### Task D6: API — Relationships CRUD routes + autoUpdateTension goroutine

**Files:**
- Modify: `internal/api/routes_phase_d.go`
- Modify: `internal/api/server.go`
- Modify: `internal/api/routes_phase_d_test.go`
- Modify: `internal/api/routes.go`

- [ ] **Step 1: Write the failing tests**

Add to `internal/api/routes_phase_d_test.go`:

```go
func TestRelationshipsCRUD(t *testing.T) {
    s := newTestServer(t)
    campID, _ := seedCampaign(t, s.db)

    // Create
    body := `{"entity_a":"Thorin","entity_b":"Iron Guild","score":3,"disposition":"warm","notes":"owes debt"}`
    req := httptest.NewRequest(http.MethodPost,
        fmt.Sprintf("/api/campaigns/%d/relationships", campID),
        strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    s.ServeHTTP(w, req)
    assert.Equal(t, http.StatusCreated, w.Code)

    var rel map[string]any
    require.NoError(t, json.Unmarshal(w.Body.Bytes(), &rel))
    assert.Equal(t, "Thorin", rel["entity_a"])

    relID := int64(rel["id"].(float64))

    // List
    req = httptest.NewRequest(http.MethodGet,
        fmt.Sprintf("/api/campaigns/%d/relationships", campID), nil)
    w = httptest.NewRecorder()
    s.ServeHTTP(w, req)
    assert.Equal(t, http.StatusOK, w.Code)
    var list []map[string]any
    require.NoError(t, json.Unmarshal(w.Body.Bytes(), &list))
    assert.Len(t, list, 1)

    // Patch
    body = `{"score":-2,"disposition":"cold","notes":"betrayed"}`
    req = httptest.NewRequest(http.MethodPatch,
        fmt.Sprintf("/api/relationships/%d", relID),
        strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w = httptest.NewRecorder()
    s.ServeHTTP(w, req)
    assert.Equal(t, http.StatusNoContent, w.Code)

    // Delete
    req = httptest.NewRequest(http.MethodDelete,
        fmt.Sprintf("/api/relationships/%d", relID), nil)
    w = httptest.NewRecorder()
    s.ServeHTTP(w, req)
    assert.Equal(t, http.StatusNoContent, w.Code)
}
```

- [ ] **Step 2: Run test — expect 404**

Run: `go test ./internal/api/... -run "TestRelationshipsCRUD" -v`
Expected: FAIL — 404

- [ ] **Step 3: Add relationship handlers to routes_phase_d.go**

Add to `internal/api/routes_phase_d.go`:

```go
// handleListRelationships returns all relationships for a campaign.
// GET /api/campaigns/{id}/relationships
func (s *Server) handleListRelationships(w http.ResponseWriter, r *http.Request) {
    id, ok := parsePathID(r, "id")
    if !ok {
        http.Error(w, "invalid campaign id", http.StatusBadRequest)
        return
    }
    list, err := s.db.ListRelationships(id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if list == nil {
        list = []db.Relationship{}
    }
    writeJSON(w, list)
}

// handleCreateRelationship creates a new relationship.
// POST /api/campaigns/{id}/relationships
func (s *Server) handleCreateRelationship(w http.ResponseWriter, r *http.Request) {
    id, ok := parsePathID(r, "id")
    if !ok {
        http.Error(w, "invalid campaign id", http.StatusBadRequest)
        return
    }
    var body struct {
        EntityA     string `json:"entity_a"`
        EntityB     string `json:"entity_b"`
        Score       int    `json:"score"`
        Disposition string `json:"disposition"`
        Notes       string `json:"notes"`
    }
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        http.Error(w, "invalid json", http.StatusBadRequest)
        return
    }
    if body.EntityA == "" || body.EntityB == "" {
        http.Error(w, "entity_a and entity_b are required", http.StatusBadRequest)
        return
    }
    if body.Disposition == "" {
        body.Disposition = "neutral"
    }
    rel, err := s.db.CreateRelationship(id, body.EntityA, body.EntityB, body.Score, body.Disposition, body.Notes)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    s.bus.Publish(Event{Type: EventRelationshipUpdated, Payload: map[string]any{
        "campaign_id":     id,
        "relationship_id": rel.ID,
    }})
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(rel)
}

// handlePatchRelationship updates score, disposition, and notes on a relationship.
// PATCH /api/relationships/{id}
func (s *Server) handlePatchRelationship(w http.ResponseWriter, r *http.Request) {
    id, ok := parsePathID(r, "id")
    if !ok {
        http.Error(w, "invalid id", http.StatusBadRequest)
        return
    }
    var body struct {
        Score       int    `json:"score"`
        Disposition string `json:"disposition"`
        Notes       string `json:"notes"`
    }
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        http.Error(w, "invalid json", http.StatusBadRequest)
        return
    }
    if err := s.db.UpdateRelationship(id, body.Score, body.Disposition, body.Notes); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    s.bus.Publish(Event{Type: EventRelationshipUpdated, Payload: map[string]any{"relationship_id": id}})
    w.WriteHeader(http.StatusNoContent)
}

// handleDeleteRelationship removes a relationship.
// DELETE /api/relationships/{id}
func (s *Server) handleDeleteRelationship(w http.ResponseWriter, r *http.Request) {
    id, ok := parsePathID(r, "id")
    if !ok {
        http.Error(w, "invalid id", http.StatusBadRequest)
        return
    }
    if err := s.db.DeleteRelationship(id); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}
```

Add the import for the db package at the top of routes_phase_d.go (since Relationship type is used):

```go
import (
    "encoding/json"
    "math/rand/v2"
    "net/http"

    "github.com/digitalghost404/inkandbone/internal/db"
)
```

- [ ] **Step 4: Register relationship routes in server.go**

Add to the Phase D block:

```go
s.mux.HandleFunc("GET /api/campaigns/{id}/relationships", s.handleListRelationships)
s.mux.HandleFunc("POST /api/campaigns/{id}/relationships", s.handleCreateRelationship)
s.mux.HandleFunc("PATCH /api/relationships/{id}", s.handlePatchRelationship)
s.mux.HandleFunc("DELETE /api/relationships/{id}", s.handleDeleteRelationship)
```

- [ ] **Step 5: Add autoUpdateTension to routes.go**

In `internal/api/routes.go`, add the `autoUpdateTension` function (near other auto* functions):

```go
// autoUpdateTension applies rule-based tension adjustments after a GM response.
// Rules: combat start → +2, failed roll logged this turn → +1, GM text contains crisis keywords → +1,
// objective completed → -1. All capped to 1-10 by UpdateTension.
func (s *Server) autoUpdateTension(ctx context.Context, sessionID int64, gmText string, rollFailed bool) {
    delta := 0
    if rollFailed {
        delta++
    }
    // Simple keyword scan for crisis signals
    crisisWords := []string{"ambush", "surrounded", "dying", "collapse", "betrayal", "captured", "overwhelmed"}
    lower := strings.ToLower(gmText)
    for _, word := range crisisWords {
        if strings.Contains(lower, word) {
            delta++
            break // only +1 regardless of how many crisis words appear
        }
    }
    if delta == 0 {
        return
    }
    current, err := s.db.GetTension(sessionID)
    if err != nil {
        return
    }
    newLevel := current + delta
    if err := s.db.UpdateTension(sessionID, newLevel); err != nil {
        return
    }
    clamped, _ := s.db.GetTension(sessionID)
    s.bus.Publish(Event{Type: EventTensionUpdated, Payload: map[string]any{
        "session_id":    sessionID,
        "tension_level": clamped,
    }})
}
```

Then in `handleGMRespondStream`, after the existing goroutine launches (line ~963), add:

```go
rollFailed := roll != nil && !roll.Success
go s.autoUpdateTension(context.Background(), id, fullText, rollFailed)
```

- [ ] **Step 6: Run test**

Run: `go test ./internal/api/... -run "TestRelationshipsCRUD" -v`
Expected: PASS

- [ ] **Step 7: Run full suite**

Run: `make test`
Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add internal/api/routes_phase_d.go internal/api/server.go internal/api/routes_phase_d_test.go internal/api/routes.go
git commit -m "feat(api): relationships CRUD routes + autoUpdateTension goroutine"
```

---

### Task D7: Frontend — Oracle panel, tension gauge, relationships panel

**Files:**
- Create: `web/src/OraclePanel.tsx`
- Create: `web/src/RelationshipsPanel.tsx`
- Modify: `web/src/types.ts`
- Modify: `web/src/App.tsx` (or wherever panels are registered/tab-listed)

- [ ] **Step 1: Add types**

In `web/src/types.ts`, add:

```typescript
export interface OracleResult {
  action: string;
  theme: string;
  action_roll: number;
  theme_roll: number;
  session_id: number;
}

export interface Relationship {
  id: number;
  campaign_id: number;
  entity_a: string;
  entity_b: string;
  score: number;         // -10 to +10
  disposition: string;  // hostile, cold, neutral, warm, allied
  notes: string;
  created_at: string;
}
```

- [ ] **Step 2: Add API functions to api.ts**

```typescript
export async function rollOracle(sessionId: number): Promise<OracleResult> {
  const res = await fetch(`/api/sessions/${sessionId}/oracle-roll`, { method: 'POST' });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function getTension(sessionId: number): Promise<{ session_id: number; tension_level: number }> {
  const res = await fetch(`/api/sessions/${sessionId}/tension`);
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function setTension(sessionId: number, level: number): Promise<void> {
  const res = await fetch(`/api/sessions/${sessionId}/tension`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ tension_level: level }),
  });
  if (!res.ok) throw new Error(await res.text());
}

export async function listRelationships(campaignId: number): Promise<Relationship[]> {
  const res = await fetch(`/api/campaigns/${campaignId}/relationships`);
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function createRelationship(campaignId: number, data: {
  entity_a: string; entity_b: string; score: number; disposition: string; notes: string;
}): Promise<Relationship> {
  const res = await fetch(`/api/campaigns/${campaignId}/relationships`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function updateRelationship(id: number, data: {
  score: number; disposition: string; notes: string;
}): Promise<void> {
  const res = await fetch(`/api/relationships/${id}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error(await res.text());
}

export async function deleteRelationship(id: number): Promise<void> {
  const res = await fetch(`/api/relationships/${id}`, { method: 'DELETE' });
  if (!res.ok) throw new Error(await res.text());
}
```

- [ ] **Step 3: Create OraclePanel.tsx**

Create `web/src/OraclePanel.tsx`:

```tsx
import React from 'react';
import { rollOracle, getTension, setTension } from './api';
import type { OracleResult } from './types';

interface Props {
  sessionId: number | null;
  lastEvent: unknown;
}

const TENSION_COLORS = ['#4CAF50','#8BC34A','#CDDC39','#FFEB3B','#FFC107','#FF9800','#FF5722','#F44336','#E91E63','#9C27B0'];
const TENSION_LABELS = ['1','2','3','4','5','6','7','8','9','10'];

export default function OraclePanel({ sessionId, lastEvent }: Props) {
  const [rolling, setRolling] = React.useState(false);
  const [result, setResult] = React.useState<OracleResult | null>(null);
  const [history, setHistory] = React.useState<OracleResult[]>([]);
  const [tension, setTensionState] = React.useState(5);

  React.useEffect(() => {
    if (!sessionId) return;
    getTension(sessionId).then(t => setTensionState(t.tension_level)).catch(() => {});
  }, [sessionId]);

  React.useEffect(() => {
    if (!lastEvent || !sessionId) return;
    const ev = lastEvent as { type: string; payload: { tension_level?: number } };
    if (ev.type === 'tension_updated' && ev.payload.tension_level !== undefined) {
      setTensionState(ev.payload.tension_level);
    }
  }, [lastEvent, sessionId]);

  const handleRoll = async () => {
    if (!sessionId) return;
    setRolling(true);
    try {
      const r = await rollOracle(sessionId);
      setResult(r);
      setHistory(prev => [r, ...prev].slice(0, 10));
    } catch (e) {
      console.error(e);
    } finally {
      setRolling(false);
    }
  };

  const handleTensionChange = async (level: number) => {
    if (!sessionId) return;
    setTensionState(level);
    await setTension(sessionId, level).catch(() => {});
  };

  if (!sessionId) return <div className="panel-empty">No active session</div>;

  return (
    <div className="oracle-panel">
      <div className="tension-section">
        <div className="tension-label">
          Tension <span style={{ color: TENSION_COLORS[tension - 1] }}>{tension}/10</span>
        </div>
        <div className="tension-track">
          {TENSION_LABELS.map((_, i) => (
            <button
              key={i}
              className={`tension-pip ${tension > i ? 'active' : ''}`}
              style={{ background: tension > i ? TENSION_COLORS[i] : undefined }}
              onClick={() => handleTensionChange(i + 1)}
              title={`Set tension to ${i + 1}`}
            />
          ))}
        </div>
      </div>

      <div className="oracle-section">
        <h3>Oracle</h3>
        <button className="oracle-roll-btn" onClick={handleRoll} disabled={rolling}>
          {rolling ? 'Rolling…' : '🎲 Roll Oracle'}
        </button>
        {result && (
          <div className="oracle-result">
            <span className="oracle-word action">{result.action}</span>
            <span className="oracle-plus"> + </span>
            <span className="oracle-word theme">{result.theme}</span>
            <div className="oracle-rolls">({result.action_roll}, {result.theme_roll})</div>
          </div>
        )}
        {history.length > 1 && (
          <div className="oracle-history">
            <div className="oracle-history-label">Recent</div>
            {history.slice(1).map((h, i) => (
              <div key={i} className="oracle-history-item">
                {h.action} + {h.theme}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
```

Add CSS for OraclePanel (in the relevant CSS file):

```css
.oracle-panel { padding: 12px; }
.tension-section { margin-bottom: 16px; }
.tension-label { font-size: 0.85rem; color: var(--text-muted, #888); margin-bottom: 6px; }
.tension-track { display: flex; gap: 4px; }
.tension-pip {
  width: 20px; height: 20px; border-radius: 3px;
  border: 1px solid var(--border, #444); background: var(--bg-input, #1a1a1a);
  cursor: pointer; padding: 0;
}
.tension-pip.active { border-color: transparent; }
.oracle-section h3 { font-size: 0.9rem; color: var(--gold, #c9a84c); margin: 0 0 8px; }
.oracle-roll-btn {
  background: rgba(201,168,76,0.1); border: 1px solid var(--gold, #c9a84c);
  color: var(--gold, #c9a84c); padding: 6px 14px; border-radius: 3px;
  cursor: pointer; font-size: 0.85rem;
}
.oracle-roll-btn:disabled { opacity: 0.5; }
.oracle-result { margin: 10px 0; font-size: 1.1rem; text-align: center; }
.oracle-word.action { color: #7ec8e3; font-weight: 600; }
.oracle-word.theme { color: #e8c99a; font-weight: 600; }
.oracle-rolls { font-size: 0.7rem; color: var(--text-muted, #888); }
.oracle-history { margin-top: 8px; font-size: 0.75rem; color: var(--text-muted, #888); }
.oracle-history-label { font-weight: 600; margin-bottom: 4px; }
.oracle-history-item { padding: 2px 0; }
```

- [ ] **Step 4: Create RelationshipsPanel.tsx**

Create `web/src/RelationshipsPanel.tsx`:

```tsx
import React from 'react';
import { listRelationships, createRelationship, updateRelationship, deleteRelationship } from './api';
import type { Relationship } from './types';

interface Props {
  campaignId: number | null;
  lastEvent: unknown;
}

const DISPOSITIONS = ['hostile', 'cold', 'neutral', 'warm', 'allied'];
const SCORE_COLORS: Record<string, string> = {
  hostile: '#F44336', cold: '#FF9800', neutral: '#9E9E9E', warm: '#4CAF50', allied: '#2196F3'
};

export default function RelationshipsPanel({ campaignId, lastEvent }: Props) {
  const [relationships, setRelationships] = React.useState<Relationship[]>([]);
  const [adding, setAdding] = React.useState(false);
  const [form, setForm] = React.useState({ entity_a: '', entity_b: '', score: 0, disposition: 'neutral', notes: '' });

  const load = React.useCallback(() => {
    if (!campaignId) return;
    listRelationships(campaignId).then(setRelationships).catch(() => {});
  }, [campaignId]);

  React.useEffect(() => { load(); }, [load]);

  React.useEffect(() => {
    if (!lastEvent) return;
    const ev = lastEvent as { type: string };
    if (ev.type === 'relationship_updated') load();
  }, [lastEvent, load]);

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!campaignId || !form.entity_a || !form.entity_b) return;
    await createRelationship(campaignId, form);
    setForm({ entity_a: '', entity_b: '', score: 0, disposition: 'neutral', notes: '' });
    setAdding(false);
    load();
  };

  const handleScoreChange = async (rel: Relationship, delta: number) => {
    const newScore = Math.max(-10, Math.min(10, rel.score + delta));
    const disposition = scoreToDisposition(newScore);
    await updateRelationship(rel.id, { score: newScore, disposition, notes: rel.notes });
    load();
  };

  const handleDelete = async (id: number) => {
    await deleteRelationship(id);
    load();
  };

  if (!campaignId) return <div className="panel-empty">No active campaign</div>;

  return (
    <div className="relationships-panel">
      <div className="relationships-header">
        <h3>Factions & Relationships</h3>
        <button className="add-btn" onClick={() => setAdding(!adding)}>
          {adding ? '✕' : '+ Add'}
        </button>
      </div>

      {adding && (
        <form onSubmit={handleAdd} className="relationship-form">
          <input placeholder="Entity A" value={form.entity_a} onChange={e => setForm(p => ({ ...p, entity_a: e.target.value }))} required />
          <input placeholder="Entity B" value={form.entity_b} onChange={e => setForm(p => ({ ...p, entity_b: e.target.value }))} required />
          <select value={form.disposition} onChange={e => setForm(p => ({ ...p, disposition: e.target.value }))}>
            {DISPOSITIONS.map(d => <option key={d} value={d}>{d}</option>)}
          </select>
          <input placeholder="Notes" value={form.notes} onChange={e => setForm(p => ({ ...p, notes: e.target.value }))} />
          <button type="submit">Add</button>
        </form>
      )}

      <div className="relationships-list">
        {relationships.length === 0 && <div className="panel-empty">No relationships tracked</div>}
        {relationships.map(rel => (
          <div key={rel.id} className="relationship-card">
            <div className="rel-entities">
              <span>{rel.entity_a}</span>
              <span className="rel-disposition" style={{ color: SCORE_COLORS[rel.disposition] }}>
                {rel.disposition}
              </span>
              <span>{rel.entity_b}</span>
            </div>
            <div className="rel-score-row">
              <button className="score-btn" onClick={() => handleScoreChange(rel, -1)}>−</button>
              <div className="rel-score-bar">
                <div
                  className="rel-score-fill"
                  style={{
                    width: `${((rel.score + 10) / 20) * 100}%`,
                    background: SCORE_COLORS[rel.disposition],
                  }}
                />
              </div>
              <button className="score-btn" onClick={() => handleScoreChange(rel, +1)}>+</button>
              <span className="rel-score-num">{rel.score > 0 ? '+' : ''}{rel.score}</span>
              <button className="del-btn" onClick={() => handleDelete(rel.id)}>✕</button>
            </div>
            {rel.notes && <div className="rel-notes">{rel.notes}</div>}
          </div>
        ))}
      </div>
    </div>
  );
}

function scoreToDisposition(score: number): string {
  if (score <= -6) return 'hostile';
  if (score <= -2) return 'cold';
  if (score <= 2) return 'neutral';
  if (score <= 6) return 'warm';
  return 'allied';
}
```

Add CSS:

```css
.relationships-panel { padding: 12px; }
.relationships-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
.relationships-header h3 { margin: 0; font-size: 0.9rem; color: var(--gold, #c9a84c); }
.add-btn {
  background: rgba(201,168,76,0.1); border: 1px solid var(--gold, #c9a84c);
  color: var(--gold, #c9a84c); padding: 3px 8px; border-radius: 3px; cursor: pointer; font-size: 0.8rem;
}
.relationship-form { display: flex; flex-direction: column; gap: 5px; margin-bottom: 10px; }
.relationship-form input, .relationship-form select {
  background: var(--bg-input, #1a1a1a); border: 1px solid var(--border, #333);
  color: var(--text, #e0d5b7); padding: 4px 6px; border-radius: 3px; font-size: 0.8rem;
}
.relationship-card { margin-bottom: 8px; padding: 8px; background: rgba(0,0,0,0.2); border-radius: 4px; }
.rel-entities { display: flex; justify-content: space-between; font-size: 0.85rem; margin-bottom: 5px; }
.rel-disposition { font-size: 0.75rem; font-style: italic; }
.rel-score-row { display: flex; align-items: center; gap: 5px; }
.score-btn {
  width: 20px; height: 20px; border-radius: 3px;
  background: rgba(255,255,255,0.05); border: 1px solid var(--border, #444);
  color: var(--text, #e0d5b7); cursor: pointer; font-size: 0.85rem; padding: 0;
}
.rel-score-bar { flex: 1; height: 6px; background: var(--border, #333); border-radius: 3px; overflow: hidden; }
.rel-score-fill { height: 100%; border-radius: 3px; transition: width 0.2s; }
.rel-score-num { font-size: 0.75rem; color: var(--text-muted, #888); width: 24px; text-align: right; }
.del-btn { background: none; border: none; color: #666; cursor: pointer; font-size: 0.75rem; }
.del-btn:hover { color: #F44336; }
.rel-notes { font-size: 0.75rem; color: var(--text-muted, #888); margin-top: 4px; }
```

- [ ] **Step 5: Register panels in App.tsx**

In `web/src/App.tsx`, import the new panels and add them to the panel tab list (follow the existing pattern for how NPCsPanel, ObjectivesPanel, etc. are registered — look at the current tab array or panel switch statement and add Oracle and Relationships in the same way).

- [ ] **Step 6: Build + test**

Run: `make test && make build`
Expected: PASS + no TypeScript errors

- [ ] **Step 7: Commit**

```bash
git add web/src/OraclePanel.tsx web/src/RelationshipsPanel.tsx web/src/types.ts web/src/api.ts web/src/App.tsx
git commit -m "feat(ui): Oracle + Tension panel, Relationships faction tracker"
```
