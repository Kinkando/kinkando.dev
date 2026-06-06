import { useAchievements } from '../../queries/useQuest'
import type { Achievement } from '../../lib/api/types'

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })
}

function BadgeCard({ a, isNew }: { a: Achievement; isNew: boolean }) {
  if (a.unlocked) {
    return (
      <div
        className={`flex flex-col items-center rounded-xl border bg-gray-900 p-4 text-center transition-colors ${
          isNew
            ? 'border-indigo-500 ring-2 ring-indigo-500/60'
            : 'border-indigo-900/50'
        }`}
      >
        <span className="text-4xl">{a.icon}</span>
        <span className="mt-2 text-sm font-semibold text-gray-100">
          {a.title}
        </span>
        <span className="mt-0.5 text-xs text-gray-500">{a.description}</span>
        <span className="mt-2 text-[11px] font-medium text-indigo-400">
          {isNew
            ? '✨ Just unlocked!'
            : a.unlocked_at && `Unlocked ${formatDate(a.unlocked_at)}`}
        </span>
      </div>
    )
  }

  const pct = a.target > 0 ? Math.min((a.progress / a.target) * 100, 100) : 0
  return (
    <div className="flex flex-col items-center rounded-xl border border-gray-800 bg-gray-900 p-4 text-center">
      <span className="text-4xl opacity-40 grayscale">{a.icon}</span>
      <span className="mt-2 text-sm font-semibold text-gray-400">
        {a.title}
      </span>
      <span className="mt-0.5 text-xs text-gray-600">{a.description}</span>
      <div className="mt-3 h-1.5 w-full overflow-hidden rounded-full bg-gray-800">
        <div
          className="h-full rounded-full bg-gray-600 transition-all"
          style={{ width: `${pct}%` }}
        />
      </div>
      <span className="mt-1 text-[11px] text-gray-600">
        {a.progress} / {a.target}
      </span>
    </div>
  )
}

export default function AchievementsTab() {
  const { data, isLoading } = useAchievements()

  if (isLoading) {
    return <p className="py-12 text-center text-sm text-gray-500">Loading…</p>
  }
  if (!data || data.achievements.length === 0) {
    return (
      <p className="py-12 text-center text-sm text-gray-500">
        No achievements yet.
      </p>
    )
  }

  const newlySet = new Set(data.newly_unlocked)

  // Group by category, preserving the API's order.
  const categories: string[] = []
  const byCategory = new Map<string, Achievement[]>()
  for (const a of data.achievements) {
    if (!byCategory.has(a.category)) {
      byCategory.set(a.category, [])
      categories.push(a.category)
    }
    byCategory.get(a.category)!.push(a)
  }

  return (
    <div className="space-y-8">
      {/* Summary */}
      <div className="flex items-center justify-between rounded-xl border border-gray-800 bg-gray-900 p-5">
        <div>
          <h3 className="text-sm font-medium text-gray-300">Badges</h3>
          <p className="mt-0.5 text-xs text-gray-500">
            Milestones earned across your journey
          </p>
        </div>
        <span className="text-lg font-black text-indigo-400">
          {data.unlocked_count}
          <span className="text-sm font-medium text-gray-600">
            {' '}
            / {data.total}
          </span>
        </span>
      </div>

      {categories.map((cat) => (
        <div key={cat}>
          <h4 className="mb-3 text-xs font-semibold tracking-wider text-gray-500 uppercase">
            {cat}
          </h4>
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
            {byCategory.get(cat)!.map((a) => (
              <BadgeCard key={a.code} a={a} isNew={newlySet.has(a.code)} />
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}
