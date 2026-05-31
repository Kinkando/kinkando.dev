import { useState } from 'react'
import type { FormEvent } from 'react'
import type { Category, RecordType } from '../../lib/api/types'
import {
  useCategories,
  useCreateCategory,
  useUpdateCategory,
  useDeleteCategory,
} from '../../queries/useFinance'
import { getIcon, ICON_NAMES } from '../../lib/icons'
import { ChevronDown } from 'lucide-react'

const DEFAULT_COLORS = [
  '#6366f1',
  '#22c55e',
  '#ef4444',
  '#f59e0b',
  '#3b82f6',
  '#ec4899',
  '#14b8a6',
  '#f97316',
  '#8b5cf6',
  '#06b6d4',
]

function CategoryRow({
  cat,
  onEdit,
}: {
  cat: Category
  onEdit: (cat: Category) => void
}) {
  const deleteMutation = useDeleteCategory()
  const Icon = getIcon(cat.icon)

  return (
    <li className="flex items-center gap-3 rounded-lg border border-gray-800 bg-gray-950 px-3 py-2">
      <span
        className="flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-md"
        style={{ backgroundColor: cat.color + '26', color: cat.color }}
      >
        <Icon size={14} />
      </span>
      <span className="flex-1 text-sm text-gray-200">{cat.name}</span>
      <button
        onClick={() => onEdit(cat)}
        className="text-xs text-gray-500 hover:text-gray-300"
      >
        Edit
      </button>
      <button
        onClick={() => {
          if (confirm(`Delete "${cat.name}"?`)) {
            deleteMutation.mutate(cat.id)
          }
        }}
        disabled={deleteMutation.isPending}
        className="text-xs text-gray-600 hover:text-red-400"
      >
        ✕
      </button>
    </li>
  )
}

