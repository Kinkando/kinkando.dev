# frontend/CLAUDE.md

## Stack

- React 19 + Vite 6 (SPA, no SSR)
- TypeScript (strict)
- Tailwind CSS v4 (`@tailwindcss/vite` plugin; `@import "tailwindcss"` in `index.css`; no `tailwind.config` or `postcss.config`)
- TanStack Query v5 — data fetching and cache
- @dnd-kit — drag and drop (Kanban)
- Firebase Auth — email+password and Google sign-in
- react-router-dom v7 — client-side routing

## Dev server

```sh
pnpm install
pnpm dev   # http://localhost:3000 — port is FIXED (backend CORS allows only :3000)
```

## Environment variables

Prefix: `VITE_*`. Read via `import.meta.env.VITE_*` through `src/config/env.ts`. Never use `process.env`. Never commit `.env.local`.

| Variable | Purpose |
|---|---|
| `VITE_API_URL` | Backend base URL (e.g. `http://localhost:8080`) |
| `VITE_FIREBASE_API_KEY` | Firebase config |
| `VITE_FIREBASE_AUTH_DOMAIN` | Firebase config |
| `VITE_FIREBASE_PROJECT_ID` | Firebase config |
| `VITE_FIREBASE_STORAGE_BUCKET` | Firebase config |
| `VITE_FIREBASE_MESSAGING_SENDER_ID` | Firebase config |
| `VITE_FIREBASE_APP_ID` | Firebase config |

## Backend contract

- Base: `<VITE_API_URL>/api/v1`
- Response envelope: `{ "data": ... }` on success; `{ "error": "..." }` on failure
- `DELETE` and `PATCH .../move` return `204 No Content` (no body)
- Auth: `Authorization: Bearer <firebase id token>` on all protected routes
- **User provisioning**: after any sign-in, `POST /api/v1/users` must be called once per session before finance endpoints work. `AuthContext` handles this automatically via `provisionUser()`.

## Directory layout

```
src/
  config/env.ts              # import.meta.env.VITE_* → typed object
  lib/
    firebase.ts              # Firebase app + auth singleton, getIdToken helper
    api/
      client.ts              # apiFetch wrapper (envelope unwrap, Bearer auth, ApiError)
      types.ts               # DTO TypeScript types
      users|portfolio|finance|kanban.ts
  auth/
    AuthContext.tsx           # onAuthStateChanged, ensureUserOnce, useAuth()
    ProtectedRoute.tsx        # redirects to /login?redirect=... if unauthenticated
  queries/                   # TanStack Query hooks
  components/
    NavBar.tsx, AuthForm.tsx
    finance/, kanban/
  pages/
```

## Rules

- Never read from `process.env`; always use `import.meta.env.VITE_*` via `src/config/env.ts`
- Never commit `.env.local`
- Port 3000 is required by backend CORS — do not change `server.port` in `vite.config.ts`
- Finance DTO is named `FinanceRecord` (not `Record`) in TypeScript to avoid collision with the TS builtin
