import { useState, useEffect, useCallback, useRef } from 'react'
import type { ReactNode } from 'react'
import ReactMarkdown from 'react-markdown'
import { useWebSocket } from './useWebSocket'
import { fetchContext, sendMessage, gmRespondStream, generateMap, createMapPin, patchSession, fetchRuleset, patchCampaign, suggestAdvances } from './api'
import type { GameContext, Message, Session } from './types'
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
import AudioControls, { getAudioMuted } from './AudioControls'
import { XPSuggestionsPanel } from './XPSuggestionsPanel'
import type { XPSpendSuggestionsEvent } from './types'
import { playDiceRoll, playNotification, playCombatStart } from './audio/sounds'
import { setAmbientTrack } from './audio/ambient'
import { wgTalentDescription } from './wgTalentData'
import { fetchTalentDescription } from './api'
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

// Ensure "What do you do?" at the end of GM responses is always its own paragraph
// and rendered bold+italic gold to stand out as the player prompt cue.
function normalizeGMContent(text: string): string {
  return text.replace(/\s*(\*\*)?What do you do\??(\*\*)?\s*$/, '\n\n**What do you do?**')
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
          <ReactMarkdown>{normalizeGMContent(m.content)}</ReactMarkdown>
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

// ── Scene Tag Picker ────────────────────────────────────────

const SCENE_TAGS = ['tavern', 'dungeon', 'forest', 'city', 'ocean', 'cave', 'castle', 'rain', 'night', 'battle', 'market', 'temple', 'ruins']

interface SceneTagPickerProps {
  session: Session
  onUpdate: (tags: string) => void
}

function SceneTagPicker({ session, onUpdate }: SceneTagPickerProps) {
  const activeTags = session.scene_tags ? session.scene_tags.split(',').filter(Boolean) : []

  async function toggleTag(tag: string) {
    const newTags = activeTags.includes(tag)
      ? activeTags.filter(t => t !== tag)
      : [...activeTags, tag]
    const tagsStr = newTags.join(',')
    try {
      await patchSession(session.id, { scene_tags: tagsStr })
      onUpdate(tagsStr)
      setAmbientTrack(newTags[0] ?? null)
    } catch (err) {
      console.error('Failed to update scene tags:', err)
    }
  }

  return (
    <div className="scene-tag-picker">
      {SCENE_TAGS.map(tag => (
        <button
          key={tag}
          className={`scene-tag${activeTags.includes(tag) ? ' active' : ''}`}
          onClick={() => toggleTag(tag)}
          title={tag}
        >
          {tag}
        </button>
      ))}
    </div>
  )
}

interface ChronicleNightTrackerProps {
  campaign: import('./types').Campaign
  onUpdate: (night: number) => void
}

function ChronicleNightTracker({ campaign, onUpdate }: ChronicleNightTrackerProps) {
  const night = campaign.chronicle_night ?? 1

  async function adjust(delta: number) {
    const next = Math.max(1, night + delta)
    try {
      await patchCampaign(campaign.id, { chronicle_night: next })
      onUpdate(next)
    } catch (err) {
      console.error('Failed to update chronicle night:', err)
    }
  }

  const days = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday']
  const dayLabel = days[(night - 1) % 7]

  return (
    <div className="chronicle-night-tracker">
      <button className="chronicle-btn" onClick={() => adjust(-1)} disabled={night <= 1}>−</button>
      <span className="chronicle-label" title={`Chronicle night ${night}`}>
        Night {night} <span className="chronicle-day">— {dayLabel}</span>
      </span>
      <button className="chronicle-btn" onClick={() => adjust(1)}>+</button>
    </div>
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
  const [xpSuggestionsEvent, setXPSuggestionsEvent] = useState<XPSpendSuggestionsEvent | null>(null)
  const [xpPanelDismissed, setXpPanelDismissed] = useState(false)
  const [suggestingXP, setSuggestingXP] = useState(false)
  const [showTalentsPanel, setShowTalentsPanel] = useState(false)
  const [aiTalentDescs, setAiTalentDescs] = useState<Record<string, string>>({})
  const [rulesetName, setRulesetName] = useState<string | null>(null)
  const scrollRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme)
    localStorage.setItem('theme', theme)
  }, [theme])

  useEffect(() => {
    const rulesetId = ctx?.campaign?.ruleset_id
    if (rulesetId == null) {
      setRulesetName(null)
      return
    }
    fetchRuleset(rulesetId)
      .then((rs) => setRulesetName(rs.name.toLowerCase()))
      .catch(() => setRulesetName(null))
  }, [ctx?.campaign?.ruleset_id])

  useEffect(() => {
    if (rulesetName === 'vtm') {
      // VtM uses a single fixed ambient track regardless of scene tags.
      setAmbientTrack('vtm/ambient')
      return
    }
    const tags = ctx?.session?.scene_tags ?? ''
    const firstTag = tags.split(',').filter(Boolean)[0] ?? null
    setAmbientTrack(firstTag)
  }, [ctx?.session?.scene_tags, rulesetName])

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

  const handleEvent = useCallback((data: unknown) => {
    loadContext()
    const event = data as { type?: string }
    if (!getAudioMuted()) {
      if (event?.type === 'dice_rolled') playDiceRoll()
      else if (event?.type === 'message_created') playNotification()
      else if (event?.type === 'combat_started') playCombatStart()
    }
    if (event?.type === 'xp_spend_suggestions') {
      setXPSuggestionsEvent((data as { payload: XPSpendSuggestionsEvent }).payload)
      setXpPanelDismissed(false)
    }
  }, [loadContext])
  const { lastEvent } = useWebSocket(WS_URL, handleEvent)

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [messages, streamingText])

  // When the talents panel opens, fetch AI descriptions for any talent/power
  // that has no static description.
  useEffect(() => {
    if (!showTalentsPanel || !ctx?.character) return
    let charData: Record<string, unknown> = {}
    try { charData = JSON.parse(ctx.character.data_json || '{}') } catch { /* ignore */ }
    const system = ctx?.campaign?.ruleset_id ? 'wrath_glory' : 'wrath_glory' // best effort
    const allNames: string[] = []
    const talentsStr = String(charData.talents ?? '').trim()
    const powersStr = String(charData.powers ?? '').trim()
    for (const s of [talentsStr, powersStr]) {
      if (s) s.split(/[|\n]/).map(t => t.trim().replace(/^[-•]\s*/, '')).filter(Boolean).forEach(n => allNames.push(n))
    }
    const unknown = allNames.filter(n => !wgTalentDescription(n) && !aiTalentDescs[n])
    if (unknown.length === 0) return
    unknown.forEach(name => {
      fetchTalentDescription(name, system).then(desc => {
        if (desc) setAiTalentDescs(prev => ({ ...prev, [name]: desc }))
      })
    })
  }, [showTalentsPanel, ctx?.character]) // eslint-disable-line react-hooks/exhaustive-deps

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

  const handleSpendXP = useCallback(async (characterId: number, field: string, newValue: number) => {
    const res = await fetch(`/api/characters/${characterId}/advance`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ field, new_value: newValue }),
    })
    if (!res.ok) {
      const text = await res.text()
      throw new Error(text || 'Advance failed')
    }
    loadContext()
  }, [loadContext])

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
        {ctx.character && (
          <button
            className={`h-actions-btn${showTalentsPanel ? ' active' : ''}`}
            onClick={() => setShowTalentsPanel((v) => !v)}
            title="Character talents & psychic powers"
          >
            ✦ Talents
          </button>
        )}
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
        {ctx?.character && aiEnabled && (
          <button
            className={`xp-available-badge${suggestingXP ? ' xp-loading' : ''}`}
            disabled={suggestingXP}
            onClick={async () => {
              if (xpSuggestionsEvent && xpPanelDismissed) {
                setXpPanelDismissed(false)
                return
              }
              setSuggestingXP(true)
              try {
                await suggestAdvances(ctx.character!.id)
                setXpPanelDismissed(false)
              } catch {
                // silently ignore — panel will appear when WS event arrives
              } finally {
                setSuggestingXP(false)
              }
            }}
            title={xpSuggestionsEvent && xpPanelDismissed
              ? `Advancement available — ${xpSuggestionsEvent.current_xp} ${xpSuggestionsEvent.xp_label}`
              : 'Request advancement suggestions'}
          >
            {suggestingXP ? '...' : '⬆ Advance'}
          </button>
        )}
        <AudioControls />
      </header>

      {manageOpen && (
        <ManagePanel
          activeCampaignId={ctx?.campaign?.id ?? null}
          activeCharacterId={ctx?.character?.id ?? null}
          activeSessionId={ctx?.session?.id ?? null}
          initialTab={manageTab}
          onTabChange={setManageTab}
          onClose={() => setManageOpen(false)}
          onContextChanged={() => { loadContext(); setManageOpen(false); setXPSuggestionsEvent(null) }}
          onCampaignActivated={() => { setMessages([]); loadContext(); setManageOpen(false); setXPSuggestionsEvent(null) }}
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

        {/* Talents & Powers Overlay */}
        {showTalentsPanel && ctx.character && (() => {
          let charData: Record<string, unknown> = {}
          try { charData = JSON.parse(ctx.character.data_json || '{}') } catch { /* ignore */ }
          const talents = String(charData.talents ?? '').trim()
          const powers = String(charData.powers ?? '').trim()
          const talentRanks = (charData.talent_ranks ?? {}) as Record<string, number>
          return (
            <div className="talents-overlay">
              <div className="talents-overlay-header">
                <span>Talents &amp; Powers — {ctx.character.name}</span>
                <button onClick={() => setShowTalentsPanel(false)}>×</button>
              </div>
              <div className="talents-overlay-body">
                <div className="talents-section">
                  <div className="talents-section-title">Talents</div>
                  {talents
                    ? talents.split(/[|\n]/).map(s => s.trim()).filter(Boolean).map((t, i) => {
                        const name = t.replace(/^[-•]\s*/, '')
                        const rank = talentRanks[name] ?? 1
                        const desc = wgTalentDescription(name) || aiTalentDescs[name] || ''
                        return (
                          <div key={i} className="talents-entry">
                            <div className="talents-entry-name">
                              {name}{rank > 1 && <span className="talents-rank-badge">Rank {rank}</span>}
                            </div>
                            {desc
                              ? <div className="talents-entry-desc">{desc}</div>
                              : <div className="talents-entry-desc talents-entry-loading">Loading description…</div>
                            }
                          </div>
                        )
                      })
                    : <div className="talents-empty">No talents recorded.</div>
                  }
                </div>
                {powers && (
                  <div className="talents-section">
                    <div className="talents-section-title">Psychic Powers</div>
                    {powers.split(/[|\n]/).map(s => s.trim()).filter(Boolean).map((p, i) => {
                      const name = p.replace(/^[-•]\s*/, '')
                      const desc = wgTalentDescription(name) || aiTalentDescs[name] || ''
                      return (
                        <div key={i} className="talents-entry">
                          <div className="talents-entry-name">{name}</div>
                          {desc
                            ? <div className="talents-entry-desc">{desc}</div>
                            : <div className="talents-entry-desc talents-entry-loading">Loading description…</div>
                          }
                        </div>
                      )
                    })}
                  </div>
                )}
              </div>
            </div>
          )
        })()}

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
            <InventoryPanel
              characterId={ctx.character.id}
              characterCurrencyBalance={ctx.character.currency_balance ?? 0}
              characterCurrencyLabel={ctx.character.currency_label ?? 'Gold'}
              lastEvent={lastEvent}
            />
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
                {ctx.session && rulesetName !== 'vtm' && (
                  <SceneTagPicker
                    session={ctx.session}
                    onUpdate={(tags) => {
                      setCtx(prev => prev && prev.session
                        ? { ...prev, session: { ...prev.session, scene_tags: tags } }
                        : prev
                      )
                    }}
                  />
                )}
                {ctx.campaign && rulesetName === 'vtm' && (
                  <ChronicleNightTracker
                    campaign={ctx.campaign}
                    onUpdate={(night) => {
                      setCtx(prev => prev && prev.campaign
                        ? { ...prev, campaign: { ...prev.campaign, chronicle_night: night } }
                        : prev
                      )
                    }}
                  />
                )}
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
                <ReactMarkdown>{normalizeGMContent(streamingText)}</ReactMarkdown>
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

        <XPSuggestionsPanel
          event={xpPanelDismissed ? null : xpSuggestionsEvent}
          onDismiss={() => { setXPSuggestionsEvent(null); setXpPanelDismissed(false) }}
          onHide={() => setXpPanelDismissed(true)}
          onSpend={handleSpendXP}
        />

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
              <ObjectivesPanel campaignId={ctx?.campaign?.id ?? null} sessionId={ctx?.session?.id ?? null} lastEvent={lastEvent} />
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
