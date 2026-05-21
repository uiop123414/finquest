import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  Cell, Pie, PieChart, ResponsiveContainer, Tooltip,
  LineChart, Line, XAxis, YAxis, CartesianGrid, Legend,
} from 'recharts'
import api from '../api/client'
import type { AnalyticsSummary, OverTimePoint } from '../types'

const COLORS = ['#6366f1', '#8b5cf6', '#a78bfa', '#c4b5fd', '#e0e7ff', '#4f46e5', '#4338ca']

function buildMonthOptions() {
  const opts: { value: string; label: string }[] = [{ value: '', label: 'Всё время' }]
  const now = new Date()
  for (let i = 0; i < 12; i++) {
    const d = new Date(now.getFullYear(), now.getMonth() - i, 1)
    const value = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
    const label = d.toLocaleDateString('ru-RU', { month: 'long', year: 'numeric' })
    opts.push({ value, label })
  }
  return opts
}

const MONTHS = buildMonthOptions()

function assess(income: number, expense: number): { text: string; color: string; emoji: string } {
  if (income === 0 && expense === 0) return { text: 'Нет данных за период', color: 'text-gray-400', emoji: '📭' }
  const savingsRate = income > 0 ? (income - expense) / income : -1
  if (savingsRate >= 0.3) return { text: `Отлично! Вы откладываете ${(savingsRate * 100).toFixed(0)}% доходов — так держать.`, color: 'text-green-700', emoji: '🏆' }
  if (savingsRate >= 0.1) return { text: `Хорошо. Норма сбережений ${(savingsRate * 100).toFixed(0)}% — есть куда расти.`, color: 'text-indigo-700', emoji: '👍' }
  if (savingsRate >= 0) return { text: `Осторожно: вы тратите почти весь доход. Попробуйте сократить расходы.`, color: 'text-orange-600', emoji: '⚠️' }
  return { text: `Расходы превышают доходы на ${Math.abs(savingsRate * 100).toFixed(0)}%. Обратите внимание на бюджет.`, color: 'text-red-600', emoji: '🚨' }
}

export default function DashboardPage() {
  const [period, setPeriod] = useState('')
  const [summary, setSummary] = useState<AnalyticsSummary | null>(null)
  const [overtime, setOvertime] = useState<OverTimePoint[]>([])
  const [error, setError] = useState('')

  useEffect(() => {
    const params = period ? { period } : {}
    setError('')
    setSummary(null)
    api.get<AnalyticsSummary>('/analytics/summary', { params })
      .then((r) => setSummary(r.data))
      .catch(() => setError('Не удалось загрузить данные'))
  }, [period])

  useEffect(() => {
    api.get<OverTimePoint[]>('/analytics/over-time')
      .then((r) => setOvertime([...r.data].reverse()))
      .catch(() => {})
  }, [])

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

  return (
    <div className="space-y-6">
      {/* Header + period filter */}
      <div className="flex items-center justify-between flex-wrap gap-3">
        <h1 className="text-2xl font-bold text-gray-800">Дашборд</h1>
        <select
          value={period}
          onChange={(e) => setPeriod(e.target.value)}
          className="border border-gray-200 rounded-lg px-3 py-1.5 text-sm text-gray-600 bg-white"
        >
          {MONTHS.map((m) => <option key={m.value} value={m.value}>{m.label}</option>)}
        </select>
      </div>

      {/* KPI cards */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        <StatCard label="Доходы" value={fmt(summary.income)} color="text-green-600" />
        <StatCard label="Расходы" value={fmt(summary.expense)} color="text-red-500" />
        <StatCard label="Баланс" value={fmt(summary.balance)} color={summary.balance >= 0 ? 'text-indigo-600' : 'text-red-500'} />
        <StatCard label="Норма сбережений" value={savingsRate !== null ? `${savingsRate}%` : '—'} color="text-gray-700" />
      </div>

      {/* Assessment banner */}
      <div className={`bg-white rounded-2xl p-4 shadow-sm border border-gray-100 flex items-center gap-3`}>
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
                    {summary.by_category.map((_, i) => (
                      <Cell key={i} fill={COLORS[i % COLORS.length]} />
                    ))}
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

          {/* Top category insight */}
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
                <p className="text-xs text-green-500">Среднемес. доход</p>
                <p className="font-bold text-green-700">{overtime.length ? fmt(overtime.reduce((s, p) => s + p.income, 0) / overtime.length) : '—'}</p>
              </div>
              <div className="bg-red-50 rounded-xl p-3 text-center">
                <p className="text-xs text-red-400">Среднемес. расход</p>
                <p className="font-bold text-red-600">{overtime.length ? fmt(overtime.reduce((s, p) => s + p.expense, 0) / overtime.length) : '—'}</p>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Line chart — always shown if we have history */}
      {overtime.length > 0 && (
        <div className="bg-white rounded-2xl p-5 shadow-sm border border-gray-100">
          <h2 className="text-sm font-semibold text-gray-700 mb-4">Динамика за последние {overtime.length} мес.</h2>
          <ResponsiveContainer width="100%" height={220}>
            <LineChart data={overtime} margin={{ top: 4, right: 16, left: 0, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
              <XAxis dataKey="period" tick={{ fontSize: 11 }} />
              <YAxis tick={{ fontSize: 11 }} tickFormatter={(v) => `${(v / 1000).toFixed(0)}k`} />
              <Tooltip formatter={(v: number) => fmt(v)} />
              <Legend iconType="circle" wrapperStyle={{ fontSize: 12 }} />
              <Line type="monotone" dataKey="income" name="Доходы" stroke="#22c55e" strokeWidth={2} dot={false} />
              <Line type="monotone" dataKey="expense" name="Расходы" stroke="#ef4444" strokeWidth={2} dot={false} />
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}
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
