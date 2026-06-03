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

  return (
    <div className="space-y-6">
      {/* Add Quest button */}
      <div className="flex justify-end">
        <button
          onClick={() => setDialog({ mode: 'create' })}
          className="cursor-pointer rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500"
        >
          Add Quest
        </button>
      </div>

      {/* Quest list */}
      <div className="rounded-xl border border-gray-800 bg-gray-900">
        {quests.length === 0 ? (
          <p className="px-5 py-10 text-center text-sm text-gray-500">
            {cfg.emptyText}
          </p>
        ) : (
          <ul className="divide-y divide-gray-800">
            {quests.map((q) => {
              const isAutoQ = q.source_type !== 'manual'
              const route = questSourceRoute(q.source_type)
              return (
                <li
                  key={q.id}
                  className={`px-5 py-4 first:rounded-t-xl last:rounded-b-xl ${
                    q.completed
                      ? 'bg-green-900'
                      : q.is_active
                        ? 'bg-gray-900'
                        : 'bg-gray-950'
                  }`}
                >
                  <div className="flex items-center gap-3">
                    {/* Count controls — only for manual quests */}
                    {isAutoQ ? (
                      <div className="flex items-center gap-1">
                        <div className="w-7"></div>
                        <span className="min-w-[2.5rem] text-center text-sm font-semibold text-gray-100">
                          {q.current_count}/{q.target_count}
                        </span>
                        <div className="w-7"></div>
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
                        <span className="min-w-[2.5rem] text-center text-sm font-semibold text-gray-100">
                          {q.current_count}/{q.target_count}
                        </span>
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
                      <div className="flex flex-wrap items-center gap-1.5">
                        <p
                          className={`text-sm font-medium text-gray-100 ${route ? 'transition-colors hover:text-indigo-400' : ''}`}
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
                      {q.description && (
                        <p className="mt-0.5 text-xs text-gray-600">
                          {q.description}
                        </p>
                      )}
                      {/* Progress bar */}
                      <div className="mt-2 h-1.5 w-full overflow-hidden rounded-full bg-gray-800">
                        <div
                          className={`h-full rounded-full transition-all ${cfg.accentBar}`}
                          style={{
                            width: `${Math.min((q.current_count / q.target_count) * 100, 100)}%`,
                          }}
                        />
                      </div>
                    </div>

                    {q.xp_reward > 0 && (
                      <span
                        className={`shrink-0 text-xs font-semibold ${
                          q.completed ? 'text-amber-600/60' : 'text-amber-500'
                        }`}
                      >
                        +{q.xp_reward} XP
                      </span>
                    )}

                    <QuestRowMenu
                      isActive={q.is_active}
                      onEdit={() => setDialog({ mode: 'edit', quest: q })}
                      onToggleActive={() => handleToggleActive(q)}
                      onDelete={() => setDeleteConfirm(q.id)}
                    />
                  </div>
                </li>
              )
            })}
          </ul>
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
