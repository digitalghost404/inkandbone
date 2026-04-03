import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, cleanup } from '@testing-library/react'
import App from './App'
import type { GameContext } from './types'

const mockCtx: GameContext = {
  campaign: { id: 1, ruleset_id: 1, name: 'Greyhawk', description: '', active: true, created_at: '' },
  character: { id: 1, campaign_id: 1, name: 'Zara', data_json: '{}', portrait_path: '', created_at: '' },
  session: { id: 1, campaign_id: 1, title: 'Session 1', date: '2026-04-03', summary: '', created_at: '' },
  recent_messages: [
    { id: 1, session_id: 1, role: 'assistant', content: 'You enter the tavern.', created_at: '' },
    { id: 2, session_id: 1, role: 'user', content: 'I look for a table.', created_at: '' },
  ],
  active_combat: null,
}

class MockWebSocket {
  onmessage = null; onclose: (() => void) | null = null; onerror = null
  close = vi.fn()
}

describe('App', () => {
  beforeEach(() => {
    vi.stubGlobal('WebSocket', MockWebSocket)
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(mockCtx),
    }))
  })

  afterEach(() => {
    cleanup()
    vi.unstubAllGlobals()
  })

  it('renders campaign name in state bar', async () => {
    render(<App />)
    expect(await screen.findByText('Greyhawk')).toBeInTheDocument()
  })

  it('renders character name in state bar', async () => {
    render(<App />)
    expect(await screen.findByText('Zara')).toBeInTheDocument()
  })

  it('renders session title in state bar', async () => {
    render(<App />)
    expect(await screen.findByText('Session 1')).toBeInTheDocument()
  })

  it('renders session log messages', async () => {
    render(<App />)
    expect(await screen.findByText('You enter the tavern.')).toBeInTheDocument()
    expect(await screen.findByText('I look for a table.')).toBeInTheDocument()
  })
})
