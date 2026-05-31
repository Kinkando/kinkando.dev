import type { MonthlySummary } from '../../lib/api/types'

function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(amount)
}

export default function SummaryPanel({ summary }: { summary: MonthlySummary }) {
  return (
    <div className="flex flex-col gap-4">
      <div className="grid grid-cols-3 gap-3">
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <p className="mb-1 text-xs text-gray-500">Income</p>
          <p className="text-lg font-semibold text-green-400">
            {formatCurrency(summary.income)}
          </p>
        </div>
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <p className="mb-1 text-xs text-gray-500">Expenses</p>
          <p className="text-lg font-semibold text-red-400">
            {formatCurrency(summary.expense)}
          </p>
        </div>
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <p className="mb-1 text-xs text-gray-500">Net</p>
          <p
            className={`text-lg font-semibold ${summary.net >= 0 ? 'text-indigo-400' : 'text-red-400'}`}
          >
            {formatCurrency(summary.net)}
          </p>
        </div>
      </div>
      {summary.categories.length > 0 && (
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <p className="mb-3 text-sm font-medium text-gray-300">By category</p>
          <ul className="flex flex-col gap-2">
            {summary.categories.map((cat, i) => (
              <li key={i} className="flex items-center justify-between text-sm">
                <span className="text-gray-400">{cat.category}</span>
                <span
                  className={`font-medium ${cat.type === 'income' ? 'text-green-400' : 'text-red-400'}`}
                >
                  {cat.type === 'income' ? '+' : '-'}
                  {formatCurrency(cat.total)}
                </span>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  )
}
