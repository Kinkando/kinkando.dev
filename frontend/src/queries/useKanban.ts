import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { fetchBoard, createCard, moveCard, deleteCard } from '../lib/api/kanban'
import type {
  CreateCardInput,
  MoveCardInput,
  KanbanBoard,
} from '../lib/api/types'
import { keys } from './keys'

export function useBoard() {
  return useQuery({
    queryKey: keys.kanbanBoard,
    queryFn: fetchBoard,
  })
}

export function useCreateCard() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateCardInput) => createCard(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.kanbanBoard })
    },
  })
}

export function useMoveCard() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: MoveCardInput }) =>
      moveCard(id, input),
    onMutate: async ({ id, input }) => {
      await queryClient.cancelQueries({ queryKey: keys.kanbanBoard })
      const prev = queryClient.getQueryData<KanbanBoard>(keys.kanbanBoard)
      if (prev) {
        queryClient.setQueryData<KanbanBoard>(keys.kanbanBoard, (old) => {
          if (!old) return old
          const cards = old.cards.map((c) =>
            c.id === id
              ? { ...c, column_id: input.column_id, order: input.order }
              : c,
          )
          return { ...old, cards }
        })
      }
      return { prev }
    },
    onError: (_err, _vars, ctx) => {
      if (ctx?.prev) queryClient.setQueryData(keys.kanbanBoard, ctx.prev)
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: keys.kanbanBoard })
    },
  })
}

export function useDeleteCard() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deleteCard(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.kanbanBoard })
    },
  })
}
