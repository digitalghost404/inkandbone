# Button Theme, Inventory Currency & Favicon Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix unstyled Reanalyze buttons, replace the Vite favicon with a grimoire-themed quill & inkwell, and add AI-managed currency tracking to the inventory panel with a 5-second undo toast.

**Architecture:** DB migration adds two columns to `characters`. A new `autoUpdateCurrency` goroutine (same pattern as `autoExtractItems`) extracts explicit currency deltas from GM text and publishes an enriched `character_updated` event. The frontend inventory panel renders a pinned currency row, inline editing, and an undo toast driven by that event.

**Tech Stack:** Go 1.22, SQLite (mattn/go-sqlite3), React 18 + TypeScript, existing `ai.Completer` interface, WebSocket event bus.

---

## File Map

| File | Action | Responsibility |
|------|--------|----------------|
| `internal/db/migrations/015_currency.sql` | Create | Add `currency_balance` + `currency_label` to characters |
| `internal/db/queries_core.go` | Modify | Update `Character` struct, `GetCharacter`, `ListCharacters`, add `UpdateCharacterCurrencyBalance` + `UpdateCharacterCurrencyLabel` |
| `internal/api/routes_character.go` | Modify | Extend `handlePatchCharacter` to accept currency fields |
| `internal/api/routes_character_test.go` | Modify | Add tests for currency patch path |
| `internal/api/routes.go` | Modify | Add `autoUpdateCurrency` goroutine, wire into `handleGMRespondStream` |
| `web/src/types.ts` | Modify | Add `currency_balance` + `currency_label` to `Character` interface |
| `web/src/api.ts` | Modify | Add `patchCurrency` function |
| `web/src/InventoryPanel.tsx` | Modify | Currency row, inline edit, undo toast |
| `web/src/App.css` | Modify | Add `.reanalyze-btn` + currency row + toast styles |
| `web/src/NPCRosterPanel.tsx` | Modify | Rename `npc-reanalyze-btn` → `reanalyze-btn` |
| `web/src/ObjectivesPanel.tsx` | Modify | Rename `npc-reanalyze-btn` → `reanalyze-btn` |
| `web/public/favicon.svg` | Replace | Quill & inkwell SVG |

---

## Task 1: DB Migration

**Files:**
- Create: `internal/db/migrations/015_currency.sql`

- [ ] **Step 1: Write the migration**

```sql
-- 015_currency.sql
ALTER TABLE characters ADD COLUMN currency_balance INTEGER NOT NULL DEFAULT 0;
ALTER TABLE characters ADD COLUMN currency_label TEXT NOT NULL DEFAULT 'Gold';
```

- [ ] **Step 2: Verify it applies**

Run: `cd /home/digitalghost/projects/inkandbone && go test ./internal/db/... -v -run TestOpen 2>&1 | tail -5`
Expected: PASS (migrations run on `db.Open`)

- [ ] **Step 3: Commit**

```bash
git add internal/db/migrations/015_currency.sql
git commit -m "feat: add currency_balance and currency_label to characters"
```

---

## Task 2: DB Layer — Character Struct & Queries

**Files:**
- Modify: `internal/db/queries_core.go`

- [ ] **Step 1: Write a failing test**

Add to a new file `internal/db/queries_core_test.go` (or append to the closest existing test file):

