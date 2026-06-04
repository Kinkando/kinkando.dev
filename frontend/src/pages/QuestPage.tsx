import { useEffect, useRef, useState } from 'react'
import { useDocumentTitle } from '../hooks/useDocumentTitle'
import { useQuestOverview, useQuestHistory } from '../queries/useQuest'
import type { QuestType } from '../lib/api/types'
import DashboardTab from '../components/quest/DashboardTab'
import QuestTab from '../components/quest/QuestTab'
import HistoryTab from '../components/quest/HistoryTab'
import QuestFormDialog from '../components/quest/QuestFormDialog'

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
  const [menuOpen, setMenuOpen] = useState<QuestType | null>(null)
  const [createType, setCreateType] = useState<QuestType | null>(null)

  const dailyMenuRef = useRef<HTMLDivElement>(null)
  const weeklyMenuRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function onClickOutside(e: MouseEvent) {
      const ref = menuOpen === 'daily' ? dailyMenuRef : weeklyMenuRef
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setMenuOpen(null)
      }
    }
    if (menuOpen) document.addEventListener('mousedown', onClickOutside)
    return () => document.removeEventListener('mousedown', onClickOutside)
  }, [menuOpen])

  const overviewQuery = useQuestOverview()
  const historyQuery = useQuestHistory(100)

  const xp = overviewQuery.data?.xp

  return (
    <main className="mx-auto max-w-5xl px-4 py-8 sm:px-6 sm:py-12">
      {/* Page header */}
      <div className="mb-8 flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="text-2xl font-black tracking-tight text-gray-100">
            Quest Log
          </h1>
          <p className="mt-0.5 text-sm text-gray-500">
            Track your daily habits and weekly goals
          </p>
        </div>
        {xp && (
          <div className="flex shrink-0 items-center gap-2 rounded-lg border border-amber-900/40 bg-amber-950/20 px-3 py-1.5">
            <span className="text-xs font-medium text-amber-600">AR</span>
            <span className="text-lg font-black text-amber-400">
              {xp.level}
            </span>
            <span className="text-xs text-gray-600">
              · {xp.total_xp.toLocaleString()} XP
            </span>
          </div>
        )}
      </div>

      {/* Tab bar */}
      <div className="mb-6 flex gap-1 rounded-lg border border-gray-800 bg-gray-900 p-1">
        {TABS.map(({ key, label }) => {
          const isActive = tab === key

          if (key === 'daily' || key === 'weekly') {
            const menuRef = key === 'daily' ? dailyMenuRef : weeklyMenuRef
            return (
              <div
                key={key}
                ref={menuRef}
                className={`relative flex flex-1 items-center rounded-md transition-colors ${
                  isActive ? 'bg-gray-800' : ''
                }`}
              >
                <button
                  onClick={() => setTab(key)}
                  className={`flex-1 cursor-pointer py-1.5 pl-3 text-left text-sm font-medium transition-colors ${
                    isActive
                      ? 'text-gray-100'
                      : 'text-gray-500 hover:text-gray-300'
                  }`}
                >
                  {label}
                </button>
                <button
                  onClick={(e) => {
                    e.stopPropagation()
                    setMenuOpen(menuOpen === key ? null : key)
                  }}
                  className={`cursor-pointer px-2 py-1.5 text-sm transition-colors ${
                    isActive
                      ? 'text-gray-400 hover:text-gray-200'
                      : 'text-gray-600 hover:text-gray-400'
                  }`}
                  title={`${label} actions`}
                  aria-label={`${label} actions`}
                >
                  ⋮
                </button>
                {menuOpen === key && (
                  <div className="absolute top-full left-0 z-20 mt-1 min-w-44 rounded-lg border border-gray-700 bg-gray-900 py-1 shadow-xl">
                    <button
                      className="w-full cursor-pointer px-3 py-1.5 text-left text-sm text-gray-300 hover:bg-gray-800"
                      onClick={() => {
                        setMenuOpen(null)
                        setTab(key)
                        setCreateType(key)
                      }}
                    >
                      + New Quest
                    </button>
                    <div className="my-1 border-t border-gray-800" />
                    <button
                      disabled
                      className="flex w-full cursor-not-allowed items-center justify-between px-3 py-1.5 text-left text-sm text-gray-600"
                    >
                      <span>Import Template</span>
                      <span className="rounded bg-gray-800 px-1.5 py-0.5 text-[10px] font-medium text-gray-500">
                        Soon
                      </span>
                    </button>
                    <button
                      disabled
                      className="flex w-full cursor-not-allowed items-center justify-between px-3 py-1.5 text-left text-sm text-gray-600"
                    >
                      <span>Bulk Actions</span>
                      <span className="rounded bg-gray-800 px-1.5 py-0.5 text-[10px] font-medium text-gray-500">
                        Soon
                      </span>
                    </button>
                  </div>
                )}
              </div>
            )
          }

          return (
            <button
              key={key}
              onClick={() => setTab(key)}
              className={`flex-1 cursor-pointer rounded-md py-1.5 text-sm font-medium transition-colors ${
                isActive
                  ? 'bg-gray-800 text-gray-100'
                  : 'text-gray-500 hover:text-gray-300'
              }`}
            >
              {label}
            </button>
          )
        })}
      </div>

      {/* Tab content */}
      {tab === 'dashboard' && <DashboardTab overview={overviewQuery.data} />}
      {tab === 'daily' && (
        <QuestTab type="daily" quests={overviewQuery.data?.daily ?? []} />
      )}
      {tab === 'weekly' && (
        <QuestTab type="weekly" quests={overviewQuery.data?.weekly ?? []} />
      )}
      {tab === 'history' && <HistoryTab events={historyQuery.data ?? []} />}

      {createType !== null && (
        <QuestFormDialog
          type={createType}
          onClose={() => setCreateType(null)}
        />
      )}
    </main>
  )
}
