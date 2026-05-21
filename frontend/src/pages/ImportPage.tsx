import { useRef, useState } from 'react'
import api from '../api/client'

export default function ImportPage() {
  const inputRef = useRef<HTMLInputElement>(null)
  const [status, setStatus] = useState<{ imported: number; total: number } | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleUpload(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return

    setError('')
    setStatus(null)
    setLoading(true)

    const fd = new FormData()
    fd.append('file', file)

    try {
      const { data } = await api.post<{ imported: number; total: number }>('/transactions/import', fd, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      setStatus(data)
    } catch {
      setError('Ошибка при импорте. Проверьте формат файла.')
    } finally {
      setLoading(false)
      if (inputRef.current) inputRef.current.value = ''
    }
  }

  return (
    <div className="space-y-6 max-w-lg">
      <h1 className="text-2xl font-bold text-gray-800">Импорт CSV</h1>

      <div className="bg-white rounded-2xl p-6 shadow-sm border border-gray-100 space-y-4">
        <p className="text-sm text-gray-500">
          Загрузите CSV-файл с транзакциями. Ожидаемые колонки:
        </p>
        <code className="block bg-gray-50 rounded-lg px-4 py-3 text-xs text-gray-600">
          date,amount,type,note<br />
          2024-01-15,1500.00,expense,Магнит<br />
          2024-01-16,50000.00,income,Зарплата
        </code>

        <label className="block">
          <span className="sr-only">Выберите файл</span>
          <input
            ref={inputRef}
            type="file"
            accept=".csv"
            onChange={handleUpload}
            className="block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:text-sm file:font-semibold file:bg-indigo-50 file:text-indigo-600 hover:file:bg-indigo-100 cursor-pointer"
          />
        </label>

        {loading && <p className="text-sm text-indigo-500">Импортируем и категоризируем...</p>}

        {status && (
          <div className="bg-green-50 border border-green-200 rounded-lg px-4 py-3 text-sm text-green-700">
            ✓ Импортировано {status.imported} из {status.total} строк
          </div>
        )}

        {error && (
          <div className="bg-red-50 border border-red-200 rounded-lg px-4 py-3 text-sm text-red-600">
            {error}
          </div>
        )}
      </div>
    </div>
  )
}
