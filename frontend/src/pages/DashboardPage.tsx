import { useEffect, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  CartesianGrid, Cell, Legend, Line, LineChart,
  Pie, PieChart, ResponsiveContainer, Tooltip, XAxis, YAxis,
} from 'recharts'
import api from '../api/client'
import type { AnalyticsSummary, OverTimePoint } from '../types'

const COLORS = ['#6366f1', '#8b5cf6', '#a78bfa', '#c4b5fd', '#e0e7ff', '#4f46e5', '#4338ca']

// ── Filter options ────────────────────────────────────────────────────────────
type FilterValue = 'all' | '1y' | '6m' | string // string = "YYYY-MM"

function buildFilterOptions() {
  const fixed: { value: FilterValue; label: string }[] = [
    { value: 'all', label: 'Всё время' },
    { value: '1y',  label: 'Последний год' },
    { value: '6m',  label: 'Последние 6 мес.' },
  ]
  const months: { value: FilterValue; label: string }[] = []
  const now = new Date()
  for (let i = 0; i < 12; i++) {
    const d = new Date(now.getFullYear(), now.getMonth() - i, 1)
    const value = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
    const label = d.toLocaleDateString('ru-RU', { month: 'long', year: 'numeric' })
    months.push({ value, label })
  }
  return [...fixed, ...months]
}

const FILTER_OPTIONS = buildFilterOptions()

function isMonthFilter(v: FilterValue) {
  return /^\d{4}-\d{2}$/.test(v)
}

function assess(income: number, expense: number) {
  if (income === 0 && expense === 0) return { text: 'Нет данных за период', color: 'text-gray-400', emoji: '📭' }
  const r = income > 0 ? (income - expense) / income : -1
  if (r >= 0.3) return { text: `Отлично! Норма сбережений ${(r * 100).toFixed(0)}% — так держать.`, color: 'text-green-700', emoji: '🏆' }
  if (r >= 0.1) return { text: `Хорошо. Норма сбережений ${(r * 100).toFixed(0)}% — есть куда расти.`, color: 'text-indigo-700', emoji: '👍' }
  if (r >= 0)   return { text: 'Расходы почти равны доходам. Сократите необязательные траты на 10–15%.', color: 'text-orange-600', emoji: '⚠️' }
  return { text: `Расходы превышают доходы на ${Math.abs(r * 100).toFixed(0)}%. Нужен бюджет.`, color: 'text-red-600', emoji: '🚨' }
}

