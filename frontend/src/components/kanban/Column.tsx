import { useState } from 'react'
import type { FormEvent } from 'react'
import { SortableContext, verticalListSortingStrategy } from '@dnd-kit/sortable'
import { useDroppable } from '@dnd-kit/core'
import type {
  Column as ColumnType,
  Card as CardType,
} from '../../lib/api/types'
import { useCreateCard } from '../../queries/useKanban'
import KanbanCard from './Card'

type Props = {
  column: ColumnType
  cards: CardType[]
}

export default function KanbanColumn({ column, cards }: Props) {
  const { setNodeRef, isOver } = useDroppable({ id: column.id })
  const createCard = useCreateCard()
  const [adding, setAdding] = useState(false)
  const [title, setTitle] = useState('')

  async function handleAdd(e: FormEvent) {
    e.preventDefault()
    if (!title.trim()) return
    await createCard.mutateAsync({
      column_id: column.id,
      title: title.trim(),
      content: '',
    })
    setTitle('')
    setAdding(false)
  }

  return (
    <div
      className={`flex w-72 shrink-0 flex-col gap-3 rounded-xl border bg-gray-900 p-4 ${
        isOver ? 'border-indigo-500' : 'border-gray-800'
      }`}
    >
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-gray-300">{column.name}</h3>
        <span className="text-xs text-gray-600">{cards.length}</span>
      </div>
      <div ref={setNodeRef} className="flex min-h-8 flex-col gap-2">
        <SortableContext
          items={cards.map((c) => c.id)}
          strategy={verticalListSortingStrategy}
        >
          {cards.map((card) => (
            <KanbanCard key={card.id} card={card} />
          ))}
        </SortableContext>
      </div>
      {adding ? (
        <form onSubmit={handleAdd} className="flex flex-col gap-2">
          <input
            autoFocus
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Card title"
            className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none"
          />
          <div className="flex gap-2">
            <button
              type="submit"
              disabled={createCard.isPending}
              className="flex-1 rounded-lg bg-indigo-600 py-1.5 text-sm text-white hover:bg-indigo-500 disabled:opacity-50"
            >
              Add
            </button>
            <button
              type="button"
              onClick={() => setAdding(false)}
              className="flex-1 rounded-lg bg-gray-800 py-1.5 text-sm text-gray-400 hover:bg-gray-700"
            >
              Cancel
            </button>
          </div>
        </form>
      ) : (
        <button
          onClick={() => setAdding(true)}
          className="text-left text-sm text-gray-500 hover:text-gray-300"
        >
          + Add card
        </button>
      )}
    </div>
  )
}
