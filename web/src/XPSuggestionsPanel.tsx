import { useState } from 'react'
import type { XPSpendSuggestionsEvent } from './types'

interface Props {
  event: XPSpendSuggestionsEvent | null
  onDismiss: () => void
  onSpend: (characterId: number, field: string, newValue: number) => Promise<void>
}

export function XPSuggestionsPanel({ event, onDismiss, onSpend }: Props) {
  const [error, setError] = useState<string | null>(null)
  const [spending, setSpending] = useState(false)

  if (!event) return null

  const handleSpend = async (field: string, newValue: number) => {
    setError(null)
    setSpending(true)
    try {
      await onSpend(event.character_id, field, newValue)
      onDismiss()
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to spend XP')
    } finally {
      setSpending(false)
    }
  }

  return (
    <div className="xp-suggestions-panel">
      <div className="xp-suggestions-header">
        <span>Advancement Suggestions — {event.current_xp} {event.xp_label} available</span>
        <button className="xp-dismiss-btn" onClick={onDismiss} title="Dismiss">×</button>
      </div>

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
              onClick={() => handleSpend(sg.field, sg.new_value)}
            >
              Spend
            </button>
          </div>
        ))}
      </div>
    </div>
  )
}
