import { useEffect, useState } from 'react'
import api from '../api/client'
import type { Category, Transaction } from '../types'

const PAGE_SIZE = 20

export default function TransactionsPage() {
  const [txs, setTxs] = useState<Transaction[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [filterCat, setFilterCat] = useState('')
  const [page, setPage] = useState(0)
  const [hasMore, setHasMore] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState({ amount: '', type: 'expense', category_id: '', date: '', note: '' })

  function load(p: number, cat: string) {
    api.get<Transaction[]>('/transactions', {
      params: {
        limit: PAGE_SIZE + 1, // fetch one extra to know if there's a next page
        offset: p * PAGE_SIZE,
        ...(cat ? { category_id: cat } : {}),
      },
    }).then((r) => {
      const data = r.data
      setHasMore(data.length > PAGE_SIZE)
      setTxs(data.slice(0, PAGE_SIZE))
    })
  }

  useEffect(() => {
    api.get<Category[]>('/categories').then((r) => setCategories(r.data))
  }, [])

  useEffect(() => {
    setPage(0)
    load(0, filterCat)
  }, [filterCat])

  useEffect(() => {
    load(page, filterCat)
  }, [page])

  async function addTransaction(e: React.FormEvent) {
    e.preventDefault()
    await api.post('/transactions', {
      amount: parseFloat(form.amount),
      type: form.type,
      category_id: form.category_id || null,
      date: form.date,
      note: form.note,
    })
    setShowForm(false)
    setForm({ amount: '', type: 'expense', category_id: '', date: '', note: '' })
    setPage(0)
    load(0, filterCat)
  }

  async function deleteTransaction(id: string) {
    await api.delete(`/transactions/${id}`)
    load(page, filterCat)
  }

  const catName = (id: string | null) => categories.find((c) => c.id === id)?.name ?? '—'
  const fmt = (n: number) => n.toLocaleString('ru-RU', { maximumFractionDigits: 0 }) + ' ₽'

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-800">Транзакции</h1>
        <button onClick={() => setShowForm(true)} className="bg-indigo-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-indigo-700">
          + Добавить
        </button>
      </div>

      {/* Filters */}
      <div className="flex gap-2 items-center">
        <label className="text-sm text-gray-500">Категория:</label>
        <select value={filterCat} onChange={(e) => setFilterCat(e.target.value)} className="border border-gray-200 rounded-lg px-3 py-1.5 text-sm">
          <option value="">Все</option>
          {categories.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
        </select>
      </div>

      {/* Add form modal */}
      {showForm && (
        <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-40">
          <form onSubmit={addTransaction} className="bg-white rounded-2xl p-6 w-full max-w-md shadow-xl space-y-3">
            <h2 className="font-semibold text-gray-800">Новая транзакция</h2>
            <input type="number" placeholder="Сумма" value={form.amount} onChange={(e) => setForm({ ...form, amount: e.target.value })} className="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm" required />
            <select value={form.type} onChange={(e) => setForm({ ...form, type: e.target.value })} className="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm">
              <option value="expense">Расход</option>
              <option value="income">Доход</option>
            </select>
            <select value={form.category_id} onChange={(e) => setForm({ ...form, category_id: e.target.value })} className="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm">
              <option value="">Без категории</option>
              {categories.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
            </select>
            <input type="date" value={form.date} onChange={(e) => setForm({ ...form, date: e.target.value })} className="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm" required />
            <input type="text" placeholder="Заметка" value={form.note} onChange={(e) => setForm({ ...form, note: e.target.value })} className="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm" />
            <div className="flex gap-2 pt-1">
              <button type="submit" className="flex-1 bg-indigo-600 text-white py-2 rounded-lg text-sm font-medium">Добавить</button>
              <button type="button" onClick={() => setShowForm(false)} className="flex-1 bg-gray-100 text-gray-600 py-2 rounded-lg text-sm">Отмена</button>
            </div>
          </form>
        </div>
      )}

      {/* Table */}
      <div className="bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 text-gray-500 text-xs uppercase">
            <tr>
              <th className="px-4 py-3 text-left">Дата</th>
              <th className="px-4 py-3 text-left">Заметка</th>
              <th className="px-4 py-3 text-left">Категория</th>
              <th className="px-4 py-3 text-right">Сумма</th>
              <th className="px-4 py-3" />
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {txs.map((tx) => (
              <tr key={tx.id} className="hover:bg-gray-50">
                <td className="px-4 py-3 text-gray-500">{tx.date.slice(0, 10)}</td>
                <td className="px-4 py-3 text-gray-700">{tx.note || '—'}</td>
                <td className="px-4 py-3 text-gray-500">{catName(tx.category_id)}</td>
                <td className={`px-4 py-3 text-right font-medium ${tx.type === 'income' ? 'text-green-600' : 'text-red-500'}`}>
                  {tx.type === 'income' ? '+' : '-'}{fmt(tx.amount)}
                </td>
                <td className="px-4 py-3 text-right">
                  <button onClick={() => deleteTransaction(tx.id)} className="text-gray-300 hover:text-red-400 text-xs">✕</button>
                </td>
              </tr>
            ))}
            {txs.length === 0 && (
              <tr><td colSpan={5} className="px-4 py-8 text-center text-gray-400">Нет транзакций</td></tr>
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      <div className="flex items-center justify-between text-sm text-gray-500">
        <button
          onClick={() => setPage((p) => Math.max(0, p - 1))}
          disabled={page === 0}
          className="px-4 py-2 rounded-lg border border-gray-200 disabled:opacity-30 hover:bg-gray-50 disabled:cursor-default"
        >
          ← Назад
        </button>
        <span>Страница {page + 1}</span>
        <button
          onClick={() => setPage((p) => p + 1)}
          disabled={!hasMore}
          className="px-4 py-2 rounded-lg border border-gray-200 disabled:opacity-30 hover:bg-gray-50 disabled:cursor-default"
        >
          Вперёд →
        </button>
      </div>
    </div>
  )
}
