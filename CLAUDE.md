# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

All commands run from the repo root (pnpm workspace, `pnpm@11.9.0`):

```bash
# ─── Apps (TokenJoy 客户侧) ───
pnpm start               # ensure-infra + backend(:8010) + frontend(:5173) + dev-mock
pnpm reset               # Reset apps databases (local mode, full seed)
pnpm reset saas          # Reset apps databases (SaaS multi-tenant mode)
pnpm reset [local|saas] [--empty|--minimal|--full]

# ─── SMS (内部运营) ───
pnpm start sms           # sms backend(:8020) + frontend(:5174)
pnpm reset sms           # Reset sms databases + seed

# ─── All ───
pnpm start all           # 并行启动 apps + sms
pnpm reset all           # 重置全部

# ─── 基础设施 ───
pnpm infra               # docker compose up (postgres:5510 + redis:6310 + newapi-apps:3010 + newapi-sms:3020)
pnpm infra:down          # docker compose down

# ─── 测试 ───
pnpm test                # apps 全量测试 (frontend vitest + backend go test)
pnpm test:sms            # sms 全量测试 (frontend vitest + backend go test)
pnpm test:integration    # apps backend 集成测试
pnpm test:sms:integration # sms backend 集成测试
pnpm test:e2e            # apps frontend Playwright E2E
pnpm test:sms:e2e        # sms frontend Playwright E2E

# ─── 质量 ───
pnpm lint                # apps lint (frontend eslint + backend go vet)
pnpm lint:sms            # sms lint (frontend eslint + backend go vet)
pnpm verify              # CI: lint + test + build
pnpm verify gate         # Gateway + webhook smoke
pnpm generate:permissions

# ─── 构建 ───
pnpm build               # apps frontend build
pnpm build:sms           # sms frontend build

# ─── 单文件测试 ───
# Frontend:
pnpm -F @tokenjoy/frontend exec vitest run tests/features/auth/use-login-page.test.ts
# Backend:
cd apps/backend && go test -tags=testhook ./tests/domain/gateway/... -run TestPrecheckRejectsZeroBudget -v
# SMS backend:
cd sms/backend && go test ./tests/domain/auth/... -v
```

## Architecture

pnpm monorepo with two products + shared contracts:

```
mytokenjoy/
├── apps/                    ← 客户侧产品 (TokenJoy Local + SaaS)
│   ├── frontend/            ← React SPA (@tokenjoy/frontend)
│   ├── backend/             ← Go 后端 (@tokenjoy/backend, module: github.com/tokenjoy/backend)
│   ├── newapi/              ← NewAPI Docker 构建
│   └── dev-mock-llm/       ← 本地模拟 LLM
├── sms/                     ← 内部运营产品
│   ├── frontend/            ← React SPA (@sms/frontend)
│   ├── backend/             ← Go 后端 (@sms/backend, module: sms/backend)
│   ├── newapi/              ← NewAPI 配置
│   └── docs/               ← SMS 产品文档
├── packages/contracts/      ← 跨产品共享契约
├── docker-compose.yml       ← 统一基础设施
├── scripts/                 ← 开发脚本
│   ├── dev.sh              ← 主调度器 (start/reset/test 路由)
│   ├── dev-sms.sh          ← SMS start/reset
│   ├── lib/common.sh       ← 共享路径
│   ├── lib/db-reset.sh     ← DB reset 共享函数
│   └── postgres-init/      ← 容器首次初始化脚本
├── apps/docs/              ← TokenJoy 产品文档
└── docs/                   ← 仓库级/跨产品文档
```

### 端口总表

| 服务 | apps (TokenJoy) | sms |
|------|----------------|-----|
| Postgres | 5510 (共用) | 5510 (共用) |
| Redis | 6310 (共用) | 6310 (共用) |
| Backend | 8010 | 8020 |
| Frontend | 5173 | 5174 |
| NewAPI | 3010 | 3020 |

### 数据库隔离

一个 Postgres 容器，6 个 database：
- apps: `tokenjoy` (owner: tokenjoy), `newapi` (owner: tokenjoy), `logs` (owner: tokenjoy)
- sms: `sms` (owner: sms), `sms_newapi` (owner: sms), `sms_logs` (owner: sms)

Redis db number: apps NewAPI=0, sms NewAPI=1, apps backend=2, sms backend=3

### Frontend (`apps/frontend/`)

React 19 SPA — Vite 8, TypeScript 6, TailwindCSS v4 (CSS-first, `@tailwindcss/vite` plugin).

- **Routing:** react-router v7 (`import from 'react-router'`, NOT `'react-router-dom'`). Routes in `config/routes.ts`.
- **State:** Zustand v5 stores co-located with features.
- **UI:** shadcn/ui in `components/ui/`, Radix primitives, lucide-react icons.
- **API layer:** Custom fetch in `api/client.ts` (`/api` base). Vite proxies `/api` to backend:8010.
- **Testing:** Vitest 4 + @testing-library/react. Tests in `tests/`.
- **Path alias:** `@/*` → `./src/*`, `@tests/*` → `./tests/*`

