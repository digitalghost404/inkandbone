import type { GameContext, WorldNote, DiceRoll, TimelineEntry } from './types'

export interface CampaignMap {
  id: number;
  campaign_id: number;
  image_path: string;
  created_at: string;
}

export interface MapPin {
  id: number;
  map_id: number;
  x: number;
  y: number;
  label: string;
  note: string;
  color: string;
  created_at: string;
}

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

export async function fetchTimeline(sessionId: number): Promise<TimelineEntry[]> {
  const url = `/api/sessions/${sessionId}/timeline`
  const res = await fetch(url)
  if (!res.ok) throw new Error(`GET ${url} failed: ${res.status}`)
  return res.json()
}

export async function fetchMaps(campaignId: number): Promise<CampaignMap[]> {
  const url = `/api/campaigns/${campaignId}/maps`
  const res = await fetch(url)
  if (!res.ok) throw new Error(`GET ${url} failed: ${res.status}`)
  return res.json()
}

export async function fetchMapPins(mapId: number): Promise<MapPin[]> {
  const url = `/api/maps/${mapId}/pins`
  const res = await fetch(url)
  if (!res.ok) throw new Error(`GET ${url} failed: ${res.status}`)
  return res.json()
}

export async function patchSessionSummary(sessionId: number, summary: string): Promise<void> {
  const res = await fetch(`/api/sessions/${sessionId}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ summary }),
  })
  if (!res.ok) throw new Error(`PATCH /api/sessions/${sessionId} failed: ${res.status}`)
}

export async function generateRecap(sessionId: number): Promise<{ summary: string }> {
  const url = `/api/sessions/${sessionId}/recap`
  const res = await fetch(url, { method: 'POST' })
  if (!res.ok) throw new Error(`POST ${url} failed: ${res.status}`)
  return res.json()
}

export async function draftWorldNote(campaignId: number, hint: string): Promise<{ id: number; title: string; content: string }> {
  const url = `/api/campaigns/${campaignId}/world-notes/draft`
  const res = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ hint }),
  })
  if (!res.ok) throw new Error(`POST ${url} failed: ${res.status}`)
  return res.json()
}

export async function uploadMap(campaignId: number, file: File): Promise<CampaignMap> {
  const url = `/api/campaigns/${campaignId}/maps`
  const form = new FormData()
  form.append('image', file)
  const res = await fetch(url, {
    method: 'POST',
    body: form,
  })
  if (!res.ok) throw new Error(`POST ${url} failed: ${res.status}`)
  return res.json()
}

export interface Character {
  id: number;
  name: string;
  data_json: string;
  campaign_id: number;
  portrait_path: string;
}

export interface Ruleset {
  id: number;
  name: string;
  schema_json: string;
}

export async function fetchRuleset(rulesetId: number): Promise<Ruleset> {
  const res = await fetch(`/api/rulesets/${rulesetId}`)
  if (!res.ok) throw new Error(`fetchRuleset failed: ${res.status}`)
  return res.json()
}

export async function patchCharacter(characterId: number, updates: Record<string, unknown>): Promise<Character> {
  const res = await fetch(`/api/characters/${characterId}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ data_json: JSON.stringify(updates) }),
  })
  if (!res.ok) throw new Error(`patchCharacter failed: ${res.status}`)
  return res.json()
}

export async function uploadPortrait(characterId: number, file: File): Promise<Character> {
  const form = new FormData()
  form.append('portrait', file)
  const res = await fetch(`/api/characters/${characterId}/portrait`, {
    method: 'POST',
    body: form,
  })
  if (!res.ok) throw new Error(`uploadPortrait failed: ${res.status}`)
  return res.json()
}

export async function ingestRulebook(rulesetId: number, text: string): Promise<{ chunks_created: number }> {
  const res = await fetch(`/api/rulesets/${rulesetId}/rulebook`, {
    method: 'POST',
    headers: { 'Content-Type': 'text/plain' },
    body: text,
  })
  if (!res.ok) throw new Error(`ingestRulebook failed: ${res.status}`)
  return res.json()
}
