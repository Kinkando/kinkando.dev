# Frontend Directory Layout

```
src/
  App.tsx                         # Route definitions
  main.tsx                        # React entry point
  index.css                       # Tailwind import + global styles
  vite-env.d.ts
  config/
    env.ts                        # import.meta.env.VITE_* → typed object
  lib/
    firebase.ts                   # Firebase app + auth singleton, getIdToken helper
    date.ts                       # Date utilities
    icons.ts                      # Icon registry
    kanban.ts                     # Kanban utilities
    news.ts                       # News mock data
    api/
      client.ts                   # apiFetch wrapper (envelope unwrap, Bearer auth, ApiError)
      types.ts                    # DTO TypeScript types
      users.ts
      portfolio.ts
      finance.ts
      kanban.ts
      health.ts
      chat.ts
  auth/
    AuthContext.tsx                # onAuthStateChanged, ensureUserOnce, useAuth()
    ProtectedRoute.tsx             # redirects to /login?redirect=... if unauthenticated
  hooks/
    useDocumentTitle.ts
    useIsMobile.ts
  queries/                        # TanStack Query hooks
    keys.ts
    useFinance.ts
    useKanban.ts
    useHealth.ts
    usePortfolio.ts
  components/
    NavBar.tsx
    AuthForm.tsx
    chat/                         # ChatInput, MessageBubble
    finance/                      # CategoryManager, RecordForm, RecordList, SummaryPanel
    health/                       # DashboardTab, ExerciseTab, ProgressTab, SettingsTab, bmi.ts
    kanban/                       # Board, BoardSwitcher, Card, CardModal, Column, FilterBar, StatsBar
    news/                         # FeaturedNewsCard, NewsCard, NewsCategoryTabs
  pages/
    LoginPage.tsx
    RegisterPage.tsx
    PortfolioPage.tsx
    FinancePage.tsx
    KanbanPage.tsx
    KanbanArchivePage.tsx
    HealthPage.tsx
    ChatPage.tsx
    NewsPage.tsx
```

## Backend contract

- Base: `<VITE_API_URL>/api/v1`
- Success envelope: `{ "data": ... }`; error: `{ "error": "..." }`
- `DELETE` and `PATCH .../move` return `204 No Content`
- Auth: `Authorization: Bearer <firebase id token>` on protected routes
- **User provisioning**: after sign-in, `POST /api/v1/users` must be called once per session before finance endpoints work. `AuthContext` handles this via `provisionUser()`.
