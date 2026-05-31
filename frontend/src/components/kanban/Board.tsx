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
import KanbanColumn from './Column'
import KanbanCard from './Card'

type Props = {
  data: KanbanBoardType
}

export default function KanbanBoard({ data }: Props) {
  const moveCard = useMoveCard()
  const [activeCard, setActiveCard] = useState<CardType | null>(null)
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 8 } }),
  )

  const columns = [...data.columns].sort((a, b) => a.order - b.order)

  function getCardsForColumn(columnId: string) {
    return data.cards
      .filter((c) => c.column_id === columnId)
      .sort((a, b) => a.order - b.order)
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
          />
        ))}
      </div>
      <DragOverlay>
        {activeCard ? <KanbanCard card={activeCard} /> : null}
      </DragOverlay>
    </DndContext>
  )
}
