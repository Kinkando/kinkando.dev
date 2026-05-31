'use client';

import { type FormEvent, useState } from 'react';

import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import type { CreateRecordInput, RecordType } from '@/types/finance';

interface Props {
  onSubmit: (input: CreateRecordInput) => Promise<void>;
  onCancel: () => void;
}

export function RecordForm({ onSubmit, onCancel }: Props) {
  const [type, setType] = useState<RecordType>('expense');
  const [amount, setAmount] = useState('');
  const [category, setCategory] = useState('');
  const [note, setNote] = useState('');
  const [date, setDate] = useState(new Date().toISOString().slice(0, 10));
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      await onSubmit({
        type,
        amount: parseFloat(amount),
        category,
        note,
        date
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create record');
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {error && <p className="rounded bg-red-50 p-3 text-sm text-red-600">{error}</p>}

      <div className="flex gap-2">
        <button
          type="button"
          onClick={() => setType('expense')}
          className={`flex-1 rounded-lg py-2 text-sm font-medium ${type === 'expense' ? 'bg-red-100 text-red-700' : 'bg-gray-100 text-gray-500'}`}
        >
          Expense
        </button>
        <button
          type="button"
          onClick={() => setType('income')}
          className={`flex-1 rounded-lg py-2 text-sm font-medium ${type === 'income' ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'}`}
        >
          Income
        </button>
      </div>

      <Input type="number" step="0.01" min="0.01" placeholder="Amount" value={amount} onChange={(e) => setAmount(e.target.value)} required />
      <Input placeholder="Category" value={category} onChange={(e) => setCategory(e.target.value)} required />
      <Input placeholder="Note (optional)" value={note} onChange={(e) => setNote(e.target.value)} />
      <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} required />

      <div className="flex gap-2">
        <Button type="submit" className="flex-1" disabled={loading}>
          {loading ? 'Saving...' : 'Add Record'}
        </Button>
        <Button type="button" variant="secondary" onClick={onCancel}>
          Cancel
        </Button>
      </div>
    </form>
  );
}
