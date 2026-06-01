# CLAUDE.md

## Project Structure

```
kinkando.dev/
├── backend/       # Go (Fiber v2) — API server
├── frontend/      # React + Vite (SPA) — web client
├── docs/          # Extended reference docs
└── plans/         # Gitignored working plans
```

## Docs Index

- [Commit Policy](docs/commit-policy.md)
- [Git Conventions](docs/git-conventions.md)
- [Response Style](docs/response-style.md)
- [Backend Commands](docs/backend-commands.md)
- [Backend Architecture](docs/backend-architecture.md)
- [Backend Database & Migrations](docs/backend-database.md)
- [Backend Environment Variables](docs/backend-env.md)
- [Frontend Stack & Dev Server](docs/frontend-stack.md)
- [Frontend Environment Variables](docs/frontend-env.md)
- [Frontend Directory Layout](docs/frontend-structure.md)
- [Frontend Rules](docs/frontend-rules.md)

## Module CLAUDE.md files

- [backend/CLAUDE.md](backend/CLAUDE.md)
- [frontend/CLAUDE.md](frontend/CLAUDE.md)

## After every edit — mandatory

**1. Schema changes** (any new or modified migration):

```bash
cd backend
set -a && source .env && set +a          # load POSTGRES_DSN (needed if POSTGRES_MIGRATION_URL absent)
make run-db-migrations DB_MIGRATION_URL="$POSTGRES_DSN"
make gen-sql-builder-windows             # or gen-sql-builder-macos on macOS/Linux
```

Stage migration + `gen/` + code together in one commit.

**2. Commit** (every edit, not just migrations):

Stage only relevant files. Message format: `type: why` + co-author line.  
See [Commit Policy](docs/commit-policy.md) for the full rules.

## Planning

Save finalized plans to `plans/` — gitignored, not committed.
