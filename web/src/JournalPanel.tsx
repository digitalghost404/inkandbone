import { useState, useEffect } from 'react'
import { patchSessionSummary, generateRecap } from './api'

interface JournalPanelProps {
  session: { id: number; summary: string } | null
  lastEvent: unknown
  aiEnabled: boolean
}

interface SessionUpdatedPayload {
  session_id: number
  summary: string
}

interface SessionUpdatedEvent {
  type: 'session_updated'
  payload: SessionUpdatedPayload
}

function isSessionUpdatedEvent(ev: unknown): ev is SessionUpdatedEvent {
  if (typeof ev !== 'object' || ev === null) return false
  const e = ev as Record<string, unknown>
  if (e['type'] !== 'session_updated') return false
  const payload = e['payload']
  if (typeof payload !== 'object' || payload === null) return false
  const p = payload as Record<string, unknown>
  return typeof p['session_id'] === 'number' && typeof p['summary'] === 'string'
}

export function JournalPanel({ session, lastEvent, aiEnabled }: JournalPanelProps) {
  const [draft, setDraft] = useState(session?.summary ?? '')

  useEffect(() => {
    if (session) {
      setDraft(session.summary)
    }
  // Intentionally omit session.summary: reset draft only when session changes,
  // not on every re-render with a server-side summary update.
  }, [session?.id]) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (!isSessionUpdatedEvent(lastEvent)) return
    if (lastEvent.payload.session_id !== session?.id) return
    setDraft(lastEvent.payload.summary)
  }, [lastEvent, session?.id])

  if (session === null) return null

  function handleBlur() {
    patchSessionSummary(session!.id, draft).catch(console.error)
  }

  async function handleGenerateRecap() {
    const result = await generateRecap(session!.id)
    setDraft(result.summary)
    await patchSessionSummary(session!.id, result.summary).catch(console.error)
  }

  return (
    <section className="panel journal-panel">
      <h2>Session Journal</h2>
      <textarea
        rows={6}
        style={{ width: '100%' }}
        value={draft}
        onChange={(e) => setDraft(e.target.value)}
        onBlur={handleBlur}
      />
      <button disabled={!aiEnabled} onClick={handleGenerateRecap}>
        Generate recap
      </button>
    </section>
  )
}
