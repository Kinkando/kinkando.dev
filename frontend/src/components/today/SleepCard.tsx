import { useNavigate } from 'react-router-dom'
import type { SleepLog } from '../../lib/api/types'

type Props = {
  sleepLogs: SleepLog[] | undefined
  today: string // YYYY-MM-DD (Bangkok)
}

function scoreColor(score: number): string {
  if (score >= 80) return 'text-emerald-400'
  if (score >= 60) return 'text-yellow-400'
  return 'text-red-400'
}

function formatDuration(minutes: number): string {
  const h = Math.floor(minutes / 60)
  const m = minutes % 60
  return m > 0 ? `${h}h ${m}m` : `${h}h`
}

/** Returns the YYYY-MM-DD string for yesterday relative to the given Bangkok date. */
function yesterday(today: string): string {
  const d = new Date(today + 'T12:00:00')
  d.setDate(d.getDate() - 1)
  return d.toISOString().slice(0, 10)
}

export default function SleepCard({ sleepLogs, today }: Props) {
  const navigate = useNavigate()

  // Logs returned oldest-first; latest is the last element.
  const latest = sleepLogs?.[sleepLogs.length - 1]
  const logDate = latest?.logged_at.slice(0, 10)
  const isRecent = logDate === today || logDate === yesterday(today)

  return (
    <div
      className="cursor-pointer rounded-xl border border-gray-800 bg-gray-900 p-4 transition-colors hover:border-gray-700"
      onClick={() => navigate('/health/sleep')}
    >
      <p className="mb-3 text-xs font-semibold tracking-widest text-purple-600 uppercase">
        Sleep
      </p>

      {latest ? (
        <>
          <div className="flex items-end gap-3">
            <p className="text-2xl font-black text-gray-100">
              {formatDuration(latest.duration_minutes)}
            </p>
            {latest.score != null && (
              <p
                className={`mb-0.5 text-lg font-bold ${scoreColor(latest.score)}`}
              >
                {latest.score}
              </p>
            )}
          </div>
          {isRecent ? (
            <p className="mt-2 text-xs text-gray-500">✓ Last night</p>
          ) : (
            <p className="mt-2 text-xs text-amber-500">
              No recent sleep logged — tap to update.
            </p>
          )}
        </>
      ) : (
        <p className="text-sm text-gray-500">No sleep logged yet.</p>
      )}
    </div>
  )
}
