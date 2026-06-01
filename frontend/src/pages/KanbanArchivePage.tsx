import { useState } from 'react'
import type { Card } from '../lib/api/types'
import { PRIORITY_META } from '../lib/kanban'
import {
  useBoards,
  useArchive,
  useUnarchiveCard,
  useDeleteCard,
} from '../queries/useKanban'

type Tab = 'completed' | 'general'

const REASON_META: Record<string, { label: string; className: string }> = {
  completed: {
    label: 'Completed',
    className: 'bg-emerald-950/50 text-emerald-400',
  },
  cancelled: {
    label: 'Cancelled',
    className: 'bg-red-950/50 text-red-400',
  },
  duplicate: {
    label: 'Duplicate',
    className: 'bg-purple-950/50 text-purple-400',
  },
  stale: {
    label: 'Stale',
    className: 'bg-gray-800 text-gray-400',
  },
}

// Group cards by month/year of the relevant date.
function groupByMonth(
  cards: Card[],
  dateKey: 'completed_at' | 'archived_at',
): Map<string, Card[]> {
  const map = new Map<string, Card[]>()
  for (const card of cards) {
    const raw = card[dateKey]
    if (!raw) continue
    const d = new Date(raw)
    const label = d.toLocaleDateString(undefined, {
      month: 'long',
      year: 'numeric',
    })
    const group = map.get(label) ?? []
    group.push(card)
    map.set(label, group)
  }
  return map
}

function formatDate(iso?: string) {
  if (!iso) return ''
  return new Date(iso).toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })
}

// Get unique [year, month] pairs from cards, sorted descending.
function getAvailableMonths(
  cards: Card[],
  dateKey: 'completed_at' | 'archived_at',
): { year: number; month: number; label: string }[] {
  const seen = new Set<string>()
  const result: { year: number; month: number; label: string }[] = []
  for (const card of cards) {
    const raw = card[dateKey]
    if (!raw) continue
    const d = new Date(raw)
    const year = d.getFullYear()
    const month = d.getMonth() + 1
    const key = `${year}-${month}`
    if (!seen.has(key)) {
      seen.add(key)
      result.push({
        year,
        month,
        label: d.toLocaleDateString(undefined, {
          month: 'long',
          year: 'numeric',
        }),
      })
    }
  }
  return result.sort((a, b) => b.year - a.year || b.month - a.month)
}

export default function KanbanArchivePage() {
  const [tab, setTab] = useState<Tab>('completed')
  const [selectedBoardId, setSelectedBoardId] = useState('')
  const [selectedYear, setSelectedYear] = useState(0)
  const [selectedMonth, setSelectedMonth] = useState(0)
  const [reasonFilter, setReasonFilter] = useState('')

  const boards = useBoards()
  const boardId = selectedBoardId || (boards.data?.[0]?.id ?? '')

  const archiveQuery = useArchive(boardId, {
    reason: tab === 'completed' ? 'completed' : 'general',
    year: selectedYear || undefined,
    month: selectedMonth || undefined,
  })

  const unarchiveCard = useUnarchiveCard(boardId)
  const deleteCard = useDeleteCard(boardId)

  const allCards = archiveQuery.data ?? []

  // Client-side reason filter (general tab only).
  const filteredCards =
    tab === 'general' && reasonFilter
      ? allCards.filter((c) => c.archive_reason === reasonFilter)
      : allCards

  const dateKey: 'completed_at' | 'archived_at' =
    tab === 'completed' ? 'completed_at' : 'archived_at'

  const grouped = groupByMonth(filteredCards, dateKey)
  const availableMonths = getAvailableMonths(allCards, dateKey)

  function resetFilters() {
    setSelectedYear(0)
    setSelectedMonth(0)
    setReasonFilter('')
  }

  function handleTabChange(newTab: Tab) {
    setTab(newTab)
    resetFilters()
  }

  if (boards.isLoading) {
    return (
      <div className="flex h-64 items-center justify-center text-gray-500">
        Loading…
      </div>
    )
  }

  return (
    <div className="mx-auto max-w-3xl px-6 py-8">
      {/* Header */}
      <div className="mb-6 flex flex-wrap items-center justify-between gap-4">
        <h1 className="text-xl font-bold text-gray-100">Archive</h1>

        {/* Board selector */}
        {(boards.data?.length ?? 0) > 1 && (
          <select
            value={boardId}
            onChange={(e) => {
              setSelectedBoardId(e.target.value)
              resetFilters()
            }}
            className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-1.5 text-sm text-gray-100 focus:border-indigo-500 focus:outline-none"
          >
            {boards.data?.map((b) => (
              <option key={b.id} value={b.id}>
                {b.name}
              </option>
            ))}
          </select>
        )}
      </div>

      {/* Tabs */}
      <div className="mb-6 flex gap-1 rounded-lg border border-gray-800 bg-gray-900 p-1">
        {(['completed', 'general'] as Tab[]).map((t) => (
          <button
            key={t}
            onClick={() => handleTabChange(t)}
            className={`flex-1 rounded-md py-1.5 text-sm font-medium transition-colors ${
              tab === t
                ? 'bg-gray-800 text-gray-100'
                : 'text-gray-500 hover:text-gray-300'
            }`}
          >
            {t === 'completed' ? '✓ Completed' : '⊙ General'}
          </button>
        ))}
      </div>

      {/* Filters */}
      <div className="mb-6 flex flex-wrap gap-3">
        {/* Month filter */}
        <select
          value={
            selectedYear && selectedMonth
              ? `${selectedYear}-${selectedMonth}`
              : ''
          }
          onChange={(e) => {
            if (!e.target.value) {
              setSelectedYear(0)
              setSelectedMonth(0)
            } else {
              const [y, m] = e.target.value.split('-').map(Number)
              setSelectedYear(y)
              setSelectedMonth(m)
            }
          }}
          className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-1.5 text-sm text-gray-300 focus:border-indigo-500 focus:outline-none"
        >
          <option value="">All months</option>
          {availableMonths.map((m) => (
            <option key={`${m.year}-${m.month}`} value={`${m.year}-${m.month}`}>
              {m.label}
            </option>
          ))}
        </select>

        {/* Reason filter (general tab only) */}
        {tab === 'general' && (
          <select
            value={reasonFilter}
            onChange={(e) => setReasonFilter(e.target.value)}
            className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-1.5 text-sm text-gray-300 focus:border-indigo-500 focus:outline-none"
          >
            <option value="">All reasons</option>
            <option value="cancelled">Cancelled</option>
            <option value="duplicate">Duplicate</option>
            <option value="stale">Stale</option>
          </select>
        )}
      </div>

      {/* Content */}
      {archiveQuery.isLoading ? (
        <div className="flex h-32 items-center justify-center text-gray-500">
          Loading…
        </div>
      ) : filteredCards.length === 0 ? (
        <div className="flex h-32 flex-col items-center justify-center gap-2 text-center">
          <p className="text-gray-500">No archived cards here.</p>
          {tab === 'completed' && (
            <p className="text-xs text-gray-600">
              Cards moved to the Done column will appear here after archiving.
            </p>
          )}
        </div>
      ) : (
        <div className="flex flex-col gap-8">
          {Array.from(grouped.entries()).map(([month, monthCards]) => (
            <section key={month}>
              <h2 className="mb-3 text-xs font-semibold tracking-widest text-gray-500 uppercase">
                {month}
              </h2>
              <div className="flex flex-col gap-2">
                {monthCards.map((card) => (
                  <ArchivedCard
                    key={card.id}
                    card={card}
                    tab={tab}
                    onRestore={() => unarchiveCard.mutate(card.id)}
                    onDelete={() => deleteCard.mutate(card.id)}
                  />
                ))}
              </div>
            </section>
          ))}
        </div>
      )}
    </div>
  )
}

