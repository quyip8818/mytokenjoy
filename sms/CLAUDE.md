# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

All commands run from the monorepo root (`/Users/duian/side-project/mytokenjoy`):

```bash
# Full stack (infra + backend + frontend)
pnpm start sms          # Postgres + Redis + NewAPI-SMS + backend + frontend
pnpm reset sms          # Wipe sms databases, re-seed, restart

# Backend (sms/backend/, Go 1.23)
cd sms/backend
make dev                # air hot-reload (or fallback go run)
make run                # go run ./cmd/server
make seed               # seed demo data
make lint               # go vet + gofmt check
make format             # gofmt -w .
make test               # go test ./tests/...
make build              # compile to bin/server + bin/seed

# Frontend (sms/frontend/)
pnpm -F @sms/frontend dev       # Vite dev server (port 9100)
pnpm -F @sms/frontend build     # tsc + vite build
pnpm -F @sms/frontend test      # vitest
pnpm -F @sms/frontend lint      # eslint
```

## Architecture

SMS (Supplier Management System) 是一个 AI 模型供应商管理系统，管理供应商、合同、采购、模型目录和供应商评估。

### 部署拓扑

```
Frontend (React, :9100) → Backend (Go, :8020) → PostgreSQL (:5510, db=sms)
                                               → NewAPI-SMS (:3020, db=sms_newapi)
```

### Backend (`sms/backend/`)

Go 模块 `sms/backend`，chi/v5 路由，pgx/v5 数据库。

```
cmd/server/          — HTTP 入口
cmd/seed/            — 数据库 seed
schema.sql           — DDL（根目录，非 internal/）
internal/
  app/               — DI 组装（pool → stores → services → router）
  config/            — 环境变量配置
  domain/            — 业务逻辑（10 个子域）
    auth/            — 登录、JWT、刷新、角色
    supplier/        — 供应商 CRUD + 联系人
    model/           — AI 模型目录（创建/更新时触发 NewAPI 同步）
    contract/        — 合同管理 + 附件上传
    order/           — 采购订单（状态机：pending→confirmed→completed/cancelled）
    evaluation/      — 供应商评估（加权评分 → 等级）
    dashboard/       — 统计卡片 + 图表
    user/            — 用户管理（admin/buyer/viewer）
    newapisync/      — NewAPI 价格同步（本地 → NewAPI ModelRatio）
    oauth/           — OAuth2 client_credentials（供外部系统调用）
    sync/            — 模型目录导出 API（供 tokenjoy pull）
  http/
    handler/         — 每个 domain 一个 handler.go
    middleware/      — auth, cors, logger, oauth_guard, role
    helpers/         — 参数解析、错误映射
    response/        — 统一 JSON 响应
  integration/newapi/ — NewAPI HTTP 客户端 + TokenStore
  store/postgres/    — SQL 实现
```

**路由结构：**
- `/api/auth/*` — 公开（登录、刷新、登出）
- `/api/oauth/*` — OAuth2 token endpoint
- `/api/sync/*` — OAuth2 守卫（scope: sync:read）
- `/api/suppliers|models|contracts|purchase-orders|evaluations|dashboard/*` — JWT 认证
- `/api/users/*` — admin 角色
- `/api/newapi/*` — admin 角色（手动触发同步）

**NewAPI 同步机制：**
- SMS 是模型价格的 SOT（单一数据源）
- 模型创建/更新时自动异步同步到 NewAPI
- 手动触发：`POST /api/newapi/sync`
- 价格公式：`modelRatio = inputPrice / 2`，`completionRatio = outputPrice / inputPrice`
- Token 从 `sms_newapi.users` 表读取，401 时自动刷新

### Frontend (`sms/frontend/`)

React 19 SPA — Vite 8, TanStack Query 5, Zustand 5, react-router v7, Radix UI, Tailwind CSS 4。

```
src/
  config/routes.ts     — 路由定义（单一数据源）
  api/                 — fetch 客户端 + domain API 模块
  features/            — 按业务域组织（hooks/components）
  routes/              — 页面入口
  components/ui/       — 基础 UI 组件
  components/layout/   — 布局组件
```

**认证：** JWT access token 在 localStorage，refresh token 为 httpOnly cookie。401 时自动刷新。

**路径别名：** `@/*` → `./src/*`

## Database

- PostgreSQL，端口 5510，数据库 `sms`
- Schema 在 `sms/backend/schema.sql`（根目录不是 internal/）
- 不需要 migration，直接修改 schema.sql 然后 `pnpm reset sms`
- 11 张表：users, sessions, suppliers, supplier_contacts, models, contracts, contract_attachments, purchase_orders, evaluations, evaluation_weights, oauth_clients

## Environment

后端必需的环境变量（`.env` 文件在 `sms/backend/.env`，已 gitignore）：

- `DATABASE_URL` — PostgreSQL 连接（必填）
- `PORT` — 后端端口（默认 8080，实际用 8020）
- `CORS_ORIGINS` — 前端地址（http://localhost:9100）
- `JWT_SECRET` — JWT 签名密钥
- `NEWAPI_BASE_URL` — NewAPI 地址（http://localhost:3020，空则禁用同步）

## Key Documentation

- `sms/docs/project-overview.md` — 项目总览
- `sms/docs/ai-model-supplier-management-design.md` — 详细设计文档
- `sms/docs/plan/newapi-integration.md` — NewAPI 集成设计
