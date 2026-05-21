import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import AchievementsPage from './pages/AchievementsPage'
import DashboardPage from './pages/DashboardPage'
import GoalsPage from './pages/GoalsPage'
import ImportPage from './pages/ImportPage'
import LoginPage from './pages/LoginPage'
import TransactionsPage from './pages/TransactionsPage'
import Layout from './components/Layout'

function RequireAuth({ children }: { children: React.ReactNode }) {
  return localStorage.getItem('access_token') ? <>{children}</> : <Navigate to="/login" replace />
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route
          path="/"
          element={
            <RequireAuth>
              <Layout />
            </RequireAuth>
          }
        >
          <Route index element={<DashboardPage />} />
          <Route path="transactions" element={<TransactionsPage />} />
          <Route path="import" element={<ImportPage />} />
          <Route path="achievements" element={<AchievementsPage />} />
          <Route path="goals" element={<GoalsPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
