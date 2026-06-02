import { useState } from 'react'
import { useDocumentTitle } from '../hooks/useDocumentTitle'
import MedicinesTab from '../components/medicine/MedicinesTab'
import HistoryTab from '../components/medicine/HistoryTab'

type Tab = 'medicines' | 'history'

const TABS: { key: Tab; label: string }[] = [
  { key: 'medicines', label: 'Medicines' },
  { key: 'history', label: 'History' },
]

export default function MedicinePage() {
  useDocumentTitle('Medicine Tracker')
  const [tab, setTab] = useState<Tab>('medicines')

  return (
    <main className="mx-auto max-w-5xl px-6 py-12">
      <h1 className="mb-6 text-2xl font-bold text-gray-100">
        Medicine Tracker
      </h1>

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
      {tab === 'medicines' && <MedicinesTab />}
      {tab === 'history' && <HistoryTab />}
    </main>
  )
}
