import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, cleanup, waitFor, fireEvent } from '@testing-library/react'
import { WorldNotesPanel } from './WorldNotesPanel'
import type { WorldNote } from './types'

const notes: WorldNote[] = [
  { id: 1, campaign_id: 1, title: 'Tavern', content: 'A seedy place.', category: 'location', tags_json: '[]', created_at: '' },
  { id: 2, campaign_id: 1, title: 'Dragon', content: 'Ancient red dragon.', category: 'npc', tags_json: '[]', created_at: '' },
]

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

describe('WorldNotesPanel', () => {
  it('renders notes returned by fetch', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve(notes) }))
    render(<WorldNotesPanel campaignId={1} />)
    expect(await screen.findByText('Tavern')).toBeInTheDocument()
    expect(screen.getByText('Dragon')).toBeInTheDocument()
  })

  it('shows empty state when no notes', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve([]) }))
    render(<WorldNotesPanel campaignId={1} />)
    expect(await screen.findByText('No notes found.')).toBeInTheDocument()
  })

  it('calls fetch with q param when search input changes', async () => {
    const mockFetch = vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve([]) })
    vi.stubGlobal('fetch', mockFetch)
    render(<WorldNotesPanel campaignId={1} />)
    await screen.findByText('No notes found.')
    fireEvent.change(screen.getByRole('searchbox'), { target: { value: 'tavern' } })
    await waitFor(() => {
      expect(mockFetch).toHaveBeenLastCalledWith('/api/campaigns/1/world-notes?q=tavern')
    })
  })
})
