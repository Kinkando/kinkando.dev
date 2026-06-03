import { useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import type {
  QuestType,
  SourceType,
  DailyQuestStatus,
  CreateQuestInput,
  UpdateQuestInput,
} from '../../lib/api/types'
import { useCreateQuest, useUpdateQuest } from '../../queries/useQuest'
import { QUEST_TYPE_CONFIG, questToForm, type FormState } from './questConfig'

type Props = {
  type: QuestType
  initial?: DailyQuestStatus
  onClose: () => void
}

const inputClass =
  'w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none'

const labelClass = 'mb-1 block text-xs font-medium text-gray-400'

export default function QuestFormDialog({ type, initial, onClose }: Props) {
  const isEdit = !!initial
  const cfg = QUEST_TYPE_CONFIG[type]

  const [form, setForm] = useState<FormState>(
    initial ? questToForm(initial) : cfg.defaultForm,
  )
  const [error, setError] = useState('')

  const createQuest = useCreateQuest()
  const updateQuest = useUpdateQuest()
  const isPending = createQuest.isPending || updateQuest.isPending

  const backdropRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [onClose])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')

    if (!form.title.trim()) {
      setError('Title is required.')
      return
    }
    const xp = parseInt(form.xp_reward, 10)
    if (isNaN(xp) || xp < 0) {
      setError('XP reward must be 0 or greater.')
      return
    }
    const target = parseInt(form.target_count, 10)
    if (isNaN(target) || target < 1) {
      setError('Target count must be at least 1.')
      return
    }

    try {
      if (isEdit) {
        const input: UpdateQuestInput = {
          source_type: form.source_type,
          title: form.title.trim(),
          description: form.description.trim(),
          xp_reward: xp,
          target_count: target,
          is_active: initial!.is_active,
        }
        await updateQuest.mutateAsync({ id: initial!.id, input })
      } else {
        const input: CreateQuestInput = {
          type,
          source_type: form.source_type,
          title: form.title.trim(),
          description: form.description.trim(),
          xp_reward: xp,
          target_count: target,
        }
        await createQuest.mutateAsync(input)
      }
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong.')
    }
  }

  const isAuto = form.source_type !== 'manual'

  const autoSource =
    form.source_type === 'medicine'
      ? 'take a medicine'
      : form.source_type === 'supplement'
        ? 'take a supplement'
        : form.source_type === 'weight'
          ? 'log your weight'
          : 'finish a workout session'

  return createPortal(
    <div
      ref={backdropRef}
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
      onMouseDown={(e) => {
        if (e.target === backdropRef.current) onClose()
      }}
    >
      <div className="max-h-[90vh] w-full max-w-lg overflow-y-auto rounded-xl border border-gray-700 bg-gray-900 p-6 shadow-2xl">
        <h2 className="mb-4 text-base font-semibold text-gray-100">
          {isEdit ? 'Edit ' : 'Add '}
          {cfg.titleNoun}
        </h2>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div className="sm:col-span-2">
              <label className={labelClass}>Title</label>
              <input
                className={inputClass}
                placeholder={cfg.titlePlaceholder}
                value={form.title}
                onChange={(e) => setForm({ ...form, title: e.target.value })}
              />
            </div>
            <div className="sm:col-span-2">
              <label className={labelClass}>Description (optional)</label>
              <input
                className={inputClass}
                placeholder="Optional notes"
                value={form.description}
                onChange={(e) =>
                  setForm({ ...form, description: e.target.value })
                }
              />
            </div>
            <div>
              <label className={labelClass}>Source</label>
              <select
                className={inputClass}
                value={form.source_type}
                onChange={(e) =>
                  setForm({
                    ...form,
                    source_type: e.target.value as SourceType,
                  })
                }
              >
                <option value="manual">{cfg.manualOptionLabel}</option>
                <option value="medicine">Medicine (auto on take)</option>
                <option value="supplement">Supplement (auto on take)</option>
                <option value="workout">Workout (auto on finish)</option>
                <option value="weight">Weight (auto on log)</option>
              </select>
            </div>
            <div>
              <label className={labelClass}>Target count / {cfg.period}</label>
              <input
                className={inputClass}
                type="number"
                min="1"
                placeholder={cfg.targetPlaceholder}
                value={form.target_count}
                onChange={(e) =>
                  setForm({ ...form, target_count: e.target.value })
                }
              />
            </div>
            <div>
              <label className={labelClass}>XP Reward (on completion)</label>
              <input
                className={inputClass}
                type="number"
                min="0"
                placeholder="10"
                value={form.xp_reward}
                onChange={(e) =>
                  setForm({ ...form, xp_reward: e.target.value })
                }
              />
            </div>
          </div>

          {isAuto && (
            <p className="rounded-lg border border-amber-900/30 bg-amber-950/20 px-3 py-2 text-xs text-amber-600">
              This quest will {cfg.autoLead} when you {autoSource}.{' '}
              {cfg.autoTail}
            </p>
          )}

          {error && <p className="text-sm text-red-400">{error}</p>}

          <div className="flex justify-end gap-2 pt-1">
            <button
              type="submit"
              disabled={isPending}
              className="cursor-pointer rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
            >
              {isPending ? 'Saving…' : isEdit ? 'Update' : 'Add Quest'}
            </button>
            <button
              type="button"
              onClick={onClose}
              className="cursor-pointer rounded-lg bg-gray-800 px-4 py-2 text-sm font-medium text-gray-400 hover:bg-gray-700"
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>,
    document.body,
  )
}
