import { useState } from 'react'
import type {
  WeeklyQuestStatus,
  CreateQuestInput,
  UpdateQuestInput,
} from '../../lib/api/types'
import {
  useIncrementWeekly,
  useDecrementWeekly,
  useCreateQuest,
  useUpdateQuest,
  useDeleteQuest,
} from '../../queries/useQuest'

type Props = {
  weekly: WeeklyQuestStatus[]
}

const inputClass =
  'w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none'

const labelClass = 'mb-1 block text-xs font-medium text-gray-400'

type FormState = {
  title: string
  description: string
  xp_reward: string
  target_count: string
}

const defaultForm: FormState = {
  title: '',
  description: '',
  xp_reward: '30',
  target_count: '3',
}

function questToForm(q: WeeklyQuestStatus): FormState {
  return {
    title: q.title,
    description: q.description,
    xp_reward: String(q.xp_reward),
    target_count: String(q.target_count),
  }
}

export default function WeeklyTab({ weekly }: Props) {
  const [form, setForm] = useState<FormState>(defaultForm)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [formError, setFormError] = useState('')
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null)

  const incrementWeekly = useIncrementWeekly()
  const decrementWeekly = useDecrementWeekly()
  const createQuest = useCreateQuest()
  const updateQuest = useUpdateQuest()
  const deleteQuest = useDeleteQuest()

  const isEditing = editingId !== null
  const formLoading = createQuest.isPending || updateQuest.isPending

  function handleEdit(q: WeeklyQuestStatus) {
    setEditingId(q.id)
    setForm(questToForm(q))
    setFormError('')
  }

  function handleCancelEdit() {
    setEditingId(null)
    setForm(defaultForm)
    setFormError('')
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
    const target = parseInt(form.target_count, 10)
    if (isNaN(target) || target < 1) {
      setFormError('Target count must be at least 1.')
      return
    }

    try {
      if (isEditing) {
        const existing = weekly.find((q) => q.id === editingId)!
        const input: UpdateQuestInput = {
          title: form.title.trim(),
          description: form.description.trim(),
          xp_reward: xp,
          target_count: target,
          is_active: existing.is_active,
        }
        await updateQuest.mutateAsync({ id: editingId!, input })
        setEditingId(null)
      } else {
        const input: CreateQuestInput = {
          type: 'weekly',
          title: form.title.trim(),
          description: form.description.trim(),
          xp_reward: xp,
          target_count: target,
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

  return (
    <div className="space-y-6">
      {/* Create / edit form */}
      <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
        <h3 className="mb-4 text-sm font-medium text-gray-300">
          {isEditing ? 'Edit Weekly Quest' : 'Add Weekly Quest'}
        </h3>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div className="sm:col-span-2">
              <label className={labelClass}>Title</label>
              <input
                className={inputClass}
                placeholder="e.g. Workout"
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
              <label className={labelClass}>Target count / week</label>
              <input
                className={inputClass}
                type="number"
                min="1"
                placeholder="3"
                value={form.target_count}
                onChange={(e) =>
                  setForm({ ...form, target_count: e.target.value })
                }
              />
            </div>
            <div>
              <label className={labelClass}>XP Reward (on completion)</label>
              <input
                className={inputClass}
                type="number"
                min="0"
                placeholder="30"
                value={form.xp_reward}
                onChange={(e) =>
                  setForm({ ...form, xp_reward: e.target.value })
                }
              />
            </div>
          </div>
          {formError && <p className="text-sm text-red-400">{formError}</p>}
          <div className="flex gap-2">
            <button
              type="submit"
              disabled={formLoading}
              className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
            >
              {formLoading ? 'Saving…' : isEditing ? 'Update' : 'Add Quest'}
            </button>
            {isEditing && (
              <button
                type="button"
                onClick={handleCancelEdit}
                className="rounded-lg bg-gray-800 px-4 py-2 text-sm font-medium text-gray-300 hover:bg-gray-700"
              >
                Cancel
              </button>
            )}
          </div>
        </form>
      </div>

      {/* Quest list */}
      <div className="rounded-xl border border-gray-800 bg-gray-900">
        {weekly.length === 0 ? (
          <p className="px-5 py-8 text-center text-sm text-gray-500">
            No weekly quests yet. Add one above.
          </p>
        ) : (
          <ul className="divide-y divide-gray-800">
            {weekly.map((q) => (
              <li key={q.id} className="px-5 py-4">
                <div className="flex items-center gap-3">
                  {/* Count controls */}
                  <div className="flex items-center gap-1">
                    <button
                      onClick={() => decrementWeekly.mutate(q.id)}
                      disabled={
                        decrementWeekly.isPending || q.current_count === 0
                      }
                      className="flex h-7 w-7 items-center justify-center rounded-md bg-gray-800 text-sm font-bold text-gray-300 hover:bg-gray-700 disabled:opacity-40"
                      aria-label="Decrement"
                    >
                      −
                    </button>
                    <span className="min-w-[2.5rem] text-center text-sm font-semibold text-gray-100">
                      {q.current_count}/{q.target_count}
                    </span>
                    <button
                      onClick={() => incrementWeekly.mutate(q.id)}
                      disabled={incrementWeekly.isPending}
                      className="flex h-7 w-7 items-center justify-center rounded-md bg-gray-800 text-sm font-bold text-gray-300 hover:bg-gray-700 disabled:opacity-40"
                      aria-label="Increment"
                    >
                      +
                    </button>
                  </div>

                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2">
                      <p className="text-sm font-medium text-gray-100">
                        {q.title}
                      </p>
                      {q.completed && (
                        <span className="rounded bg-indigo-900/60 px-1.5 py-0.5 text-xs font-medium text-indigo-400">
                          Done
                        </span>
                      )}
                      {!q.is_active && (
                        <span className="rounded bg-gray-800 px-1.5 py-0.5 text-xs text-gray-500">
                          inactive
                        </span>
                      )}
                    </div>
                    {q.description && (
                      <p className="text-xs text-gray-600">{q.description}</p>
                    )}
                    {/* Progress bar */}
                    <div className="mt-1.5 h-1.5 w-full overflow-hidden rounded-full bg-gray-800">
                      <div
                        className="h-full rounded-full bg-indigo-500 transition-all"
                        style={{
                          width: `${Math.min(
                            (q.current_count / q.target_count) * 100,
                            100,
                          )}%`,
                        }}
                      />
                    </div>
                  </div>

                  {q.xp_reward > 0 && (
                    <span
                      className={`shrink-0 text-xs font-medium ${
                        q.completed ? 'text-indigo-400' : 'text-yellow-600'
                      }`}
                    >
                      +{q.xp_reward} XP
                    </span>
                  )}

                  <button
                    onClick={() => handleEdit(q)}
                    className="shrink-0 text-xs text-gray-400 hover:text-gray-100"
                  >
                    Edit
                  </button>
                  <button
                    onClick={() => setDeleteConfirm(q.id)}
                    className="shrink-0 text-xs text-red-500 hover:text-red-400"
                  >
                    Delete
                  </button>
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>

      {/* Delete confirm modal */}
      {deleteConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div className="w-full max-w-sm rounded-xl border border-gray-700 bg-gray-900 p-6 shadow-2xl">
            <p className="mb-4 text-sm text-gray-300">
              Delete this weekly quest? All completions and XP events will be
              removed.
            </p>
            <div className="flex gap-2">
              <button
                onClick={() => handleDelete(deleteConfirm)}
                disabled={deleteQuest.isPending}
                className="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-500 disabled:opacity-50"
              >
                {deleteQuest.isPending ? 'Deleting…' : 'Delete'}
              </button>
              <button
                onClick={() => setDeleteConfirm(null)}
                className="rounded-lg bg-gray-800 px-4 py-2 text-sm font-medium text-gray-300 hover:bg-gray-700"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
