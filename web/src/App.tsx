import { useState, useEffect, useCallback, useRef } from 'react'
import type { ReactNode } from 'react'
import ReactMarkdown from 'react-markdown'
import { useWebSocket } from './useWebSocket'
import { fetchContext, sendMessage, gmRespondStream, generateMap, createMapPin } from './api'
import type { GameContext, Message } from './types'
import { CombatPanel } from './CombatPanel'
import { WorldNotesPanel } from './WorldNotesPanel'
import { DiceHistoryPanel } from './DiceHistoryPanel'
import { DiceRoller } from './DiceRoller'
import { MapPanel } from './MapPanel'
import { JournalPanel } from './JournalPanel'
import { CharacterSheetPanel } from './CharacterSheetPanel'
import { NPCRosterPanel } from './NPCRosterPanel'
import { ObjectivesPanel } from './ObjectivesPanel'
import { InventoryPanel } from './InventoryPanel'
import { ManagePanel } from './ManagePanel'
import { OraclePanel } from './OraclePanel'
import { RelationshipsPanel } from './RelationshipsPanel'
import './App.css'

const WS_URL = `ws://${window.location.host}/ws`

// ── Turn Order Strip ────────────────────────────────────────

interface TurnOrderStripProps {
  combatants: GameContext['active_combat'] extends null ? never : NonNullable<GameContext['active_combat']>['combatants']
}

function TurnOrderStrip({ combatants }: TurnOrderStripProps) {
  return (
    <div className="turn-strip">
      {combatants.map((c, idx) => {
        const isDead = c.hp_current <= 0
        const isActive = idx === 0
        return (
          <div
            key={c.id}
            className={`turn-chip${isActive ? ' active-turn' : ''}${isDead ? ' dead' : ''}`}
          >
            {c.name} ({c.initiative})
          </div>
        )
      })}
    </div>
  )
}

// ── Pin Placement Modal ─────────────────────────────────────

interface PinPlacementModalProps {
  mapId: number
  mapImagePath: string
  defaultLabel: string
  onClose: () => void
}

