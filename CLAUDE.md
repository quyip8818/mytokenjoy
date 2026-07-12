# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

All commands run from the repo root (pnpm workspace, `pnpm@11.9.0`):

```bash
# Full-stack
pnpm start              # Backend + frontend together (waits for backend healthz)
pnpm verify             # Full CI check: lint + test + build
pnpm generate:permissions  # Regenerate permission keys from packages/contracts manifest

# Frontend (apps/frontend)
pnpm -F @tokenjoy/frontend start     # Vite dev server
pnpm -F @tokenjoy/frontend build     # tsc + vite build
pnpm -F @tokenjoy/frontend test      # vitest run
pnpm -F @tokenjoy/frontend test:e2e  # Playwright

# Single frontend test:
pnpm -F @tokenjoy/frontend exec vitest run tests/features/auth/use-login-page.test.ts

# Backend (apps/backend, from apps/backend/)
make start              # go run ./cmd/server (reads .env)
make test-unit          # go test -tags=testhook ./tests/... (requires PostgreSQL)
make lint               # go vet + gofmt check
make format             # gofmt -w .

# Prerequisites: pnpm start:postgres (or DATABASE_URL)

# Single backend test:
cd apps/backend && go test ./tests/domain/gateway/... -run TestPrecheckRejectsZeroBudget -v

# NewAPI (apps/newapi)
pnpm start:newapi       # docker compose up
pnpm verify:gate        # 通路冒烟（自建 Key + Gateway + webhook）
pnpm verify:integration # 入账 + Toggle/Rotate/Revoke + metrics（需 NEW_API_ADMIN_TOKEN）
```

## Architecture

pnpm monorepo with apps under `apps/` and shared contracts under `packages/`:

### Contracts (`packages/contracts/`)

Cross-app JSON contracts and codegen. Permission manifest: `permission/manifest.json` → `pnpm generate:permissions` → backend `keys.go` + frontend `permission-keys.ts`.

### Frontend (`apps/frontend/`)

React 19 SPA — Vite, TypeScript, TailwindCSS v4 (CSS-first, no tailwind.config).

- **Routing:** react-router v7 (`import from 'react-router'`, NOT `'react-router-dom'`). Routes defined in `config/routes.ts` via `ROUTE_DEFINITIONS` (single source of truth).
- **State:** Zustand v5 stores co-located with features.
- **UI:** shadcn/ui in `components/ui/`, Radix primitives, lucide-react icons. `cn()` from `lib/utils.ts`.
- **API layer:** Custom fetch in `api/client.ts` (`/api` base). Domain namespaces in `api/*.ts`. Vite proxies `/api` to backend.
- **Testing:** Vitest + @testing-library/react. Tests in `tests/`. Use `createMockApis()` + `renderHookWithProviders` from `@tests/utils`.
- **Path alias:** `@/*` → `./src/*`, `@tests/*` → `./tests/*`

Key conventions:

- Route pages: `routes/{domain}/{page}.tsx` — compose only, delegate to `features/{domain}/hooks/use-{page}-page.ts`
- Page hooks use `useInjectedApis(injectedApis?)` for testability; other code uses `useApis()`
- Shared domain UI: `components/{domain}/` (2+ consumers); page-only: `routes/{domain}/components/`
- Never import API functions directly in business code — go through the DI layer
- Workflows (dialogs/forms): `features/workflow/workflows/`, opened via `useWorkflow().open()`

### Backend (`apps/backend/`)

Go 1.24 — chi router, PostgreSQL (pgx v5), env config (caarlos0/env).

Module: `github.com/tokenjoy/backend`

```
cmd/server/              — entrypoint
internal/
  app/                   — application wiring (DI)
  config/                — env-based configuration
  domain/                — business logic by subdomain:
    adminport/, audit/, billing/, budget/, company/, dashboard/,
    grants/, keys/, memberanalytics/, models/, org/, gateway/,
    newapisync/, usage/
  http/handler/          — HTTP handlers (one package per domain)
  http/middleware/       — auth, RBAC, company resolve, CORS
  http/httputil/         — response/decode helpers
  identity/              — authz, credentials, session tokens
  infra/                 — worker, notification, permission manifest
  integration/           — external: newapi (admin_port_adapter), datasource (feishu)
  pkg/                   — shared utilities (budget calc, org helpers, newapiunits, tree)
  store/                 — repository interfaces + implementations:
    postgres/            — PostgreSQL (production + tests)
seed/                    — demo bootstrap + contract IDs (see docs/Backend.md §5.3)
tests/                   — ALL unit tests (mirrors internal/ structure)
  testutil/              — test helpers, fixtures, stubs
```

**Store pattern:** Production and tests both use `postgres.New`. Tests use per-schema isolation via `testutil.NewTestStore` / `NewTestApp` (see `docs/Backend.md` §5).

