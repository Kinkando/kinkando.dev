import { useState } from 'react'
import type {
  DailyQuestStatus,
  CreateQuestInput,
  UpdateQuestInput,
  SourceType,
} from '../../lib/api/types'
import {
  useCompleteDaily,
  useUncompleteDaily,
  useCreateQuest,
  useUpdateQuest,
  useDeleteQuest,
} from '../../queries/useQuest'

type Props = {
  daily: DailyQuestStatus[]
}

const inputClass =
  'w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none'

const labelClass = 'mb-1 block text-xs font-medium text-gray-400'

type FormState = {
  title: string
  description: string
  xp_reward: string
  source_type: SourceType
}

const defaultForm: FormState = {
  title: '',
  description: '',
  xp_reward: '10',
  source_type: 'manual',
}

function questToForm(q: DailyQuestStatus): FormState {
  return {
    title: q.title,
    description: q.description,
    xp_reward: String(q.xp_reward),
    source_type: q.source_type,
  }
}

const SOURCE_LABELS: Record<SourceType, string> = {
  manual: 'Manual',
  medicine: 'Medicine (auto)',
  workout: 'Workout (auto)',
}

export default function DailyTab({ daily }: Props) {
  const [form, setForm] = useState<FormState>(defaultForm)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [formError, setFormError] = useState('')
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null)

  const completeDaily = useCompleteDaily()
  const uncompleteDaily = useUncompleteDaily()
  const createQuest = useCreateQuest()
  const updateQuest = useUpdateQuest()
  const deleteQuest = useDeleteQuest()

  const isEditing = editingId !== null
  const formLoading = createQuest.isPending || updateQuest.isPending

  function handleEdit(q: DailyQuestStatus) {
    setEditingId(q.id)
    setForm(questToForm(q))
    setFormError('')
  }

  function handleCancelEdit() {
    setEditingId(null)
    setForm(defaultForm)
    setFormError('')
  }

  function handleToggle(q: DailyQuestStatus) {
    if (q.completed_today) {
      uncompleteDaily.mutate(q.id)
    } else {
      completeDaily.mutate(q.id)
    }
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setFormError('')

    if (!form.title.trim()) {
      setFormError('Title is required.')
      return
    }
    const xp = parseInt(form.xp_reward, 10)
    if (isNaN(xp) || xp < 0) {
      setFormError('XP reward must be 0 or greater.')
      return
    }

    try {
      if (isEditing) {
        const existing = daily.find((q) => q.id === editingId)!
        const input: UpdateQuestInput = {
          source_type: form.source_type,
          title: form.title.trim(),
          description: form.description.trim(),
          xp_reward: xp,
          target_count: 1,
          is_active: existing.is_active,
        }
        await updateQuest.mutateAsync({ id: editingId!, input })
        setEditingId(null)
      } else {
        const input: CreateQuestInput = {
          type: 'daily',
          source_type: form.source_type,
          title: form.title.trim(),
          description: form.description.trim(),
          xp_reward: xp,
          target_count: 1,
        }
        await createQuest.mutateAsync(input)
      }
      setForm(defaultForm)
    } catch (err) {
      setFormError(err instanceof Error ? err.message : 'Something went wrong.')
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

  const isAuto = form.source_type !== 'manual'

  return (
    <div className="space-y-6">
      {/* Create / edit form */}
      <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
        <h3 className="mb-4 text-sm font-semibold text-gray-300">
          {isEditing ? 'Edit Daily Quest' : 'Add Daily Quest'}
        </h3>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div className="sm:col-span-2">
              <label className={labelClass}>Title</label>
              <input
                className={inputClass}
                placeholder="e.g. Take thyroid medication"
                value={form.title}
                onChange={(e) => setForm({ ...form, title: e.target.value })}
              />
            </div>
            <div className="sm:col-span-2">
              <label className={labelClass}>Description (optional)</label>
              <input
                className={inputClass}
                placeholder="Optional notes"
                value={form.description}
                onChange={(e) =>
                  setForm({ ...form, description: e.target.value })
                }
              />
            </div>
            <div>
              <label className={labelClass}>Source</label>
              <select
                className={inputClass}
                value={form.source_type}
                onChange={(e) =>
                  setForm({
                    ...form,
                    source_type: e.target.value as SourceType,
                  })
                }
              >
                <option value="manual">Manual (check off yourself)</option>
                <option value="medicine">Medicine (auto on take)</option>
                <option value="workout">Workout (auto on finish)</option>
              </select>
            </div>
            <div>
              <label className={labelClass}>XP Reward</label>
              <input
                className={inputClass}
                type="number"
                min="0"
                placeholder="10"
                value={form.xp_reward}
                onChange={(e) =>
                  setForm({ ...form, xp_reward: e.target.value })
                }
              />
            </div>
          </div>
          {isAuto && (
            <p className="rounded-lg border border-amber-900/30 bg-amber-950/20 px-3 py-2 text-xs text-amber-600">
              This quest will auto-complete when you{' '}
              {form.source_type === 'medicine'
                ? 'take a medicine'
                : 'finish a workout session'}
              . No manual checkbox will be shown.
            </p>
          )}
          {formError && <p className="text-sm text-red-400">{formError}</p>}
          <div className="flex justify-end gap-2">
            {isEditing && (
              <button
                type="button"
                onClick={handleCancelEdit}
                className="cursor-pointer rounded-lg bg-gray-800 px-4 py-2 text-sm font-medium text-gray-300 hover:bg-gray-700"
              >
                Cancel
              </button>
            )}
            <button
              type="submit"
              disabled={formLoading}
              className="cursor-pointer rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
            >
              {formLoading ? 'Saving…' : isEditing ? 'Update' : 'Add Quest'}
            </button>
          </div>
        </form>
      </div>

      {/* Quest list */}
      <div className="rounded-xl border border-gray-800 bg-gray-900">
        {daily.length === 0 ? (
          <p className="px-5 py-10 text-center text-sm text-gray-500">
            No daily quests yet. Add one above.
          </p>
        ) : (
          <ul className="divide-y divide-gray-800">
            {daily.map((q) => {
              const isAuto = q.source_type !== 'manual'
              return (
                <li
                  key={q.id}
                  className="group flex items-center gap-3 px-5 py-3.5"
                >
                  {/* Check toggle — only for manual quests */}
                  {isAuto ? (
                    <div
                      className={`flex h-6 w-6 shrink-0 items-center justify-center rounded-full border ${
                        q.completed_today
                          ? 'border-sky-500 bg-sky-500 text-white'
                          : 'border-gray-700 text-gray-700'
                      }`}
                    >
                      <svg
                        className="h-3.5 w-3.5"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                        strokeWidth={3}
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          d="M5 13l4 4L19 7"
                        />
                      </svg>
                    </div>
                  ) : (
                    <button
                      onClick={() => handleToggle(q)}
                      disabled={
                        completeDaily.isPending || uncompleteDaily.isPending
                      }
                      className={`flex h-6 w-6 shrink-0 cursor-pointer items-center justify-center rounded-full border transition-colors disabled:opacity-50 ${
                        q.completed_today
                          ? 'border-sky-500 bg-sky-500 text-white'
                          : 'border-gray-600 text-transparent hover:border-sky-400'
                      }`}
                      aria-label={
                        q.completed_today ? 'Uncheck quest' : 'Complete quest'
                      }
                    >
                      <svg
                        className="h-3.5 w-3.5"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                        strokeWidth={3}
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          d="M5 13l4 4L19 7"
                        />
                      </svg>
                    </button>
                  )}

                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-1.5">
                      <p
                        className={`text-sm font-medium ${
                          q.completed_today
                            ? 'text-gray-500 line-through'
                            : 'text-gray-100'
                        }`}
                      >
                        {q.title}
                      </p>
                      {isAuto && (
                        <span className="rounded bg-gray-800 px-1.5 py-0.5 text-xs text-gray-500">
                          ⚙ {SOURCE_LABELS[q.source_type]}
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
                  </div>

                  {q.xp_reward > 0 && (
                    <span
                      className={`shrink-0 text-xs font-semibold ${
                        q.completed_today
                          ? 'text-amber-600/60'
                          : 'text-amber-500'
                      }`}
                    >
                      +{q.xp_reward} XP
                    </span>
                  )}

                  <button
                    onClick={() => handleEdit(q)}
                    className="shrink-0 cursor-pointer text-xs text-gray-600 opacity-0 transition-opacity group-hover:opacity-100 hover:text-gray-100"
                  >
                    Edit
                  </button>
                  <button
                    onClick={() => setDeleteConfirm(q.id)}
                    className="shrink-0 cursor-pointer text-xs text-red-700 opacity-0 transition-opacity group-hover:opacity-100 hover:text-red-400"
                  >
                    Delete
                  </button>
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
            <p className="mb-4 text-sm text-gray-300">
              Delete this daily quest? All completions and XP events will be
              removed.
            </p>
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
    </div>
  )
}
