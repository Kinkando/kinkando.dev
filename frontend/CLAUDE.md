# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Dev server (localhost:3000)
pnpm dev

# Production build
pnpm next build

# Lint
pnpm lint

# Build for Cloudflare Pages
pnpm build

# Preview Cloudflare build locally
pnpm preview

# Deploy to Cloudflare Pages
pnpm deploy
```

## Environment Setup

Copy `.env.example` to `.env.local` and fill in values:

- `NEXT_PUBLIC_API_URL` — Backend API base URL (default: `http://localhost:8080`)
- `NEXT_PUBLIC_FIREBASE_*` — Firebase project config (6 vars: API_KEY, AUTH_DOMAIN, PROJECT_ID, STORAGE_BUCKET, MESSAGING_SENDER_ID, APP_ID)

Firebase credentials are required for auth to work at runtime, but the app builds fine without them (Firebase is lazily initialized via dynamic imports).

## Architecture

Next.js 14 App Router with TypeScript and Tailwind CSS. Deploys to Cloudflare Pages via `@cloudflare/next-on-pages`.

### Directory structure

```
src/
  app/              # App Router pages and layouts
  components/       # React components by feature area
    ui/             # Generic primitives (Button, Input, Card, Modal, Spinner)
    layout/         # Navbar, Sidebar, DashboardHeader
    auth/           # AuthGuard, LoginForm, RegisterForm
    finance/        # MonthPicker, MonthlySummary, RecordList, RecordForm
    kanban/         # Board, Column, CardItem, AddCardForm
    portfolio/      # ProjectCard, SkillList
  contexts/         # React contexts (AuthContext)
  hooks/            # Custom hooks (useAuth, useFinance, useKanban)
  lib/              # Firebase init, API client
    api/            # Per-feature API functions (finance, kanban, portfolio)
  types/            # TypeScript interfaces mirroring backend models
```

### Key patterns

- **All data fetching is client-side** — backend CORS only allows browser origins (`localhost:3000`, `*.pages.dev`). Firebase tokens are also only available client-side.
- **Firebase is lazily initialized** via dynamic imports in `lib/firebase.ts` to avoid SSR/prerender errors. Never import `firebase/*` at module level outside of `lib/firebase.ts`.
- **Auth flow**: `AuthProvider` in root layout wraps the app. Dashboard routes are protected by `AuthGuard` (client-side redirect). `useAuth()` hook provides `user`, `loading`, `signIn`, `signUp`, `signOut`.
- **API client**: `lib/api.ts` exports `apiFetch<T>()` which auto-attaches Firebase ID token and unwraps `{ data: T }` response envelope. Feature-specific functions in `lib/api/*.ts`.
- **No middleware.ts** — Next.js edge middleware can't access Firebase client auth state.

### Routes

| Route                | Auth      | Description                           |
| -------------------- | --------- | ------------------------------------- |
| `/`                  | Public    | Portfolio landing (projects + skills) |
| `/login`             | Public    | Email/password login                  |
| `/register`          | Public    | Account creation                      |
| `/dashboard`         | Protected | Overview with links to features       |
| `/dashboard/finance` | Protected | Monthly finance records + summary     |
| `/dashboard/kanban`  | Protected | Kanban board with columns and cards   |
