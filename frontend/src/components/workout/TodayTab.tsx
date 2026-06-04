import { useState, useEffect, useRef } from 'react'
import type {
  WorkoutSession,
  WorkoutScheduleEntry,
  WorkoutSessionExercise,
  ExerciseSection,
  UpdateSessionExerciseInput,
  AddSessionExerciseInput,
} from '../../lib/api/types'
import {
  useGenerateSession,
  useCreateSession,
  useUpdateSessionExercise,
  useBulkUpdateSessionExercises,
  useUpdateSession,
  useDeleteSession,
  useFinishSession,
  useAddSessionExercise,
  useDeleteSessionExercise,
} from '../../queries/useWorkout'
import { WORKOUT_TYPE_LABELS, QUICK_START_TYPES } from '../../lib/workout'

const inputClass =
  'w-full rounded-md border border-gray-700 bg-gray-800 px-2 py-1.5 text-xs text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none'

const SECTION_LABELS: Record<ExerciseSection, string> = {
  warmup: '🔥 Warm-up',
  main: '🏋️ Main',
  cooldown: '❄️ Cool-down',
}

const SECTIONS: ExerciseSection[] = ['warmup', 'main', 'cooldown']

const DAY_NAMES = [
  'Sunday',
  'Monday',
  'Tuesday',
  'Wednesday',
  'Thursday',
  'Friday',
  'Saturday',
]

function todayStr() {
  return new Date().toISOString().slice(0, 10)
}

// ── Exercise log row ───────────────────────────────────────────────────────────

type ExerciseLogState = {
  actual_sets: string
  actual_reps: string
  actual_duration_seconds: string
  weight_kg: string
  completed: boolean
  notes: string
}

function exToState(ex: WorkoutSessionExercise): ExerciseLogState {
  return {
    actual_sets:
      ex.actual_sets != null
        ? String(ex.actual_sets)
        : ex.target_sets != null
          ? String(ex.target_sets)
          : '',
    actual_reps:
      ex.actual_reps != null
        ? String(ex.actual_reps)
        : ex.target_reps != null
          ? String(ex.target_reps)
          : '',
    actual_duration_seconds:
      ex.actual_duration_seconds != null
        ? String(ex.actual_duration_seconds)
        : ex.target_duration_seconds != null
          ? String(ex.target_duration_seconds)
          : '',
    weight_kg: ex.weight_kg != null ? String(ex.weight_kg) : '',
    completed: ex.completed,
    notes: ex.notes ?? '',
  }
}

