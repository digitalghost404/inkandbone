import { useState, useEffect, useCallback } from 'react'
import { fetchWorldNotes } from './api'
import type { WorldNote } from './types'

interface Props {
  campaignId: number
  lastEvent: unknown
}

function parseTags(json: string): string[] {
  try { return JSON.parse(json) as string[] }
  catch { return [] }
}

export function WorldNotesPanel({ campaignId, lastEvent }: Props) {
  const [notes, setNotes] = useState<WorldNote[]>([])
  const [query, setQuery] = useState('')
  const [activeTag, setActiveTag] = useState<string | null>(null)

  const loadNotes = useCallback(() => {
    fetchWorldNotes(campaignId, query || undefined, activeTag || undefined)
      .then(setNotes)
      .catch(() => setNotes([]))
  }, [campaignId, query, activeTag])

  useEffect(() => {
    loadNotes()
  }, [loadNotes])

  useEffect(() => {
    const ev = lastEvent as { type?: string } | null
    if (ev?.type === 'world_note_updated' || ev?.type === 'world_note_created') {
      loadNotes()
    }
  }, [lastEvent, loadNotes])

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
        notes.map((n) => {
          const tags = parseTags(n.tags_json)
          return (
            <div key={n.id} className="world-note">
              <div className="note-header">
                <span className="note-title">{n.title}</span>
                {n.category && <span className="note-category">{n.category}</span>}
              </div>
              {tags.length > 0 && (
                <div className="tag-pills">
                  {tags.map((tag) => (
                    <button
                      key={tag}
                      className={`tag-pill${activeTag === tag ? ' active' : ''}`}
                      onClick={() => setActiveTag((t) => (t === tag ? null : tag))}
                    >
                      {tag}
                    </button>
                  ))}
                </div>
              )}
              <p className="note-content">{n.content}</p>
            </div>
          )
        })
      )}
    </section>
  )
}
