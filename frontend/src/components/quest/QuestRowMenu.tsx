import { useEffect, useRef, useState } from 'react'

type Props = {
  isActive: boolean
  onEdit: () => void
  onToggleActive: () => void
  onDelete: () => void
}

export default function QuestRowMenu({
  isActive,
  onEdit,
  onToggleActive,
  onDelete,
}: Props) {
  const [menuOpen, setMenuOpen] = useState(false)
  const [openUpward, setOpenUpward] = useState(false)
  const menuRef = useRef<HTMLDivElement>(null)
  const buttonRef = useRef<HTMLButtonElement>(null)

  useEffect(() => {
    function onClickOutside(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuOpen(false)
      }
    }
    if (menuOpen) document.addEventListener('mousedown', onClickOutside)
    return () => document.removeEventListener('mousedown', onClickOutside)
  }, [menuOpen])

  function handleOpen() {
    if (buttonRef.current) {
      const rect = buttonRef.current.getBoundingClientRect()
      setOpenUpward(window.innerHeight - rect.bottom < 120)
    }
    setMenuOpen((o) => !o)
  }

  return (
    <div className="relative shrink-0" ref={menuRef}>
      <button
        ref={buttonRef}
        onClick={handleOpen}
        className="cursor-pointer px-1 text-gray-600 hover:text-gray-300"
        title="Quest options"
        aria-label="Quest options"
      >
        ⋮
      </button>
      {menuOpen && (
        <div
          className={`absolute ${openUpward ? 'bottom-6' : 'top-6'} right-0 z-20 min-w-32 rounded-lg border border-gray-700 bg-gray-900 py-1 shadow-xl`}
        >
          <button
            className="w-full cursor-pointer px-3 py-1.5 text-left text-sm text-gray-300 hover:bg-gray-800"
            onClick={() => {
              setMenuOpen(false)
              onEdit()
            }}
          >
            Edit
          </button>
          <button
            className="w-full cursor-pointer px-3 py-1.5 text-left text-sm text-gray-300 hover:bg-gray-800"
            onClick={() => {
              setMenuOpen(false)
              onToggleActive()
            }}
          >
            {isActive ? 'Disable' : 'Enable'}
          </button>
          <button
            className="w-full cursor-pointer px-3 py-1.5 text-left text-sm text-red-400 hover:bg-gray-800"
            onClick={() => {
              setMenuOpen(false)
              onDelete()
            }}
          >
            Delete
          </button>
        </div>
      )}
    </div>
  )
}
