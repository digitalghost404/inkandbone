import type { GameContext, WorldNote, DiceRoll, TimelineEntry, SessionNPC } from './types'

export interface CampaignMap {
  id: number;
  campaign_id: number;
  name: string;
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
  form.append('name', file.name.replace(/\.[^.]+$/, ''))
  const res = await fetch(url, {
    method: 'POST',
    body: form,
  })
  if (!res.ok) throw new Error(`POST ${url} failed: ${res.status}`)
  return res.json()
}

export interface Ruleset {
  id: number;
  name: string;
  schema_json: string;
  version: string;
}

export async function fetchRuleset(rulesetId: number): Promise<Ruleset> {
  const res = await fetch(`/api/rulesets/${rulesetId}`)
  if (!res.ok) throw new Error(`fetchRuleset failed: ${res.status}`)
  return res.json()
}

export async function patchCharacter(characterId: number, updates: Record<string, unknown>): Promise<void> {
  const res = await fetch(`/api/characters/${characterId}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ data_json: JSON.stringify(updates) }),
  })
  if (!res.ok) throw new Error(`patchCharacter failed: ${res.status}`)
}

export async function uploadPortrait(characterId: number, file: File): Promise<{ portrait_path: string }> {
  const form = new FormData()
  form.append('portrait', file)
  const res = await fetch(`/api/characters/${characterId}/portrait`, {
    method: 'POST',
    body: form,
  })
  if (!res.ok) throw new Error(`uploadPortrait failed: ${res.status}`)
  return res.json()
}

export async function sendMessage(sessionId: number, content: string, whisper?: boolean): Promise<void> {
  const body: Record<string, unknown> = { role: 'user', content }
  if (whisper) body['whisper'] = true
  const res = await fetch(`/api/sessions/${sessionId}/messages`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error(`sendMessage failed: ${res.status}`)
}

export async function generateMap(campaignId: number, name: string, context: string): Promise<CampaignMap> {
  const res = await fetch(`/api/campaigns/${campaignId}/maps/generate`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, context }),
  })
  if (!res.ok) throw new Error(`generateMap failed: ${res.status}`)
  return res.json()
}

export async function gmRespondStream(
  sessionId: number,
  onChunk: (text: string) => void,
): Promise<string> {
  const res = await fetch(`/api/sessions/${sessionId}/gm-respond-stream`, { method: 'POST' })
  if (!res.ok) throw new Error(`gmRespondStream failed: ${res.status}`)
  const reader = res.body?.getReader()
  if (!reader) return ''
  const decoder = new TextDecoder()
  let accumulated = ''
  let buffer = ''
  while (true) {
    const { done, value } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true })
    const lines = buffer.split('\n')
    buffer = lines.pop() ?? ''
    for (const line of lines) {
      if (line.startsWith('data: ')) {
        const chunk = line.slice(6)
        accumulated += chunk
        onChunk(chunk)
      }
    }
  }
  // flush remaining buffer
  if (buffer.startsWith('data: ')) {
    const chunk = buffer.slice(6)
    accumulated += chunk
    onChunk(chunk)
  }
  return accumulated
}

export async function rollDice(
  sessionId: number,
  expression: string,
): Promise<{ expression: string; result: number; rolls: number[] }> {
  const res = await fetch(`/api/sessions/${sessionId}/dice-rolls`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ expression }),
  })
  if (!res.ok) throw new Error(`rollDice failed: ${res.status}`)
  return res.json()
}

export async function patchCombatant(
  combatantId: number,
  updates: { conditions_json?: string; hp_current?: number },
): Promise<void> {
  const res = await fetch(`/api/combatants/${combatantId}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(updates),
  })
  if (!res.ok) throw new Error(`patchCombatant failed: ${res.status}`)
}

export async function createMapPin(
  mapId: number,
  pin: { x: number; y: number; label: string; note: string; color: string },
): Promise<MapPin> {
  const res = await fetch(`/api/maps/${mapId}/pins`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(pin),
  })
  if (!res.ok) throw new Error(`createMapPin failed: ${res.status}`)
  return res.json()
}

export async function fetchNPCs(sessionId: number): Promise<SessionNPC[]> {
  const res = await fetch(`/api/sessions/${sessionId}/npcs`)
  if (!res.ok) throw new Error(`fetchNPCs failed: ${res.status}`)
  return res.json()
}

export async function createNPC(sessionId: number, name: string, note: string): Promise<SessionNPC> {
  const res = await fetch(`/api/sessions/${sessionId}/npcs`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, note }),
  })
  if (!res.ok) throw new Error(`createNPC failed: ${res.status}`)
  return res.json()
}

export async function patchNPC(npcId: number, note: string): Promise<void> {
  const res = await fetch(`/api/npcs/${npcId}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ note }),
  })
  if (!res.ok) throw new Error(`patchNPC failed: ${res.status}`)
}

export async function deleteNPC(npcId: number): Promise<void> {
  const res = await fetch(`/api/npcs/${npcId}`, { method: 'DELETE' })
  if (!res.ok) throw new Error(`deleteNPC failed: ${res.status}`)
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
