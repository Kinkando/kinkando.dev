import { useState } from 'react'
import { useDocumentTitle } from '../../hooks/useDocumentTitle'
import type { MedicineSourceType } from '../../lib/api/types'
import MedicinesTab from './MedicinesTab'
import HistoryTab from './HistoryTab'

type Tab = 'list' | 'history'

type Props = {
  sourceType: MedicineSourceType
  title: string
  /** Label for the list tab (e.g. "Medicines" / "Supplements"). */
  listLabel: string
}

/**
 * Shared tracker UI for medicines and supplements — identical engine, scoped to
 * one source_type. Mounted by both MedicinePage and SupplementPage.
 */
export default function MedicineTracker({
  sourceType,
  title,
  listLabel,
}: Props) {
  useDocumentTitle(title)
  const [tab, setTab] = useState<Tab>('list')

  const TABS: { key: Tab; label: string }[] = [
    { key: 'list', label: listLabel },
    { key: 'history', label: 'History' },
  ]

  return (
    <main className="mx-auto max-w-5xl px-6 py-12">
      <h1 className="mb-6 text-2xl font-bold text-gray-100">{title}</h1>

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
      {tab === 'list' && <MedicinesTab sourceType={sourceType} />}
      {tab === 'history' && <HistoryTab sourceType={sourceType} />}
    </main>
  )
}
