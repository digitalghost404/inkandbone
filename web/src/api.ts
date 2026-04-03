import type { GameContext, WorldNote, DiceRoll } from './types'

export async function fetchContext(): Promise<GameContext> {
  const res = await fetch('/api/context')
  if (!res.ok) throw new Error(`GET /api/context failed: ${res.status}`)
  return res.json()
}

export async function fetchWorldNotes(campaignId: number, q?: string, tag?: string): Promise<WorldNote[]> {
  const params = new URLSearchParams()
  if (q) params.set('q', q)
  if (tag) params.set('tag', tag)
  const qs = params.toString()
  const url = qs
    ? `/api/campaigns/${campaignId}/world-notes?${qs}`
    : `/api/campaigns/${campaignId}/world-notes`
  const res = await fetch(url)
  if (!res.ok) throw new Error(`GET ${url} failed: ${res.status}`)
  return res.json()
}

export async function fetchDiceRolls(sessionId: number): Promise<DiceRoll[]> {
  const url = `/api/sessions/${sessionId}/dice-rolls`
  const res = await fetch(url)
  if (!res.ok) throw new Error(`GET ${url} failed: ${res.status}`)
  return res.json()
}
