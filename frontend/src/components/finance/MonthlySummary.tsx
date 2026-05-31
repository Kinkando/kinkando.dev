import { Card } from '@/components/ui/Card';
import type { MonthlySummary as MonthlySummaryType } from '@/types/finance';

interface Props {
  summary: MonthlySummaryType;
}

export function MonthlySummary({ summary }: Props) {
  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
      <Card>
        <p className="text-sm text-gray-500">Income</p>
        <p className="text-2xl font-bold text-green-600">+{summary.income.toLocaleString(undefined, { minimumFractionDigits: 2 })}</p>
      </Card>
      <Card>
        <p className="text-sm text-gray-500">Expense</p>
        <p className="text-2xl font-bold text-red-600">-{summary.expense.toLocaleString(undefined, { minimumFractionDigits: 2 })}</p>
      </Card>
      <Card>
        <p className="text-sm text-gray-500">Net</p>
        <p className={`text-2xl font-bold ${summary.net >= 0 ? 'text-green-600' : 'text-red-600'}`}>
          {summary.net >= 0 ? '+' : ''}
          {summary.net.toLocaleString(undefined, { minimumFractionDigits: 2 })}
        </p>
      </Card>
    </div>
  );
}
