'use client';

import { useEffect } from 'react';

import { Board } from '@/components/kanban/Board';
import { Spinner } from '@/components/ui/Spinner';
import { useKanban } from '@/hooks/useKanban';

export default function KanbanPage() {
  const { boardData, loading, error, fetchBoard, addCard, removeCard } = useKanban();

  useEffect(() => {
    fetchBoard();
  }, [fetchBoard]);

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Kanban Board</h1>
      </div>

      {error && <p className="mb-4 rounded bg-red-50 p-3 text-sm text-red-600">{error}</p>}

      {loading ? (
        <div className="flex justify-center py-12">
          <Spinner className="h-8 w-8" />
        </div>
      ) : boardData ? (
        <Board data={boardData} onAddCard={addCard} onDeleteCard={removeCard} />
      ) : null}
    </div>
  );
}
