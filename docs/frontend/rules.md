# Frontend Rules

## Environment & configuration

- Never read from `process.env`; always use `import.meta.env.VITE_*` via `src/config/env.ts`
- Never commit `.env.local`
- Port 3000 is required by backend CORS — do not change `server.port` in `vite.config.ts`

## Types

- `FinanceRecord` (not `Record`) to avoid collision with the TS builtin
- For domain fields with a fixed set of valid string values, define a **TypeScript string-literal union** in `src/lib/api/types.ts` (e.g. `type MealType = 'breakfast' | 'lunch' | 'dinner' | 'snack'`). Never use plain `string` for these fields in DTOs or component props.

## Enum label/meta maps

- Centralized label maps live in **`src/lib/<domain>.ts`** (e.g. `workout.ts`, `medicine.ts`, `health.ts`, `finance.ts`). Import from there; never define a `Record<…, string>` map inline inside a component.
- Always type maps as a **closed `Record<Union, …>`** — not `Record<string, …>`. This makes the compiler flag a missing entry when the union is widened.
- If a domain lib file does not yet exist for a new union, create it in `src/lib/` following the existing pattern.

## Formatters

- Use **`formatDate` / `formatTime`** from `src/lib/date.ts` for all date/time presentation. Never inline an `Intl.DateTimeFormat` or `toLocaleDateString` call in a component. Pass the optional `opts` parameter for non-default formats (e.g. `{ month: 'short', day: 'numeric' }`).
- Use **`formatCurrency` / `formatNumber`** from `src/lib/format.ts` for all currency and large-number display.

## className composition

- Use **`cn()`** from `src/lib/cn.ts` to compose conditional class strings. Prefer it over template literals with ternaries for any `className` with more than one conditional branch.
- Every `<button>` must include `cursor-pointer` in its `className`. For static strings use a literal class; for dynamic strings (`cn(...)`) include it in the base classes. Exception: `cursor-grab` / `cursor-not-allowed` can replace `cursor-pointer` when the semantic is clearly different (e.g. drag handles, permanently disabled controls).

## Icons

- Utility icons (chevrons, close, download, etc.) come from **lucide-react**. Import the named component; do not inline SVGs for icons that exist in lucide-react.
- Brand icons not available in lucide-react (GitHub, Google, etc.) live in **`src/components/icons/`** as typed React components (`className?: string` prop, `fill="currentColor"`). Never inline `<svg>` markup for brand marks directly in pages or feature components.

## Component layout

- Reuse existing components where possible.
- Bottom action buttons in dialogs, modals, and forms (Save/Cancel, Delete/Cancel, etc.) must be right-aligned: wrap them in `<div className="flex justify-end gap-2">`. Do not use `flex-1` on those buttons; size them by their content with `px-4 py-2` (or `px-4 py-1.5` for compact forms).