**Multi-tenant:** `company_id` is the tenant boundary, carried via `domain/company.Context` in request context. Platform (SaaS admin) is a separate auth layer.

**NewAPI integration:** Domain talks to NewAPI Admin via `domain/adminport.Port` (adapter in `integration/newapi/admin_port_adapter.go`); quota conversion in `pkg/newapiunits/`. `domain/newapisync/` syncs PlatformKey/ProviderKey; `domain/gateway/` runs `/v1` precheck then reverse-proxies. Precheck validates: key validity → key status → model whitelist → budget → forward.

### NewAPI (`apps/newapi/`)

Docker-based LLM API gateway upstream (NewAPI). Configured via `.env`. Backend HTTP client and `admin_port_adapter` live in `internal/integration/newapi/`.

## Testing Patterns (Backend)

- Tests live in `tests/` (external test packages, e.g., `package gateway_test`)
- Use `testutil.NewTestStore(t, opts...)` or `testutil.NewTestApp(t, mutate)` for store/app
- Requires PostgreSQL: `pnpm start:postgres` before `make test-unit`
- **Dev loop:** `make test-fast` (from `apps/backend/`, pure `tests/pkg/...`, no Postgres) for pkg changes; `go test -tags=testhook ./tests/domain/<域>/...` or `./tests/http/middleware/...` for a single domain; **`make test-unit`** before commit/PR
- **SSOT patterns:** GET contracts → `tests/handler/core/contract_test.go`; write smoke → `mutating_contract_test.go`; middleware unit → `tests/http/middleware/` (`stubs_test.go` + `middleware_test.go`, chi + stub, not full `NewApp`); newapisync outbox → `tests/domain/newapisync/outbox_*.go`
- Use `testutil.Ctx()` for a default company context
- Use `testutil.CtxForCompany(id)` for specific company
- Config options: `testutil.WithNewAPIEnabled(true)`, `testutil.WithSupportSaas(true)`, etc.
- Org service: `orgfix.NewService(t, cfg, st)` from `tests/testutil/org`
- Gateway scenarios: `gatewaytf.BuildGatewayScenario(t, opts)` from `tests/testutil/gateway`
- HTTP handler tests use `testutil/http` with real chi router + seeded store
- Float pointer helper: `budgetfix.FloatPtr` from `tests/testutil/budget/ptr.go`
- The `-tags=testhook` build tag activates test hooks in `internal/app/testhook.go` and `testhook_registry.go` (`BuildRegistry`, `MustNewAPISync`)

## Key Documentation

- `docs/plan.md` — Engineering backlog (single source for pending work)
- `docs/PRD.md` — Product requirements (authoritative PRD)
- `docs/Frontend.md` — Frontend development guide and API contract
- `docs/Backend.md` — Backend design document (index)
- `docs/Backend-测试优化.md` — Test coverage + speed optimization (PR1/PR2 done, PR3 backlog)
- `docs/Backend-架构.md` — Layering, naming (Gateway / NewAPISync / PlatformKey), Store, Worker
- `docs/Backend-配置架构.md` — Config load, production contract, bootstrap, Clock
- `docs/Backend-业务时钟与账期.md` — Business clock, dual period keys, guards
- `docs/Backend-预算.md` — Budget subsystem design
- `docs/Backend-存储架构.md` — Storage layer design
- `DESIGN.md` — Design system tokens and visual conventions

## Environment Variables

- `VITE_API_PROXY_TARGET=http://localhost:8080` — Frontend proxy target
- `DATABASE_URL` — PostgreSQL connection (required for tests and production)
- `DATA_SOURCE_CREDENTIAL_KEY` — Required credential encryption key (32-byte hex or base64)
- `DEPLOY_ENV` — `local` / `staging` / `production` (`production` triggers fail-fast production contract)
- `BOOTSTRAP_MODE` — `none` / `minimal` / `demo` (empty DB bootstrap policy)
- `SECURE_COOKIE` — Set-Cookie Secure flag (required `true` when `DEPLOY_ENV=production`)
- `CLOCK_ANCHOR` — Optional `YYYY-MM-DD` for fixed dashboard clock and seed reference date
- `NEW_API_ENABLED=true` — Enable NewAPI integration
- `NEW_API_GATEWAY_ENABLED=true` — Enable `/v1` Gateway
- `NEW_API_BASE_URL` / `NEW_API_ADMIN_TOKEN` — NewAPI service credentials
- `PLATFORM_SHARED_NEW_API_GROUP` — SaaS shared NewAPI group (default `platform_shared`)
- `SESSION_SECRET` — JWT session signing key
- `SUPPORT_SAAS=true` — Multi-tenant SaaS mode
