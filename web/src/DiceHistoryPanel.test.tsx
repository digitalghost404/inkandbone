import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, cleanup } from '@testing-library/react'
import { DiceHistoryPanel } from './DiceHistoryPanel'
import type { DiceRoll } from './types'

const rolls: DiceRoll[] = [
  { id: 1, session_id: 1, expression: '1d20+5', result: 18, breakdown_json: '[]', created_at: '' },
  { id: 2, session_id: 1, expression: '2d6', result: 7, breakdown_json: '[]', created_at: '' },
]

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

describe('DiceHistoryPanel', () => {
  it('renders dice rolls', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve(rolls) }))
    render(<DiceHistoryPanel sessionId={1} />)
    expect(await screen.findByText('1d20+5')).toBeInTheDocument()
    expect(screen.getByText('18')).toBeInTheDocument()
    expect(screen.getByText('2d6')).toBeInTheDocument()
    expect(screen.getByText('7')).toBeInTheDocument()
  })

  it('shows empty state when no rolls', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve([]) }))
    render(<DiceHistoryPanel sessionId={1} />)
    expect(await screen.findByText('No rolls yet.')).toBeInTheDocument()
  })
})
