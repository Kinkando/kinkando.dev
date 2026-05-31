import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  listBoards,
  createBoard,
  updateBoard,
  deleteBoard,
  fetchBoard,
  fetchBoardStats,
  createCard,
  updateCard,
  moveCard,
  deleteCard,
} from '../lib/api/kanban'
import type {
  Board,
  CreateBoardInput,
  UpdateBoardInput,
  CreateCardInput,
  UpdateCardInput,
  MoveCardInput,
  KanbanBoard,
} from '../lib/api/types'
import { keys } from './keys'

export function useBoards() {
  return useQuery({
    queryKey: keys.kanbanBoards,
    queryFn: listBoards,
  })
}

export function useBoard(boardId: string) {
  return useQuery({
    queryKey: keys.kanbanBoard(boardId),
    queryFn: () => fetchBoard(boardId),
    enabled: !!boardId,
  })
}

export function useBoardStats(boardId: string) {
  return useQuery({
    queryKey: keys.kanbanStats(boardId),
    queryFn: () => fetchBoardStats(boardId),
    enabled: !!boardId,
  })
}

export function useCreateBoard() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateBoardInput) => createBoard(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.kanbanBoards })
    },
  })
}

export function useUpdateBoard() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateBoardInput }) =>
      updateBoard(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.kanbanBoards })
    },
  })
}

export function useDeleteBoard() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deleteBoard(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.kanbanBoards })
    },
  })
}

export function useCreateCard(boardId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateCardInput) => createCard(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.kanbanBoard(boardId) })
      queryClient.invalidateQueries({ queryKey: keys.kanbanStats(boardId) })
    },
  })
}

export function useUpdateCard(boardId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateCardInput }) =>
      updateCard(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.kanbanBoard(boardId) })
      queryClient.invalidateQueries({ queryKey: keys.kanbanStats(boardId) })
    },
  })
}

export function useMoveCard(boardId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: MoveCardInput }) =>
      moveCard(id, input),
    onMutate: async ({ id, input }) => {
      const key = keys.kanbanBoard(boardId)
      await queryClient.cancelQueries({ queryKey: key })
      const prev = queryClient.getQueryData<KanbanBoard>(key)
      if (prev) {
        queryClient.setQueryData<KanbanBoard>(key, (old) => {
          if (!old) return old
          const moving = old.cards.find((c) => c.id === id)
          if (!moving) return old
          const oldCol = moving.column_id
          const newCol = input.column_id
          const oldOrder = moving.order
          const newOrder = input.order
          const cards = old.cards.map((c) => {
            if (c.id === id) return { ...c, column_id: newCol, order: newOrder }
            if (oldCol === newCol) {
              if (
                oldOrder < newOrder &&
                c.column_id === oldCol &&
                c.order > oldOrder &&
                c.order <= newOrder
              )
                return { ...c, order: c.order - 1 }
              if (
                oldOrder > newOrder &&
                c.column_id === oldCol &&
                c.order >= newOrder &&
                c.order < oldOrder
              )
                return { ...c, order: c.order + 1 }
            } else {
              if (c.column_id === oldCol && c.order > oldOrder)
                return { ...c, order: c.order - 1 }
              if (c.column_id === newCol && c.order >= newOrder)
                return { ...c, order: c.order + 1 }
            }
            return c
          })
          return { ...old, cards }
        })
      }
      return { prev }
    },
    onError: (_err, _vars, ctx) => {
      if (ctx?.prev)
        queryClient.setQueryData(keys.kanbanBoard(boardId), ctx.prev)
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: keys.kanbanBoard(boardId) })
    },
  })
}

export function useDeleteCard(boardId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deleteCard(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.kanbanBoard(boardId) })
      queryClient.invalidateQueries({ queryKey: keys.kanbanStats(boardId) })
    },
  })
}