```go
func TestCharacterCurrency(t *testing.T) {
    d, err := Open(":memory:")
    require.NoError(t, err)
    defer d.Close()

    rsID, err := d.CreateRuleset("test", "{}", "1")
    require.NoError(t, err)
    campID, err := d.CreateCampaign(rsID, "Camp", "")
    require.NoError(t, err)
    charID, err := d.CreateCharacter(campID, "Hero")
    require.NoError(t, err)

    // Defaults
    c, err := d.GetCharacter(charID)
    require.NoError(t, err)
    assert.Equal(t, int64(0), c.CurrencyBalance)
    assert.Equal(t, "Gold", c.CurrencyLabel)

    // Update balance
    err = d.UpdateCharacterCurrencyBalance(charID, 50)
    require.NoError(t, err)
    c, err = d.GetCharacter(charID)
    require.NoError(t, err)
    assert.Equal(t, int64(50), c.CurrencyBalance)

    // Update label
    err = d.UpdateCharacterCurrencyLabel(charID, "Coin")
    require.NoError(t, err)
    c, err = d.GetCharacter(charID)
    require.NoError(t, err)
    assert.Equal(t, "Coin", c.CurrencyLabel)

    // ListCharacters includes currency
    chars, err := d.ListCharacters(campID)
    require.NoError(t, err)
    require.Len(t, chars, 1)
    assert.Equal(t, int64(50), chars[0].CurrencyBalance)
    assert.Equal(t, "Coin", chars[0].CurrencyLabel)
}
```

- [ ] **Step 2: Run to confirm it fails**

Run: `go test ./internal/db/... -v -run TestCharacterCurrency`
Expected: FAIL — `c.CurrencyBalance` undefined

- [ ] **Step 3: Update `Character` struct**

In `internal/db/queries_core.go`, replace the `Character` struct (lines 136-143):

```go
type Character struct {
	ID              int64  `json:"id"`
	CampaignID      int64  `json:"campaign_id"`
	Name            string `json:"name"`
	DataJSON        string `json:"data_json"`
	PortraitPath    string `json:"portrait_path"`
	CurrencyBalance int64  `json:"currency_balance"`
	CurrencyLabel   string `json:"currency_label"`
	CreatedAt       string `json:"created_at"`
}
```

- [ ] **Step 4: Update `GetCharacter`**

Replace the `GetCharacter` function body:

```go
func (d *DB) GetCharacter(id int64) (*Character, error) {
	c := &Character{}
	err := d.db.QueryRow(
		"SELECT id, campaign_id, name, data_json, portrait_path, currency_balance, currency_label, created_at FROM characters WHERE id = ?", id,
	).Scan(&c.ID, &c.CampaignID, &c.Name, &c.DataJSON, &c.PortraitPath, &c.CurrencyBalance, &c.CurrencyLabel, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return c, err
}
```

- [ ] **Step 5: Update `ListCharacters`**

Replace the query and scan in `ListCharacters`:

```go
func (d *DB) ListCharacters(campaignID int64) ([]Character, error) {
	rows, err := d.db.Query(
		"SELECT id, campaign_id, name, data_json, portrait_path, currency_balance, currency_label, created_at FROM characters WHERE campaign_id = ? ORDER BY name",
		campaignID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Character
	for rows.Next() {
		var c Character
		if err := rows.Scan(&c.ID, &c.CampaignID, &c.Name, &c.DataJSON, &c.PortraitPath, &c.CurrencyBalance, &c.CurrencyLabel, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
```

- [ ] **Step 6: Add `UpdateCharacterCurrencyBalance` and `UpdateCharacterCurrencyLabel`**

Add after `UpdateCharacterPortrait` in `internal/db/queries_core.go`:

```go
func (d *DB) UpdateCharacterCurrencyBalance(id int64, balance int64) error {
	res, err := d.db.Exec("UPDATE characters SET currency_balance = ? WHERE id = ?", balance, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("character %d not found", id)
	}
	return nil
}

func (d *DB) UpdateCharacterCurrencyLabel(id int64, label string) error {
	res, err := d.db.Exec("UPDATE characters SET currency_label = ? WHERE id = ?", label, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("character %d not found", id)
	}
	return nil
}
```

- [ ] **Step 7: Run test to confirm it passes**

Run: `go test ./internal/db/... -v -run TestCharacterCurrency`
Expected: PASS

- [ ] **Step 8: Run full DB test suite**

Run: `go test ./internal/db/... -v`
Expected: all PASS

- [ ] **Step 9: Commit**

```bash
git add internal/db/queries_core.go
git commit -m "feat: add currency fields to Character struct and queries"
```

