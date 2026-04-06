import type { GameContext, WorldNote, DiceRoll, TimelineEntry, SessionNPC, Objective, Item, XPEntry, Relationship } from './types'

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

export async function patchSession(sessionId: number, updates: { scene_tags?: string; summary?: string; notes?: string }): Promise<void> {
  const res = await fetch(`/api/sessions/${sessionId}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(updates),
  })
  if (!res.ok) throw new Error(`patchSession failed: ${res.status}`)
}

export async function patchSessionSummary(sessionId: number, summary: string): Promise<void> {
  const res = await fetch(`/api/sessions/${sessionId}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ summary }),
  })
  if (!res.ok) throw new Error(`PATCH /api/sessions/${sessionId} failed: ${res.status}`)
}

export async function patchSessionNotes(sessionId: number, notes: string): Promise<void> {
  const res = await fetch(`/api/sessions/${sessionId}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ notes }),
  })
  if (!res.ok) throw new Error(`patchSessionNotes failed: ${res.status}`)
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

export async function patchWorldNotePersonality(noteId: number, personalityJson: string): Promise<void> {
  const url = `/api/world-notes/${noteId}/personality`
  const res = await fetch(url, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ personality_json: personalityJson }),
  })
  if (!res.ok) throw new Error(`PATCH ${url} failed: ${res.status}`)
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

export async function fetchObjectives(campaignId: number): Promise<Objective[]> {
  const res = await fetch(`/api/campaigns/${campaignId}/objectives`)
  if (!res.ok) throw new Error(`fetchObjectives failed: ${res.status}`)
  return res.json()
}

export async function createObjective(campaignId: number, title: string, description: string, parentId?: number): Promise<Objective> {
  const res = await fetch(`/api/campaigns/${campaignId}/objectives`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ title, description, parent_id: parentId ?? null }),
  })
  if (!res.ok) throw new Error(`createObjective failed: ${res.status}`)
  return res.json()
}

export async function patchObjective(id: number, status: string): Promise<void> {
  const res = await fetch(`/api/objectives/${id}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ status }),
  })
  if (!res.ok) throw new Error(`patchObjective failed: ${res.status}`)
}

export async function deleteObjective(id: number): Promise<void> {
  const res = await fetch(`/api/objectives/${id}`, { method: 'DELETE' })
  if (!res.ok) throw new Error(`deleteObjective failed: ${res.status}`)
}

export async function fetchItems(characterId: number): Promise<Item[]> {
  const res = await fetch(`/api/characters/${characterId}/items`)
  if (!res.ok) throw new Error(`fetchItems failed: ${res.status}`)
  return res.json()
}

export async function createItem(
  characterId: number,
  name: string,
  description: string,
  quantity: number,
): Promise<Item> {
  const res = await fetch(`/api/characters/${characterId}/items`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, description, quantity }),
  })
  if (!res.ok) throw new Error(`createItem failed: ${res.status}`)
  return res.json()
}

