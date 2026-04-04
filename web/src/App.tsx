import { useState, useEffect, useCallback, useRef } from 'react'
import type { ReactNode } from 'react'
import { useWebSocket } from './useWebSocket'
import { fetchContext, sendMessage, gmRespond, generateMap } from './api'
import type { GameContext, Message } from './types'
import { CombatPanel } from './CombatPanel'
import { WorldNotesPanel } from './WorldNotesPanel'
import { DiceHistoryPanel } from './DiceHistoryPanel'
import { MapPanel } from './MapPanel'
import { JournalPanel } from './JournalPanel'
import { CharacterSheetPanel } from './CharacterSheetPanel'
import './App.css'

const WS_URL = `ws://${window.location.host}/ws`

function ProseJournal({ messages, characterName }: { messages: Message[]; characterName: string }) {
  if (messages.length === 0) {
    return <p className="empty">The story has not yet begun.</p>
  }

  const nodes: ReactNode[] = []
  messages.forEach((m, i) => {
    if (m.role === 'assistant') {
      nodes.push(
        <p key={m.id} className="prose-gm">{m.content}</p>
      )
    } else {
      nodes.push(
        <div key={m.id} className="prose-player">
          <div className="prose-player-label">{characterName} speaks</div>
          <p className="prose-player-text">{m.content}</p>
        </div>
      )
      // Decorative divider after each player turn (except the last message)
      if (i < messages.length - 1) {
        nodes.push(
          <div key={`div-${m.id}`} className="prose-divider">◆</div>
        )
      }
    }
  })
  return <>{nodes}</>
}

