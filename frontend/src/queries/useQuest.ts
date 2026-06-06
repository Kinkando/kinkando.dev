import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  fetchQuestOverview,
  fetchQuestStreaks,
  fetchQuests,
  createQuest,
  updateQuest,
  deleteQuest,
  activateQuest,
  deactivateQuest,
  incrementQuest,
  decrementQuest,
  fetchQuestHistory,
} from '../lib/api/quest'
import type { CreateQuestInput, UpdateQuestInput } from '../lib/api/types'
import { keys } from './keys'

export function useQuestOverview() {
  return useQuery({ queryKey: keys.questOverview, queryFn: fetchQuestOverview })
}

export function useQuestStreaks() {
  return useQuery({ queryKey: keys.questStreaks, queryFn: fetchQuestStreaks })
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

export function useActivateQuest() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => activateQuest(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.questOverview })
      queryClient.invalidateQueries({ queryKey: keys.questList('daily') })
      queryClient.invalidateQueries({ queryKey: keys.questList('weekly') })
    },
  })
}

export function useDeactivateQuest() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deactivateQuest(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.questOverview })
      queryClient.invalidateQueries({ queryKey: keys.questList('daily') })
      queryClient.invalidateQueries({ queryKey: keys.questList('weekly') })
    },
  })
}

// ── Action mutations ──────────────────────────────────────────────────────────

export function useIncrementQuest() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => incrementQuest(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.questOverview })
      queryClient.invalidateQueries({ queryKey: keys.questHistory })
    },
  })
}

export function useDecrementQuest() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => decrementQuest(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.questOverview })
      queryClient.invalidateQueries({ queryKey: keys.questHistory })
    },
  })
}
