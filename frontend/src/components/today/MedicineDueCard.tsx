import { useNavigate } from 'react-router-dom'
import type { Medicine, MedicineIntake } from '../../lib/api/types'

type Props = {
  medicines: Medicine[] | undefined
  intakes: MedicineIntake[] | undefined
}

export default function MedicineDueCard({ medicines, intakes }: Props) {
  const navigate = useNavigate()

  const active = medicines?.filter((m) => m.archived_at === null) ?? []
  const withReminder = active.filter((m) => m.reminder_enabled)
  const takenIds = new Set(intakes?.map((i) => i.medicine_id) ?? [])
  const takenCount = withReminder.filter((m) => takenIds.has(m.id)).length
  const lowStock = active.filter(
    (m) => m.stock_quantity > 0 && m.stock_quantity <= m.low_stock_threshold,
  )

  return (
    <div
      className="cursor-pointer rounded-xl border border-gray-800 bg-gray-900 p-4 transition-colors hover:border-gray-700"
      onClick={() => navigate('/health/medicine')}
    >
      <p className="mb-3 text-xs font-semibold tracking-widest text-rose-600 uppercase">
        Medicine &amp; Supplements
      </p>

      <div className="grid grid-cols-2 gap-2">
        <div className="rounded-lg bg-gray-800/60 px-3 py-2 text-center">
          <p className="text-xs text-gray-500">Taken today</p>
          <p className="text-lg font-bold text-gray-200">
            {takenCount}
            <span className="ml-0.5 text-xs text-gray-500">
              / {withReminder.length}
            </span>
          </p>
        </div>
        <div className="rounded-lg bg-gray-800/60 px-3 py-2 text-center">
          <p className="text-xs text-gray-500">Active</p>
          <p className="text-lg font-bold text-gray-200">{active.length}</p>
        </div>
      </div>

      {lowStock.length > 0 && (
        <div className="mt-3 flex items-center gap-1.5 rounded-lg bg-amber-950/30 px-3 py-1.5 text-xs text-amber-400">
          <span>⚠</span>
          <span>
            {lowStock.length} item{lowStock.length > 1 ? 's' : ''} low on stock
          </span>
        </div>
      )}
    </div>
  )
}
