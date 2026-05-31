'use client';

import { useCallback, useState } from 'react';

import * as kanbanApi from '@/lib/api/kanban';
import type { BoardData, CreateCardInput, MoveCardInput } from '@/types/kanban';

export function useKanban() {
  const [boardData, setBoardData] = useState<BoardData | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchBoard = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await kanbanApi.getBoard();
      setBoardData(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch board');
    } finally {
      setLoading(false);
    }
  }, []);

  const addCard = useCallback(
    async (input: CreateCardInput) => {
      await kanbanApi.createCard(input);
      await fetchBoard();
    },
    [fetchBoard]
  );

  const moveCard = useCallback(
    async (id: string, input: MoveCardInput) => {
      await kanbanApi.moveCard(id, input);
      await fetchBoard();
    },
    [fetchBoard]
  );

  const removeCard = useCallback(
    async (id: string) => {
      await kanbanApi.deleteCard(id);
      await fetchBoard();
    },
    [fetchBoard]
  );

  return { boardData, loading, error, fetchBoard, addCard, moveCard, removeCard };
}
