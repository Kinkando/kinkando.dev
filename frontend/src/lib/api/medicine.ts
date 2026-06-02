import { apiFetch } from './client'
import type {
  Medicine,
  CreateMedicineInput,
  UpdateMedicineInput,
  TakeMedicineInput,
  AdjustStockInput,
  MedicineIntake,
  MedicineStockAdjustment,
  TakeResponse,
  AdjustStockResponse,
} from './types'

export function fetchMedicines(
  includeArchived = false,
): Promise<Medicine[] | undefined> {
  return apiFetch<Medicine[]>('/medicines', {
    auth: true,
    query: includeArchived ? { include_archived: 'true' } : undefined,
  })
}

export function createMedicine(
  input: CreateMedicineInput,
): Promise<Medicine | undefined> {
  return apiFetch<Medicine>('/medicines', {
    method: 'POST',
    body: input,
    auth: true,
  })
}

export function updateMedicine(
  id: string,
  input: UpdateMedicineInput,
): Promise<Medicine | undefined> {
  return apiFetch<Medicine>(`/medicines/${id}`, {
    method: 'PATCH',
    body: input,
    auth: true,
  })
}

export function archiveMedicine(id: string): Promise<Medicine | undefined> {
  return apiFetch<Medicine>(`/medicines/${id}/archive`, {
    method: 'POST',
    auth: true,
  })
}

export function unarchiveMedicine(id: string): Promise<Medicine | undefined> {
  return apiFetch<Medicine>(`/medicines/${id}/unarchive`, {
    method: 'POST',
    auth: true,
  })
}

export function takeMedicine(
  id: string,
  input: TakeMedicineInput,
): Promise<TakeResponse | undefined> {
  return apiFetch<TakeResponse>(`/medicines/${id}/take`, {
    method: 'POST',
    body: input,
    auth: true,
  })
}

export function adjustStock(
  id: string,
  input: AdjustStockInput,
): Promise<AdjustStockResponse | undefined> {
  return apiFetch<AdjustStockResponse>(`/medicines/${id}/stock`, {
    method: 'POST',
    body: input,
    auth: true,
  })
}

export function fetchIntakes(opts?: {
  date?: string
  limit?: number
}): Promise<MedicineIntake[] | undefined> {
  const query: Record<string, string> = {}
  if (opts?.date) query.date = opts.date
  if (opts?.limit) query.limit = String(opts.limit)
  return apiFetch<MedicineIntake[]>('/medicines/intakes', {
    auth: true,
    query: Object.keys(query).length > 0 ? query : undefined,
  })
}

export function fetchStockAdjustments(opts?: {
  date?: string
  limit?: number
}): Promise<MedicineStockAdjustment[] | undefined> {
  const query: Record<string, string> = {}
  if (opts?.date) query.date = opts.date
  if (opts?.limit) query.limit = String(opts.limit)
  return apiFetch<MedicineStockAdjustment[]>('/medicines/stock-adjustments', {
    auth: true,
    query: Object.keys(query).length > 0 ? query : undefined,
  })
}

export function fetchMedicineIntakes(
  medicineId: string,
  opts?: { date?: string; limit?: number },
): Promise<MedicineIntake[] | undefined> {
  const query: Record<string, string> = {}
  if (opts?.date) query.date = opts.date
  if (opts?.limit) query.limit = String(opts.limit)
  return apiFetch<MedicineIntake[]>(`/medicines/${medicineId}/intakes`, {
    auth: true,
    query: Object.keys(query).length > 0 ? query : undefined,
  })
}

export function fetchMedicineStockAdjustments(
  medicineId: string,
  opts?: { date?: string; limit?: number },
): Promise<MedicineStockAdjustment[] | undefined> {
  const query: Record<string, string> = {}
  if (opts?.date) query.date = opts.date
  if (opts?.limit) query.limit = String(opts.limit)
  return apiFetch<MedicineStockAdjustment[]>(
    `/medicines/${medicineId}/stock-adjustments`,
    {
      auth: true,
      query: Object.keys(query).length > 0 ? query : undefined,
    },
  )
}
