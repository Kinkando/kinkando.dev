import { useState } from 'react'
import type { SleepLog } from '../../lib/api/types'
import {
  useCreateSleepLog,
  useUpdateSleepLog,
  useDeleteSleepLog,
  useSleepLogs,
} from '../../queries/useHealth'
import { todayDate, addDays } from '../../lib/date'

const inputClass =
  'w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none'

const inputDisabledClass =
  'w-full rounded-lg border border-gray-700 bg-gray-900 px-3 py-2 text-sm text-gray-500 cursor-not-allowed'

const labelClass = 'mb-1 block text-xs font-medium text-gray-400'

/** Default bedtime = 22:00 local time yesterday */
function defaultBedtime() {
  const d = new Date()
  d.setDate(d.getDate() - 1)
  return `${d.toISOString().slice(0, 10)}T22:00`
}

/** Default wake time = 06:00 local time today */
function defaultWakeTime() {
  return `${todayDate()}T06:00`
}

function formatDuration(minutes: number): string {
  const h = Math.floor(minutes / 60)
  const m = minutes % 60
  return m > 0 ? `${h}h ${m}m` : `${h}h`
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })
}

/** Convert a datetime-local value (YYYY-MM-DDTHH:mm) to RFC3339 (local tz offset) */
function localToRFC3339(localDT: string): string {
  if (!localDT) return ''
  const d = new Date(localDT)
  if (isNaN(d.getTime())) return ''
  const pad = (n: number) => String(n).padStart(2, '0')
  const off = -d.getTimezoneOffset()
  const sign = off >= 0 ? '+' : '-'
  const absOff = Math.abs(off)
  return (
    `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}` +
    `T${pad(d.getHours())}:${pad(d.getMinutes())}:00` +
    `${sign}${pad(Math.floor(absOff / 60))}:${pad(absOff % 60)}`
  )
}

/** Parse an RFC3339 string back to datetime-local format (YYYY-MM-DDTHH:mm) */
function rfc3339ToLocal(rfc: string): string {
  if (!rfc) return ''
  const d = new Date(rfc)
  if (isNaN(d.getTime())) return ''
  const pad = (n: number) => String(n).padStart(2, '0')
  return (
    `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}` +
    `T${pad(d.getHours())}:${pad(d.getMinutes())}`
  )
}

type FormState = {
  started_at: string // datetime-local YYYY-MM-DDTHH:mm
  ended_at: string // datetime-local YYYY-MM-DDTHH:mm
  score: string // 0–100
  notes: string
  logged_at: string // YYYY-MM-DD
}

const defaultForm: FormState = {
  started_at: defaultBedtime(),
  ended_at: defaultWakeTime(),
  score: '',
  notes: '',
  logged_at: todayDate(),
}

function logToForm(log: SleepLog): FormState {
  return {
    started_at: rfc3339ToLocal(log.started_at),
    ended_at: rfc3339ToLocal(log.ended_at),
    score: log.score != null ? String(log.score) : '',
    notes: log.notes ?? '',
    logged_at: log.logged_at.slice(0, 10),
  }
}

function scoreColor(score: number): string {
  if (score >= 80) return 'text-green-400'
  if (score >= 60) return 'text-yellow-400'
  return 'text-red-400'
}

