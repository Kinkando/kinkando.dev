# Frontend Environment Variables

Prefix: `VITE_*`. Read via `import.meta.env.VITE_*` through `src/config/env.ts`. Never use `process.env`. Never commit `.env.local`.

The `VITE_FIREBASE_*` vars are also injected into `public/firebase-messaging-sw.js` at build/dev time by the `firebase-sw-env` Vite plugin (`vite.config.ts`) — do not hardcode values in that file.

| Variable | Purpose |
|---|---|
| `VITE_API_URL` | Backend base URL (e.g. `http://localhost:8080`) |
| `VITE_FIREBASE_API_KEY` | Firebase config |
| `VITE_FIREBASE_AUTH_DOMAIN` | Firebase config |
| `VITE_FIREBASE_PROJECT_ID` | Firebase config |
| `VITE_FIREBASE_STORAGE_BUCKET` | Firebase config |
| `VITE_FIREBASE_MESSAGING_SENDER_ID` | Firebase config |
| `VITE_FIREBASE_APP_ID` | Firebase config |
| `VITE_FIREBASE_VAPID_KEY` | FCM Web Push VAPID key — Firebase Console → Project Settings → Cloud Messaging → Web Push certificates |
