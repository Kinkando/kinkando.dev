import { ApiError, apiFetch } from './client'
import env from '../../config/env'
import { getIdToken } from '../firebase'
import type {
  ArchiveCardInput,
  ArchiveFilter,
  Attachment,
  Board,
  BoardStats,
  Card,
  Column,
  CreateBoardInput,
  CreateCardInput,
  CreateColumnInput,
  DeleteColumnInput,
  KanbanBoard,
  MoveCardInput,
  ReorderColumnsInput,
  UpdateBoardInput,
  UpdateCardInput,
  UpdateColumnInput,
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

export function fetchArchive(
  boardId: string,
  filter: ArchiveFilter = {},
): Promise<Card[] | undefined> {
  const params: Record<string, string> = {}
  if (filter.reason) params.reason = filter.reason
  if (filter.month) params.month = String(filter.month)
  if (filter.year) params.year = String(filter.year)
  return apiFetch<Card[]>(`/kanban/boards/${boardId}/archive`, {
    auth: true,
    query: Object.keys(params).length > 0 ? params : undefined,
  })
}

export function createColumn(
  input: CreateColumnInput,
): Promise<Column | undefined> {
  return apiFetch<Column>('/kanban/columns', {
    method: 'POST',
    body: input,
    auth: true,
  })
}

export function updateColumn(
  id: string,
  input: UpdateColumnInput,
): Promise<undefined> {
  return apiFetch<undefined>(`/kanban/columns/${id}`, {
    method: 'PATCH',
    body: input,
    auth: true,
  })
}

export function reorderColumns(
  boardId: string,
  input: ReorderColumnsInput,
): Promise<undefined> {
  return apiFetch<undefined>(`/kanban/boards/${boardId}/columns/reorder`, {
    method: 'PATCH',
    body: input,
    auth: true,
  })
}

export function deleteColumn(
  id: string,
  input: DeleteColumnInput,
): Promise<undefined> {
  return apiFetch<undefined>(`/kanban/columns/${id}`, {
    method: 'DELETE',
    body: input,
    auth: true,
  })
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

export function archiveCard(
  id: string,
  input: ArchiveCardInput,
): Promise<Card | undefined> {
  return apiFetch<Card>(`/kanban/cards/${id}/archive`, {
    method: 'PATCH',
    body: input,
    auth: true,
  })
}

export function unarchiveCard(id: string): Promise<Card | undefined> {
  return apiFetch<Card>(`/kanban/cards/${id}/unarchive`, {
    method: 'PATCH',
    auth: true,
  })
}

// uploadAndAttachFile streams the file to the backend as a multipart form;
// the backend pushes the bytes to Supabase Storage and persists the metadata.
// We bypass apiFetch here because it forces a JSON Content-Type.
export async function uploadAndAttachFile(
  cardId: string,
  file: File,
): Promise<Attachment | undefined> {
  const token = await getIdToken()
  if (!token) throw new Error('not signed in')

  const form = new FormData()
  form.append('file', file)

  const res = await fetch(
    `${env.apiUrl}/api/v1/kanban/cards/${cardId}/attachments`,
    {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}` },
      body: form,
    },
  )

  if (res.status === 204) return undefined
  const json = await res.json()
  if (!res.ok) {
    throw new ApiError(
      res.status,
      (json as { error?: string }).error ?? `HTTP ${res.status}`,
    )
  }
  return (json as { data: Attachment }).data
}

// removeAttachment deletes the metadata + the underlying object server-side.
export function removeAttachment(
  cardId: string,
  attachmentId: string,
): Promise<undefined> {
  return apiFetch<undefined>(
    `/kanban/cards/${cardId}/attachments/${attachmentId}`,
    { method: 'DELETE', auth: true },
  )
}
