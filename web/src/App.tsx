import { useState, useEffect, useCallback } from 'react'
import { useWebSocket } from './useWebSocket'
import { fetchContext } from './api'
import type { GameContext, Message } from './types'
import { CombatPanel } from './CombatPanel'
import { SessionTimeline } from './SessionTimeline'
import { WorldNotesPanel } from './WorldNotesPanel'
import { DiceHistoryPanel } from './DiceHistoryPanel'
import './App.css'

const WS_URL = `ws://${window.location.host}/ws`

export default function App() {
  const [ctx, setCtx] = useState<GameContext | null>(null)
  const [messages, setMessages] = useState<Message[]>([])
  const [error, setError] = useState<string | null>(null)

  const loadContext = useCallback(() => {
    fetchContext()
      .then((data) => {
        setCtx(data)
        setMessages(data.recent_messages ?? [])
      })
      .catch(() => setError('Could not load game state'))
  }, [])

  useEffect(() => {
    loadContext()
  }, [loadContext])

  const handleEvent = useCallback(
    (_data: unknown) => {
      loadContext()
    },
    [loadContext],
  )

  const { lastEvent } = useWebSocket(WS_URL, handleEvent)

  if (error) return <div className="error">{error}</div>
  if (!ctx) return <div className="loading">Loading…</div>

  return (
    <div className="dashboard">
      <header className="state-bar">
        <span className="campaign">{ctx.campaign?.name ?? 'No campaign'}</span>
        <span className="separator">·</span>
        {ctx.character?.portrait_path && (
          <img
            className="portrait"
            src={`/api/files/${ctx.character.portrait_path}`}
            alt={ctx.character.name}
          />
        )}
        <span className="character">{ctx.character?.name ?? 'No character'}</span>
        <span className="separator">·</span>
        <span className="session">{ctx.session?.title ?? 'No session'}</span>
      </header>

      <main className="panels">
        <section className="panel messages">
          <h2>Session Log</h2>
          {messages.length === 0 ? (
            <p className="empty">No messages yet.</p>
          ) : (
            messages.map((m) => (
              <div key={m.id} className={`message ${m.role}`}>
                <span className="role">{m.role}</span>
                <span className="content">{m.content}</span>
              </div>
            ))
          )}
        </section>

        {ctx.active_combat && <CombatPanel combat={ctx.active_combat} />}

        {ctx.session && (
          <SessionTimeline sessionId={ctx.session.id} lastEvent={lastEvent} />
        )}

        {ctx.campaign && (
          <WorldNotesPanel campaignId={ctx.campaign.id} lastEvent={lastEvent} />
        )}

        {ctx.session && (
          <DiceHistoryPanel sessionId={ctx.session.id} lastEvent={lastEvent} />
        )}
      </main>
    </div>
  )
}
