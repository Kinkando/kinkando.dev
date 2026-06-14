import {
  deleteObject,
  getDownloadURL,
  ref as storageRef,
  uploadBytes,
} from 'firebase/storage'
import { apiFetch } from './client'
import { auth, storage } from '../firebase'
import type {
  AddAttachmentInput,
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

// uploadAndAttachFile pushes a file to Firebase Storage under the signed-in
// user's namespace, then registers the resulting URL on the card.
export async function uploadAndAttachFile(
  cardId: string,
  file: File,
): Promise<Attachment | undefined> {
  const uid = auth.currentUser?.uid
  if (!uid) throw new Error('not signed in')
  // Path includes uid + cardId so Storage rules can scope writes per user/card.
  const path = `kanban/${uid}/${cardId}/${Date.now()}_${file.name}`
  const ref = storageRef(storage, path)
  await uploadBytes(ref, file, { contentType: file.type || undefined })
  const url = await getDownloadURL(ref)
  const input: AddAttachmentInput = {
    name: file.name,
    url,
    storage_path: path,
    size: file.size,
    content_type: file.type || 'application/octet-stream',
  }
  return apiFetch<Attachment>(`/kanban/cards/${cardId}/attachments`, {
    method: 'POST',
    body: input,
    auth: true,
  })
}

// removeAttachment unregisters the metadata on the card, then best-effort
// deletes the underlying object from Firebase Storage.
export async function removeAttachment(
  cardId: string,
  attachmentId: string,
): Promise<void> {
  const removed = await apiFetch<Attachment>(
    `/kanban/cards/${cardId}/attachments/${attachmentId}`,
    { method: 'DELETE', auth: true },
  )
  if (removed?.storage_path) {
    try {
      await deleteObject(storageRef(storage, removed.storage_path))
    } catch {
      // The storage object may already be gone; the DB row is authoritative.
    }
  }
}
