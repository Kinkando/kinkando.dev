import { useState } from 'react'
import type { FormEvent } from 'react'
import type { CreateRecordInput, RecordType } from '../../lib/api/types'
import { useCreateRecord } from '../../queries/useFinance'

export default function RecordForm({ month }: { month: string }) {
  const mutation = useCreateRecord(month)
  const [type, setType] = useState<RecordType>('income')
  const [amount, setAmount] = useState('')
  const [category, setCategory] = useState('')
  const [note, setNote] = useState('')
  const [date, setDate] = useState(() => new Date().toISOString().slice(0, 10))

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    const input: CreateRecordInput = {
      type,
      amount: parseFloat(amount),
      category,
      note,
      date,
    }
    await mutation.mutateAsync(input)
    setAmount('')
    setCategory('')
    setNote('')
  }

  const inputClass =
    'px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-indigo-500'

  return (
    <form
      onSubmit={handleSubmit}
      className="flex flex-col gap-3 rounded-xl border border-gray-800 bg-gray-900 p-5"
    >
      <div className="flex gap-2">
        {(['income', 'expense'] as RecordType[]).map((t) => (
          <button
            key={t}
            type="button"
            onClick={() => setType(t)}
            className={`flex-1 rounded-lg py-1.5 text-sm font-medium ${
              type === t
                ? t === 'income'
                  ? 'bg-green-700 text-white'
                  : 'bg-red-700 text-white'
                : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
            }`}
          >
            {t.charAt(0).toUpperCase() + t.slice(1)}
          </button>
        ))}
      </div>
      <input
        type="number"
        placeholder="Amount"
        value={amount}
        min="0.01"
        step="0.01"
        required
        onChange={(e) => setAmount(e.target.value)}
        className={inputClass}
      />
      <input
        type="text"
        placeholder="Category"
        value={category}
        required
        onChange={(e) => setCategory(e.target.value)}
        className={inputClass}
      />
      <input
        type="text"
        placeholder="Note (optional)"
        value={note}
        onChange={(e) => setNote(e.target.value)}
        className={inputClass}
      />
      <input
        type="date"
        value={date}
        required
        onChange={(e) => setDate(e.target.value)}
        className={inputClass}
      />
      <button
        type="submit"
        disabled={mutation.isPending}
        className="rounded-lg bg-indigo-600 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
      >
        {mutation.isPending ? 'Adding…' : 'Add record'}
      </button>
      {mutation.isError && (
        <p className="text-sm text-red-400">{String(mutation.error)}</p>
      )}
    </form>
  )
}
