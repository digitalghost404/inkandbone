import type { CombatSnapshot } from './types'

interface Props {
  combat: CombatSnapshot
}

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

export function CombatPanel({ combat }: Props) {
  const { encounter, combatants } = combat
  return (
    <section className="panel combat-panel">
      <h2>Combat: {encounter.name}</h2>
      {combatants.map((c, idx) => {
        const pct = c.hp_max > 0 ? Math.max(0, Math.round((c.hp_current / c.hp_max) * 100)) : 0
        const colorClass = hpBarClass(c.hp_current, c.hp_max)
        const conditions = parseConditions(c.conditions_json)
        const isActive = idx === 0
        return (
          <div
            key={c.id}
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
            {conditions.length > 0 && (
              <div className="conditions">
                {conditions.map((cond) => (
                  <span key={cond} className="condition-badge">
                    {cond}
                  </span>
                ))}
              </div>
            )}
          </div>
        )
      })}
    </section>
  )
}
