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
import type { WeightLog, HealthProfile } from '../../lib/api/types'
import { useCreateWeightLog, useDeleteWeightLog } from '../../queries/useHealth'
import { todayDate } from '../../lib/date'

type Props = {
  weightLogs: WeightLog[] | undefined
  profile: HealthProfile | null | undefined
}

const inputClass =
  'w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none'

const labelClass = 'mb-1 block text-xs font-medium text-gray-400'

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })
}

export default function ProgressTab({ weightLogs, profile }: Props) {
  const [weight, setWeight] = useState('')
  const [loggedAt, setLoggedAt] = useState(todayDate)
  const [error, setError] = useState('')
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null)

  const createWeight = useCreateWeightLog()
  const deleteWeight = useDeleteWeightLog()

  const chartData =
    weightLogs?.map((w) => ({
      date: new Date(w.logged_at).toLocaleDateString(undefined, {
        month: 'short',
        day: 'numeric',
      }),
      weight: w.weight,
    })) ?? []

  // Derive an approximate goal reference weight for the chart based on goal direction
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

  // Pre-fill weight input with the latest logged weight once data loads
  useEffect(() => {
    if (latestWeight != null) {
      setWeight((cur) => (cur === '' ? String(latestWeight) : cur))
    }
  }, [latestWeight])

  // True when a weight entry already exists for today (Asia/Bangkok date)
  const loggedToday =
    weightLogs?.some((w) => w.logged_at.slice(0, 10) === todayDate()) ?? false

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    const w = parseFloat(weight)
    if (!weight || isNaN(w) || w <= 0) {
      setError('Enter a valid weight.')
      return
    }
    try {
      await createWeight.mutateAsync({
        weight: w,
        logged_at: loggedAt || todayDate(),
      })
      setWeight('')
      setLoggedAt(todayDate())
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong.')
    }
  }

  async function handleDelete(id: string) {
    try {
      await deleteWeight.mutateAsync(id)
      setDeleteConfirm(null)
    } catch {
      // ignore
    }
  }

  return (
    <div className="space-y-6">
      {/* Add weight form */}
      <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
        <h3 className="mb-4 text-sm font-medium text-gray-300">Log Weight</h3>
        {loggedToday ? (
          <p className="text-sm text-indigo-400">
            ✓ You've already logged your weight today
            {latestWeight != null ? ` (${latestWeight} kg)` : ''}.
          </p>
        ) : (
          <>
            <form
              onSubmit={handleSubmit}
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
                  value={weight}
                  onChange={(e) => setWeight(e.target.value)}
                />
              </div>
              <div className="w-40">
                <label className={labelClass}>Date</label>
                <input
                  className={inputClass}
                  type="date"
                  value={loggedAt}
                  onChange={(e) => setLoggedAt(e.target.value)}
                />
              </div>
              <button
                type="submit"
                disabled={createWeight.isPending}
                className="cursor-pointer rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
              >
                {createWeight.isPending ? 'Saving…' : 'Add'}
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
            No weight entries yet. Log your first entry above.
          </p>
        </div>
      )}

      {/* Weight log table */}
      {weightLogs && weightLogs.length > 0 && (
        <div className="rounded-xl border border-gray-800 bg-gray-900">
          <ul className="divide-y divide-gray-800">
            {[...weightLogs].reverse().map((w) => (
              <li
                key={w.id}
                className="flex items-center justify-between px-5 py-3"
              >
                <span className="text-sm text-gray-100">{w.weight} kg</span>
                <span className="text-xs text-gray-500">
                  {formatDate(w.logged_at)}
                </span>
                <button
                  onClick={() => setDeleteConfirm(w.id)}
                  className="cursor-pointer text-xs text-red-500 hover:text-red-400"
                >
                  Delete
                </button>
              </li>
            ))}
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
                disabled={deleteWeight.isPending}
                className="cursor-pointer rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-500 disabled:opacity-50"
              >
                {deleteWeight.isPending ? 'Deleting…' : 'Delete'}
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