---

## Task 3: API — Extend PATCH /characters/{id}

**Files:**
- Modify: `internal/api/routes_character.go`
- Modify: `internal/api/routes_character_test.go`

- [ ] **Step 1: Write failing tests**

Append to `internal/api/routes_character_test.go`:

```go
func TestPatchCharacter_currency(t *testing.T) {
	s := newTestServer(t)
	campID, _ := seedCampaign(t, s.db)
	charID, err := s.db.CreateCharacter(campID, "Kael")
	require.NoError(t, err)

	ch := s.bus.Subscribe()

	body := `{"currency_balance":75,"currency_label":"Coin"}`
	req := httptest.NewRequest(http.MethodPatch,
		fmt.Sprintf("/api/characters/%d", charID),
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)

	char, err := s.db.GetCharacter(charID)
	require.NoError(t, err)
	assert.Equal(t, int64(75), char.CurrencyBalance)
	assert.Equal(t, "Coin", char.CurrencyLabel)

	var got Event
	select {
	case got = <-ch:
	default:
		t.Fatal("expected character_updated event")
	}
	assert.Equal(t, EventCharacterUpdated, got.Type)
	payload := got.Payload.(map[string]any)
	assert.Equal(t, charID, payload["id"])
}

func TestPatchCharacter_currencyBalanceOnly(t *testing.T) {
	s := newTestServer(t)
	campID, _ := seedCampaign(t, s.db)
	charID, err := s.db.CreateCharacter(campID, "Kael")
	require.NoError(t, err)

	body := `{"currency_balance":30}`
	req := httptest.NewRequest(http.MethodPatch,
		fmt.Sprintf("/api/characters/%d", charID),
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)

	char, err := s.db.GetCharacter(charID)
	require.NoError(t, err)
	assert.Equal(t, int64(30), char.CurrencyBalance)
	assert.Equal(t, "Gold", char.CurrencyLabel) // label unchanged
}
```

- [ ] **Step 2: Run to confirm they fail**

Run: `go test ./internal/api/... -v -run TestPatchCharacter_currency`
Expected: FAIL

- [ ] **Step 3: Update `handlePatchCharacter`**

Replace `handlePatchCharacter` in `internal/api/routes_character.go`:

