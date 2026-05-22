import { useState } from 'react'
import { Link, Outlet, useNavigate } from 'react-router-dom'
import XpBar from './XpBar'

const NAV_LINKS = [
  { to: '/', label: 'Дашборд' },
  { to: '/transactions', label: 'Транзакции' },
  { to: '/import', label: 'Импорт' },
  { to: '/achievements', label: 'Ачивки' },
  { to: '/goals', label: 'Цели' },
  { to: '/investments', label: 'Инвестиции' },
  { to: '/credits', label: 'Кредиты' },
]

export default function Layout() {
  const navigate = useNavigate()
  const [menuOpen, setMenuOpen] = useState(false)

  function logout() {
    localStorage.clear()
    navigate('/login')
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200 px-4 py-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            {/* Hamburger (mobile only) */}
            <button
              className="md:hidden p-1.5 rounded-lg text-gray-500 hover:bg-gray-100"
              onClick={() => setMenuOpen((o) => !o)}
              aria-label="Меню"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                {menuOpen
                  ? <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  : <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
                }
              </svg>
            </button>
            <Link to="/" className="text-xl font-bold text-indigo-600 hover:text-indigo-700">FinQuest</Link>
            {/* Desktop nav */}
            <nav className="hidden md:flex gap-4 text-sm ml-2">
              {NAV_LINKS.map((l) => (
                <Link key={l.to} to={l.to} className="text-gray-600 hover:text-indigo-600">{l.label}</Link>
              ))}
            </nav>
          </div>
          <div className="flex items-center gap-3">
            <div className="hidden sm:block"><XpBar /></div>
            <button onClick={logout} className="text-sm text-gray-500 hover:text-red-500">Выйти</button>
          </div>
        </div>

        {/* Mobile menu */}
        {menuOpen && (
          <nav className="md:hidden mt-3 pt-3 pb-1 border-t border-gray-100 flex flex-col gap-0.5">
            {NAV_LINKS.map((l) => (
              <Link
                key={l.to}
                to={l.to}
                className="text-sm text-gray-700 hover:text-indigo-600 hover:bg-indigo-50 px-3 py-2.5 rounded-lg"
                onClick={() => setMenuOpen(false)}
              >
                {l.label}
              </Link>
            ))}
            <div className="mt-2 pt-2 border-t border-gray-100 px-3 py-1">
              <XpBar />
            </div>
          </nav>
        )}
      </header>
      <main className="max-w-5xl mx-auto px-4 sm:px-6 py-6 sm:py-8">
        <Outlet />
      </main>
    </div>
  )
}
