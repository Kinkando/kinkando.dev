import { useState, useRef, useEffect } from 'react'
import type { FormEvent } from 'react'
import type { CreateRecordInput, RecordType } from '../../lib/api/types'
import { useCreateRecord, useCategories } from '../../queries/useFinance'
import { getIcon } from '../../lib/icons'
import { ChevronDown } from 'lucide-react'

export default function RecordForm({ month }: { month: string }) {
  const mutation = useCreateRecord(month)
  const { data: categories } = useCategories()
  const [type, setType] = useState<RecordType>('income')
  const [amount, setAmount] = useState('')
  const [categoryID, setCategoryID] = useState('')
  const [note, setNote] = useState('')
  const [date, setDate] = useState(() => new Date().toISOString().slice(0, 10))
  const [dropdownOpen, setDropdownOpen] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(e.target as Node)
      ) {
        setDropdownOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const filtered = (categories ?? []).filter((c) => c.type === type)

  function handleTypeChange(t: RecordType) {
    setType(t)
    setCategoryID('')
    setDropdownOpen(false)
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
        <div ref={dropdownRef} className="relative">
          <button
            type="button"
            onClick={() => setDropdownOpen((o) => !o)}
            className={`${inputClass} flex w-full items-center justify-between gap-2`}
          >
            {categoryID ? (
              (() => {
                const cat = filtered.find((c) => c.id === categoryID)!
                const Icon = getIcon(cat.icon)
                return (
                  <span className="flex items-center gap-2">
                    <span
                      className="flex h-5 w-5 items-center justify-center rounded"
                      style={{ color: cat.color }}
                    >
                      <Icon size={14} />
                    </span>
                    <span style={{ color: cat.color }}>{cat.name}</span>
                  </span>
                )
              })()
            ) : (
              <span className="text-gray-500">Select category…</span>
            )}
            <ChevronDown size={14} className="shrink-0 text-gray-400" />
          </button>
          {dropdownOpen && (
            <ul className="absolute z-10 mt-1 w-full rounded-lg border border-gray-700 bg-gray-800 py-1 shadow-lg">
              {filtered.map((cat) => {
                const Icon = getIcon(cat.icon)
                return (
                  <li key={cat.id}>
                    <button
                      type="button"
                      onClick={() => {
                        setCategoryID(cat.id)
                        setDropdownOpen(false)
                      }}
                      className={`flex w-full items-center gap-2 px-3 py-2 text-sm hover:bg-gray-700 ${categoryID === cat.id ? 'bg-gray-700' : ''}`}
                    >
                      <span
                        className="flex h-5 w-5 shrink-0 items-center justify-center rounded"
                        style={{ color: cat.color }}
                      >
                        <Icon size={14} />
                      </span>
                      <span style={{ color: cat.color }}>{cat.name}</span>
                    </button>
                  </li>
                )
              })}
            </ul>
          )}
        </div>
      )}
      <input
        type="text"
        required
        className="sr-only"
        value={categoryID}
        readOnly
        tabIndex={-1}
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
