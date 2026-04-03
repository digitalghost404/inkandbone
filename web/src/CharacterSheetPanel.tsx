import { useState, useEffect, useRef } from 'react'
import { fetchRuleset, patchCharacter, uploadPortrait } from './api'
import type { Character, Ruleset } from './api'

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

export function CharacterSheetPanel({ character, rulesetId, lastEvent }: CharacterSheetPanelProps) {
  const [ruleset, setRuleset] = useState<Ruleset | null>(null)
  const [fields, setFields] = useState<Record<string, string>>({})
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Load ruleset when rulesetId changes
  useEffect(() => {
    if (rulesetId === null) return
    fetchRuleset(rulesetId)
      .then(setRuleset)
      .catch(console.error)
  }, [rulesetId])

  // Sync fields when character changes (reset on new character)
  useEffect(() => {
    if (!character) return
    try {
      const data = JSON.parse(character.data_json || '{}') as Record<string, unknown>
      setFields(Object.fromEntries(Object.entries(data).map(([k, v]) => [k, String(v ?? '')])))
    } catch {
      setFields({})
    }
  }, [character?.id])

  // Apply character_updated WS event
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
      return JSON.parse(ruleset?.schema_json ?? '[]') as SchemaField[]
    } catch {
      return []
    }
  })()

  function handleChange(key: string, value: string) {
    const next = { ...fields, [key]: value }
    setFields(next)
    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => {
      // Convert numeric fields back to numbers before patching
      const updates: Record<string, unknown> = {}
      schema.forEach((f) => {
        const v = next[f.key]
        updates[f.key] = f.type === 'number' ? (v === '' ? null : Number(v)) : v
      })
      patchCharacter(character!.id, updates).catch(console.error)
    }, 500)
  }

  function handlePortraitChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    uploadPortrait(character!.id, file).catch(console.error)
  }

  return (
    <section className="panel character-sheet-panel">
      <h2>Character Sheet — {character.name}</h2>

      <div className="portrait-upload">
        {character.portrait_path ? (
          <img
            className="portrait-large"
            src={`/api/files/${character.portrait_path}`}
            alt={character.name}
          />
        ) : (
          <div className="portrait-placeholder">No portrait</div>
        )}
        <label>
          <input type="file" accept="image/*" onChange={handlePortraitChange} style={{ display: 'none' }} />
          Change portrait
        </label>
      </div>

      {schema.map((field) => (
        <div key={field.key} className="field-row">
          <label htmlFor={`field-${field.key}`}>{field.label}</label>
          {field.type === 'textarea' ? (
            <textarea
              id={`field-${field.key}`}
              value={fields[field.key] ?? ''}
              onChange={(e) => handleChange(field.key, e.target.value)}
            />
          ) : (
            <input
              id={`field-${field.key}`}
              type={field.type}
              value={fields[field.key] ?? ''}
              onChange={(e) => handleChange(field.key, e.target.value)}
            />
          )}
        </div>
      ))}
    </section>
  )
}