export default function DashboardPage() {
  const [filter, setFilter] = useState<FilterValue>('1y')
  const [summary, setSummary] = useState<AnalyticsSummary | null>(null)
  const [overtime, setOvertime] = useState<OverTimePoint[]>([])
  const [error, setError] = useState('')

  // ── Summary: use period= for specific month, range= otherwise ──────────────
  useEffect(() => {
    setError('')
    setSummary(null)
    const params = isMonthFilter(filter)
      ? { period: filter }
      : filter === 'all' ? {} : { range: filter }
    api.get<AnalyticsSummary>('/analytics/summary', { params })
      .then((r) => setSummary(r.data))
      .catch(() => setError('Не удалось загрузить данные'))
  }, [filter])

  // ── Chart: period= for daily, range= for monthly ───────────────────────────
  useEffect(() => {
    const params = isMonthFilter(filter)
      ? { period: filter }                          // daily for this month
      : filter === 'all' ? { range: 'all' }
      : { range: filter }                           // 6m / 1y monthly
    api.get<OverTimePoint[]>('/analytics/over-time', { params })
      .then((r) => setOvertime(r.data))
      .catch(() => setOvertime([]))
  }, [filter])

  const fmt = (n: number) =>
    n.toLocaleString('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 0 })

  if (error) return <div className="text-red-400 bg-red-50 rounded-2xl p-6 text-sm">{error}</div>
  if (!summary) return <div className="text-gray-400 text-sm">Загрузка...</div>

  const isEmpty = summary.income === 0 && summary.expense === 0
  const assessment = assess(summary.income, summary.expense)
  const topCat = summary.by_category[0]
  const savingsRate = summary.income > 0
    ? ((summary.income - summary.expense) / summary.income * 100).toFixed(0)
    : null

  const chartLabel = isMonthFilter(filter)
    ? `Динамика по дням — ${filter}`
    : filter === '6m' ? 'Динамика за 6 месяцев'
    : filter === '1y' ? 'Динамика за год'
    : 'Динамика за всё время'

  const xAxisLabel = isMonthFilter(filter) ? 'День' : 'Месяц'

  return (
    <div className="space-y-6">
      {/* Header + filter */}
      <div className="flex items-center justify-between flex-wrap gap-3">
        <h1 className="text-2xl font-bold text-gray-800">Дашборд</h1>
        <select
          value={filter}
          onChange={(e) => setFilter(e.target.value as FilterValue)}
          className="border border-gray-200 rounded-lg px-3 py-1.5 text-sm text-gray-600 bg-white"
        >
          {FILTER_OPTIONS.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
        </select>
      </div>

      {/* KPI */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        <StatCard label="Доходы"   value={fmt(summary.income)}  color="text-green-600" />
        <StatCard label="Расходы"  value={fmt(summary.expense)} color="text-red-500" />
        <StatCard label="Баланс"   value={fmt(summary.balance)} color={summary.balance >= 0 ? 'text-indigo-600' : 'text-red-500'} />
        <StatCard label="Норма сбережений" value={savingsRate !== null ? `${savingsRate}%` : '—'} color="text-gray-700" />
      </div>

      {/* Assessment */}
      <div className="bg-white rounded-2xl p-4 shadow-sm border border-gray-100 flex items-center gap-3">
        <span className="text-2xl">{assessment.emoji}</span>
        <p className={`text-sm font-medium ${assessment.color}`}>{assessment.text}</p>
      </div>

      {isEmpty ? (
        <div className="bg-white rounded-2xl p-10 shadow-sm border border-gray-100 text-center">
          <p className="text-4xl mb-3">💰</p>
          <p className="font-semibold text-gray-700 mb-1">Транзакций пока нет</p>
          <p className="text-sm text-gray-400 mb-5">Добавьте первую или импортируйте CSV</p>
          <div className="flex justify-center gap-3">
            <Link to="/transactions" className="bg-indigo-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-indigo-700">+ Добавить</Link>
            <Link to="/import" className="bg-gray-100 text-gray-600 px-4 py-2 rounded-lg text-sm font-medium hover:bg-gray-200">Импорт CSV</Link>
          </div>
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Pie chart */}
          <div className="bg-white rounded-2xl p-5 shadow-sm border border-gray-100">
            <h2 className="text-sm font-semibold text-gray-700 mb-4">Расходы по категориям</h2>
            <div className="flex gap-4 items-center">
              <ResponsiveContainer width={180} height={180}>
                <PieChart>
                  <Pie data={summary.by_category} dataKey="amount" nameKey="category" cx="50%" cy="50%" outerRadius={80} innerRadius={40}>
                    {summary.by_category.map((_, i) => <Cell key={i} fill={COLORS[i % COLORS.length]} />)}
                  </Pie>
                  <Tooltip formatter={(v: number) => fmt(v)} />
                </PieChart>
              </ResponsiveContainer>
              <ul className="space-y-1.5 text-xs flex-1">
                {summary.by_category.slice(0, 7).map((item, i) => (
                  <li key={i} className="flex items-center gap-2">
                    <span className="w-2.5 h-2.5 rounded-full flex-shrink-0" style={{ background: COLORS[i % COLORS.length] }} />
                    <span className="text-gray-600 truncate flex-1">{item.category}</span>
                    <span className="font-medium text-gray-700">{fmt(item.amount)}</span>
                  </li>
                ))}
              </ul>
            </div>
          </div>

          {/* Insight */}
          <div className="bg-white rounded-2xl p-5 shadow-sm border border-gray-100 flex flex-col gap-4">
            <h2 className="text-sm font-semibold text-gray-700">Детали</h2>
            {topCat && (
              <div className="bg-indigo-50 rounded-xl p-4">
                <p className="text-xs text-indigo-500 mb-0.5">Главная статья расходов</p>
                <p className="font-bold text-indigo-800 text-lg">{topCat.category}</p>
                <p className="text-sm text-indigo-600">{fmt(topCat.amount)}</p>
                {summary.expense > 0 && (
                  <p className="text-xs text-indigo-400 mt-1">{(topCat.amount / summary.expense * 100).toFixed(0)}% от всех расходов</p>
                )}
              </div>
            )}
            <div className="grid grid-cols-2 gap-3">
              <div className="bg-green-50 rounded-xl p-3 text-center">
                <p className="text-xs text-green-500">{isMonthFilter(filter) ? 'Доход за месяц' : 'Среднемес. доход'}</p>
                <p className="font-bold text-green-700 text-sm">
                  {overtime.length
                    ? fmt(isMonthFilter(filter)
                        ? overtime.reduce((s, p) => s + p.income, 0)
                        : overtime.reduce((s, p) => s + p.income, 0) / overtime.length)
                    : '—'}
                </p>
              </div>
              <div className="bg-red-50 rounded-xl p-3 text-center">
                <p className="text-xs text-red-400">{isMonthFilter(filter) ? 'Расход за месяц' : 'Среднемес. расход'}</p>
                <p className="font-bold text-red-600 text-sm">
                  {overtime.length
                    ? fmt(isMonthFilter(filter)
                        ? overtime.reduce((s, p) => s + p.expense, 0)
                        : overtime.reduce((s, p) => s + p.expense, 0) / overtime.length)
                    : '—'}
                </p>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Line chart — period-aware */}
      {overtime.length > 0 && (
        <div className="bg-white rounded-2xl p-5 shadow-sm border border-gray-100">
          <h2 className="text-sm font-semibold text-gray-700 mb-4">
            {chartLabel}
            <span className="ml-2 text-xs text-gray-400 font-normal">({overtime.length} {isMonthFilter(filter) ? 'дн.' : 'мес.'})</span>
          </h2>
          <ResponsiveContainer width="100%" height={220}>
            <LineChart data={overtime} margin={{ top: 4, right: 16, left: 0, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
              <XAxis dataKey="period" tick={{ fontSize: 10 }} label={{ value: xAxisLabel, position: 'insideBottomRight', offset: -4, fontSize: 10 }} />
              <YAxis tick={{ fontSize: 11 }} tickFormatter={(v) => `${(v / 1000).toFixed(0)}k`} />
              <Tooltip formatter={(v: number) => fmt(v)} />
              <Legend iconType="circle" wrapperStyle={{ fontSize: 12 }} />
              <Line type="monotone" dataKey="income"  name="Доходы"  stroke="#22c55e" strokeWidth={2} dot={overtime.length < 15} />
              <Line type="monotone" dataKey="expense" name="Расходы" stroke="#ef4444" strokeWidth={2} dot={overtime.length < 15} />
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}

      {/* AI advice */}
      <AIAdviceWidget />
    </div>
  )
}

function StatCard({ label, value, color }: { label: string; value: string; color: string }) {
  return (
    <div className="bg-white rounded-2xl p-4 shadow-sm border border-gray-100">
      <p className="text-xs text-gray-500 mb-1">{label}</p>
      <p className={`text-lg font-bold ${color}`}>{value}</p>
    </div>
  )
}

function AIAdviceWidget() {
  const [advice, setAdvice] = useState('')
  const [loading, setLoading] = useState(false)
  const [loaded, setLoaded] = useState(false)
  const abortRef = useRef<AbortController | null>(null)

  function load() {
    if (loading) return
    abortRef.current?.abort()
    abortRef.current = new AbortController()
    setLoading(true)
    api.get<{ advice: string }>('/ai/advice', { signal: abortRef.current.signal })
      .then((r) => { setAdvice(r.data.advice); setLoaded(true) })
      .catch(() => setAdvice('Не удалось получить совет. Проверьте ANTHROPIC_API_KEY.'))
      .finally(() => setLoading(false))
  }

  return (
    <div className="bg-white rounded-2xl p-5 shadow-sm border border-gray-100">
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <span className="text-lg">🤖</span>
          <h2 className="text-sm font-semibold text-gray-700">Мнение нейросети</h2>
          <span className="text-xs text-gray-400">учитывает транзакции, депозиты, кредиты и цели</span>
        </div>
        <button
          onClick={load}
          disabled={loading}
          className="text-xs bg-indigo-50 text-indigo-600 px-3 py-1.5 rounded-lg hover:bg-indigo-100 disabled:opacity-50 font-medium"
        >
          {loading ? 'Анализирую...' : loaded ? 'Обновить' : 'Получить совет'}
        </button>
      </div>
      {advice ? (
        <p className="text-sm text-gray-700 leading-relaxed whitespace-pre-line">{advice}</p>
      ) : (
        <p className="text-sm text-gray-400">
          Нажмите кнопку — нейросеть проанализирует транзакции, депозиты, кредиты и цели за последние 30 дней и даст персональный совет.
        </p>
      )}
    </div>
  )
}
