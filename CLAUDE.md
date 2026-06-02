# CLAUDE.md

## Project Structure

```
kinkando.dev/
├── backend/       # Go (Fiber v2) — API server
├── frontend/      # React + Vite (SPA) — web client
├── docs/          # Extended reference docs
└── plans/         # Gitignored working plans
```

## Docs

- [Documentation Index](docs/README.md)

## Modules

- [backend/CLAUDE.md](backend/CLAUDE.md)
- [frontend/CLAUDE.md](frontend/CLAUDE.md)

## Planning

Save finalized plans to `plans/` — gitignored, not committed.

## Refactoring

After completing any edit, check for opportunities to refactor and reuse existing functions or utilities rather than duplicating logic. This applies to both frontend and backend. If a newly written piece of logic already exists elsewhere, replace the duplicate with a call to the shared function.

## Committing

After completing any task, always create a git commit without waiting to be asked.
