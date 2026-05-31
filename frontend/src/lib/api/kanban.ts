import { apiFetch } from './client'
import type { KanbanBoard, Card, CreateCardInput, MoveCardInput } from './types'

export function fetchBoard(): Promise<KanbanBoard | undefined> {
  return apiFetch<KanbanBoard>('/kanban/board', { auth: true })
}

export function createCard(input: CreateCardInput): Promise<Card | undefined> {
  return apiFetch<Card>('/kanban/cards', {
    method: 'POST',
    body: input,
    auth: true,
  })
}

export function moveCard(id: string, input: MoveCardInput): Promise<undefined> {
  return apiFetch<undefined>(`/kanban/cards/${id}/move`, {
    method: 'PATCH',
    body: input,
    auth: true,
  })
}

export function deleteCard(id: string): Promise<undefined> {
  return apiFetch<undefined>(`/kanban/cards/${id}`, {
    method: 'DELETE',
    auth: true,
  })
}
