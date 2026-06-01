import type { FinanceRecord } from '../../lib/api/types'
import { useDeleteRecord } from '../../queries/useFinance'
import { getIcon } from '../../lib/icons'

function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'THB',
  }).format(amount)
}

function formatGroupDate(date: string): string {
  return new Date(date + 'T00:00:00').toLocaleDateString('en-US', {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
  })
}

function formatTime(ts: string): string {
  return new Date(ts).toLocaleTimeString('en-US', {
    hour: 'numeric',
    minute: '2-digit',
  })
}

function groupByDate(records: FinanceRecord[]): [string, FinanceRecord[]][] {
  const map = new Map<string, FinanceRecord[]>()
  for (const r of records) {
    const key = r.date.slice(0, 10)
    const group = map.get(key)
    if (group) group.push(r)
    else map.set(key, [r])
  }
  return [...map.entries()].sort(([a], [b]) => b.localeCompare(a))
}

function dailyNet(records: FinanceRecord[]): number {
  return records.reduce(
    (sum, r) => sum + (r.type === 'income' ? r.amount : -r.amount),
    0,
  )
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

  const groups = groupByDate(records)

  return (
    <div className="flex flex-col gap-4">
      {groups.map(([date, groupRecords]) => {
        const net = dailyNet(groupRecords)
        return (
          <div key={date}>
            <div className="mb-2 flex items-center justify-between px-1">
              <span className="text-xs font-semibold tracking-wide text-gray-400 uppercase">
                {formatGroupDate(date)}
              </span>
              <span
                className={`text-xs font-semibold ${net >= 0 ? 'text-green-400' : 'text-red-400'}`}
              >
                {net >= 0 ? '+' : ''}
                {formatCurrency(net)}
              </span>
            </div>
            <ul className="flex flex-col gap-2">
              {groupRecords.map((record) => {
                const catName = record.category?.name
                const catIcon = record.category?.icon
                const catColor = record.category?.color
                const Icon = catIcon ? getIcon(catIcon) : null

                return (
                  <li
                    key={record.id}
                    className="rounded-xl border border-gray-800 bg-gray-900 px-4 py-3"
                  >
                    <div className="flex items-center gap-3">
                      {/* Category icon */}
                      {Icon && catColor ? (
                        <span
                          className="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-lg"
                          style={{
                            backgroundColor: catColor + '26',
                            color: catColor,
                          }}
                        >
                          <Icon size={15} />
                        </span>
                      ) : catColor ? (
                        <span
                          className="h-3 w-3 flex-shrink-0 rounded-full"
                          style={{ backgroundColor: catColor }}
                        />
                      ) : null}

                      {/* Info */}
                      <div className="min-w-0 flex-1">
                        <div className="flex flex-wrap items-center gap-1.5">
                          <span
                            className={`rounded-full px-2 py-0.5 text-xs font-medium ${
                              record.type === 'income'
                                ? 'bg-green-900 text-green-300'
                                : 'bg-red-900 text-red-300'
                            }`}
                          >
                            {record.type}
                          </span>
                          {catName && (
                            <span className="truncate text-sm font-medium text-gray-200">
                              {catName}
                            </span>
                          )}
                        </div>
                        <div className="mt-0.5 flex items-center gap-2">
                          <span className="text-xs text-gray-500">
                            {formatTime(record.created_at)}
                          </span>
                          {record.note && (
                            <span className="truncate text-xs text-gray-500">
                              {record.note}
                            </span>
                          )}
                        </div>
                      </div>

                      {/* Amount + delete */}
                      <div className="flex flex-shrink-0 items-center gap-2">
                        <span
                          className={`text-sm font-semibold ${record.type === 'income' ? 'text-green-400' : 'text-red-400'}`}
                        >
                          {record.type === 'income' ? '+' : '-'}
                          {formatCurrency(record.amount)}
                        </span>
                        <button
                          onClick={() => deleteMutation.mutate(record.id)}
                          disabled={deleteMutation.isPending}
                          className="text-xs text-gray-600 hover:text-red-400"
                        >
                          ✕
                        </button>
                      </div>
                    </div>
                  </li>
                )
              })}
            </ul>
          </div>
        )
      })}
    </div>
  )
}