```go
func (s *Server) handlePatchCharacter(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePathID(r, "id")
	if !ok {
		http.Error(w, "invalid character id", http.StatusBadRequest)
		return
	}
	var body struct {
		DataJSON        *string `json:"data_json"`
		CurrencyBalance *int64  `json:"currency_balance"`
		CurrencyLabel   *string `json:"currency_label"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if body.DataJSON != nil {
		if err := s.db.UpdateCharacterData(id, *body.DataJSON); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if body.CurrencyBalance != nil {
		balance := *body.CurrencyBalance
		if balance < 0 {
			balance = 0
		}
		if err := s.db.UpdateCharacterCurrencyBalance(id, balance); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if body.CurrencyLabel != nil {
		if err := s.db.UpdateCharacterCurrencyLabel(id, *body.CurrencyLabel); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	s.bus.Publish(Event{Type: EventCharacterUpdated, Payload: map[string]any{"id": id}})
	w.WriteHeader(http.StatusNoContent)
}
```

- [ ] **Step 4: Run tests to confirm they pass**

Run: `go test ./internal/api/... -v -run TestPatchCharacter`
Expected: all TestPatchCharacter* PASS

- [ ] **Step 5: Commit**

```bash
git add internal/api/routes_character.go internal/api/routes_character_test.go
git commit -m "feat: extend PATCH /characters/{id} to accept currency fields"
```

---

## Task 4: autoUpdateCurrency Goroutine

**Files:**
- Modify: `internal/api/routes.go`

- [ ] **Step 1: Add the goroutine**

Add the following function at the end of `internal/api/routes.go` (after `autoExtractItems`):

```go
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
```

- [ ] **Step 2: Wire into `handleGMRespondStream`**

In `internal/api/routes.go`, find the block that launches goroutines after GM response (around line 1073-1084). Add one new line after `autoExtractItems`:

```go
go s.autoExtractItems(context.Background(), id, fullText)
go s.autoUpdateCurrency(context.Background(), id, fullText)
```

- [ ] **Step 3: Build to check compilation**

Run: `cd /home/digitalghost/projects/inkandbone && go build ./...`
Expected: no errors

- [ ] **Step 4: Run full API test suite**

Run: `go test ./internal/api/... -v 2>&1 | tail -20`
Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/api/routes.go
git commit -m "feat: add autoUpdateCurrency goroutine with undo-toast event payload"
```

---

## Task 5: Frontend Types & API

**Files:**
- Modify: `web/src/types.ts`
- Modify: `web/src/api.ts`

- [ ] **Step 1: Update `Character` interface in `types.ts`**

In `web/src/types.ts`, replace the `Character` interface:

```typescript
export interface Character {
  id: number
  campaign_id: number
  name: string
  data_json: string
  portrait_path: string
  currency_balance: number
  currency_label: string
  created_at: string
}
```

- [ ] **Step 2: Add `patchCurrency` to `api.ts`**

Append to `web/src/api.ts`:

```typescript
export async function patchCurrency(
  characterId: number,
  updates: { currency_balance?: number; currency_label?: string },
): Promise<void> {
  const res = await fetch(`/api/characters/${characterId}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(updates),
  })
  if (!res.ok) throw new Error(`patchCurrency failed: ${res.status}`)
}
```

- [ ] **Step 3: Build TypeScript**

Run: `cd /home/digitalghost/projects/inkandbone/web && npm run build 2>&1 | tail -10`
Expected: no TypeScript errors

- [ ] **Step 4: Commit**

```bash
git add web/src/types.ts web/src/api.ts
git commit -m "feat: add currency fields to Character type and patchCurrency api"
```

---

## Task 6: InventoryPanel — Currency Row & Undo Toast

**Files:**
- Modify: `web/src/InventoryPanel.tsx`
- Modify: `web/src/App.css`

- [ ] **Step 1: Add CSS for currency row and toast**

In `web/src/App.css`, find the `/* ── Management Panel` comment and insert before it:

```css
/* ── Currency Row ────────────────────────────────────────── */

.currency-row {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.3rem 0 0.4rem;
  border-bottom: 1px solid var(--border);
  margin-bottom: 0.4rem;
}

.currency-icon { color: var(--gold); font-size: 12px; flex-shrink: 0; }

.currency-label-text {
  font-family: var(--serif);
  font-size: 11px;
  letter-spacing: 1px;
  color: var(--text-dim);
  cursor: pointer;
  transition: color 0.12s;
}
.currency-label-text:hover { color: var(--gold); }

.currency-value {
  font-family: var(--serif);
  font-size: 14px;
  color: var(--gold);
  margin-left: auto;
  cursor: pointer;
  transition: color 0.12s;
}
.currency-value:hover { color: var(--text); }

.currency-edit-btn {
  background: none;
  border: none;
  color: var(--text-dim);
  font-size: 10px;
  cursor: pointer;
  padding: 0 0.1rem;
  transition: color 0.12s;
  flex-shrink: 0;
}
.currency-edit-btn:hover { color: var(--gold); }

.currency-input {
  background: none;
  border: none;
  border-bottom: 1px solid var(--gold-dim);
  color: var(--gold);
  font-family: var(--serif);
  font-size: 14px;
  outline: none;
  width: 70px;
  text-align: right;
}

.currency-label-input {
  background: none;
  border: none;
  border-bottom: 1px solid var(--gold-dim);
  color: var(--text-dim);
  font-family: var(--serif);
  font-size: 11px;
  letter-spacing: 1px;
  outline: none;
  width: 60px;
}

/* ── Currency Undo Toast ─────────────────────────────────── */

