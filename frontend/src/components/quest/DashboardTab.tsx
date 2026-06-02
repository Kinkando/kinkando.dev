import type { QuestOverview } from '../../lib/api/types'

type Props = {
  overview: QuestOverview | undefined
}

function XPBar({ value, max }: { value: number; max: number }) {
  const pct = max > 0 ? Math.min((value / max) * 100, 100) : 0
  return (
    <div className="h-2 w-full overflow-hidden rounded-full bg-gray-800">
      <div
        className="h-full rounded-full bg-indigo-500 transition-all duration-300"
        style={{ width: `${pct}%` }}
      />
    </div>
  )
}

export default function DashboardTab({ overview }: Props) {
  if (!overview) {
    return <p className="py-12 text-center text-sm text-gray-500">Loading…</p>
  }

  const {
    xp,
    daily_done,
    daily_total,
    weekly_done,
    weekly_total,
    daily,
    weekly,
  } = overview

  return (
    <div className="space-y-6">
      {/* XP / Level card */}
      <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
        <div className="mb-3 flex items-center justify-between">
          <div>
            <p className="text-xs font-medium tracking-wider text-gray-500 uppercase">
              Level
            </p>
            <p className="text-3xl font-bold text-indigo-400">{xp.level}</p>
          </div>
          <div className="text-right">
            <p className="text-xs text-gray-500">Total XP</p>
            <p className="text-lg font-semibold text-gray-100">
              {xp.total_xp.toLocaleString()}
            </p>
          </div>
        </div>
        <XPBar value={xp.xp_into_level} max={xp.xp_for_level} />
        <p className="mt-1.5 text-right text-xs text-gray-500">
          {xp.xp_into_level} / {xp.xp_for_level} XP &mdash; {xp.xp_to_next} to
          next level
        </p>
      </div>

      {/* Progress summary */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        {/* Daily progress */}
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
          <div className="mb-3 flex items-center justify-between">
            <h3 className="text-sm font-medium text-gray-300">Daily Quests</h3>
            <span className="text-sm font-semibold text-gray-100">
              {daily_done} / {daily_total}
            </span>
          </div>
          <XPBar value={daily_done} max={daily_total || 1} />
          <ul className="mt-4 space-y-1.5">
            {daily.length === 0 && (
              <li className="text-xs text-gray-600">No daily quests yet.</li>
            )}
            {daily.map((q) => (
              <li key={q.id} className="flex items-center gap-2">
                <span
                  className={`flex h-4 w-4 shrink-0 items-center justify-center rounded-full border text-xs ${
                    q.completed_today
                      ? 'border-indigo-500 bg-indigo-500 text-white'
                      : 'border-gray-700 text-gray-700'
                  }`}
                >
                  {q.completed_today ? '✓' : ''}
                </span>
                <span
                  className={`text-sm ${
                    q.completed_today
                      ? 'text-gray-500 line-through'
                      : 'text-gray-300'
                  }`}
                >
                  {q.title}
                </span>
                {q.xp_reward > 0 && (
                  <span className="ml-auto shrink-0 text-xs text-yellow-600">
                    +{q.xp_reward} XP
                  </span>
                )}
              </li>
            ))}
          </ul>
        </div>

        {/* Weekly progress */}
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
          <div className="mb-3 flex items-center justify-between">
            <h3 className="text-sm font-medium text-gray-300">Weekly Quests</h3>
            <span className="text-sm font-semibold text-gray-100">
              {weekly_done} / {weekly_total}
            </span>
          </div>
          <XPBar value={weekly_done} max={weekly_total || 1} />
          <ul className="mt-4 space-y-2">
            {weekly.length === 0 && (
              <li className="text-xs text-gray-600">No weekly quests yet.</li>
            )}
            {weekly.map((q) => (
              <li key={q.id} className="space-y-1">
                <div className="flex items-center gap-2">
                  <span
                    className={`text-sm ${
                      q.completed
                        ? 'text-gray-500 line-through'
                        : 'text-gray-300'
                    }`}
                  >
                    {q.title}
                  </span>
                  {q.completed && (
                    <span className="rounded bg-indigo-900/60 px-1.5 py-0.5 text-xs font-medium text-indigo-400">
                      Done
                    </span>
                  )}
                  <span className="ml-auto shrink-0 text-xs text-gray-500">
                    {q.current_count}/{q.target_count}
                  </span>
                </div>
                <div className="h-1 w-full overflow-hidden rounded-full bg-gray-800">
                  <div
                    className="h-full rounded-full bg-indigo-500 transition-all"
                    style={{
                      width: `${Math.min((q.current_count / q.target_count) * 100, 100)}%`,
                    }}
                  />
                </div>
              </li>
            ))}
          </ul>
        </div>
      </div>
    </div>
  )
}
