import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  fetchProfile,
  upsertProfile,
  fetchWeightLogs,
  createWeightLog,
  deleteWeightLog,
  fetchExercises,
  createExercise,
  updateExercise,
  deleteExercise,
} from '../lib/api/health'
import type {
  UpsertProfileInput,
  CreateWeightInput,
  CreateExerciseInput,
  UpdateExerciseInput,
} from '../lib/api/types'
import { keys } from './keys'

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

export function useWeightLogs() {
  return useQuery({
    queryKey: keys.healthWeight,
    queryFn: fetchWeightLogs,
  })
}

export function useCreateWeightLog() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateWeightInput) => createWeightLog(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.healthWeight })
    },
  })
}

export function useDeleteWeightLog() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deleteWeightLog(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.healthWeight })
    },
  })
}

export function useExercises() {
  return useQuery({
    queryKey: keys.healthExercises,
    queryFn: fetchExercises,
  })
}

export function useCreateExercise() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateExerciseInput) => createExercise(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.healthExercises })
    },
  })
}

export function useUpdateExercise() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateExerciseInput }) =>
      updateExercise(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.healthExercises })
    },
  })
}

export function useDeleteExercise() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deleteExercise(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.healthExercises })
    },
  })
}
