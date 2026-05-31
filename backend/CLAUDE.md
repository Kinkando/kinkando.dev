# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run the server
go run ./cmd/main.go

# Build
go build -o server ./cmd/main.go

# Build Docker image
docker build -t personal-dashboard .

# Run tests
go test ./...

# Run a single package's tests
go test ./internal/finance/...

# Lint (requires golangci-lint)
golangci-lint run
```

## Database Migrations

Migrations are managed with **[dbmate](https://github.com/amacneil/dbmate)** and live in `migrations/`.

### File format

Every migration file **must** contain both sections:

```sql
-- migrate:up
-- forward SQL here (CREATE TABLE, ALTER TABLE, CREATE INDEX …)

-- migrate:down
-- reverse SQL here (DROP TABLE, DROP INDEX …)
```

### Naming convention

```
NNN_<snake_case_description>.sql
```

Examples: `001_init.sql`, `002_add_currency_to_finance_records.sql`

### Checklist — when you add, edit, or remove a migration

1. **Add** a new numbered file in `migrations/` with `-- migrate:up` and `-- migrate:down` sections.  
   Never edit an already-applied migration; create a new one instead.
2. **Apply** the migration:
   ```bash
   make run-db-migrations
   ```
3. **Regenerate** the jet SQL builder so generated Go code matches the new schema:
   ```bash
   # macOS / Linux
   make gen-sql-builder-macos

   # Windows (Git Bash / WSL)
   make gen-sql-builder-windows
   ```
   This rewrites everything under `gen/<dbname>/public/`.
4. **Update** any affected repository code in `internal/*/repository/repo.go` to use the new/changed table or model types.
5. **Commit** all related files together in a single commit:
   - The new/changed migration file (`migrations/NNN_*.sql`)
   - The regenerated jet files (`gen/`)
   - Any updated repository, service, handler, or model code
   - Updated `CLAUDE.md` if the migration workflow section itself changed

### Rollback the most recent migration

```bash
make rollback-db-migration
```

After a rollback, re-run `make gen-sql-builder-*` so the generated code reflects the rolled-back schema, then update any affected repository code.

### Environment variables required

| Variable | Used by |
|---|---|
| `POSTGRES_DSN` | Go server (app runtime) |
| `POSTGRES_MIGRATION_URL` | dbmate (`make run-db-migrations`, `make rollback-db-migration`) |
| `POSTGRES_DSN` | jet codegen (`make gen-sql-builder-*`) |

Set both in `.env` (copy from `.env.example`).

## Environment Setup

Copy `.env.example` to `.env` and fill in values:

- `POSTGRES_DSN` — PostgreSQL DSN for the app server (usually with `sslmode=require`)
- `POSTGRES_MIGRATION_URL` — PostgreSQL DSN for dbmate migrations (can use `sslmode=disable` for direct Supabase connections)
- `MONGO_URI` / `MONGO_DB` — MongoDB Atlas URI and database name
- `FIREBASE_CREDENTIALS` — Firebase service-account JSON (inline, not a file path)
- `MCP_USER_FIREBASE_UID` — Firebase UID of the single user for the MCP server (see MCP server section below)

## Architecture

This is a Go backend for a personal dashboard. Entry point is `cmd/main.go`, which wires all dependencies and starts a [Fiber](https://gofiber.io/) HTTP server on `PORT` (default 8080).

### Module layout

Each feature lives under `internal/<feature>/` and follows a consistent three-layer pattern:

| Layer | Location | Responsibility |
|---|---|---|
| Handler | `internal/<feature>/handler/handler.go` | HTTP routing, request parsing, validation |
| Service | `internal/<feature>/service/service.go` | Business logic (currently thin pass-throughs) |
| Repository | `internal/<feature>/repository/repo.go` | Database queries |

Models/types are defined in `internal/<feature>/model.go`.

### Features

- **Finance** (`/api/v1/finance`) — income/expense records stored in **PostgreSQL** (`finance_records` table). All endpoints require auth. Supports CRUD on records and monthly summaries grouped by category.
- **Kanban** (`/api/v1/kanban`) — board/column/card data stored in **MongoDB** (collections: `boards`, `columns`, `cards`). Board is auto-created with default columns ("To Do", "In Progress", "Done") on first access per user.
- **Portfolio** (`/api/v1/portfolio`) — static data, no auth required, no database.

### MCP server

A second binary at `cmd/mcp` exposes kanban and finance operations as MCP tools over **stdio** transport. It shares the same `internal/` repositories and services as the HTTP server — no HTTP round-trip, no Firebase JWT needed at runtime.

Run it:
```bash
go run ./cmd/mcp
```

Build it:
```bash
go build -o mcp-server ./cmd/mcp
```

**Identity:** Set `MCP_USER_FIREBASE_UID` in `.env` to your Firebase UID. The binary resolves this once at startup to the internal `uuid.UUID` via the `users` table. The user must have signed into the web app at least once (which provisions the `users` row).

**Tools exposed (8 total):**
- Finance: `finance_list_records`, `finance_create_record`, `finance_delete_record`, `finance_monthly_summary`
- Kanban: `kanban_get_board`, `kanban_create_card`, `kanban_move_card`, `kanban_delete_card`

**Logging:** all log output goes to **stderr**; stdout is reserved for the MCP stdio protocol.

**Claude Code config** (example `~/.claude/mcp_servers.json` entry):
```json
{
  "kinkando": {
    "command": "go",
    "args": ["run", "./cmd/mcp"],
    "cwd": "/path/to/kinkando.dev/backend"
  }
}
```

### Authentication

Firebase Auth JWT verification via `internal/auth/middleware.go`. The `authMW.Require()` middleware validates the `Authorization: Bearer <id_token>` header, verifies it against Firebase, and stores the Firebase UID in `c.Locals("user_id")`. Use `auth.GetUserID(c)` in handlers to retrieve it.

Finance handlers additionally parse the UID as a `uuid.UUID` (requires a matching row in the `users` table); Kanban uses the raw Firebase UID string directly.

### Databases

- **PostgreSQL** (`pkg/postgres`) — `pgxpool` for connection management; `pgx/v5/stdlib` wraps the pool into a `*sql.DB` for jet. Schema lives in `migrations/`. Generated query builder code is in `gen/postgres/public/`.
- **MongoDB** (`pkg/mongo`) — used only by Kanban. No schema migrations needed.

### SQL query builder (go-jet/jet)

The finance repository uses [go-jet/jet](https://github.com/go-jet/jet) for type-safe SQL. The generated table and model structs live in:

```
gen/
└── <dbname>/          ← actual database name (e.g. "kinkando")
    └── public/
        ├── model/   ← Go structs mirroring each table row
        └── table/   ← Type-safe column and table references
```

These files are **generated** — do not edit them by hand. Regenerate with `make gen-sql-builder-*` after any schema change.

### Config

Loaded from environment (or `.env`) via `config/config.go` using `caarlos0/env`. All fields are required except `PORT` and `APP_ENV`.

### CORS

Allowed origins: `http://localhost:3000` and `https://*.pages.dev`.
