'use client';

import { useDroppable } from '@dnd-kit/core';
import { SortableContext, verticalListSortingStrategy } from '@dnd-kit/sortable';

import type { Column as ColumnType, CreateCardInput, KanbanCard } from '@/types/kanban';

import { AddCardForm } from './AddCardForm';
import { CardItem } from './CardItem';

interface Props {
  column: ColumnType;
  cards: KanbanCard[];
  onAddCard: (input: CreateCardInput) => Promise<void>;
  onDeleteCard: (id: string) => Promise<void>;
}

export function Column({ column, cards, onAddCard, onDeleteCard }: Props) {
  // Register this column as a drop target so empty columns accept dragged cards.
  const { setNodeRef, isOver } = useDroppable({ id: column.id });

  return (
    <div className="flex w-72 flex-shrink-0 flex-col rounded-lg bg-gray-50 p-3">
      <h3 className="mb-3 text-sm font-semibold text-gray-700">
        {column.name} <span className="font-normal text-gray-400">({cards.length})</span>
      </h3>

      <SortableContext items={cards.map((c) => c.id)} strategy={verticalListSortingStrategy}>
        <div ref={setNodeRef} className={`flex-1 space-y-2 rounded-md transition-colors ${isOver ? 'bg-blue-50' : ''}`}>
          {cards.map((card) => (
            <CardItem key={card.id} card={card} onDelete={onDeleteCard} />
          ))}
        </div>
      </SortableContext>

      <div className="mt-3">
        <AddCardForm columnId={column.id} onAdd={onAddCard} />
      </div>
    </div>
  );
}
