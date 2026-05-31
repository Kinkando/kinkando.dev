import { SortableContext, verticalListSortingStrategy } from '@dnd-kit/sortable'
import { useDroppable } from '@dnd-kit/core'
import type {
  Column as ColumnType,
  Card as CardType,
} from '../../lib/api/types'
import KanbanCard from './Card'

type Props = {
  column: ColumnType
  cards: CardType[]
  boardId: string
  onAddCard: (columnId: string) => void
  onEditCard: (card: CardType) => void
}

export default function KanbanColumn({
  column,
  cards,
  boardId,
  onAddCard,
  onEditCard,
}: Props) {
  const { setNodeRef, isOver } = useDroppable({ id: column.id })

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
            <KanbanCard
              key={card.id}
              card={card}
              boardId={boardId}
              onEdit={onEditCard}
            />
          ))}
        </SortableContext>
      </div>

      <button
        onClick={() => onAddCard(column.id)}
        className="text-left text-sm text-gray-500 hover:text-gray-300"
      >
        + Add card
      </button>
    </div>
  )
}
