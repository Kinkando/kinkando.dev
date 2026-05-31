'use client';

import { useEffect, useState } from 'react';

import { MonthlySummary } from '@/components/finance/MonthlySummary';
import { MonthPicker } from '@/components/finance/MonthPicker';
import { RecordForm } from '@/components/finance/RecordForm';
import { RecordList } from '@/components/finance/RecordList';
import { Button } from '@/components/ui/Button';
import { Modal } from '@/components/ui/Modal';
import { Spinner } from '@/components/ui/Spinner';
import { useFinance } from '@/hooks/useFinance';

function currentMonth() {
  const now = new Date();
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`;
}

export default function FinancePage() {
  const [month, setMonth] = useState(currentMonth);
  const [showForm, setShowForm] = useState(false);
  const { records, summary, loading, error, fetchRecords, addRecord, removeRecord } = useFinance();

  useEffect(() => {
    fetchRecords(month);
  }, [month, fetchRecords]);

  return (
    <div>
      <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <h1 className="text-2xl font-bold text-gray-900">Finance</h1>
        <div className="flex flex-wrap items-center gap-3">
          <MonthPicker value={month} onChange={setMonth} />
          <Button onClick={() => setShowForm(true)}>+ Add Record</Button>
        </div>
      </div>

      {error && <p className="mb-4 rounded bg-red-50 p-3 text-sm text-red-600">{error}</p>}

      {loading ? (
        <div className="flex justify-center py-12">
          <Spinner className="h-8 w-8" />
        </div>
      ) : (
        <>
          {summary && (
            <div className="mb-6">
              <MonthlySummary summary={summary} />
            </div>
          )}
          <div className="rounded-lg border border-gray-200 bg-white">
            <RecordList records={records} onDelete={(id) => removeRecord(id, month)} />
          </div>
        </>
      )}

      <Modal open={showForm} onClose={() => setShowForm(false)} title="Add Record">
        <RecordForm
          onSubmit={async (input) => {
            await addRecord(input, month);
            setShowForm(false);
          }}
          onCancel={() => setShowForm(false)}
        />
      </Modal>
    </div>
  );
}
