import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import type { QuestType, DailyQuestStatus } from '../../lib/api/types'
import {
  useIncrementQuest,
  useDecrementQuest,
  useDeleteQuest,
  useActivateQuest,
  useDeactivateQuest,
} from '../../queries/useQuest'
import {
  QUEST_TYPE_CONFIG,
  SOURCE_LABELS,
  questSourceRoute,
} from './questConfig'
import QuestActionsMenu from './QuestActionsMenu'
import QuestFormDialog from './QuestFormDialog'
import QuestRowMenu from './QuestRowMenu'

type Props = {
  type: QuestType
  quests: DailyQuestStatus[]
}

type DialogState =
  | { mode: 'create' }
  | { mode: 'edit'; quest: DailyQuestStatus }
  | null

function SectionHeader({ label, count }: { label: string; count: number }) {
  return (
    <div className="border-t border-gray-800 px-3 pt-4 pb-2 text-xs font-semibold tracking-wider text-gray-500 uppercase sm:px-5">
      {label} · {count}
    </div>
  )
}

export default function QuestTab({ type, quests }: Props) {
  const cfg = QUEST_TYPE_CONFIG[type]
  const navigate = useNavigate()

  const [dialog, setDialog] = useState<DialogState>(null)
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null)

  const incrementQuest = useIncrementQuest()
  const decrementQuest = useDecrementQuest()
  const deleteQuest = useDeleteQuest()
  const activateQuest = useActivateQuest()
  const deactivateQuest = useDeactivateQuest()

  function handleToggleActive(q: DailyQuestStatus) {
    if (q.is_active) {
      deactivateQuest.mutate(q.id)
    } else {
      activateQuest.mutate(q.id)
    }
  }

  async function handleDelete(id: string) {
    try {
      await deleteQuest.mutateAsync(id)
      setDeleteConfirm(null)
    } catch {
      // ignore
    }
  }

  function renderRow(q: DailyQuestStatus, isFirst: boolean, isLast: boolean) {
    const isAutoQ = q.source_type !== 'manual'
    const route = questSourceRoute(q.source_type)

    const titleColor = !q.is_active
      ? 'text-gray-500'
      : q.completed
        ? 'text-gray-400'
        : 'text-gray-100'

    const titleBadges = (
      <div className="flex flex-wrap items-center gap-1.5">
        {q.completed && (
          <span className="shrink-0 text-sm text-gray-400">✓</span>
        )}
        <p
          className={`text-sm font-medium ${titleColor} ${route ? 'transition-colors hover:text-indigo-400' : ''}`}
        >
          {q.title}
        </p>
        {isAutoQ && (
          <span className="rounded bg-gray-800 px-1.5 py-0.5 text-xs text-gray-500">
            ⚙ {SOURCE_LABELS[q.source_type]}
          </span>
        )}
        {q.completed && (
          <span
            className={`rounded px-1.5 py-0.5 text-xs font-medium ${cfg.accentBadge}`}
          >
            Complete
          </span>
        )}
        {!q.is_active && (
          <span className="rounded bg-gray-800 px-1.5 py-0.5 text-xs text-gray-500">
            inactive
          </span>
        )}
      </div>
    )

    const progressBar = (
      <div className="h-1.5 w-full overflow-hidden rounded-full bg-gray-800">
        <div
          className={`h-full rounded-full transition-all ${cfg.accentBar} ${q.completed || !q.is_active ? 'opacity-40' : ''}`}
          style={{
            width: `${Math.min((q.current_count / q.target_count) * 100, 100)}%`,
          }}
        />
      </div>
    )

    const countDisplay = (
      <span className="min-w-[2.5rem] text-center text-sm font-semibold text-gray-100">
        {q.current_count}/{q.target_count}
      </span>
    )

    const xpBadge = q.xp_reward > 0 && (
      <span
        className={`shrink-0 text-xs font-semibold ${
          q.completed || !q.is_active ? 'text-amber-600/40' : 'text-amber-500'
        }`}
      >
        +{q.xp_reward} XP
      </span>
    )

    const rowMenu = (
      <QuestRowMenu
        isActive={q.is_active}
        onEdit={() => setDialog({ mode: 'edit', quest: q })}
        onToggleActive={() => handleToggleActive(q)}
        onDelete={() => setDeleteConfirm(q.id)}
      />
    )

    return (
      <li
        key={q.id}
        className={`px-3 py-3 sm:px-5 sm:py-4 ${isFirst ? 'rounded-t-xl' : ''} ${isLast ? 'rounded-b-xl' : ''} ${
          !q.is_active ? 'bg-gray-950 opacity-70' : 'bg-gray-900'
        }`}
      >
        {/* Mobile layout: title row + controls/bar row */}
        <div className="sm:hidden">
          <div className="flex items-start gap-2">
            <div
              className={`min-w-0 flex-1 ${route ? 'cursor-pointer' : ''}`}
              onClick={route ? () => navigate(route) : undefined}
            >
              {titleBadges}
              {q.description && (
                <p className="mt-0.5 text-xs text-gray-600">{q.description}</p>
              )}
            </div>
            {isAutoQ && countDisplay}
            {xpBadge}
            {rowMenu}
          </div>
          <div className="mt-2 flex items-center gap-3">
            {!isAutoQ && (
              <div className="flex items-center gap-1">
                <button
                  onClick={() => decrementQuest.mutate(q.id)}
                  disabled={
                    decrementQuest.isPending ||
                    q.current_count === 0 ||
                    !q.is_active
                  }
                  className={`flex h-9 w-9 ${!q.is_active || q.current_count === 0 ? 'cursor-not-allowed' : 'cursor-pointer'} items-center justify-center rounded-md bg-gray-800 text-sm font-bold text-gray-300 hover:bg-gray-700 disabled:opacity-40`}
                  aria-label="Decrement"
                >
                  −
                </button>
                {countDisplay}
                <button
                  onClick={() => incrementQuest.mutate(q.id)}
                  disabled={incrementQuest.isPending || !q.is_active}
                  className={`flex h-9 w-9 ${!q.is_active ? 'cursor-not-allowed' : 'cursor-pointer'} items-center justify-center rounded-md bg-gray-800 text-sm font-bold text-gray-300 hover:bg-gray-700 disabled:opacity-40`}
                  aria-label="Increment"
                >
                  +
                </button>
              </div>
            )}
            <div className="min-w-0 flex-1">{progressBar}</div>
          </div>
        </div>

        {/* Desktop layout: single row */}
        <div className="hidden sm:flex sm:items-center sm:gap-3">
          {isAutoQ ? (
            <div className="flex items-center gap-1">
              <div className="w-7" />
              {countDisplay}
              <div className="w-7" />
            </div>
          ) : (
            <div className="flex items-center gap-1">
              <button
                onClick={() => decrementQuest.mutate(q.id)}
                disabled={
                  decrementQuest.isPending ||
                  q.current_count === 0 ||
                  !q.is_active
                }
                className={`flex h-7 w-7 ${!q.is_active || q.current_count === 0 ? 'cursor-not-allowed' : 'cursor-pointer'} items-center justify-center rounded-md bg-gray-800 text-sm font-bold text-gray-300 hover:bg-gray-700 disabled:opacity-40`}
                aria-label="Decrement"
              >
                −
              </button>
              {countDisplay}
              <button
                onClick={() => incrementQuest.mutate(q.id)}
                disabled={incrementQuest.isPending || !q.is_active}
                className={`flex h-7 w-7 ${!q.is_active ? 'cursor-not-allowed' : 'cursor-pointer'} items-center justify-center rounded-md bg-gray-800 text-sm font-bold text-gray-300 hover:bg-gray-700 disabled:opacity-40`}
                aria-label="Increment"
              >
                +
              </button>
            </div>
          )}
          <div
            className={`min-w-0 flex-1 ${route ? 'cursor-pointer' : ''}`}
            onClick={route ? () => navigate(route) : undefined}
          >
            {titleBadges}
            {q.description && (
              <p className="mt-0.5 text-xs text-gray-600">{q.description}</p>
            )}
            <div className="mt-2">{progressBar}</div>
          </div>
          {xpBadge}
          {rowMenu}
        </div>
      </li>
    )
  }

  const activeQuests = quests.filter((q) => q.is_active && !q.completed)
  const completedQuests = quests.filter((q) => q.is_active && q.completed)
  const inactiveQuests = quests.filter((q) => !q.is_active)

  // Determine which is the very first / last visible row across all sections
  const allOrdered = [...activeQuests, ...completedQuests, ...inactiveQuests]

  return (
    <div className="space-y-6">
      {/* Quest actions */}
      <div className="flex justify-end">
        <QuestActionsMenu onNewQuest={() => setDialog({ mode: 'create' })} />
      </div>

      {/* Quest list */}
      <div className="rounded-xl border border-gray-800 bg-gray-900">
        {quests.length === 0 ? (
          <p className="px-5 py-10 text-center text-sm text-gray-500">
            {cfg.emptyText}
          </p>
        ) : (
          <>
            {/* Active / In Progress */}
            {activeQuests.length > 0 && (
              <ul className="divide-y divide-gray-800">
                {activeQuests.map((q) =>
                  renderRow(
                    q,
                    q === allOrdered[0],
                    q === allOrdered[allOrdered.length - 1],
                  ),
                )}
              </ul>
            )}

            {/* Completed */}
            {completedQuests.length > 0 && (
              <>
                <SectionHeader
                  label="Completed"
                  count={completedQuests.length}
                />
                <ul className="divide-y divide-gray-800">
                  {completedQuests.map((q) =>
                    renderRow(
                      q,
                      false,
                      q === allOrdered[allOrdered.length - 1],
                    ),
                  )}
                </ul>
              </>
            )}

            {/* Inactive */}
            {inactiveQuests.length > 0 && (
              <>
                <SectionHeader label="Inactive" count={inactiveQuests.length} />
                <ul className="divide-y divide-gray-800">
                  {inactiveQuests.map((q, i) =>
                    renderRow(q, false, i === inactiveQuests.length - 1),
                  )}
                </ul>
              </>
            )}
          </>
        )}
      </div>

      {/* Delete confirm modal */}
      {deleteConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div className="w-full max-w-sm rounded-xl border border-gray-700 bg-gray-900 p-6 shadow-2xl">
            <p className="mb-4 text-sm text-gray-300">{cfg.deleteText}</p>
            <div className="flex justify-end gap-2">
              <button
                onClick={() => setDeleteConfirm(null)}
                className="cursor-pointer rounded-lg bg-gray-800 px-4 py-2 text-sm font-medium text-gray-300 hover:bg-gray-700"
              >
                Cancel
              </button>
              <button
                onClick={() => handleDelete(deleteConfirm)}
                disabled={deleteQuest.isPending}
                className="cursor-pointer rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-500 disabled:opacity-50"
              >
                {deleteQuest.isPending ? 'Deleting…' : 'Delete'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Create / edit dialog */}
      {dialog !== null && (
        <QuestFormDialog
          type={type}
          initial={dialog.mode === 'edit' ? dialog.quest : undefined}
          onClose={() => setDialog(null)}
        />
      )}
    </div>
  )
}