export default function SleepTab() {
  const defaultFrom = addDays(todayDate(), -29)
  const defaultTo = todayDate()

  const [from, setFrom] = useState(defaultFrom)
  const [to, setTo] = useState(defaultTo)

  const { data: sleepLogs } = useSleepLogs({ from, to })

  const [form, setForm] = useState<FormState>(defaultForm)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [error, setError] = useState('')
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null)

  const createSleepLog = useCreateSleepLog()
  const updateSleepLog = useUpdateSleepLog()
  const deleteSleepLog = useDeleteSleepLog()

  const isEditing = editingId !== null
  const loading = createSleepLog.isPending || updateSleepLog.isPending

  const loggedToday =
    sleepLogs?.some((l) => l.logged_at.slice(0, 10) === todayDate()) ?? false

  // The log currently being edited (to determine today-gating)
  const editingLog = editingId
    ? sleepLogs?.find((l) => l.id === editingId)
    : null
  const editingIsToday = editingLog?.logged_at.slice(0, 10) === todayDate()

  function handleEdit(log: SleepLog) {
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

    if (!form.started_at || !form.ended_at) {
      setError('Bedtime and wake time are required.')
      return
    }

    const startedAtRFC = localToRFC3339(form.started_at)
    const endedAtRFC = localToRFC3339(form.ended_at)

    if (!startedAtRFC || !endedAtRFC) {
      setError('Invalid date/time format.')
      return
    }

    const scoreVal = form.score ? parseInt(form.score, 10) : null
    if (scoreVal !== null && (scoreVal < 0 || scoreVal > 100)) {
      setError('Score must be between 0 and 100.')
      return
    }

    const payload = {
      started_at: startedAtRFC,
      ended_at: endedAtRFC,
      score: scoreVal,
      notes: form.notes.trim() || null,
      logged_at: form.logged_at || undefined,
    }

    try {
      if (isEditing) {
        await updateSleepLog.mutateAsync({ id: editingId!, input: payload })
        setEditingId(null)
      } else {
        await createSleepLog.mutateAsync(payload)
      }
      setForm(defaultForm)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong.')
    }
  }

  async function handleDelete(id: string) {
    try {
      await deleteSleepLog.mutateAsync(id)
      setDeleteConfirm(null)
    } catch {
      // ignore — mutation error shown in console
    }
  }

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

      {/* Create / edit form */}
      <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
        <h3 className="mb-4 text-sm font-medium text-gray-300">
          {isEditing
            ? editingIsToday
              ? 'Edit Sleep Entry'
              : 'Edit Note'
            : 'Log Sleep'}
        </h3>
        {loggedToday && !isEditing ? (
          <p className="text-sm text-indigo-400">
            ✓ You've already logged sleep for today.
          </p>
        ) : (
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <label className={labelClass}>Bedtime</label>
                {isEditing && !editingIsToday ? (
                  <input
                    className={inputDisabledClass}
                    type="datetime-local"
                    value={form.started_at}
                    disabled
                    readOnly
                  />
                ) : (
                  <input
                    className={inputClass}
                    type="datetime-local"
                    value={form.started_at}
                    onChange={(e) =>
                      setForm({ ...form, started_at: e.target.value })
                    }
                  />
                )}
              </div>
              <div>
                <label className={labelClass}>Wake time</label>
                {isEditing && !editingIsToday ? (
                  <input
                    className={inputDisabledClass}
                    type="datetime-local"
                    value={form.ended_at}
                    disabled
                    readOnly
                  />
                ) : (
                  <input
                    className={inputClass}
                    type="datetime-local"
                    value={form.ended_at}
                    onChange={(e) =>
                      setForm({ ...form, ended_at: e.target.value })
                    }
                  />
                )}
              </div>
              <div>
                <label className={labelClass}>
                  Sleep score (0–100, Samsung Health)
                </label>
                {isEditing && !editingIsToday ? (
                  <input
                    className={inputDisabledClass}
                    type="number"
                    value={form.score}
                    disabled
                    readOnly
                  />
                ) : (
                  <input
                    className={inputClass}
                    type="number"
                    min="0"
                    max="100"
                    placeholder="Optional"
                    value={form.score}
                    onChange={(e) =>
                      setForm({ ...form, score: e.target.value })
                    }
                  />
                )}
              </div>
              <div>
                <label className={labelClass}>Night-of date</label>
                {isEditing && !editingIsToday ? (
                  <input
                    className={inputDisabledClass}
                    type="date"
                    value={form.logged_at}
                    disabled
                    readOnly
                  />
                ) : (
                  <input
                    className={inputClass}
                    type="date"
                    value={form.logged_at}
                    onChange={(e) =>
                      setForm({ ...form, logged_at: e.target.value })
                    }
                  />
                )}
              </div>
              <div className="sm:col-span-2">
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
                {loading ? 'Saving…' : isEditing ? 'Update' : 'Log Sleep'}
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
        )}
      </div>

      {/* Sleep log list */}
      <div className="rounded-xl border border-gray-800 bg-gray-900">
        {!sleepLogs || sleepLogs.length === 0 ? (
          <p className="px-5 py-8 text-center text-sm text-gray-500">
            No sleep logged in this range.
          </p>
        ) : (
          <ul className="divide-y divide-gray-800">
            {sleepLogs.map((log) => {
              const isToday = log.logged_at.slice(0, 10) === todayDate()
              return (
                <li
                  key={log.id}
                  className="flex items-center gap-3 px-5 py-3.5"
                >
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium text-gray-100">
                        {formatDuration(log.duration_minutes)}
                      </span>
                      {log.score != null && (
                        <span
                          className={`rounded bg-gray-800 px-2 py-0.5 text-xs font-medium ${scoreColor(log.score)}`}
                        >
                          {log.score}
                        </span>
                      )}
                    </div>
                    <p className="text-xs text-gray-500">
                      {[
                        new Date(log.started_at).toLocaleTimeString(undefined, {
                          hour: '2-digit',
                          minute: '2-digit',
                        }) +
                          ' → ' +
                          new Date(log.ended_at).toLocaleTimeString(undefined, {
                            hour: '2-digit',
                            minute: '2-digit',
                          }),
                        log.notes,
                      ]
                        .filter(Boolean)
                        .join(' · ')}
                    </p>
                  </div>
                  <span className="shrink-0 text-xs text-gray-600">
                    {formatDate(log.logged_at)}
                  </span>
                  <button
                    onClick={() => handleEdit(log)}
                    className="shrink-0 cursor-pointer text-xs text-gray-400 hover:text-gray-100"
                  >
                    Edit
                  </button>
                  {isToday && (
                    <button
                      onClick={() => setDeleteConfirm(log.id)}
                      className="shrink-0 cursor-pointer text-xs text-red-500 hover:text-red-400"
                    >
                      Delete
                    </button>
                  )}
                </li>
              )
            })}
          </ul>
        )}
      </div>

      {/* Delete confirm */}
      {deleteConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div className="w-full max-w-sm rounded-xl border border-gray-700 bg-gray-900 p-6 shadow-2xl">
            <p className="mb-4 text-sm text-gray-300">
              Delete this sleep entry? This cannot be undone.
            </p>
            <div className="flex justify-end gap-2">
              <button
                onClick={() => handleDelete(deleteConfirm)}
                disabled={deleteSleepLog.isPending}
                className="cursor-pointer rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-500 disabled:opacity-50"
              >
                {deleteSleepLog.isPending ? 'Deleting…' : 'Delete'}
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
