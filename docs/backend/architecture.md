# Backend Architecture

Entry point: `cmd/main.go` — wires all dependencies, starts [Fiber](https://gofiber.io/) on `PORT` (default 8080).

## Module layout

Each feature lives under `internal/<feature>/` with a consistent three-layer pattern:

| Layer | Location | Responsibility |
|---|---|---|
| Handler | `internal/<feature>/handler/handler.go` | HTTP routing, request parsing, validation |
| Service | `internal/<feature>/service/service.go` | Business logic |
| Repository | `internal/<feature>/repository/repo.go` | Database queries |

Models/types: `internal/<feature>/model.go`

**Enum convention**: for any field with a fixed set of valid string values, declare a named type and a `const` block — never use bare untyped string constants:

```go
// correct
type MealType string

const (
    MealTypeBreakfast MealType = "breakfast"
    MealTypeLunch     MealType = "lunch"
)

// wrong — no type wrapper
const MealTypeBreakfast = "breakfast"
```

Request structs use `validate` struct tags with `github.com/go-playground/validator/v10`. Enum fields use `oneof` to constrain valid values — no separate `ValidX` helper functions:

```go
type CreateFoodLogInput struct {
    MealType MealType `json:"meal_type" validate:"required,oneof=breakfast lunch dinner snack"`
}
```

The matching frontend types in `src/lib/api/types.ts` must be TypeScript string literal unions (e.g. `type MealType = 'breakfast' | 'lunch'`), not plain `string`.

## Features

| Feature | Route prefix | Storage | Notes |
|---|---|---|---|
| Finance | `/api/v1/finance` | PostgreSQL | CRUD records + monthly summaries by category |
| Kanban | `/api/v1/kanban` | MongoDB | Board auto-created on first access per user |
| Health | `/api/v1/health` | PostgreSQL | Weight logs, food logs, sleep logs, profiles |
| Medicine | `/api/v1/medicines` | PostgreSQL | Stock, intake tracking, dose reminders |
| Quest | `/api/v1/quest` | PostgreSQL | Daily/weekly quests with XP rewards |
| Workout | `/api/v1/workout` | PostgreSQL | Sessions, presets, schedule |
| AI Chat | `/api/v1/ai-chat` | — | Powered by Gemini; see `internal/gemini/` |
| Notification | `/api/v1/notifications` | PostgreSQL | Per-user settings (LINE/Discord/Web Push) + FCM tokens |
| LINE | `/api/v1/line` | — | LINE Messaging API webhook |
| Portfolio | `/api/v1/portfolio` | — | Static data, no auth |
| User | `/api/v1/users` | PostgreSQL | User provisioning |
| **Cron** | `/api/v1/cron` | — | Batch reminder jobs; auth via `X-Cron-Secret` (not Firebase) |

Finance handlers parse the UID as `uuid.UUID` (requires a `users` row); Kanban uses the raw Firebase UID string.

## Databases

- **PostgreSQL** (`pkg/postgres`) — `pgxpool`; `pgx/v5/stdlib` wraps for jet. Schema in `migrations/`. Generated code in `gen/kinkando/public/`.
- **MongoDB** (`pkg/mongo`) — Kanban only. No migrations.

## SQL query builder (go-jet/jet)

All repositories use [go-jet/jet](https://github.com/go-jet/jet) for type-safe SQL. Generated structs:

```
gen/
└── kinkando/
    └── public/
        ├── model/   ← Go structs mirroring each table row
        └── table/   ← Type-safe column and table references
```

**Do not edit these by hand.** Regenerate with `make gen-sql-builder-*` after every schema change (run without asking — see [database.md](database.md)). Raw SQL strings (`db.QueryContext` with a plain string) are forbidden for any table with a generated Jet file — see [rules.md](rules.md#sql-queries--jet-only).

## Authentication

Two auth mechanisms — choose based on the caller:

| Mechanism | Middleware | Caller |
|---|---|---|
| Firebase JWT | `authMW.Require()` | Web/mobile app users |
| Shared secret | `middleware.CronAuth(cfg.CronSecret)` | External cron worker (`X-Cron-Secret` header) |

Firebase handlers use `auth.GetUserID(c)` to resolve the internal UUID from the Firebase UID.

## MCP server

Served at `POST /mcp` (Streamable HTTP, JSON). No separate process. Always mounted.

**Auth:** `Authorization: Bearer <firebase-id-token>` — the same Firebase ID token used by the web app. Each request acts as the user the token belongs to.

**Multi-user:** a single shared MCP server serves all users. A **receiving middleware** (`mcpserver.userMiddleware`) resolves the Firebase UID per request and injects the app user into `context`. Tool handlers read identity from `ctx` — no per-user server instances.

**LINE users:** the LINE webhook resolves the sender's `line_id` to the matching app user, so AI replies act as that user. Unlinked accounts receive link instructions.

**Tools (32):**
- Finance: `finance_list_records`, `finance_list_categories`, `finance_create_category`, `finance_delete_category`, `finance_create_record`, `finance_delete_record`, `finance_monthly_summary`
- Kanban: `kanban_get_board`, `kanban_create_card`, `kanban_update_card`, `kanban_move_card`, `kanban_delete_card`, `kanban_board_stats`, `kanban_archive_card`, `kanban_unarchive_card`, `kanban_list_archived_cards`
- Workout: `workout_list_sessions`, `workout_list_presets`, `workout_get_preset`, `workout_get_schedule`, `workout_create_preset`, `workout_update_preset`, `workout_delete_preset`, `workout_start_session`, `workout_update_session`, `workout_log_exercise`, `workout_add_exercise`
- Food: `food_list_logs`, `food_log_meal`, `food_update_meal`, `food_delete_meal`
- Sleep: `sleep_list_logs`, `sleep_log_night`, `sleep_update_night`, `sleep_delete_night`
- Medicine: `medicine_list`, `medicine_take`, `medicine_adjust_stock`, `medicine_list_intakes`, `medicine_list_stock_adjustments`

**Config example (Claude Code or any MCP client):**
```json
{
  "kinkando": {
    "type": "http",
    "url": "http://localhost:8080/mcp",
    "headers": { "Authorization": "Bearer <firebase-id-token>" }
  }
}
```

## Config

Loaded from `.env` via `config/config.go` using `caarlos0/env`. All fields required except `PORT` and `APP_ENV`.

## CORS

Allowed origins: `http://localhost:3000` and `https://*.pages.dev`.
