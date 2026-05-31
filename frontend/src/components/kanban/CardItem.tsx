'use client';

import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';

import type { KanbanCard } from '@/types/kanban';

interface Props {
  card: KanbanCard;
  onDelete: (id: string) => Promise<void>;
  /** True when rendered inside DragOverlay (no transform/transition needed). */
  isDragging?: boolean;
}

export function CardItem({ card, onDelete, isDragging }: Props) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging: isSortableDragging } = useSortable({ id: card.id });

  const style: React.CSSProperties = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isSortableDragging ? 0.4 : 1
  };

  return (
    <div ref={setNodeRef} style={isDragging ? undefined : style} className="rounded-lg border border-gray-200 bg-white p-3 shadow-sm">
      <div className="flex items-start gap-2">
        {/* Drag handle — only spreads drag listeners so the delete button stays clickable */}
        <span
          {...attributes}
          {...listeners}
          className="mt-0.5 cursor-grab touch-none select-none text-gray-300 active:cursor-grabbing"
          aria-label="Drag card"
        >
          ⠿
        </span>

        <div className="min-w-0 flex-1">
          <div className="flex items-start justify-between gap-2">
            <h4 className="text-sm font-medium text-gray-900">{card.title}</h4>
            <button
              onClick={(e) => {
                e.stopPropagation();
                onDelete(card.id);
              }}
              className="text-xs text-gray-400 hover:text-red-500"
              title="Delete card"
            >
              &times;
            </button>
          </div>
          {card.content && <p className="mt-1 text-xs text-gray-500">{card.content}</p>}
        </div>
      </div>
    </div>
  );
}
