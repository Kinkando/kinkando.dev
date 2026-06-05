import type { XPEvent } from '../../lib/api/types'

type Props = {
  events: XPEvent[]
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString(undefined, {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
  })
}

function weekKey(isoDate: string) {
  const d = new Date(isoDate)
  const diff = (d.getDay() + 6) % 7
  const mon = new Date(d)
  mon.setDate(d.getDate() - diff)
  return mon.toISOString().slice(0, 10)
}

function dayKey(isoDate: string) {
  const d = new Date(isoDate)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

function formatWeek(mondayISO: string) {
  const start = new Date(mondayISO)
  const end = new Date(start)
  end.setDate(start.getDate() + 6)
  const opts: Intl.DateTimeFormatOptions = { month: 'short', day: 'numeric' }
  return `${start.toLocaleDateString(undefined, opts)} – ${end.toLocaleDateString(undefined, opts)}`
}

function sourceBadge(source: string): { label: string; classes: string } {
  switch (source) {
    case 'daily':
      return { label: 'Daily', classes: 'bg-sky-900/50 text-sky-400' }
    case 'weekly':
      return { label: 'Weekly', classes: 'bg-violet-900/50 text-violet-400' }
    case 'daily_bonus':
      return { label: 'Daily Bonus', classes: 'bg-amber-900/50 text-amber-400' }
    case 'weekly_bonus':
      return {
        label: 'Weekly Bonus',
        classes: 'bg-amber-900/50 text-amber-400',
      }
    default:
      return { label: source, classes: 'bg-gray-800 text-gray-400' }
  }
}

type DayGroup = { dayKey: string; events: XPEvent[] }
type WeekGroup = { weekKey: string; dayGroups: DayGroup[] }

export default function HistoryTab({ events }: Props) {
  if (events.length === 0) {
    return (
      <div className="rounded-xl border border-gray-800 bg-gray-900 px-5 py-16 text-center">
        <p className="text-2xl">📜</p>
        <p className="mt-2 text-sm text-gray-500">
          No XP events yet. Complete quests to start earning XP.
        </p>
      </div>
    )
  }

  const weekGroups: WeekGroup[] = []
  const weekEntryMap = new Map<
    string,
    { entry: WeekGroup; dayMap: Map<string, XPEvent[]> }
  >()

  for (const ev of events) {
    const wk = weekKey(ev.created_at)
    const dk = dayKey(ev.created_at)

    if (!weekEntryMap.has(wk)) {
      const entry: WeekGroup = { weekKey: wk, dayGroups: [] }
      weekEntryMap.set(wk, { entry, dayMap: new Map() })
      weekGroups.push(entry)
    }

    const { entry: weekEntry, dayMap } = weekEntryMap.get(wk)!

    if (!dayMap.has(dk)) {
      const dayEvents: XPEvent[] = []
      dayMap.set(dk, dayEvents)
      weekEntry.dayGroups.push({ dayKey: dk, events: dayEvents })
    }

    dayMap.get(dk)!.push(ev)
  }

  return (
    <div className="space-y-4">
      {weekGroups.map(({ weekKey: wk, dayGroups }) => {
        const weekTotal = dayGroups.reduce(
          (sum, dg) => sum + dg.events.reduce((s, e) => s + e.xp, 0),
          0,
        )
        return (
          <div
            key={wk}
            className="rounded-xl border border-gray-800 bg-gray-900"
          >
            {/* Week header */}
            <div className="flex items-center justify-between border-b border-gray-800 px-5 py-3">
              <div className="flex items-center gap-2">
                <span className="text-xs font-semibold tracking-wider text-gray-500 uppercase">
                  Week of
                </span>
                <span className="text-xs font-medium text-gray-400">
                  {formatWeek(wk)}
                </span>
              </div>
              <span className="rounded-full bg-amber-900/30 px-2.5 py-0.5 text-xs font-bold text-amber-400">
                +{weekTotal} XP
              </span>
            </div>

            {/* Day sub-groups */}
            {dayGroups.map(({ dayKey: dk, events: dayEvents }, di) => {
              const dayTotal = dayEvents.reduce((sum, e) => sum + e.xp, 0)
              return (
                <div
                  key={dk}
                  className={di > 0 ? 'border-t border-gray-800' : ''}
                >
                  {/* Day sub-header */}
                  <div className="flex items-center justify-between bg-gray-800/30 px-5 py-2">
                    <span className="text-xs font-semibold text-gray-400">
                      {formatDate(dayEvents[0].created_at)}
                    </span>
                    <span className="text-xs font-semibold text-amber-500/80">
                      +{dayTotal} XP
                    </span>
                  </div>

                  {/* Event list */}
                  <ul className="divide-y divide-gray-800/60">
                    {dayEvents.map((ev) => (
                      <li
                        key={ev.id}
                        className="flex items-center gap-3 px-5 py-3"
                      >
                        <span
                          className={`shrink-0 rounded px-2 py-0.5 text-xs font-semibold ${sourceBadge(ev.source).classes}`}
                        >
                          {sourceBadge(ev.source).label}
                        </span>
                        <span className="min-w-0 flex-1 truncate text-sm text-gray-300">
                          {ev.quest_title}
                        </span>
                        <span className="shrink-0 text-sm font-bold text-amber-400">
                          +{ev.xp} XP
                        </span>
                      </li>
                    ))}
                  </ul>
                </div>
              )
            })}
          </div>
        )
      })}
    </div>
  )
}
