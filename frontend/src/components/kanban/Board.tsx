import { useState } from 'react'
import {
  DndContext,
  DragOverlay,
  PointerSensor,
  useSensor,
  useSensors,
  type DragStartEvent,
  type DragEndEvent,
} from '@dnd-kit/core'
import { arrayMove } from '@dnd-kit/sortable'
import type {
  KanbanBoard as KanbanBoardType,
  Card as CardType,
} from '../../lib/api/types'
import { useMoveCard } from '../../queries/useKanban'
import { isOverdue, isDueSoon } from '../../lib/kanban'
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
  const [activeCard, setActiveCard] = useState<CardType | null>(null)
  const [modalState, setModalState] = useState<CardModalState>(null)
  const [filter, setFilter] = useState<FilterState>(DEFAULT_FILTER)

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 8 } }),
  )

  const columns = [...data.columns].sort((a, b) => a.order - b.order)

  // Collect all unique tags across the board for the filter UI
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
    setActiveCard(data.cards.find((c) => c.id === active.id) ?? null)
  }

  function onDragEnd({ active, over }: DragEndEvent) {
    setActiveCard(null)
    if (!over || active.id === over.id) return

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

  return (
    <>
      <StatsBar boardId={boardId} />
      <FilterBar filter={filter} onChange={setFilter} allTags={allTags} />

      <DndContext
        sensors={sensors}
        onDragStart={onDragStart}
        onDragEnd={onDragEnd}
      >
        <div className="flex gap-5 overflow-x-auto pb-4">
          {columns.map((column) => (
            <KanbanColumn
              key={column.id}
              column={column}
              cards={getCardsForColumn(column.id)}
              boardId={boardId}
              onAddCard={(colId) =>
                setModalState({ mode: 'create', columnId: colId })
              }
              onEditCard={(card) => setModalState({ mode: 'edit', card })}
            />
          ))}
        </div>
        <DragOverlay dropAnimation={null}>
          {activeCard ? (
            <KanbanCard card={activeCard} boardId={boardId} onEdit={() => {}} />
          ) : null}
        </DragOverlay>
      </DndContext>

      {modalState && (
        <CardModal
          boardId={boardId}
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