function ExerciseLogRow({
  exercise: ex,
  sessionId,
  state,
  onStateChange,
  onDelete,
  isLocked,
}: {
  exercise: WorkoutSessionExercise
  sessionId: string
  state: ExerciseLogState
  onStateChange: (s: ExerciseLogState) => void
  onDelete: (exId: string) => void
  isLocked: boolean
}) {
  const [saving, setSaving] = useState(false)
  const [saved, setSaved] = useState(false)
  const updateExercise = useUpdateSessionExercise()

  const isTimed = ex.target_duration_seconds != null && ex.target_reps == null

  async function handleSave() {
    setSaving(true)
    setSaved(false)
    const input: UpdateSessionExerciseInput = {
      actual_sets: state.actual_sets ? parseInt(state.actual_sets, 10) : null,
      actual_reps: state.actual_reps ? parseInt(state.actual_reps, 10) : null,
      actual_duration_seconds: state.actual_duration_seconds
        ? parseInt(state.actual_duration_seconds, 10)
        : null,
      weight_kg: state.weight_kg ? parseFloat(state.weight_kg) : null,
      completed: state.completed,
      notes: state.notes.trim() || null,
    }
    try {
      await updateExercise.mutateAsync({ sessionId, exId: ex.id, input })
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
    } catch {
      // ignore — user can retry
    } finally {
      setSaving(false)
    }
  }

  return (
    <div
      className={`rounded-lg border p-3 ${
        state.completed
          ? 'border-green-800/60 bg-green-900/10'
          : 'border-gray-700 bg-gray-800/30'
      }`}
    >
      <div className="mb-2">
        <div className="flex items-center justify-between gap-2">
          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              checked={state.completed}
              disabled={isLocked}
              onChange={(e) =>
                !isLocked &&
                onStateChange({ ...state, completed: e.target.checked })
              }
              className="h-4 w-4 accent-indigo-500 disabled:opacity-60"
            />
            <span
              className={`text-sm font-medium ${state.completed ? 'text-gray-500 line-through' : 'text-gray-100'}`}
            >
              {ex.name}
            </span>
          </div>
          {!isLocked && (
            <button
              onClick={() => onDelete(ex.id)}
              className="shrink-0 cursor-pointer text-xs text-gray-600 hover:text-red-400"
              title="Remove exercise"
            >
              ✕
            </button>
          )}
        </div>
        {ex.target_muscles && (
          <p className="mt-0.5 pl-6 text-xs text-gray-500">
            {ex.target_muscles}
          </p>
        )}
        {ex.instructions && (
          <p className="mt-0.5 pl-6 text-xs text-gray-600 italic">
            {ex.instructions}
          </p>
        )}
        <div className="mt-1 flex flex-wrap gap-2 pl-6 text-xs text-gray-600">
          {ex.target_sets && <span>Target: {ex.target_sets} sets</span>}
          {ex.target_reps && <span>× {ex.target_reps} reps</span>}
          {ex.target_duration_seconds && (
            <span>× {ex.target_duration_seconds}s</span>
          )}
          {ex.rest_seconds && <span>Rest: {ex.rest_seconds}s</span>}
        </div>
      </div>

      <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
        {!isTimed ? (
          <>
            <div>
              <label className="mb-1 block text-xs text-gray-500">
                Actual sets
              </label>
              <input
                className={inputClass}
                type="number"
                min="0"
                placeholder={ex.target_sets ? String(ex.target_sets) : '—'}
                value={state.actual_sets}
                disabled={isLocked}
                onChange={(e) =>
                  onStateChange({ ...state, actual_sets: e.target.value })
                }
              />
            </div>
            <div>
              <label className="mb-1 block text-xs text-gray-500">
                Actual reps
              </label>
              <input
                className={inputClass}
                type="number"
                min="0"
                placeholder={ex.target_reps ? String(ex.target_reps) : '—'}
                value={state.actual_reps}
                disabled={isLocked}
                onChange={(e) =>
                  onStateChange({ ...state, actual_reps: e.target.value })
                }
              />
            </div>
          </>
        ) : (
          <div className="col-span-2">
            <label className="mb-1 block text-xs text-gray-500">
              Actual duration (sec)
            </label>
            <input
              className={inputClass}
              type="number"
              min="0"
              placeholder={
                ex.target_duration_seconds
                  ? String(ex.target_duration_seconds)
                  : '—'
              }
              value={state.actual_duration_seconds}
              disabled={isLocked}
              onChange={(e) =>
                onStateChange({
                  ...state,
                  actual_duration_seconds: e.target.value,
                })
              }
            />
          </div>
        )}
        <div>
          <label className="mb-1 block text-xs text-gray-500">
            Equipment Weight (kg)
          </label>
          <input
            className={inputClass}
            type="number"
            min="0"
            step="0.05"
            placeholder="—"
            value={state.weight_kg}
            disabled={isLocked}
            onChange={(e) =>
              onStateChange({ ...state, weight_kg: e.target.value })
            }
          />
        </div>
        <div className="col-span-2 sm:col-span-4">
          <label className="mb-1 block text-xs text-gray-500">Notes</label>
          <input
            className={inputClass}
            placeholder="Optional"
            value={state.notes}
            disabled={isLocked}
            onChange={(e) => onStateChange({ ...state, notes: e.target.value })}
          />
        </div>
      </div>

      {!isLocked && (
        <div className="mt-2 flex items-center gap-2">
          <button
            onClick={handleSave}
            disabled={saving}
            className="cursor-pointer rounded-md bg-gray-700 px-3 py-1 text-xs font-medium text-gray-200 hover:bg-gray-600 disabled:opacity-50"
          >
            {saving ? 'Saving…' : 'Save'}
          </button>
          {saved && <span className="text-xs text-green-400">✓</span>}
        </div>
      )}
    </div>
  )
}