function CategoryForm({
  initial,
  typeDefault,
  onDone,
}: {
  initial?: Category
  typeDefault: RecordType
  onDone: () => void
}) {
  const createMutation = useCreateCategory()
  const updateMutation = useUpdateCategory()

  const [name, setName] = useState(initial?.name ?? '')
  const [type, setType] = useState<RecordType>(initial?.type ?? typeDefault)
  const [icon, setIcon] = useState(initial?.icon ?? 'CircleDollarSign')
  const [color, setColor] = useState(initial?.color ?? '#6366f1')

  const isEdit = !!initial
  const isPending = createMutation.isPending || updateMutation.isPending

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    if (isEdit) {
      await updateMutation.mutateAsync({
        id: initial.id,
        input: { name, icon, color },
      })
    } else {
      await createMutation.mutateAsync({ name, type, icon, color })
    }
    onDone()
  }

  const inputClass =
    'px-2 py-1.5 bg-gray-800 border border-gray-700 rounded-lg text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-indigo-500'

  return (
    <form
      onSubmit={handleSubmit}
      className="flex flex-col gap-3 rounded-lg border border-gray-700 bg-gray-950 p-3"
    >
      {!isEdit && (
        <div className="flex gap-2">
          {(['income', 'expense'] as RecordType[]).map((t) => (
            <button
              key={t}
              type="button"
              onClick={() => setType(t)}
              className={`flex-1 rounded-lg py-1 text-xs font-medium ${
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
      )}
      <input
        type="text"
        placeholder="Category name"
        value={name}
        required
        onChange={(e) => setName(e.target.value)}
        className={inputClass}
      />
      <div className="flex flex-col gap-1.5">
        <p className="text-xs text-gray-500">Icon</p>
        <div className="grid grid-cols-5 gap-1">
          {ICON_NAMES.map((n) => {
            const Ic = getIcon(n)
            return (
              <button
                key={n}
                type="button"
                title={n}
                onClick={() => setIcon(n)}
                className={`flex items-center justify-center rounded-md p-1.5 text-sm ${
                  icon === n
                    ? 'bg-indigo-600 text-white'
                    : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
                }`}
              >
                <Ic size={14} />
              </button>
            )
          })}
        </div>
      </div>
      <div className="flex flex-col gap-1.5">
        <p className="text-xs text-gray-500">Color</p>
        <div className="flex flex-wrap gap-1.5">
          {DEFAULT_COLORS.map((c) => (
            <button
              key={c}
              type="button"
              onClick={() => setColor(c)}
              className={`h-6 w-6 rounded-full ${color === c ? 'ring-2 ring-white ring-offset-1 ring-offset-gray-900' : ''}`}
              style={{ backgroundColor: c }}
            />
          ))}
          <input
            type="color"
            value={color}
            onChange={(e) => setColor(e.target.value)}
            className="h-6 w-6 cursor-pointer rounded-full border-0 bg-transparent p-0"
            title="Custom color"
          />
        </div>
      </div>
      <div className="flex gap-2">
        <button
          type="submit"
          disabled={isPending}
          className="flex-1 rounded-lg bg-indigo-600 py-1.5 text-xs font-medium text-white hover:bg-indigo-500 disabled:opacity-50"
        >
          {isPending ? 'Saving…' : isEdit ? 'Save' : 'Add category'}
        </button>
        <button
          type="button"
          onClick={onDone}
          className="rounded-lg bg-gray-800 px-3 py-1.5 text-xs text-gray-400 hover:bg-gray-700"
        >
          Cancel
        </button>
      </div>
      {(createMutation.isError || updateMutation.isError) && (
        <p className="text-xs text-red-400">
          {String(createMutation.error ?? updateMutation.error)}
        </p>
      )}
    </form>
  )
}

export default function CategoryManager() {
  const { data: categories, isLoading } = useCategories()
  const [expanded, setExpanded] = useState(false)
  const [adding, setAdding] = useState<RecordType | null>(null)
  const [editing, setEditing] = useState<Category | null>(null)

  const income = (categories ?? []).filter((c) => c.type === 'income')
  const expense = (categories ?? []).filter((c) => c.type === 'expense')

  return (
    <div className="flex flex-col rounded-xl border border-gray-800 bg-gray-900">
      <button
        type="button"
        onClick={() => setExpanded((v) => !v)}
        className="flex items-center justify-between px-5 py-4 text-left"
      >
        <h2 className="text-sm font-semibold text-gray-300">Categories</h2>
        <ChevronDown
          size={16}
          className={`text-gray-500 transition-transform duration-200 ${expanded ? 'rotate-180' : ''}`}
        />
      </button>

      {expanded && (
        <div className="flex flex-col gap-4 border-t border-gray-800 px-5 pt-4 pb-5">
          {isLoading ? (
            <p className="text-sm text-gray-500">Loading categories…</p>
          ) : (
            (['income', 'expense'] as RecordType[]).map((t) => {
              const list = t === 'income' ? income : expense
              return (
                <div key={t} className="flex flex-col gap-2">
                  <div className="flex items-center justify-between">
                    <span
                      className={`text-xs font-medium ${t === 'income' ? 'text-green-400' : 'text-red-400'}`}
                    >
                      {t.charAt(0).toUpperCase() + t.slice(1)}
                    </span>
                    {adding !== t && editing?.type !== t && (
                      <button
                        onClick={() => setAdding(t)}
                        className="text-xs text-gray-500 hover:text-indigo-400"
                      >
                        + Add
                      </button>
                    )}
                  </div>

                  {list.length > 0 && (
                    <ul className="flex flex-col gap-1.5">
                      {list.map((cat) =>
                        editing?.id === cat.id ? (
                          <CategoryForm
                            key={cat.id}
                            initial={cat}
                            typeDefault={t}
                            onDone={() => setEditing(null)}
                          />
                        ) : (
                          <CategoryRow
                            key={cat.id}
                            cat={cat}
                            onEdit={(c) => {
                              setAdding(null)
                              setEditing(c)
                            }}
                          />
                        ),
                      )}
                    </ul>
                  )}

                  {adding === t && (
                    <CategoryForm
                      typeDefault={t}
                      onDone={() => setAdding(null)}
                    />
                  )}
                </div>
              )
            })
          )}
        </div>
      )}
    </div>
  )
}
