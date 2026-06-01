import { apiFetch } from './client'
import type {
  HealthProfile,
  UpsertProfileInput,
  WeightLog,
  CreateWeightInput,
  HealthExercise,
  CreateExerciseInput,
  UpdateExerciseInput,
} from './types'

export function fetchProfile(): Promise<HealthProfile | null | undefined> {
  return apiFetch<HealthProfile | null>('/health/profile', { auth: true })
}

export function upsertProfile(
  input: UpsertProfileInput,
): Promise<HealthProfile | undefined> {
  return apiFetch<HealthProfile>('/health/profile', {
    method: 'PUT',
    body: input,
    auth: true,
  })
}

export function fetchWeightLogs(): Promise<WeightLog[] | undefined> {
  return apiFetch<WeightLog[]>('/health/weight', { auth: true })
}

export function createWeightLog(
  input: CreateWeightInput,
): Promise<WeightLog | undefined> {
  return apiFetch<WeightLog>('/health/weight', {
    method: 'POST',
    body: input,
    auth: true,
  })
}

export function deleteWeightLog(id: string): Promise<undefined> {
  return apiFetch<undefined>(`/health/weight/${id}`, {
    method: 'DELETE',
    auth: true,
  })
}

export function fetchExercises(): Promise<HealthExercise[] | undefined> {
  return apiFetch<HealthExercise[]>('/health/exercises', { auth: true })
}

export function createExercise(
  input: CreateExerciseInput,
): Promise<HealthExercise | undefined> {
  return apiFetch<HealthExercise>('/health/exercises', {
    method: 'POST',
    body: input,
    auth: true,
  })
}

export function updateExercise(
  id: string,
  input: UpdateExerciseInput,
): Promise<HealthExercise | undefined> {
  return apiFetch<HealthExercise>(`/health/exercises/${id}`, {
    method: 'PATCH',
    body: input,
    auth: true,
  })
}

export function deleteExercise(id: string): Promise<undefined> {
  return apiFetch<undefined>(`/health/exercises/${id}`, {
    method: 'DELETE',
    auth: true,
  })
}
