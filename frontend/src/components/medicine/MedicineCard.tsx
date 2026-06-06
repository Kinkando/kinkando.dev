import { useState } from 'react'
import { TriangleAlert } from 'lucide-react'
import type { Medicine } from '../../lib/api/types'
import {
  useArchiveMedicine,
  useUnarchiveMedicine,
} from '../../queries/useMedicine'
import {
  isLowStock,
  estimatedDaysRemaining,
  requiredDailyDoses,
} from '../../lib/medicine'
import TakeDialog from './TakeDialog'
import AdjustStockDialog from './AdjustStockDialog'
import MedicineFormDialog from './MedicineFormDialog'

type Props = {
  medicine: Medicine
  takenToday: number
}

const SOURCE_BADGE: Record<string, string> = {
  medication: 'bg-indigo-900/50 text-indigo-300',
  supplement: 'bg-emerald-900/50 text-emerald-300',
}

const SOURCE_LABELS: Record<string, string> = {
  medication: 'Medication',
  supplement: 'Supplement',
}

const TIMING_LABELS: Record<string, string> = {
  before_meal: 'Before meal',
  after_meal: 'After meal',
  before_breakfast: 'Before breakfast',
  after_breakfast: 'After breakfast',
  before_lunch: 'Before lunch',
  after_lunch: 'After lunch',
  before_dinner: 'Before dinner',
  after_dinner: 'After dinner',
  before_bed: 'Before bed',
  anytime: 'Anytime',
}

const FREQUENCY_LABELS: Record<string, string> = {
  daily: 'Daily',
  weekly: 'Weekly',
  as_needed: 'As needed',
  custom: 'Custom',
}

