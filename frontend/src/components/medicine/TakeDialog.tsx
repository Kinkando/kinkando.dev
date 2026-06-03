import { useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import type { Medicine } from '../../lib/api/types'
import { useTakeMedicine } from '../../queries/useMedicine'
import { ApiError } from '../../lib/api/client'

type Props = {
  medicine: Medicine
  onClose: () => void
}

const inputClass =
  'w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none'
const labelClass = 'mb-1 block text-xs font-medium text-gray-400'

export default function TakeDialog({ medicine, onClose }: Props) {
  const [quantity, setQuantity] = useState(String(medicine.dosage_amount))
  const [note, setNote] = useState('')
  const [error, setError] = useState('')
  const [showOverride, setShowOverride] = useState(false)

  const takeMedicine = useTakeMedicine()
  const backdropRef = useRef<HTMLDivElement>(null)

  async function submit(allowNegative: boolean) {
    const qty = parseFloat(quantity)
    if (isNaN(qty) || qty <= 0) {
      setError('Quantity must be greater than 0.')
      return
    }
    setError('')
    try {
      await takeMedicine.mutateAsync({
        id: medicine.id,
        input: {
          quantity_taken: qty,
          note: note.trim() || null,
          allow_negative: allowNegative,
        },
      })
      onClose()
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        setShowOverride(true)
        return
      }
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
      <div className="w-full max-w-sm rounded-xl border border-gray-700 bg-gray-900 p-6 shadow-2xl">
        <h2 className="mb-1 text-base font-semibold text-gray-100">
          Record Intake
        </h2>
        <p className="mb-4 text-sm text-gray-400">{medicine.name}</p>

        <div className="space-y-4">
          <div>
            <label className={labelClass}>
              Quantity taken ({medicine.stock_unit})
            </label>
            <input
              className={inputClass}
              type="number"
              min="0.5"
              step="0.5"
              value={quantity}
              onChange={(e) => {
                setQuantity(e.target.value)
                setShowOverride(false)
              }}
              autoFocus
            />
            <p className="mt-1 text-xs text-gray-500">
              Current stock: {medicine.stock_quantity} {medicine.stock_unit}
            </p>
          </div>

          <div>
            <label className={labelClass}>Note (optional)</label>
            <input
              className={inputClass}
              placeholder="Optional"
              value={note}
              onChange={(e) => setNote(e.target.value)}
            />
          </div>
        </div>

        {error && <p className="mt-3 text-sm text-red-400">{error}</p>}

        {showOverride && (
          <div className="mt-3 rounded-lg border border-yellow-600/40 bg-yellow-900/20 p-3">
            <p className="text-sm text-yellow-300">
              Stock will go below zero. Confirm override?
            </p>
            <div className="mt-2 flex justify-end gap-2">
              <button
                onClick={() => submit(true)}
                disabled={takeMedicine.isPending}
                className="cursor-pointer rounded-lg bg-yellow-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-yellow-500 disabled:opacity-50"
              >
                {takeMedicine.isPending ? 'Saving…' : 'Yes, record anyway'}
              </button>
              <button
                onClick={() => setShowOverride(false)}
                className="cursor-pointer rounded-lg bg-gray-800 px-3 py-1.5 text-xs font-medium text-gray-300 hover:bg-gray-700"
              >
                Cancel
              </button>
            </div>
          </div>
        )}

        {!showOverride && (
          <div className="mt-4 flex justify-end gap-2">
            <button
              onClick={() => submit(false)}
              disabled={takeMedicine.isPending}
              className="cursor-pointer rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
            >
              {takeMedicine.isPending ? 'Saving…' : 'Confirm'}
            </button>
            <button
              onClick={onClose}
              className="cursor-pointer rounded-lg bg-gray-800 px-4 py-2 text-sm font-medium text-gray-300 hover:bg-gray-700"
            >
              Cancel
            </button>
          </div>
        )}
      </div>
    </div>,
    document.body,
  )
}
