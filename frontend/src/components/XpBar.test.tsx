import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import XpBar from './XpBar'

vi.mock('../api/client', () => ({
  default: {
    get: vi.fn().mockResolvedValue({
      data: { xp_total: 150, level: 2, level_progress_pct: 50, achievements: [] },
    }),
  },
}))

describe('XpBar', () => {
  it('renders level and xp', async () => {
    render(<XpBar />)
    expect(await screen.findByText('Lv.2')).toBeInTheDocument()
    expect(await screen.findByText('150 XP')).toBeInTheDocument()
  })

  it('renders progress bar', async () => {
    render(<XpBar />)
    const bar = await screen.findByTestId('xp-bar')
    expect(bar).toBeInTheDocument()
  })
})
