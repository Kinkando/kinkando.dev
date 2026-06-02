import { useState } from 'react'
import type { FoodLog, MealType } from '../../lib/api/types'
import {
  useCreateFoodLog,
  useUpdateFoodLog,
  useDeleteFoodLog,
} from '../../queries/useHealth'

type Props = {
  foodLogs: FoodLog[] | undefined
}

const MEAL_TYPES: MealType[] = ['breakfast', 'lunch', 'dinner', 'snack']

const MEAL_LABELS: Record<MealType, string> = {
  breakfast: 'Breakfast',
  lunch: 'Lunch',
  dinner: 'Dinner',
  snack: 'Snack',
}

const inputClass =
  'w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none'

const labelClass = 'mb-1 block text-xs font-medium text-gray-400'

function todayStr() {
  return new Date().toISOString().slice(0, 10)
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })
}

type FormState = {
  name: string
  meal_type: MealType
  calories: string
  protein_g: string
  carbs_g: string
  fat_g: string
  notes: string
  consumed_at: string
}

const defaultForm: FormState = {
  name: '',
  meal_type: 'breakfast',
  calories: '',
  protein_g: '',
  carbs_g: '',
  fat_g: '',
  notes: '',
  consumed_at: todayStr(),
}

function logToForm(log: FoodLog): FormState {
  return {
    name: log.name,
    meal_type: log.meal_type,
    calories: log.calories != null ? String(log.calories) : '',
    protein_g: log.protein_g != null ? String(log.protein_g) : '',
    carbs_g: log.carbs_g != null ? String(log.carbs_g) : '',
    fat_g: log.fat_g != null ? String(log.fat_g) : '',
    notes: log.notes ?? '',
    consumed_at: log.consumed_at.slice(0, 10),
  }
}

