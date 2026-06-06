import { apiFetch } from './client'
import type {
  HealthProfile,
  UpsertProfileInput,
  WeightLog,
  CreateWeightInput,
  UpdateWeightInput,
  FoodLog,
  CreateFoodInput,
  UpdateFoodInput,
  SleepLog,
  CreateSleepInput,
  UpdateSleepInput,
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

export function fetchWeightLogs(params?: {
  from?: string
  to?: string
}): Promise<WeightLog[] | undefined> {
  const query: Record<string, string> = {}
  if (params?.from) query.from = params.from
  if (params?.to) query.to = params.to
  return apiFetch<WeightLog[]>('/health/weight', { auth: true, query })
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

export function updateWeightLog(
  id: string,
  input: UpdateWeightInput,
): Promise<WeightLog | undefined> {
  return apiFetch<WeightLog>(`/health/weight/${id}`, {
    method: 'PATCH',
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

// ── Food ─────────────────────────────────────────────────────────────────────

export function fetchFoodLogs(): Promise<FoodLog[] | undefined> {
  return apiFetch<FoodLog[]>('/health/food', { auth: true })
}

export function createFoodLog(
  input: CreateFoodInput,
): Promise<FoodLog | undefined> {
  return apiFetch<FoodLog>('/health/food', {
    method: 'POST',
    body: input,
    auth: true,
  })
}

export function updateFoodLog(
  id: string,
  input: UpdateFoodInput,
): Promise<FoodLog | undefined> {
  return apiFetch<FoodLog>(`/health/food/${id}`, {
    method: 'PATCH',
    body: input,
    auth: true,
  })
}

export function deleteFoodLog(id: string): Promise<undefined> {
  return apiFetch<undefined>(`/health/food/${id}`, {
    method: 'DELETE',
    auth: true,
  })
}

// ── Sleep ─────────────────────────────────────────────────────────────────────

export function fetchSleepLogs(params?: {
  from?: string
  to?: string
}): Promise<SleepLog[] | undefined> {
  const query: Record<string, string> = {}
  if (params?.from) query.from = params.from
  if (params?.to) query.to = params.to
  return apiFetch<SleepLog[]>('/health/sleep', { auth: true, query })
}

export function createSleepLog(
  input: CreateSleepInput,
): Promise<SleepLog | undefined> {
  return apiFetch<SleepLog>('/health/sleep', {
    method: 'POST',
    body: input,
    auth: true,
  })
}

export function updateSleepLog(
  id: string,
  input: UpdateSleepInput,
): Promise<SleepLog | undefined> {
  return apiFetch<SleepLog>(`/health/sleep/${id}`, {
    method: 'PATCH',
    body: input,
    auth: true,
  })
}

export function deleteSleepLog(id: string): Promise<undefined> {
  return apiFetch<undefined>(`/health/sleep/${id}`, {
    method: 'DELETE',
    auth: true,
  })
}
