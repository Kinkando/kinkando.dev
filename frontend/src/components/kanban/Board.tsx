'use client';

import type { BoardData } from '@/types/kanban';
import type { CreateCardInput } from '@/types/kanban';

import { Column } from './Column';

interface Props {
  data: BoardData;
  onAddCard: (input: CreateCardInput) => Promise<void>;
  onDeleteCard: (id: string) => Promise<void>;
}

export function Board({ data, onAddCard, onDeleteCard }: Props) {
  const { columns, cards } = data;

  return (
    <div className="flex gap-4 overflow-x-auto pb-4">
      {columns.map((col) => {
        const colCards = cards.filter((c) => c.column_id === col.id).sort((a, b) => a.order - b.order);

        return <Column key={col.id} column={col} cards={colCards} onAddCard={onAddCard} onDeleteCard={onDeleteCard} />;
      })}
    </div>
  );
}
