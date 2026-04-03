import { useState, useEffect } from 'react'
import { fetchDiceRolls } from './api'
import type { DiceRoll } from './types'

interface Props {
  sessionId: number
}

export function DiceHistoryPanel({ sessionId }: Props) {
  const [rolls, setRolls] = useState<DiceRoll[]>([])

  useEffect(() => {
    let ignored = false
    fetchDiceRolls(sessionId)
      .then((data) => { if (!ignored) setRolls(data) })
      .catch(() => { if (!ignored) setRolls([]) })
    return () => { ignored = true }
  }, [sessionId])

  return (
    <section className="panel dice-history">
      <h2>Dice History</h2>
      {rolls.length === 0 ? (
        <p className="empty">No rolls yet.</p>
      ) : (
        rolls.map((r) => (
          <div key={r.id} className="dice-roll">
            <span className="expression">{r.expression}</span>
            <span className="result">{r.result}</span>
          </div>
        ))
      )}
    </section>
  )
}
