# Frontend Directory Layout

```
src/
  App.tsx                         # Route definitions
  main.tsx                        # React entry point
  index.css                       # Tailwind import + global styles (animate-fade-in, etc.)
  vite-env.d.ts
  config/
    env.ts                        # import.meta.env.VITE_* → typed object
  lib/
    firebase.ts                   # Firebase app + auth singleton, getIdToken helper
    date.ts                       # Date utilities: todayDate, addDays, formatDate, formatTime, calculateAge
    format.ts                     # Generic formatters: formatCurrency (THB), formatNumber
    cn.ts                         # cn() — conditional className helper (no dependency)
    icons.ts                      # Icon registry: getIcon(), ICON_NAMES
    kanban.ts                     # Kanban label/meta maps (PRIORITY_META, ARCHIVE_REASON_META, etc.)
    workout.ts                    # Workout label maps (WORKOUT_TYPE_LABELS, SECTIONS, SECTION_LABELS, SECTION_LABELS_EMOJI)
    medicine.ts                   # Medicine label/badge/status maps (SOURCE_*, FREQUENCY_*, TIMING_*, STATUS_*, ADJ_*)
    health.ts                     # Health label maps (GENDER_LABELS, GOAL_LABELS, MEAL_LABELS, MEAL_TYPES, etc.)
    finance.ts                    # Finance label/meta maps (RECORD_TYPES, RECORD_TYPE_META)
    news.ts                       # News category style map (CATEGORY_STYLE)
    messaging.ts                  # Messaging helpers
    api/
      client.ts                   # apiFetch wrapper (envelope unwrap, Bearer auth, ApiError)
      types.ts                    # DTO TypeScript types (all string-literal unions live here)
      users.ts
      portfolio.ts
      finance.ts
      kanban.ts
      health.ts
      workout.ts
      medicine.ts
      news.ts
      quest.ts
      notifications.ts
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
    useWorkout.ts
    useMedicine.ts
    usePortfolio.ts
    useQuest.ts
    useNews.ts
    useUser.ts
    useNotifications.ts
  components/
    NavBar.tsx
    AuthForm.tsx
    InstallButton.tsx
    LoadingScreen.tsx
    TopLoader.tsx
    icons/                        # Brand SVG icon components (not in lucide-react)
      GitHubIcon.tsx
      GoogleIcon.tsx
    chat/                         # ChatInput, MessageBubble
    finance/                      # CategoryManager, RecordForm, RecordList, SummaryPanel
    health/                       # DashboardTab, FoodTab, ProgressTab, SettingsTab, SleepTab, bmi.ts
    kanban/                       # Board, BoardSwitcher, Card, CardModal, Column, FilterBar, StatsBar
    medicine/                     # AdjustStockDialog, HistoryTab, MedicineCard, MedicineFormDialog,
                                  # MedicineTracker, MedicinesTab, TakeDialog
    news/                         # FeaturedNewsCard, NewsCard, NewsCategoryTabs
    quest/                        # AchievementsTab, DashboardTab, HistoryTab, QuestActionsMenu,
                                  # QuestFormDialog, QuestRowMenu, QuestTab, StreaksSection, questConfig.ts
    today/                        # CaloriesCard, MedicineDueCard, QuestSummaryCard, SleepCard,
                                  # WeightCard, WorkoutTodayCard
    workout/                      # HistoryTab, PresetForm, PresetsTab, ScheduleTab, TodayTab
  pages/
    LoginPage.tsx
    RegisterPage.tsx
    AccountPage.tsx
    PortfolioPage.tsx
    TodayPage.tsx
    FinancePage.tsx
    HealthPage.tsx
    FoodPage.tsx
    SleepPage.tsx
    WorkoutPage.tsx
    MedicinePage.tsx
    SupplementPage.tsx
    QuestPage.tsx
    KanbanPage.tsx
    KanbanArchivePage.tsx
    NewsPage.tsx
    NotificationsPage.tsx
    ChatPage.tsx
```

## lib/ conventions

- **`src/lib/api/types.ts`** is the single source of truth for all DTO types and string-literal union types. Never define unions inline in components or lib files — import them.
- **`src/lib/<domain>.ts`** holds label maps and meta objects for that domain. Always type them as a **closed `Record<Union, …>`** (not `Record<string, …>`), so adding a new union value forces a compile error at the map sites.
- **`src/lib/date.ts`** — use `formatDate` / `formatTime` everywhere. Never inline an `Intl` formatter in a component.
- **`src/lib/format.ts`** — use `formatCurrency` / `formatNumber` everywhere. Never duplicate THB/number formatters.
- **`src/lib/cn.ts`** — use `cn()` for conditional className composition. No external dependency needed.

## components/icons/ convention

Brand icons that are unavailable in lucide-react (e.g. GitHub, Google) live in `src/components/icons/` as typed React components. Never inline `<svg>` markup for brand marks directly in pages or feature components.

## Backend contract

- Base: `<VITE_API_URL>/api/v1`
- Success envelope: `{ "data": ... }`; error: `{ "error": "..." }`
- `DELETE` and `PATCH .../move` return `204 No Content`
- Auth: `Authorization: Bearer <firebase id token>` on protected routes
- **User provisioning**: after sign-in, `POST /api/v1/users` must be called once per session before finance endpoints work. `AuthContext` handles this via `provisionUser()`.
