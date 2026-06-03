import { useState } from 'react'
import type {
  DailyQuestStatus,
  CreateQuestInput,
  UpdateQuestInput,
  SourceType,
} from '../../lib/api/types'
import {
  useIncrementQuest,
  useDecrementQuest,
  useCreateQuest,
  useUpdateQuest,
  useDeleteQuest,
  useActivateQuest,
  useDeactivateQuest,
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
  target_count: string
  source_type: SourceType
}

const defaultForm: FormState = {
  title: '',
  description: '',
  xp_reward: '10',
  target_count: '1',
  source_type: 'manual',
}

function questToForm(q: DailyQuestStatus): FormState {
  return {
    title: q.title,
    description: q.description,
    xp_reward: String(q.xp_reward),
    target_count: String(q.target_count),
    source_type: q.source_type,
  }
}

const SOURCE_LABELS: Record<SourceType, string> = {
  manual: 'Manual',
  medicine: 'Medicine (auto)',
  workout: 'Workout (auto)',
  supplement: 'Supplement (auto)',
}

export default function DailyTab({ daily }: Props) {
  const [form, setForm] = useState<FormState>(defaultForm)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [formError, setFormError] = useState('')
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null)

  const incrementQuest = useIncrementQuest()
  const decrementQuest = useDecrementQuest()
  const createQuest = useCreateQuest()
  const updateQuest = useUpdateQuest()
  const deleteQuest = useDeleteQuest()
  const activateQuest = useActivateQuest()
  const deactivateQuest = useDeactivateQuest()

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

  function handleToggleActive(q: DailyQuestStatus) {
    if (q.is_active) {
      deactivateQuest.mutate(q.id)
    } else {
      activateQuest.mutate(q.id)
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
    const target = parseInt(form.target_count, 10)
    if (isNaN(target) || target < 1) {
      setFormError('Target count must be at least 1.')
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
          target_count: target,
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
                <option value="supplement">Supplement (auto on take)</option>
                <option value="workout">Workout (auto on finish)</option>
              </select>
            </div>
            <div>
              <label className={labelClass}>Target count / day</label>
              <input
                className={inputClass}
                type="number"
                min="1"
                placeholder="1"
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
                : form.source_type === 'supplement'
                  ? 'take a supplement'
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
              const isAutoQ = q.source_type !== 'manual'
              return (
                <li
                  key={q.id}
                  className={`group px-5 py-4 ${!q.is_active ? 'opacity-50' : ''}`}
                >
                  <div className="flex items-center gap-3">
                    {/* Count controls — only for manual quests */}
                    {isAutoQ ? (
                      <div className="flex items-center gap-1">
                        <span className="min-w-[2.5rem] text-center text-sm font-semibold text-gray-100">
                          {q.current_count}/{q.target_count}
                        </span>
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

                    <div className="min-w-0 flex-1">
                      <div className="flex flex-wrap items-center gap-1.5">
                        <p className="text-sm font-medium text-gray-100">
                          {q.title}
                        </p>
                        {isAutoQ && (
                          <span className="rounded bg-gray-800 px-1.5 py-0.5 text-xs text-gray-500">
                            ⚙ {SOURCE_LABELS[q.source_type]}
                          </span>
                        )}
                        {q.completed && (
                          <span className="rounded bg-sky-900/60 px-1.5 py-0.5 text-xs font-medium text-sky-400">
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
                          className="h-full rounded-full bg-sky-500 transition-all"
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

                    <button
                      role="switch"
                      aria-checked={q.is_active}
                      onClick={() => handleToggleActive(q)}
                      className={`relative h-5 w-9 shrink-0 cursor-pointer rounded-full transition-colors focus:outline-none disabled:opacity-40 ${q.is_active ? 'bg-indigo-600' : 'bg-gray-700'}`}
                    >
                      <span
                        className={`absolute top-0.5 left-0.5 h-4 w-4 rounded-full bg-white shadow transition-transform ${q.is_active ? 'translate-x-4' : 'translate-x-0'}`}
                      />
                    </button>
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
