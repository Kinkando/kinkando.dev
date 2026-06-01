import { apiFetch } from './client'
import type {
  WorkoutPreset,
  CreatePresetInput,
  UpdatePresetInput,
  WorkoutScheduleEntry,
  SetScheduleInput,
  WorkoutSession,
  WorkoutSessionExercise,
  UpdateSessionInput,
  UpdateSessionExerciseInput,
  AddSessionExerciseInput,
} from './types'

// ── Presets ────────────────────────────────────────────────────────────────────

export function fetchPresets(): Promise<WorkoutPreset[] | undefined> {
  return apiFetch<WorkoutPreset[]>('/workout/presets', { auth: true })
}

export function fetchPreset(id: string): Promise<WorkoutPreset | undefined> {
  return apiFetch<WorkoutPreset>(`/workout/presets/${id}`, { auth: true })
}

export function createPreset(
  input: CreatePresetInput,
): Promise<WorkoutPreset | undefined> {
  return apiFetch<WorkoutPreset>('/workout/presets', {
    method: 'POST',
    body: input,
    auth: true,
  })
}

export function updatePreset(
  id: string,
  input: UpdatePresetInput,
): Promise<WorkoutPreset | undefined> {
  return apiFetch<WorkoutPreset>(`/workout/presets/${id}`, {
    method: 'PUT',
    body: input,
    auth: true,
  })
}

export function deletePreset(id: string): Promise<undefined> {
  return apiFetch<undefined>(`/workout/presets/${id}`, {
    method: 'DELETE',
    auth: true,
  })
}

// ── Schedule ───────────────────────────────────────────────────────────────────

export function fetchSchedule(): Promise<WorkoutScheduleEntry[] | undefined> {
  return apiFetch<WorkoutScheduleEntry[]>('/workout/schedule', { auth: true })
}

export function setSchedule(
  input: SetScheduleInput,
): Promise<WorkoutScheduleEntry[] | undefined> {
  return apiFetch<WorkoutScheduleEntry[]>('/workout/schedule', {
    method: 'PUT',
    body: input,
    auth: true,
  })
}

// ── Sessions ───────────────────────────────────────────────────────────────────

export function fetchSessions(params?: {
  from?: string
  to?: string
}): Promise<WorkoutSession[] | undefined> {
  const query: Record<string, string> = {}
  if (params?.from) query.from = params.from
  if (params?.to) query.to = params.to
  return apiFetch<WorkoutSession[]>('/workout/sessions', { auth: true, query })
}

export function fetchSession(id: string): Promise<WorkoutSession | undefined> {
  return apiFetch<WorkoutSession>(`/workout/sessions/${id}`, { auth: true })
}

export function generateSession(
  date?: string,
): Promise<WorkoutSession | undefined> {
  return apiFetch<WorkoutSession>('/workout/sessions/generate', {
    method: 'POST',
    body: { date: date ?? '' },
    auth: true,
  })
}

export function createSession(input: {
  preset_id?: string
  type?: string
  date?: string
  name?: string
}): Promise<WorkoutSession | undefined> {
  return apiFetch<WorkoutSession>('/workout/sessions', {
    method: 'POST',
    body: input,
    auth: true,
  })
}

export function addSessionExercise(
  sessionId: string,
  input: AddSessionExerciseInput,
): Promise<WorkoutSessionExercise | undefined> {
  return apiFetch<WorkoutSessionExercise>(
    `/workout/sessions/${sessionId}/exercises`,
    {
      method: 'POST',
      body: input,
      auth: true,
    },
  )
}

export function deleteSessionExercise(
  sessionId: string,
  exId: string,
): Promise<undefined> {
  return apiFetch<undefined>(
    `/workout/sessions/${sessionId}/exercises/${exId}`,
    {
      method: 'DELETE',
      auth: true,
    },
  )
}

export function updateSession(
  id: string,
  input: UpdateSessionInput,
): Promise<WorkoutSession | undefined> {
  return apiFetch<WorkoutSession>(`/workout/sessions/${id}`, {
    method: 'PATCH',
    body: input,
    auth: true,
  })
}

export function updateSessionExercise(
  sessionId: string,
  exId: string,
  input: UpdateSessionExerciseInput,
): Promise<WorkoutSessionExercise | undefined> {
  return apiFetch<WorkoutSessionExercise>(
    `/workout/sessions/${sessionId}/exercises/${exId}`,
    {
      method: 'PATCH',
      body: input,
      auth: true,
    },
  )
}

export function deleteSession(id: string): Promise<undefined> {
  return apiFetch<undefined>(`/workout/sessions/${id}`, {
    method: 'DELETE',
    auth: true,
  })
}
