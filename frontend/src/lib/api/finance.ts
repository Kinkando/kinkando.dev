import { apiFetch } from './client'
import type { FinanceRecord, CreateRecordInput, MonthlySummary } from './types'

export function fetchRecords(
  month: string,
): Promise<FinanceRecord[] | undefined> {
  return apiFetch<FinanceRecord[]>('/finance/records', {
    auth: true,
    query: { month },
  })
}

export function createRecord(
  input: CreateRecordInput,
): Promise<FinanceRecord | undefined> {
  return apiFetch<FinanceRecord>('/finance/records', {
    method: 'POST',
    body: input,
    auth: true,
  })
}

export function deleteRecord(id: string): Promise<undefined> {
  return apiFetch<undefined>(`/finance/records/${id}`, {
    method: 'DELETE',
    auth: true,
  })
}

export function fetchSummary(
  month: string,
): Promise<MonthlySummary | undefined> {
  return apiFetch<MonthlySummary>('/finance/summary', {
    auth: true,
    query: { month },
  })
}
