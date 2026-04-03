import { describe, it, expect, vi, afterEach } from 'vitest'
import { fetchContext } from './api'

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
