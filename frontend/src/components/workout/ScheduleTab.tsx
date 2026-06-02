import { useEffect, useState } from 'react'
import type { WorkoutPreset, WorkoutScheduleEntry } from '../../lib/api/types'
import { useSetSchedule } from '../../queries/useWorkout'

const DAYS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

const inputClass =
  'w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 focus:border-indigo-500 focus:outline-none'

type Props = {
  schedule: WorkoutScheduleEntry[] | undefined
  presets: WorkoutPreset[] | undefined
}

export default function ScheduleTab({ schedule, presets }: Props) {
  // Local state: index 0=Sun … 6=Sat, value = preset id or ''
  const [entries, setEntries] = useState<string[]>(Array(7).fill(''))
  const [saved, setSaved] = useState(false)
  const [error, setError] = useState('')

  const setSchedule = useSetSchedule()

  // Sync remote schedule into local state whenever it changes.
  useEffect(() => {
    const next = Array(7).fill('')
    for (const e of schedule ?? []) {
      if (e.day_of_week >= 0 && e.day_of_week <= 6) {
        next[e.day_of_week] = e.preset_id
      }
    }
    setEntries(next)
  }, [schedule])

  async function handleSave() {
    setError('')
    setSaved(false)
    const payload = entries
      .map((presetId, dayOfWeek) => ({
        day_of_week: dayOfWeek,
        preset_id: presetId,
      }))
      .filter((e) => e.preset_id !== '')

    try {
      await setSchedule.mutateAsync({ entries: payload })
      setSaved(true)
      setTimeout(() => setSaved(false), 2500)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save schedule.')
    }
  }

  return (
    <div className="space-y-4">
      <p className="text-sm text-gray-400">
        Assign a preset to each weekday. Days without a preset are rest days.
      </p>

      <div className="rounded-xl border border-gray-800 bg-gray-900">
        <div className="divide-y divide-gray-800">
          {DAYS.map((day, i) => (
            <div key={i} className="flex items-center gap-4 px-5 py-3">
              <span className="w-10 shrink-0 text-sm font-medium text-gray-400">
                {day}
              </span>
              <select
                className={inputClass}
                value={entries[i]}
                onChange={(e) => {
                  const next = [...entries]
                  next[i] = e.target.value
                  setEntries(next)
                }}
              >
                <option value="">— Rest day —</option>
                {(presets ?? []).map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.name}
                  </option>
                ))}
              </select>
            </div>
          ))}
        </div>
      </div>

      {error && <p className="text-sm text-red-400">{error}</p>}

      <div className="flex items-center gap-3">
        <button
          onClick={handleSave}
          disabled={setSchedule.isPending}
          className="cursor-pointer rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
        >
          {setSchedule.isPending ? 'Saving…' : 'Save Schedule'}
        </button>
        {saved && <span className="text-sm text-green-400">Saved ✓</span>}
      </div>
    </div>
  )
}
