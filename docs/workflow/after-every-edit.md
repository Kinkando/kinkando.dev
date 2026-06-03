# After Every Edit — Mandatory

## 1. Schema changes (any new or modified migration)

```bash
cd backend
set -a && source .env && set +a          # load POSTGRES_DSN (needed if POSTGRES_MIGRATION_URL absent)
make run-db-migrations-windows
make gen-sql-builder-windows             # or gen-sql-builder-macos on macOS/Linux
```

**Never edit `gen/` files by hand.** They are overwritten entirely on every codegen run — manual edits will be lost and create drift between the real schema and the generated types. If running the migration requires a permission the agent doesn't have, stop and ask the user to run both commands before writing any repository code.

Stage migration + `gen/` + code together in one commit.

## 2. Refactor / reuse

Check for duplicate logic. If the newly written code already exists elsewhere, replace it with a call to the shared function. Applies to both frontend and backend.

## 3. Commit (every edit, not just migrations) — **non-negotiable**

> **Do not finish a turn without committing.** This applies after plan-mode
> implementation, bug fixes, style tweaks — everything. No exceptions, no
> waiting to be asked.

Stage only relevant files. Message format: `type: why` + co-author line.
See [Commit Policy](commit-policy.md) for the full rules.