// ── Add-exercise form ──────────────────────────────────────────────────────────

type AddExerciseFormState = {
  section: ExerciseSection
  name: string
  target_sets: string
  target_reps: string
  target_duration_seconds: string
  rest_seconds: string
}

const defaultAddForm = (): AddExerciseFormState => ({
  section: 'main',
  name: '',
  target_sets: '',
  target_reps: '',
  target_duration_seconds: '',
  rest_seconds: '',
})

function AddExerciseForm({
  sessionId,
  onAdded,
  isLocked,
}: {
  sessionId: string
  onAdded: () => void
  isLocked: boolean
}) {
  const [open, setOpen] = useState(false)
  const [form, setForm] = useState<AddExerciseFormState>(defaultAddForm())
  const [error, setError] = useState('')
  const addExercise = useAddSessionExercise()

  function reset() {
    setForm(defaultAddForm())
    setError('')
    setOpen(false)
  }

  async function handleAdd() {
    setError('')
    if (!form.name.trim()) {
      setError('Exercise name is required.')
      return
    }
    const input: AddSessionExerciseInput = {
      section: form.section,
      name: form.name.trim(),
      target_muscles: null,
      instructions: null,
      target_sets: form.target_sets ? parseInt(form.target_sets, 10) : null,
      target_reps: form.target_reps ? parseInt(form.target_reps, 10) : null,
      target_duration_seconds: form.target_duration_seconds
        ? parseInt(form.target_duration_seconds, 10)
        : null,
      rest_seconds: form.rest_seconds ? parseInt(form.rest_seconds, 10) : null,
    }
    try {
      await addExercise.mutateAsync({ sessionId, input })
      reset()
      onAdded()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to add exercise.')
    }
  }

  if (isLocked) return null

  if (!open) {
    return (
      <button
        onClick={() => setOpen(true)}
        className="mt-2 w-full cursor-pointer rounded-lg border border-dashed border-gray-700 py-2 text-xs font-medium text-gray-500 hover:border-gray-500 hover:text-gray-300"
      >
        + Add exercise
      </button>
    )
  }

  return (
    <div className="mt-2 rounded-lg border border-gray-700 bg-gray-800/50 p-3">
      <p className="mb-3 text-xs font-medium text-gray-300">New exercise</p>
      <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
        <div className="col-span-2">
          <label className="mb-1 block text-xs text-gray-500">Name *</label>
          <input
            className={inputClass}
            placeholder="e.g. Push-up"
            value={form.name}
            onChange={(e) => setForm({ ...form, name: e.target.value })}
          />
        </div>
        <div>
          <label className="mb-1 block text-xs text-gray-500">Section</label>
          <select
            className={inputClass}
            value={form.section}
            onChange={(e) =>
              setForm({ ...form, section: e.target.value as ExerciseSection })
            }
          >
            {SECTIONS.map((s) => (
              <option key={s} value={s}>
                {SECTION_LABELS[s]}
              </option>
            ))}
          </select>
        </div>
        <div>
          <label className="mb-1 block text-xs text-gray-500">Sets</label>
          <input
            className={inputClass}
            type="number"
            min="1"
            placeholder="3"
            value={form.target_sets}
            onChange={(e) => setForm({ ...form, target_sets: e.target.value })}
          />
        </div>
        <div>
          <label className="mb-1 block text-xs text-gray-500">Reps</label>
          <input
            className={inputClass}
            type="number"
            min="1"
            placeholder="10"
            value={form.target_reps}
            onChange={(e) => setForm({ ...form, target_reps: e.target.value })}
          />
        </div>
        <div>
          <label className="mb-1 block text-xs text-gray-500">
            Duration (sec)
          </label>
          <input
            className={inputClass}
            type="number"
            min="1"
            placeholder="60"
            value={form.target_duration_seconds}
            onChange={(e) =>
              setForm({ ...form, target_duration_seconds: e.target.value })
            }
          />
        </div>
        <div>
          <label className="mb-1 block text-xs text-gray-500">Rest (sec)</label>
          <input
            className={inputClass}
            type="number"
            min="0"
            placeholder="60"
            value={form.rest_seconds}
            onChange={(e) => setForm({ ...form, rest_seconds: e.target.value })}
          />
        </div>
      </div>
      {error && <p className="mt-2 text-xs text-red-400">{error}</p>}
      <div className="mt-3 flex gap-2">
        <button
          onClick={handleAdd}
          disabled={addExercise.isPending}
          className="cursor-pointer rounded-lg bg-indigo-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
        >
          {addExercise.isPending ? 'Adding…' : 'Add'}
        </button>
        <button
          onClick={reset}
          className="cursor-pointer rounded-lg bg-gray-700 px-3 py-1.5 text-xs font-medium text-gray-300 hover:bg-gray-600"
        >
          Cancel
        </button>
      </div>
    </div>
  )
}

