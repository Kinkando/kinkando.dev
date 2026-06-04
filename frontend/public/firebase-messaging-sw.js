// Firebase Cloud Messaging service worker.
// Handles background push notifications when the app is not in the foreground.
//
// NOTE: This file is served at /firebase-messaging-sw.js (site root) and runs
// in a service-worker context where import.meta.env is NOT available.
// Copy your VITE_FIREBASE_* values into the firebaseConfig object below.
// These are public client-side values — safe to commit.

importScripts('https://www.gstatic.com/firebasejs/10.12.0/firebase-app-compat.js')
importScripts('https://www.gstatic.com/firebasejs/10.12.0/firebase-messaging-compat.js')

// TODO: fill in your Firebase project config (matches VITE_FIREBASE_* env vars)
const firebaseConfig = {
  apiKey: '',
  authDomain: '',
  projectId: '',
  storageBucket: '',
  messagingSenderId: '',
  appId: '',
}

firebase.initializeApp(firebaseConfig)

const messaging = firebase.messaging()

messaging.onBackgroundMessage((payload) => {
  const title = payload.notification?.title ?? 'Notification'
  const body = payload.notification?.body ?? ''

  self.registration.showNotification(title, {
    body,
    icon: '/logo.png',
  })
})
