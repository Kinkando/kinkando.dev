import { getMessaging, getToken, onMessage } from 'firebase/messaging'
import { app } from './firebase'
import env from '../config/env'

const messaging = getMessaging(app)

/**
 * Returns true when the browser supports Web Push notifications.
 * Blocks Safari < 16.1 and contexts where Notification is unavailable.
 */
export function isPushSupported(): boolean {
  return (
    typeof window !== 'undefined' &&
    'Notification' in window &&
    'serviceWorker' in navigator &&
    'PushManager' in window
  )
}

/**
 * Registers the Firebase messaging service worker and returns the registration.
 * Shared by requestPushToken and getCurrentToken.
 */
async function registerServiceWorker(): Promise<ServiceWorkerRegistration> {
  return navigator.serviceWorker.register('/scripts/firebase-messaging-sw.js')
}

/**
 * Requests notification permission, registers the service worker, and returns
 * an FCM device token. Returns null when permission is denied or an error occurs.
 */
export async function requestPushToken(): Promise<string | null> {
  if (!isPushSupported()) return null

  const permission = await Notification.requestPermission()
  if (permission !== 'granted') return null

  try {
    const registration = await registerServiceWorker()
    const token = await getToken(messaging, {
      vapidKey: env.vapidKey,
      serviceWorkerRegistration: registration,
    })
    return token ?? null
  } catch {
    return null
  }
}

/**
 * Returns the current FCM device token without prompting for permission.
 * Returns null when permission has not been granted or an error occurs.
 * Used for silent token refresh on app load and for the device status check.
 */
export async function getCurrentToken(): Promise<string | null> {
  if (!isPushSupported()) return null
  if (Notification.permission !== 'granted') return null

  try {
    const registration = await registerServiceWorker()
    const token = await getToken(messaging, {
      vapidKey: env.vapidKey,
      serviceWorkerRegistration: registration,
    })
    return token ?? null
  } catch {
    return null
  }
}

/**
 * Displays a notification via the active service-worker registration.
 * This works in the foreground (where the browser won't auto-display)
 * and produces the same OS-level toast as the background SW handler.
 *
 * Races against a 3-second timeout so the call never hangs indefinitely
 * when the service worker is not yet registered.
 */
export async function showLocalNotification(
  title: string,
  body: string,
): Promise<void> {
  const timeout = new Promise<never>((_, reject) =>
    setTimeout(() => reject(new Error('SW not ready')), 3000),
  )
  let reg: ServiceWorkerRegistration
  try {
    reg = await Promise.race([navigator.serviceWorker.ready, timeout])
  } catch {
    return
  }
  await reg.showNotification(title, { body, icon: '/images/logo.png' })
}

/**
 * Registers a foreground message handler that fires when a push message
 * arrives while the tab is in the foreground.
 *
 * Reads from payload.data (data-only messages) so it matches the backend
 * that sends { data: { title, body } } without a Notification field.
 *
 * Returns a cleanup function — call it to unsubscribe.
 */
export function onForegroundMessage(
  cb: (title: string, body: string) => void,
): () => void {
  return onMessage(messaging, (payload) => {
    const title = (payload.data?.['title'] as string | undefined) ?? ''
    const body = (payload.data?.['body'] as string | undefined) ?? ''
    cb(title, body)
  })
}
