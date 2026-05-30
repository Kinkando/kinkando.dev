## ── Variables ────────────────────────────────────────────────────────────────
## Override DB_MIGRATION_URL on the command line or via .env if needed.
## Example: make run-db-migrations DB_MIGRATION_URL=postgresql://...
DB_MIGRATION_URL ?= $(POSTGRES_MIGRATION_URL)
MIGRATIONS_DIR   ?= ./migrations

## ── Database migrations ───────────────────────────────────────────────────────
run-db-migrations:
	dbmate --url $(DB_MIGRATION_URL) --migrations-dir $(MIGRATIONS_DIR) --no-dump-schema migrate

rollback-db-migration:
	dbmate --url $(DB_MIGRATION_URL) --migrations-dir $(MIGRATIONS_DIR) --no-dump-schema rollback

## ── SQL builder code generation ──────────────────────────────────────────────
gen-sql-builder-macos:
	. ./script/commands/gen_sql_builder.sh

gen-sql-builder-windows:
	bash script/commands/gen_sql_builder.sh

.PHONY: run-db-migrations rollback-db-migration gen-sql-builder-macos gen-sql-builder-windows
