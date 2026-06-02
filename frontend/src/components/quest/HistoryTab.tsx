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

function formatWeek(mondayISO: string) {
  const start = new Date(mondayISO)
  const end = new Date(start)
  end.setDate(start.getDate() + 6)
  const opts: Intl.DateTimeFormatOptions = { month: 'short', day: 'numeric' }
  return `${start.toLocaleDateString(undefined, opts)} – ${end.toLocaleDateString(undefined, opts)}`
}

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

  // Group by Monday of the week (events arrive newest-first from API)
  const groups: { weekKey: string; events: XPEvent[] }[] = []
  const seenWeeks = new Map<string, XPEvent[]>()

  for (const ev of events) {
    const wk = weekKey(ev.created_at)
    if (!seenWeeks.has(wk)) {
      const arr: XPEvent[] = []
      seenWeeks.set(wk, arr)
      groups.push({ weekKey: wk, events: arr })
    }
    seenWeeks.get(wk)!.push(ev)
  }

  return (
    <div className="space-y-4">
      {groups.map(({ weekKey: wk, events: weekEvents }) => {
        const weekTotal = weekEvents.reduce((sum, e) => sum + e.xp, 0)
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

            {/* Event list */}
            <ul className="divide-y divide-gray-800/60">
              {weekEvents.map((ev) => (
                <li key={ev.id} className="flex items-center gap-3 px-5 py-3.5">
                  {/* Quest-type badge */}
                  <span
                    className={`shrink-0 rounded px-2 py-0.5 text-xs font-semibold ${
                      ev.source === 'daily'
                        ? 'bg-sky-900/50 text-sky-400'
                        : 'bg-violet-900/50 text-violet-400'
                    }`}
                  >
                    {ev.source === 'daily' ? 'Daily' : 'Weekly'}
                  </span>

                  {/* Quest title */}
                  <span className="min-w-0 flex-1 truncate text-sm text-gray-300">
                    {ev.quest_title}
                  </span>

                  {/* Date */}
                  <span className="shrink-0 text-xs text-gray-600">
                    {formatDate(ev.created_at)}
                  </span>

                  {/* XP earned */}
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
}
