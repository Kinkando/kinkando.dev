import { useNavigate } from 'react-router-dom'
import type { WorkoutScheduleEntry, WorkoutSession } from '../../lib/api/types'

type Props = {
  schedule: WorkoutScheduleEntry[] | undefined
  sessions: WorkoutSession[] | undefined
  /** JS day-of-week derived from today's Bangkok date: 0=Sun … 6=Sat */
  dayOfWeek: number
}

export default function WorkoutTodayCard({
  schedule,
  sessions,
  dayOfWeek,
}: Props) {
  const navigate = useNavigate()

  const scheduled = schedule?.find((e) => e.day_of_week === dayOfWeek)
  // Sessions are already filtered to today by the query; first finished session wins.
  const session = sessions?.find((s) => s.completed_at != null) ?? sessions?.[0]
  const isDone = session?.completed_at != null

  return (
    <div
      className="cursor-pointer rounded-xl border border-gray-800 bg-gray-900 p-4 transition-colors hover:border-gray-700"
      onClick={() => navigate('/health/workout')}
    >
      <p className="mb-3 text-xs font-semibold tracking-widest text-emerald-600 uppercase">
        Workout Today
      </p>

      {scheduled ? (
        <>
          <p className="text-sm font-semibold text-gray-200">
            {scheduled.preset_name}
          </p>
          <p className="mt-0.5 text-xs text-gray-500 capitalize">
            {scheduled.preset_type}
          </p>
          <div
            className={`mt-3 inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium ${
              isDone
                ? 'bg-emerald-900/40 text-emerald-400'
                : 'bg-gray-800 text-gray-400'
            }`}
          >
            {isDone ? '✓ Done' : '◦ Pending'}
          </div>
        </>
      ) : (
        <p className="text-sm text-gray-500">
          Rest day — no workout scheduled.
        </p>
      )}
    </div>
  )
}
