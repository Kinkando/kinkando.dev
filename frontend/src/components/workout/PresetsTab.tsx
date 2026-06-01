import { useState } from 'react'
import type { WorkoutPreset, CreatePresetInput } from '../../lib/api/types'
import {
  useCreatePreset,
  useUpdatePreset,
  useDeletePreset,
} from '../../queries/useWorkout'
import PresetForm from './PresetForm'

const TYPE_LABELS = {
  weight_training: 'Weight Training',
  body_weight: 'Body Weight',
} as const

type Props = {
  presets: WorkoutPreset[] | undefined
}

export default function PresetsTab({ presets }: Props) {
  const [showForm, setShowForm] = useState(false)
  const [editingPreset, setEditingPreset] = useState<WorkoutPreset | null>(null)
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null)

  const createPreset = useCreatePreset()
  const updatePreset = useUpdatePreset()
  const deletePreset = useDeletePreset()

  const saving = createPreset.isPending || updatePreset.isPending

  function handleNew() {
    setEditingPreset(null)
    setShowForm(true)
  }

  function handleEdit(preset: WorkoutPreset) {
    setEditingPreset(preset)
    setShowForm(true)
  }

  function handleCancelForm() {
    setShowForm(false)
    setEditingPreset(null)
  }

  async function handleSave(input: CreatePresetInput) {
    if (editingPreset) {
      await updatePreset.mutateAsync({ id: editingPreset.id, input })
    } else {
      await createPreset.mutateAsync(input)
    }
    setShowForm(false)
    setEditingPreset(null)
  }

  async function handleDelete(id: string) {
    await deletePreset.mutateAsync(id)
    setDeleteConfirm(null)
  }

  if (showForm) {
    return (
      <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
        <h3 className="mb-4 text-sm font-medium text-gray-300">
          {editingPreset ? 'Edit Preset' : 'New Preset'}
        </h3>
        <PresetForm
          preset={editingPreset ?? undefined}
          onSave={handleSave}
          onCancel={handleCancelForm}
          saving={saving}
        />
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <p className="text-sm text-gray-400">
          {presets?.length
            ? `${presets.length} preset${presets.length > 1 ? 's' : ''}`
            : 'No presets yet'}
        </p>
        <button
          onClick={handleNew}
          className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500"
        >
          New Preset
        </button>
      </div>

      {!presets || presets.length === 0 ? (
        <div className="rounded-xl border border-dashed border-gray-700 py-12 text-center">
          <p className="mb-2 text-sm text-gray-500">No presets yet.</p>
          <p className="text-xs text-gray-600">
            Create a preset to save your workout routines.
          </p>
        </div>
      ) : (
        <div className="space-y-3">
          {presets.map((preset) => (
            <div
              key={preset.id}
              className="rounded-xl border border-gray-800 bg-gray-900 p-4"
            >
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <span className="rounded bg-indigo-900/50 px-2 py-0.5 text-xs text-indigo-400">
                      {TYPE_LABELS[preset.type]}
                    </span>
                    <h3 className="truncate text-sm font-medium text-gray-100">
                      {preset.name}
                    </h3>
                  </div>
                  {preset.description && (
                    <p className="mt-1 text-xs text-gray-500">
                      {preset.description}
                    </p>
                  )}
                  <p className="mt-1 text-xs text-gray-600">
                    {preset.exercises.length} exercise
                    {preset.exercises.length !== 1 ? 's' : ''}
                    {preset.exercises.length > 0 && (
                      <>
                        {' '}
                        ·{' '}
                        {[
                          preset.exercises.filter((e) => e.section === 'warmup')
                            .length > 0 &&
                            `${preset.exercises.filter((e) => e.section === 'warmup').length} warm-up`,
                          preset.exercises.filter((e) => e.section === 'main')
                            .length > 0 &&
                            `${preset.exercises.filter((e) => e.section === 'main').length} main`,
                          preset.exercises.filter(
                            (e) => e.section === 'cooldown',
                          ).length > 0 &&
                            `${preset.exercises.filter((e) => e.section === 'cooldown').length} cool-down`,
                        ]
                          .filter(Boolean)
                          .join(', ')}
                      </>
                    )}
                  </p>
                </div>
                <div className="flex shrink-0 gap-2">
                  <button
                    onClick={() => handleEdit(preset)}
                    className="text-xs text-gray-400 hover:text-gray-100"
                  >
                    Edit
                  </button>
                  <button
                    onClick={() => setDeleteConfirm(preset.id)}
                    className="text-xs text-red-500 hover:text-red-400"
                  >
                    Delete
                  </button>
                </div>
              </div>

              {/* Exercise preview */}
              {preset.exercises.length > 0 && (
                <div className="mt-3 space-y-0.5">
                  {preset.exercises.slice(0, 5).map((ex, i) => (
                    <div
                      key={i}
                      className="flex items-baseline gap-2 text-xs text-gray-600"
                    >
                      <span className="shrink-0 text-gray-700">
                        {ex.section === 'warmup'
                          ? '🔥'
                          : ex.section === 'cooldown'
                            ? '❄️'
                            : '·'}
                      </span>
                      <span className="text-gray-500">{ex.name}</span>
                      {(ex.sets || ex.reps) && (
                        <span>
                          {ex.sets}×{ex.reps}
                        </span>
                      )}
                      {ex.duration_seconds && !ex.reps && (
                        <span>{ex.duration_seconds}s</span>
                      )}
                    </div>
                  ))}
                  {preset.exercises.length > 5 && (
                    <p className="text-xs text-gray-700">
                      +{preset.exercises.length - 5} more…
                    </p>
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {/* Delete confirm modal */}
      {deleteConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div className="w-full max-w-sm rounded-xl border border-gray-700 bg-gray-900 p-6 shadow-2xl">
            <p className="mb-4 text-sm text-gray-300">
              Delete this preset? Any scheduled days using it will also be
              cleared. This cannot be undone.
            </p>
            <div className="flex gap-2">
              <button
                onClick={() => handleDelete(deleteConfirm)}
                disabled={deletePreset.isPending}
                className="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-500 disabled:opacity-50"
              >
                {deletePreset.isPending ? 'Deleting…' : 'Delete'}
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