export default function App() {
  const [ctx, setCtx] = useState<GameContext | null>(null)
  const [messages, setMessages] = useState<Message[]>([])
  const [error, setError] = useState<string | null>(null)
  const [aiEnabled, setAiEnabled] = useState(false)
  const [mapOpen, setMapOpen] = useState(false)
  const [rightTab, setRightTab] = useState<'notes' | 'journal'>('notes')
  const [input, setInput] = useState('')
  const [sending, setSending] = useState(false)
  const [gmResponding, setGmResponding] = useState(false)
  const [generatingMap, setGeneratingMap] = useState(false)
  const scrollRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    fetch('/api/health')
      .then((r) => r.json())
      .then((data: { ai_enabled: boolean }) => setAiEnabled(data.ai_enabled))
      .catch(() => setAiEnabled(false))
  }, [])

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

  const handleEvent = useCallback((_data: unknown) => { loadContext() }, [loadContext])
  const { lastEvent } = useWebSocket(WS_URL, handleEvent)

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [messages])

  const handleSend = useCallback(async () => {
    const text = input.trim()
    if (!text || !ctx?.session || sending) return
    setSending(true)
    setInput('')
    try {
      await sendMessage(ctx.session.id, text)
      loadContext()
      setGmResponding(true)
      await gmRespond(ctx.session.id)
      loadContext()
    } catch {
      setInput(text)
    } finally {
      setSending(false)
      setGmResponding(false)
    }
  }, [input, ctx, sending, loadContext])

  const handleGenerateMap = useCallback(async () => {
    if (!ctx?.campaign || !aiEnabled || generatingMap) return
    setGeneratingMap(true)
    setMapOpen(true)
    const recentText = messages.slice(-6).map(m => `[${m.role}]: ${m.content}`).join('\n')
    const context = `Campaign: ${ctx.campaign.name}\n\n${recentText}`
    const mapName = ctx.session?.title ?? ctx.campaign.name
    try {
      await generateMap(ctx.campaign.id, mapName, context)
    } finally {
      setGeneratingMap(false)
    }
  }, [ctx, aiEnabled, generatingMap, messages, setMapOpen])

  if (error) return <div className="error">{error}</div>
  if (!ctx) return <div className="loading">Loading…</div>

  const sessionTitle = ctx.session?.title?.toUpperCase() ?? ''
  const sessionDate = ctx.session?.date
    ? new Date(ctx.session.date).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })
    : ''

  return (
    <div className="grimoire">
      <header className="grimoire-header">
        <span className="h-campaign">{ctx.campaign?.name ?? 'No campaign'}</span>
        <span className="h-sep">›</span>
        <span className="h-char">{ctx.character?.name ?? 'No character'}</span>
        <span className="h-sep">›</span>
        <span className="h-session">{ctx.session?.title ?? 'No session'}</span>
      </header>

      <div className="grimoire-body">

        {/* Left Sidebar */}
        <aside className="sidebar-left">
          <CharacterSheetPanel
            character={ctx?.character ?? null}
            rulesetId={ctx?.campaign?.ruleset_id ?? null}
            lastEvent={lastEvent}
          />
          <hr className="sidebar-rule" />
          {ctx.session && (
            <DiceHistoryPanel sessionId={ctx.session.id} lastEvent={lastEvent} />
          )}
        </aside>

        {/* Center Column */}
        <main className="story-center">
          <div className="story-scroll" ref={scrollRef}>
            {sessionTitle && (
              <>
                <div className="session-title">✦ {sessionTitle} ✦</div>
                {sessionDate && <div className="session-date">{sessionDate}</div>}
              </>
            )}
            {ctx.active_combat && <CombatPanel combat={ctx.active_combat} />}
            <ProseJournal messages={messages} characterName={ctx.character?.name ?? 'Player'} />
            {gmResponding && <p className="gm-thinking">▸ The GM is narrating…</p>}
          </div>

          <div className="player-input-bar">
            <textarea
              className="player-input-field"
              placeholder="What do you do?"
              value={input}
              disabled={sending || !ctx.session}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' && !e.shiftKey) {
                  e.preventDefault()
                  handleSend()
                }
              }}
              rows={3}
            />
            <button
              type="button"
              className="player-input-send"
              disabled={sending || !input.trim() || !ctx.session}
              onClick={handleSend}
            >
              {sending ? '…' : '↵'}
            </button>
          </div>

          <div className="map-drawer">
            <div className="map-drawer-handle-row">
              <button
                type="button"
                className="map-drawer-handle"
                onClick={() => setMapOpen((o) => !o)}
              >
                {mapOpen
                  ? '[ ▴ COLLAPSE ]'
                  : `[ ${ctx.campaign?.name?.toUpperCase() ?? 'THE IRONLANDS'} ▾ ]`}
              </button>
              {aiEnabled && (
                <button
                  type="button"
                  className="map-generate-btn"
                  onClick={handleGenerateMap}
                  disabled={generatingMap}
                  title="Generate a map with AI"
                >
                  {generatingMap ? '…' : '✦ Generate Map'}
                </button>
              )}
            </div>
            <div className={`map-drawer-content${mapOpen ? ' open' : ''}`}>
              <div className="map-drawer-inner">
                <MapPanel campaignId={ctx?.campaign?.id ?? null} lastEvent={lastEvent} />
              </div>
            </div>
          </div>
        </main>

        {/* Right Sidebar */}
        <aside className="sidebar-right">
          <div className="tab-bar">
            <button
              className={`tab-btn${rightTab === 'notes' ? ' active' : ''}`}
              onClick={() => setRightTab('notes')}
            >
              Notes
            </button>
            <button
              className={`tab-btn${rightTab === 'journal' ? ' active' : ''}`}
              onClick={() => setRightTab('journal')}
            >
              Journal
            </button>
          </div>
          <div className="tab-content">
            {rightTab === 'notes' && ctx.campaign && (
              <WorldNotesPanel
                campaignId={ctx.campaign.id}
                lastEvent={lastEvent}
                aiEnabled={aiEnabled}
              />
            )}
            {rightTab === 'journal' && (
              <JournalPanel
                session={ctx?.session ?? null}
                lastEvent={lastEvent}
                aiEnabled={aiEnabled}
              />
            )}
          </div>
        </aside>

      </div>
    </div>
  )
}
