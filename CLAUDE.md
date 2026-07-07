# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

All commands run from the repo root (pnpm workspace, `pnpm@9.15.9`):

```bash
# Full-stack
pnpm start              # Backend + frontend together (waits for backend healthz)
pnpm test               # All tests (frontend + backend unit)
pnpm lint               # ESLint (frontend) + go vet/gofmt (backend)
pnpm format             # Prettier + gofmt

# Frontend (apps/frontend)
pnpm start:frontend     # Vite dev server with HMR
pnpm build              # tsc -b && vite build
pnpm test:frontend      # vitest run (once)

# Single frontend test:
pnpm --filter @tokenjoy/frontend exec vitest run tests/lib/org.test.ts

# Backend (apps/backend)
pnpm start:backend      # go run ./cmd/server
pnpm test:backend       # go test ./tests/...
pnpm test:backend:integration  # go test -tags=integration ./tests/store/postgres/...
pnpm lint:backend       # go vet + gofmt check

# Relay (apps/newapi)
pnpm start:relay        # docker compose up
pnpm gate:verify        # apps/newapi/scripts/gate-verify.sh
```

## Architecture

pnpm monorepo with three apps under `apps/`:

### Frontend (`apps/frontend/`)

React 19 SPA — Vite 8, TypeScript 6, TailwindCSS v4 (CSS-first, no tailwind.config).

- **Routing:** react-router v7 (import from `'react-router'`, not `'react-router-dom'`). Routes nested under `<AdminLayout />`.
- **State:** Zustand v5 stores co-located with features.
- **UI:** shadcn/ui primitives in `components/ui/`, Radix primitives, lucide-react icons. `cn()` from `lib/utils.ts`.
- **API layer:** Custom fetch in `api/client.ts` (`/api` base). Domain namespaces in `api/*.ts`. No react-query.
- **Mocking:** MSW v2 when `VITE_ENABLE_MOCKS=true`. Handlers in `mocks/handlers/`, fixtures in `mocks/fixtures/`.
- **Testing:** Vitest + @testing-library/react + jsdom. Tests in `apps/frontend/tests/`. Use `createMockApis()` + `renderHookWithProviders` from `@tests/utils`.
- **Path alias:** `@/*` → `./src/*`

Key conventions (from `.cursor/rules/frontend-structure.mdc`):
- Route pages: `routes/{domain}/{page}.tsx` — compose only, delegate to `hooks/use-{page}-page.ts`
- Page hooks use `useInjectedApis(injectedApis?)` for testability; other code uses `useApis()`
- Shared domain UI: `components/{domain}/` (2+ consumers); page-only UI: `routes/{domain}/components/`
- Routes defined in `config/routes.ts` via `ROUTE_DEFINITIONS` (single source of truth)
- Never import API functions directly in business code — go through the DI layer

### Backend (`apps/backend/`)

Go 1.24 service — chi router, PostgreSQL (pgx v5), env config (caarlos0/env).

Module: `github.com/tokenjoy/backend`

```
cmd/server/         — entrypoint
internal/
  app/              — application wiring
  config/           — env-based configuration
  http/handler/     — HTTP handlers
  http/middleware/  — auth, logging middleware
  http/response/    — response helpers
  domain/           — business logic (org, keys, models, dashboard, audit, budget, relay, session)
  permission/       — RBAC
  store/postgres/   — persistence + migrations
  integration/newapi/ — New API relay integration
  seed/             — demo seed data
  worker/           — background jobs
  pkg/              — shared utilities (roleutil, budgetutil, queryutil, etc.)
tests/              — all tests (mirrors internal/ domains)
```

The backend implements 81 REST endpoints matching `docs/Frontend-API契约.md`. First version uses in-memory stores (reset on restart); PostgreSQL is optional.

### Relay (`apps/newapi/`)

Docker-based LLM API relay service. Configured via `.env` (see `.env.example`). The backend integrates with it through `internal/integration/newapi/`.

## Key Documentation

- `docs/Frontend-API契约.md` — API contract (81 endpoints, request/response schemas)
- `docs/Frontend-开发指南.md` — Frontend development guide
- `docs/Backend-设计.md` — Backend design document
- `docs/Backend-test.md` — Backend testing guide
- `DESIGN.md` — Design system tokens and visual conventions

## Environment Variables

- `VITE_ENABLE_MOCKS=true|false` — Toggle MSW browser mocks (Vercel deploys with `true`)
- `VITE_API_PROXY_TARGET=http://localhost:8080` — Proxy `/api` to real backend in dev
