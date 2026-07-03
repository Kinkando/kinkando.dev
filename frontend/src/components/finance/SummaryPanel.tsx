import { useState } from 'react'
import { Eye, EyeOff } from 'lucide-react'
import type { MonthlySummary } from '../../lib/api/types'
import { getIcon } from '../../lib/icons'
import { formatCurrency } from '../../lib/format'
import { cn } from '../../lib/cn'

const MASK = '••••••'

export default function SummaryPanel({ summary }: { summary: MonthlySummary }) {
  const [revealed, setRevealed] = useState(false)

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <p className="text-sm font-medium text-gray-300">Summary</p>
        <button
          type="button"
          onClick={() => setRevealed((v) => !v)}
          aria-label={revealed ? 'Hide amounts' : 'Show amounts'}
          aria-pressed={revealed}
          className="flex cursor-pointer items-center gap-1.5 rounded-lg border border-gray-800 bg-gray-900 px-2.5 py-1.5 text-xs text-gray-400 transition-colors hover:text-gray-200"
        >
          {revealed ? <EyeOff size={14} /> : <Eye size={14} />}
          {revealed ? 'Hide' : 'Show'}
        </button>
      </div>
      <div className="grid grid-cols-1 gap-3 md:grid-cols-3 lg:grid-cols-1">
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <p className="mb-1 text-xs text-gray-500">Income</p>
          <p className="truncate text-right text-lg font-semibold text-green-400">
            {revealed ? formatCurrency(summary.income) : MASK}
          </p>
        </div>
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <p className="mb-1 text-xs text-gray-500">Expenses</p>
          <p className="truncate text-right text-lg font-semibold text-red-400">
            {revealed ? formatCurrency(summary.expense) : MASK}
          </p>
        </div>
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <p className="mb-1 text-xs text-gray-500">Net</p>
          <p
            className={cn(
              'truncate text-right text-lg font-semibold',
              summary.net >= 0 ? 'text-indigo-400' : 'text-red-400',
            )}
          >
            {revealed ? formatCurrency(summary.net) : MASK}
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
                    className={cn(
                      'font-medium',
                      cat.type === 'income' ? 'text-green-400' : 'text-red-400',
                    )}
                  >
                    {cat.type === 'income' ? '+' : '-'}
                    {revealed ? formatCurrency(cat.total) : MASK}
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
