import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { useDeleteCard } from '../../queries/useKanban'
import type { Card as CardType } from '../../lib/api/types'

export default function KanbanCard({ card }: { card: CardType }) {
  const deleteCard = useDeleteCard()
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: card.id,
  })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.4 : 1,
  }

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...attributes}
      className="group rounded-lg border border-gray-700 bg-gray-800 p-3"
    >
      <div className="flex items-start gap-2">
        <div
          {...listeners}
          className="mt-0.5 flex-shrink-0 cursor-grab text-gray-600 select-none hover:text-gray-400"
        >
          ⠿
        </div>
        <div className="min-w-0 flex-1">
          <p className="text-sm break-words text-gray-200">{card.title}</p>
          {card.content && (
            <p className="mt-1 text-xs break-words text-gray-500">
              {card.content}
            </p>
          )}
        </div>
        <button
          onPointerDown={(e) => e.stopPropagation()}
          onClick={() => deleteCard.mutate(card.id)}
          className="flex-shrink-0 text-xs text-gray-600 opacity-0 group-hover:opacity-100 hover:text-red-400"
        >
          ✕
        </button>
      </div>
    </div>
  )
}