// ---- Archived card row -----------------------------------------------------

function ArchivedCard({
  card,
  tab,
  onRestore,
  onDelete,
}: {
  card: Card
  tab: Tab
  onRestore: () => void
  onDelete: () => void
}) {
  const priorityMeta =
    card.priority && card.priority !== 'none'
      ? PRIORITY_META[card.priority]
      : null

  const reasonMeta = card.archive_reason
    ? (REASON_META[card.archive_reason] ?? null)
    : null

  const dateLabel =
    tab === 'completed'
      ? card.completed_at
        ? `Completed ${formatDate(card.completed_at)}`
        : ''
      : card.archived_at
        ? `Archived ${formatDate(card.archived_at)}`
        : ''

  return (
    <div className="group flex items-start gap-3 rounded-xl border border-gray-800 bg-gray-900 p-4">
      <div className="min-w-0 flex-1">
        <p className="text-sm font-medium text-gray-200">{card.title}</p>
        {card.description && (
          <p className="mt-0.5 line-clamp-2 text-xs text-gray-500">
            {card.description}
          </p>
        )}
        <div className="mt-2 flex flex-wrap items-center gap-2">
          {reasonMeta && (
            <span
              className={`rounded px-1.5 py-0.5 text-xs font-medium ${reasonMeta.className}`}
            >
              {reasonMeta.label}
            </span>
          )}
          {priorityMeta && (
            <span
              className="rounded px-1.5 py-0.5 text-xs font-medium"
              style={{
                backgroundColor: priorityMeta.color + '26',
                color: priorityMeta.color,
              }}
            >
              {priorityMeta.label}
            </span>
          )}
          {card.tags.slice(0, 3).map((tag) => (
            <span
              key={tag}
              className="rounded-full bg-gray-800 px-2 py-0.5 text-xs text-gray-400"
            >
              {tag}
            </span>
          ))}
          {dateLabel && (
            <span className="text-xs text-gray-600">{dateLabel}</span>
          )}
        </div>
      </div>

      {/* Actions */}
      <div className="flex flex-shrink-0 items-center gap-2 opacity-0 group-hover:opacity-100">
        <button
          onClick={onRestore}
          className="rounded-md border border-gray-700 px-2.5 py-1 text-xs text-gray-300 hover:border-indigo-500 hover:text-indigo-400"
        >
          Restore
        </button>
        <button
          onClick={onDelete}
          className="rounded-md border border-gray-700 px-2.5 py-1 text-xs text-gray-500 hover:border-red-700 hover:text-red-400"
        >
          Delete
        </button>
      </div>
    </div>
  )
}