.currency-toast {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--surface);
  border: 1px solid var(--gold-dim);
  border-radius: 2px;
  padding: 0.3rem 0.6rem;
  margin-bottom: 0.4rem;
  font-family: var(--serif);
  font-size: 11px;
  color: var(--gold);
  letter-spacing: 0.5px;
  animation: toast-in 0.15s ease;
}

@keyframes toast-in {
  from { opacity: 0; transform: translateY(-4px); }
  to   { opacity: 1; transform: translateY(0); }
}

.currency-toast-undo {
  background: none;
  border: 1px solid var(--border);
  border-radius: 2px;
  color: var(--text-dim);
  font-family: var(--serif);
  font-size: 10px;
  letter-spacing: 1px;
  padding: 0.1rem 0.4rem;
  cursor: pointer;
  transition: color 0.12s, border-color 0.12s;
}
.currency-toast-undo:hover { color: var(--gold); border-color: var(--gold-dim); }
```

- [ ] **Step 2: Rewrite `InventoryPanel.tsx`**

Replace the entire contents of `web/src/InventoryPanel.tsx`:

```typescript
import { useState, useEffect, useRef } from 'react'
import { fetchItems, createItem, patchItem, deleteItem, patchCurrency } from './api'
import type { Item } from './types'

interface InventoryPanelProps {
  characterId: number | null
  characterCurrencyBalance: number
  characterCurrencyLabel: string
  lastEvent: unknown
}

interface CurrencyToast {
  delta: number
  label: string
  prevBalance: number
  newBalance: number
}

