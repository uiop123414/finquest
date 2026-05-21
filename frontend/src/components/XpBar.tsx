import { useEffect, useState } from 'react'
import api from '../api/client'
import type { GamificationProfile } from '../types'

export default function XpBar() {
  const [profile, setProfile] = useState<GamificationProfile | null>(null)

  useEffect(() => {
    api.get<GamificationProfile>('/gamification/profile')
      .then((r) => setProfile(r.data))
      .catch(() => null)
  }, [])

  if (!profile) return null

  return (
    <div className="flex items-center gap-2" data-testid="xp-bar">
      <span className="text-xs font-semibold text-indigo-600">Lv.{profile.level}</span>
      <div className="w-24 h-2 bg-gray-200 rounded-full overflow-hidden">
        <div
          className="h-full bg-indigo-500 rounded-full transition-all"
          style={{ width: `${profile.level_progress_pct}%` }}
        />
      </div>
      <span className="text-xs text-gray-500">{profile.xp_total} XP</span>
    </div>
  )
}
