import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, cleanup, waitFor, fireEvent } from '@testing-library/react'
import { JournalPanel } from './JournalPanel'

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

describe('JournalPanel', () => {
  it('renders nothing when session is null', () => {
    const { container } = render(
      <JournalPanel session={null} lastEvent={null} aiEnabled={false} />
    )
    expect(container.firstChild).toBeNull()
  })

  it('renders textarea with session summary', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve({}) }))
    render(
      <JournalPanel
        session={{ id: 1, summary: 'We fought a dragon.' }}
        lastEvent={null}
        aiEnabled={false}
      />
    )
    const textarea = screen.getByRole('textbox') as HTMLTextAreaElement
    expect(textarea.value).toBe('We fought a dragon.')
  })

  it('calls patchSessionSummary on blur', async () => {
    const mockFetch = vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve({}) })
    vi.stubGlobal('fetch', mockFetch)
    render(
      <JournalPanel
        session={{ id: 42, summary: 'Initial summary.' }}
        lastEvent={null}
        aiEnabled={false}
      />
    )
    const textarea = screen.getByRole('textbox')
    fireEvent.change(textarea, { target: { value: 'Updated summary.' } })
    fireEvent.blur(textarea)
    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/sessions/42',
        expect.objectContaining({
          method: 'PATCH',
          body: JSON.stringify({ summary: 'Updated summary.' }),
        })
      )
    })
  })

  it('generate recap button is disabled when aiEnabled=false', () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve({}) }))
    render(
      <JournalPanel
        session={{ id: 1, summary: 'Some summary.' }}
        lastEvent={null}
        aiEnabled={false}
      />
    )
    const button = screen.getByRole('button', { name: /generate recap/i })
    expect(button).toBeDisabled()
  })

  it('generate recap button calls generateRecap and updates textarea when aiEnabled=true', async () => {
    const mockFetch = vi.fn()
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve({ summary: 'AI generated recap.' }) }) // generateRecap
    vi.stubGlobal('fetch', mockFetch)
    render(
      <JournalPanel
        session={{ id: 5, summary: 'Old summary.' }}
        lastEvent={null}
        aiEnabled={true}
      />
    )
    const button = screen.getByRole('button', { name: /generate recap/i })
    expect(button).not.toBeDisabled()
    fireEvent.click(button)
    await waitFor(() => {
      const textarea = screen.getByRole('textbox') as HTMLTextAreaElement
      expect(textarea.value).toBe('AI generated recap.')
    })
    expect(mockFetch).toHaveBeenCalledWith(
      '/api/sessions/5/recap',
      expect.objectContaining({ method: 'POST' })
    )
    expect(mockFetch).toHaveBeenCalledTimes(1)
  })

  it('session_updated WS event updates draft', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve({}) }))
    const { rerender } = render(
      <JournalPanel
        session={{ id: 3, summary: 'Initial.' }}
        lastEvent={null}
        aiEnabled={false}
      />
    )
    rerender(
      <JournalPanel
        session={{ id: 3, summary: 'Initial.' }}
        lastEvent={{ type: 'session_updated', payload: { session_id: 3, summary: 'WS updated summary.' } }}
        aiEnabled={false}
      />
    )
    await waitFor(() => {
      const textarea = screen.getByRole('textbox') as HTMLTextAreaElement
      expect(textarea.value).toBe('WS updated summary.')
    })
  })

  it('session_updated event with wrong session_id does not update draft', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve({}) }))
    const { rerender } = render(
      <JournalPanel
        session={{ id: 3, summary: 'Initial.' }}
        lastEvent={null}
        aiEnabled={false}
      />
    )
    rerender(
      <JournalPanel
        session={{ id: 3, summary: 'Initial.' }}
        lastEvent={{ type: 'session_updated', payload: { session_id: 99, summary: 'Other session.' } }}
        aiEnabled={false}
      />
    )
    const textarea = screen.getByRole('textbox') as HTMLTextAreaElement
    expect(textarea.value).toBe('Initial.')
  })
})
