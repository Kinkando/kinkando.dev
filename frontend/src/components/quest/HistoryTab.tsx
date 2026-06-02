import type { XPEvent } from '../../lib/api/types'

type Props = {
  events: XPEvent[]
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString(undefined, {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })
}

function weekKey(isoDate: string) {
  // Returns the Monday of the week containing this date (YYYY-MM-DD)
  const d = new Date(isoDate)
  const day = d.getDay()
  const diff = (day + 6) % 7 // days since Monday
  const mon = new Date(d)
  mon.setDate(d.getDate() - diff)
  return mon.toISOString().slice(0, 10)
}

function formatWeek(mondayISO: string) {
  const start = new Date(mondayISO)
  const end = new Date(start)
  end.setDate(start.getDate() + 6)
  const opts: Intl.DateTimeFormatOptions = { month: 'short', day: 'numeric' }
  return `Week of ${start.toLocaleDateString(undefined, opts)} – ${end.toLocaleDateString(undefined, opts)}`
}

export default function HistoryTab({ events }: Props) {
  if (events.length === 0) {
    return (
      <div className="rounded-xl border border-gray-800 bg-gray-900 px-5 py-12 text-center">
        <p className="text-sm text-gray-500">
          No XP events yet. Complete quests to start earning XP.
        </p>
      </div>
    )
  }

  // Group events by the Monday of the week they belong to
  const groups: { weekKey: string; events: XPEvent[] }[] = []
  const seenWeeks = new Map<string, XPEvent[]>()

  for (const event of events) {
    const wk = weekKey(event.created_at)
    if (!seenWeeks.has(wk)) {
      const arr: XPEvent[] = []
      seenWeeks.set(wk, arr)
      groups.push({ weekKey: wk, events: arr })
    }
    seenWeeks.get(wk)!.push(event)
  }

  return (
    <div className="space-y-6">
      {groups.map(({ weekKey: wk, events: weekEvents }) => {
        const weekTotal = weekEvents.reduce((sum, e) => sum + e.xp, 0)
        return (
          <div
            key={wk}
            className="rounded-xl border border-gray-800 bg-gray-900"
          >
            {/* Week header */}
            <div className="flex items-center justify-between border-b border-gray-800 px-5 py-3">
              <span className="text-xs font-medium text-gray-500">
                {formatWeek(wk)}
              </span>
              <span className="text-xs font-semibold text-indigo-400">
                +{weekTotal} XP
              </span>
            </div>
            {/* Events */}
            <ul className="divide-y divide-gray-800">
              {weekEvents.map((event) => (
                <li
                  key={event.id}
                  className="flex items-center gap-3 px-5 py-3.5"
                >
                  {/* Source badge */}
                  <span
                    className={`shrink-0 rounded px-1.5 py-0.5 text-xs font-medium ${
                      event.source === 'daily'
                        ? 'bg-sky-900/50 text-sky-400'
                        : 'bg-purple-900/50 text-purple-400'
                    }`}
                  >
                    {event.source === 'daily' ? 'Daily' : 'Weekly'}
                  </span>

                  {/* Quest title */}
                  <span className="min-w-0 flex-1 truncate text-sm text-gray-300">
                    {event.quest_title}
                  </span>

                  {/* Date */}
                  <span className="shrink-0 text-xs text-gray-600">
                    {formatDate(event.created_at)}
                  </span>

                  {/* XP */}
                  <span className="shrink-0 text-sm font-semibold text-yellow-500">
                    +{event.xp} XP
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
