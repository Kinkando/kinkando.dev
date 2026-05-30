#!/usr/bin/env bash
# rollback_db_migration.sh
# Rolls back the most recently applied migration using dbmate.
# Run via: make rollback-db-migration
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

if [ -z "${POSTGRES_MIGRATION_URL:-}" ]; then
  echo "Error: POSTGRES_MIGRATION_URL is not set. Add it to .env or export it before running." >&2
  exit 1
fi

echo "→ Rolling back the most recent migration..."
cd "$ROOT_DIR"

dbmate \
  --url "$POSTGRES_MIGRATION_URL" \
  --migrations-dir './migrations' \
  --no-dump-schema \
  rollback

echo "✓ Rollback complete."
