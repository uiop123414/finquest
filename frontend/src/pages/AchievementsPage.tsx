import { useEffect, useState } from 'react'
import api from '../api/client'
import type { Achievement, GamificationProfile } from '../types'

export default function AchievementsPage() {
  const [achievements, setAchievements] = useState<Achievement[]>([])

  useEffect(() => {
    api.get<GamificationProfile>('/gamification/profile').then((r) => setAchievements(r.data.achievements))
  }, [])

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-gray-800">Достижения</h1>
      <div className="grid grid-cols-2 sm:grid-cols-3 gap-4">
        {achievements.map((a) => (
          <AchievementCard key={a.id} achievement={a} />
        ))}
      </div>
    </div>
  )
}

function AchievementCard({ achievement: a }: { achievement: Achievement }) {
  const unlocked = !!a.earned_at

  return (
    <div className={`rounded-2xl p-5 border transition-all ${unlocked ? 'bg-white border-indigo-200 shadow-sm' : 'bg-gray-50 border-gray-100 opacity-60'}`}>
      <div className={`text-3xl mb-2 ${unlocked ? '' : 'grayscale'}`}>🏆</div>
      <p className={`font-semibold text-sm ${unlocked ? 'text-gray-800' : 'text-gray-400'}`}>{a.name}</p>
      <p className="text-xs text-gray-400 mt-1">{a.description}</p>
      {unlocked && (
        <p className="text-xs text-indigo-400 mt-2">{new Date(a.earned_at!).toLocaleDateString('ru-RU')}</p>
      )}
    </div>
  )
}
