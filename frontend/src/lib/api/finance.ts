import { apiFetch } from '@/lib/api';
import type { CreateRecordInput, FinanceRecord, MonthlySummary } from '@/types/finance';

export function listRecords(month: string): Promise<FinanceRecord[]> {
  return apiFetch(`/api/v1/finance/records?month=${month}`);
}

export function createRecord(input: CreateRecordInput): Promise<FinanceRecord> {
  return apiFetch('/api/v1/finance/records', {
    method: 'POST',
    body: JSON.stringify(input)
  });
}

export function deleteRecord(id: string): Promise<void> {
  return apiFetch(`/api/v1/finance/records/${id}`, { method: 'DELETE' });
}

export function getMonthlySummary(month: string): Promise<MonthlySummary> {
  return apiFetch(`/api/v1/finance/summary?month=${month}`);
}
