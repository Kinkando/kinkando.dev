# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Structure

```
kinkando.dev/
├── backend/       # Go (Fiber v2) — API server
├── frontend/      # Next.js 14 (App Router) — web client
└── CLAUDE.md      # This file (root-level instructions)
```

See `backend/CLAUDE.md` and `frontend/CLAUDE.md` for module-specific guidance.

## Commit Policy

**Always commit when an edit is completed.** After finishing any task that changes files:

1. Stage only the files relevant to the change — never use `git add -A` or `git add .`
2. Never commit files that contain secrets (`.env`, credentials, API keys, connection strings)
3. Write a concise commit message: start with a verb (`feat:`, `fix:`, `refactor:`, `docs:`, `chore:`), explain the *why* in 1-2 sentences
4. End every commit message with:
   ```
   Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
   ```
5. Do **not** push unless explicitly asked

## Git Conventions

- Branch naming: `feat/<name>`, `fix/<name>`, `chore/<name>`
- One logical change per commit — don't bundle unrelated changes
- Never amend a commit unless explicitly asked
- Never force push to `main`
