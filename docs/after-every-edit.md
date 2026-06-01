# After Every Edit — Mandatory

## 1. Schema changes (any new or modified migration)

```bash
cd backend
set -a && source .env && set +a          # load POSTGRES_DSN (needed if POSTGRES_MIGRATION_URL absent)
make run-db-migrations DB_MIGRATION_URL="$POSTGRES_DSN"
make gen-sql-builder-windows             # or gen-sql-builder-macos on macOS/Linux
```

Stage migration + `gen/` + code together in one commit.

## 2. Commit (every edit, not just migrations)

Stage only relevant files. Message format: `type: why` + co-author line.
See [Commit Policy](commit-policy.md) for the full rules.
