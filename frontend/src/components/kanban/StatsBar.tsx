import { useBoardStats } from '../../queries/useKanban'
import { PRIORITY_META } from '../../lib/kanban'
import type { Priority } from '../../lib/api/types'

type Props = {
  boardId: string
}

export default function StatsBar({ boardId }: Props) {
  const { data: stats } = useBoardStats(boardId)
  if (!stats || stats.total === 0) return null

  const priorityEntries = (
    Object.entries(stats.by_priority) as [Priority, number][]
  ).filter(([p, n]) => p !== 'none' && n > 0)

  return (
    <div className="mb-4 flex flex-wrap items-center gap-2">
      {/* Total */}
      <div className="rounded-lg border border-gray-800 bg-gray-900 px-3 py-1.5">
        <span className="text-xs text-gray-500">Total </span>
        <span className="text-sm font-semibold text-gray-200">
          {stats.total}
        </span>
      </div>

      {/* Overdue */}
      {stats.overdue > 0 && (
        <div className="rounded-lg border border-red-900/60 bg-red-950/30 px-3 py-1.5">
          <span className="text-xs text-red-400">Overdue </span>
          <span className="text-sm font-semibold text-red-300">
            {stats.overdue}
          </span>
        </div>
      )}

      {/* By column */}
      {Object.entries(stats.by_column).map(([col, count]) => (
        <div
          key={col}
          className="rounded-lg border border-gray-800 bg-gray-900 px-3 py-1.5"
        >
          <span className="text-xs text-gray-500">{col} </span>
          <span className="text-sm font-semibold text-gray-200">{count}</span>
        </div>
      ))}

      {/* By priority */}
      {priorityEntries.map(([p, n]) => {
        const meta = PRIORITY_META[p]
        return (
          <div
            key={p}
            className="rounded-lg px-3 py-1.5"
            style={{
              backgroundColor: meta.color + '1a',
              border: `1px solid ${meta.color}40`,
            }}
          >
            <span className="text-xs" style={{ color: meta.color }}>
              {meta.label}{' '}
            </span>
            <span
              className="text-sm font-semibold"
              style={{ color: meta.color }}
            >
              {n}
            </span>
          </div>
        )
      })}
    </div>
  )
}
