# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

Run from the repository root (pnpm workspace):

- `pnpm install` — Install dependencies
- `pnpm start` / `pnpm dev` — Start Vite dev server with HMR
- `pnpm build` — TypeScript type-check (`tsc -b`) then Vite production build
- `pnpm lint` — ESLint across all TS/TSX files
- `pnpm format` — Format all files with Prettier
- `pnpm test` — Run vitest in watch mode
- `pnpm test:run` — Run vitest once (CI-friendly)
- `pnpm preview` — Serve the production build locally

To run a single test file: `pnpm --filter @tokenjoy/frontend exec vitest run tests/lib/org.test.ts`

## Architecture

pnpm monorepo. Frontend app lives in `apps/frontend/`.

Single-page React app built with Vite 8, React 19, and TypeScript 6.

**Entry point:** `apps/frontend/src/main.tsx` → starts MSW service worker when `USE_MOCKS` is true (`apps/frontend/src/config/app.ts`), then renders `<App />` into `#root`.

**Routing:** `react-router` v7 (imported from `'react-router'`, not `'react-router-dom'`). All routes are nested under `<AdminLayout />` which provides sidebar + header + `<Outlet />`. Route page components live in `apps/frontend/src/routes/org/`.

**Styling:** TailwindCSS v4 via the `@tailwindcss/vite` plugin. CSS-first configuration in `apps/frontend/src/index.css` — no separate tailwind.config file.

**UI components:** shadcn/ui primitives in `apps/frontend/src/components/ui/` (uses `class-variance-authority`, `tailwind-merge`, `lucide-react` icons). The `cn()` utility in `apps/frontend/src/lib/utils.ts` merges class names.

**State management:** Zustand v5 — stores are co-located with the features that use them (no central store directory).

**Data fetching:** Custom fetch wrapper in `apps/frontend/src/api/client.ts` with a `request<T>()` generic function (base URL: `/api`). Domain-specific API methods are in `apps/frontend/src/api/` (`org.ts`, `keys.ts`, `budget.ts`, `models.ts`, `dashboard.ts`, `audit.ts`, `session.ts`), organized as namespaced objects (`dataSourceApi`, `syncApi`, `departmentApi`, `memberApi`, `roleApi`, etc.). No react-query/SWR — fetches happen directly in effects or event handlers. API contract: `docs/Frontend-API契约.md`.

**API mocking:** MSW v2 intercepts `/api/*` requests when `USE_MOCKS` is true. Handlers in `apps/frontend/src/mocks/handlers/` (aggregated by `handlers/index.ts`), fixtures in `apps/frontend/src/mocks/fixtures/` (re-exported via `mocks/data.ts`). Local backend proxy: set `VITE_API_PROXY_TARGET` in `vite.config.ts`.

**Testing:** Vitest + `@testing-library/react` + jsdom. Setup file: `apps/frontend/tests/setup.ts`. MSW is available for mocking API calls in tests. All specs live under `apps/frontend/tests/` (mirror `src/` paths); use `@tests/utils` for providers and `createMockApis`.

**Path alias:** `@/*` resolves to `./src/*` (configured in both `apps/frontend/vite.config.ts` and `apps/frontend/tsconfig.app.json`).

**TypeScript config:** Project references — `tsconfig.app.json` (app code, ES2023, bundler resolution, strict) and `tsconfig.node.json` (tooling). Strict rules: `noUnusedLocals`, `noUnusedParameters`, `noFallthroughCasesInSwitch`, `erasableSyntaxOnly`.

**ESLint:** Flat config (ESLint 10) with typescript-eslint, react-hooks, and react-refresh plugins.
