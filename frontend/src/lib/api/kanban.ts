import { apiFetch } from './client'
import type {
  Board,
  BoardStats,
  Card,
  CreateBoardInput,
  CreateCardInput,
  KanbanBoard,
  MoveCardInput,
  UpdateBoardInput,
  UpdateCardInput,
} from './types'

export function listBoards(): Promise<Board[] | undefined> {
  return apiFetch<Board[]>('/kanban/boards', { auth: true })
}

export function createBoard(
  input: CreateBoardInput,
): Promise<Board | undefined> {
  return apiFetch<Board>('/kanban/boards', {
    method: 'POST',
    body: input,
    auth: true,
  })
}

export function updateBoard(
  id: string,
  input: UpdateBoardInput,
): Promise<undefined> {
  return apiFetch<undefined>(`/kanban/boards/${id}`, {
    method: 'PATCH',
    body: input,
    auth: true,
  })
}

export function deleteBoard(id: string): Promise<undefined> {
  return apiFetch<undefined>(`/kanban/boards/${id}`, {
    method: 'DELETE',
    auth: true,
  })
}

export function fetchBoard(boardId: string): Promise<KanbanBoard | undefined> {
  return apiFetch<KanbanBoard>(`/kanban/boards/${boardId}`, { auth: true })
}

export function fetchBoardStats(
  boardId: string,
): Promise<BoardStats | undefined> {
  return apiFetch<BoardStats>(`/kanban/boards/${boardId}/stats`, { auth: true })
}

export function createCard(input: CreateCardInput): Promise<Card | undefined> {
  return apiFetch<Card>('/kanban/cards', {
    method: 'POST',
    body: input,
    auth: true,
  })
}

export function updateCard(
  id: string,
  input: UpdateCardInput,
): Promise<Card | undefined> {
  return apiFetch<Card>(`/kanban/cards/${id}`, {
    method: 'PATCH',
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
