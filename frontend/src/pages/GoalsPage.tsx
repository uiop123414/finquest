import { useEffect, useState } from 'react'
import api from '../api/client'
import type { Goal } from '../types'

type Modal =
  | { type: 'create' }
  | { type: 'deposit'; goal: Goal }
  | { type: 'edit'; goal: Goal }
  | null

const PAGE_SIZE = 5

export default function GoalsPage() {
  const [goals, setGoals] = useState<Goal[]>([])
  const [modal, setModal] = useState<Modal>(null)
  const [activePage, setActivePage] = useState(0)
  const [donePage, setDonePage] = useState(0)

  // form states
  const [createForm, setCreateForm] = useState({ name: '', target_amount: '', current_amount: '', deadline: '' })
  const [depositAmount, setDepositAmount] = useState('')
  const [editForm, setEditForm] = useState({ name: '', target_amount: '', current_amount: '', deadline: '' })

  function load() {
    api.get<Goal[]>('/goals').then((r) => setGoals(r.data))
  }

  useEffect(() => { load() }, [])

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    await api.post('/goals', {
      name: createForm.name,
      target_amount: parseFloat(createForm.target_amount),
      current_amount: parseFloat(createForm.current_amount || '0'),
      deadline: createForm.deadline,
    })
    setModal(null)
    setCreateForm({ name: '', target_amount: '', current_amount: '', deadline: '' })
    load()
  }

  async function handleDeposit(e: React.FormEvent) {
    e.preventDefault()
    if (modal?.type !== 'deposit') return
    const delta = parseFloat(depositAmount)
    if (!delta || delta <= 0) return
    await api.patch(`/goals/${modal.goal.id}`, {
      current_amount: modal.goal.current_amount + delta,
    })
    setModal(null)
    setDepositAmount('')
    load()
  }

  async function handleEdit(e: React.FormEvent) {
    e.preventDefault()
    if (modal?.type !== 'edit') return
    await api.patch(`/goals/${modal.goal.id}`, {
      name: editForm.name,
      target_amount: parseFloat(editForm.target_amount),
      current_amount: parseFloat(editForm.current_amount),
      deadline: editForm.deadline,
    })
    setModal(null)
    load()
  }

  async function toggleComplete(goal: Goal) {
    await api.patch(`/goals/${goal.id}`, { completed: !goal.completed_at })
    load()
  }

  async function deleteGoal(id: string) {
    if (!confirm('Удалить цель?')) return
    await api.delete(`/goals/${id}`)
    load()
  }

  function openEdit(goal: Goal) {
    setEditForm({
      name: goal.name,
      target_amount: String(goal.target_amount),
      current_amount: String(goal.current_amount),
      deadline: goal.deadline.slice(0, 10),
    })
    setModal({ type: 'edit', goal })
  }

  const fmt = (n: number) => n.toLocaleString('ru-RU', { maximumFractionDigits: 0 }) + ' ₽'

  const active = goals.filter((g) => !g.completed_at)
  const done = goals.filter((g) => g.completed_at)
  const activeSlice = active.slice(activePage * PAGE_SIZE, (activePage + 1) * PAGE_SIZE)
  const doneSlice = done.slice(donePage * PAGE_SIZE, (donePage + 1) * PAGE_SIZE)

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-800">Цели</h1>
        <button
          onClick={() => setModal({ type: 'create' })}
          className="bg-indigo-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-indigo-700"
        >
          + Добавить цель
        </button>
      </div>

      {/* Active goals */}
      <div className="space-y-3">
        {activeSlice.map((g) => <GoalCard key={g.id} goal={g} fmt={fmt} onDeposit={() => { setDepositAmount(''); setModal({ type: 'deposit', goal: g }) }} onEdit={() => openEdit(g)} onToggle={() => toggleComplete(g)} onDelete={() => deleteGoal(g.id)} />)}
        {active.length === 0 && (
          <div className="bg-white rounded-2xl p-8 shadow-sm border border-gray-100 text-center text-gray-400 text-sm">
            Активных целей нет. Создайте первую!
          </div>
        )}
      </div>
      {active.length > PAGE_SIZE && (
        <Pager page={activePage} total={active.length} pageSize={PAGE_SIZE} onChange={setActivePage} />
      )}

      {/* Completed goals */}
      {done.length > 0 && (
        <div>
          <h2 className="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-3">Выполнено</h2>
          <div className="space-y-3">
            {doneSlice.map((g) => <GoalCard key={g.id} goal={g} fmt={fmt} onDeposit={() => { setDepositAmount(''); setModal({ type: 'deposit', goal: g }) }} onEdit={() => openEdit(g)} onToggle={() => toggleComplete(g)} onDelete={() => deleteGoal(g.id)} />)}
          </div>
          {done.length > PAGE_SIZE && (
            <div className="mt-3">
              <Pager page={donePage} total={done.length} pageSize={PAGE_SIZE} onChange={setDonePage} />
            </div>
          )}
        </div>
      )}

      {/* ── Modals ────────────────────────────────────────────────────── */}
      {modal && (
        <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-40 px-4" onClick={() => setModal(null)}>
          <div className="w-full flex justify-center" onClick={(e) => e.stopPropagation()}>

            {/* Create */}
            {modal.type === 'create' && (
              <form onSubmit={handleCreate} className="bg-white rounded-2xl p-6 w-full max-w-md shadow-xl space-y-3">
                <h2 className="font-semibold text-gray-800">Новая цель</h2>
                <input type="text" placeholder="Название" value={createForm.name} onChange={(e) => setCreateForm({ ...createForm, name: e.target.value })} className="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm" required />
                <input type="number" placeholder="Целевая сумма" value={createForm.target_amount} onChange={(e) => setCreateForm({ ...createForm, target_amount: e.target.value })} className="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm" required min="1" />
                <input type="number" placeholder="Уже накоплено (необязательно)" value={createForm.current_amount} onChange={(e) => setCreateForm({ ...createForm, current_amount: e.target.value })} className="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm" min="0" />
                <input type="date" value={createForm.deadline} onChange={(e) => setCreateForm({ ...createForm, deadline: e.target.value })} className="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm" required />
                <div className="flex gap-2 pt-1">
                  <button type="submit" className="flex-1 bg-indigo-600 text-white py-2 rounded-lg text-sm font-medium">Создать</button>
                  <button type="button" onClick={() => setModal(null)} className="flex-1 bg-gray-100 text-gray-600 py-2 rounded-lg text-sm">Отмена</button>
                </div>
              </form>
            )}

            {/* Deposit */}
            {modal.type === 'deposit' && (
              <form onSubmit={handleDeposit} className="bg-white rounded-2xl p-6 w-full max-w-sm mx-4 shadow-xl space-y-3">
                <h2 className="font-semibold text-gray-800">Пополнить «{modal.goal.name}»</h2>
                <p className="text-xs text-gray-400">Текущий прогресс: {fmt(modal.goal.current_amount)} / {fmt(modal.goal.target_amount)}</p>
                <input
                  type="number"
                  placeholder="Сумма пополнения"
                  value={depositAmount}
                  onChange={(e) => setDepositAmount(e.target.value)}
                  className="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm"
                  required min="1" autoFocus
                />
                <div className="flex gap-2 pt-1">
                  <button type="submit" className="flex-1 bg-green-600 text-white py-2 rounded-lg text-sm font-medium">Пополнить</button>
                  <button type="button" onClick={() => setModal(null)} className="flex-1 bg-gray-100 text-gray-600 py-2 rounded-lg text-sm">Отмена</button>
                </div>
              </form>
            )}

            {/* Edit */}
            {modal.type === 'edit' && (
              <form onSubmit={handleEdit} className="bg-white rounded-2xl p-6 w-full max-w-md shadow-xl space-y-3">
                <h2 className="font-semibold text-gray-800">Редактировать цель</h2>
                <input type="text" placeholder="Название" value={editForm.name} onChange={(e) => setEditForm({ ...editForm, name: e.target.value })} className="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm" required />
                <input type="number" placeholder="Целевая сумма" value={editForm.target_amount} onChange={(e) => setEditForm({ ...editForm, target_amount: e.target.value })} className="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm" required min="1" />
                <input type="number" placeholder="Накоплено" value={editForm.current_amount} onChange={(e) => setEditForm({ ...editForm, current_amount: e.target.value })} className="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm" min="0" />
                <input type="date" value={editForm.deadline} onChange={(e) => setEditForm({ ...editForm, deadline: e.target.value })} className="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm" required />
                <div className="flex gap-2 pt-1">
                  <button type="submit" className="flex-1 bg-indigo-600 text-white py-2 rounded-lg text-sm font-medium">Сохранить</button>
                  <button type="button" onClick={() => setModal(null)} className="flex-1 bg-gray-100 text-gray-600 py-2 rounded-lg text-sm">Отмена</button>
                </div>
              </form>
            )}

          </div>
        </div>
      )}
    </div>
  )
}

