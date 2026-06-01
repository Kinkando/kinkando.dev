# Frontend Stack & Dev Server

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

Port 3000 is required by backend CORS — do not change `server.port` in `vite.config.ts`.
