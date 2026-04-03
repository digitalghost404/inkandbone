import { describe, it, expect, vi, afterEach } from 'vitest'
import { fetchContext, fetchWorldNotes, fetchDiceRolls } from './api'

afterEach(() => vi.restoreAllMocks())

describe('fetchContext', () => {
  it('returns parsed GameContext on success', async () => {
    const payload = {
      campaign: null,
      character: null,
      session: null,
      recent_messages: [],
      active_combat: null,
    }
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(payload),
    }))

    const ctx = await fetchContext()
    expect(ctx.campaign).toBeNull()
    expect(ctx.recent_messages).toEqual([])
  })

  it('throws on non-ok response', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: false,
      status: 500,
    }))

    await expect(fetchContext()).rejects.toThrow('GET /api/context failed: 500')
  })
})

describe('fetchWorldNotes', () => {
  it('returns parsed WorldNote array on success', async () => {
    const notes = [
      { id: 1, campaign_id: 1, title: 'Tavern', content: 'A seedy place.', category: 'location', tags_json: '[]', created_at: '' },
    ]
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(notes),
    }))

    const result = await fetchWorldNotes(1)
    expect(result).toHaveLength(1)
    expect(result[0].title).toBe('Tavern')
  })

  it('appends q param when query is provided', async () => {
    const mockFetch = vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve([]) })
    vi.stubGlobal('fetch', mockFetch)

    await fetchWorldNotes(1, 'dragon')
    expect(mockFetch).toHaveBeenCalledWith('/api/campaigns/1/world-notes?q=dragon')
  })

  it('omits q param when query is empty string', async () => {
    const mockFetch = vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve([]) })
    vi.stubGlobal('fetch', mockFetch)

    await fetchWorldNotes(1, '')
    expect(mockFetch).toHaveBeenCalledWith('/api/campaigns/1/world-notes')
  })

  it('throws on non-ok response', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false, status: 404 }))
    await expect(fetchWorldNotes(1)).rejects.toThrow('failed: 404')
  })

  it('appends tag param when tag is provided', async () => {
    const mockFetch = vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve([]) })
    vi.stubGlobal('fetch', mockFetch)
    await fetchWorldNotes(1, undefined, 'npc')
    expect(mockFetch).toHaveBeenCalledWith('/api/campaigns/1/world-notes?tag=npc')
  })

  it('appends both q and tag when both provided', async () => {
    const mockFetch = vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve([]) })
    vi.stubGlobal('fetch', mockFetch)
    await fetchWorldNotes(1, 'tavern', 'location')
    expect(mockFetch).toHaveBeenCalledWith('/api/campaigns/1/world-notes?q=tavern&tag=location')
  })
})

describe('fetchDiceRolls', () => {
  it('returns parsed DiceRoll array on success', async () => {
    const rolls = [
      { id: 1, session_id: 1, expression: '1d20+5', result: 18, breakdown_json: '[]', created_at: '' },
    ]
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(rolls),
    }))

    const result = await fetchDiceRolls(1)
    expect(result).toHaveLength(1)
    expect(result[0].expression).toBe('1d20+5')
    expect(result[0].result).toBe(18)
  })

  it('throws on non-ok response', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false, status: 500 }))
    await expect(fetchDiceRolls(1)).rejects.toThrow('failed: 500')
  })
})
