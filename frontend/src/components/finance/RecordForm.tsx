import { useState, useRef, useEffect } from 'react'
import type { FormEvent } from 'react'
import type { CreateRecordInput, RecordType } from '../../lib/api/types'
import {
  useCreateRecord,
  useCategories,
  useFinanceNotes,
} from '../../queries/useFinance'
import { getIcon } from '../../lib/icons'
import { RECORD_TYPES, RECORD_TYPE_META } from '../../lib/finance'
import { cn } from '../../lib/cn'
import { ChevronDown } from 'lucide-react'
import { todayDate } from '../../lib/date'

export default function RecordForm({ month }: { month: string }) {
  const mutation = useCreateRecord(month)
  const { data: categories } = useCategories()
  const { data: notes } = useFinanceNotes()
  const [type, setType] = useState<RecordType>('expense')
  const [amount, setAmount] = useState('')
  const [categoryID, setCategoryID] = useState('')
  const [note, setNote] = useState('')
  const [date, setDate] = useState(todayDate)
  const [dropdownOpen, setDropdownOpen] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)
  const [noteOpen, setNoteOpen] = useState(false)
  const noteRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(e.target as Node)
      ) {
        setDropdownOpen(false)
      }
      if (noteRef.current && !noteRef.current.contains(e.target as Node)) {
        setNoteOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const filtered = (categories ?? []).filter((c) => c.type === type)

  const filteredNotes =
    note.length > 0
      ? (notes ?? []).filter((n) =>
          n.toLowerCase().includes(note.toLowerCase()),
        )
      : []
  const showNoteDropdown =
    noteOpen &&
    filteredNotes.length > 0 &&
    !(filteredNotes.length === 1 && filteredNotes[0] === note)

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
    setNoteOpen(false)
  }

  const inputClass =
    'px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-indigo-500'

  return (
    <form
      onSubmit={handleSubmit}
      className="flex flex-col gap-3 rounded-xl border border-gray-800 bg-gray-900 p-5"
    >
      <div className="flex gap-2">
        {RECORD_TYPES.map((t) => (
          <button
            key={t}
            type="button"
            onClick={() => handleTypeChange(t)}
            className={cn(
              'flex-1 cursor-pointer rounded-lg py-1.5 text-sm font-medium',
              type === t
                ? RECORD_TYPE_META[t].activeClass
                : 'bg-gray-800 text-gray-400 hover:bg-gray-700',
            )}
          >
            {RECORD_TYPE_META[t].label}
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
            className={`cursor-pointer ${inputClass} flex w-full items-center justify-between gap-2`}
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
            <ul className="absolute z-10 mt-1 max-h-48 w-full overflow-y-auto rounded-lg border border-gray-700 bg-gray-800 py-1 shadow-lg">
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
                      className={`flex w-full cursor-pointer items-center gap-2 px-3 py-2 text-sm hover:bg-gray-700 ${categoryID === cat.id ? 'bg-gray-700' : ''}`}
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
      <div ref={noteRef} className="relative">
        <input
          type="text"
          placeholder="Note (optional)"
          value={note}
          autoComplete="off"
          onChange={(e) => {
            const v = e.target.value
            setNote(v)
            setNoteOpen(v.length > 0)
          }}
          onFocus={() => setNoteOpen(note.length > 0)}
          className={`${inputClass} w-full`}
        />
        {showNoteDropdown && (
          <ul className="absolute z-10 mt-1 max-h-48 w-full overflow-y-auto rounded-lg border border-gray-700 bg-gray-800 py-1 shadow-lg">
            {filteredNotes.map((n) => (
              <li key={n}>
                <button
                  type="button"
                  onClick={() => {
                    setNote(n)
                    setNoteOpen(false)
                  }}
                  className="flex w-full cursor-pointer items-center px-3 py-2 text-left text-sm text-gray-100 hover:bg-gray-700"
                >
                  {n}
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>
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
        className="cursor-pointer rounded-lg bg-indigo-600 py-2 text-sm font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
      >
        {mutation.isPending ? 'Adding…' : 'Add record'}
      </button>
      {mutation.isError && (
        <p className="text-sm text-red-400">{String(mutation.error)}</p>
      )}
    </form>
  )
}