export function InventoryPanel({
  characterId,
  characterCurrencyBalance,
  characterCurrencyLabel,
  lastEvent,
}: InventoryPanelProps) {
  const [items, setItems] = useState<Item[]>([])
  const [addName, setAddName] = useState('')
  const [saving, setSaving] = useState(false)

  // Currency state (mirrors props but allows inline edit)
  const [balance, setBalance] = useState(characterCurrencyBalance)
  const [label, setLabel] = useState(characterCurrencyLabel)
  const [editingBalance, setEditingBalance] = useState(false)
  const [editingLabel, setEditingLabel] = useState(false)
  const [balanceDraft, setBalanceDraft] = useState('')
  const [labelDraft, setLabelDraft] = useState('')

  // Toast
  const [toast, setToast] = useState<CurrencyToast | null>(null)
  const toastTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Sync currency from parent when not editing
  useEffect(() => {
    if (!editingBalance) setBalance(characterCurrencyBalance)
  }, [characterCurrencyBalance, editingBalance])

  useEffect(() => {
    if (!editingLabel) setLabel(characterCurrencyLabel)
  }, [characterCurrencyLabel, editingLabel])

  useEffect(() => {
    if (characterId === null) return
    fetchItems(characterId).then(setItems).catch(() => setItems([]))
  }, [characterId])

  useEffect(() => {
    const ev = lastEvent as { type?: string; payload?: Record<string, unknown> } | null
    if (!ev) return

    if (ev.type === 'item_updated' && characterId !== null) {
      fetchItems(characterId).then(setItems).catch(() => {})
    }

    if (ev.type === 'character_updated') {
      const p = ev.payload
      if (p && typeof p.currency_delta === 'number' && p.currency_delta !== 0) {
        const delta = p.currency_delta as number
        const newBal = p.currency_balance as number
        const lbl = (p.currency_label as string) ?? label

        // Clear any existing timer
        if (toastTimer.current) clearTimeout(toastTimer.current)

        setToast({
          delta,
          label: lbl,
          prevBalance: newBal - delta,
          newBalance: newBal,
        })
        setBalance(newBal)
        setLabel(lbl)

        toastTimer.current = setTimeout(() => setToast(null), 5000)
      } else if (p && typeof p.currency_balance === 'number') {
        // Manual patch (no delta) — just sync silently
        setBalance(p.currency_balance as number)
        if (typeof p.currency_label === 'string') setLabel(p.currency_label as string)
      }
    }
  }, [lastEvent, characterId, label])

  async function handleUndoToast() {
    if (!toast || characterId === null) return
    if (toastTimer.current) clearTimeout(toastTimer.current)
    setToast(null)
    setBalance(toast.prevBalance)
    try {
      await patchCurrency(characterId, { currency_balance: toast.prevBalance })
    } catch (err) {
      console.error(err)
    }
  }

  async function handleBalanceSave() {
    const parsed = parseInt(balanceDraft, 10)
    if (isNaN(parsed) || characterId === null) {
      setEditingBalance(false)
      return
    }
    const clamped = Math.max(0, parsed)
    setBalance(clamped)
    setEditingBalance(false)
    try {
      await patchCurrency(characterId, { currency_balance: clamped })
    } catch (err) {
      console.error(err)
    }
  }

  async function handleLabelSave() {
    const trimmed = labelDraft.trim()
    if (!trimmed || characterId === null) {
      setEditingLabel(false)
      return
    }
    setLabel(trimmed)
    setEditingLabel(false)
    try {
      await patchCurrency(characterId, { currency_label: trimmed })
    } catch (err) {
      console.error(err)
    }
  }

  async function handleAdd() {
    if (!addName.trim() || characterId === null) return
    setSaving(true)
    try {
      const item = await createItem(characterId, addName.trim(), '', 1)
      setItems((prev) => [...prev, item])
      setAddName('')
    } catch (err) {
      console.error(err)
    } finally {
      setSaving(false)
    }
  }

  async function handleEquipToggle(item: Item) {
    try {
      await patchItem(item.id, { equipped: !item.equipped })
      setItems((prev) =>
        prev.map((i) => (i.id === item.id ? { ...i, equipped: !item.equipped } : i))
      )
    } catch (err) {
      console.error(err)
    }
  }

  async function handleDelete(id: number) {
    try {
      await deleteItem(id)
      setItems((prev) => prev.filter((i) => i.id !== id))
    } catch (err) {
      console.error(err)
    }
  }

  if (characterId === null) return null

  const sorted = [
    ...items.filter((i) => i.equipped),
    ...items.filter((i) => !i.equipped),
  ]

  const deltaSign = toast && toast.delta > 0 ? '+' : ''

  return (
    <div className="inventory-panel">
      <div className="inventory-panel-label">Inventory</div>

      {/* Currency row */}
      <div className="currency-row">
        <span className="currency-icon">◈</span>

        {editingLabel ? (
          <input
            className="currency-label-input"
            value={labelDraft}
            onChange={(e) => setLabelDraft(e.target.value)}
            onBlur={handleLabelSave}
            onKeyDown={(e) => {
              if (e.key === 'Enter') handleLabelSave()
              if (e.key === 'Escape') setEditingLabel(false)
            }}
            autoFocus
          />
        ) : (
          <span
            className="currency-label-text"
            onClick={() => { setLabelDraft(label); setEditingLabel(true) }}
            title="Click to rename"
          >
            {label}
          </span>
        )}

        {editingBalance ? (
          <input
            className="currency-input"
            value={balanceDraft}
            onChange={(e) => setBalanceDraft(e.target.value)}
            onBlur={handleBalanceSave}
            onKeyDown={(e) => {
              if (e.key === 'Enter') handleBalanceSave()
              if (e.key === 'Escape') setEditingBalance(false)
            }}
            autoFocus
          />
        ) : (
          <span
            className="currency-value"
            onClick={() => { setBalanceDraft(String(balance)); setEditingBalance(true) }}
            title="Click to correct"
          >
            {balance}
          </span>
        )}

        <button
          className="currency-edit-btn"
          title="Correct balance"
          onClick={() => { setBalanceDraft(String(balance)); setEditingBalance(true) }}
        >
          ✎
        </button>
      </div>

      {/* Undo toast */}
      {toast && (
        <div className="currency-toast">
          <span>AI: {deltaSign}{toast.delta} {toast.label}</span>
          <button className="currency-toast-undo" onClick={handleUndoToast}>Undo</button>
        </div>
      )}

      {sorted.length === 0 && (
        <p className="empty">No items yet.</p>
      )}

      {sorted.map((item) => (
        <div key={item.id} className={`inventory-item${item.equipped ? ' equipped' : ''}`}>
          <div className="inventory-item-row">
            {item.equipped && <span className="inventory-equipped-icon" title="Equipped">⚔</span>}
            <span className="inventory-item-name">{item.name}</span>
            {item.quantity > 1 && (
              <span className="inventory-item-qty">×{item.quantity}</span>
            )}
            <div className="inventory-actions">
              <button
                className={`inventory-equip-btn${item.equipped ? ' active' : ''}`}
                onClick={() => handleEquipToggle(item)}
                title={item.equipped ? 'Unequip' : 'Equip'}
              >
                {item.equipped ? '▣' : '□'}
              </button>
              <button
                className="inventory-delete-btn"
                onClick={() => handleDelete(item.id)}
                title="Remove item"
              >
                ×
              </button>
            </div>
          </div>
          {item.description && (
            <div className="inventory-item-desc">{item.description}</div>
          )}
        </div>
      ))}

      <div className="inventory-add-form">
        <input
          className="inventory-add-input"
          placeholder="Item name…"
          value={addName}
          onChange={(e) => setAddName(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter') handleAdd()
          }}
        />
        <button
          className="inventory-add-btn"
          onClick={handleAdd}
          disabled={saving || !addName.trim()}
        >
          {saving ? '…' : 'Add'}
        </button>
      </div>
    </div>
  )
}
```

- [ ] **Step 3: Update `App.tsx` to pass currency props to `InventoryPanel`**

Find where `InventoryPanel` is rendered in `web/src/App.tsx`. It currently receives `characterId` and `lastEvent`. Add the two new props from the active character state:

```tsx
<InventoryPanel
  characterId={activeCharacterId}
  characterCurrencyBalance={activeCharacter?.currency_balance ?? 0}
  characterCurrencyLabel={activeCharacter?.currency_label ?? 'Gold'}
  lastEvent={lastEvent}
