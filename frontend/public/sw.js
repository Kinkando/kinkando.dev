// kinkando.dev app service worker — installability + basic offline shell.
//
// Scope: '/' (registered from src/main.tsx in production only).
// This is SEPARATE from the Firebase messaging SW at /scripts/ — that one keeps
// handling push. This SW deliberately ignores /scripts/*, cross-origin requests
// and the backend API so push and data calls are never intercepted.

const CACHE = 'kinkando-shell-v1'
const PRECACHE = ['/', '/index.html']

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches
      .open(CACHE)
      .then((cache) => cache.addAll(PRECACHE))
      .then(() => self.skipWaiting()),
  )
})

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches
      .keys()
      .then((keys) =>
        Promise.all(keys.filter((k) => k !== CACHE).map((k) => caches.delete(k))),
      )
      .then(() => self.clients.claim()),
  )
})

self.addEventListener('fetch', (event) => {
  const { request } = event
  if (request.method !== 'GET') return

  const url = new URL(request.url)

  // Leave cross-origin (incl. backend API), the FCM SW, and /api to the network.
  if (url.origin !== self.location.origin) return
  if (url.pathname.startsWith('/scripts/')) return
  if (url.pathname.startsWith('/api')) return

  // Navigations: network-first, fall back to the cached app shell when offline.
  if (request.mode === 'navigate') {
    event.respondWith(
      fetch(request).catch(
        () =>
          caches
            .match('/index.html')
            .then((cached) => cached || caches.match('/')),
      ),
    )
    return
  }

  // Static assets (hashed by Vite): cache-first, then network and cache.
  event.respondWith(
    caches.match(request).then(
      (cached) =>
        cached ||
        fetch(request).then((resp) => {
          if (resp.ok && resp.type === 'basic') {
            const copy = resp.clone()
            caches.open(CACHE).then((cache) => cache.put(request, copy))
          }
          return resp
        }),
    ),
  )
})