function PinPlacementModal({ mapId, mapImagePath, defaultLabel, onClose }: PinPlacementModalProps) {
  const [label, setLabel] = useState(defaultLabel.slice(0, 60))
  const [note, setNote] = useState('')
  const [pos, setPos] = useState<{ x: number; y: number } | null>(null)
  const [saving, setSaving] = useState(false)
  const imgRef = useRef<HTMLImageElement>(null)

  function handleImageClick(e: React.MouseEvent<HTMLImageElement>) {
    const rect = imgRef.current?.getBoundingClientRect()
    if (!rect) return
    setPos({
      x: (e.clientX - rect.left) / rect.width,
      y: (e.clientY - rect.top) / rect.height,
    })
  }

  async function handleSubmit() {
    if (!pos) return
    setSaving(true)
    try {
      await createMapPin(mapId, { x: pos.x, y: pos.y, label, note, color: '#c9a84c' })
      onClose()
    } catch (err) {
      console.error(err)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="pin-modal-backdrop" onClick={(e) => { if (e.target === e.currentTarget) onClose() }}>
      <div className="pin-modal">
        <div className="pin-modal-header">
          <span>Place Map Pin</span>
          <button className="pin-modal-close" onClick={onClose}>×</button>
        </div>
        <p className="pin-modal-hint">Click on the map to place the pin</p>
        <div className="pin-modal-map-wrap">
          <img
            ref={imgRef}
            src={`/api/files/${mapImagePath}`}
            alt="Map"
            className="pin-modal-map"
            onClick={handleImageClick}
          />
          {pos && (
            <div
              className="pin-modal-marker"
              style={{ left: `${pos.x * 100}%`, top: `${pos.y * 100}%` }}
            >
              ✦
            </div>
          )}
        </div>
        <input
          className="pin-modal-input"
          value={label}
          onChange={(e) => setLabel(e.target.value)}
          placeholder="Label…"
        />
        <textarea
          className="pin-modal-textarea"
          value={note}
          onChange={(e) => setNote(e.target.value)}
          placeholder="Note…"
          rows={3}
        />
        <button
          className="pin-modal-submit"
          onClick={handleSubmit}
          disabled={!pos || saving || !label.trim()}
        >
          {saving ? 'Saving…' : 'Place Pin'}
        </button>
      </div>
    </div>
  )
}

// ── Prose Journal ───────────────────────────────────────────

function highlightText(text: string, query: string): ReactNode {
  if (!query) return text
  const lower = text.toLowerCase()
  const lowerQ = query.toLowerCase()
  const parts: ReactNode[] = []
  let start = 0
  let idx = lower.indexOf(lowerQ, start)
  while (idx !== -1) {
    if (idx > start) parts.push(text.slice(start, idx))
    parts.push(<mark key={idx}>{text.slice(idx, idx + query.length)}</mark>)
    start = idx + query.length
    idx = lower.indexOf(lowerQ, start)
  }
  if (start < text.length) parts.push(text.slice(start))
  return <>{parts}</>
}

interface ProseJournalProps {
  messages: Message[]
  characterName: string
  searchQuery?: string
  activeMapId: number | null
  activeMapImagePath: string | null
}

function ProseJournal({
  messages,
  characterName,
  searchQuery = '',
  activeMapId,
  activeMapImagePath,
}: ProseJournalProps) {
  const [pinModal, setPinModal] = useState<{ content: string } | null>(null)

  if (messages.length === 0) {
    return <p className="empty">The story has not yet begun.</p>
  }

  const nodes: ReactNode[] = []
  messages.forEach((m, i) => {
    if (m.role === 'assistant') {
      nodes.push(
        <div key={m.id} className="prose-gm prose-gm-wrap">
          <ReactMarkdown>{m.content}</ReactMarkdown>
          {activeMapId !== null && activeMapImagePath !== null && (
            <button
              className="prose-pin-btn"
              title="Place as map pin"
              onClick={() => setPinModal({ content: m.content.replace(/[#*_`\[\]]/g, '').slice(0, 60) })}
            >
              📍
            </button>
          )}
        </div>
      )
    } else {
      const isWhisper = m.whisper === true
      nodes.push(
        <div key={m.id} className={`prose-player${isWhisper ? ' prose-player--whisper' : ''}`}>
          <div className="prose-player-label">{characterName} speaks</div>
          <p className="prose-player-text">
            {searchQuery ? highlightText(m.content, searchQuery) : m.content}
          </p>
        </div>
      )
      if (i < messages.length - 1) {
        nodes.push(
          <div key={`div-${m.id}`} className="prose-divider">◆</div>
        )
      }
    }
  })

  return (
    <>
      {nodes}
      {pinModal && activeMapId !== null && activeMapImagePath !== null && (
        <PinPlacementModal
          mapId={activeMapId}
          mapImagePath={activeMapImagePath}
          defaultLabel={pinModal.content}
          onClose={() => setPinModal(null)}
        />
      )}
    </>
  )
}

// ── App ────────────────────────────────────────────────────

export default function App() {
  const [ctx, setCtx] = useState<GameContext | null>(null)
  const [messages, setMessages] = useState<Message[]>([])
  const [error, setError] = useState<string | null>(null)
  const [aiEnabled, setAiEnabled] = useState(false)
  const [mapOpen, setMapOpen] = useState(false)
  const [rightTab, setRightTab] = useState<'notes' | 'journal' | 'npcs' | 'objectives' | 'oracle' | 'relationships'>('notes')
  const [input, setInput] = useState('')
  const [sending, setSending] = useState(false)
  const [gmResponding, setGmResponding] = useState(false)
  const [streamingText, setStreamingText] = useState('')
  const [generatingMap, setGeneratingMap] = useState(false)
  const [whisperMode, setWhisperMode] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [showPlayerHistory, setShowPlayerHistory] = useState(false)
  const [theme, setTheme] = useState(() => localStorage.getItem('theme') ?? 'worn-grimoire')
  const [activeMapId, setActiveMapId] = useState<number | null>(null)
  const [activeMapImagePath, setActiveMapImagePath] = useState<string | null>(null)
  const [manageOpen, setManageOpen] = useState(false)
  const [manageTab, setManageTab] = useState<'campaigns' | 'characters' | 'sessions' | 'rulebooks'>('campaigns')
  const scrollRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme)
    localStorage.setItem('theme', theme)
  }, [theme])

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
  }, [messages, streamingText])

  const handleSend = useCallback(async () => {
    const text = input.trim()
    if (!text || !ctx?.session || sending) return
    const isWhisper = whisperMode
    setSending(true)
    setInput('')
    setWhisperMode(false)
    try {
      await sendMessage(ctx.session.id, text, isWhisper)
      loadContext()
      if (!isWhisper) {
        setGmResponding(true)
        setStreamingText('')
        await gmRespondStream(ctx.session.id, (chunk) => {
          setStreamingText((prev) => prev + chunk)
        })
        setStreamingText('')
        loadContext()
      }
    } catch {
      setInput(text)
    } finally {
      setSending(false)
      setGmResponding(false)
    }
  }, [input, ctx, sending, loadContext, whisperMode])

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
  }, [ctx, aiEnabled, generatingMap, messages])

  function handleExport() {
    if (!ctx) return
    const sessionDate = ctx.session?.date
      ? new Date(ctx.session.date).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })
      : ''
    const lines: string[] = []
    lines.push(`# ${ctx.session?.title ?? 'Session'}`)
    if (sessionDate) lines.push(`*${sessionDate}*`)
    lines.push('')
    messages.forEach(m => {
      if (m.whisper) return
      if (m.role === 'assistant') {
        lines.push(m.content)
      } else {
        lines.push(`> **${ctx.character?.name ?? 'Player'}:** ${m.content}`)
      }
      lines.push('')
    })
    const blob = new Blob([lines.join('\n')], { type: 'text/markdown' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${(ctx.session?.title ?? 'session').replace(/\s+/g, '-').toLowerCase()}.md`
    a.click()
    URL.revokeObjectURL(url)
  }

  if (error) return <div className="error">{error}</div>
  if (!ctx) return <div className="loading">Loading…</div>

  const sessionTitle = ctx.session?.title?.toUpperCase() ?? ''
  const sessionDate = ctx.session?.date
    ? new Date(ctx.session.date).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })
    : ''

  const displayMessages = searchQuery
    ? messages.filter(m => m.content.toLowerCase().includes(searchQuery.toLowerCase()))
    : messages

  return (
    <div className="grimoire">
      <header className="grimoire-header">
        <span className="h-campaign">{ctx.campaign?.name ?? 'No campaign'}</span>
        <span className="h-sep">›</span>
        <span className="h-char">{ctx.character?.name ?? 'No character'}</span>
        <span className="h-sep">›</span>
        <span className="h-session">{ctx.session?.title ?? 'No session'}</span>
        <button
          className="h-theme"
          onClick={() => setTheme(t => t === 'worn-grimoire' ? 'parchment' : 'worn-grimoire')}
          title="Toggle theme"
        >
          {theme === 'worn-grimoire' ? '☀' : '🌙'}
        </button>
        <button
          className={`h-actions-btn${showPlayerHistory ? ' active' : ''}`}
          onClick={() => setShowPlayerHistory((v) => !v)}
          title="Your actions"
        >
          ⚔ Actions
        </button>
        <button className="h-export" onClick={handleExport} title="Export session">
          ↓ Export
        </button>
        <button
          className="h-manage"
          onClick={() => setManageOpen(true)}
          title="Manage campaigns, characters, sessions"
        >
          ⚙ Manage
        </button>
      </header>

      {manageOpen && (
        <ManagePanel
          activeCampaignId={ctx?.campaign?.id ?? null}
          activeCharacterId={ctx?.character?.id ?? null}
          activeSessionId={ctx?.session?.id ?? null}
          initialTab={manageTab}
          onTabChange={setManageTab}
          onClose={() => setManageOpen(false)}
          onContextChanged={() => { loadContext(); setManageOpen(false) }}
        />
      )}

      <div className="grimoire-body">

        {/* Player History Overlay */}
        {showPlayerHistory && (
          <div className="player-history-overlay">
            <div className="player-history-header">
              <span>Your Actions</span>
              <button onClick={() => setShowPlayerHistory(false)}>×</button>
            </div>
            <div className="player-history-list">
              {messages.filter(m => m.role === 'user' && !m.whisper).map(m => (
                <div key={m.id} className="player-history-item">
                  <p>{m.content}</p>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Left Sidebar */}
        <aside className="sidebar-left">
          <CharacterSheetPanel
            character={ctx?.character ?? null}
            rulesetId={ctx?.campaign?.ruleset_id ?? null}
            lastEvent={lastEvent}
          />
          {ctx.session && (
            <DiceRoller sessionId={ctx.session.id} />
          )}
          {ctx.character && (
            <InventoryPanel characterId={ctx.character.id} lastEvent={lastEvent} />
          )}
          <hr className="sidebar-rule" />
          {ctx.session && (
            <DiceHistoryPanel sessionId={ctx.session.id} lastEvent={lastEvent} />
          )}
        </aside>

        {/* Center Column */}
        <main className="story-center">
          {ctx.active_combat && (
            <TurnOrderStrip combatants={ctx.active_combat.combatants} />
          )}

          <div className="story-search-bar">
            <input
              type="search"
              placeholder="Search story…"
              value={searchQuery}
              onChange={e => setSearchQuery(e.target.value)}
            />
            {searchQuery && (
              <button onClick={() => setSearchQuery('')}>×</button>
            )}
          </div>

          <div className="story-scroll" ref={scrollRef}>
            {sessionTitle && (
              <>
                <div className="session-title">✦ {sessionTitle} ✦</div>
                {sessionDate && <div className="session-date">{sessionDate}</div>}
              </>
            )}
            {ctx.active_combat && <CombatPanel combat={ctx.active_combat} />}
            <ProseJournal
              messages={displayMessages}
              characterName={ctx.character?.name ?? 'Player'}
              searchQuery={searchQuery}
              activeMapId={activeMapId}
              activeMapImagePath={activeMapImagePath}
            />
            {streamingText && (
              <div className="prose-gm streaming">
                <ReactMarkdown>{streamingText}</ReactMarkdown>
              </div>
            )}
            {gmResponding && !streamingText && (
              <p className="gm-thinking">▸ The GM is narrating…</p>
            )}
          </div>

          <div className="player-input-bar">
            <button
              type="button"
              className={`whisper-toggle${whisperMode ? ' active' : ''}`}
              onClick={() => setWhisperMode((v) => !v)}
              title={whisperMode ? 'Whisper mode on — GM will not respond' : 'Enable whisper mode'}
            >
              🔒
            </button>
            <textarea
              className={`player-input-field${whisperMode ? ' whisper-active' : ''}`}
              placeholder={whisperMode ? 'Whisper (private, no GM response)…' : 'What do you do?'}
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
                <MapPanel
                  campaignId={ctx?.campaign?.id ?? null}
                  lastEvent={lastEvent}
                  onActiveMapChange={(mapId, imagePath) => {
                    setActiveMapId(mapId)
                    setActiveMapImagePath(imagePath)
                  }}
                />
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
            <button
              className={`tab-btn${rightTab === 'npcs' ? ' active' : ''}`}
              onClick={() => setRightTab('npcs')}
            >
              NPCs
            </button>
            <button
              className={`tab-btn${rightTab === 'objectives' ? ' active' : ''}`}
              onClick={() => setRightTab('objectives')}
            >
              Objectives
            </button>
            <button
              className={`tab-btn${rightTab === 'oracle' ? ' active' : ''}`}
              onClick={() => setRightTab('oracle')}
            >
              Oracle
            </button>
            <button
              className={`tab-btn${rightTab === 'relationships' ? ' active' : ''}`}
              onClick={() => setRightTab('relationships')}
            >
              Relations
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
                campaignId={ctx?.campaign?.id ?? null}
                lastEvent={lastEvent}
                aiEnabled={aiEnabled}
              />
            )}
            {rightTab === 'npcs' && (
              <NPCRosterPanel
                sessionId={ctx?.session?.id ?? null}
                lastEvent={lastEvent}
              />
            )}
            {rightTab === 'objectives' && (
              <ObjectivesPanel campaignId={ctx?.campaign?.id ?? null} lastEvent={lastEvent} />
            )}
            {rightTab === 'oracle' && ctx.session && (
              <OraclePanel sessionId={ctx.session.id} />
            )}
            {rightTab === 'relationships' && ctx.campaign && (
              <RelationshipsPanel campaignId={ctx.campaign.id} />
            )}
          </div>
        </aside>

      </div>
    </div>
  )
}
