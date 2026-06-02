## ── Database migrations ───────────────────────────────────────────────────────
run-db-migrations-macos:
	. ./script/commands/run_db_migration.sh

run-db-migrations-windows:
	bash script/commands/run_db_migration.sh

rollback-db-migration-macos:
	. ./script/commands/rollback_db_migration.sh

rollback-db-migration-windows:
	bash script/commands/rollback_db_migration.sh

## ── SQL builder code generation ──────────────────────────────────────────────
gen-sql-builder-macos:
	. ./script/commands/gen_sql_builder.sh

gen-sql-builder-windows:
	bash script/commands/gen_sql_builder.sh

.PHONY: run-db-migrations-windows run-db-migrations-macos rollback-db-migration-macos rollback-db-migration-windows gen-sql-builder-macos gen-sql-builder-windows
