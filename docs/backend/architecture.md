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

## Features

| Feature | Route prefix | Storage | Notes |
|---|---|---|---|
| Finance | `/api/v1/finance` | PostgreSQL | CRUD records + monthly summaries by category |
| Kanban | `/api/v1/kanban` | MongoDB | Board auto-created on first access per user |
| Health | `/api/v1/health` | PostgreSQL | Weight logs, food logs, sleep logs, profiles |
| AI Chat | `/api/v1/aichat` | — | Powered by Gemini; see `internal/gemini/` |
| LINE | `/api/v1/line` | — | LINE Messaging API webhook |
| Portfolio | `/api/v1/portfolio` | — | Static data, no auth |
| User | `/api/v1/users` | PostgreSQL | User provisioning |

## Authentication

Firebase Auth JWT via `internal/auth/middleware.go`. `authMW.Require()` validates `Authorization: Bearer <id_token>`, stores Firebase UID in `c.Locals("user_id")`. Use `auth.GetUserID(c)` in handlers.

Finance handlers parse the UID as `uuid.UUID` (requires a `users` row); Kanban uses the raw Firebase UID string.

## Databases

- **PostgreSQL** (`pkg/postgres`) — `pgxpool`; `pgx/v5/stdlib` wraps for jet. Schema in `migrations/`. Generated code in `gen/kinkando/public/`.
- **MongoDB** (`pkg/mongo`) — Kanban only. No migrations.

## SQL query builder (go-jet/jet)

Finance (and Health) repositories use [go-jet/jet](https://github.com/go-jet/jet) for type-safe SQL. Generated structs:

```
gen/
└── kinkando/
    └── public/
        ├── model/   ← Go structs mirroring each table row
        └── table/   ← Type-safe column and table references
```

Do not edit these by hand. Regenerate with `make gen-sql-builder-*` after schema changes.

## MCP server

Served at `POST /mcp` (Streamable HTTP, JSON). No separate process.

**Enabled when** both `MCP_USER_FIREBASE_UID` and `MCP_AUTH_TOKEN` are set. Otherwise `/mcp` returns 404.

**Auth:** `Authorization: Bearer <MCP_AUTH_TOKEN>`

**Tools (32):**
- Finance: `finance_list_records`, `finance_list_categories`, `finance_create_category`, `finance_delete_category`, `finance_create_record`, `finance_delete_record`, `finance_monthly_summary`
- Kanban: `kanban_get_board`, `kanban_create_card`, `kanban_update_card`, `kanban_move_card`, `kanban_delete_card`, `kanban_board_stats`, `kanban_archive_card`, `kanban_unarchive_card`, `kanban_list_archived_cards`
- Workout: `workout_list_sessions`, `workout_list_presets`, `workout_get_preset`, `workout_get_schedule`, `workout_create_preset`, `workout_update_preset`, `workout_delete_preset`, `workout_start_session`, `workout_update_session`, `workout_log_exercise`, `workout_add_exercise`
- Food: `food_list_logs`, `food_log_meal`, `food_update_meal`, `food_delete_meal`
- Sleep: `sleep_list_logs`, `sleep_log_night`, `sleep_update_night`, `sleep_delete_night`
- Medicine: `medicine_list`, `medicine_take`, `medicine_adjust_stock`, `medicine_list_intakes`, `medicine_list_stock_adjustments`

**Config example:**
```json
{
  "kinkando": {
    "type": "http",
    "url": "http://localhost:8080/mcp",
    "headers": { "Authorization": "Bearer <MCP_AUTH_TOKEN>" }
  }
}
```

## Config

Loaded from `.env` via `config/config.go` using `caarlos0/env`. All fields required except `PORT` and `APP_ENV`.

## CORS

Allowed origins: `http://localhost:3000` and `https://*.pages.dev`.
