import { useState, useEffect } from 'react'
import { fetchObjectives, patchObjective, deleteObjective, createObjective } from './api'
import type { Objective } from './types'

interface ObjectivesPanelProps {
  campaignId: number | null
  lastEvent: unknown
}

export function ObjectivesPanel({ campaignId, lastEvent }: ObjectivesPanelProps) {
  const [objectives, setObjectives] = useState<Objective[]>([])
  const [subTaskForm, setSubTaskForm] = useState<{ parentId: number; title: string } | null>(null)

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
      // Remove objective and any of its sub-tasks from local state
      setObjectives((prev) => prev.filter((o) => o.id !== id && o.parent_id !== id))
    } catch (err) {
      console.error(err)
    }
  }

  async function handleAddSubTask() {
    if (!subTaskForm || !campaignId || !subTaskForm.title.trim()) return
    try {
      await createObjective(campaignId, subTaskForm.title.trim(), '', subTaskForm.parentId)
      setSubTaskForm(null)
      load()
    } catch (err) {
      console.error(err)
    }
  }

  if (campaignId === null) return <p className="empty">No active campaign.</p>

  const parents = objectives.filter((o) => o.parent_id === null)
  const childrenOf = (parentId: number) => objectives.filter((o) => o.parent_id === parentId)
  const active = parents.filter((o) => o.status === 'active')
  const resolved = parents.filter((o) => o.status !== 'active')

  if (objectives.length === 0) {
    return (
      <div className="objectives-panel">
        <p className="empty">No objectives yet — the GM will assign them.</p>
      </div>
    )
  }

  function renderParent(obj: Objective) {
    const children = childrenOf(obj.id)
    return (
      <div key={obj.id}>
        <ObjectiveCard
          objective={obj}
          onStatus={handleStatus}
          onDelete={handleDelete}
          onAddSubTask={() => setSubTaskForm({ parentId: obj.id, title: '' })}
        />
        {children.length > 0 && (
          <div className="subtask-list">
            {children.map((child) => (
              <ObjectiveCard
                key={child.id}
                objective={child}
                onStatus={handleStatus}
                onDelete={handleDelete}
              />
            ))}
          </div>
        )}
        {subTaskForm?.parentId === obj.id && (
          <div className="subtask-add-form">
            <input
              className="subtask-input"
              value={subTaskForm.title}
              onChange={(e) =>
                setSubTaskForm((prev) => prev ? { ...prev, title: e.target.value } : null)
              }
              onKeyDown={(e) => {
                if (e.key === 'Enter') handleAddSubTask()
                if (e.key === 'Escape') setSubTaskForm(null)
              }}
              placeholder="Sub-task title…"
              autoFocus
            />
            <button className="obj-btn obj-btn--complete" onClick={handleAddSubTask}>Add</button>
            <button className="obj-btn" onClick={() => setSubTaskForm(null)}>✕</button>
          </div>
        )}
      </div>
    )
  }

  return (
    <div className="objectives-panel">
      {active.length > 0 && (
        <section>
          <div className="objectives-section-label">Active</div>
          {active.map(renderParent)}
        </section>
      )}
      {resolved.length > 0 && (
        <section>
          <div className="objectives-section-label objectives-section-label--dim">Resolved</div>
          {resolved.map(renderParent)}
        </section>
      )}
    </div>
  )
}

interface CardProps {
  objective: Objective
  onStatus: (id: number, status: 'active' | 'completed' | 'failed') => void
  onDelete: (id: number) => void
  onAddSubTask?: () => void
}

function ObjectiveCard({ objective, onStatus, onDelete, onAddSubTask }: CardProps) {
  const { id, title, description, status } = objective
  const isActive = status === 'active'

  return (
    <div className={`objective-card ${status}`}>
      <div className="objective-header">
        <span className={`objective-status objective-status--${status}`}>
          {status === 'active' ? '◈ Active' : status === 'completed' ? '✓ Done' : '✗ Failed'}
        </span>
        <div className="objective-header-actions">
          {onAddSubTask && isActive && (
            <button
              className="objective-subtask-btn"
              onClick={onAddSubTask}
              title="Add sub-task"
              aria-label="Add sub-task"
            >
              ＋
            </button>
          )}
          <button
            className="objective-delete-btn"
            onClick={() => onDelete(id)}
            title="Delete objective"
            aria-label="Delete objective"
          >
            ×
          </button>
        </div>
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
