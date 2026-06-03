import { useNavigate } from 'react-router-dom'
import type { QuestOverview, SourceType } from '../../lib/api/types'
import { questSourceRoute } from './questConfig'

type Props = {
  overview: QuestOverview | undefined
}

function XPBar({
  value,
  max,
  gold,
}: {
  value: number
  max: number
  gold?: boolean
}) {
  const pct = max > 0 ? Math.min((value / max) * 100, 100) : 0
  return (
    <div className="h-2.5 w-full overflow-hidden rounded-full bg-gray-800">
      <div
        className={`h-full rounded-full transition-all duration-500 ${gold ? 'bg-gradient-to-r from-amber-500 to-yellow-400' : 'bg-indigo-500'}`}
        style={{ width: `${pct}%` }}
      />
    </div>
  )
}

function SourceBadge({ source }: { source: SourceType }) {
  if (source === 'manual') return null
  const label =
    source === 'medicine'
      ? '⚙ Auto · Medicine'
      : source === 'supplement'
        ? '⚙ Auto · Supplement'
        : source === 'weight'
          ? '⚙ Auto · Weight'
          : '⚙ Auto · Workout'
  return (
    <span className="rounded bg-gray-800 px-1.5 py-0.5 text-xs text-gray-500">
      {label}
    </span>
  )
}

export default function DashboardTab({ overview }: Props) {
  const navigate = useNavigate()

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
      {/* Adventure Rank / Level card */}
      <div className="relative overflow-hidden rounded-xl border border-amber-900/40 bg-gradient-to-br from-gray-900 via-gray-900 to-amber-950/30 p-6">
        {/* decorative glow */}
        <div className="pointer-events-none absolute -top-8 -right-8 h-40 w-40 rounded-full bg-amber-500/5 blur-2xl" />
        <div className="mb-4 flex items-start justify-between">
          <div>
            <p className="mb-1 text-xs font-semibold tracking-[0.2em] text-amber-600 uppercase">
              Adventure Rank
            </p>
            <div className="flex items-baseline gap-2">
              <span className="text-5xl font-black text-amber-400">
                {xp.level}
              </span>
              <span className="text-sm text-gray-500">/ ∞</span>
            </div>
          </div>
          <div className="text-right">
            <p className="text-xs text-gray-600">Total XP</p>
            <p className="text-xl font-bold text-amber-300">
              {xp.total_xp.toLocaleString()}
            </p>
            <p className="mt-0.5 text-xs text-gray-600">
              {xp.xp_to_next} to next rank
            </p>
          </div>
        </div>
        <XPBar value={xp.xp_into_level} max={xp.xp_for_level} gold />
        <p className="mt-2 text-right text-xs text-amber-700/80">
          {xp.xp_into_level} / {xp.xp_for_level} XP
        </p>
      </div>

      {/* Daily + Weekly summary grid */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        {/* Daily */}
        <div className="rounded-xl border border-sky-900/40 bg-gray-900 p-5">
          <div className="mb-3 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <span className="rounded-full bg-sky-900/60 px-2 py-0.5 text-xs font-semibold text-sky-400">
                Daily
              </span>
              <h3 className="text-sm font-medium text-gray-300">Quests</h3>
            </div>
            <span className="text-sm font-bold text-gray-100">
              {daily_done} / {daily_total}
            </span>
          </div>
          <XPBar value={daily_done} max={daily_total || 1} />
          <ul className="mt-4 space-y-2.5">
            {daily.length === 0 && (
              <li className="text-xs text-gray-600">No daily quests yet.</li>
            )}
            {daily.map((q) => {
              const route = questSourceRoute(q.source_type)
              return (
                <li key={q.id} className="space-y-1">
                  <div
                    className={`flex items-center gap-2 ${route ? 'cursor-pointer' : ''}`}
                    onClick={route ? () => navigate(route) : undefined}
                  >
                    <span
                      className={`min-w-0 flex-1 truncate text-sm transition-colors ${
                        q.completed
                          ? 'text-gray-500 line-through'
                          : route
                            ? 'text-gray-300 hover:text-indigo-400'
                            : 'text-gray-300'
                      }`}
                    >
                      {q.title}
                    </span>
                    <SourceBadge source={q.source_type} />
                    {q.completed && (
                      <span className="shrink-0 rounded bg-sky-900/60 px-1.5 py-0.5 text-xs font-medium text-sky-400">
                        ✓
                      </span>
                    )}
                    <span className="shrink-0 text-xs text-gray-500">
                      {q.current_count}/{q.target_count}
                    </span>
                  </div>
                  <div className="h-1.5 w-full overflow-hidden rounded-full bg-gray-800">
                    <div
                      className="h-full rounded-full bg-sky-500 transition-all"
                      style={{
                        width: `${Math.min((q.current_count / q.target_count) * 100, 100)}%`,
                      }}
                    />
                  </div>
                </li>
              )
            })}
          </ul>
        </div>

        {/* Weekly */}
        <div className="rounded-xl border border-violet-900/40 bg-gray-900 p-5">
          <div className="mb-3 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <span className="rounded-full bg-violet-900/60 px-2 py-0.5 text-xs font-semibold text-violet-400">
                Weekly
              </span>
              <h3 className="text-sm font-medium text-gray-300">Quests</h3>
            </div>
            <span className="text-sm font-bold text-gray-100">
              {weekly_done} / {weekly_total}
            </span>
          </div>
          <XPBar value={weekly_done} max={weekly_total || 1} />
          <ul className="mt-4 space-y-2.5">
            {weekly.length === 0 && (
              <li className="text-xs text-gray-600">No weekly quests yet.</li>
            )}
            {weekly.map((q) => {
              const route = questSourceRoute(q.source_type)
              return (
                <li key={q.id} className="space-y-1">
                  <div
                    className={`flex items-center gap-2 ${route ? 'cursor-pointer' : ''}`}
                    onClick={route ? () => navigate(route) : undefined}
                  >
                    <span
                      className={`min-w-0 flex-1 truncate text-sm transition-colors ${
                        q.completed
                          ? 'text-gray-500 line-through'
                          : route
                            ? 'text-gray-300 hover:text-indigo-400'
                            : 'text-gray-300'
                      }`}
                    >
                      {q.title}
                    </span>
                    <SourceBadge source={q.source_type} />
                    {q.completed && (
                      <span className="shrink-0 rounded bg-violet-900/60 px-1.5 py-0.5 text-xs font-medium text-violet-400">
                        ✓
                      </span>
                    )}
                    <span className="shrink-0 text-xs text-gray-500">
                      {q.current_count}/{q.target_count}
                    </span>
                  </div>
                  <div className="h-1.5 w-full overflow-hidden rounded-full bg-gray-800">
                    <div
                      className="h-full rounded-full bg-violet-500 transition-all"
                      style={{
                        width: `${Math.min((q.current_count / q.target_count) * 100, 100)}%`,
                      }}
                    />
                  </div>
                </li>
              )
            })}
          </ul>
        </div>
      </div>
    </div>
  )
}
