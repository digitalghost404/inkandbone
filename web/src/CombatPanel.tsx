import { useState } from 'react'
import type { CombatSnapshot, Combatant } from './types'
import { patchCombatant, advanceTurn } from './api'

interface Props {
  combat: CombatSnapshot
}

const STANDARD_CONDITIONS = [
  'Poisoned', 'Prone', 'Stunned', 'Blinded',
  'Exhausted', 'Frightened', 'Paralyzed', 'Invisible',
]

function hpBarClass(current: number, max: number): string {
  if (max === 0) return 'hp-bar-green'
  const ratio = current / max
  if (ratio > 0.5) return 'hp-bar-green'
  if (ratio > 0.25) return 'hp-bar-yellow'
  return 'hp-bar-red'
}

function parseConditions(json: string): string[] {
  try {
    return JSON.parse(json) as string[]
  } catch {
    return []
  }
}

function CombatantRow({ c, isActive }: { c: Combatant; isActive: boolean }) {
  const [conditions, setConditions] = useState<string[]>(() => parseConditions(c.conditions_json))
  const [showDropdown, setShowDropdown] = useState(false)

  const pct = c.hp_max > 0 ? Math.max(0, Math.round((c.hp_current / c.hp_max) * 100)) : 0
  const colorClass = hpBarClass(c.hp_current, c.hp_max)

  function removeCondition(cond: string) {
    const next = conditions.filter((x) => x !== cond)
    setConditions(next)
    patchCombatant(c.id, { conditions_json: JSON.stringify(next) }).catch(console.error)
  }

  function addCondition(cond: string) {
    if (conditions.includes(cond)) return
    const next = [...conditions, cond]
    setConditions(next)
    patchCombatant(c.id, { conditions_json: JSON.stringify(next) }).catch(console.error)
    setShowDropdown(false)
  }

  const available = STANDARD_CONDITIONS.filter((s) => !conditions.includes(s))

  return (
    <div
      className={`combatant-card ${c.is_player ? 'player' : 'enemy'} ${isActive ? 'active-turn' : ''}`}
    >
      <div className="combatant-header">
        <span className="combatant-name">{c.name}</span>
        <span className="combatant-init">Init {c.initiative}</span>
      </div>
      <div className="hp-bar-track">
        <div className={`hp-bar-fill ${colorClass}`} style={{ width: `${pct}%` }} />
      </div>
      <div className="hp-label">
        {c.hp_current} / {c.hp_max} HP
      </div>
      {(conditions.length > 0 || true) && (
        <div className="conditions">
          {conditions.map((cond) => (
            <button
              key={cond}
              className="condition-badge condition-badge-btn"
              onClick={() => removeCondition(cond)}
              title={`Remove ${cond}`}
            >
              {cond} ×
            </button>
          ))}
          <div className="condition-add-wrap">
            <button
              className="condition-add-btn"
              onClick={() => setShowDropdown((v) => !v)}
              title="Add condition"
            >
              +
            </button>
            {showDropdown && available.length > 0 && (
              <div className="condition-dropdown">
                {available.map((cond) => (
                  <button
                    key={cond}
                    className="condition-dropdown-item"
                    onClick={() => addCondition(cond)}
                  >
                    {cond}
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

export function CombatPanel({ combat }: Props) {
  const { encounter, combatants } = combat

  async function handleNextTurn() {
    try {
      await advanceTurn(encounter.id)
    } catch (err) {
      console.error('advanceTurn failed:', err)
    }
  }

  return (
    <div className="combat-grimoire">
      <h2>⚔ {encounter.name}</h2>
      {combatants.map((c, idx) => (
        <CombatantRow
          key={c.id}
          c={c}
          isActive={idx === encounter.active_turn_index}
        />
      ))}
      <button className="next-turn-btn" onClick={handleNextTurn}>
        Next Turn →
      </button>
    </div>
  )
}
