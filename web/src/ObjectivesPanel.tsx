import { useState, useEffect } from 'react'
import { fetchObjectives, patchObjective, deleteObjective } from './api'
import type { Objective } from './types'

interface ObjectivesPanelProps {
  campaignId: number | null
  lastEvent: unknown
}

export function ObjectivesPanel({ campaignId, lastEvent }: ObjectivesPanelProps) {
  const [objectives, setObjectives] = useState<Objective[]>([])

  function load() {
    if (campaignId === null) return
    fetchObjectives(campaignId).then(setObjectives).catch(() => setObjectives([]))
  }

  useEffect(() => {
    load()
  }, [campaignId])

  useEffect(() => {
    const ev = lastEvent as { type?: string } | null
    if (ev?.type === 'objective_updated' && campaignId !== null) {
      load()
    }
  }, [lastEvent, campaignId])

  async function handleStatus(id: number, status: 'active' | 'completed' | 'failed') {
    try {
      await patchObjective(id, status)
      setObjectives((prev) =>
        prev.map((o) => (o.id === id ? { ...o, status } : o))
      )
    } catch (err) {
      console.error(err)
    }
  }

  async function handleDelete(id: number) {
    try {
      await deleteObjective(id)
      setObjectives((prev) => prev.filter((o) => o.id !== id))
    } catch (err) {
      console.error(err)
    }
  }

  if (campaignId === null) return <p className="empty">No active campaign.</p>

  const active = objectives.filter((o) => o.status === 'active')
  const resolved = objectives.filter((o) => o.status === 'completed' || o.status === 'failed')

  if (objectives.length === 0) {
    return (
      <div className="objectives-panel">
        <p className="empty">No objectives yet — the GM will assign them.</p>
      </div>
    )
  }

  return (
    <div className="objectives-panel">
      {active.length > 0 && (
        <section>
          <div className="objectives-section-label">Active</div>
          {active.map((obj) => (
            <ObjectiveCard
              key={obj.id}
              objective={obj}
              onStatus={handleStatus}
              onDelete={handleDelete}
            />
          ))}
        </section>
      )}
      {resolved.length > 0 && (
        <section>
          <div className="objectives-section-label objectives-section-label--dim">Resolved</div>
          {resolved.map((obj) => (
            <ObjectiveCard
              key={obj.id}
              objective={obj}
              onStatus={handleStatus}
              onDelete={handleDelete}
            />
          ))}
        </section>
      )}
    </div>
  )
}

interface CardProps {
  objective: Objective
  onStatus: (id: number, status: 'active' | 'completed' | 'failed') => void
  onDelete: (id: number) => void
}

function ObjectiveCard({ objective, onStatus, onDelete }: CardProps) {
  const { id, title, description, status } = objective
  const isActive = status === 'active'

  return (
    <div className={`objective-card ${status}`}>
      <div className="objective-header">
        <span className={`objective-status objective-status--${status}`}>
          {status === 'active' ? '◈ Active' : status === 'completed' ? '✓ Done' : '✗ Failed'}
        </span>
        <button
          className="objective-delete-btn"
          onClick={() => onDelete(id)}
          title="Delete objective"
          aria-label="Delete objective"
        >
          ×
        </button>
      </div>
      <div className="objective-title">{title}</div>
      {description && <div className="objective-desc">{description}</div>}
      <div className="objective-actions">
        {isActive ? (
          <>
            <button
              className="obj-btn obj-btn--complete"
              onClick={() => onStatus(id, 'completed')}
              title="Mark completed"
              aria-label="Mark completed"
            >
              ✓
            </button>
            <button
              className="obj-btn obj-btn--fail"
              onClick={() => onStatus(id, 'failed')}
              title="Mark failed"
              aria-label="Mark failed"
            >
              ✗
            </button>
          </>
        ) : (
          <button
            className="obj-btn obj-btn--reopen"
            onClick={() => onStatus(id, 'active')}
            title="Reopen"
            aria-label="Reopen objective"
          >
            ↺ Reopen
          </button>
        )}
      </div>
    </div>
  )
}