export default function MedicineCard({ medicine: med, takenToday }: Props) {
  const [showTake, setShowTake] = useState(false)
  const [showAdjust, setShowAdjust] = useState(false)
  const [showEdit, setShowEdit] = useState(false)
  const [archiveConfirm, setArchiveConfirm] = useState(false)

  const archiveMedicine = useArchiveMedicine()
  const unarchiveMedicine = useUnarchiveMedicine()

  const isArchived = med.archived_at != null
  const lowStock = isLowStock(med)
  const daysLeft = estimatedDaysRemaining(med)
  const requiredDoses = requiredDailyDoses(med) // null for non-daily frequencies
  const doseComplete = requiredDoses !== null && takenToday >= requiredDoses

  const frequencyLabel =
    FREQUENCY_LABELS[med.frequency_type] ?? med.frequency_type

  async function handleToggleArchive() {
    if (isArchived) {
      await unarchiveMedicine.mutateAsync(med.id)
    } else {
      await archiveMedicine.mutateAsync(med.id)
    }
    setArchiveConfirm(false)
  }

  return (
    <>
      <div
        className={`rounded-xl border bg-gray-900 p-4 ${isArchived ? 'border-gray-700 opacity-60' : 'border-gray-800'}`}
      >
        {/* Header */}
        <div className="mb-2 flex items-start justify-between gap-2">
          <div className="min-w-0 flex-1">
            <p className="truncate text-sm font-semibold text-gray-100">
              {med.name}
            </p>
            {med.generic_name && (
              <p className="truncate text-xs text-gray-500">
                {med.generic_name}
              </p>
            )}
          </div>
          <span
            className={`shrink-0 rounded px-1.5 py-0.5 text-xs ${SOURCE_BADGE[med.source_type] ?? 'bg-gray-700 text-gray-400'}`}
          >
            {SOURCE_LABELS[med.source_type] ?? med.source_type}
          </span>
          {isArchived && (
            <span className="shrink-0 rounded bg-gray-700 px-1.5 py-0.5 text-xs text-gray-400">
              Archived
            </span>
          )}
        </div>

        {/* Description */}
        {med.description && (
          <p className="mb-3 line-clamp-2 text-xs text-gray-500">
            {med.description}
          </p>
        )}

        {/* Stats */}
        <div className="mb-3 space-y-1.5">
          {/* Stock */}
          <div className="flex items-center justify-between">
            <span className="text-xs text-gray-500">Stock</span>
            <span
              className={`text-sm font-medium ${lowStock ? 'text-yellow-400' : 'text-gray-100'}`}
            >
              {med.stock_quantity} {med.stock_unit}
            </span>
          </div>

          {/* Low stock warning */}
          {lowStock && !isArchived && (
            <div className="flex items-center gap-1.5 rounded-md bg-yellow-900/30 px-2 py-1">
              <TriangleAlert
                className="h-3.5 w-3.5 text-yellow-400"
                strokeWidth={2}
              />
              <span className="text-xs text-yellow-300">
                Low stock (≤ {med.low_stock_threshold} {med.stock_unit})
              </span>
            </div>
          )}

          {/* Days remaining */}
          <div className="flex items-center justify-between">
            <span className="text-xs text-gray-500">Est. days left</span>
            <span className="text-xs text-gray-400">
              {daysLeft === null ? 'Unknown' : `${daysLeft} days`}
            </span>
          </div>

          {/* Dosage */}
          <div className="flex items-center justify-between">
            <span className="text-xs text-gray-500">Dosage</span>
            <span className="text-xs text-gray-400">
              {med.dosage_amount}
              {med.dosage_unit ? ` ${med.dosage_unit}` : ''} · {frequencyLabel}
            </span>
          </div>

          {/* Timing */}
          {med.timing && (
            <div className="flex items-center justify-between">
              <span className="text-xs text-gray-500">Timing</span>
              <span className="text-xs text-gray-400">
                {TIMING_LABELS[med.timing] ?? med.timing}
              </span>
            </div>
          )}

          {/* Today's intake progress (daily frequency only) */}
          {requiredDoses !== null && (
            <div className="flex items-center justify-between">
              <span className="text-xs text-gray-500">Today's Intake</span>
              <span
                className={`text-xs font-medium ${doseComplete ? 'text-emerald-400' : 'text-gray-400'}`}
              >
                {Math.min(takenToday, requiredDoses)}/{requiredDoses} doses
              </span>
            </div>
          )}
        </div>

        {/* Actions */}
        {!isArchived && (
          <div className="flex flex-wrap gap-2 border-t border-gray-800 pt-3">
            <button
              onClick={() => !doseComplete && setShowTake(true)}
              disabled={doseComplete}
              className="cursor-pointer rounded-lg bg-indigo-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-indigo-500 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {doseComplete ? 'Daily Dose Completed' : 'Take Medicine'}
            </button>
            <button
              onClick={() => setShowAdjust(true)}
              className="cursor-pointer rounded-lg bg-gray-800 px-3 py-1.5 text-xs font-medium text-gray-300 hover:bg-gray-700"
            >
              Adjust Stock
            </button>
            <button
              onClick={() => setShowEdit(true)}
              className="cursor-pointer rounded-lg bg-gray-800 px-3 py-1.5 text-xs font-medium text-gray-300 hover:bg-gray-700"
            >
              Edit
            </button>
            <button
              onClick={() => setArchiveConfirm(true)}
              className="cursor-pointer rounded-lg bg-gray-800 px-3 py-1.5 text-xs font-medium text-gray-500 hover:bg-gray-700 hover:text-gray-300"
            >
              Archive
            </button>
          </div>
        )}
        {isArchived && (
          <div className="flex gap-2 border-t border-gray-800 pt-3">
            <button
              onClick={() => unarchiveMedicine.mutate(med.id)}
              disabled={unarchiveMedicine.isPending}
              className="cursor-pointer rounded-lg bg-gray-800 px-3 py-1.5 text-xs font-medium text-gray-300 hover:bg-gray-700 disabled:opacity-50"
            >
              {unarchiveMedicine.isPending ? 'Unarchiving…' : 'Unarchive'}
            </button>
          </div>
        )}
      </div>

      {/* Archive confirm */}
      {archiveConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div className="w-full max-w-sm rounded-xl border border-gray-700 bg-gray-900 p-6 shadow-2xl">
            <p className="mb-4 text-sm text-gray-300">
              Archive{' '}
              <span className="font-medium text-gray-100">{med.name}</span>? It
              will be hidden from the active list.
            </p>
            <div className="flex justify-end gap-2">
              <button
                onClick={handleToggleArchive}
                disabled={archiveMedicine.isPending}
                className="cursor-pointer rounded-lg bg-gray-600 px-4 py-2 text-sm font-medium text-white hover:bg-gray-500 disabled:opacity-50"
              >
                {archiveMedicine.isPending ? 'Archiving…' : 'Archive'}
              </button>
              <button
                onClick={() => setArchiveConfirm(false)}
                className="cursor-pointer rounded-lg bg-gray-800 px-4 py-2 text-sm font-medium text-gray-300 hover:bg-gray-700"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      {showTake && (
        <TakeDialog medicine={med} onClose={() => setShowTake(false)} />
      )}
      {showAdjust && (
        <AdjustStockDialog
          medicine={med}
          onClose={() => setShowAdjust(false)}
        />
      )}
      {showEdit && (
        <MedicineFormDialog initial={med} onClose={() => setShowEdit(false)} />
      )}
    </>
  )
}
