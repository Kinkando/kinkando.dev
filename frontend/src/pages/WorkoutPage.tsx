import { useState } from 'react'
import { useDocumentTitle } from '../hooks/useDocumentTitle'
import { usePresets, useSchedule, useSessions } from '../queries/useWorkout'
import PresetsTab from '../components/workout/PresetsTab'
import ScheduleTab from '../components/workout/ScheduleTab'
import TodayTab from '../components/workout/TodayTab'
import HistoryTab from '../components/workout/HistoryTab'

type Tab = 'today' | 'presets' | 'schedule' | 'history'

const TABS: { key: Tab; label: string }[] = [
  { key: 'today', label: 'Today' },
  { key: 'presets', label: 'Presets' },
  { key: 'schedule', label: 'Schedule' },
  { key: 'history', label: 'History' },
]

function todayStr() {
  return new Date().toISOString().slice(0, 10)
}

export default function WorkoutPage() {
  useDocumentTitle('Workout')
  const [tab, setTab] = useState<Tab>('today')

  const presetsQuery = usePresets()
  const scheduleQuery = useSchedule()

  // Today's sessions: from = today, to = today
  const today = todayStr()
  const todaySessionsQuery = useSessions({ from: today, to: today })

  // History: last 60 days (default — no params)
  const historyQuery = useSessions()

  function refreshSessions() {
    todaySessionsQuery.refetch()
    historyQuery.refetch()
  }

  return (
    <main className="mx-auto max-w-5xl px-6 py-12">
      <h1 className="mb-6 text-2xl font-bold text-gray-100">Workout</h1>

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
      {tab === 'today' && (
        <TodayTab
          todaySessions={todaySessionsQuery.data ?? undefined}
          schedule={scheduleQuery.data ?? undefined}
          onSessionChange={refreshSessions}
        />
      )}
      {tab === 'presets' && (
        <PresetsTab
          presets={presetsQuery.data ?? undefined}
          onStarted={() => {
            refreshSessions()
            setTab('today')
          }}
        />
      )}
      {tab === 'schedule' && (
        <ScheduleTab
          schedule={scheduleQuery.data ?? undefined}
          presets={presetsQuery.data ?? undefined}
        />
      )}
      {tab === 'history' && (
        <HistoryTab sessions={historyQuery.data ?? undefined} />
      )}
    </main>
  )
}
