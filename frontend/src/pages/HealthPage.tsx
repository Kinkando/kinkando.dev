import { useSearchParams } from 'react-router-dom'
import { useDocumentTitle } from '../hooks/useDocumentTitle'
import { useHealthProfile, useWeightLogs } from '../queries/useHealth'
import DashboardTab from '../components/health/DashboardTab'
import ProgressTab from '../components/health/ProgressTab'
import SettingsTab from '../components/health/SettingsTab'

type Tab = 'dashboard' | 'progress' | 'settings'

const TABS: { key: Tab; label: string }[] = [
  { key: 'dashboard', label: 'Dashboard' },
  { key: 'progress', label: 'Progress' },
  { key: 'settings', label: 'Settings' },
]

const TAB_KEYS = new Set<string>(TABS.map((t) => t.key))

function toTab(value: string | null): Tab {
  return value !== null && TAB_KEYS.has(value) ? (value as Tab) : 'dashboard'
}

export default function HealthPage() {
  useDocumentTitle('Health')
  const [searchParams, setSearchParams] = useSearchParams()
  const tab = toTab(searchParams.get('tab'))

  function setTab(key: Tab) {
    setSearchParams(key === 'dashboard' ? {} : { tab: key }, { replace: true })
  }

  const profileQuery = useHealthProfile()
  const weightQuery = useWeightLogs()

  return (
    <main className="mx-auto max-w-5xl px-6 py-12">
      <h1 className="mb-6 text-2xl font-bold text-gray-100">Health</h1>

      {/* Tab bar */}
      <div className="mb-6 flex gap-1 rounded-lg border border-gray-800 bg-gray-900 p-1">
        {TABS.map(({ key, label }) => (
          <button
            key={key}
            onClick={() => setTab(key)}
            className={`flex-1 cursor-pointer rounded-md py-1.5 text-sm font-medium transition-colors ${
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
      {tab === 'dashboard' && (
        <DashboardTab
          profile={profileQuery.data}
          weightLogs={weightQuery.data}
          onGoToSettings={() => setTab('settings')}
        />
      )}
      {tab === 'progress' && (
        <ProgressTab
          weightLogs={weightQuery.data}
          profile={profileQuery.data}
        />
      )}
      {tab === 'settings' && <SettingsTab profile={profileQuery.data} />}
    </main>
  )
}
