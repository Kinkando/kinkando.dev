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
 * Requests notification permission, registers the service worker, and returns
 * an FCM device token. Returns null when permission is denied or an error occurs.
 */
export async function requestPushToken(): Promise<string | null> {
  if (!isPushSupported()) return null

  const permission = await Notification.requestPermission()
  if (permission !== 'granted') return null

  try {
    const registration = await navigator.serviceWorker.register(
      '/firebase-messaging-sw.js',
    )
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
 */
export async function showLocalNotification(
  title: string,
  body: string,
): Promise<void> {
  const reg = await navigator.serviceWorker.ready
  await reg.showNotification(title, { body, icon: '/logo.png' })
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
