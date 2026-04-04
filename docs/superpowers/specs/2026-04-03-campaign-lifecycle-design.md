# Campaign Lifecycle: Close, Delete, Reopen

**Date:** 2026-04-03  
**Status:** Approved

## Overview

Add three campaign lifecycle operations to ink & bone:

- `close_campaign` — soft-close a campaign (preserves data, marks inactive)
- `delete_campaign` — hard-delete a campaign and all related data
- `set_active` update — automatically reopens a closed campaign when switching to it

`list_campaigns` requires no changes; it already returns `id` and `active` for every campaign.

---

## Schema

No migration needed. The `campaigns` table already has `active INTEGER NOT NULL DEFAULT 1`. Close sets it to `0`; reopen sets it back to `1`.

---

## Tools

### `close_campaign`

**Parameters:**
- `campaign_id` (number, optional) — defaults to active campaign

**Behaviour:**
1. Resolve campaign ID (param or `active_campaign_id` setting).
2. Check `active_session_id` setting — if it belongs to this campaign, return an error: `"end your current session before closing the campaign"`.
3. Call `db.CloseCampaign(id)` — sets `campaigns.active = 0`.
4. Clear `active_campaign_id`, `active_session_id`, `active_character_id` from settings if they belong to this campaign.
5. Publish `campaign_closed` event.
6. Return `"campaign <id> closed: <name>"`.

**DB method:** `CloseCampaign(id int64) error`

---

### `delete_campaign`

**Parameters:**
- `campaign_id` (number, **required**)
- `confirm` (bool, **required**)

**Behaviour without `confirm: true`:**
Return an error listing what will be deleted:
```
campaign <id> "<name>" and all its data will be permanently deleted:
  - <N> sessions, <N> characters, <N> world notes, <N> maps
call delete_campaign again with confirm: true to proceed
```

**Behaviour with `confirm: true`:**
1. Call `db.DeleteCampaign(id)` in a single transaction cascading in this order:
   `dice_rolls → messages → combatants → combat_encounters → sessions → world_notes → map_pins → maps → characters → campaigns`
2. Clear `active_campaign_id`, `active_session_id`, `active_character_id` from settings if they belonged to this campaign.
3. Publish `campaign_deleted` event.
4. Return `"campaign <id> deleted"`.

**DB method:** `DeleteCampaign(id int64) error`

**Note:** `campaign_id` is required (not defaulted) as an intentional second safety check alongside `confirm`.

---

### `set_active` update

When `campaign_id` is provided and that campaign has `active = 0`:
1. Call `db.ReopenCampaign(id)` — sets `campaigns.active = 1`.
2. Set `active_campaign_id` in settings.
3. Publish `campaign_reopened` event.

Existing behaviour (just updating settings) is unchanged for already-open campaigns.

**DB method:** `ReopenCampaign(id int64) error`

---

## Events

| Event | Payload |
|---|---|
| `campaign_closed` | `{ campaign_id }` |
| `campaign_deleted` | `{ campaign_id }` |
| `campaign_reopened` | `{ campaign_id }` |

Add these constants to `internal/api/events.go`.

---

## Files Changed

| File | Change |
|---|---|
| `internal/db/queries_core.go` | Add `CloseCampaign`, `DeleteCampaign`, `ReopenCampaign` |
| `internal/api/events.go` | Add 3 event type constants |
| `internal/mcp/lifecycle.go` | Add `handleCloseCampaign`, `handleDeleteCampaign` |
| `internal/mcp/campaign.go` | Update `handleSetActive` to reopen closed campaigns |
| `internal/mcp/server.go` | Register `close_campaign` and `delete_campaign` tools |

---

## Out of Scope

- Archiving/exporting campaign data before delete
- Undo/restore after delete
