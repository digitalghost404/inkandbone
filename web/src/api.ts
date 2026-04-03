import type { GameContext } from './types'

export async function fetchContext(): Promise<GameContext> {
  const res = await fetch('/api/context')
  if (!res.ok) throw new Error(`GET /api/context failed: ${res.status}`)
  return res.json()
}
