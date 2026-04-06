import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor, cleanup } from '@testing-library/react'
import { XPSuggestionsPanel } from './XPSuggestionsPanel'
import type { XPSpendSuggestionsEvent } from './types'

beforeEach(() => {
  vi.restoreAllMocks()
})

afterEach(() => {
  cleanup()
})

const mockEvent: XPSpendSuggestionsEvent = {
  character_id: 7,
  character_name: 'Brother Cato',
  current_xp: 20,
  xp_label: 'XP',
  suggestions: [
    {
      field: 'toughness',
      display_name: 'Toughness',
      current_value: 4,
      new_value: 5,
      xp_cost: 20,
      reasoning: 'More wounds and resilience.',
    },
    {
      field: 'talent:Iron Will',
      display_name: 'Iron Will (Talent)',
      current_value: 0,
      new_value: 1,
      xp_cost: 20,
      reasoning: 'Reinforces your frontline build.',
    },
  ],
}

describe('XPSuggestionsPanel', () => {
  it('renders nothing when event is null', () => {
    const { container } = render(
      <XPSuggestionsPanel event={null} onDismiss={() => {}} onSpend={() => Promise.resolve()} />
    )
    expect(container.firstChild).toBeNull()
  })

  it('renders suggestion cards', () => {
    render(
      <XPSuggestionsPanel event={mockEvent} onDismiss={() => {}} onSpend={() => Promise.resolve()} />
    )
    expect(screen.getByText('Toughness')).toBeTruthy()
    expect(screen.getByText('Iron Will (Talent)')).toBeTruthy()
    expect(screen.getAllByText('Spend').length).toBe(2)
  })

  it('shows current → new value arrow', () => {
    render(
      <XPSuggestionsPanel event={mockEvent} onDismiss={() => {}} onSpend={() => Promise.resolve()} />
    )
    expect(screen.getByText('4 → 5')).toBeTruthy()
  })

  it('shows XP cost badge', () => {
    render(
      <XPSuggestionsPanel event={mockEvent} onDismiss={() => {}} onSpend={() => Promise.resolve()} />
    )
    expect(screen.getAllByText(/20 XP/).length).toBeGreaterThan(0)
  })

  it('calls onDismiss when X is clicked', () => {
    const onDismiss = vi.fn()
    render(
      <XPSuggestionsPanel event={mockEvent} onDismiss={onDismiss} onSpend={() => Promise.resolve()} />
    )
    fireEvent.click(screen.getByTitle('Dismiss'))
    expect(onDismiss).toHaveBeenCalledOnce()
  })

  it('calls onSpend with correct args and dismisses on success', async () => {
    const onSpend = vi.fn().mockResolvedValue(undefined)
    const onDismiss = vi.fn()
    render(
      <XPSuggestionsPanel event={mockEvent} onDismiss={onDismiss} onSpend={onSpend} />
    )
    fireEvent.click(screen.getAllByText('Spend')[0])
    await waitFor(() => {
      expect(onSpend).toHaveBeenCalledWith(7, 'toughness', 5)
      expect(onDismiss).toHaveBeenCalledOnce()
    })
  })

  it('shows error toast on spend failure', async () => {
    const onSpend = vi.fn().mockRejectedValue(new Error('not enough XP'))
    const onDismiss = vi.fn()
    render(
      <XPSuggestionsPanel event={mockEvent} onDismiss={onDismiss} onSpend={onSpend} />
    )
    fireEvent.click(screen.getAllByText('Spend')[0])
    await waitFor(() => {
      expect(screen.getByText(/not enough XP/i)).toBeTruthy()
    })
    expect(onDismiss).not.toHaveBeenCalled()
  })
})
