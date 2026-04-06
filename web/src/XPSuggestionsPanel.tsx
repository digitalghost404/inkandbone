import { useState, useEffect, useRef } from 'react'
import type { XPSpendSuggestionsEvent, XPSuggestion } from './types'

interface Props {
  event: XPSpendSuggestionsEvent | null
  onDismiss: () => void
  onSpend: (characterId: number, field: string, newValue: number) => Promise<void>
}

export function XPSuggestionsPanel({ event, onDismiss, onSpend }: Props) {
  const [error, setError] = useState<string | null>(null)
  const [spending, setSpending] = useState(false)
  const [success, setSuccess] = useState(false)
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current)
    }
  }, [])

  if (!event) return null

  const handleSpend = async (sg: XPSuggestion) => {
    setError(null)
    setSpending(true)
    let succeeded = false
    try {
      await onSpend(event.character_id, sg.field, sg.new_value)
      succeeded = true
      setSuccess(true)
      timerRef.current = setTimeout(() => onDismiss(), 600)
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to spend XP')
    } finally {
      if (!succeeded) setSpending(false)
    }
  }

  return (
    <div className={`xp-suggestions-panel${success ? ' xp-success-flash' : ''}`}>
      <div className="xp-suggestions-header">
        <span>Advancement Suggestions — {event.current_xp} {event.xp_label} available</span>
        <button className="xp-dismiss-btn" onClick={onDismiss} aria-label="Dismiss">×</button>
      </div>

      {success && <div className="xp-success-toast">Spent!</div>}
      {error && <div className="xp-error-toast">{error}</div>}

      <div className="xp-suggestions-list">
        {event.suggestions.map((sg) => (
          <div key={sg.field} className="xp-suggestion-card">
            <div className="xp-suggestion-top">
              <span className="xp-suggestion-name">{sg.display_name}</span>
              <span className="xp-suggestion-arrow">{sg.current_value} → {sg.new_value}</span>
              <span className="xp-cost-badge">{sg.xp_cost} {event.xp_label}</span>
            </div>
            <div className="xp-suggestion-reason">{sg.reasoning}</div>
            <button
              className="xp-spend-btn"
              disabled={spending}
              onClick={() => handleSpend(sg)}
            >
              Spend
            </button>
          </div>
        ))}
      </div>
    </div>
  )
}
