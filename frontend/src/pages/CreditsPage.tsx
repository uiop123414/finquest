import { useEffect, useState } from 'react'
import api from '../api/client'
import type { Credit } from '../types'

type Modal = { type: 'create' } | { type: 'edit'; cr: Credit } | null

const TYPE_LABELS: Record<string, string> = { consumer: 'Потребительский кредит', card: 'Кредитная карта' }

export default function CreditsPage() {
  const [credits, setCredits] = useState<Credit[]>([])
  const [modal, setModal] = useState<Modal>(null)
  const [form, setForm] = useState({
    type: 'consumer', bank_name: '', total_amount: '', remaining_balance: '',
    interest_rate: '', monthly_payment: '', note: '',
  })

  function load() {
    api.get<Credit[]>('/credits').then((r) => setCredits(r.data))
  }
  useEffect(() => { load() }, [])

  function openCreate() {
    setForm({ type: 'consumer', bank_name: '', total_amount: '', remaining_balance: '', interest_rate: '', monthly_payment: '', note: '' })
    setModal({ type: 'create' })
  }

  function openEdit(cr: Credit) {
    setForm({
      type: cr.type,
      bank_name: cr.bank_name,
      total_amount: String(cr.total_amount),
      remaining_balance: String(cr.remaining_balance),
      interest_rate: String(cr.interest_rate),
      monthly_payment: String(cr.monthly_payment),
      note: cr.note,
    })
    setModal({ type: 'edit', cr })
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const payload = {
      type: form.type,
      bank_name: form.bank_name,
      total_amount: parseFloat(form.total_amount),
      remaining_balance: parseFloat(form.remaining_balance || '0'),
      interest_rate: parseFloat(form.interest_rate),
      monthly_payment: parseFloat(form.monthly_payment || '0'),
      note: form.note,
    }
    if (modal?.type === 'create') {
      await api.post('/credits', payload)
    } else if (modal?.type === 'edit') {
      await api.patch(`/credits/${modal.cr.id}`, payload)
    }
    setModal(null)
    load()
  }

  async function del(id: string) {
    if (!confirm('Удалить запись?')) return
    await api.delete(`/credits/${id}`)
    load()
  }

  const fmt = (n: number) => n.toLocaleString('ru-RU', { maximumFractionDigits: 0 }) + ' ₽'
  const totalDebt = credits.reduce((s, c) => s + c.remaining_balance, 0)
  const totalPayment = credits.reduce((s, c) => s + c.monthly_payment, 0)
  const avgRate = credits.length
    ? credits.reduce((s, c) => s + c.interest_rate, 0) / credits.length
    : 0

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-800">Кредиты</h1>
          <p className="text-sm text-gray-400 mt-0.5">Потребительские кредиты и кредитные карты</p>
        </div>
        <button onClick={openCreate} className="bg-red-500 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-red-600">
          + Добавить
        </button>
      </div>

      {/* Summary */}
      {credits.length > 0 && (
        <div className="grid grid-cols-2 sm:grid-cols-3 gap-4">
          <div className="bg-white rounded-2xl p-4 shadow-sm border border-gray-100">
            <p className="text-xs text-gray-500 mb-1">Общий долг</p>
            <p className="text-lg font-bold text-red-500">{fmt(totalDebt)}</p>
          </div>
          <div className="bg-white rounded-2xl p-4 shadow-sm border border-gray-100">
            <p className="text-xs text-gray-500 mb-1">Ежемес. платёж</p>
            <p className="text-lg font-bold text-orange-500">{fmt(totalPayment)}</p>
          </div>
          <div className="bg-white rounded-2xl p-4 shadow-sm border border-gray-100">
            <p className="text-xs text-gray-500 mb-1">Средняя ставка</p>
            <p className="text-lg font-bold text-gray-700">{avgRate.toFixed(1)}%</p>
          </div>
        </div>
      )}

      {/* List */}
      <div className="space-y-3">
        {credits.map((cr) => {
          const pct = cr.total_amount > 0 ? ((cr.total_amount - cr.remaining_balance) / cr.total_amount) * 100 : 0
          return (
            <div key={cr.id} className="bg-white rounded-2xl p-5 shadow-sm border border-gray-100">
              <div className="flex justify-between items-start mb-3">
                <div>
                  <div className="flex items-center gap-2">
                    <p className="font-semibold text-gray-800">{cr.bank_name}</p>
                    <span className={`text-xs px-2 py-0.5 rounded-full ${cr.type === 'card' ? 'bg-purple-100 text-purple-700' : 'bg-blue-100 text-blue-700'}`}>
                      {TYPE_LABELS[cr.type]}
                    </span>
                  </div>
                  <p className="text-xs text-gray-400 mt-0.5">{cr.interest_rate}% годовых · платёж {fmt(cr.monthly_payment)}/мес</p>
                  {cr.note && <p className="text-xs text-gray-400">{cr.note}</p>}
                </div>
                <div className="text-right">
                  <p className="text-lg font-bold text-red-500">{fmt(cr.remaining_balance)}</p>
                  <p className="text-xs text-gray-400">из {fmt(cr.total_amount)}</p>
                </div>
              </div>
              {/* Repayment progress */}
              <div className="w-full h-2 bg-gray-100 rounded-full overflow-hidden mb-1">
                <div className="h-full bg-green-400 rounded-full transition-all" style={{ width: `${pct}%` }} />
              </div>
              <p className="text-xs text-gray-400 text-right mb-3">Погашено {pct.toFixed(0)}%</p>
              <div className="flex gap-2">
                <button onClick={() => openEdit(cr)} className="text-xs bg-gray-50 text-gray-600 px-3 py-1.5 rounded-lg hover:bg-gray-100">Редактировать</button>
                <button onClick={() => del(cr.id)} className="text-xs text-gray-300 hover:text-red-400 ml-auto px-2">Удалить</button>
              </div>
            </div>
          )
        })}
        {credits.length === 0 && (
          <div className="bg-white rounded-2xl p-10 shadow-sm border border-gray-100 text-center">
            <p className="text-3xl mb-3">💳</p>
            <p className="font-semibold text-gray-700 mb-1">Кредитов нет</p>
            <p className="text-sm text-gray-400">Добавьте кредит или кредитную карту для отслеживания долга</p>
          </div>
        )}
      </div>

      {/* Modal */}
      {modal && (
        <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-40 px-4" onClick={() => setModal(null)}>
          <form onSubmit={handleSubmit} className="bg-white rounded-2xl p-6 w-full max-w-md shadow-xl space-y-3" onClick={(e) => e.stopPropagation()}>
            <h2 className="font-semibold text-gray-800">{modal.type === 'create' ? 'Новый кредит' : 'Редактировать кредит'}</h2>
            <select value={form.type} onChange={(e) => setForm({ ...form, type: e.target.value })} className="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm">
              <option value="consumer">Потребительский кредит</option>
              <option value="card">Кредитная карта</option>
            </select>
            <input type="text" placeholder="Банк" value={form.bank_name} onChange={(e) => setForm({ ...form, bank_name: e.target.value })} className="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm" required />
            <div className="grid grid-cols-2 gap-2">
              <input type="number" placeholder={form.type === 'card' ? 'Лимит, ₽' : 'Сумма кредита, ₽'} value={form.total_amount} onChange={(e) => setForm({ ...form, total_amount: e.target.value })} className="border border-gray-200 rounded-lg px-3 py-2 text-sm" required min="1" />
              <input type="number" placeholder={form.type === 'card' ? 'Долг сейчас, ₽' : 'Остаток долга, ₽'} value={form.remaining_balance} onChange={(e) => setForm({ ...form, remaining_balance: e.target.value })} className="border border-gray-200 rounded-lg px-3 py-2 text-sm" min="0" />
            </div>
            <div className="grid grid-cols-2 gap-2">
              <input type="number" placeholder="Ставка, % год." value={form.interest_rate} onChange={(e) => setForm({ ...form, interest_rate: e.target.value })} className="border border-gray-200 rounded-lg px-3 py-2 text-sm" required min="0" step="0.1" />
              <input type="number" placeholder="Платёж, ₽/мес" value={form.monthly_payment} onChange={(e) => setForm({ ...form, monthly_payment: e.target.value })} className="border border-gray-200 rounded-lg px-3 py-2 text-sm" min="0" />
            </div>
            <input type="text" placeholder="Заметка (необязательно)" value={form.note} onChange={(e) => setForm({ ...form, note: e.target.value })} className="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm" />
            <div className="flex gap-2 pt-1">
              <button type="submit" className="flex-1 bg-red-500 text-white py-2 rounded-lg text-sm font-medium">{modal.type === 'create' ? 'Добавить' : 'Сохранить'}</button>
              <button type="button" onClick={() => setModal(null)} className="flex-1 bg-gray-100 text-gray-600 py-2 rounded-lg text-sm">Отмена</button>
            </div>
          </form>
        </div>
      )}
    </div>
  )
}
