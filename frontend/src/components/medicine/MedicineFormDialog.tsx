import { useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import type {
  Medicine,
  CreateMedicineInput,
  MedicineSourceType,
  FrequencyType,
  MedicineTiming,
} from '../../lib/api/types'
import { useCreateMedicine, useUpdateMedicine } from '../../queries/useMedicine'

type Props = {
  initial?: Medicine
  onClose: () => void
}

const inputClass =
  'w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none'
const labelClass = 'mb-1 block text-xs font-medium text-gray-400'

const FREQUENCY_TYPES: FrequencyType[] = [
  'daily',
  'weekly',
  'as_needed',
  'custom',
]
const FREQUENCY_LABELS: Record<FrequencyType, string> = {
  daily: 'Daily',
  weekly: 'Weekly',
  as_needed: 'As needed',
  custom: 'Custom',
}

const SOURCE_TYPES: MedicineSourceType[] = ['medication', 'supplement']
const SOURCE_LABELS: Record<MedicineSourceType, string> = {
  medication: 'Medication',
  supplement: 'Supplement',
}

const TIMINGS: Array<{ value: MedicineTiming | ''; label: string }> = [
  { value: '', label: 'Not specified' },
  { value: 'anytime', label: 'Anytime' },
  { value: 'before_meal', label: 'Before meal' },
  { value: 'after_meal', label: 'After meal' },
  { value: 'before_breakfast', label: 'Before breakfast' },
  { value: 'after_breakfast', label: 'After breakfast' },
  { value: 'before_lunch', label: 'Before lunch' },
  { value: 'after_lunch', label: 'After lunch' },
  { value: 'before_dinner', label: 'Before dinner' },
  { value: 'after_dinner', label: 'After dinner' },
  { value: 'before_bed', label: 'Before bed' },
]

type FormState = {
  name: string
  source_type: MedicineSourceType
  generic_name: string
  description: string
  stock_quantity: string
  stock_unit: string
  dosage_amount: string
  dosage_unit: string
  frequency_type: FrequencyType
  frequency_value: string
  timing: MedicineTiming | ''
  start_date: string
  end_date: string
  low_stock_threshold: string
  note: string
}

const defaultForm: FormState = {
  name: '',
  source_type: 'medication',
  generic_name: '',
  description: '',
  stock_quantity: '0',
  stock_unit: 'tablet',
  dosage_amount: '1',
  dosage_unit: 'tablet',
  frequency_type: 'daily',
  frequency_value: '',
  timing: '',
  start_date: '',
  end_date: '',
  low_stock_threshold: '7',
  note: '',
}

function medicineToForm(m: Medicine): FormState {
  return {
    name: m.name,
    source_type: m.source_type,
    generic_name: m.generic_name ?? '',
    description: m.description ?? '',
    stock_quantity: String(m.stock_quantity),
    stock_unit: m.stock_unit,
    dosage_amount: String(m.dosage_amount),
    dosage_unit: m.dosage_unit ?? '',
    frequency_type: m.frequency_type,
    frequency_value: m.frequency_value != null ? String(m.frequency_value) : '',
    timing: m.timing ?? '',
    start_date: m.start_date ? m.start_date.slice(0, 10) : '',
    end_date: m.end_date ? m.end_date.slice(0, 10) : '',
    low_stock_threshold: String(m.low_stock_threshold),
    note: m.note ?? '',
  }
}

export default function MedicineFormDialog({ initial, onClose }: Props) {
  const isEdit = !!initial
  const [form, setForm] = useState<FormState>(
    initial ? medicineToForm(initial) : defaultForm,
  )
  const [error, setError] = useState('')

  const createMedicine = useCreateMedicine()
  const updateMedicine = useUpdateMedicine()
  const isPending = createMedicine.isPending || updateMedicine.isPending

  const backdropRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [onClose])

  function buildPayload(): CreateMedicineInput {
    const threshold = parseFloat(form.low_stock_threshold)
    return {
      name: form.name.trim(),
      source_type: form.source_type,
      generic_name: form.generic_name.trim() || null,
      description: form.description.trim() || null,
      stock_quantity: parseFloat(form.stock_quantity) || 0,
      stock_unit: form.stock_unit.trim(),
      dosage_amount: parseFloat(form.dosage_amount) || 0,
      dosage_unit: form.dosage_unit.trim() || null,
      frequency_type: form.frequency_type,
      frequency_value: form.frequency_value
        ? parseInt(form.frequency_value, 10)
        : null,
      timing: form.timing || null,
      start_date: form.start_date || undefined,
      end_date: form.end_date || undefined,
      low_stock_threshold: isNaN(threshold) ? 7 : threshold,
      note: form.note.trim() || null,
    }
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')

    if (!form.name.trim()) {
      setError('Name is required.')
      return
    }
    const dosage = parseFloat(form.dosage_amount)
    if (isNaN(dosage) || dosage <= 0) {
      setError('Dosage amount must be greater than 0.')
      return
    }
    const stock = parseFloat(form.stock_quantity)
    if (isNaN(stock) || stock < 0) {
      setError('Stock quantity must be 0 or more.')
      return
    }
    if (!form.stock_unit.trim()) {
      setError('Stock unit is required.')
      return
    }

    try {
      if (isEdit) {
        await updateMedicine.mutateAsync({
          id: initial!.id,
          input: buildPayload(),
        })
      } else {
        await createMedicine.mutateAsync(buildPayload())
      }
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong.')
    }
  }

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
          {isEdit ? 'Edit Medicine' : 'Add Medicine'}
        </h2>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            {/* Name */}
            <div className="sm:col-span-2">
              <label className={labelClass}>Name *</label>
              <input
                className={inputClass}
                placeholder="e.g. Paracetamol"
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
              />
            </div>

            {/* Type */}
            <div>
              <label className={labelClass}>Type *</label>
              <select
                className={inputClass}
                value={form.source_type}
                onChange={(e) =>
                  setForm({
                    ...form,
                    source_type: e.target.value as MedicineSourceType,
                  })
                }
              >
                {SOURCE_TYPES.map((s) => (
                  <option key={s} value={s}>
                    {SOURCE_LABELS[s]}
                  </option>
                ))}
              </select>
            </div>

            {/* Generic name */}
            <div>
              <label className={labelClass}>Generic name</label>
              <input
                className={inputClass}
                placeholder="Optional"
                value={form.generic_name}
                onChange={(e) =>
                  setForm({ ...form, generic_name: e.target.value })
                }
              />
            </div>

            {/* Dosage */}
            <div>
              <label className={labelClass}>Dosage amount *</label>
              <input
                className={inputClass}
                type="number"
                min="0.01"
                step="0.01"
                placeholder="e.g. 0.5"
                value={form.dosage_amount}
                onChange={(e) =>
                  setForm({ ...form, dosage_amount: e.target.value })
                }
              />
            </div>

            {/* Dosage unit */}
            <div>
              <label className={labelClass}>Dosage unit</label>
              <input
                className={inputClass}
                placeholder="e.g. tablet"
                value={form.dosage_unit}
                onChange={(e) =>
                  setForm({ ...form, dosage_unit: e.target.value })
                }
              />
            </div>

            {/* Frequency */}
            <div>
              <label className={labelClass}>Frequency *</label>
              <select
                className={inputClass}
                value={form.frequency_type}
                onChange={(e) =>
                  setForm({
                    ...form,
                    frequency_type: e.target.value as FrequencyType,
                  })
                }
              >
                {FREQUENCY_TYPES.map((f) => (
                  <option key={f} value={f}>
                    {FREQUENCY_LABELS[f]}
                  </option>
                ))}
              </select>
            </div>

            {/* Frequency value */}
            {(form.frequency_type === 'daily' ||
              form.frequency_type === 'weekly' ||
              form.frequency_type === 'custom') && (
              <div>
                <label className={labelClass}>
                  Times per {form.frequency_type === 'weekly' ? 'week' : 'day'}
                </label>
                <input
                  className={inputClass}
                  type="number"
                  min="1"
                  step="1"
                  placeholder="e.g. 2"
                  value={form.frequency_value}
                  onChange={(e) =>
                    setForm({ ...form, frequency_value: e.target.value })
                  }
                />
              </div>
            )}

            {/* Timing */}
            <div>
              <label className={labelClass}>Timing</label>
              <select
                className={inputClass}
                value={form.timing}
                onChange={(e) =>
                  setForm({
                    ...form,
                    timing: e.target.value as MedicineTiming | '',
                  })
                }
              >
                {TIMINGS.map(({ value, label }) => (
                  <option key={value} value={value}>
                    {label}
                  </option>
                ))}
              </select>
            </div>

            {/* Stock quantity */}
            <div>
              <label className={labelClass}>Current stock *</label>
              <input
                className={inputClass}
                type="number"
                min="0"
                step="0.01"
                placeholder="e.g. 30"
                value={form.stock_quantity}
                onChange={(e) =>
                  setForm({ ...form, stock_quantity: e.target.value })
                }
              />
            </div>

            {/* Stock unit */}
            <div>
              <label className={labelClass}>Stock unit *</label>
              <input
                className={inputClass}
                placeholder="e.g. tablet"
                value={form.stock_unit}
                onChange={(e) =>
                  setForm({ ...form, stock_unit: e.target.value })
                }
              />
            </div>

            {/* Low stock threshold */}
            <div>
              <label className={labelClass}>Low stock warning at</label>
              <input
                className={inputClass}
                type="number"
                min="0"
                step="0.01"
                value={form.low_stock_threshold}
                onChange={(e) =>
                  setForm({ ...form, low_stock_threshold: e.target.value })
                }
              />
            </div>

            {/* Start / end date */}
            <div>
              <label className={labelClass}>Start date</label>
              <input
                className={inputClass}
                type="date"
                value={form.start_date}
                onChange={(e) =>
                  setForm({ ...form, start_date: e.target.value })
                }
              />
            </div>
            <div>
              <label className={labelClass}>End date</label>
              <input
                className={inputClass}
                type="date"
                value={form.end_date}
                onChange={(e) => setForm({ ...form, end_date: e.target.value })}
              />
            </div>

            {/* Description */}
            <div className="sm:col-span-2">
              <label className={labelClass}>Description</label>
              <textarea
                className={`${inputClass} resize-none`}
                rows={2}
                placeholder="Optional"
                value={form.description}
                onChange={(e) =>
                  setForm({ ...form, description: e.target.value })
                }
              />
            </div>

            {/* Note */}
            <div className="sm:col-span-2">
              <label className={labelClass}>Note</label>
              <input
                className={inputClass}
                placeholder="Optional"
                value={form.note}
                onChange={(e) => setForm({ ...form, note: e.target.value })}
              />
            </div>
          </div>

          {error && <p className="text-sm text-red-400">{error}</p>}

          <div className="flex justify-end gap-2 pt-1">
            <button
              type="submit"
              disabled={isPending}
              className="cursor-pointer rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
            >
              {isPending ? 'Saving…' : isEdit ? 'Save Changes' : 'Add Medicine'}
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
