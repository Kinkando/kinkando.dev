import { useQuestStreaks } from '../../queries/useQuest'
import type { HeatmapDay } from '../../lib/api/types'
import { todayDate, addDays, dayOfWeek } from '../../lib/date'

const MONTHS = [
  'Jan',
  'Feb',
  'Mar',
  'Apr',
  'May',
  'Jun',
  'Jul',
  'Aug',
  'Sep',
  'Oct',
  'Nov',
  'Dec',
]

// Shade buckets (dark theme + indigo accents). Ratio = completed / total.
const SHADES = [
  'bg-gray-800', // empty: no quests, or none completed
  'bg-indigo-900', // (0, 0.5)
  'bg-indigo-700', // [0.5, 0.75)
  'bg-indigo-500', // [0.75, 1)
  'bg-indigo-400', // perfect (= 1)
]

function shadeFor(cell: HeatmapDay | undefined): string {
  if (!cell || cell.total === 0) return SHADES[0]
  const r = cell.completed / cell.total
  if (r <= 0) return SHADES[0]
  if (r < 0.5) return SHADES[1]
  if (r < 0.75) return SHADES[2]
  if (r < 1) return SHADES[3]
  return SHADES[4]
}

type WeekCell = { date: string; cell: HeatmapDay | undefined }

function StatCard({
  icon,
  label,
  value,
  suffix,
}: {
  icon: string
  label: string
  value: number
  suffix?: string
}) {
  return (
    <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
      <div className="flex items-center gap-1.5 text-xs text-gray-500">
        <span className="text-sm">{icon}</span>
        {label}
      </div>
      <p className="mt-1 text-2xl font-black text-gray-100">
        {value}
        {suffix && (
          <span className="ml-1 text-sm font-medium text-gray-500">
            {suffix}
          </span>
        )}
      </p>
    </div>
  )
}

export default function StreaksSection() {
  const { data, isLoading } = useQuestStreaks()

  return (
    <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
      <h3 className="mb-4 text-sm font-semibold text-gray-300">Consistency</h3>

      {isLoading && (
        <p className="py-8 text-center text-sm text-gray-500">Loading…</p>
      )}

      {!isLoading && data && (
        <>
          {/* Streak stat cards */}
          <div className="mb-5 grid grid-cols-3 gap-3">
            <StatCard
              icon="🔥"
              label="Current"
              value={data.current_streak}
              suffix="days"
            />
            <StatCard
              icon="🏆"
              label="Longest"
              value={data.longest_streak}
              suffix="days"
            />
            <StatCard
              icon="✅"
              label="Perfect"
              value={data.perfect_days}
              suffix="days"
            />
          </div>

          {data.days.length === 0 ? (
            <p className="py-6 text-center text-sm text-gray-500">
              Complete daily quests to start your streak.
            </p>
          ) : (
            <Heatmap days={data.days} />
          )}
        </>
      )}
    </div>
  )
}

function Heatmap({ days }: { days: HeatmapDay[] }) {
  const byDate = new Map(days.map((d) => [d.date, d]))

  const today = todayDate()
  const windowStart = addDays(today, -363)
  // Back up to the Sunday that starts the first grid column.
  const gridStart = addDays(windowStart, -dayOfWeek(windowStart))

  // Flat list of cells from gridStart..today (inclusive), column-major chunking.
  const cells: WeekCell[] = []
  for (let d = gridStart; d <= today; d = addDays(d, 1)) {
    cells.push({ date: d, cell: byDate.get(d) })
  }
  const weeks: WeekCell[][] = []
  for (let i = 0; i < cells.length; i += 7) {
    weeks.push(cells.slice(i, i + 7))
  }

  // Month label for each week column — shown when the month changes.
  let prevMonth = ''
  const monthLabels = weeks.map((week) => {
    const m = week[0].date.slice(0, 7)
    if (m === prevMonth) return ''
    prevMonth = m
    return MONTHS[Number(week[0].date.slice(5, 7)) - 1]
  })

  const weekdayLabels = ['', 'Mon', '', 'Wed', '', 'Fri', '']

  return (
    <div className="overflow-x-auto">
      <div className="inline-flex flex-col gap-1">
        {/* Month labels */}
        <div className="flex gap-1 pl-8">
          {monthLabels.map((label, wi) => (
            <div key={wi} className="w-3">
              <span className="block text-[10px] whitespace-nowrap text-gray-600">
                {label}
              </span>
            </div>
          ))}
        </div>

        {/* Weekday labels + week columns */}
        <div className="flex gap-1">
          <div className="flex w-7 flex-col gap-1 pr-1 text-right text-[10px] leading-3 text-gray-600">
            {weekdayLabels.map((l, i) => (
              <div key={i} className="h-3">
                {l}
              </div>
            ))}
          </div>

          {weeks.map((week, wi) => (
            <div key={wi} className="flex flex-col gap-1">
              {week.map(({ date, cell }) => (
                <div
                  key={date}
                  title={
                    cell
                      ? `${date} · ${cell.completed}/${cell.total} quests`
                      : `${date} · no daily quests`
                  }
                  className={`h-3 w-3 rounded-sm ${shadeFor(cell)}`}
                />
              ))}
            </div>
          ))}
        </div>

        {/* Legend */}
        <div className="mt-1 flex items-center gap-1 pl-8 text-[10px] text-gray-600">
          <span>Less</span>
          {SHADES.map((s, i) => (
            <span key={i} className={`h-3 w-3 rounded-sm ${s}`} />
          ))}
          <span>More</span>
        </div>
      </div>
    </div>
  )
}
