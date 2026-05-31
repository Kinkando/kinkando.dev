'use client';

import type { DragEndEvent, DragStartEvent } from '@dnd-kit/core';
import { DndContext, DragOverlay, PointerSensor, useSensor, useSensors } from '@dnd-kit/core';
import { useState } from 'react';

import type { BoardData, CreateCardInput, KanbanCard, MoveCardInput } from '@/types/kanban';

import { CardItem } from './CardItem';
import { Column } from './Column';

interface Props {
  data: BoardData;
  onAddCard: (input: CreateCardInput) => Promise<void>;
  onMoveCard: (id: string, input: MoveCardInput) => Promise<void>;
  onDeleteCard: (id: string) => Promise<void>;
}

export function Board({ data, onAddCard, onMoveCard, onDeleteCard }: Props) {
  const { columns, cards } = data;
  const [activeCard, setActiveCard] = useState<KanbanCard | null>(null);

  // Require a small drag distance before activating so click handlers still fire.
  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 8 } }));

  function handleDragStart({ active }: DragStartEvent) {
    const card = cards.find((c) => c.id === active.id);
    setActiveCard(card ?? null);
  }

  function handleDragEnd({ active, over }: DragEndEvent) {
    setActiveCard(null);
    if (!over || active.id === over.id) return;

    const activeId = String(active.id);
    const overId = String(over.id);

    // Determine which column the card was dropped onto.
    // `over` can be another card or an empty column droppable.
    const overCard = cards.find((c) => c.id === overId);
    const destColumnId = overCard ? overCard.column_id : overId;

    // Compute destination order: insert after the target card (or append to empty column).
    const destCards = cards.filter((c) => c.column_id === destColumnId && c.id !== activeId).sort((a, b) => a.order - b.order);
    const overIndex = overCard ? destCards.findIndex((c) => c.id === overId) : -1;
    const destOrder = overIndex >= 0 ? overIndex + 1 : destCards.length;

    onMoveCard(activeId, { column_id: destColumnId, order: destOrder });
  }

  return (
    <DndContext sensors={sensors} onDragStart={handleDragStart} onDragEnd={handleDragEnd}>
      <div className="flex gap-4 overflow-x-auto pb-4">
        {columns.map((col) => {
          const colCards = cards.filter((c) => c.column_id === col.id).sort((a, b) => a.order - b.order);
          return <Column key={col.id} column={col} cards={colCards} onAddCard={onAddCard} onDeleteCard={onDeleteCard} />;
        })}
      </div>

      {/* Ghost card shown under the cursor while dragging */}
      <DragOverlay>{activeCard ? <CardItem card={activeCard} onDelete={onDeleteCard} isDragging /> : null}</DragOverlay>
    </DndContext>
  );
}
