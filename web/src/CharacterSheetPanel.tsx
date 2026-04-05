import { useState, useEffect, useRef, type ChangeEvent } from 'react'
import { fetchRuleset, patchCharacter, uploadPortrait } from './api'
import type { Ruleset } from './api'
import type { Character } from './types'

interface SchemaField {
  key: string
  label: string
  type: 'text' | 'number' | 'textarea'
}

interface CharacterSheetPanelProps {
  character: Character | null
  rulesetId: number | null
  lastEvent: unknown
}

interface CharacterUpdatedPayload {
  id: number
  data_json?: string
  portrait_path?: string
}

interface CharacterUpdatedEvent {
  type: 'character_updated'
  payload: CharacterUpdatedPayload
}

function isCharacterUpdatedEvent(ev: unknown): ev is CharacterUpdatedEvent {
  if (typeof ev !== 'object' || ev === null) return false
  const e = ev as Record<string, unknown>
  if (e['type'] !== 'character_updated') return false
  const p = e['payload']
  if (typeof p !== 'object' || p === null) return false
  return typeof (p as Record<string, unknown>)['id'] === 'number'
}

const ATTRIBUTE_KEYS = new Set(['edge', 'heart', 'iron', 'shadow', 'wits'])
const TRACK_KEYS     = new Set(['health', 'spirit', 'supply', 'momentum'])

function AttributePips({ value }: { value: number }) {
  return (
    <div className="attr-pips">
      {[1, 2, 3].map((i) => (
        <div key={i} className={`pip ${i <= value ? 'filled' : 'empty'}`} />
      ))}
    </div>
  )
}

function TrackBar({ fieldKey, value }: { fieldKey: string; value: number }) {
  const isMomentum = fieldKey === 'momentum'
  const max = isMomentum ? 10 : 5
  const filled = Math.max(0, Math.min(max, value))
  const colorClass = isMomentum ? 'filled-momentum' : 'filled-health'
  const displayValue = isMomentum ? value : `${value}/${max}`

  return (
    <div className="track-row">
      <div className="track-header">
        <span className="track-label">{fieldKey}</span>
        <span className="track-value">{displayValue}</span>
      </div>
      <div className="track-segments">
        {Array.from({ length: max }, (_, i) => (
          <div
            key={i}
            className={`track-seg ${i < filled ? colorClass : 'empty-seg'}`}
          />
        ))}
      </div>
    </div>
  )
}

