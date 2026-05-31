import { apiFetch } from '@/lib/api';
import type { BoardData, CreateCardInput, KanbanCard, MoveCardInput } from '@/types/kanban';

export function getBoard(): Promise<BoardData> {
  return apiFetch('/api/v1/kanban/board');
}

export function createCard(input: CreateCardInput): Promise<KanbanCard> {
  return apiFetch('/api/v1/kanban/cards', {
    method: 'POST',
    body: JSON.stringify(input)
  });
}

export function moveCard(id: string, input: MoveCardInput): Promise<void> {
  return apiFetch(`/api/v1/kanban/cards/${id}/move`, {
    method: 'PATCH',
    body: JSON.stringify(input)
  });
}

export function deleteCard(id: string): Promise<void> {
  return apiFetch(`/api/v1/kanban/cards/${id}`, { method: 'DELETE' });
}
