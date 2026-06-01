import { createPortal } from 'react-dom'
import { useEffect, useRef, useState } from 'react'
import {
  SortableContext,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import type {
  Column as ColumnType,
  Card as CardType,
} from '../../lib/api/types'
import { useUpdateColumn, useDeleteColumn } from '../../queries/useKanban'
import KanbanCard from './Card'

type Props = {
  column: ColumnType
  otherColumns: ColumnType[]
  cards: CardType[]
  boardId: string
  onAddCard: (columnId: string) => void
  onEditCard: (card: CardType) => void
}

export default function KanbanColumn({
  column,
  otherColumns,
  cards,
  boardId,
  onAddCard,
  onEditCard,
}: Props) {
  const updateColumn = useUpdateColumn(boardId)
  const deleteColumn = useDeleteColumn(boardId)

  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: column.id, data: { type: 'column' } })

  const style: React.CSSProperties = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.4 : 1,
  }

  const [menuOpen, setMenuOpen] = useState(false)
  const [renaming, setRenaming] = useState(false)
  const [nameInput, setNameInput] = useState(column.name)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)

  const menuRef = useRef<HTMLDivElement>(null)
  const renameRef = useRef<HTMLInputElement>(null)

  // Close menu on outside click.
  useEffect(() => {
    function onClickOutside(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuOpen(false)
      }
    }
    if (menuOpen) document.addEventListener('mousedown', onClickOutside)
    return () => document.removeEventListener('mousedown', onClickOutside)
  }, [menuOpen])

  // Focus rename input when entering rename mode.
  useEffect(() => {
    if (renaming) renameRef.current?.select()
  }, [renaming])

  function startRename() {
    setMenuOpen(false)
    setNameInput(column.name)
    setRenaming(true)
  }

  function commitRename() {
    const trimmed = nameInput.trim()
    if (trimmed && trimmed !== column.name) {
      updateColumn.mutate({ id: column.id, input: { name: trimmed } })
    }
    setRenaming(false)
  }

  function handleRenameKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Enter') commitRename()
    if (e.key === 'Escape') setRenaming(false)
  }

  const isDone = column.type === 'done'

  return (
    <>
      <div
        ref={setNodeRef}
        style={style}
        className="flex w-72 shrink-0 flex-col gap-3 rounded-xl border bg-gray-900 p-4"
      >
        {/* Column header */}
        <div className="flex items-center gap-2">
          {/* Drag handle */}
          <button
            {...attributes}
            {...listeners}
            className="flex-shrink-0 cursor-grab touch-none text-sm text-gray-600 select-none hover:text-gray-400"
            title="Drag to reorder column"
          >
            ⠿
          </button>

          {/* Name / rename input */}
          {renaming ? (
            <input
              ref={renameRef}
              value={nameInput}
              onChange={(e) => setNameInput(e.target.value)}
              onBlur={commitRename}
              onKeyDown={handleRenameKeyDown}
              className="flex-1 rounded bg-gray-800 px-2 py-0.5 text-sm font-semibold text-gray-100 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
            />
          ) : (
            <h3
              className="flex-1 truncate text-sm font-semibold text-gray-300"
              onDoubleClick={startRename}
              title="Double-click to rename"
            >
              {column.name}
              {isDone && (
                <span className="ml-1.5 text-xs font-normal text-emerald-500">
                  ✓
                </span>
              )}
            </h3>
          )}

          {/* Card count */}
          <span className="flex-shrink-0 text-xs text-gray-600">
            {cards.length}
          </span>

          {/* Column menu */}
          <div className="relative flex-shrink-0" ref={menuRef}>
            <button
              onClick={() => setMenuOpen((o) => !o)}
              className="text-gray-600 hover:text-gray-300"
              title="Column options"
            >
              ⋯
            </button>
            {menuOpen && (
              <div className="absolute top-6 right-0 z-20 min-w-32 rounded-lg border border-gray-700 bg-gray-900 py-1 shadow-xl">
                <button
                  className="w-full px-3 py-1.5 text-left text-sm text-gray-300 hover:bg-gray-800"
                  onClick={() => onAddCard(column.id)}
                >
                  Add card
                </button>
                <button
                  className="w-full px-3 py-1.5 text-left text-sm text-gray-300 hover:bg-gray-800"
                  onClick={startRename}
                >
                  Rename
                </button>
                {!column.is_system && (
                  <button
                    className="w-full px-3 py-1.5 text-left text-sm text-red-400 hover:bg-gray-800"
                    onClick={() => {
                      setMenuOpen(false)
                      setShowDeleteDialog(true)
                    }}
                  >
                    Delete column
                  </button>
                )}
              </div>
            )}
          </div>
        </div>

        {/* Cards */}
        <div className="flex min-h-8 flex-col gap-2">
          <SortableContext
            items={cards.map((c) => c.id)}
            strategy={verticalListSortingStrategy}
          >
            {cards.map((card) => (
              <KanbanCard
                key={card.id}
                card={card}
                boardId={boardId}
                columnType={column.type}
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

      {showDeleteDialog && (
        <DeleteColumnDialog
          column={column}
          otherColumns={otherColumns}
          onConfirm={(action, targetColumnId) => {
            deleteColumn.mutate(
              {
                id: column.id,
                input: { action, target_column_id: targetColumnId },
              },
              { onSettled: () => setShowDeleteDialog(false) },
            )
          }}
          onClose={() => setShowDeleteDialog(false)}
        />
      )}
    </>
  )
}

// ---- Delete column dialog --------------------------------------------------

type DeleteColumnDialogProps = {
  column: ColumnType
  otherColumns: ColumnType[]
  onConfirm: (action: 'move' | 'archive', targetColumnId?: string) => void
  onClose: () => void
}

function DeleteColumnDialog({
  column,
  otherColumns,
  onConfirm,
  onClose,
}: DeleteColumnDialogProps) {
  const [action, setAction] = useState<'move' | 'archive'>('move')
  const [targetColumnId, setTargetColumnId] = useState(
    otherColumns[0]?.id ?? '',
  )

  const backdropRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [onClose])

  function handleConfirm() {
    if (action === 'move' && !targetColumnId) return
    onConfirm(action, action === 'move' ? targetColumnId : undefined)
  }

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
          Delete &ldquo;{column.name}&rdquo;
        </h2>
        <p className="mb-5 text-sm text-gray-400">
          This column has {/* card count shown in parent */} cards. What should
          happen to them?
        </p>

        <div className="mb-4 flex flex-col gap-3">
          {/* Move option */}
          <label className="flex cursor-pointer items-start gap-3 rounded-lg border border-gray-700 p-3 has-[:checked]:border-indigo-500">
            <input
              type="radio"
              name="action"
              value="move"
              checked={action === 'move'}
              onChange={() => setAction('move')}
              className="mt-0.5 accent-indigo-500"
            />
            <div className="flex-1">
              <span className="text-sm font-medium text-gray-200">
                Move cards to another column
              </span>
              {action === 'move' && (
                <select
                  value={targetColumnId}
                  onChange={(e) => setTargetColumnId(e.target.value)}
                  className="mt-2 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-1.5 text-sm text-gray-100 focus:border-indigo-500 focus:outline-none"
                >
                  {otherColumns.map((col) => (
                    <option key={col.id} value={col.id}>
                      {col.name}
                    </option>
                  ))}
                </select>
              )}
            </div>
          </label>

          {/* Archive option */}
          <label className="flex cursor-pointer items-start gap-3 rounded-lg border border-gray-700 p-3 has-[:checked]:border-indigo-500">
            <input
              type="radio"
              name="action"
              value="archive"
              checked={action === 'archive'}
              onChange={() => setAction('archive')}
              className="mt-0.5 accent-indigo-500"
            />
            <div>
              <span className="text-sm font-medium text-gray-200">
                Archive all cards
              </span>
              <p className="mt-0.5 text-xs text-gray-500">
                Cards will be marked as stale and moved to the archive.
              </p>
            </div>
          </label>
        </div>

        <div className="flex gap-2">
          <button
            onClick={handleConfirm}
            disabled={action === 'move' && !targetColumnId}
            className="flex-1 rounded-lg bg-red-600 py-2 text-sm font-medium text-white hover:bg-red-500 disabled:opacity-50"
          >
            Delete Column
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
