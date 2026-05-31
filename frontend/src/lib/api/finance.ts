import { apiFetch } from './client';
import type { CreateRecordInput, FinanceRecord, FinanceSummary } from './types';

export function listRecords(month: string): Promise<FinanceRecord[]> {
  return apiFetch<FinanceRecord[]>('/finance/records', { auth: true, query: { month } });
}

export function createRecord(input: CreateRecordInput): Promise<FinanceRecord> {
  return apiFetch<FinanceRecord>('/finance/records', { method: 'POST', auth: true, body: input });
}

export function deleteRecord(id: string): Promise<void> {
  return apiFetch<void>(`/finance/records/${id}`, { method: 'DELETE', auth: true });
}

export function getSummary(month: string): Promise<FinanceSummary> {
  return apiFetch<FinanceSummary>('/finance/summary', { auth: true, query: { month } });
}
