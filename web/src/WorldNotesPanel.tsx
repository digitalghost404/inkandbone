import { useState, useEffect, useCallback } from 'react'
import { fetchWorldNotes } from './api'
import type { WorldNote } from './types'

interface Props {
  campaignId: number
}

export function WorldNotesPanel({ campaignId }: Props) {
  const [notes, setNotes] = useState<WorldNote[]>([])
  const [query, setQuery] = useState('')

  const load = useCallback(() => {
    fetchWorldNotes(campaignId, query || undefined)
      .then(setNotes)
      .catch(() => setNotes([]))
  }, [campaignId, query])

  useEffect(() => {
    load()
  }, [load])

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
