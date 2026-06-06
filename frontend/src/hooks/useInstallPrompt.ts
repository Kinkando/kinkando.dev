import { useEffect, useState } from 'react'

// The beforeinstallprompt event is not in the standard lib DOM types.
type BeforeInstallPromptEvent = Event & {
  prompt: () => Promise<void>
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>
}

// Module-level singleton: capture the deferred prompt at import time so it isn't
// missed before React mounts. Components subscribe via the hook below.
let deferred: BeforeInstallPromptEvent | null = null
const listeners = new Set<() => void>()

function emit() {
  listeners.forEach((l) => l())
}

if (typeof window !== 'undefined') {
  window.addEventListener('beforeinstallprompt', (e) => {
    e.preventDefault()
    deferred = e as BeforeInstallPromptEvent
    emit()
  })
  window.addEventListener('appinstalled', () => {
    deferred = null
    emit()
  })
}

function isStandalone(): boolean {
  if (typeof window === 'undefined') return false
  return (
    window.matchMedia?.('(display-mode: standalone)').matches ||
    // iOS Safari standalone flag (non-standard).
    (navigator as unknown as { standalone?: boolean }).standalone === true
  )
}

function isIOSSafari(): boolean {
  if (typeof navigator === 'undefined') return false
  const ua = navigator.userAgent
  return (
    /iphone|ipad|ipod/i.test(ua) &&
    /safari/i.test(ua) &&
    !/crios|fxios/i.test(ua)
  )
}

/**
 * Exposes install state for the in-app Install button.
 * - `canInstall`: a native install prompt is available and the app isn't installed.
 * - `iosHint`: iOS Safari (no beforeinstallprompt) and not yet installed — show
 *   a manual "Share → Add to Home Screen" hint instead.
 * - `promptInstall`: fire the native prompt (no-op when unavailable).
 */
export function useInstallPrompt() {
  const [, force] = useState(0)

  useEffect(() => {
    const l = () => force((n) => n + 1)
    listeners.add(l)
    return () => {
      listeners.delete(l)
    }
  }, [])

  const installed = isStandalone()

  return {
    canInstall: deferred !== null && !installed,
    iosHint: isIOSSafari() && !installed,
    async promptInstall() {
      if (!deferred) return
      await deferred.prompt()
      await deferred.userChoice
      deferred = null
      emit()
    },
  }
}
