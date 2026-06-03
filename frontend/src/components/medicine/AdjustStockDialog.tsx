import { useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import type { Medicine, AdjustmentType } from '../../lib/api/types'
import { useAdjustStock } from '../../queries/useMedicine'

type Props = {
  medicine: Medicine
  onClose: () => void
}

const inputClass =
  'w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none'
const labelClass = 'mb-1 block text-xs font-medium text-gray-400'

const TYPES: Array<{ value: AdjustmentType; label: string; desc: string }> = [
  { value: 'add', label: 'Add stock', desc: 'Increase stock (e.g. restocked)' },
  {
    value: 'remove',
    label: 'Remove stock',
    desc: 'Decrease stock (e.g. discarded)',
  },
  {
    value: 'correction',
    label: 'Set exact',
    desc: 'Correct stock to an exact amount',
  },
]

export default function AdjustStockDialog({ medicine, onClose }: Props) {
  const [type, setType] = useState<AdjustmentType>('add')
  const [quantity, setQuantity] = useState('')
  const [reason, setReason] = useState('')
  const [error, setError] = useState('')

  const adjustStock = useAdjustStock()
  const backdropRef = useRef<HTMLDivElement>(null)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')

    const qty = parseFloat(quantity)
    if (isNaN(qty) || qty <= 0) {
      setError('Quantity must be greater than 0.')
      return
    }

    try {
      await adjustStock.mutateAsync({
        id: medicine.id,
        input: {
          type,
          quantity: qty,
          reason: reason.trim() || null,
        },
      })
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong.')
    }
  }

  function previewStock() {
    const qty = parseFloat(quantity) || 0
    if (qty <= 0) return null
    switch (type) {
      case 'add':
        return medicine.stock_quantity + qty
      case 'remove':
        return Math.max(0, medicine.stock_quantity - qty)
      case 'correction':
        return qty
    }
  }

  const preview = previewStock()

  return createPortal(
    <div
      ref={backdropRef}
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
      onMouseDown={(e) => {
        if (e.target === backdropRef.current) onClose()
      }}
    >
      <div className="w-full max-w-sm rounded-xl border border-gray-700 bg-gray-900 p-6 shadow-2xl">
        <h2 className="mb-1 text-base font-semibold text-gray-100">
          Adjust Stock
        </h2>
        <p className="mb-4 text-sm text-gray-400">
          {medicine.name} — current: {medicine.stock_quantity}{' '}
          {medicine.stock_unit}
        </p>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Type selector */}
          <div className="flex gap-2">
            {TYPES.map(({ value, label }) => (
              <button
                key={value}
                type="button"
                onClick={() => setType(value)}
                className={`flex-1 cursor-pointer rounded-lg border px-2 py-2 text-xs font-medium transition-colors ${
                  type === value
                    ? 'border-indigo-500 bg-indigo-600/20 text-indigo-300'
                    : 'border-gray-700 bg-gray-800 text-gray-400 hover:text-gray-200'
                }`}
              >
                {label}
              </button>
            ))}
          </div>

          {/* Quantity */}
          <div>
            <label className={labelClass}>
              {type === 'correction' ? 'New stock amount' : 'Quantity'} (
              {medicine.stock_unit})
            </label>
            <input
              className={inputClass}
              type="number"
              min="0.01"
              step="1"
              placeholder="e.g. 30"
              value={quantity}
              onChange={(e) => setQuantity(e.target.value)}
              autoFocus
            />
            {preview != null && (
              <p className="mt-1 text-xs text-gray-500">
                After:{' '}
                <span className="text-gray-300">
                  {preview} {medicine.stock_unit}
                </span>
              </p>
            )}
          </div>

          {/* Reason */}
          <div>
            <label className={labelClass}>Reason (optional)</label>
            <input
              className={inputClass}
              placeholder="e.g. Bought new pack"
              value={reason}
              onChange={(e) => setReason(e.target.value)}
            />
          </div>

          {error && <p className="text-sm text-red-400">{error}</p>}

          <div className="flex justify-end gap-2 pt-1">
            <button
              type="submit"
              disabled={adjustStock.isPending}
              className="cursor-pointer rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
            >
              {adjustStock.isPending ? 'Saving…' : 'Save'}
            </button>
            <button
              type="button"
              onClick={onClose}
              className="cursor-pointer rounded-lg bg-gray-800 px-4 py-2 text-sm font-medium text-gray-300 hover:bg-gray-700"
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
