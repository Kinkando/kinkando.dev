'use client';

import { Button } from '@/components/ui/Button';
import type { FinanceRecord } from '@/types/finance';

interface Props {
  records: FinanceRecord[];
  onDelete: (id: string) => void;
}

export function RecordList({ records, onDelete }: Props) {
  if (records.length === 0) {
    return <p className="py-8 text-center text-sm text-gray-500">No records for this month.</p>;
  }

  const typeBadge = (type: FinanceRecord['type']) => (
    <span
      className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${
        type === 'income' ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'
      }`}
    >
      {type}
    </span>
  );

  const amount = (value: number) => value.toLocaleString(undefined, { minimumFractionDigits: 2 });

  return (
    <>
      {/* Table layout — sm and up */}
      <div className="hidden overflow-x-auto sm:block">
        <table className="w-full text-left text-sm">
          <thead>
            <tr className="border-b border-gray-200 text-gray-500">
              <th className="px-3 py-2 font-medium">Date</th>
              <th className="px-3 py-2 font-medium">Type</th>
              <th className="px-3 py-2 font-medium">Category</th>
              <th className="px-3 py-2 font-medium">Note</th>
              <th className="px-3 py-2 text-right font-medium">Amount</th>
              <th className="px-3 py-2" />
            </tr>
          </thead>
          <tbody>
            {records.map((rec) => (
              <tr key={rec.id} className="border-b border-gray-100">
                <td className="px-3 py-2">{rec.date.slice(0, 10)}</td>
                <td className="px-3 py-2">{typeBadge(rec.type)}</td>
                <td className="px-3 py-2">{rec.category}</td>
                <td className="px-3 py-2 text-gray-500">{rec.note}</td>
                <td className="px-3 py-2 text-right font-mono">{amount(rec.amount)}</td>
                <td className="px-3 py-2">
                  <Button variant="danger" className="px-2 py-1 text-xs" onClick={() => onDelete(rec.id)}>
                    Delete
                  </Button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Stacked cards — below sm */}
      <div className="divide-y divide-gray-100 sm:hidden">
        {records.map((rec) => (
          <div key={rec.id} className="flex flex-col gap-2 p-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                {typeBadge(rec.type)}
                <span className="text-sm text-gray-500">{rec.date.slice(0, 10)}</span>
              </div>
              <span className="font-mono text-sm">{amount(rec.amount)}</span>
            </div>
            <div className="text-sm text-gray-900">{rec.category}</div>
            {rec.note && <div className="text-sm text-gray-500">{rec.note}</div>}
            <div>
              <Button variant="danger" className="px-2 py-1 text-xs" onClick={() => onDelete(rec.id)}>
                Delete
              </Button>
            </div>
          </div>
        ))}
      </div>
    </>
  );
}