export function CharacterSheetPanel({ character, rulesetId, lastEvent }: CharacterSheetPanelProps) {
  const [ruleset, setRuleset] = useState<Ruleset | null>(null)
  const [fields, setFields] = useState<Record<string, string>>({})
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (rulesetId === null) return
    fetchRuleset(rulesetId)
      .then(setRuleset)
      .catch(console.error)
  }, [rulesetId])

  useEffect(() => {
    if (!character) return
    try {
      const data = JSON.parse(character.data_json || '{}') as Record<string, unknown>
      setFields(Object.fromEntries(Object.entries(data).map(([k, v]) => [k, String(v ?? '')])))
    } catch {
      setFields({})
    }
  }, [character?.id])

  useEffect(() => {
    if (!isCharacterUpdatedEvent(lastEvent)) return
    if (lastEvent.payload.id !== character?.id) return
    if (lastEvent.payload.data_json) {
      try {
        const data = JSON.parse(lastEvent.payload.data_json) as Record<string, unknown>
        setFields(Object.fromEntries(Object.entries(data).map(([k, v]) => [k, String(v ?? '')])))
      } catch { /* ignore */ }
    }
  }, [lastEvent, character?.id])

  if (!character) return null

  const schema: SchemaField[] = (() => {
    try {
      const parsed = JSON.parse(ruleset?.schema_json ?? '[]') as unknown
      if (!Array.isArray(parsed) && typeof parsed === 'object' && parsed !== null) {
        const legacy = parsed as Record<string, unknown>
        if (Array.isArray(legacy['fields'])) {
          return (legacy['fields'] as string[]).map((key) => ({
            key,
            label: key.charAt(0).toUpperCase() + key.slice(1).replace(/_/g, ' '),
            type: 'text' as const,
          }))
        }
      }
      if (!Array.isArray(parsed)) return []
      return parsed as SchemaField[]
    } catch {
      return []
    }
  })()

  function handleChange(key: string, value: string) {
    const next = { ...fields, [key]: value }
    setFields(next)
    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => {
      const updates: Record<string, unknown> = {}
      schema.forEach((f) => {
        const v = next[f.key]
        updates[f.key] = f.type === 'number' ? (v === '' ? null : Number(v)) : v
      })
      patchCharacter(character!.id, updates).catch(console.error)
    }, 500)
  }

  function handlePortraitChange(e: ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    uploadPortrait(character!.id, file).catch(console.error)
  }

  const attributeFields = schema.filter((f) => ATTRIBUTE_KEYS.has(f.key))
  const trackFields     = schema.filter((f) => TRACK_KEYS.has(f.key))
  const otherFields     = schema.filter((f) => !ATTRIBUTE_KEYS.has(f.key) && !TRACK_KEYS.has(f.key))

  return (
    <>
      {/* Portrait */}
      <div className="portrait-wrap">
        {character.portrait_path ? (
          <img
            className="portrait-circle"
            src={`/api/files/${character.portrait_path}`}
            alt={character.name}
          />
        ) : (
          <div className="portrait-placeholder-circle">{character.name[0] ?? '?'}</div>
        )}
        <label className="portrait-change">
          <input
            ref={fileInputRef}
            type="file"
            accept="image/*"
            onChange={handlePortraitChange}
            style={{ display: 'none' }}
          />
          Change portrait
        </label>
      </div>

      {/* Attributes — pip dots */}
      {attributeFields.length > 0 && (
        <div>
          {attributeFields.map((f) => (
            <div key={f.key} className="attr-row">
              <span className="attr-label">{f.key}</span>
              <AttributePips value={Number(fields[f.key] ?? 0)} />
            </div>
          ))}
        </div>
      )}

      {/* Tracks — segmented bars */}
      {trackFields.length > 0 && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
          {trackFields.map((f) => (
            <TrackBar
              key={f.key}
              fieldKey={f.key}
              value={Number(fields[f.key] ?? 0)}
            />
          ))}
        </div>
      )}

      {/* Other fields — number fields in 2-col grid, text/textarea full width */}
      {(() => {
        const numFields = otherFields.filter((f) => f.type === 'number')
        const wideFields = otherFields.filter((f) => f.type !== 'number')
        const labelStyle: React.CSSProperties = { fontSize: '9px', textTransform: 'uppercase', letterSpacing: '1.5px', color: 'var(--gold-dim)', fontFamily: 'var(--serif)', display: 'flex', flexDirection: 'column', gap: '2px' }
        const inputStyle: React.CSSProperties = { background: 'var(--surface)', border: '1px solid var(--border)', color: 'var(--text)', fontSize: '12px', padding: '0.15rem 0.3rem', width: '100%' }
        return (
          <>
            {numFields.length > 0 && (
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.35rem 0.5rem' }}>
                {numFields.map((field) => (
                  <label key={field.key} style={labelStyle}>
                    {field.label}
                    <input
                      type="number"
                      value={fields[field.key] ?? ''}
                      onChange={(e) => handleChange(field.key, e.target.value)}
                      style={inputStyle}
                    />
                  </label>
                ))}
              </div>
            )}
            {wideFields.map((field) => (
              <label key={field.key} style={{ ...labelStyle, display: 'flex', flexDirection: 'column', gap: '2px' }}>
                {field.label}
                {field.type === 'textarea' ? (
                  <textarea
                    value={fields[field.key] ?? ''}
                    onChange={(e) => handleChange(field.key, e.target.value)}
                    style={{ ...inputStyle, fontFamily: 'inherit', resize: 'vertical', minHeight: '3rem' }}
                  />
                ) : (
                  <input
                    type="text"
                    value={fields[field.key] ?? ''}
                    onChange={(e) => handleChange(field.key, e.target.value)}
                    style={inputStyle}
                  />
                )}
              </label>
            ))}
          </>
        )
      })()}
    </>
  )
}
