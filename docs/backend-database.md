# Backend Database & Migrations

## Schema Change Rule

**Every time you modify the PostgreSQL schema**, run these two commands before writing any repository code:

```bash
make run-db-migrations        # apply the migration
make gen-sql-builder-windows  # regenerate gen/kinkando/public/ (use gen-sql-builder-macos on macOS/Linux)
```

Skipping codegen causes `undefined: table.XxxYyy` build errors.

## Migration tool: dbmate

Migrations live in `migrations/`. Every file **must** contain both sections:

```sql
-- migrate:up
-- forward SQL (CREATE TABLE, ALTER TABLE, CREATE INDEX …)

-- migrate:down
-- reverse SQL (DROP TABLE, DROP INDEX …)
```

### Naming

```
NNN_<snake_case_description>.sql
```

Examples: `001_init.sql`, `002_add_currency_to_finance_records.sql`

## Checklist — adding/editing a migration

1. Add a new numbered file in `migrations/` with both `migrate:up` and `migrate:down`. Never edit an already-applied migration.
2. Apply: `make run-db-migrations`
3. Regenerate jet:
   ```bash
   make gen-sql-builder-macos    # macOS / Linux
   make gen-sql-builder-windows  # Windows
   ```
4. Update affected `internal/*/repository/repo.go` files.
5. Commit all together: migration file, `gen/`, updated repository/service/handler code.

## Rollback

```bash
make rollback-db-migration
```

After rollback, re-run `make gen-sql-builder-*` and update affected repository code.

## Environment variables

| Variable | Used by |
|---|---|
| `POSTGRES_DSN` | Go server (runtime) and jet codegen |
| `POSTGRES_MIGRATION_URL` | dbmate (`make run-db-migrations`, `make rollback-db-migration`) |
