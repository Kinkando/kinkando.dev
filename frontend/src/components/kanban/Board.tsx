import { useRef, useState } from 'react'
import {
  DndContext,
  DragOverlay,
  PointerSensor,
  useSensor,
  useSensors,
  type DragStartEvent,
  type DragEndEvent,
} from '@dnd-kit/core'
import {
  SortableContext,
  arrayMove,
  horizontalListSortingStrategy,
} from '@dnd-kit/sortable'
import type {
  Column as ColumnType,
  KanbanBoard as KanbanBoardType,
  Card as CardType,
} from '../../lib/api/types'
import {
  useMoveCard,
  useReorderColumns,
  useCreateColumn,
} from '../../queries/useKanban'
import { isOverdue, isDueSoon } from '../../lib/kanban'
import { useIsMobile } from '../../hooks/useIsMobile'
import KanbanColumn from './Column'
import KanbanCard from './Card'
import CardModal from './CardModal'
import FilterBar, { type FilterState, DEFAULT_FILTER } from './FilterBar'
import StatsBar from './StatsBar'

type CardModalState =
  | { mode: 'create'; columnId: string }
  | { mode: 'edit'; card: CardType }
  | null

type Props = {
  boardId: string
  data: KanbanBoardType
}

export default function KanbanBoard({ boardId, data }: Props) {
  const moveCard = useMoveCard(boardId)
  const reorderColumns = useReorderColumns(boardId)
  const createColumn = useCreateColumn(boardId)
  const isMobile = useIsMobile()

  const [activeCard, setActiveCard] = useState<CardType | null>(null)
  const [activeColumn, setActiveColumn] = useState<ColumnType | null>(null)
  const [modalState, setModalState] = useState<CardModalState>(null)
  const [filter, setFilter] = useState<FilterState>(DEFAULT_FILTER)
  const [addingColumn, setAddingColumn] = useState(false)
  const [newColumnName, setNewColumnName] = useState('')
  const addColumnInputRef = useRef<HTMLInputElement>(null)

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 8 } }),
  )

  const columns = [...data.columns].sort((a, b) => a.order - b.order)

  // Collect all unique tags across the board for the filter UI.
  const allTags = Array.from(
    new Set(data.cards.flatMap((c) => c.tags ?? [])),
  ).sort()

  function applyFilter(cards: CardType[]): CardType[] {
    return cards.filter((c) => {
      if (
        filter.search &&
        !c.title.toLowerCase().includes(filter.search.toLowerCase()) &&
        !c.description?.toLowerCase().includes(filter.search.toLowerCase())
      )
        return false

      if (
        filter.priorities.length > 0 &&
        !filter.priorities.includes(c.priority)
      )
        return false

      if (
        filter.tags.length > 0 &&
        !filter.tags.some((t) => c.tags?.includes(t))
      )
        return false

      if (filter.dueStatus === 'overdue' && !isOverdue(c)) return false
      if (filter.dueStatus === 'due-soon' && !isDueSoon(c)) return false
      if (filter.dueStatus === 'no-date' && !!c.due_date) return false

      return true
    })
  }

  function getCardsForColumn(columnId: string) {
    const col = data.cards
      .filter((c) => c.column_id === columnId)
      .sort((a, b) => a.order - b.order)
    return applyFilter(col)
  }

  function onDragStart({ active }: DragStartEvent) {
    if (active.data.current?.type === 'column') {
      setActiveColumn(columns.find((c) => c.id === active.id) ?? null)
    } else {
      setActiveCard(data.cards.find((c) => c.id === active.id) ?? null)
    }
  }

  function onDragEnd({ active, over }: DragEndEvent) {
    setActiveCard(null)
    setActiveColumn(null)

    if (!over || active.id === over.id) return

    // Column reorder
    if (active.data.current?.type === 'column') {
      const targetCol = columns.find((c) => c.id === String(over.id))
      if (!targetCol) return
      const oldIdx = columns.findIndex((c) => c.id === String(active.id))
      const newIdx = columns.findIndex((c) => c.id === String(over.id))
      if (oldIdx === -1 || newIdx === -1 || oldIdx === newIdx) return
      const reordered = arrayMove(columns, oldIdx, newIdx)
      reorderColumns.mutate({ column_ids: reordered.map((c) => c.id) })
      return
    }

    // Card move (existing logic)
    const card = data.cards.find((c) => c.id === active.id)
    if (!card) return

    const overCard = data.cards.find((c) => c.id === over.id)
    const targetColumnId = overCard ? overCard.column_id : String(over.id)
    if (!columns.some((col) => col.id === targetColumnId)) return

    const columnCards = data.cards
      .filter((c) => c.column_id === targetColumnId)
      .sort((a, b) => a.order - b.order)

    let newOrder: number
    if (overCard) {
      const sameColumn = card.column_id === targetColumnId
      if (sameColumn) {
        const oldIdx = columnCards.findIndex((c) => c.id === card.id)
        const newIdx = columnCards.findIndex((c) => c.id === overCard.id)
        if (oldIdx === newIdx) return
        const reordered = arrayMove(columnCards, oldIdx, newIdx)
        newOrder = reordered.findIndex((c) => c.id === card.id)
      } else {
        const insertIdx = columnCards.findIndex((c) => c.id === overCard.id)
        newOrder = insertIdx === -1 ? columnCards.length : insertIdx
      }
    } else {
      if (card.column_id === targetColumnId) return
      newOrder = columnCards.length
    }

    moveCard.mutate({
      id: card.id,
      input: { column_id: targetColumnId, order: newOrder },
    })
  }

  function handleAddColumn() {
    const name = newColumnName.trim()
    if (!name) {
      setAddingColumn(false)
      return
    }
    createColumn.mutate(
      { board_id: boardId, name },
      {
        onSettled: () => {
          setAddingColumn(false)
          setNewColumnName('')
        },
      },
    )
  }

  function handleAddColumnKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Enter') handleAddColumn()
    if (e.key === 'Escape') {
      setAddingColumn(false)
      setNewColumnName('')
    }
  }

  // Find column type for a given card (used in drag overlay).
  function cardColumnType(card: CardType) {
    return columns.find((c) => c.id === card.column_id)?.type ?? 'custom'
  }

  return (
    <>
      <StatsBar boardId={boardId} />
      <FilterBar filter={filter} onChange={setFilter} allTags={allTags} />

      <DndContext
        sensors={isMobile ? [] : sensors}
        onDragStart={onDragStart}
        onDragEnd={onDragEnd}
      >
        <div className="flex gap-5 overflow-x-auto pb-4">
          <SortableContext
            items={columns.map((c) => c.id)}
            strategy={horizontalListSortingStrategy}
          >
            {columns.map((column) => (
              <KanbanColumn
                key={column.id}
                column={column}
                otherColumns={columns.filter((c) => c.id !== column.id)}
                cards={getCardsForColumn(column.id)}
                boardId={boardId}
                onAddCard={(colId) =>
                  setModalState({ mode: 'create', columnId: colId })
                }
                onEditCard={(card) => setModalState({ mode: 'edit', card })}
              />
            ))}
          </SortableContext>

          {/* Add column */}
          {addingColumn ? (
            <div className="flex w-72 shrink-0 flex-col gap-2 rounded-xl border border-dashed border-gray-700 bg-gray-900 p-4">
              <input
                ref={addColumnInputRef}
                autoFocus
                value={newColumnName}
                onChange={(e) => setNewColumnName(e.target.value)}
                onKeyDown={handleAddColumnKeyDown}
                onBlur={handleAddColumn}
                placeholder="Column name"
                className="w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none"
              />
              <div className="flex gap-2">
                <button
                  onMouseDown={(e) => e.preventDefault()}
                  onClick={handleAddColumn}
                  className="flex-1 rounded-lg bg-indigo-600 py-1.5 text-sm font-medium text-white hover:bg-indigo-500"
                >
                  Add
                </button>
                <button
                  onMouseDown={(e) => e.preventDefault()}
                  onClick={() => {
                    setAddingColumn(false)
                    setNewColumnName('')
                  }}
                  className="flex-1 rounded-lg bg-gray-800 py-1.5 text-sm font-medium text-gray-400 hover:bg-gray-700"
                >
                  Cancel
                </button>
              </div>
            </div>
          ) : (
            <button
              onClick={() => setAddingColumn(true)}
              className="flex h-fit w-72 shrink-0 items-center gap-2 rounded-xl border border-dashed border-gray-700 px-4 py-3 text-sm text-gray-500 hover:border-gray-600 hover:text-gray-400"
            >
              + Add column
            </button>
          )}
        </div>

        <DragOverlay dropAnimation={null}>
          {activeColumn ? (
            <div className="w-72 rounded-xl border border-indigo-500 bg-gray-900 p-4 opacity-90 shadow-2xl">
              <p className="text-sm font-semibold text-gray-300">
                {activeColumn.name}
              </p>
            </div>
          ) : activeCard ? (
            <KanbanCard
              card={activeCard}
              boardId={boardId}
              columnType={cardColumnType(activeCard)}
              onEdit={() => {}}
            />
          ) : null}
        </DragOverlay>
      </DndContext>

      {modalState && (
        <CardModal
          boardId={boardId}
          columns={columns}
          cards={data.cards}
          columnId={
            modalState.mode === 'create' ? modalState.columnId : undefined
          }
          initial={modalState.mode === 'edit' ? modalState.card : undefined}
          onClose={() => setModalState(null)}
        />
      )}
    </>
  )
}
