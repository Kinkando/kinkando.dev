import { useNavigate } from 'react-router-dom'
import type { WeightLog } from '../../lib/api/types'

type Props = {
  weightLogs: WeightLog[] | undefined
  today: string // YYYY-MM-DD (Bangkok)
}

export default function WeightCard({ weightLogs, today }: Props) {
  const navigate = useNavigate()

  // Logs are returned oldest-first; latest is the last element.
  const latest = weightLogs?.[weightLogs.length - 1]
  const loggedToday = latest?.logged_at.slice(0, 10) === today

  return (
    <div
      className="cursor-pointer rounded-xl border border-gray-800 bg-gray-900 p-4 transition-colors hover:border-gray-700"
      onClick={() => navigate('/health')}
    >
      <p className="mb-3 text-xs font-semibold tracking-widest text-indigo-600 uppercase">
        Weight
      </p>

      {latest ? (
        <>
          <p className="text-2xl font-black text-gray-100">
            {latest.weight}
            <span className="ml-1 text-sm font-normal text-gray-500">kg</span>
          </p>
          {loggedToday ? (
            <p className="mt-2 text-xs text-emerald-500">✓ Logged today</p>
          ) : (
            <p className="mt-2 text-xs text-amber-500">
              Not logged today — tap to update.
            </p>
          )}
        </>
      ) : (
        <p className="text-sm text-gray-500">No weight logged yet.</p>
      )}
    </div>
  )
}
