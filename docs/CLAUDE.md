# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

Run from the repository root (pnpm workspace):

- `pnpm install` — Install dependencies
- `pnpm start` / `pnpm dev` — Start backend + frontend (Vite proxy to `:8080`)
- `pnpm build` — TypeScript type-check (`tsc -b`) then Vite production build
- `pnpm lint` — ESLint across all TS/TSX files
- `pnpm format` — Format all files with Prettier
- `pnpm test` — Run vitest once + backend unit tests
- `pnpm verify` — lint + test + build（PR 前快捷命令）
- `pnpm preview` — Serve the production build locally (frontend)

To run a single test file: `pnpm --filter @tokenjoy/frontend exec vitest run tests/lib/org.test.ts`

## Architecture

pnpm monorepo. Frontend app lives in `apps/frontend/`. Backend API lives in `apps/backend/`.

Single-page React app built with Vite 8, React 19, and TypeScript 6.

**Entry point:** `apps/frontend/src/main.tsx` → renders `<App />` into `#root`.

**Routing:** `react-router` v7 (imported from `'react-router'`, not `'react-router-dom'`). All routes are nested under `<AdminLayout />` which provides sidebar + header + `<Outlet />`. Route page components live in `apps/frontend/src/routes/org/`.

**Styling:** TailwindCSS v4 via the `@tailwindcss/vite` plugin. CSS-first configuration in `apps/frontend/src/index.css` — no separate tailwind.config file.

**UI components:** shadcn/ui primitives in `apps/frontend/src/components/ui/` (uses `class-variance-authority`, `tailwind-merge`, `lucide-react` icons). The `cn()` utility in `apps/frontend/src/lib/utils.ts` merges class names.

**State management:** Zustand v5 — stores are co-located with the features that use them (no central store directory).

**Data fetching:** Custom fetch wrapper in `apps/frontend/src/api/client.ts` with a `request<T>()` generic function (base URL: `/api`). Domain-specific API methods are in `apps/frontend/src/api/` (`org.ts`, `keys.ts`, `budget.ts`, `models.ts`, `dashboard.ts`, `audit.ts`, `session.ts`), organized as namespaced objects (`dataSourceApi`, `syncApi`, `departmentApi`, `memberApi`, `roleApi`, etc.). TanStack Query via `features/query`. API contract: `docs/Frontend-API契约.md`.

**Local backend proxy:** `apps/frontend/.env.development` sets `VITE_API_PROXY_TARGET=http://localhost:8080`. Root `pnpm start` runs backend and frontend together.

**Testing:** Vitest + `@testing-library/react` + jsdom. Setup file: `apps/frontend/tests/setup.ts`. Tests use `createMockApis()` for API injection. All specs live under `apps/frontend/tests/` (mirror `src/` paths); use `@tests/utils` for providers and `createMockApis`.

**Path alias:** `@/*` resolves to `./src/*` (configured in both `apps/frontend/vite.config.ts` and `apps/frontend/tsconfig.app.json`).

**TypeScript config:** Project references — `tsconfig.app.json` (app code, ES2023, bundler resolution, strict) and `tsconfig.node.json` (tooling). Strict rules: `noUnusedLocals`, `noUnusedParameters`, `noFallthroughCasesInSwitch`, `erasableSyntaxOnly`.

**ESLint:** Flat config (ESLint 10) with typescript-eslint, react-hooks, and react-refresh plugins.