export default function FoodTab({ foodLogs }: Props) {
  const [form, setForm] = useState<FormState>(defaultForm)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [error, setError] = useState('')
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null)

  const createFoodLog = useCreateFoodLog()
  const updateFoodLog = useUpdateFoodLog()
  const deleteFoodLog = useDeleteFoodLog()

  const isEditing = editingId !== null
  const loading = createFoodLog.isPending || updateFoodLog.isPending

  function handleEdit(log: FoodLog) {
    setEditingId(log.id)
    setForm(logToForm(log))
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
      meal_type: form.meal_type,
      calories: form.calories ? parseInt(form.calories, 10) : null,
      protein_g: form.protein_g ? parseFloat(form.protein_g) : null,
      carbs_g: form.carbs_g ? parseFloat(form.carbs_g) : null,
      fat_g: form.fat_g ? parseFloat(form.fat_g) : null,
      notes: form.notes.trim() || null,
      consumed_at: form.consumed_at || todayStr(),
    }

    try {
      if (isEditing) {
        await updateFoodLog.mutateAsync({ id: editingId!, input: payload })
        setEditingId(null)
      } else {
        await createFoodLog.mutateAsync(payload)
      }
      setForm(defaultForm)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong.')
    }
  }

  async function handleDelete(id: string) {
    try {
      await deleteFoodLog.mutateAsync(id)
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
          {isEditing ? 'Edit Entry' : 'Log Food'}
        </h3>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <label className={labelClass}>Name</label>
              <input
                className={inputClass}
                placeholder="e.g. Grilled chicken"
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
              />
            </div>
            <div>
              <label className={labelClass}>Meal type</label>
              <select
                className={inputClass}
                value={form.meal_type}
                onChange={(e) =>
                  setForm({ ...form, meal_type: e.target.value as MealType })
                }
              >
                {MEAL_TYPES.map((t) => (
                  <option key={t} value={t}>
                    {MEAL_LABELS[t]}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className={labelClass}>Calories</label>
              <input
                className={inputClass}
                type="number"
                min="0"
                placeholder="Optional"
                value={form.calories}
                onChange={(e) => setForm({ ...form, calories: e.target.value })}
              />
            </div>
            <div>
              <label className={labelClass}>Protein (g)</label>
              <input
                className={inputClass}
                type="number"
                min="0"
                step="0.1"
                placeholder="Optional"
                value={form.protein_g}
                onChange={(e) =>
                  setForm({ ...form, protein_g: e.target.value })
                }
              />
            </div>
            <div>
              <label className={labelClass}>Carbs (g)</label>
              <input
                className={inputClass}
                type="number"
                min="0"
                step="0.1"
                placeholder="Optional"
                value={form.carbs_g}
                onChange={(e) => setForm({ ...form, carbs_g: e.target.value })}
              />
            </div>
            <div>
              <label className={labelClass}>Fat (g)</label>
              <input
                className={inputClass}
                type="number"
                min="0"
                step="0.1"
                placeholder="Optional"
                value={form.fat_g}
                onChange={(e) => setForm({ ...form, fat_g: e.target.value })}
              />
            </div>
            <div>
              <label className={labelClass}>Date</label>
              <input
                className={inputClass}
                type="date"
                value={form.consumed_at}
                onChange={(e) =>
                  setForm({ ...form, consumed_at: e.target.value })
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

          <div className="flex justify-end gap-2">
            <button
              type="submit"
              disabled={loading}
              className="cursor-pointer rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
            >
              {loading ? 'Saving…' : isEditing ? 'Update' : 'Log Food'}
            </button>
            {isEditing && (
              <button
                type="button"
                onClick={handleCancelEdit}
                className="cursor-pointer rounded-lg bg-gray-800 px-4 py-2 text-sm font-medium text-gray-300 hover:bg-gray-700"
              >
                Cancel
              </button>
            )}
          </div>
        </form>
      </div>

      {/* Food log list */}
      <div className="rounded-xl border border-gray-800 bg-gray-900">
        {!foodLogs || foodLogs.length === 0 ? (
          <p className="px-5 py-8 text-center text-sm text-gray-500">
            No food logged yet.
          </p>
        ) : (
          <ul className="divide-y divide-gray-800">
            {foodLogs.map((log) => (
              <li key={log.id} className="flex items-center gap-3 px-5 py-3.5">
                <span className="rounded bg-gray-800 px-2 py-0.5 text-xs text-indigo-400">
                  {MEAL_LABELS[log.meal_type]}
                </span>
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm text-gray-100">{log.name}</p>
                  <p className="text-xs text-gray-500">
                    {[
                      log.calories != null ? `${log.calories} kcal` : null,
                      log.protein_g != null ? `P ${log.protein_g}g` : null,
                      log.carbs_g != null ? `C ${log.carbs_g}g` : null,
                      log.fat_g != null ? `F ${log.fat_g}g` : null,
                      log.notes,
                    ]
                      .filter(Boolean)
                      .join(' · ')}
                  </p>
                </div>
                <span className="shrink-0 text-xs text-gray-600">
                  {formatDate(log.consumed_at)}
                </span>
                <button
                  onClick={() => handleEdit(log)}
                  className="shrink-0 cursor-pointer text-xs text-gray-400 hover:text-gray-100"
                >
                  Edit
                </button>
                <button
                  onClick={() => setDeleteConfirm(log.id)}
                  className="shrink-0 cursor-pointer text-xs text-red-500 hover:text-red-400"
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
              Delete this food entry? This cannot be undone.
            </p>
            <div className="flex justify-end gap-2">
              <button
                onClick={() => handleDelete(deleteConfirm)}
                disabled={deleteFoodLog.isPending}
                className="cursor-pointer rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-500 disabled:opacity-50"
              >
                {deleteFoodLog.isPending ? 'Deleting…' : 'Delete'}
              </button>
              <button
                onClick={() => setDeleteConfirm(null)}
                className="cursor-pointer rounded-lg bg-gray-800 px-4 py-2 text-sm font-medium text-gray-300 hover:bg-gray-700"
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
