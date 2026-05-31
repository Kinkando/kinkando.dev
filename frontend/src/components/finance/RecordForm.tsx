import { useState } from 'react'
import type { FormEvent } from 'react'
import type { CreateRecordInput, RecordType } from '../../lib/api/types'
import { useCreateRecord, useCategories } from '../../queries/useFinance'
import { getIcon } from '../../lib/icons'

export default function RecordForm({ month }: { month: string }) {
  const mutation = useCreateRecord(month)
  const { data: categories } = useCategories()
  const [type, setType] = useState<RecordType>('income')
  const [amount, setAmount] = useState('')
  const [categoryID, setCategoryID] = useState('')
  const [note, setNote] = useState('')
  const [date, setDate] = useState(() => new Date().toISOString().slice(0, 10))

  const filtered = (categories ?? []).filter((c) => c.type === type)

  function handleTypeChange(t: RecordType) {
    setType(t)
    setCategoryID('')
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    const input: CreateRecordInput = {
      type,
      amount: parseFloat(amount),
      category_id: categoryID,
      note,
      date,
    }
    await mutation.mutateAsync(input)
    setAmount('')
    setCategoryID('')
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
            onClick={() => handleTypeChange(t)}
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
      {filtered.length === 0 ? (
        <p className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-500">
          No {type} categories — add one below
        </p>
      ) : (
        <select
          value={categoryID}
          required
          onChange={(e) => setCategoryID(e.target.value)}
          className={inputClass}
        >
          <option value="">Select category…</option>
          {filtered.map((cat) => {
            const Icon = getIcon(cat.icon)
            return (
              <option key={cat.id} value={cat.id}>
                {cat.name}
              </option>
            )
          })}
        </select>
      )}
      {categoryID &&
        filtered.length > 0 &&
        (() => {
          const cat = filtered.find((c) => c.id === categoryID)
          if (!cat) return null
          const Icon = getIcon(cat.icon)
          return (
            <div className="flex items-center gap-2 text-sm text-gray-400">
              <span
                className="flex h-5 w-5 items-center justify-center rounded"
                style={{ color: cat.color }}
              >
                <Icon size={14} />
              </span>
              <span style={{ color: cat.color }}>{cat.name}</span>
            </div>
          )
        })()}
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
        disabled={mutation.isPending || filtered.length === 0}
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
