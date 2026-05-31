import { useState } from 'react'
import RecordForm from '../components/finance/RecordForm'
import RecordList from '../components/finance/RecordList'
import SummaryPanel from '../components/finance/SummaryPanel'
import CategoryManager from '../components/finance/CategoryManager'
import { useRecords, useSummary } from '../queries/useFinance'
import { todayMonth } from '../lib/date'

export default function FinancePage() {
  const [month, setMonth] = useState(todayMonth)
  const { data: records, isLoading: loadingRecords } = useRecords(month)
  const { data: summary, isLoading: loadingSummary } = useSummary(month)

  return (
    <main className="mx-auto max-w-5xl px-6 py-12">
      <div className="mb-8 flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-100">Finance</h1>
        <input
          type="month"
          value={month}
          onChange={(e) => setMonth(e.target.value)}
          className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 focus:border-indigo-500 focus:outline-none"
        />
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <div className="flex flex-col gap-6 lg:col-span-1">
          <RecordForm month={month} />
          {loadingSummary ? (
            <p className="text-sm text-gray-500">Loading summary…</p>
          ) : summary ? (
            <SummaryPanel summary={summary} />
          ) : null}
          <CategoryManager />
        </div>
        <div className="lg:col-span-2">
          {loadingRecords ? (
            <p className="text-sm text-gray-500">Loading records…</p>
          ) : (
            <RecordList records={records ?? []} month={month} />
          )}
        </div>
      </div>
    </main>
  )
}