/>
```

(Replace `activeCharacterId` and `activeCharacter` with whatever variable names `App.tsx` currently uses for the character — check `App.tsx` to verify the exact variable names before editing.)

- [ ] **Step 4: Build to check for TypeScript errors**

Run: `cd /home/digitalghost/projects/inkandbone/web && npm run build 2>&1 | tail -15`
Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add web/src/InventoryPanel.tsx web/src/App.css web/src/App.tsx
git commit -m "feat: add currency row and AI undo toast to InventoryPanel"
```

---

## Task 7: Reanalyze Button Theme

**Files:**
- Modify: `web/src/App.css`
- Modify: `web/src/NPCRosterPanel.tsx`
- Modify: `web/src/ObjectivesPanel.tsx`

- [ ] **Step 1: Add `.reanalyze-btn` CSS**

In `web/src/App.css`, find the `.npc-add-btn` block. Add before it:

```css
/* ── Reanalyze Button ────────────────────────────────────── */

.reanalyze-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-family: var(--serif);
  font-size: 11px;
  color: var(--text-dim);
  letter-spacing: 1px;
  padding: 0.3rem 0;
  width: 100%;
  text-align: left;
  transition: color 0.15s;
}
.reanalyze-btn:hover:not(:disabled) { color: var(--gold); }
.reanalyze-btn:disabled { opacity: 0.4; cursor: not-allowed; }
```

- [ ] **Step 2: Rename class in `NPCRosterPanel.tsx`**

In `web/src/NPCRosterPanel.tsx` line 89, change:

```tsx
className="npc-reanalyze-btn"
```
to:
```tsx
className="reanalyze-btn"
```

- [ ] **Step 3: Rename class in `ObjectivesPanel.tsx`**

