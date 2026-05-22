import { Link, Outlet, useNavigate } from 'react-router-dom'
import XpBar from './XpBar'

export default function Layout() {
  const navigate = useNavigate()

  function logout() {
    localStorage.clear()
    navigate('/login')
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200 px-6 py-3 flex items-center justify-between">
        <div className="flex items-center gap-6">
          <Link to="/" className="text-xl font-bold text-indigo-600 hover:text-indigo-700">FinQuest</Link>
          <nav className="flex gap-4 text-sm">
            <Link to="/" className="text-gray-600 hover:text-indigo-600">Дашборд</Link>
            <Link to="/transactions" className="text-gray-600 hover:text-indigo-600">Транзакции</Link>
            <Link to="/import" className="text-gray-600 hover:text-indigo-600">Импорт</Link>
            <Link to="/achievements" className="text-gray-600 hover:text-indigo-600">Ачивки</Link>
            <Link to="/goals" className="text-gray-600 hover:text-indigo-600">Цели</Link>
            <Link to="/investments" className="text-gray-600 hover:text-indigo-600">Инвестиции</Link>
            <Link to="/credits" className="text-gray-600 hover:text-indigo-600">Кредиты</Link>
          </nav>
        </div>
        <div className="flex items-center gap-4">
          <XpBar />
          <button onClick={logout} className="text-sm text-gray-500 hover:text-red-500">Выйти</button>
        </div>
      </header>
      <main className="max-w-5xl mx-auto px-6 py-8">
        <Outlet />
      </main>
    </div>
  )
}
