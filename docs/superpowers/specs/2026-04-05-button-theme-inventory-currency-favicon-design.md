# Design: Button Theme, Inventory Currency, Favicon

**Date:** 2026-04-05  
**Status:** Approved

---

## Overview

Four related improvements to the ink & bone UI:

1. Fix unstyled Reanalyze buttons in NPCRosterPanel and ObjectivesPanel
2. Replace the default Vite lightning bolt favicon with a thematic quill & inkwell
3. Add AI-managed currency tracking to the character inventory
4. Show a 5-second undo toast when the AI updates currency

---

## Section 1 — Reanalyze Button Theme

### Problem

`npc-reanalyze-btn` is used in both `NPCRosterPanel.tsx` (line 89) and `ObjectivesPanel.tsx` (line 136) but has no CSS definition in `App.css`. The button renders as an unstyled browser default — wrong font, wrong color, no hover state.

### Solution

- Rename the class to `reanalyze-btn` (shared, not NPC-specific)
- Define it once in `App.css` as a ghost/borderless button matching `.map-generate-btn`
- Update both panel files to use `className="reanalyze-btn"`

### CSS

```css
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

---

## Section 2 — Favicon

### Problem

`web/public/favicon.svg` is the default Vite purple lightning bolt — no relation to the grimoire theme.

### Solution

Replace with a quill-and-inkwell SVG using grimoire color tokens:
- Inkwell: `#3a3020` body, `#5a4a2a` rim
- Quill shaft + tip: `#c9a84c` (gold)
- Feather barbs: `#8b7355` (text-dim) with gold overlay
- Background: transparent (browser fills with `#0f0e0a` on dark tabs)

The design must read clearly at 16×16 (browser tab size).

---

## Section 3 — Currency: Database & Backend

### Data Model

New migration (next increment after current highest) adds two columns to `characters`:

```sql
ALTER TABLE characters ADD COLUMN currency_balance INTEGER NOT NULL DEFAULT 0;
ALTER TABLE characters ADD COLUMN currency_label TEXT NOT NULL DEFAULT 'Gold';
```

### API

Extend existing `PATCH /api/characters/{id}` handler to accept and persist `currency_balance` and `currency_label` fields. No new routes.

### Goroutine: `autoUpdateCurrency`

Fires after every GM response, same pattern as `autoExtractItems`.

**Prompt:** Send GM response text to Claude with a focused extraction prompt:
> "Extract any explicit currency transaction from the following text. Return JSON: `{delta: number}` where delta is positive for gains and negative for costs. Only extract when a specific number AND a currency word appear together (e.g. '30 gold', '15 coin', '5 dollars'). If no explicit transaction exists, return `{delta: 0}`."

**Logic:**
1. Parse delta from response
2. If delta == 0, return early — no update, no toast
3. Apply: `new_balance = MAX(0, current_balance + delta)`
4. Persist via `UPDATE characters SET currency_balance = ? WHERE id = ?`
5. Broadcast `character_updated` WebSocket event with payload: `{ character_id, delta, label, new_balance }`

**Error handling:** Any parse failure or Claude error returns early with no update. Never corrupt the balance silently.

---

## Section 4 — Inventory Panel

### Currency Row

Pinned above all items. Layout: `◈  Gold  [balance]  ✎`

- `◈` — decorative gold icon (color: `var(--gold)`)
- Label ("Gold") — editable inline on click; saves on blur/Enter via PATCH
- Balance — editable inline on click; saves on blur/Enter via PATCH
- `✎` — small pencil button that triggers inline edit mode for manual correction

Inline edit style: borderless input, `border-bottom: 1px solid var(--gold-dim)`, gold text, serif font.

### Undo Toast

When the panel receives a `character_updated` WebSocket event with a non-zero `delta`:

- Show a small toast inside the inventory panel (not a global overlay)
- Content: `"AI: +30 Gold  [Undo]"` (or `"AI: −15 Gold  [Undo]"` for costs)
- Auto-dismisses after 5 seconds
- If user clicks Undo: PATCH `currency_balance` back to `new_balance - delta`, hide toast immediately

Toast styling matches grimoire theme: dark surface, gold text, serif font, subtle border.

### Item Actions

No changes to existing item behavior. The AI goroutine `autoExtractItems` already handles item removal/decrement from GM narrative. The `×` delete button remains the manual override.

### WebSocket

Add `character_updated` to the panel's event listener (alongside existing `item_updated`) to trigger a currency balance refresh.

---

## Files Changed

| File | Change |
|------|--------|
| `web/src/App.css` | Add `.reanalyze-btn` class; add currency row + toast styles |
| `web/src/NPCRosterPanel.tsx` | Rename `npc-reanalyze-btn` → `reanalyze-btn` |
| `web/src/ObjectivesPanel.tsx` | Rename `npc-reanalyze-btn` → `reanalyze-btn` |
| `web/public/favicon.svg` | Replace with quill & inkwell SVG |
| `web/src/InventoryPanel.tsx` | Add currency row, inline edit, undo toast |
| `web/src/types.ts` | Add `currency_balance` and `currency_label` to `Character` interface |
| `web/src/api.ts` | Extend `patchCharacter` to accept currency fields |
| `internal/db/migrations/` | New migration: add currency columns to characters |
| `internal/api/` | Extend PATCH /characters/{id} handler; add `autoUpdateCurrency` goroutine |

---

## Out of Scope

- Per-system currency denomination (multi-denomination is a future concern)
- Currency history / transaction log (chat history serves this role)
- Shop/merchant UI (buy flow is GM-narrated; AI handles the deduction)
