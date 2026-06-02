import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  fetchQuestOverview,
  fetchQuests,
  createQuest,
  updateQuest,
  deleteQuest,
  completeDaily,
  uncompleteDaily,
  incrementWeekly,
  decrementWeekly,
  fetchQuestHistory,
} from '../lib/api/quest'
import type { CreateQuestInput, UpdateQuestInput } from '../lib/api/types'
import { keys } from './keys'

export function useQuestOverview() {
  return useQuery({
    queryKey: keys.questOverview,
    queryFn: fetchQuestOverview,
  })
}

export function useQuests(type: string) {
  return useQuery({
    queryKey: keys.questList(type),
    queryFn: () => fetchQuests(type),
  })
}

export function useQuestHistory(limit = 50) {
  return useQuery({
    queryKey: keys.questHistory,
    queryFn: () => fetchQuestHistory(limit),
  })
}

// ── CRUD mutations ────────────────────────────────────────────────────────────

export function useCreateQuest() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateQuestInput) => createQuest(input),
    onSuccess: (_data, input) => {
      queryClient.invalidateQueries({ queryKey: keys.questOverview })
      queryClient.invalidateQueries({ queryKey: keys.questList(input.type) })
    },
  })
}

export function useUpdateQuest() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateQuestInput }) =>
      updateQuest(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.questOverview })
      queryClient.invalidateQueries({ queryKey: keys.questList('daily') })
      queryClient.invalidateQueries({ queryKey: keys.questList('weekly') })
    },
  })
}

export function useDeleteQuest() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deleteQuest(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.questOverview })
      queryClient.invalidateQueries({ queryKey: keys.questList('daily') })
      queryClient.invalidateQueries({ queryKey: keys.questList('weekly') })
    },
  })
}

// ── Action mutations ──────────────────────────────────────────────────────────

export function useCompleteDaily() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => completeDaily(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.questOverview })
      queryClient.invalidateQueries({ queryKey: keys.questHistory })
    },
  })
}

export function useUncompleteDaily() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => uncompleteDaily(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.questOverview })
      queryClient.invalidateQueries({ queryKey: keys.questHistory })
    },
  })
}

export function useIncrementWeekly() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => incrementWeekly(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.questOverview })
      queryClient.invalidateQueries({ queryKey: keys.questHistory })
    },
  })
}

export function useDecrementWeekly() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => decrementWeekly(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.questOverview })
      queryClient.invalidateQueries({ queryKey: keys.questHistory })
    },
  })
}
