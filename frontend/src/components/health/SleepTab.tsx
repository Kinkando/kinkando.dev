import { useState } from 'react'
import type { SleepLog } from '../../lib/api/types'
import {
  useCreateSleepLog,
  useUpdateSleepLog,
  useDeleteSleepLog,
} from '../../queries/useHealth'

type Props = {
  sleepLogs: SleepLog[] | undefined
}

const inputClass =
  'w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none'

const labelClass = 'mb-1 block text-xs font-medium text-gray-400'

/** Today's date as YYYY-MM-DD in local time */
function todayStr() {
  return new Date().toISOString().slice(0, 10)
}

/** Default bedtime = 22:00 local time today */
function defaultBedtime() {
  return `${todayStr()}T22:00`
}

/** Default wake time = 06:00 local time today */
function defaultWakeTime() {
  return `${todayStr()}T06:00`
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
  logged_at: todayStr(),
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

export default function SleepTab({ sleepLogs }: Props) {
  const [form, setForm] = useState<FormState>(defaultForm)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [error, setError] = useState('')
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null)

  const createSleepLog = useCreateSleepLog()
  const updateSleepLog = useUpdateSleepLog()
  const deleteSleepLog = useDeleteSleepLog()

  const isEditing = editingId !== null
  const loading = createSleepLog.isPending || updateSleepLog.isPending

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
      {/* Create / edit form */}
      <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
        <h3 className="mb-4 text-sm font-medium text-gray-300">
          {isEditing ? 'Edit Sleep Entry' : 'Log Sleep'}
        </h3>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <label className={labelClass}>Bedtime</label>
              <input
                className={inputClass}
                type="datetime-local"
                value={form.started_at}
                onChange={(e) =>
                  setForm({ ...form, started_at: e.target.value })
                }
              />
            </div>
            <div>
              <label className={labelClass}>Wake time</label>
              <input
                className={inputClass}
                type="datetime-local"
                value={form.ended_at}
                onChange={(e) => setForm({ ...form, ended_at: e.target.value })}
              />
            </div>
            <div>
              <label className={labelClass}>
                Sleep score (0–100, Samsung Health)
              </label>
              <input
                className={inputClass}
                type="number"
                min="0"
                max="100"
                placeholder="Optional"
                value={form.score}
                onChange={(e) => setForm({ ...form, score: e.target.value })}
              />
            </div>
            <div>
              <label className={labelClass}>Night-of date</label>
              <input
                className={inputClass}
                type="date"
                value={form.logged_at}
                onChange={(e) =>
                  setForm({ ...form, logged_at: e.target.value })
                }
              />
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

          <div className="flex gap-2">
            <button
              type="submit"
              disabled={loading}
              className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
            >
              {loading ? 'Saving…' : isEditing ? 'Update' : 'Log Sleep'}
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

      {/* Sleep log list */}
      <div className="rounded-xl border border-gray-800 bg-gray-900">
        {!sleepLogs || sleepLogs.length === 0 ? (
          <p className="px-5 py-8 text-center text-sm text-gray-500">
            No sleep logged yet.
          </p>
        ) : (
          <ul className="divide-y divide-gray-800">
            {sleepLogs.map((log) => (
              <li key={log.id} className="flex items-center gap-3 px-5 py-3.5">
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
                  className="shrink-0 text-xs text-gray-400 hover:text-gray-100"
                >
                  Edit
                </button>
                <button
                  onClick={() => setDeleteConfirm(log.id)}
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
              Delete this sleep entry? This cannot be undone.
            </p>
            <div className="flex gap-2">
              <button
                onClick={() => handleDelete(deleteConfirm)}
                disabled={deleteSleepLog.isPending}
                className="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-500 disabled:opacity-50"
              >
                {deleteSleepLog.isPending ? 'Deleting…' : 'Delete'}
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
