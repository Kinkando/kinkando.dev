// Firebase Cloud Messaging service worker.
// Handles background push notifications when the app is not in the foreground.
//
// NOTE: This file is served at /scripts/firebase-messaging-sw.js and runs
// in a service-worker context where import.meta.env is NOT available.
// Values below are injected at build/dev time from VITE_FIREBASE_* env vars by
// the `firebase-sw-env` Vite plugin (see vite.config.ts). Do not hardcode real
// values here — the committed file ships only placeholder tokens.

importScripts('https://www.gstatic.com/firebasejs/10.12.0/firebase-app-compat.js')
importScripts('https://www.gstatic.com/firebasejs/10.12.0/firebase-messaging-compat.js')

const firebaseConfig = {
  apiKey: '__VITE_FIREBASE_API_KEY__',
  authDomain: '__VITE_FIREBASE_AUTH_DOMAIN__',
  projectId: '__VITE_FIREBASE_PROJECT_ID__',
  storageBucket: '__VITE_FIREBASE_STORAGE_BUCKET__',
  messagingSenderId: '__VITE_FIREBASE_MESSAGING_SENDER_ID__',
  appId: '__VITE_FIREBASE_APP_ID__',
}

firebase.initializeApp(firebaseConfig)

const messaging = firebase.messaging()

// Messages are data-only (no Notification field on the server payload), so we
// render the notification here ourselves to avoid browser auto-display duplicates.
messaging.onBackgroundMessage((payload) => {
  const title = payload.data?.title ?? 'Notification'
  const body = payload.data?.body ?? ''

  self.registration.showNotification(title, {
    body,
    icon: '../images/logo.png',
  })
})
