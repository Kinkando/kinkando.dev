import { useState } from 'react'
import type {
  WorkoutPreset,
  WorkoutType,
  ExerciseSection,
  WorkoutPresetExerciseInput,
  CreatePresetInput,
} from '../../lib/api/types'
import {
  PRESET_TYPES,
  WORKOUT_TYPE_LABELS,
  SECTIONS,
  SECTION_LABELS,
} from '../../lib/workout'

type Props = {
  preset?: WorkoutPreset
  onSave: (input: CreatePresetInput) => Promise<void>
  onCancel: () => void
  saving: boolean
}

const inputClass =
  'w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none'
const labelClass = 'mb-1 block text-xs font-medium text-gray-400'
const smallInputClass =
  'w-full rounded-md border border-gray-700 bg-gray-800 px-2 py-1.5 text-xs text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none'

type ExerciseRow = {
  section: ExerciseSection
  name: string
  target_muscles: string
  instructions: string
  sets: string
  reps: string
  duration_seconds: string
  rest_seconds: string
  weight_kg: string
  equipment: string
  notes: string
}

const defaultExerciseRow = (): ExerciseRow => ({
  section: 'main',
  name: '',
  target_muscles: '',
  instructions: '',
  sets: '',
  reps: '',
  duration_seconds: '',
  rest_seconds: '',
  weight_kg: '',
  equipment: '',
  notes: '',
})

function presetToForm(preset: WorkoutPreset): {
  name: string
  type: WorkoutType
  description: string
  exercises: ExerciseRow[]
} {
  return {
    name: preset.name,
    type: preset.type,
    description: preset.description ?? '',
    exercises: preset.exercises.map((ex) => ({
      section: ex.section,
      name: ex.name,
      target_muscles: ex.target_muscles ?? '',
      instructions: ex.instructions ?? '',
      sets: ex.sets != null ? String(ex.sets) : '',
      reps: ex.reps != null ? String(ex.reps) : '',
      duration_seconds:
        ex.duration_seconds != null ? String(ex.duration_seconds) : '',
      rest_seconds: ex.rest_seconds != null ? String(ex.rest_seconds) : '',
      weight_kg: ex.weight_kg != null ? String(ex.weight_kg) : '',
      equipment: ex.equipment ?? '',
      notes: ex.notes ?? '',
    })),
  }
}

