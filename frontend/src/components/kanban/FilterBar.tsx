import type { Priority } from '../../lib/api/types'
import { PRIORITY_META } from '../../lib/kanban'

export type FilterState = {
  search: string
  priorities: Priority[]
  tags: string[]
  dueStatus: 'all' | 'overdue' | 'due-soon' | 'no-date'
}

export const DEFAULT_FILTER: FilterState = {
  search: '',
  priorities: [],
  tags: [],
  dueStatus: 'all',
}

type Props = {
  filter: FilterState
  onChange: (f: FilterState) => void
  allTags: string[]
}

const dueOptions: { value: FilterState['dueStatus']; label: string }[] = [
  { value: 'all', label: 'All' },
  { value: 'overdue', label: 'Overdue' },
  { value: 'due-soon', label: 'Due soon' },
  { value: 'no-date', label: 'No date' },
]

export default function FilterBar({ filter, onChange, allTags }: Props) {
  function togglePriority(p: Priority) {
    const next = filter.priorities.includes(p)
      ? filter.priorities.filter((x) => x !== p)
      : [...filter.priorities, p]
    onChange({ ...filter, priorities: next })
  }

  function toggleTag(tag: string) {
    const next = filter.tags.includes(tag)
      ? filter.tags.filter((t) => t !== tag)
      : [...filter.tags, tag]
    onChange({ ...filter, tags: next })
  }

  const hasFilters =
    !!filter.search ||
    filter.priorities.length > 0 ||
    filter.tags.length > 0 ||
    filter.dueStatus !== 'all'

  return (
    <div className="mb-4 flex flex-wrap items-center gap-2">
      {/* Search */}
      <input
        value={filter.search}
        onChange={(e) => onChange({ ...filter, search: e.target.value })}
        placeholder="Search cards…"
        className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-1.5 text-sm text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none"
      />

      {/* Priority filter */}
      <div className="flex gap-1">
        {(
          Object.entries(PRIORITY_META) as [
            Priority,
            { label: string; color: string },
          ][]
        )
          .filter(([p]) => p !== 'none')
          .map(([p, meta]) => (
            <button
              key={p}
              onClick={() => togglePriority(p)}
              className={`rounded-lg px-2.5 py-1 text-xs font-medium transition-all ${
                filter.priorities.includes(p)
                  ? 'ring-1'
                  : 'opacity-50 hover:opacity-80'
              }`}
              style={
                {
                  backgroundColor: meta.color + '26',
                  color: meta.color,
                  '--tw-ring-color': meta.color,
                } as React.CSSProperties
              }
            >
              {meta.label}
            </button>
          ))}
      </div>

      {/* Due status */}
      <div className="flex gap-0.5 rounded-lg border border-gray-700 bg-gray-800 p-0.5">
        {dueOptions.map((opt) => (
          <button
            key={opt.value}
            onClick={() => onChange({ ...filter, dueStatus: opt.value })}
            className={`rounded-md px-2.5 py-1 text-xs transition-colors ${
              filter.dueStatus === opt.value
                ? 'bg-gray-700 text-gray-100'
                : 'text-gray-500 hover:text-gray-300'
            }`}
          >
            {opt.label}
          </button>
        ))}
      </div>

      {/* Tags */}
      {allTags.length > 0 && (
        <div className="flex flex-wrap gap-1">
          {allTags.map((tag) => (
            <button
              key={tag}
              onClick={() => toggleTag(tag)}
              className={`rounded-full px-2.5 py-0.5 text-xs transition-colors ${
                filter.tags.includes(tag)
                  ? 'bg-indigo-600 text-white'
                  : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
              }`}
            >
              {tag}
            </button>
          ))}
        </div>
      )}

      {/* Clear */}
      {hasFilters && (
        <button
          onClick={() => onChange(DEFAULT_FILTER)}
          className="text-xs text-gray-500 hover:text-gray-300"
        >
          Clear filters
        </button>
      )}
    </div>
  )
}
