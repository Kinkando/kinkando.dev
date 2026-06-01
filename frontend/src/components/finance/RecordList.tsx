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
                    className={`relative rounded-xl border px-3 pt-2 pb-3 ${
                      record.type === 'income'
                        ? 'border-green-900/40 bg-green-950/30'
                        : 'border-red-900/40 bg-red-950/30'
                    }`}
                  >
                    {/* Delete — top right */}
                    <button
                      onClick={() => deleteMutation.mutate(record.id)}
                      disabled={deleteMutation.isPending}
                      className="absolute top-2 right-2 text-xs text-gray-600 hover:text-red-400 disabled:opacity-40"
                    >
                      ✕
                    </button>

                    <div className="flex items-center gap-3 pr-5">
                      {/* Category icon — bg tinted by type */}
                      <span
                        className="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-xl"
                        style={{
                          backgroundColor:
                            record.type === 'income'
                              ? '#16a34a26'
                              : '#dc262626',
                          color:
                            catColor ??
                            (record.type === 'income' ? '#4ade80' : '#f87171'),
                        }}
                      >
                        {Icon ? (
                          <Icon size={16} />
                        ) : (
                          <span className="text-xs">?</span>
                        )}
                      </span>

                      {/* Info — priority: category › note › time */}
                      <div className="min-w-0 flex-1">
                        <span className="truncate text-sm font-medium text-gray-200">
                          {catName ?? '—'}
                        </span>
                        {record.note && (
                          <p className="mt-0.5 text-xs text-gray-400">
                            {record.note}
                          </p>
                        )}
                        <p className="mt-0.5 text-xs text-gray-600">
                          {formatTime(record.created_at)}
                        </p>
                      </div>

                      {/* Amount */}
                      <span
                        className={`shrink-0 text-sm font-semibold whitespace-nowrap ${
                          record.type === 'income'
                            ? 'text-green-400'
                            : 'text-red-400'
                        }`}
                      >
                        {record.type === 'income' ? '+' : '-'}
                        {formatCurrency(record.amount)}
                      </span>
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
