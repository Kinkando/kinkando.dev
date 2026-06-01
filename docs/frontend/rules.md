# Frontend Rules

- Never read from `process.env`; always use `import.meta.env.VITE_*` via `src/config/env.ts`
- Never commit `.env.local`
- Port 3000 is required by backend CORS — do not change `server.port` in `vite.config.ts`
- `FinanceRecord` (not `Record`) to avoid collision with the TS builtin
- Reuse existing components where possible
