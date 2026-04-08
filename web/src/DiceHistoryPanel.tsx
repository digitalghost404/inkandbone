import { useState, useEffect } from 'react'
import { fetchDiceRolls } from './api'
import type { DiceRoll } from './types'

interface Props {
  sessionId: number
  lastEvent: unknown
}

export function DiceHistoryPanel({ sessionId, lastEvent }: Props) {
  const [rolls, setRolls] = useState<DiceRoll[]>([])

  useEffect(() => {
    let ignored = false
    fetchDiceRolls(sessionId)
      .then((data) => { if (!ignored) setRolls(data) })
      .catch(() => { if (!ignored) setRolls([]) })
    return () => { ignored = true }
  }, [sessionId])

  useEffect(() => {
    const ev = lastEvent as { type?: string } | null
    if (ev?.type === 'dice_rolled') {
      fetchDiceRolls(sessionId).then(setRolls).catch(() => {})
    }
  }, [lastEvent, sessionId])

  const recent = rolls.slice(0, 5)

  if (recent.length === 0) return null

  // VtM V5 pool rolls store success counts, not pip sums. Detect by "(xN+yH)" suffix.
  const isVtMPool = (expr: string) => /\(\d+N\+\d+H\)/.test(expr)

  return (
    <div className="dice-compact">
      <div className="dice-compact-label">Dice</div>
      {recent.map((r) => (
        <div key={r.id} className="dice-compact-row">
          <span className="dice-compact-expr">{r.expression}</span>
          <span className="dice-compact-result">
            {isVtMPool(r.expression)
              ? `${r.result} ${r.result === 1 ? 'success' : 'successes'}`
              : r.result}
          </span>
        </div>
      ))}
    </div>
  )
}
