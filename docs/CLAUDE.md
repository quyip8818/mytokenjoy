# CLAUDE.md

本文件供 Claude Code 等 AI 助手快速了解仓库。详细说明见 [docs/README.md](./README.md)。

## Commands

在仓库根目录（pnpm workspace）：

- `pnpm install` — 安装依赖
- `pnpm start` / `pnpm dev` — 并发 backend + frontend（Vite 代理 `/api` → `:8080`）
- `pnpm build` — 前端 `tsc -b` + Vite 生产构建
- `pnpm lint` — ESLint（前端）+ golangci-lint（后端）
- `pnpm test` — Vitest + `go test ./tests/...`
- `pnpm verify` — lint + test + build（PR 前）
- `pnpm preview` — 本地预览前端生产构建

单测示例：`pnpm --filter @tokenjoy/frontend exec vitest run tests/lib/org.test.ts`

## Architecture

pnpm monorepo：`apps/frontend/`（前端）、`apps/backend/`（Go API）。

**前端：** Vite 8、React 19、TypeScript 6、React Router 7（从 `'react-router'` 导入）、Tailwind CSS 4、shadcn/ui、Zustand、TanStack Query。

**入口：** `apps/frontend/src/main.tsx` → `<App />`。

**路由：** 业务页在 `apps/frontend/src/routes/{domain}/`；路由定义单源 `config/routes.ts` 的 `ROUTE_DEFINITIONS`。

**数据：** `api/client.ts` 的 `request<T>()`；域 API 在 `api/{domain}.ts`；`AppApis` DI + `features/query`；契约见 `docs/Frontend-API契约.md`。

**鉴权：** Dev 在 `/login` 选成员写入 `tokenjoy_session_member` cookie；`credentials: 'include'`；401 跳转登录。

**测试：** Vitest + Testing Library；`createMockApis()` 注入，不依赖 backend 进程。用例在 `apps/frontend/tests/`。

**路径别名：** `@/*` → `src/*`；`@tests/*` → `tests/*`。

**后端：** Go + chi；`internal/app` 组合根；`internal/domain/*` 业务；`internal/store` Memory/Postgres；`tests/` 镜像 internal。详见 `docs/Backend-设计.md`。
