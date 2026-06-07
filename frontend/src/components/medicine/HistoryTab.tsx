import { useState } from 'react'
import type {
  MedicineIntake,
  MedicineStockAdjustment,
  MedicineSourceType,
} from '../../lib/api/types'
import {
  useMedicines,
  useMedicineIntakes,
  useStockAdjustments,
} from '../../queries/useMedicine'
import {
  STATUS_LABELS,
  STATUS_COLORS,
  ADJ_LABELS,
  ADJ_COLORS,
} from '../../lib/medicine'
import { formatDate, formatTime } from '../../lib/date'

const inputClass =
  'rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 focus:border-indigo-500 focus:outline-none'

function todayStr() {
  return new Date().toISOString().slice(0, 10)
}

type Props = {
  sourceType: MedicineSourceType
}

export default function HistoryTab({ sourceType }: Props) {
  const [filterDate, setFilterDate] = useState('')

  // Today's intakes always shown separately
  const todayIntakesQuery = useMedicineIntakes(sourceType, todayStr())
  // Filtered or recent intakes
  const recentIntakesQuery = useMedicineIntakes(
    sourceType,
    filterDate || undefined,
  )
  // Adjustments
  const adjustmentsQuery = useStockAdjustments(
    sourceType,
    filterDate || undefined,
  )

  // Map of medicine IDs belonging to the scoped source type (null = no scoping).
  const { data: medicines } = useMedicines(sourceType, true)
  const idSet = sourceType
    ? new Set(
        (medicines ?? [])
          .filter((m) => m.source_type === sourceType)
          .map((m) => m.id),
      )
    : null
  const matches = (medicineID: string) => !idSet || idSet.has(medicineID)

  const todayIntakes = (todayIntakesQuery.data ?? []).filter((i) =>
    matches(i.medicine_id),
  )
  const recentIntakesAll = (recentIntakesQuery.data ?? []).filter((i) =>
    matches(i.medicine_id),
  )
  const recentIntakes = filterDate
    ? recentIntakesAll
    : recentIntakesAll.slice(0, 20)
  const adjustmentsAll = (adjustmentsQuery.data ?? []).filter((a) =>
    matches(a.medicine_id),
  )
  const adjustments = filterDate ? adjustmentsAll : adjustmentsAll.slice(0, 20)

  return (
    <div className="space-y-6">
      {/* Date filter */}
      <div className="flex items-center gap-3">
        <label className="text-sm text-gray-400">Filter by date</label>
        <input
          type="date"
          className={inputClass}
          value={filterDate}
          onChange={(e) => setFilterDate(e.target.value)}
        />
        {filterDate && (
          <button
            onClick={() => setFilterDate('')}
            className="cursor-pointer text-xs text-gray-500 hover:text-gray-300"
          >
            Clear
          </button>
        )}
      </div>

      {/* Today's intakes (only when no date filter) */}
      {!filterDate && (
        <section>
          <h3 className="mb-3 text-sm font-medium text-gray-300">
            Today's intakes
          </h3>
          {todayIntakes.length === 0 ? (
            <p className="rounded-xl border border-gray-800 bg-gray-900 px-5 py-6 text-center text-sm text-gray-500">
              No intake logs today.
            </p>
          ) : (
            <IntakeList intakes={todayIntakes} />
          )}
        </section>
      )}

      {/* Intake logs */}
      <section>
        <h3 className="mb-3 text-sm font-medium text-gray-300">
          {filterDate ? 'Intake logs' : 'Recent intake logs'}
        </h3>
        {recentIntakesQuery.isLoading ? (
          <p className="py-4 text-center text-sm text-gray-500">Loading…</p>
        ) : recentIntakes.length === 0 ? (
          <p className="rounded-xl border border-gray-800 bg-gray-900 px-5 py-6 text-center text-sm text-gray-500">
            No intake logs{filterDate ? ' for this date' : ''}.
          </p>
        ) : (
          <IntakeList intakes={recentIntakes} showDate />
        )}
      </section>

      {/* Stock adjustments */}
      <section>
        <h3 className="mb-3 text-sm font-medium text-gray-300">
          {filterDate ? 'Stock adjustments' : 'Recent stock adjustments'}
        </h3>
        {adjustmentsQuery.isLoading ? (
          <p className="py-4 text-center text-sm text-gray-500">Loading…</p>
        ) : adjustments.length === 0 ? (
          <p className="rounded-xl border border-gray-800 bg-gray-900 px-5 py-6 text-center text-sm text-gray-500">
            No stock adjustments yet.
          </p>
        ) : (
          <AdjustmentList adjustments={adjustments} />
        )}
      </section>
    </div>
  )
}

function IntakeList({
  intakes,
  showDate = false,
}: {
  intakes: MedicineIntake[]
  showDate?: boolean
}) {
  return (
    <div className="rounded-xl border border-gray-800 bg-gray-900">
      <ul className="divide-y divide-gray-800">
        {intakes.map((intake) => (
          <li key={intake.id} className="flex items-center gap-3 px-5 py-3">
            <span className="w-12 shrink-0 text-xs text-gray-500">
              {showDate
                ? formatDate(intake.taken_at, {
                    month: 'short',
                    day: 'numeric',
                  })
                : formatTime(intake.taken_at)}
            </span>
            <span
              className={`shrink-0 rounded px-1.5 py-0.5 text-xs ${STATUS_COLORS[intake.status]}`}
            >
              {STATUS_LABELS[intake.status]}
            </span>
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm text-gray-100">
                {intake.medicine_name}
              </p>
              {intake.note && (
                <p className="truncate text-xs text-gray-500">{intake.note}</p>
              )}
            </div>
            <span className="shrink-0 text-xs text-gray-400">
              {intake.quantity_taken}
            </span>
            {showDate && (
              <span className="shrink-0 text-xs text-gray-600">
                {formatTime(intake.taken_at)}
              </span>
            )}
          </li>
        ))}
      </ul>
    </div>
  )
}

function AdjustmentList({
  adjustments,
}: {
  adjustments: MedicineStockAdjustment[]
}) {
  return (
    <div className="rounded-xl border border-gray-800 bg-gray-900">
      <ul className="divide-y divide-gray-800">
        {adjustments.map((adj) => (
          <li key={adj.id} className="flex items-center gap-3 px-5 py-3">
            <span className="w-12 shrink-0 text-xs text-gray-500">
              {formatDate(adj.created_at, { month: 'short', day: 'numeric' })}
            </span>
            <span
              className={`shrink-0 rounded px-1.5 py-0.5 text-xs ${ADJ_COLORS[adj.type]}`}
            >
              {ADJ_LABELS[adj.type]}
            </span>
            <div className="min-w-0 flex-1">
              <p className="truncate text-xs text-gray-500">
                {adj.stock_before} → {adj.stock_after}
              </p>
              {adj.reason && (
                <p className="truncate text-xs text-gray-600">{adj.reason}</p>
              )}
            </div>
            <span className="shrink-0 text-xs text-gray-400">
              {adj.quantity}
            </span>
            <span className="shrink-0 text-xs text-gray-600">
              {formatTime(adj.created_at)}
            </span>
          </li>
        ))}
      </ul>
    </div>
  )
}
