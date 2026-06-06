import { useState } from 'react'
import { useInstallPrompt } from '../hooks/useInstallPrompt'

type Props = {
  /** Button classes so each placement matches its surrounding menu items. */
  className?: string
  /** Extra classes for the iOS hint line. */
  hintClassName?: string
  /** Called after a successful native prompt (e.g. to close the menu/drawer). */
  onAction?: () => void
}

export default function InstallButton({
  className,
  hintClassName,
  onAction,
}: Props) {
  const { canInstall, iosHint, promptInstall } = useInstallPrompt()
  const [showHint, setShowHint] = useState(false)

  if (!canInstall && !iosHint) return null

  async function handleClick() {
    if (canInstall) {
      await promptInstall()
      onAction?.()
    } else {
      setShowHint((s) => !s)
    }
  }

  return (
    <>
      <button
        onClick={handleClick}
        className={
          className ??
          'flex w-full cursor-pointer items-center gap-2 px-4 py-2 text-left text-sm text-gray-300 transition-colors hover:bg-gray-700 hover:text-gray-100'
        }
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-4 w-4 shrink-0"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth={1.5}
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5M16.5 12L12 16.5m0 0L7.5 12m4.5 4.5V3"
          />
        </svg>
        Install app
      </button>
      {iosHint && showHint && (
        <p className={hintClassName ?? 'px-4 pb-2 text-xs text-gray-500'}>
          Tap Share → Add to Home Screen
        </p>
      )}
    </>
  )
}
