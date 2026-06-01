# Commit Policy

**Always commit when an edit is completed — no exceptions.**

1. Stage only files relevant to the change — never `git add -A` or `git add .`
2. Never commit secrets (`.env`, credentials, API keys, connection strings)
3. Commit message: start with `feat:` / `fix:` / `refactor:` / `docs:` / `chore:`, explain the *why* in 1–2 sentences
4. End every commit message with:
   ```
   Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
   ```
5. Do **not** push unless explicitly asked

## Schema-change commits

When a migration is added, the commit **must** include all of these together:

1. The migration file (`backend/migrations/NNN_*.sql`)
2. The regenerated jet code (`backend/gen/kinkando/public/`)
3. The updated repository / service / handler code

Run migration + codegen **before** committing — see [Backend Database & Migrations](../backend/database.md).
