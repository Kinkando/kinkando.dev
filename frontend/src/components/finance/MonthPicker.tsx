'use client';

interface MonthPickerProps {
  value: string; // YYYY-MM
  onChange: (month: string) => void;
}

export function MonthPicker({ value, onChange }: MonthPickerProps) {
  function shiftMonth(delta: number) {
    const [y, m] = value.split('-').map(Number);
    const d = new Date(y, m - 1 + delta, 1);
    const newMonth = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`;
    onChange(newMonth);
  }

  return (
    <div className="flex items-center gap-2">
      <button onClick={() => shiftMonth(-1)} className="rounded px-2 py-1 text-gray-600 hover:bg-gray-100">
        &larr;
      </button>
      <input type="month" value={value} onChange={(e) => onChange(e.target.value)} className="rounded border border-gray-300 px-3 py-1.5 text-sm" />
      <button onClick={() => shiftMonth(1)} className="rounded px-2 py-1 text-gray-600 hover:bg-gray-100">
        &rarr;
      </button>
    </div>
  );
}
