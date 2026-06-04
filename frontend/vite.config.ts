import { defineConfig, loadEnv, type Plugin } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import fs from 'node:fs'
import path from 'node:path'

const SW = 'firebase-messaging-sw.js'

/**
 * Vite plugin that injects VITE_FIREBASE_* env values into the FCM service
 * worker (public/firebase-messaging-sw.js) at both dev-serve and build time.
 *
 * The committed SW file uses __VITE_FIREBASE_*__ placeholder tokens so no
 * real config lives in the repository; this plugin substitutes them at the
 * point Vite serves or emits the file.
 */
function firebaseSwEnv(env: Record<string, string>): Plugin {
  const keys = [
    'VITE_FIREBASE_API_KEY',
    'VITE_FIREBASE_AUTH_DOMAIN',
    'VITE_FIREBASE_PROJECT_ID',
    'VITE_FIREBASE_STORAGE_BUCKET',
    'VITE_FIREBASE_MESSAGING_SENDER_ID',
    'VITE_FIREBASE_APP_ID',
  ]
  const fill = (code: string) =>
    keys.reduce((out, k) => out.replaceAll(`__${k}__`, env[k] ?? ''), code)

  return {
    name: 'firebase-sw-env',
    // Dev: intercept before Vite's built-in public-file handler so the browser
    // never sees the raw placeholder tokens.
    configureServer(server) {
      server.middlewares.use((req, res, next) => {
        if (req.url?.split('?')[0] === `/${SW}`) {
          const src = path.resolve(__dirname, 'public', SW)
          res.setHeader('Content-Type', 'text/javascript')
          res.end(fill(fs.readFileSync(src, 'utf-8')))
          return
        }
        next()
      })
    },
    // Build: rewrite the file Vite copied verbatim from public/ into dist/.
    closeBundle() {
      const out = path.resolve(__dirname, 'dist', SW)
      if (fs.existsSync(out)) {
        fs.writeFileSync(out, fill(fs.readFileSync(out, 'utf-8')))
      }
    },
  }
}

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), 'VITE_')
  return {
    plugins: [react(), tailwindcss(), firebaseSwEnv(env)],
    server: { port: 3000, strictPort: true },
  }
})
