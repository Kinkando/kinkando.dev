# Frontend Rules

- Never read from `process.env`; always use `import.meta.env.VITE_*` via `src/config/env.ts`
- Never commit `.env.local`
- Port 3000 is required by backend CORS — do not change `server.port` in `vite.config.ts`
- `FinanceRecord` (not `Record`) to avoid collision with the TS builtin
- Reuse existing components where possible
- Every `<button>` must include `cursor-pointer` in its `className`. For static strings use a literal class; for dynamic strings (`className={\`...\`}`) include it in the template literal. Exception: `cursor-grab` / `cursor-not-allowed` can replace `cursor-pointer` when the semantic is clearly different (e.g. drag handles, permanently disabled controls).
- Bottom action buttons in dialogs, modals, and forms (Save/Cancel, Delete/Cancel, etc.) must be right-aligned: wrap them in `<div className="flex justify-end gap-2">`. Do not use `flex-1` on those buttons; size them by their content with `px-4 py-2` (or `px-4 py-1.5` for compact forms).
- For domain fields with a fixed set of valid string values, define a TypeScript string literal union in `src/lib/api/types.ts` (e.g. `type MealType = 'breakfast' | 'lunch' | 'dinner' | 'snack'`). Never use plain `string` for these fields in DTOs or component props.
