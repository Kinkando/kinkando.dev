# Frontend Rules

- Never read from `process.env`; always use `import.meta.env.VITE_*` via `src/config/env.ts`
- Never commit `.env.local`
- Port 3000 is required by backend CORS — do not change `server.port` in `vite.config.ts`
- `FinanceRecord` (not `Record`) to avoid collision with the TS builtin
- Reuse existing components where possible
- Every `<button>` must include `cursor-pointer` in its `className`. For static strings use a literal class; for dynamic strings (`className={\`...\`}`) include it in the template literal. Exception: `cursor-grab` / `cursor-not-allowed` can replace `cursor-pointer` when the semantic is clearly different (e.g. drag handles, permanently disabled controls).
