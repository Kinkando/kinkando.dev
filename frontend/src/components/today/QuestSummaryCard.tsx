import { useNavigate } from 'react-router-dom'
import type { QuestOverview } from '../../lib/api/types'

type Props = { overview: QuestOverview | undefined }

export default function QuestSummaryCard({ overview }: Props) {
  const navigate = useNavigate()

  if (!overview) {
    return (
      <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
        <p className="text-xs font-semibold tracking-widest text-amber-600 uppercase">
          Adventure Rank
        </p>
        <p className="mt-3 text-sm text-gray-600">Loading…</p>
      </div>
    )
  }

  const { xp, daily_done, daily_total, weekly_done, weekly_total } = overview
  const pct =
    xp.xp_for_level > 0
      ? Math.min((xp.xp_into_level / xp.xp_for_level) * 100, 100)
      : 0
  const dailyComplete = daily_total > 0 && daily_done === daily_total
  const weeklyComplete = weekly_total > 0 && weekly_done === weekly_total

  return (
    <div
      className="cursor-pointer rounded-xl border border-amber-900/30 bg-gray-900 p-4 transition-colors hover:border-amber-700/50"
      onClick={() => navigate('/quest')}
    >
      <div className="mb-3 flex items-center justify-between">
        <p className="text-xs font-semibold tracking-widest text-amber-600 uppercase">
          Adventure Rank
        </p>
        <div className="flex items-baseline gap-1.5">
          <span className="text-3xl font-black text-amber-400">{xp.level}</span>
          <span className="text-xs text-gray-600">/ ∞</span>
        </div>
      </div>

      {/* XP bar */}
      <div className="mb-1 h-2 w-full overflow-hidden rounded-full bg-gray-800">
        <div
          className="h-full rounded-full bg-gradient-to-r from-amber-500 to-yellow-400 transition-all duration-500"
          style={{ width: `${pct}%` }}
        />
      </div>
      <p className="mb-4 text-right text-xs text-amber-700/80">
        {xp.xp_into_level.toLocaleString()} / {xp.xp_for_level.toLocaleString()}{' '}
        XP
      </p>

      {/* Daily + Weekly summary */}
      <div className="grid grid-cols-2 gap-2">
        <div className="rounded-lg bg-gray-800/60 px-3 py-2 text-center">
          <p className="text-xs text-gray-500">Daily</p>
          <p
            className={`text-sm font-bold ${dailyComplete ? 'text-sky-400' : 'text-gray-200'}`}
          >
            {daily_done} / {daily_total}
            {dailyComplete && <span className="ml-1">⭐</span>}
          </p>
        </div>
        <div className="rounded-lg bg-gray-800/60 px-3 py-2 text-center">
          <p className="text-xs text-gray-500">Weekly</p>
          <p
            className={`text-sm font-bold ${weeklyComplete ? 'text-violet-400' : 'text-gray-200'}`}
          >
            {weekly_done} / {weekly_total}
            {weeklyComplete && <span className="ml-1">⭐</span>}
          </p>
        </div>
      </div>
    </div>
  )
}
