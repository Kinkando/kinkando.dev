import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  fetchPresets,
  fetchPreset,
  createPreset,
  updatePreset,
  deletePreset,
  fetchSchedule,
  setSchedule,
  fetchSessions,
  fetchSession,
  generateSession,
  createSession,
  updateSession,
  updateSessionExercise,
  bulkUpdateSessionExercises,
  deleteSession,
  finishSession,
  addSessionExercise,
  deleteSessionExercise,
} from '../lib/api/workout'
import type {
  CreatePresetInput,
  UpdatePresetInput,
  SetScheduleInput,
  UpdateSessionInput,
  UpdateSessionExerciseInput,
  BulkUpdateSessionExercisesInput,
  AddSessionExerciseInput,
} from '../lib/api/types'
import { keys } from './keys'

// ── Presets ────────────────────────────────────────────────────────────────────

export function usePresets() {
  return useQuery({
    queryKey: keys.workoutPresets,
    queryFn: fetchPresets,
  })
}

export function usePreset(id: string) {
  return useQuery({
    queryKey: keys.workoutPreset(id),
    queryFn: () => fetchPreset(id),
    enabled: !!id,
  })
}

export function useCreatePreset() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: CreatePresetInput) => createPreset(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.workoutPresets })
    },
  })
}

export function useUpdatePreset() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdatePresetInput }) =>
      updatePreset(id, input),
    onSuccess: (_data, { id }) => {
      queryClient.invalidateQueries({ queryKey: keys.workoutPresets })
      queryClient.invalidateQueries({ queryKey: keys.workoutPreset(id) })
    },
  })
}

export function useDeletePreset() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deletePreset(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.workoutPresets })
      // Also invalidate schedule since deleting a preset removes it from schedule.
      queryClient.invalidateQueries({ queryKey: keys.workoutSchedule })
    },
  })
}

// ── Schedule ───────────────────────────────────────────────────────────────────

export function useSchedule() {
  return useQuery({
    queryKey: keys.workoutSchedule,
    queryFn: fetchSchedule,
  })
}

export function useSetSchedule() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: SetScheduleInput) => setSchedule(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: keys.workoutSchedule })
    },
  })
}

// ── Sessions ───────────────────────────────────────────────────────────────────

/** Builds a stable cache-key string from an optional date range. */
function rangeKey(params?: { from?: string; to?: string }) {
  return `${params?.from ?? ''}_${params?.to ?? ''}`
}

export function useSessions(params?: { from?: string; to?: string }) {
  return useQuery({
    queryKey: keys.workoutSessions(rangeKey(params)),
    queryFn: () => fetchSessions(params),
  })
}

export function useSession(id: string) {
  return useQuery({
    queryKey: keys.workoutSession(id),
    queryFn: () => fetchSession(id),
    enabled: !!id,
  })
}

export function useGenerateSession() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (date?: string) => generateSession(date),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workout', 'sessions'] })
    },
  })
}

export function useCreateSession() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: {
      preset_id?: string
      type?: string
      date?: string
      name?: string
    }) => createSession(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workout', 'sessions'] })
    },
  })
}

export function useUpdateSession() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateSessionInput }) =>
      updateSession(id, input),
    onSuccess: (_data, { id }) => {
      queryClient.invalidateQueries({ queryKey: ['workout', 'sessions'] })
      queryClient.invalidateQueries({ queryKey: keys.workoutSession(id) })
    },
  })
}

export function useUpdateSessionExercise() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      sessionId,
      exId,
      input,
    }: {
      sessionId: string
      exId: string
      input: UpdateSessionExerciseInput
    }) => updateSessionExercise(sessionId, exId, input),
    onSuccess: (_data, { sessionId }) => {
      queryClient.invalidateQueries({
        queryKey: keys.workoutSession(sessionId),
      })
      queryClient.invalidateQueries({ queryKey: ['workout', 'sessions'] })
    },
  })
}

export function useBulkUpdateSessionExercises() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      sessionId,
      input,
    }: {
      sessionId: string
      input: BulkUpdateSessionExercisesInput
    }) => bulkUpdateSessionExercises(sessionId, input),
    onSuccess: (_data, { sessionId }) => {
      queryClient.invalidateQueries({
        queryKey: keys.workoutSession(sessionId),
      })
      queryClient.invalidateQueries({ queryKey: ['workout', 'sessions'] })
    },
  })
}

export function useDeleteSession() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deleteSession(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workout', 'sessions'] })
    },
  })
}

export function useFinishSession() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => finishSession(id),
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: ['workout', 'sessions'] })
      queryClient.invalidateQueries({ queryKey: keys.workoutSession(id) })
    },
  })
}

export function useAddSessionExercise() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      sessionId,
      input,
    }: {
      sessionId: string
      input: AddSessionExerciseInput
    }) => addSessionExercise(sessionId, input),
    onSuccess: (_data, { sessionId }) => {
      queryClient.invalidateQueries({
        queryKey: keys.workoutSession(sessionId),
      })
      queryClient.invalidateQueries({ queryKey: ['workout', 'sessions'] })
    },
  })
}

export function useDeleteSessionExercise() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ sessionId, exId }: { sessionId: string; exId: string }) =>
      deleteSessionExercise(sessionId, exId),
    onSuccess: (_data, { sessionId }) => {
      queryClient.invalidateQueries({
        queryKey: keys.workoutSession(sessionId),
      })
      queryClient.invalidateQueries({ queryKey: ['workout', 'sessions'] })
    },
  })
}
