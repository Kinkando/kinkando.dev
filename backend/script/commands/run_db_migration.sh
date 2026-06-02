#!/usr/bin/env bash
# run_db_migration.sh
# Applies the most recent migration using dbmate.
# Run via: make run-db-migrations-windows or make run-db-migrations-macos
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

echo "→ Applying the most recent migration..."
cd "$ROOT_DIR"

dbmate \
  --url "$POSTGRES_DSN" \
  --migrations-dir './migrations' \
  --no-dump-schema \
  migrate

echo "✓ Migration complete."
