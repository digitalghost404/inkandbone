import { useState } from 'react'
import { rollDice } from './api'

interface DiceRollerProps {
  sessionId: number
}

const DICE = [4, 6, 8, 10, 12, 20] as const

export function DiceRoller({ sessionId }: DiceRollerProps) {
  const [flash, setFlash] = useState<{ die: number; result: number } | null>(null)

  async function handleRoll(die: number) {
    try {
      const data = await rollDice(sessionId, `d${die}`)
      setFlash({ die, result: data.result })
      setTimeout(() => setFlash(null), 1500)
    } catch (err) {
      console.error(err)
    }
  }

  return (
    <div className="dice-roller">
      <div className="dice-compact-label">Roll Dice</div>
      <div className="dice-roller-buttons">
        {DICE.map((die) => (
          <button
            key={die}
            className={`dice-btn${flash?.die === die ? ' dice-btn-flash' : ''}`}
            onClick={() => handleRoll(die)}
            title={`Roll d${die}`}
          >
            {flash?.die === die ? flash.result : `d${die}`}
          </button>
        ))}
      </div>
    </div>
  )
}
