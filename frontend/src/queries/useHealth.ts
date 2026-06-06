import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  fetchProfile,
  upsertProfile,
  fetchWeightLogs,
  createWeightLog,
  updateWeightLog,
  deleteWeightLog,
  fetchFoodLogs,
  createFoodLog,
  updateFoodLog,
  deleteFoodLog,
  fetchSleepLogs,
  createSleepLog,
  updateSleepLog,
  deleteSleepLog,
} from '../lib/api/health'
import type {
  UpsertProfileInput,
  CreateWeightInput,
  UpdateWeightInput,
  CreateFoodInput,
  UpdateFoodInput,
  CreateSleepInput,
  UpdateSleepInput,
} from '../lib/api/types'
import { keys } from './keys'

function rangeKey(params?: { from?: string; to?: string }) {
  return `${params?.from ?? ''}_${params?.to ?? ''}`
}

export function useHealthProfile() {
  return useQuery({
    queryKey: keys.healthProfile,
    queryFn: fetchProfile,
  })
}

export function useUpsertProfile() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: UpsertProfileInput) => upsertProfile(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.healthProfile })
    },
  })
}

export function useWeightLogs(params?: { from?: string; to?: string }) {
  return useQuery({
    queryKey: keys.healthWeight(rangeKey(params)),
    queryFn: () => fetchWeightLogs(params),
  })
}

export function useCreateWeightLog() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateWeightInput) => createWeightLog(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['health', 'weight'] })
    },
  })
}

export function useUpdateWeightLog() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateWeightInput }) =>
      updateWeightLog(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['health', 'weight'] })
    },
  })
}

export function useDeleteWeightLog() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deleteWeightLog(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['health', 'weight'] })
    },
  })
}

// ── Food ─────────────────────────────────────────────────────────────────────

export function useFoodLogs() {
  return useQuery({
    queryKey: keys.healthFood,
    queryFn: fetchFoodLogs,
  })
}

export function useCreateFoodLog() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateFoodInput) => createFoodLog(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.healthFood })
    },
  })
}

export function useUpdateFoodLog() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateFoodInput }) =>
      updateFoodLog(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.healthFood })
    },
  })
}

export function useDeleteFoodLog() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deleteFoodLog(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.healthFood })
    },
  })
}

// ── Sleep ─────────────────────────────────────────────────────────────────────

export function useSleepLogs(params?: { from?: string; to?: string }) {
  return useQuery({
    queryKey: keys.healthSleep(rangeKey(params)),
    queryFn: () => fetchSleepLogs(params),
  })
}

export function useCreateSleepLog() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateSleepInput) => createSleepLog(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['health', 'sleep'] })
    },
  })
}

export function useUpdateSleepLog() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateSleepInput }) =>
      updateSleepLog(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['health', 'sleep'] })
    },
  })
}

export function useDeleteSleepLog() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deleteSleepLog(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['health', 'sleep'] })
    },
  })
}
