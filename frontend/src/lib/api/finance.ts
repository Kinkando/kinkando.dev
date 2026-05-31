import { apiFetch } from './client'
import type {
  FinanceRecord,
  CreateRecordInput,
  MonthlySummary,
  Category,
  CreateCategoryInput,
  UpdateCategoryInput,
} from './types'

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

export function fetchCategories(): Promise<Category[] | undefined> {
  return apiFetch<Category[]>('/finance/categories', { auth: true })
}

export function createCategory(
  input: CreateCategoryInput,
): Promise<Category | undefined> {
  return apiFetch<Category>('/finance/categories', {
    method: 'POST',
    body: input,
    auth: true,
  })
}

export function updateCategory(
  id: string,
  input: UpdateCategoryInput,
): Promise<Category | undefined> {
  return apiFetch<Category>(`/finance/categories/${id}`, {
    method: 'PATCH',
    body: input,
    auth: true,
  })
}

export function deleteCategory(id: string): Promise<undefined> {
  return apiFetch<undefined>(`/finance/categories/${id}`, {
    method: 'DELETE',
    auth: true,
  })
}