export async function patchItem(
  id: number,
  updates: { name?: string; description?: string; quantity?: number; equipped?: boolean },
): Promise<void> {
  const res = await fetch(`/api/items/${id}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(updates),
  })
  if (!res.ok) throw new Error(`patchItem failed: ${res.status}`)
}

export async function deleteItem(id: number): Promise<void> {
  const res = await fetch(`/api/items/${id}`, { method: 'DELETE' })
  if (!res.ok) throw new Error(`deleteItem failed: ${res.status}`)
}

export async function advanceTurn(encounterId: number): Promise<void> {
  const res = await fetch(`/api/combat-encounters/${encounterId}/next-turn`, {
    method: 'POST',
  })
  if (!res.ok) throw new Error(`advanceTurn failed: ${res.status}`)
}

export async function fetchXP(sessionId: number): Promise<XPEntry[]> {
  const res = await fetch(`/api/sessions/${sessionId}/xp`)
  if (!res.ok) throw new Error(`fetchXP failed: ${res.status}`)
  return res.json()
}

export async function createXP(sessionId: number, note: string, amount?: number): Promise<XPEntry> {
  const res = await fetch(`/api/sessions/${sessionId}/xp`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ note, amount: amount ?? null }),
  })
  if (!res.ok) throw new Error(`createXP failed: ${res.status}`)
  return res.json()
}

export async function deleteXP(id: number): Promise<void> {
  const res = await fetch(`/api/xp/${id}`, { method: 'DELETE' })
  if (!res.ok) throw new Error(`deleteXP failed: ${res.status}`)
}

export async function postImprovise(sessionId: number): Promise<string> {
  const res = await fetch(`/api/sessions/${sessionId}/improvise`, { method: 'POST' })
  if (!res.ok) throw new Error('Improvise failed')
  const data = await res.json()
  return data.result
}

export async function postPreSessionBrief(campaignId: number): Promise<string> {
  const res = await fetch(`/api/campaigns/${campaignId}/pre-session-brief`, { method: 'POST' })
  if (!res.ok) throw new Error('Pre-session brief failed')
  const data = await res.json()
  return data.result
}

export async function postDetectThreads(sessionId: number): Promise<string> {
  const res = await fetch(`/api/sessions/${sessionId}/detect-threads`, { method: 'POST' })
  if (!res.ok) throw new Error('Detect threads failed')
  const data = await res.json()
  return data.result
}

export async function postCampaignAsk(campaignId: number, question: string): Promise<string> {
  const res = await fetch(`/api/campaigns/${campaignId}/ask`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ question }),
  })
  if (!res.ok) throw new Error('Campaign ask failed')
  const data = await res.json()
  return data.result
}

// --- Management API ---

export interface RulebookSource {
  source: string;
  chunks: number;
}

export async function fetchRulesets(): Promise<Ruleset[]> {
  const res = await fetch('/api/rulesets')
  if (!res.ok) throw new Error(`fetchRulesets failed: ${res.status}`)
  return res.json()
}

export async function fetchCampaigns(): Promise<import('./types').Campaign[]> {
  const res = await fetch('/api/campaigns')
  if (!res.ok) throw new Error(`fetchCampaigns failed: ${res.status}`)
  return res.json()
}

export async function fetchCharacters(campaignId: number): Promise<import('./types').Character[]> {
  const res = await fetch(`/api/campaigns/${campaignId}/characters`)
  if (!res.ok) throw new Error(`fetchCharacters failed: ${res.status}`)
  return res.json()
}

export async function fetchSessions(campaignId: number): Promise<import('./types').Session[]> {
  const res = await fetch(`/api/campaigns/${campaignId}/sessions`)
  if (!res.ok) throw new Error(`fetchSessions failed: ${res.status}`)
  return res.json()
}

export async function createCampaign(name: string, description: string, rulesetId: number): Promise<{ id: number }> {
  const res = await fetch('/api/campaigns', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, description, ruleset_id: rulesetId }),
  })
  if (!res.ok) throw new Error(`createCampaign failed: ${res.status}`)
  return res.json()
}

export async function deleteCampaign(id: number): Promise<void> {
  const res = await fetch(`/api/campaigns/${id}`, { method: 'DELETE' })
  if (!res.ok) throw new Error(`deleteCampaign failed: ${res.status}`)
}

export async function fetchCharacterOptions(rulesetId: number): Promise<Record<string, string[]>> {
  const res = await fetch(`/api/rulesets/${rulesetId}/character-options`)
  if (!res.ok) throw new Error(`fetchCharacterOptions failed: ${res.status}`)
  return res.json()
}

export async function createCharacter(
  campaignId: number,
  name: string,
  overrides?: Record<string, string>,
): Promise<import('./types').Character> {
  const res = await fetch(`/api/campaigns/${campaignId}/characters`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, overrides }),
  })
  if (!res.ok) throw new Error(`createCharacter failed: ${res.status}`)
  return res.json()
}

export async function deleteCharacter(id: number): Promise<void> {
  const res = await fetch(`/api/characters/${id}`, { method: 'DELETE' })
  if (!res.ok) throw new Error(`deleteCharacter failed: ${res.status}`)
}

export async function createSession(
  campaignId: number,
  title: string,
  date: string,
): Promise<import('./types').Session> {
  const res = await fetch(`/api/campaigns/${campaignId}/sessions`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ title, date }),
  })
  if (!res.ok) throw new Error(`createSession failed: ${res.status}`)
  return res.json()
}

export async function deleteSession(id: number): Promise<void> {
  const res = await fetch(`/api/sessions/${id}`, { method: 'DELETE' })
  if (!res.ok) throw new Error(`deleteSession failed: ${res.status}`)
}

export async function patchSettings(settings: {
  campaign_id?: number | null;
  character_id?: number | null;
  session_id?: number | null;
}): Promise<void> {
  const body: Record<string, number> = {}
  if (settings.campaign_id !== undefined) body['campaign_id'] = settings.campaign_id ?? 0
  if (settings.character_id !== undefined) body['character_id'] = settings.character_id ?? 0
  if (settings.session_id !== undefined) body['session_id'] = settings.session_id ?? 0
  const res = await fetch('/api/settings', {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error(`patchSettings failed: ${res.status}`)
}

export async function fetchRulebookSources(rulesetId: number): Promise<RulebookSource[]> {
  const res = await fetch(`/api/rulesets/${rulesetId}/rulebook`)
  if (!res.ok) throw new Error(`fetchRulebookSources failed: ${res.status}`)
  return res.json()
}

export async function uploadRulebook(
  rulesetId: number,
  file: File,
  source: string,
): Promise<{ chunks_created: number; source: string }> {
  const form = new FormData()
  form.append('rulebook', file)
  form.append('source', source)
  const res = await fetch(`/api/rulesets/${rulesetId}/rulebook`, {
    method: 'POST',
    body: form,
  })
  if (!res.ok) throw new Error(`uploadRulebook failed: ${res.status}`)
  return res.json()
}

// Oracle
export async function postOracleRoll(table: string, roll: number, rulesetId?: number): Promise<{ result: string; table: string; roll: number }> {
  const res = await fetch('/api/oracle/roll', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ table, roll, ruleset_id: rulesetId }),
  })
  if (!res.ok) throw new Error('Oracle roll failed')
  return res.json()
}

// Tension
export async function getTension(sessionId: number): Promise<number> {
  const res = await fetch(`/api/sessions/${sessionId}/tension`)
  if (!res.ok) throw new Error('Get tension failed')
  const data = await res.json()
  return data.tension_level
}

export async function patchTension(sessionId: number, level: number): Promise<void> {
  const res = await fetch(`/api/sessions/${sessionId}/tension`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ tension_level: level }),
  })
  if (!res.ok) throw new Error('Patch tension failed')
}

// Relationships
export async function listRelationships(campaignId: number): Promise<Relationship[]> {
  const res = await fetch(`/api/campaigns/${campaignId}/relationships`)
  if (!res.ok) throw new Error('List relationships failed')
  return res.json()
}

export async function createRelationship(campaignId: number, fromName: string, toName: string, type: string, description: string): Promise<{ id: number }> {
  const res = await fetch(`/api/campaigns/${campaignId}/relationships`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ from_name: fromName, to_name: toName, relationship_type: type, description }),
  })
  if (!res.ok) throw new Error('Create relationship failed')
  return res.json()
}

export async function updateRelationship(id: number, type: string, description: string): Promise<void> {
  const res = await fetch(`/api/relationships/${id}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ relationship_type: type, description }),
  })
  if (!res.ok) throw new Error('Update relationship failed')
}

export async function deleteRelationship(id: number): Promise<void> {
  const res = await fetch(`/api/relationships/${id}`, { method: 'DELETE' })
  if (!res.ok) throw new Error('Delete relationship failed')
}

export async function reanalyzeSession(sessionId: number): Promise<void> {
  const res = await fetch(`/api/sessions/${sessionId}/reanalyze`, { method: 'POST' })
  if (!res.ok) throw new Error('Reanalyze failed')
}

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

export async function fetchTalentDescription(name: string, system = 'wrath_glory'): Promise<string> {
  const res = await fetch(`/api/talent-description?name=${encodeURIComponent(name)}&system=${encodeURIComponent(system)}`)
  if (!res.ok) return ''
  const data = await res.json() as { description: string }
  return data.description ?? ''
}
