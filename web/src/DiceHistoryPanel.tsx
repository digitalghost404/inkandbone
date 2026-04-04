import { useState, useEffect } from 'react'
import { fetchDiceRolls } from './api'
import type { DiceRoll } from './types'

interface Props {
  sessionId: number
  lastEvent: unknown
}

function parseBreakdown(json: string): number[] {
  try { return JSON.parse(json) as number[] }
  catch { return [] }
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
      fetchDiceRolls(sessionId)
        .then(setRolls)
        .catch(() => {})
    }
  }, [lastEvent, sessionId])

  return (
    <section className="panel dice-history">
      <h2>Dice History</h2>
      {rolls.length === 0 ? (
        <p className="empty">No rolls yet.</p>
      ) : (
        rolls.map((r) => {
          const breakdown = parseBreakdown(r.breakdown_json)
          return (
            <div key={r.id} className="dice-roll">
              <div className="roll-top">
                <span className="expression">{r.expression}</span>
                <span className="result">{r.result}</span>
              </div>
              {breakdown.length > 0 && (
                <div className="breakdown">
                  {breakdown.map((d, i) => (
                    <span key={i} className="die-badge">[{d}]</span>
                  ))}
                </div>
              )}
            </div>
          )
        })
      )}
    </section>
  )
}
