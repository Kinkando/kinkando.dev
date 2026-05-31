import type { FinanceRecord } from '../../lib/api/types'
import { useDeleteRecord } from '../../queries/useFinance'
import { getIcon } from '../../lib/icons'

function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(amount)
}

function formatDate(date: string): string {
  return new Date(date).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
  })
}

export default function RecordList({
  records,
  month,
}: {
  records: FinanceRecord[]
  month: string
}) {
  const deleteMutation = useDeleteRecord(month)

  if (!records.length) {
    return (
      <p className="py-4 text-sm text-gray-500">No records for this month.</p>
    )
  }

  return (
    <ul className="flex flex-col gap-2">
      {records.map((record) => {
        const catName = record.category?.name ?? record.category_name
        const catIcon = record.category?.icon
        const catColor = record.category?.color
        const Icon = catIcon ? getIcon(catIcon) : null

        return (
          <li
            key={record.id}
            className="flex items-center gap-3 rounded-xl border border-gray-800 bg-gray-900 px-4 py-3"
          >
            <span
              className={`rounded-full px-2 py-0.5 text-xs font-medium ${
                record.type === 'income'
                  ? 'bg-green-900 text-green-300'
                  : 'bg-red-900 text-red-300'
              }`}
            >
              {record.type}
            </span>
            <span className="text-sm text-gray-400">
              {formatDate(record.date)}
            </span>
            <span className="flex flex-1 items-center gap-1.5 text-sm font-medium text-gray-200">
              {Icon && catColor && (
                <Icon size={13} style={{ color: catColor }} />
              )}
              {catColor && (
                <span
                  className="h-2 w-2 flex-shrink-0 rounded-full"
                  style={{ backgroundColor: catColor }}
                />
              )}
              {catName}
            </span>
            {record.note && (
              <span className="max-w-32 truncate text-xs text-gray-500">
                {record.note}
              </span>
            )}
            <span
              className={`text-sm font-semibold ${record.type === 'income' ? 'text-green-400' : 'text-red-400'}`}
            >
              {record.type === 'income' ? '+' : '-'}
              {formatCurrency(record.amount)}
            </span>
            <button
              onClick={() => deleteMutation.mutate(record.id)}
              disabled={deleteMutation.isPending}
              className="ml-1 text-xs text-gray-600 hover:text-red-400"
            >
              ✕
            </button>
          </li>
        )
      })}
    </ul>
  )
}
