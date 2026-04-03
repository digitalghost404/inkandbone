import { useState, useEffect } from 'react'
import { fetchWorldNotes } from './api'
import type { WorldNote } from './types'

interface Props {
  campaignId: number
}

export function WorldNotesPanel({ campaignId }: Props) {
  const [notes, setNotes] = useState<WorldNote[]>([])
  const [query, setQuery] = useState('')

  useEffect(() => {
    let ignored = false
    fetchWorldNotes(campaignId, query || undefined)
      .then((data) => { if (!ignored) setNotes(data) })
      .catch(() => { if (!ignored) setNotes([]) })
    return () => { ignored = true }
  }, [campaignId, query])

  return (
    <section className="panel world-notes">
      <h2>World Notes</h2>
      <input
        className="search"
        type="search"
        placeholder="Search notes…"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
      />
      {notes.length === 0 ? (
        <p className="empty">No notes found.</p>
      ) : (
        notes.map((n) => (
          <div key={n.id} className="world-note">
            <div className="note-header">
              <span className="note-title">{n.title}</span>
              {n.category && <span className="note-category">{n.category}</span>}
            </div>
            <p className="note-content">{n.content}</p>
          </div>
        ))
      )}
    </section>
  )
}
