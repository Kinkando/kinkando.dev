import { createPortal } from 'react-dom'
import { useEffect, useRef, useState } from 'react'
import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { useDeleteCard, useArchiveCard } from '../../queries/useKanban'
import type {
  ArchiveReason,
  Card as CardType,
  ColumnType,
} from '../../lib/api/types'
import { PRIORITY_META, isOverdue, isDueSoon } from '../../lib/kanban'

type Props = {
  card: CardType
  boardId: string
  columnType: ColumnType
  onEdit: (card: CardType) => void
}

export default function KanbanCard({
  card,
  boardId,
  columnType,
  onEdit,
}: Props) {
  const deleteCard = useDeleteCard(boardId)
  const archiveCard = useArchiveCard(boardId)
  const [showReasonPicker, setShowReasonPicker] = useState(false)
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)

  const { attributes, listeners, setNodeRef, transform, isDragging } =
    useSortable({ id: card.id, data: { type: 'card' } })

  const style: React.CSSProperties = {
    transform: CSS.Transform.toString(transform),
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

  function handleArchiveClick(e: React.MouseEvent) {
    e.stopPropagation()
    if (columnType === 'done') {
      archiveCard.mutate({ id: card.id, input: {} })
    } else {
      setShowReasonPicker(true)
    }
  }

  const cardStyle: React.CSSProperties = {
    ...style,
    ...(priorityMeta ? { borderLeftColor: priorityMeta.color } : undefined),
  }

  return (
    <>
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
            className="mt-0.5 flex-shrink-0 cursor-grab text-sm text-gray-600 select-none hover:text-gray-400"
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

            {/* Completed indicator */}
            {card.completed_at && (
              <p className="mt-1 text-xs text-emerald-600">
                ✓ Completed {formatDate(card.completed_at)}
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

          {/* Action buttons (visible on hover) */}
          <div
            className="flex flex-shrink-0 flex-col gap-1 opacity-0 group-hover:opacity-100"
            onPointerDown={(e) => e.stopPropagation()}
          >
            {/* Archive */}
            <button
              onClick={handleArchiveClick}
              className="text-xs text-gray-600 hover:text-amber-400"
              title={columnType === 'done' ? 'Archive as completed' : 'Archive'}
            >
              ⊙
            </button>
            {/* Delete */}
            <button
              onClick={() => setShowDeleteConfirm(true)}
              className="text-xs text-gray-600 hover:text-red-400"
              title="Delete permanently"
            >
              ✕
            </button>
          </div>
        </div>
      </div>

      {showDeleteConfirm && (
        <ConfirmDeleteModal
          title={card.title}
          onConfirm={() => {
            deleteCard.mutate(card.id)
            setShowDeleteConfirm(false)
          }}
          onClose={() => setShowDeleteConfirm(false)}
        />
      )}

      {showReasonPicker && (
        <ArchiveReasonModal
          onConfirm={(reason) => {
            archiveCard.mutate({ id: card.id, input: { reason } })
            setShowReasonPicker(false)
          }}
          onClose={() => setShowReasonPicker(false)}
        />
      )}
    </>
  )
}

// ---- Confirm permanent delete ----------------------------------------------

function ConfirmDeleteModal({
  title,
  onConfirm,
  onClose,
}: {
  title: string
  onConfirm: () => void
  onClose: () => void
}) {
  const backdropRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [onClose])

  return createPortal(
    <div
      ref={backdropRef}
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
      onMouseDown={(e) => {
        if (e.target === backdropRef.current) onClose()
      }}
    >
      <div className="w-full max-w-sm rounded-xl border border-gray-700 bg-gray-900 p-6 shadow-2xl">
        <h2 className="mb-1 text-base font-semibold text-gray-100">
          Delete card?
        </h2>
        <p className="mb-5 text-sm text-gray-400">
          &ldquo;{title}&rdquo; will be permanently deleted. This cannot be
          undone.
        </p>
        <div className="flex gap-2">
          <button
            onClick={onConfirm}
            className="flex-1 rounded-lg bg-red-600 py-2 text-sm font-medium text-white hover:bg-red-500"
          >
            Delete
          </button>
          <button
            onClick={onClose}
            className="flex-1 rounded-lg bg-gray-800 py-2 text-sm font-medium text-gray-400 hover:bg-gray-700"
          >
            Cancel
          </button>
        </div>
      </div>
    </div>,
    document.body,
  )
}

// ---- Archive reason picker -------------------------------------------------

const REASONS: {
  value: 'cancelled' | 'duplicate' | 'stale'
  label: string
  desc: string
}[] = [
  {
    value: 'cancelled',
    label: 'Cancelled',
    desc: 'This work is no longer needed.',
  },
  {
    value: 'duplicate',
    label: 'Duplicate',
    desc: 'There is another card covering this.',
  },
  {
    value: 'stale',
    label: 'Stale',
    desc: 'This has been inactive for too long.',
  },
]

function ArchiveReasonModal({
  onConfirm,
  onClose,
}: {
  onConfirm: (reason: Exclude<ArchiveReason, 'completed'>) => void
  onClose: () => void
}) {
  const [selected, setSelected] = useState<'cancelled' | 'duplicate' | 'stale'>(
    'cancelled',
  )
  const backdropRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [onClose])

  return createPortal(
    <div
      ref={backdropRef}
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
      onMouseDown={(e) => {
        if (e.target === backdropRef.current) onClose()
      }}
    >
      <div className="w-full max-w-sm rounded-xl border border-gray-700 bg-gray-900 p-6 shadow-2xl">
        <h2 className="mb-1 text-base font-semibold text-gray-100">
          Archive card
        </h2>
        <p className="mb-4 text-sm text-gray-400">
          Why is this card being archived?
        </p>

        <div className="mb-5 flex flex-col gap-2">
          {REASONS.map((r) => (
            <label
              key={r.value}
              className="flex cursor-pointer items-start gap-3 rounded-lg border border-gray-700 p-3 has-[:checked]:border-indigo-500"
            >
              <input
                type="radio"
                name="reason"
                value={r.value}
                checked={selected === r.value}
                onChange={() => setSelected(r.value)}
                className="mt-0.5 accent-indigo-500"
              />
              <div>
                <span className="text-sm font-medium text-gray-200">
                  {r.label}
                </span>
                <p className="mt-0.5 text-xs text-gray-500">{r.desc}</p>
              </div>
            </label>
          ))}
        </div>

        <div className="flex gap-2">
          <button
            onClick={() => onConfirm(selected)}
            className="flex-1 rounded-lg bg-amber-600 py-2 text-sm font-medium text-white hover:bg-amber-500"
          >
            Archive
          </button>
          <button
            onClick={onClose}
            className="flex-1 rounded-lg bg-gray-800 py-2 text-sm font-medium text-gray-400 hover:bg-gray-700"
          >
            Cancel
          </button>
        </div>
      </div>
    </div>,
    document.body,
  )
}
