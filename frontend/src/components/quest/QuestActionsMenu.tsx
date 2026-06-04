import { useEffect, useRef, useState } from 'react'

type Props = {
  isActive: boolean
  label: string
  onNewQuest: () => void
}

export default function QuestActionsMenu({
  isActive,
  label,
  onNewQuest,
}: Props) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function onClickOutside(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    if (open) document.addEventListener('mousedown', onClickOutside)
    return () => document.removeEventListener('mousedown', onClickOutside)
  }, [open])

  return (
    <div className="relative shrink-0" ref={ref}>
      <button
        onClick={(e) => {
          e.stopPropagation()
          setOpen((o) => !o)
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
      {open && (
        <div className="absolute top-full right-0 z-20 mt-1 min-w-44 rounded-lg border border-gray-700 bg-gray-900 py-1 shadow-xl">
          <button
            className="w-full cursor-pointer px-3 py-1.5 text-left text-sm text-gray-300 hover:bg-gray-800"
            onClick={() => {
              setOpen(false)
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
