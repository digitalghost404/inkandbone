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
  const res = await fetch(`/api/campaigns/${campaignId}/maps`);
  return res.json();
}

export async function fetchMapPins(mapId: number): Promise<MapPin[]> {
  const res = await fetch(`/api/maps/${mapId}/pins`);
  return res.json();
}

export async function patchSessionSummary(sessionId: number, summary: string): Promise<void> {
  await fetch(`/api/sessions/${sessionId}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ summary }),
  });
}

export async function generateRecap(sessionId: number): Promise<{ summary: string }> {
  const res = await fetch(`/api/sessions/${sessionId}/recap`, { method: 'POST' });
  return res.json();
}

export async function draftWorldNote(campaignId: number, hint: string): Promise<{ id: number; title: string; content: string }> {
  const res = await fetch(`/api/campaigns/${campaignId}/world-notes/draft`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ hint }),
  });
  return res.json();
}

export async function uploadMap(campaignId: number, file: File): Promise<CampaignMap> {
  const form = new FormData();
  form.append('image', file);
  const res = await fetch(`/api/campaigns/${campaignId}/maps`, {
    method: 'POST',
    body: form,
  });
  return res.json();
}
