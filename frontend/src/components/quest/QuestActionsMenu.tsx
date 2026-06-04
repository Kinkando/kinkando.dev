import { useEffect, useRef, useState } from 'react'

type Props = {
  onNewQuest: () => void
}

export default function QuestActionsMenu({ onNewQuest }: Props) {
  const [menuOpen, setMenuOpen] = useState(false)
  const menuRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function onClickOutside(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuOpen(false)
      }
    }
    if (menuOpen) document.addEventListener('mousedown', onClickOutside)
    return () => document.removeEventListener('mousedown', onClickOutside)
  }, [menuOpen])

  return (
    <div className="relative shrink-0" ref={menuRef}>
      <button
        onClick={() => setMenuOpen((o) => !o)}
        className="cursor-pointer rounded-lg border border-gray-800 bg-gray-900 px-2 py-1.5 text-gray-400 hover:text-gray-200"
        title="Quest actions"
        aria-label="Quest actions"
      >
        ⋮
      </button>
      {menuOpen && (
        <div className="absolute top-9 right-0 z-20 min-w-44 rounded-lg border border-gray-700 bg-gray-900 py-1 shadow-xl">
          <button
            className="w-full cursor-pointer px-3 py-1.5 text-left text-sm text-gray-300 hover:bg-gray-800"
            onClick={() => {
              setMenuOpen(false)
              onNewQuest()
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
