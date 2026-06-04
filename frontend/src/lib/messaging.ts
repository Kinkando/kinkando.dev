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
 * Registers a foreground message handler. Returns a cleanup function.
 */
export function onForegroundMessage(
  cb: (title: string, body: string) => void,
): () => void {
  return onMessage(messaging, (payload) => {
    const title = payload.notification?.title ?? ''
    const body = payload.notification?.body ?? ''
    cb(title, body)
  })
}
