import { useState } from 'react'
import type { HealthExercise, ExerciseType } from '../../lib/api/types'
import {
  useCreateExercise,
  useUpdateExercise,
  useDeleteExercise,
} from '../../queries/useHealth'

type Props = {
  exercises: HealthExercise[] | undefined
}

const EXERCISE_TYPES: ExerciseType[] = ['cardio', 'strength', 'flexibility']

const TYPE_LABELS: Record<ExerciseType, string> = {
  cardio: 'Cardio',
  strength: 'Strength',
  flexibility: 'Flexibility',
}

const inputClass =
  'w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none'

const labelClass = 'mb-1 block text-xs font-medium text-gray-400'

function todayStr() {
  return new Date().toISOString().slice(0, 10)
}

type FormState = {
  name: string
  type: ExerciseType
  duration_minutes: string
  calories: string
  notes: string
  performed_at: string
}

const defaultForm: FormState = {
  name: '',
  type: 'cardio',
  duration_minutes: '',
  calories: '',
  notes: '',
  performed_at: todayStr(),
}

function exerciseToForm(ex: HealthExercise): FormState {
  return {
    name: ex.name,
    type: ex.type,
    duration_minutes:
      ex.duration_minutes != null ? String(ex.duration_minutes) : '',
    calories: ex.calories != null ? String(ex.calories) : '',
    notes: ex.notes ?? '',
    performed_at: ex.performed_at.slice(0, 10),
  }
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })
}

export default function ExerciseTab({ exercises }: Props) {
  const [form, setForm] = useState<FormState>(defaultForm)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [error, setError] = useState('')
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null)

  const createExercise = useCreateExercise()
  const updateExercise = useUpdateExercise()
  const deleteExercise = useDeleteExercise()

  const isEditing = editingId !== null
  const loading = createExercise.isPending || updateExercise.isPending

  function handleEdit(ex: HealthExercise) {
    setEditingId(ex.id)
    setForm(exerciseToForm(ex))
    setError('')
  }

  function handleCancelEdit() {
    setEditingId(null)
    setForm(defaultForm)
    setError('')
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')

    if (!form.name.trim()) {
      setError('Name is required.')
      return
    }

    const payload = {
      name: form.name.trim(),
      type: form.type,
      duration_minutes: form.duration_minutes
        ? parseInt(form.duration_minutes, 10)
        : null,
      calories: form.calories ? parseInt(form.calories, 10) : null,
      notes: form.notes.trim() || null,
      performed_at: form.performed_at || todayStr(),
    }

    try {
      if (isEditing) {
        await updateExercise.mutateAsync({ id: editingId!, input: payload })
        setEditingId(null)
      } else {
        await createExercise.mutateAsync(payload)
      }
      setForm(defaultForm)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong.')
    }
  }

  async function handleDelete(id: string) {
    try {
      await deleteExercise.mutateAsync(id)
      setDeleteConfirm(null)
    } catch {
      // ignore — mutation error shown in console
    }
  }

  return (
    <div className="space-y-6">
      {/* Create / edit form */}
      <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
        <h3 className="mb-4 text-sm font-medium text-gray-300">
          {isEditing ? 'Edit Exercise' : 'Log Exercise'}
        </h3>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <label className={labelClass}>Name</label>
              <input
                className={inputClass}
                placeholder="e.g. Morning Run"
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
              />
            </div>
            <div>
              <label className={labelClass}>Type</label>
              <select
                className={inputClass}
                value={form.type}
                onChange={(e) =>
                  setForm({ ...form, type: e.target.value as ExerciseType })
                }
              >
                {EXERCISE_TYPES.map((t) => (
                  <option key={t} value={t}>
                    {TYPE_LABELS[t]}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className={labelClass}>Duration (min)</label>
              <input
                className={inputClass}
                type="number"
                min="1"
                placeholder="Optional"
                value={form.duration_minutes}
                onChange={(e) =>
                  setForm({ ...form, duration_minutes: e.target.value })
                }
              />
            </div>
            <div>
              <label className={labelClass}>Calories</label>
              <input
                className={inputClass}
                type="number"
                min="1"
                placeholder="Optional"
                value={form.calories}
                onChange={(e) => setForm({ ...form, calories: e.target.value })}
              />
            </div>
            <div>
              <label className={labelClass}>Date</label>
              <input
                className={inputClass}
                type="date"
                value={form.performed_at}
                onChange={(e) =>
                  setForm({ ...form, performed_at: e.target.value })
                }
              />
            </div>
            <div>
              <label className={labelClass}>Notes</label>
              <input
                className={inputClass}
                placeholder="Optional"
                value={form.notes}
                onChange={(e) => setForm({ ...form, notes: e.target.value })}
              />
            </div>
          </div>

          {error && <p className="text-sm text-red-400">{error}</p>}

          <div className="flex gap-2">
            <button
              type="submit"
              disabled={loading}
              className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
            >
              {loading ? 'Saving…' : isEditing ? 'Update' : 'Add Exercise'}
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

      {/* Exercise list */}
      <div className="rounded-xl border border-gray-800 bg-gray-900">
        {!exercises || exercises.length === 0 ? (
          <p className="px-5 py-8 text-center text-sm text-gray-500">
            No exercises logged yet.
          </p>
        ) : (
          <ul className="divide-y divide-gray-800">
            {exercises.map((ex) => (
              <li key={ex.id} className="flex items-center gap-3 px-5 py-3.5">
                <span className="rounded bg-gray-800 px-2 py-0.5 text-xs text-indigo-400">
                  {TYPE_LABELS[ex.type]}
                </span>
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm text-gray-100">{ex.name}</p>
                  <p className="text-xs text-gray-500">
                    {[
                      ex.duration_minutes != null
                        ? `${ex.duration_minutes} min`
                        : null,
                      ex.calories != null ? `${ex.calories} kcal` : null,
                      ex.notes,
                    ]
                      .filter(Boolean)
                      .join(' · ')}
                  </p>
                </div>
                <span className="shrink-0 text-xs text-gray-600">
                  {formatDate(ex.performed_at)}
                </span>
                <button
                  onClick={() => handleEdit(ex)}
                  className="shrink-0 text-xs text-gray-400 hover:text-gray-100"
                >
                  Edit
                </button>
                <button
                  onClick={() => setDeleteConfirm(ex.id)}
                  className="shrink-0 text-xs text-red-500 hover:text-red-400"
                >
                  Delete
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>

      {/* Delete confirm */}
      {deleteConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div className="w-full max-w-sm rounded-xl border border-gray-700 bg-gray-900 p-6 shadow-2xl">
            <p className="mb-4 text-sm text-gray-300">
              Delete this exercise? This cannot be undone.
            </p>
            <div className="flex gap-2">
              <button
                onClick={() => handleDelete(deleteConfirm)}
                disabled={deleteExercise.isPending}
                className="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-500 disabled:opacity-50"
              >
                {deleteExercise.isPending ? 'Deleting…' : 'Delete'}
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
