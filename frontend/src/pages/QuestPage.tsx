import { useState } from 'react'
import { useDocumentTitle } from '../hooks/useDocumentTitle'
import { useQuestOverview, useQuestHistory } from '../queries/useQuest'
import DashboardTab from '../components/quest/DashboardTab'
import DailyTab from '../components/quest/DailyTab'
import WeeklyTab from '../components/quest/WeeklyTab'
import HistoryTab from '../components/quest/HistoryTab'

type Tab = 'dashboard' | 'daily' | 'weekly' | 'history'

const TABS: { key: Tab; label: string }[] = [
  { key: 'dashboard', label: 'Dashboard' },
  { key: 'daily', label: 'Daily' },
  { key: 'weekly', label: 'Weekly' },
  { key: 'history', label: 'History' },
]

export default function QuestPage() {
  useDocumentTitle('Quest')
  const [tab, setTab] = useState<Tab>('dashboard')

  const overviewQuery = useQuestOverview()
  const historyQuery = useQuestHistory(100)

  return (
    <main className="mx-auto max-w-5xl px-6 py-12">
      <h1 className="mb-6 text-2xl font-bold text-gray-100">Quest</h1>

      {/* Tab bar */}
      <div className="mb-6 flex gap-1 rounded-lg border border-gray-800 bg-gray-900 p-1">
        {TABS.map(({ key, label }) => (
          <button
            key={key}
            onClick={() => setTab(key)}
            className={`flex-1 rounded-md py-1.5 text-sm font-medium transition-colors ${
              tab === key
                ? 'bg-gray-800 text-gray-100'
                : 'text-gray-500 hover:text-gray-300'
            }`}
          >
            {label}
          </button>
        ))}
      </div>

      {/* Tab content */}
      {tab === 'dashboard' && <DashboardTab overview={overviewQuery.data} />}
      {tab === 'daily' && <DailyTab daily={overviewQuery.data?.daily ?? []} />}
      {tab === 'weekly' && (
        <WeeklyTab weekly={overviewQuery.data?.weekly ?? []} />
      )}
      {tab === 'history' && <HistoryTab events={historyQuery.data ?? []} />}
    </main>
  )
}
