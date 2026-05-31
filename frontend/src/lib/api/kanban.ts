import { apiFetch } from './client';
import type { BoardResponse, CreateCardInput, KanbanCard, MoveCardInput } from './types';

export function getBoard(): Promise<BoardResponse> {
  return apiFetch<BoardResponse>('/kanban/board', { auth: true });
}

export function createCard(input: CreateCardInput): Promise<KanbanCard> {
  return apiFetch<KanbanCard>('/kanban/cards', { method: 'POST', auth: true, body: input });
}

export function moveCard(id: string, input: MoveCardInput): Promise<void> {
  return apiFetch<void>(`/kanban/cards/${id}/move`, { method: 'PATCH', auth: true, body: input });
}

export function deleteCard(id: string): Promise<void> {
  return apiFetch<void>(`/kanban/cards/${id}`, { method: 'DELETE', auth: true });
}
