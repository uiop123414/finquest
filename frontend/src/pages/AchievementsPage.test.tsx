import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import AchievementsPage from './AchievementsPage'

vi.mock('../api/client', () => ({
  default: {
    get: vi.fn().mockResolvedValue({
      data: {
        xp_total: 50,
        level: 1,
        level_progress_pct: 50,
        achievements: [
          { id: '1', code: 'first_transaction', name: 'Первый шаг', description: 'Добавьте первую транзакцию', earned_at: '2024-01-01T00:00:00Z' },
          { id: '2', code: 'ten_transactions', name: 'Десятка', description: 'Добавьте 10 транзакций' },
        ],
      },
    }),
  },
}))

describe('AchievementsPage', () => {
  it('shows unlocked and locked achievements', async () => {
    render(<AchievementsPage />)
    expect(await screen.findByText('Первый шаг')).toBeInTheDocument()
    expect(await screen.findByText('Десятка')).toBeInTheDocument()
  })
})
