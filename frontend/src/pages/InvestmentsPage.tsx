import { useEffect, useState } from 'react'
import api from '../api/client'
import type { Deposit } from '../types'

type Modal = { type: 'create' } | { type: 'edit'; dep: Deposit } | null

const today = new Date().toISOString().slice(0, 10)

function yearlyIncome(dep: Deposit) {
  return dep.amount * dep.interest_rate / 100
}

function daysLeft(endDate: string) {
  return Math.ceil((new Date(endDate).getTime() - Date.now()) / 86400000)
}

export default function InvestmentsPage() {
  const [deps, setDeps] = useState<Deposit[]>([])
  const [modal, setModal] = useState<Modal>(null)
  const [form, setForm] = useState({ bank_name: '', amount: '', interest_rate: '', start_date: today, end_date: '', note: '' })

  function load() {
    api.get<Deposit[]>('/investments/deposits').then((r) => setDeps(r.data))
  }
  useEffect(() => { load() }, [])

  function openCreate() {
    setForm({ bank_name: '', amount: '', interest_rate: '', start_date: today, end_date: '', note: '' })
    setModal({ type: 'create' })
  }

  function openEdit(dep: Deposit) {
    setForm({
      bank_name: dep.bank_name,
      amount: String(dep.amount),
      interest_rate: String(dep.interest_rate),
      start_date: dep.start_date.slice(0, 10),
      end_date: dep.end_date.slice(0, 10),
      note: dep.note,
    })
    setModal({ type: 'edit', dep })
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const payload = {
      bank_name: form.bank_name,
      amount: parseFloat(form.amount),
      interest_rate: parseFloat(form.interest_rate),
      start_date: form.start_date,
      end_date: form.end_date,
      note: form.note,
    }
    if (modal?.type === 'create') {
      await api.post('/investments/deposits', payload)
    } else if (modal?.type === 'edit') {
      await api.patch(`/investments/deposits/${modal.dep.id}`, payload)
    }
    setModal(null)
    load()
  }

  async function del(id: string) {
    if (!confirm('Удалить депозит?')) return
    await api.delete(`/investments/deposits/${id}`)
    load()
  }

  const fmt = (n: number) => n.toLocaleString('ru-RU', { maximumFractionDigits: 0 }) + ' ₽'
  const totalAmount = deps.reduce((s, d) => s + d.amount, 0)
  const totalYearly = deps.reduce((s, d) => s + yearlyIncome(d), 0)

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-800">Инвестиции — Депозиты</h1>
          <p className="text-sm text-gray-400 mt-0.5">Банковские вклады и их доходность</p>
        </div>
        <button onClick={openCreate} className="bg-indigo-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-indigo-700">
          + Добавить депозит
        </button>
      </div>

      {/* Summary cards */}
      {deps.length > 0 && (
        <div className="grid grid-cols-2 sm:grid-cols-3 gap-4">
          <div className="bg-white rounded-2xl p-4 shadow-sm border border-gray-100">
            <p className="text-xs text-gray-500 mb-1">Всего в депозитах</p>
            <p className="text-lg font-bold text-indigo-600">{fmt(totalAmount)}</p>
          </div>
          <div className="bg-white rounded-2xl p-4 shadow-sm border border-gray-100">
            <p className="text-xs text-gray-500 mb-1">Доход за год</p>
            <p className="text-lg font-bold text-green-600">≈ {fmt(totalYearly)}</p>
          </div>
          <div className="bg-white rounded-2xl p-4 shadow-sm border border-gray-100">
            <p className="text-xs text-gray-500 mb-1">Активных вкладов</p>
            <p className="text-lg font-bold text-gray-700">{deps.length}</p>
          </div>
        </div>
      )}

      {/* List */}
      <div className="space-y-3">
        {deps.map((d) => {
          const dl = daysLeft(d.end_date)
          const yi = yearlyIncome(d)
          return (
            <div key={d.id} className="bg-white rounded-2xl p-5 shadow-sm border border-gray-100">
              <div className="flex justify-between items-start">
                <div>
                  <p className="font-semibold text-gray-800">{d.bank_name}</p>
                  <p className="text-xs text-gray-400 mt-0.5">
                    {d.start_date.slice(0, 10)} → {d.end_date.slice(0, 10)}
                    <span className={`ml-2 ${dl < 30 ? 'text-orange-500' : dl < 0 ? 'text-red-500' : 'text-gray-400'}`}>
                      {dl < 0 ? '(истёк)' : `${dl} дн. осталось`}
                    </span>
                  </p>
                  {d.note && <p className="text-xs text-gray-400 mt-0.5">{d.note}</p>}
                </div>
                <div className="text-right">
                  <p className="text-lg font-bold text-indigo-600">{fmt(d.amount)}</p>
                  <p className="text-xs text-gray-500">{d.interest_rate}% годовых</p>
                  <p className="text-xs text-green-600">≈ {fmt(yi)} / год</p>
                </div>
              </div>
              <div className="flex gap-2 mt-3">
                <button onClick={() => openEdit(d)} className="text-xs bg-gray-50 text-gray-600 px-3 py-1.5 rounded-lg hover:bg-gray-100">Редактировать</button>
                <button onClick={() => del(d.id)} className="text-xs text-gray-300 hover:text-red-400 ml-auto px-2">Удалить</button>
              </div>
            </div>
          )
        })}
        {deps.length === 0 && (
          <div className="bg-white rounded-2xl p-10 shadow-sm border border-gray-100 text-center">
            <p className="text-3xl mb-3">🏦</p>
            <p className="font-semibold text-gray-700 mb-1">Депозитов пока нет</p>
            <p className="text-sm text-gray-400">Добавьте банковский вклад, чтобы отслеживать доходность</p>
          </div>
        )}
      </div>

      {/* Modal */}
      {modal && (
        <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-40" onClick={() => setModal(null)}>
          <form onSubmit={handleSubmit} className="bg-white rounded-2xl p-6 w-full max-w-md shadow-xl space-y-3" onClick={(e) => e.stopPropagation()}>
            <h2 className="font-semibold text-gray-800">{modal.type === 'create' ? 'Новый депозит' : 'Редактировать депозит'}</h2>
            <input type="text" placeholder="Банк" value={form.bank_name} onChange={(e) => setForm({ ...form, bank_name: e.target.value })} className="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm" required />
            <div className="grid grid-cols-2 gap-2">
              <input type="number" placeholder="Сумма, ₽" value={form.amount} onChange={(e) => setForm({ ...form, amount: e.target.value })} className="border border-gray-200 rounded-lg px-3 py-2 text-sm" required min="1" />
              <input type="number" placeholder="Ставка, % год." value={form.interest_rate} onChange={(e) => setForm({ ...form, interest_rate: e.target.value })} className="border border-gray-200 rounded-lg px-3 py-2 text-sm" required min="0" step="0.1" />
            </div>
            <div className="grid grid-cols-2 gap-2">
              <div><label className="text-xs text-gray-500">Дата открытия</label>
                <input type="date" value={form.start_date} onChange={(e) => setForm({ ...form, start_date: e.target.value })} className="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm" required /></div>
              <div><label className="text-xs text-gray-500">Дата закрытия</label>
                <input type="date" value={form.end_date} onChange={(e) => setForm({ ...form, end_date: e.target.value })} className="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm" required /></div>
            </div>
            <input type="text" placeholder="Заметка (необязательно)" value={form.note} onChange={(e) => setForm({ ...form, note: e.target.value })} className="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm" />
            <div className="flex gap-2 pt-1">
              <button type="submit" className="flex-1 bg-indigo-600 text-white py-2 rounded-lg text-sm font-medium">{modal.type === 'create' ? 'Добавить' : 'Сохранить'}</button>
              <button type="button" onClick={() => setModal(null)} className="flex-1 bg-gray-100 text-gray-600 py-2 rounded-lg text-sm">Отмена</button>
            </div>
          </form>
        </div>
      )}
    </div>
  )
}
