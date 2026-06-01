import type { MonthlySummary } from '../../lib/api/types'
import { getIcon } from '../../lib/icons'

function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(amount)
}

export default function SummaryPanel({ summary }: { summary: MonthlySummary }) {
  return (
    <div className="flex flex-col gap-4">
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <p className="mb-1 text-xs text-gray-500">Income</p>
          <p className="truncate text-lg font-semibold text-green-400">
            {formatCurrency(summary.income)}
          </p>
        </div>
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <p className="mb-1 text-xs text-gray-500">Expenses</p>
          <p className="truncate text-lg font-semibold text-red-400">
            {formatCurrency(summary.expense)}
          </p>
        </div>
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <p className="mb-1 text-xs text-gray-500">Net</p>
          <p
            className={`truncate text-lg font-semibold ${summary.net >= 0 ? 'text-indigo-400' : 'text-red-400'}`}
          >
            {formatCurrency(summary.net)}
          </p>
        </div>
      </div>
      {summary.categories.length > 0 && (
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <p className="mb-3 text-sm font-medium text-gray-300">By category</p>
          <ul className="flex flex-col gap-2">
            {summary.categories.map((cat, i) => {
              const Icon = cat.icon ? getIcon(cat.icon) : null
              return (
                <li
                  key={i}
                  className="flex items-center justify-between gap-2 text-sm"
                >
                  <span className="flex items-center gap-2 text-gray-400">
                    {Icon && cat.color && (
                      <Icon size={13} style={{ color: cat.color }} />
                    )}
                    {cat.color && (
                      <span
                        className="h-2 w-2 flex-shrink-0 rounded-full"
                        style={{ backgroundColor: cat.color }}
                      />
                    )}
                    {cat.category}
                  </span>
                  <span
                    className={`font-medium ${cat.type === 'income' ? 'text-green-400' : 'text-red-400'}`}
                  >
                    {cat.type === 'income' ? '+' : '-'}
                    {formatCurrency(cat.total)}
                  </span>
                </li>
              )
            })}
          </ul>
        </div>
      )}
    </div>
  )
}
