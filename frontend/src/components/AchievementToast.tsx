import { useEffect, useState } from 'react'

interface Props {
  message: string
  onClose: () => void
}

export default function AchievementToast({ message, onClose }: Props) {
  const [visible, setVisible] = useState(true)

  useEffect(() => {
    const t = setTimeout(() => { setVisible(false); onClose() }, 3000)
    return () => clearTimeout(t)
  }, [onClose])

  if (!visible) return null

  return (
    <div className="fixed bottom-6 right-6 bg-indigo-600 text-white px-4 py-3 rounded-xl shadow-lg flex items-center gap-2 z-50">
      <span>🏆</span>
      <span className="text-sm font-medium">{message}</span>
    </div>
  )
}