### Frontend (`sms/frontend/`)

React 19 SPA — Vite 8, TypeScript 6, TailwindCSS v4. 结构同 apps/frontend。
- Vite proxies `/api` to backend:8020, port 5174.
- 包名 @sms/frontend，禁止与 apps/frontend 互相 import。

### Backend (`apps/backend/`)

Go 1.24 — chi router, PostgreSQL (pgx v5), env config.
Module: `github.com/tokenjoy/backend`

```
cmd/server/              — entrypoint
internal/
  app/                   — DI wiring
  config/                — env config (DefaultDatabaseURL: 127.0.0.1:5510)
  domain/                — business logic by subdomain
  http/handler/          — HTTP handlers
  http/middleware/       — auth, RBAC, company resolve, CORS
  identity/              — authz, credentials, session tokens
  integration/           — external: newapi, datasource
  pkg/                   — shared utilities
  store/postgres/        — PostgreSQL implementations
seed/                    — demo bootstrap + contract IDs
tests/                   — ALL unit tests (mirrors internal/)
  testutil/              — test helpers, fixtures
```

### Backend (`sms/backend/`)

Go 1.23 — chi router, pgx v5. Module: `sms/backend`. 独立于 apps/backend。

```
cmd/server/              — entrypoint
cmd/seed/                — seed script
internal/
  config/                — env config
  domain/                — business logic (auth, supplier, model, contract, order, evaluation, dashboard, user, newapisync)
  http/handler/          — HTTP handlers
  integration/newapi/    — NewAPI HTTP client
  store/postgres/        — SQL implementations
tests/                   — ALL tests (mirrors internal/)
```

## Testing Patterns

### Backend (apps/backend)
- Tests in `tests/` (external test packages, e.g., `package gateway_test`)
- `testutil.NewTestStore(t, opts...)` / `testutil.NewTestApp(t, mutate)` for store/app
- Requires PostgreSQL on port 5510: `pnpm infra` before `make test-unit`
- Build tag: `-tags=testhook`
- Clock fixed: `ClockAnchor = "2026-06-19"`, period = `"2026-06"`
- **改了 schema.sql 或 seed/ → bump `testTemplateVersion`** in `tests/testutil/pg/template.go`

### Backend (sms/backend)
- Tests in `sms/backend/tests/` (external test packages)
- DATABASE_URL: `postgres://sms:sms@127.0.0.1:5510/sms`
- 运行: `cd sms/backend && go test ./tests/...`

### Frontend (apps/frontend + sms/frontend)
- Vitest 4 + @testing-library/react
- Tests in `{app}/tests/`, mirrors src/ path
- API mocked via `vi.mock`
- E2E: Playwright, config in `playwright.config.ts`

## File Placement Rules

### 测试
- Frontend：`apps/frontend/tests/`（镜像 src/ 路径）
- Backend：`apps/backend/tests/`（镜像 internal/ 路径，外部测试包）
- SMS Frontend：`sms/frontend/tests/`（镜像 src/ 路径）
- SMS Backend：`sms/backend/tests/`（镜像 internal/ 路径）
- 禁止在 src/、internal/、组件旁边放测试文件

### 文档
- 根 `docs/`：仓库级/跨产品文档（策略 ADR、整体 Roadmap）
- `apps/docs/`：TokenJoy（apps）产品专属文档（plan/、adr/、todos/ 等）
- `sms/docs/`：SMS 产品专属文档
- 禁止在 apps/frontend/、apps/backend/、sms/frontend/、sms/backend/ 下新建 .md（README.md、DESIGN.md 除外）
- 禁止在项目根新建 .md（CLAUDE.md、DESIGN.md 除外）

### 后端
- 禁止在 cmd/ 放业务逻辑（仅 main 入口 + 启动编排）
- 禁止跨 domain 直接引用另一个 domain 的内部实现
- 共享内核例外：`domain/types`、`domain/grants`、`domain/company`、`domain/newapisync`
- 跨域协作通过 ports/interfaces 解耦

### 前端
- 页面入口：`routes/{domain}/{page}.tsx`（仅组合，从 features/ 导入）
- 领域特性包：`features/{domain}/`（含 hooks/、components/、lib/、index.ts）
- features/ 必须有 index.ts barrel export；外部禁止 deep import
- 禁止直接 import API 函数——通过 useApis()/useInjectedApis()
- 共享合约/类型放 packages/contracts/

### 跨产品约束
- apps/ 和 sms/ 之间禁止 Go import
- apps/ 和 sms/ 之间禁止 TypeScript import
- 共享类型/契约只能放 packages/contracts/
- 跨产品通信只能通过 HTTP API

### 语言
- 所有回复使用简体中文，所有设计文档使用中文。

### 其他
- 项目没有上线，不需要 migration，不需要向后兼容