export default function PresetForm({
  preset,
  onSave,
  onCancel,
  saving,
}: Props) {
  const initial = preset
    ? presetToForm(preset)
    : {
        name: '',
        type: 'weight_training' as WorkoutType,
        description: '',
        exercises: [defaultExerciseRow()],
      }

  const [name, setName] = useState(initial.name)
  const [type, setType] = useState<WorkoutType>(initial.type)
  const [description, setDescription] = useState(initial.description)
  const [exercises, setExercises] = useState<ExerciseRow[]>(
    initial.exercises.length > 0 ? initial.exercises : [defaultExerciseRow()],
  )
  const [error, setError] = useState('')

  function updateExercise(index: number, patch: Partial<ExerciseRow>) {
    setExercises((rows) =>
      rows.map((r, i) => (i === index ? { ...r, ...patch } : r)),
    )
  }

  function addExercise() {
    setExercises((rows) => [...rows, defaultExerciseRow()])
  }

  function removeExercise(index: number) {
    setExercises((rows) => rows.filter((_, i) => i !== index))
  }

  function moveExercise(index: number, direction: -1 | 1) {
    const next = index + direction
    if (next < 0 || next >= exercises.length) return
    setExercises((rows) => {
      const updated = [...rows]
      ;[updated[index], updated[next]] = [updated[next], updated[index]]
      return updated
    })
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')

    if (!name.trim()) {
      setError('Preset name is required.')
      return
    }

    for (const ex of exercises) {
      if (!ex.name.trim()) {
        setError('Each exercise must have a name.')
        return
      }
    }

    const payload: CreatePresetInput = {
      name: name.trim(),
      type,
      description: description.trim() || null,
      exercises: exercises.map(
        (ex): WorkoutPresetExerciseInput => ({
          section: ex.section,
          name: ex.name.trim(),
          target_muscles: ex.target_muscles.trim() || null,
          instructions: ex.instructions.trim() || null,
          sets: ex.sets ? parseInt(ex.sets, 10) : null,
          reps: ex.reps ? parseInt(ex.reps, 10) : null,
          duration_seconds: ex.duration_seconds
            ? parseInt(ex.duration_seconds, 10)
            : null,
          rest_seconds: ex.rest_seconds ? parseInt(ex.rest_seconds, 10) : null,
          weight_kg: ex.weight_kg ? parseFloat(ex.weight_kg) : null,
          equipment: ex.equipment.trim() || null,
          notes: ex.notes.trim() || null,
        }),
      ),
    }

    try {
      await onSave(payload)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong.')
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Header fields */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <div>
          <label className={labelClass}>Preset Name</label>
          <input
            className={inputClass}
            placeholder="e.g. Full Body Dumbbell"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
        </div>
        <div>
          <label className={labelClass}>Type</label>
          <select
            className={inputClass}
            value={type}
            onChange={(e) => setType(e.target.value as WorkoutType)}
          >
            {PRESET_TYPES.map((t) => (
              <option key={t} value={t}>
                {WORKOUT_TYPE_LABELS[t]}
              </option>
            ))}
          </select>
        </div>
        <div className="sm:col-span-2">
          <label className={labelClass}>Description (optional)</label>
          <input
            className={inputClass}
            placeholder="e.g. Dumbbell full-body routine, 3–4 sets per exercise"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          />
        </div>
      </div>

      {/* Exercise rows */}
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <h4 className="text-sm font-medium text-gray-300">
            Exercises ({exercises.length})
          </h4>
          <button
            type="button"
            onClick={addExercise}
            className="cursor-pointer rounded-md bg-gray-800 px-3 py-1 text-xs font-medium text-indigo-400 hover:bg-gray-700"
          >
            + Add exercise
          </button>
        </div>

        {exercises.map((ex, i) => (
          <div
            key={i}
            className="rounded-lg border border-gray-700 bg-gray-800/50 p-4"
          >
            {/* Row header */}
            <div className="mb-3 flex items-center justify-between gap-2">
              <span className="text-xs font-semibold text-gray-400">
                #{i + 1}
              </span>
              <div className="flex gap-1">
                <button
                  type="button"
                  onClick={() => moveExercise(i, -1)}
                  disabled={i === 0}
                  className="cursor-pointer rounded p-0.5 text-gray-500 hover:text-gray-200 disabled:opacity-30"
                  title="Move up"
                >
                  ↑
                </button>
                <button
                  type="button"
                  onClick={() => moveExercise(i, 1)}
                  disabled={i === exercises.length - 1}
                  className="cursor-pointer rounded p-0.5 text-gray-500 hover:text-gray-200 disabled:opacity-30"
                  title="Move down"
                >
                  ↓
                </button>
                <button
                  type="button"
                  onClick={() => removeExercise(i)}
                  className="ml-1 cursor-pointer rounded px-1.5 py-0.5 text-xs text-red-500 hover:text-red-400"
                >
                  Remove
                </button>
              </div>
            </div>

            {/* Fields grid */}
            <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
              <div className="col-span-2">
                <label className={labelClass}>Name *</label>
                <input
                  className={smallInputClass}
                  placeholder="e.g. Goblet Squat"
                  value={ex.name}
                  onChange={(e) => updateExercise(i, { name: e.target.value })}
                />
              </div>
              <div>
                <label className={labelClass}>Section</label>
                <select
                  className={smallInputClass}
                  value={ex.section}
                  onChange={(e) =>
                    updateExercise(i, {
                      section: e.target.value as ExerciseSection,
                    })
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
                <label className={labelClass}>Equipment</label>
                <input
                  className={smallInputClass}
                  placeholder="e.g. Dumbbell"
                  value={ex.equipment}
                  onChange={(e) =>
                    updateExercise(i, { equipment: e.target.value })
                  }
                />
              </div>
              <div>
                <label className={labelClass}>Sets</label>
                <input
                  className={smallInputClass}
                  type="number"
                  min="1"
                  placeholder="3"
                  value={ex.sets}
                  onChange={(e) => updateExercise(i, { sets: e.target.value })}
                />
              </div>
              <div>
                <label className={labelClass}>Reps</label>
                <input
                  className={smallInputClass}
                  type="number"
                  min="1"
                  placeholder="10"
                  value={ex.reps}
                  onChange={(e) => updateExercise(i, { reps: e.target.value })}
                />
              </div>
              <div>
                <label className={labelClass}>Duration (sec)</label>
                <input
                  className={smallInputClass}
                  type="number"
                  min="1"
                  placeholder="60"
                  value={ex.duration_seconds}
                  onChange={(e) =>
                    updateExercise(i, { duration_seconds: e.target.value })
                  }
                />
              </div>
              <div>
                <label className={labelClass}>Rest (sec)</label>
                <input
                  className={smallInputClass}
                  type="number"
                  min="0"
                  placeholder="60"
                  value={ex.rest_seconds}
                  onChange={(e) =>
                    updateExercise(i, { rest_seconds: e.target.value })
                  }
                />
              </div>
              <div>
                <label className={labelClass}>Equipment Weight (kg)</label>
                <input
                  className={smallInputClass}
                  type="number"
                  min="0"
                  step="0.1"
                  placeholder="Optional"
                  value={ex.weight_kg}
                  onChange={(e) =>
                    updateExercise(i, { weight_kg: e.target.value })
                  }
                />
              </div>
              <div className="col-span-2 sm:col-span-3">
                <label className={labelClass}>Target muscles</label>
                <input
                  className={smallInputClass}
                  placeholder="e.g. Quads, glutes"
                  value={ex.target_muscles}
                  onChange={(e) =>
                    updateExercise(i, { target_muscles: e.target.value })
                  }
                />
              </div>
              <div className="col-span-2 sm:col-span-4">
                <label className={labelClass}>Instructions</label>
                <input
                  className={smallInputClass}
                  placeholder="e.g. Keep back straight, lower slowly"
                  value={ex.instructions}
                  onChange={(e) =>
                    updateExercise(i, { instructions: e.target.value })
                  }
                />
              </div>
              <div className="col-span-2 sm:col-span-4">
                <label className={labelClass}>Notes</label>
                <input
                  className={smallInputClass}
                  placeholder="Optional"
                  value={ex.notes}
                  onChange={(e) => updateExercise(i, { notes: e.target.value })}
                />
              </div>
            </div>
          </div>
        ))}

        {exercises.length === 0 && (
          <p className="rounded-lg border border-dashed border-gray-700 py-6 text-center text-sm text-gray-500">
            No exercises yet — click &ldquo;+ Add exercise&rdquo; to start.
          </p>
        )}
      </div>

      {error && <p className="text-sm text-red-400">{error}</p>}

      <div className="flex justify-end gap-2">
        <button
          type="submit"
          disabled={saving}
          className="cursor-pointer rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
        >
          {saving ? 'Saving…' : preset ? 'Update Preset' : 'Create Preset'}
        </button>
        <button
          type="button"
          onClick={onCancel}
          className="cursor-pointer rounded-lg bg-gray-800 px-4 py-2 text-sm font-medium text-gray-300 hover:bg-gray-700"
        >
          Cancel
        </button>
      </div>
    </form>
  )
}
