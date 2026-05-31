'use client';

import type { KanbanCard } from '@/types/kanban';

interface Props {
  card: KanbanCard;
  onDelete: (id: string) => Promise<void>;
}

export function CardItem({ card, onDelete }: Props) {
  return (
    <div className="rounded-lg border border-gray-200 bg-white p-3 shadow-sm">
      <div className="flex items-start justify-between gap-2">
        <h4 className="text-sm font-medium text-gray-900">{card.title}</h4>
        <button onClick={() => onDelete(card.id)} className="text-xs text-gray-400 hover:text-red-500" title="Delete card">
          &times;
        </button>
      </div>
      {card.content && <p className="mt-1 text-xs text-gray-500">{card.content}</p>}
    </div>
  );
}
