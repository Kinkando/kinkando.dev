import { useState, useEffect } from 'react'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  ReferenceLine,
} from 'recharts'
import type { HealthProfile } from '../../lib/api/types'
import {
  useWeightLogs,
  useCreateWeightLog,
  useUpdateWeightLog,
  useDeleteWeightLog,
} from '../../queries/useHealth'
import { todayDate, addDays } from '../../lib/date'

type Props = {
  profile: HealthProfile | null | undefined
}

const inputClass =
  'w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none'

const inputDisabledClass =
  'w-full rounded-lg border border-gray-700 bg-gray-900 px-3 py-2 text-sm text-gray-500 cursor-not-allowed'

const labelClass = 'mb-1 block text-xs font-medium text-gray-400'

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })
}

type FormState = {
  weight: string
  note: string
  loggedAt: string
}

export default function ProgressTab({ profile }: Props) {
  const defaultFrom = addDays(todayDate(), -29)
  const defaultTo = todayDate()

  const [from, setFrom] = useState(defaultFrom)
  const [to, setTo] = useState(defaultTo)

  const { data: weightLogs } = useWeightLogs({ from, to })

  const [editingId, setEditingId] = useState<string | null>(null)
  const [form, setForm] = useState<FormState>({
    weight: '',
    note: '',
    loggedAt: todayDate(),
  })
  const [createNote, setCreateNote] = useState('')
  const [createWeight, setCreateWeight] = useState('')
  const [createLoggedAt, setCreateLoggedAt] = useState(todayDate())
  const [error, setError] = useState('')
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null)

  const createWeightMutation = useCreateWeightLog()
  const updateWeightMutation = useUpdateWeightLog()
  const deleteWeightMutation = useDeleteWeightLog()

  const isEditing = editingId !== null

  const chartData =
    weightLogs?.map((w) => ({
      date: new Date(w.logged_at).toLocaleDateString(undefined, {
        month: 'short',
        day: 'numeric',
      }),
      weight: w.weight,
    })) ?? []

  // Latest weight across ALL dates (from the filtered window)
  const latestWeight =
    weightLogs && weightLogs.length > 0
      ? weightLogs[weightLogs.length - 1].weight
      : null

  const goalLabel =
    profile?.goal === 'lose_weight'
      ? 'Goal: Lose Weight'
      : profile?.goal === 'maintain'
        ? 'Goal: Maintain'
        : profile?.goal === 'gain_muscle'
          ? 'Goal: Gain Muscle'
          : null

  // Pre-fill create weight input with the latest logged weight once data loads
  useEffect(() => {
    if (latestWeight != null) {
      setCreateWeight((cur) => (cur === '' ? String(latestWeight) : cur))
    }
  }, [latestWeight])

  // True when a weight entry already exists for today (Asia/Bangkok date)
  const loggedToday =
    weightLogs?.some((w) => w.logged_at.slice(0, 10) === todayDate()) ?? false

  function handleEdit(w: NonNullable<typeof weightLogs>[number]) {
    setEditingId(w.id)
    setForm({
      weight: String(w.weight),
      note: w.note ?? '',
      loggedAt: w.logged_at.slice(0, 10),
    })
    setError('')
  }

  function handleCancelEdit() {
    setEditingId(null)
    setForm({ weight: '', note: '', loggedAt: todayDate() })
    setError('')
  }

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    const w = parseFloat(createWeight)
    if (!createWeight || isNaN(w) || w <= 0) {
      setError('Enter a valid weight.')
      return
    }
    try {
      await createWeightMutation.mutateAsync({
        weight: w,
        note: createNote.trim() || null,
        logged_at: createLoggedAt || todayDate(),
      })
      setCreateWeight('')
      setCreateNote('')
      setCreateLoggedAt(todayDate())
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong.')
    }
  }

  async function handleUpdate(e: React.FormEvent) {
    e.preventDefault()
    if (!editingId) return
    setError('')
    const editingLog = weightLogs?.find((w) => w.id === editingId)
    const isToday = editingLog?.logged_at.slice(0, 10) === todayDate()

    // For today: use form weight; for non-today: preserve original weight
    const w = isToday ? parseFloat(form.weight) : (editingLog?.weight ?? 0)
    if (isToday && (isNaN(w) || w <= 0)) {
      setError('Enter a valid weight.')
      return
    }
    try {
      await updateWeightMutation.mutateAsync({
        id: editingId,
        input: {
          weight: w,
          note: form.note.trim() || null,
          logged_at: isToday
            ? form.loggedAt
            : editingLog?.logged_at.slice(0, 10),
        },
      })
      setEditingId(null)
      setForm({ weight: '', note: '', loggedAt: todayDate() })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong.')
    }
  }

  async function handleDelete(id: string) {
    try {
      await deleteWeightMutation.mutateAsync(id)
      setDeleteConfirm(null)
    } catch {
      // ignore
    }
  }

  const editingLog = editingId
    ? weightLogs?.find((w) => w.id === editingId)
    : null
  const editingIsToday = editingLog?.logged_at.slice(0, 10) === todayDate()

  return (
    <div className="space-y-6">
      {/* Date range filter */}
      <div className="flex flex-wrap items-end gap-3 rounded-xl border border-gray-800 bg-gray-900 p-5">
        <div>
          <label className={labelClass}>From</label>
          <input
            className={inputClass}
            type="date"
            value={from}
            max={to}
            onChange={(e) => setFrom(e.target.value)}
          />
        </div>
        <div>
          <label className={labelClass}>To</label>
          <input
            className={inputClass}
            type="date"
            value={to}
            min={from}
            max={todayDate()}
            onChange={(e) => setTo(e.target.value)}
          />
        </div>
      </div>

      {/* Add / Edit form */}
      <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
        <h3 className="mb-4 text-sm font-medium text-gray-300">
          {isEditing
            ? editingIsToday
              ? 'Edit Weight Entry'
              : 'Edit Note'
            : 'Log Weight'}
        </h3>

        {isEditing ? (
          <form onSubmit={handleUpdate} className="space-y-4">
            <div className="flex flex-wrap items-end gap-3">
              <div className="w-36">
                <label className={labelClass}>Weight (kg)</label>
                {editingIsToday ? (
                  <input
                    className={inputClass}
                    type="number"
                    step="0.05"
                    min="1"
                    value={form.weight}
                    onChange={(e) =>
                      setForm({ ...form, weight: e.target.value })
                    }
                  />
                ) : (
                  <input
                    className={inputDisabledClass}
                    type="number"
                    value={form.weight}
                    disabled
                    readOnly
                  />
                )}
              </div>
              <div className="w-56">
                <label className={labelClass}>Note</label>
                <input
                  className={inputClass}
                  type="text"
                  placeholder="Optional"
                  value={form.note}
                  onChange={(e) => setForm({ ...form, note: e.target.value })}
                />
              </div>
              {editingIsToday && (
                <div className="w-40">
                  <label className={labelClass}>Date</label>
                  <input
                    className={inputClass}
                    type="date"
                    value={form.loggedAt}
                    onChange={(e) =>
                      setForm({ ...form, loggedAt: e.target.value })
                    }
                  />
                </div>
              )}
            </div>
            {error && <p className="text-sm text-red-400">{error}</p>}
            <div className="flex justify-end gap-2">
              <button
                type="submit"
                disabled={updateWeightMutation.isPending}
                className="cursor-pointer rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
              >
                {updateWeightMutation.isPending ? 'Saving…' : 'Update'}
              </button>
              <button
                type="button"
                onClick={handleCancelEdit}
                className="cursor-pointer rounded-lg bg-gray-800 px-4 py-2 text-sm font-medium text-gray-300 hover:bg-gray-700"
              >
                Cancel
              </button>
            </div>
          </form>
        ) : loggedToday ? (
          <p className="text-sm text-indigo-400">
            ✓ You've already logged your weight today
            {latestWeight != null ? ` (${latestWeight} kg)` : ''}.
          </p>
        ) : (
          <>
            <form
              onSubmit={handleCreate}
              className="flex flex-wrap items-end gap-3"
            >
              <div className="w-36">
                <label className={labelClass}>Weight (kg)</label>
                <input
                  className={inputClass}
                  type="number"
                  step="0.05"
                  min="1"
                  placeholder="e.g. 72.5"
                  value={createWeight}
                  onChange={(e) => setCreateWeight(e.target.value)}
                />
              </div>
              <div className="w-56">
                <label className={labelClass}>Note</label>
                <input
                  className={inputClass}
                  type="text"
                  placeholder="Optional"
                  value={createNote}
                  onChange={(e) => setCreateNote(e.target.value)}
                />
              </div>
              <div className="w-40">
                <label className={labelClass}>Date</label>
                <input
                  className={inputClass}
                  type="date"
                  value={createLoggedAt}
                  onChange={(e) => setCreateLoggedAt(e.target.value)}
                />
              </div>
              <button
                type="submit"
                disabled={createWeightMutation.isPending}
                className="cursor-pointer rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
              >
                {createWeightMutation.isPending ? 'Saving…' : 'Add'}
              </button>
            </form>
            {error && <p className="mt-2 text-sm text-red-400">{error}</p>}
          </>
        )}
      </div>

      {/* Chart */}
      {chartData.length > 0 ? (
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
          <div className="mb-3 flex items-center justify-between">
            <h3 className="text-sm font-medium text-gray-400">
              Weight Over Time
            </h3>
            {goalLabel && (
              <span className="text-xs text-indigo-400">{goalLabel}</span>
            )}
          </div>
          <ResponsiveContainer width="100%" height={240}>
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
              <XAxis
                dataKey="date"
                tick={{ fill: '#9ca3af', fontSize: 11 }}
                tickLine={false}
                axisLine={false}
              />
              <YAxis
                domain={['auto', 'auto']}
                tick={{ fill: '#9ca3af', fontSize: 11 }}
                tickLine={false}
                axisLine={false}
                width={36}
              />
              <Tooltip
                contentStyle={{
                  background: '#1f2937',
                  border: '1px solid #374151',
                  borderRadius: '8px',
                  color: '#f3f4f6',
                  fontSize: 12,
                }}
                formatter={(value) => [`${value} kg`, 'Weight']}
              />
              {latestWeight != null && profile?.goal === 'maintain' && (
                <ReferenceLine
                  y={latestWeight}
                  stroke="#6366f1"
                  strokeDasharray="4 2"
                  label={{
                    value: 'Target',
                    fill: '#818cf8',
                    fontSize: 11,
                  }}
                />
              )}
              <Line
                type="monotone"
                dataKey="weight"
                stroke="#6366f1"
                strokeWidth={2}
                dot={{ fill: '#6366f1', r: 3 }}
                activeDot={{ r: 5 }}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      ) : (
        <div className="rounded-xl border border-gray-800 bg-gray-900 px-5 py-10 text-center">
          <p className="text-sm text-gray-500">
            No weight entries in this range. Adjust the date filter or log your
            first entry above.
          </p>
        </div>
      )}

      {/* Weight log table */}
      {weightLogs && weightLogs.length > 0 && (
        <div className="rounded-xl border border-gray-800 bg-gray-900">
          <ul className="divide-y divide-gray-800">
            {[...weightLogs].reverse().map((w) => {
              const isToday = w.logged_at.slice(0, 10) === todayDate()
              return (
                <li key={w.id} className="flex items-center gap-3 px-5 py-3">
                  <div className="min-w-0 flex-1">
                    <span className="text-sm text-gray-100">{w.weight} kg</span>
                    {w.note && (
                      <p className="text-xs text-gray-500">{w.note}</p>
                    )}
                  </div>
                  <span className="shrink-0 text-xs text-gray-600">
                    {formatDate(w.logged_at)}
                  </span>
                  <button
                    onClick={() => handleEdit(w)}
                    className="shrink-0 cursor-pointer text-xs text-gray-400 hover:text-gray-100"
                  >
                    Edit
                  </button>
                  {isToday && (
                    <button
                      onClick={() => setDeleteConfirm(w.id)}
                      className="shrink-0 cursor-pointer text-xs text-red-500 hover:text-red-400"
                    >
                      Delete
                    </button>
                  )}
                </li>
              )
            })}
          </ul>
        </div>
      )}

      {/* Delete confirm */}
      {deleteConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div className="w-full max-w-sm rounded-xl border border-gray-700 bg-gray-900 p-6 shadow-2xl">
            <p className="mb-4 text-sm text-gray-300">
              Delete this weight entry? This cannot be undone.
            </p>
            <div className="flex justify-end gap-2">
              <button
                onClick={() => handleDelete(deleteConfirm)}
                disabled={deleteWeightMutation.isPending}
                className="cursor-pointer rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-500 disabled:opacity-50"
              >
                {deleteWeightMutation.isPending ? 'Deleting…' : 'Delete'}
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