// ── Pager ───────────────────────────────────────────────────────────────────
function Pager({ page, total, pageSize, onChange }: { page: number; total: number; pageSize: number; onChange: (p: number) => void }) {
  const pages = Math.ceil(total / pageSize)
  return (
    <div className="flex items-center justify-between text-sm text-gray-500 mt-1">
      <button onClick={() => onChange(Math.max(0, page - 1))} disabled={page === 0}
        className="px-3 py-1.5 rounded-lg border border-gray-200 disabled:opacity-30 hover:bg-gray-50 disabled:cursor-default text-xs">
        ← Назад
      </button>
      <span className="text-xs">{page + 1} / {pages}</span>
      <button onClick={() => onChange(Math.min(pages - 1, page + 1))} disabled={page >= pages - 1}
        className="px-3 py-1.5 rounded-lg border border-gray-200 disabled:opacity-30 hover:bg-gray-50 disabled:cursor-default text-xs">
        Вперёд →
      </button>
    </div>
  )
}

// ── Goal card component ─────────────────────────────────────────────────────
interface GoalCardProps {
  goal: Goal
  fmt: (n: number) => string
  onDeposit: () => void
  onEdit: () => void
  onToggle: () => void
  onDelete: () => void
}

function GoalCard({ goal, fmt, onDeposit, onEdit, onToggle, onDelete }: GoalCardProps) {
  const pct = Math.min((goal.current_amount / goal.target_amount) * 100, 100)
  const done = !!goal.completed_at
  const daysLeft = Math.ceil((new Date(goal.deadline).getTime() - Date.now()) / 86400000)

  return (
    <div className={`bg-white rounded-2xl p-5 shadow-sm border transition-opacity ${done ? 'border-gray-100 opacity-60' : 'border-gray-100'}`}>
      <div className="flex justify-between items-start mb-3">
        <div>
          <div className="flex items-center gap-2">
            <p className="font-semibold text-gray-800">{goal.name}</p>
            {done && <span className="text-xs bg-green-100 text-green-700 px-2 py-0.5 rounded-full">Выполнено</span>}
          </div>
          <p className={`text-xs mt-0.5 ${daysLeft < 0 ? 'text-red-400' : daysLeft < 30 ? 'text-orange-400' : 'text-gray-400'}`}>
            {done
              ? `Закрыта ${new Date(goal.completed_at!).toLocaleDateString('ru-RU')}`
              : daysLeft < 0
                ? `Просрочено на ${Math.abs(daysLeft)} дн.`
                : `${daysLeft} дн. осталось · до ${new Date(goal.deadline).toLocaleDateString('ru-RU')}`
            }
          </p>
        </div>
        <div className="text-right">
          <p className="text-sm font-medium text-indigo-600">{fmt(goal.current_amount)}</p>
          <p className="text-xs text-gray-400">из {fmt(goal.target_amount)}</p>
        </div>
      </div>

      {/* Progress bar */}
      <div className="w-full h-2 bg-gray-100 rounded-full overflow-hidden mb-1">
        <div
          className={`h-full rounded-full transition-all ${done ? 'bg-green-500' : pct >= 100 ? 'bg-green-500' : 'bg-indigo-500'}`}
          style={{ width: `${pct}%` }}
        />
      </div>
      <p className="text-xs text-gray-400 text-right mb-3">{pct.toFixed(0)}%</p>

      {/* Action buttons */}
      <div className="flex gap-2 flex-wrap">
        {!done && (
          <button onClick={onDeposit} className="text-xs bg-green-50 text-green-700 px-3 py-1.5 rounded-lg hover:bg-green-100 font-medium">
            + Пополнить
          </button>
        )}
        <button onClick={onEdit} className="text-xs bg-gray-50 text-gray-600 px-3 py-1.5 rounded-lg hover:bg-gray-100">
          Редактировать
        </button>
        <button onClick={onToggle} className={`text-xs px-3 py-1.5 rounded-lg ${done ? 'bg-indigo-50 text-indigo-600 hover:bg-indigo-100' : 'bg-gray-50 text-gray-500 hover:bg-gray-100'}`}>
          {done ? 'Возобновить' : 'Закрыть цель'}
        </button>
        <button onClick={onDelete} className="text-xs text-gray-300 hover:text-red-400 ml-auto px-2 py-1.5">
          Удалить
        </button>
      </div>
    </div>
  )
}
