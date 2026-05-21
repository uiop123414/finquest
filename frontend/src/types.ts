export interface User {
  id: string
  email: string
  xp_total: number
  level: number
  created_at: string
}

export interface Category {
  id: string
  user_id: string | null
  name: string
  is_system: boolean
}

export interface Transaction {
  id: string
  user_id: string
  amount: number
  type: 'income' | 'expense'
  category_id: string | null
  date: string
  note: string
  created_at: string
}

export interface Achievement {
  id: string
  code: string
  name: string
  description: string
  earned_at?: string
}

export interface Goal {
  id: string
  user_id: string
  name: string
  target_amount: number
  current_amount: number
  deadline: string
  completed_at?: string
}

export interface GamificationProfile {
  xp_total: number
  level: number
  level_progress_pct: number
  achievements: Achievement[]
}

export interface AnalyticsSummary {
  income: number
  expense: number
  balance: number
  by_category: { category: string; amount: number }[]
}

export interface OverTimePoint {
  period: string
  income: number
  expense: number
}

export interface AuthResponse {
  user: User
  access_token: string
  refresh_token: string
}