In `web/src/ObjectivesPanel.tsx` line 136, change:

```tsx
className="npc-reanalyze-btn"
```
to:
```tsx
className="reanalyze-btn"
```

- [ ] **Step 4: Build to confirm no regressions**

Run: `cd /home/digitalghost/projects/inkandbone/web && npm run build 2>&1 | tail -5`
Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add web/src/App.css web/src/NPCRosterPanel.tsx web/src/ObjectivesPanel.tsx
git commit -m "fix: style reanalyze button with grimoire ghost button theme"
```

---

## Task 8: Favicon

**Files:**
- Replace: `web/public/favicon.svg`

- [ ] **Step 1: Replace the favicon**

Overwrite `web/public/favicon.svg` with the quill & inkwell SVG:

```svg
<svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 48 48" fill="none">
  <!-- Inkwell body -->
  <ellipse cx="16" cy="37" rx="8" ry="5" fill="#3a3020" stroke="#5a4a2a" stroke-width="1"/>
  <!-- Inkwell rim/opening -->
  <ellipse cx="16" cy="34" rx="6" ry="3" fill="#0f0e0a" stroke="#5a4a2a" stroke-width="0.8"/>
  <!-- Inkwell body side -->
  <path d="M10 35 L10 37 Q16 41 22 37 L22 35" fill="#1a1710" stroke="#5a4a2a" stroke-width="0.8"/>
  <!-- Quill shaft -->
  <line x1="21" y1="34" x2="40" y2="9" stroke="#c9a84c" stroke-width="1.5" stroke-linecap="round"/>
  <!-- Feather left barb -->
  <path d="M21 34 Q26 27 31 20" stroke="#8b7355" stroke-width="3.5" stroke-linecap="round" fill="none"/>
  <!-- Feather left barb highlight -->
  <path d="M23 31 Q27 25 32 18" stroke="#c9a84c" stroke-width="1" stroke-linecap="round" fill="none"/>
  <!-- Feather right barb -->
  <path d="M26 29 Q31 23 36 15" stroke="#8b7355" stroke-width="3.5" stroke-linecap="round" fill="none"/>
  <!-- Feather right barb highlight -->
  <path d="M28 26 Q33 20 38 12" stroke="#c9a84c" stroke-width="1" stroke-linecap="round" fill="none"/>
  <!-- Quill tip dot (nib) -->
  <circle cx="21" cy="34" r="1.5" fill="#c9a84c"/>
</svg>
```

- [ ] **Step 2: Verify in browser**

Run `make dev` or `cd web && npm run dev`, open the browser and confirm the tab shows the quill icon.

- [ ] **Step 3: Commit**

```bash
git add web/public/favicon.svg
git commit -m "chore: replace Vite lightning bolt favicon with grimoire quill & inkwell"
```

---

## Self-Review

**Spec coverage:**
- ✅ Section 1 (Reanalyze button) → Task 7
- ✅ Section 2 (Favicon) → Task 8
- ✅ Section 3 (DB migration) → Task 1
- ✅ Section 3 (API extend PATCH) → Task 3
- ✅ Section 3 (autoUpdateCurrency goroutine) → Task 4
- ✅ Section 3 (Option 2: undo toast) → Task 6
- ✅ Section 4 (Currency row pinned at top) → Task 6
- ✅ Section 4 (Inline edit for balance + label) → Task 6
- ✅ Section 4 (WebSocket character_updated triggers currency refresh) → Task 4 + Task 6
- ✅ Types + API layer → Task 5

**Placeholder scan:** No TBDs. All code blocks contain real implementations.

**Type consistency:**
- `CurrencyBalance int64` / `currency_balance: number` used consistently across all tasks
- `patchCurrency` defined in Task 5, used in Task 6 ✅
- `UpdateCharacterCurrencyBalance` / `UpdateCharacterCurrencyLabel` defined in Task 2, used in Tasks 3 and 4 ✅
- `CurrencyToast` interface defined and used within Task 6 only ✅