// ── Session view ──────────────────────────────────────────────────────────────

function SessionView({
  session,
  onDelete,
  onExerciseChange,
  defaultExpanded = true,
}: {
  session: WorkoutSession
  onDelete: () => void
  onExerciseChange: () => void
  defaultExpanded?: boolean
}) {
  const isLocked = session.completed_at != null
  const [bodyExpanded, setBodyExpanded] = useState(defaultExpanded)

  // Session fields — edited together, saved with one button
  const [nameValue, setNameValue] = useState(session.name)
  const [duration, setDuration] = useState(
    session.duration_minutes != null ? String(session.duration_minutes) : '',
  )
  const [notes, setNotes] = useState(session.notes ?? '')
  const [saving, setSaving] = useState(false)
  const [durationError, setDurationError] = useState('')

  // Sync fields when server data changes (after a save or external update)
  useEffect(() => {
    setNameValue(session.name)
    setDuration(
      session.duration_minutes != null ? String(session.duration_minutes) : '',
    )
    setNotes(session.notes ?? '')
  }, [session.name, session.duration_minutes, session.notes])

  // Dialogs
  const [deleteExConfirm, setDeleteExConfirm] = useState<string | null>(null)
  const [finishConfirm, setFinishConfirm] = useState(false)

  const updateSession = useUpdateSession()
  const deleteExercise = useDeleteSessionExercise()
  const bulkUpdate = useBulkUpdateSessionExercises()
  const finishSessionMutation = useFinishSession()

  const [exerciseStates, setExerciseStates] = useState<
    Map<string, ExerciseLogState>
  >(() => new Map(session.exercises.map((ex) => [ex.id, exToState(ex)])))

  useEffect(() => {
    setExerciseStates((prev) => {
      const currentIds = new Set(session.exercises.map((e) => e.id))
      const prevIds = new Set(prev.keys())
      const added = session.exercises.filter((e) => !prevIds.has(e.id))
      const removed = [...prevIds].filter((id) => !currentIds.has(id))
      if (added.length === 0 && removed.length === 0) return prev
      const next = new Map(prev)
      for (const ex of added) next.set(ex.id, exToState(ex))
      for (const id of removed) next.delete(id)
      return next
    })
  }, [session.exercises])

  const [savingAll, setSavingAll] = useState(false)
  const [savedAll, setSavedAll] = useState(false)

  function handleCheckAll() {
    setExerciseStates((prev) => {
      const next = new Map(prev)
      for (const [id, s] of next) next.set(id, { ...s, completed: true })
      return next
    })
  }

  async function handleSaveAll() {
    setSavingAll(true)
    setSavedAll(false)
    try {
      await bulkUpdate.mutateAsync({
        sessionId: session.id,
        input: {
          items: session.exercises.map((ex) => {
            const s = exerciseStates.get(ex.id) ?? exToState(ex)
            return {
              id: ex.id,
              actual_sets: s.actual_sets ? parseInt(s.actual_sets, 10) : null,
              actual_reps: s.actual_reps ? parseInt(s.actual_reps, 10) : null,
              actual_duration_seconds: s.actual_duration_seconds
                ? parseInt(s.actual_duration_seconds, 10)
                : null,
              weight_kg: s.weight_kg ? parseFloat(s.weight_kg) : null,
              completed: s.completed,
              notes: s.notes.trim() || null,
            }
          }),
        },
      })
      setSavedAll(true)
      setTimeout(() => setSavedAll(false), 2000)
    } catch {
      // ignore
    } finally {
      setSavingAll(false)
    }
  }

  async function handleSaveFields() {
    setSaving(true)
    try {
      await updateSession.mutateAsync({
        id: session.id,
        input: {
          name: nameValue.trim() || session.name,
          duration_minutes: duration ? parseInt(duration, 10) : null,
          notes: notes.trim() || null,
        },
      })
      setDurationError('')
    } finally {
      setSaving(false)
    }
  }

  function handleResetFields() {
    setNameValue(session.name)
    setDuration(
      session.duration_minutes != null ? String(session.duration_minutes) : '',
    )
    setNotes(session.notes ?? '')
    setDurationError('')
  }

  async function handleFinishedClick() {
    if (!duration.trim() && session.duration_minutes == null) {
      setDurationError('Duration is required before finishing.')
      setBodyExpanded(true)
      return
    }
    setDurationError('')
    // Auto-save any unsaved field changes before opening the confirm dialog
    const localDuration = duration.trim() ? parseInt(duration, 10) : null
    const hasUnsaved =
      (nameValue.trim() || session.name) !== session.name ||
      localDuration !== session.duration_minutes ||
      (notes.trim() || null) !== session.notes
    if (hasUnsaved) {
      setSaving(true)
      try {
        await updateSession.mutateAsync({
          id: session.id,
          input: {
            name: nameValue.trim() || session.name,
            duration_minutes: localDuration,
            notes: notes.trim() || null,
          },
        })
      } catch {
        setSaving(false)
        return
      }
      setSaving(false)
    }
    setFinishConfirm(true)
  }

  async function handleFinish() {
    await finishSessionMutation.mutateAsync(session.id)
    setFinishConfirm(false)
  }

  async function handleDeleteExercise(exId: string) {
    await deleteExercise.mutateAsync({ sessionId: session.id, exId })
    setDeleteExConfirm(null)
    onExerciseChange()
  }

  const completedCount = session.exercises.filter((e) => e.completed).length

  const bySection: Partial<Record<ExerciseSection, WorkoutSessionExercise[]>> =
    {}
  for (const ex of session.exercises) {
    ;(bySection[ex.section] ??= []).push(ex)
  }

  return (
    <div className="rounded-xl border border-gray-800 bg-gray-900">
      {/* ── Collapsible header ─────────────────────────────────────── */}
      <div className="p-4">
        <div className="flex items-start gap-3">
          {/* Left: badges + name + meta */}
          <div
            className="min-w-0 flex-1 cursor-pointer"
            onClick={() => setBodyExpanded((b) => !b)}
          >
            <div className="flex flex-wrap items-center gap-2">
              <span className="rounded bg-indigo-900/50 px-2 py-0.5 text-xs text-indigo-400">
                {WORKOUT_TYPE_LABELS[session.type]}
              </span>
              {isLocked && (
                <span className="rounded bg-green-900/50 px-2 py-0.5 text-xs text-green-400">
                  Finished
                </span>
              )}
              <span className="truncate text-sm font-medium text-gray-100">
                {session.name}
              </span>
            </div>
            <div className="mt-0.5 flex flex-wrap gap-2 text-xs text-gray-500">
              <span>
                {new Date(session.created_at).toLocaleTimeString(undefined, {
                  hour: '2-digit',
                  minute: '2-digit',
                })}
              </span>
              {session.duration_minutes != null && (
                <span>⏱ {session.duration_minutes} min</span>
              )}
              {session.exercises.length > 0 && (
                <span>
                  {completedCount}/{session.exercises.length} done
                </span>
              )}
              {isLocked && session.completed_at && (
                <span className="text-green-600">
                  ✓{' '}
                  {new Date(session.completed_at).toLocaleTimeString(
                    undefined,
                    {
                      hour: '2-digit',
                      minute: '2-digit',
                    },
                  )}
                </span>
              )}
            </div>
          </div>

          {/* Right: actions + chevron */}
          <div className="flex shrink-0 items-center gap-2">
            {!isLocked && (
              <>
                <button
                  onClick={handleFinishedClick}
                  className="cursor-pointer text-xs font-medium text-green-500 hover:text-green-400"
                >
                  Finished
                </button>
                <button
                  onClick={onDelete}
                  className="cursor-pointer text-xs text-red-500 hover:text-red-400"
                >
                  Delete
                </button>
              </>
            )}
            <button
              onClick={() => setBodyExpanded((b) => !b)}
              className="cursor-pointer text-gray-600 hover:text-gray-400"
              aria-label={bodyExpanded ? 'Collapse' : 'Expand'}
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className={`h-4 w-4 transition-transform ${bodyExpanded ? 'rotate-180' : ''}`}
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M19 9l-7 7-7-7"
                />
              </svg>
            </button>
          </div>
        </div>

        {durationError && (
          <p className="mt-2 text-xs text-red-400">{durationError}</p>
        )}
      </div>

      {/* ── Expandable body ────────────────────────────────────────── */}
      {bodyExpanded && (
        <div className="space-y-4 border-t border-gray-800 px-4 pt-4 pb-4">
          {/* Session fields */}
          <div className="space-y-3 rounded-lg border border-gray-700 bg-gray-800/30 p-3">
            <div>
              <label className="mb-1 block text-xs text-gray-500">Name</label>
              <input
                className={inputClass}
                value={nameValue}
                disabled={isLocked}
                onChange={(e) => setNameValue(e.target.value)}
              />
            </div>

            <div>
              <label className="mb-1 block text-xs text-gray-500">
                Duration (min)
                {!isLocked && <span className="text-red-400"> *</span>}
              </label>
              {isLocked ? (
                <p className="text-xs text-gray-400">
                  {session.duration_minutes != null
                    ? `${session.duration_minutes} min`
                    : '—'}
                </p>
              ) : (
                <input
                  className={`w-28 ${inputClass}`}
                  type="number"
                  min="1"
                  placeholder="e.g. 45"
                  value={duration}
                  onChange={(e) => setDuration(e.target.value)}
                />
              )}
            </div>

            <div>
              <label className="mb-1 block text-xs text-gray-500">Notes</label>
              {isLocked ? (
                <p className="text-xs text-gray-400">{notes || '—'}</p>
              ) : (
                <textarea
                  className="w-full resize-none rounded-md border border-gray-700 bg-gray-800 px-2 py-1.5 text-xs text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none"
                  rows={2}
                  placeholder="Optional"
                  value={notes}
                  onChange={(e) => setNotes(e.target.value)}
                />
              )}
            </div>

            {!isLocked && (
              <div className="flex items-center gap-2 pt-1">
                <button
                  onClick={handleSaveFields}
                  disabled={saving}
                  className="cursor-pointer rounded-md bg-indigo-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
                >
                  {saving ? 'Saving…' : 'Save'}
                </button>
                <button
                  onClick={handleResetFields}
                  disabled={saving}
                  className="cursor-pointer rounded-md bg-gray-700 px-3 py-1.5 text-xs font-medium text-gray-300 hover:bg-gray-600 disabled:opacity-50"
                >
                  Reset
                </button>
              </div>
            )}
          </div>

          {/* Bulk actions */}
          {!isLocked && session.exercises.length > 0 && (
            <div className="flex items-center gap-2">
              <button
                onClick={handleCheckAll}
                className="cursor-pointer rounded-md bg-green-900/30 px-3 py-1.5 text-xs font-medium text-green-400 hover:bg-green-900/50"
              >
                Check All
              </button>
              <button
                onClick={handleSaveAll}
                disabled={savingAll}
                className="cursor-pointer rounded-md bg-indigo-900/30 px-3 py-1.5 text-xs font-medium text-indigo-400 hover:bg-indigo-900/50 disabled:opacity-50"
              >
                {savingAll ? 'Saving…' : 'Save All'}
              </button>
              {savedAll && (
                <span className="text-xs text-green-400">✓ All saved</span>
              )}
            </div>
          )}

          {/* Exercises by section */}
          {SECTIONS.map((section) => {
            const exs = bySection[section]
            if (!exs?.length) return null
            return (
              <div key={section} className="space-y-2">
                <h4 className="text-xs font-semibold tracking-wide text-gray-500 uppercase">
                  {SECTION_LABELS[section]}
                </h4>
                {exs.map((ex) => (
                  <ExerciseLogRow
                    key={ex.id}
                    exercise={ex}
                    sessionId={session.id}
                    state={exerciseStates.get(ex.id) ?? exToState(ex)}
                    onStateChange={(s) =>
                      setExerciseStates((prev) => new Map(prev).set(ex.id, s))
                    }
                    onDelete={(exId) => setDeleteExConfirm(exId)}
                    isLocked={isLocked}
                  />
                ))}
              </div>
            )
          })}

          {/* Add exercise */}
          <AddExerciseForm
            sessionId={session.id}
            onAdded={onExerciseChange}
            isLocked={isLocked}
          />
        </div>
      )}

      {/* ── Modals ─────────────────────────────────────────────────── */}

      {/* Delete exercise confirm */}
      {deleteExConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div className="w-full max-w-sm rounded-xl border border-gray-700 bg-gray-900 p-6 shadow-2xl">
            <p className="mb-4 text-sm text-gray-300">
              Remove this exercise from the session?
            </p>
            <div className="flex justify-end gap-2">
              <button
                onClick={() => handleDeleteExercise(deleteExConfirm)}
                disabled={deleteExercise.isPending}
                className="cursor-pointer rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-500 disabled:opacity-50"
              >
                {deleteExercise.isPending ? 'Removing…' : 'Remove'}
              </button>
              <button
                onClick={() => setDeleteExConfirm(null)}
                className="cursor-pointer rounded-lg bg-gray-800 px-4 py-2 text-sm font-medium text-gray-300 hover:bg-gray-700"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Finish confirm */}
      {finishConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div className="w-full max-w-sm rounded-xl border border-gray-700 bg-gray-900 p-6 shadow-2xl">
            <p className="mb-1 text-sm font-medium text-gray-100">
              Mark workout as finished?
            </p>
            <p className="mb-4 text-xs text-gray-500">
              You won&apos;t be able to edit or delete this session afterwards.
            </p>
            <div className="flex justify-end gap-2">
              <button
                onClick={handleFinish}
                disabled={finishSessionMutation.isPending}
                className="cursor-pointer rounded-lg bg-green-600 px-4 py-2 text-sm font-medium text-white hover:bg-green-500 disabled:opacity-50"
              >
                {finishSessionMutation.isPending ? 'Finishing…' : 'Finished'}
              </button>
              <button
                onClick={() => setFinishConfirm(false)}
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

// ── TodayTab ───────────────────────────────────────────────────────────────────

type Props = {
  todaySessions: WorkoutSession[] | undefined
  schedule: WorkoutScheduleEntry[] | undefined
  onSessionChange: () => Promise<void>
}

export default function TodayTab({
  todaySessions,
  schedule,
  onSessionChange,
}: Props) {
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null)
  // Track which sessions have been auto-expanded; default-expand only the first one.
  const initializedRef = useRef(false)
  const [expandedByDefault, setExpandedByDefault] = useState<
    Record<string, boolean>
  >({})
  const [isDone, setIsDone] = useState(true)

  useEffect(() => {
    if (todaySessions && todaySessions.length > 0 && !initializedRef.current) {
      initializedRef.current = true
      const map: Record<string, boolean> = {}
      todaySessions.forEach((s, i) => {
        map[s.id] = i === 0
      })
      setExpandedByDefault(map)
    }
  }, [todaySessions])

  const generateSession = useGenerateSession()
  const createSession = useCreateSession()
  const deleteSession = useDeleteSession()

  const dayOfWeek = new Date().getDay()
  const todaySchedule = (schedule ?? []).find(
    (e) => e.day_of_week === dayOfWeek,
  )
  const generating =
    generateSession.isPending || createSession.isPending || !isDone

  function setStateDone(func: () => Promise<void>) {
    return async () => {
      setIsDone(false)
      try {
        await func()
      } finally {
        setIsDone(true)
      }
    }
  }

  async function handleGenerate() {
    await generateSession.mutateAsync(todayStr())
    await onSessionChange()
  }

  async function handleQuickStart(type: string, label: string) {
    await createSession.mutateAsync({ type, name: label, date: todayStr() })
    await onSessionChange()
  }

  async function handleDeleteSession(id: string) {
    await deleteSession.mutateAsync(id)
    setDeleteConfirm(null)
    await onSessionChange()
  }

  if (todaySessions && todaySessions.length > 0) {
    return (
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-medium text-gray-300">
            Today&apos;s Workout
          </h3>
          {todaySessions.length > 1 && (
            <span className="text-xs text-gray-500">
              {todaySessions.length} sessions today
            </span>
          )}
        </div>

        {todaySessions.map((session, i) => (
          <SessionView
            key={session.id}
            session={session}
            defaultExpanded={expandedByDefault[session.id] ?? i === 0}
            onDelete={() => setDeleteConfirm(session.id)}
            onExerciseChange={onSessionChange}
          />
        ))}

        {deleteConfirm && (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
            <div className="w-full max-w-sm rounded-xl border border-gray-700 bg-gray-900 p-6 shadow-2xl">
              <p className="mb-4 text-sm text-gray-300">
                Delete this session? All logged data will be lost.
              </p>
              <div className="flex justify-end gap-2">
                <button
                  onClick={setStateDone(() =>
                    handleDeleteSession(deleteConfirm),
                  )}
                  disabled={deleteSession.isPending}
                  className="cursor-pointer rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-500 disabled:opacity-50"
                >
                  {deleteSession.isPending ? 'Deleting…' : 'Delete'}
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

  // No session yet for today.
  return (
    <div className="space-y-4">
      {todaySchedule ? (
        <div className="rounded-xl border border-indigo-800/40 bg-indigo-900/10 p-5">
          <p className="mb-1 text-xs tracking-wide text-indigo-400 uppercase">
            Scheduled for {DAY_NAMES[dayOfWeek]}
          </p>
          <p className="text-sm font-medium text-gray-100">
            {todaySchedule.preset_name}
          </p>
          <p className="mt-0.5 text-xs text-gray-500">
            {WORKOUT_TYPE_LABELS[todaySchedule.preset_type]}
          </p>
          <button
            onClick={setStateDone(handleGenerate)}
            disabled={generating}
            className="mt-4 cursor-pointer rounded-lg bg-indigo-600 px-5 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
          >
            {generating ? 'Waiting…' : "Start today's workout →"}
          </button>
          {generateSession.isError && (
            <p className="mt-2 text-xs text-red-400">
              {generateSession.error instanceof Error
                ? generateSession.error.message
                : 'Failed to generate session.'}
            </p>
          )}
        </div>
      ) : (
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
          <p className="text-sm font-medium text-gray-300">Rest day 🎉</p>
          <p className="mt-1 text-xs text-gray-500">
            No workout scheduled for {DAY_NAMES[dayOfWeek]}. Start one below.
          </p>
        </div>
      )}

      {/* Quick Start */}
      <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
        <p className="mb-3 text-sm font-medium text-gray-300">Quick Start</p>
        <div className="grid grid-cols-2 gap-2 sm:grid-cols-3">
          {QUICK_START_TYPES.map(({ type, label }) => (
            <button
              key={type}
              onClick={setStateDone(() => handleQuickStart(type, label))}
              disabled={generating}
              className="cursor-pointer rounded-lg border border-gray-700 bg-gray-800 px-3 py-2.5 text-left text-sm font-medium text-gray-200 hover:border-gray-600 hover:bg-gray-700 disabled:opacity-50"
            >
              {label}
            </button>
          ))}
        </div>
        {createSession.isError && (
          <p className="mt-2 text-xs text-red-400">
            {createSession.error instanceof Error
              ? createSession.error.message
              : 'Failed to start session.'}
          </p>
        )}
      </div>
    </div>
  )
}
