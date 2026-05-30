#!/usr/bin/env bash
# gen_sql_builder.sh
# Generates go-jet/jet type-safe query builder code from the live PostgreSQL schema.
# Run via: make gen-sql-builder-macos  OR  make gen-sql-builder-windows
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Load .env from project root if present
ENV_FILE="$ROOT_DIR/.env"
if [ -f "$ENV_FILE" ]; then
  set -o allexport
  # shellcheck disable=SC1090
  source "$ENV_FILE"
  set +o allexport
fi

if [ -z "${POSTGRES_DSN:-}" ]; then
  echo "Error: POSTGRES_DSN is not set. Add it to .env or export it before running." >&2
  exit 1
fi

echo "→ Generating jet SQL builder from schema (schema: public)..."
cd "$ROOT_DIR"

go run github.com/go-jet/jet/v2/cmd/jet@latest \
  -dsn="$POSTGRES_DSN" \
  -schema=public \
  -path=./gen

echo "✓ Done. Files written to ./gen/<dbname>/public/"
