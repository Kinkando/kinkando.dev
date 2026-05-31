import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { useDeleteCard } from '../../queries/useKanban'
import type { Card as CardType } from '../../lib/api/types'
import { PRIORITY_META, isOverdue, isDueSoon } from '../../lib/kanban'

type Props = {
  card: CardType
  boardId: string
  onEdit: (card: CardType) => void
}

export default function KanbanCard({ card, boardId, onEdit }: Props) {
  const deleteCard = useDeleteCard(boardId)
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: card.id })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.4 : 1,
  }

  const priorityMeta =
    card.priority && card.priority !== 'none'
      ? PRIORITY_META[card.priority]
      : null
  const overdue = isOverdue(card)
  const dueSoon = !overdue && isDueSoon(card)

  function formatDate(iso: string) {
    return new Date(iso).toLocaleDateString(undefined, {
      month: 'short',
      day: 'numeric',
    })
  }

  const cardStyle: React.CSSProperties = {
    ...style,
    ...(priorityMeta ? { borderLeftColor: priorityMeta.color } : undefined),
  }

  return (
    <div
      ref={setNodeRef}
      style={cardStyle}
      {...attributes}
      className={`group rounded-lg border bg-gray-800 p-3 ${
        priorityMeta ? 'border-l-2 border-gray-700' : 'border-gray-700'
      }`}
    >
      <div className="flex items-start gap-2">
        {/* Drag handle */}
        <div
          {...listeners}
          className="mt-0.5 flex-shrink-0 cursor-grab text-gray-600 select-none hover:text-gray-400"
        >
          ⠿
        </div>

        {/* Content — clickable to edit */}
        <div
          className="min-w-0 flex-1 cursor-pointer"
          onPointerDown={(e) => e.stopPropagation()}
          onClick={() => onEdit(card)}
        >
          <p className="text-sm break-words text-gray-200">{card.title}</p>

          {card.description && (
            <p className="mt-1 line-clamp-2 text-xs text-gray-500">
              {card.description}
            </p>
          )}

          {/* Metadata row */}
          {(priorityMeta || card.due_date || card.tags.length > 0) && (
            <div className="mt-2 flex flex-wrap items-center gap-1.5">
              {priorityMeta && (
                <span
                  className="rounded px-1.5 py-0.5 text-xs font-medium"
                  style={{
                    backgroundColor: priorityMeta.color + '26',
                    color: priorityMeta.color,
                  }}
                >
                  {priorityMeta.label}
                </span>
              )}

              {card.due_date && (
                <span
                  className={`rounded px-1.5 py-0.5 text-xs ${
                    overdue
                      ? 'bg-red-950/50 text-red-400'
                      : dueSoon
                        ? 'bg-yellow-950/50 text-yellow-400'
                        : 'bg-gray-700 text-gray-400'
                  }`}
                >
                  {overdue ? '⚠ ' : ''}
                  {formatDate(card.due_date)}
                </span>
              )}

              {card.tags.slice(0, 3).map((tag) => (
                <span
                  key={tag}
                  className="rounded-full bg-gray-700 px-2 py-0.5 text-xs text-gray-300"
                >
                  {tag}
                </span>
              ))}
              {card.tags.length > 3 && (
                <span className="text-xs text-gray-500">
                  +{card.tags.length - 3}
                </span>
              )}
            </div>
          )}
        </div>

        {/* Delete */}
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
