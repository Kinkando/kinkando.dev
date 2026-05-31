'use client';

import { useCallback, useState } from 'react';

import * as financeApi from '@/lib/api/finance';
import type { CreateRecordInput, FinanceRecord, MonthlySummary } from '@/types/finance';

export function useFinance() {
  const [records, setRecords] = useState<FinanceRecord[]>([]);
  const [summary, setSummary] = useState<MonthlySummary | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchRecords = useCallback(async (month: string) => {
    setLoading(true);
    setError(null);
    try {
      const [recs, sum] = await Promise.all([financeApi.listRecords(month), financeApi.getMonthlySummary(month)]);
      setRecords(recs);
      setSummary(sum);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch records');
    } finally {
      setLoading(false);
    }
  }, []);

  const addRecord = useCallback(
    async (input: CreateRecordInput, month: string) => {
      await financeApi.createRecord(input);
      await fetchRecords(month);
    },
    [fetchRecords]
  );

  const removeRecord = useCallback(
    async (id: string, month: string) => {
      await financeApi.deleteRecord(id);
      await fetchRecords(month);
    },
    [fetchRecords]
  );

  return { records, summary, loading, error, fetchRecords, addRecord, removeRecord };
}
