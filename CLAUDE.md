# CLAUDE.md

## 语言
- 所有回复使用简体中文，设计文档使用中文
- 项目没有上线，不需要 migration，不需要向后兼容

## 架构

默认在 apps/。路径无前缀 = apps/。

- `apps/` — TokenJoy，客户侧 LLM API 管控平台。前端 React19+TailwindV4+TanStackQuery，后端 Go+PG+Redis。`SUPPORT_SAAS` flag 区分私有化/SaaS。
- `apps/newapi/` — NewAPI 网关服务（Docker 部署），后端通过 HTTP 集成。
- `apps/dev-mock-llm/` — 本地开发用模拟 LLM 服务。
- `apps/docs/` — 项目文档（架构、PRD、ADR、计划、评审）。
- `web/` — 官网，纯展示 React+TailwindV3，无路由，iframe 嵌入 App 认证。
- `sms/` — 内部供应商管理系统，独立 Go module，制品不出门。
- `packages/contracts/` — 跨产品共享契约（permission manifest、notification types）。
- apps/sms 互不 import，跨产品仅 HTTP API。

## 命令

```bash
# 开发环境
pnpm start              # 启动后端+前端（需先 pnpm reset）
pnpm reset              # 重置 Docker 容器（postgres+redis+newapi）
pnpm infra              # 仅启动基础设施容器
pnpm infra:down         # 停止基础设施容器

# 前端 (apps/frontend)
pnpm -F @tokenjoy/frontend dev       # Vite dev server
pnpm -F @tokenjoy/frontend build     # tsc + vite build
pnpm -F @tokenjoy/frontend test      # vitest run
pnpm -F @tokenjoy/frontend test:e2e  # Playwright

# 后端 (apps/backend)
make start              # air 热重载启动（读 .env）
make dev-bootstrap      # 初始化本地开发数据（需 NEW_API_ENABLED=true）
make test-unit          # go test -tags=testhook ./tests/...
make test-integration   # 集成测试
make lint               # lint-clock + lint-layer
make format             # gofmt -w
make scaffold-domain    # 脚手架新域

# 权限生成
pnpm generate:permissions  # 从 packages/contracts/permission/manifest.json 生成后端+前端常量
```

## 文件放置

### apps 前端 (apps/frontend/)
- 页面：`routes/{domain}/{page}.tsx`（仅组合）
- 领域：`features/{domain}/`（含 hooks/components/lib/index.ts，外部禁止 deep import）
- 原子 UI：`components/ui/`（无业务语义）
- API：`api/{domain}.ts`，禁止直接 import——通过 useApis()
- 测试：`apps/frontend/tests/`（镜像 src/）
- E2E：`apps/frontend/e2e/`

**前端特性目录：** account, approval, audit, auth, billing, budget, dashboard, keys, models, mydashboard, notifications, org, query, session, workflow

**路由结构：**
- `/dashboard/cost|usage` — 数据看板
- `/keys/platform` — Key 管理
- `/approvals` — 审批中心
- `/keys/provider` — 供应商 Key
- `/models/list|routing` — 模型管理
- `/budget|/budget/alerts` — 预算管理
- `/billing` — 钱包管理
- `/org/data-source|structure|roles` — 组织与权限
- `/audit/operations|calls` — 审计
- `/me/keys|usage|settings` — 我的

### apps 后端 (apps/backend/)
- `cmd/server/` — 主入口
- `cmd/dev-bootstrap/` — 开发环境初始化
- `internal/app/` — 组合根（DI wiring）
- `internal/config/` — 环境配置
- `internal/domain/` — 业务逻辑（按子域隔离）
- `internal/http/handler/` — HTTP handler（每域一包）
- `internal/http/middleware/` — 中间件（auth, RBAC, CORS）
- `internal/identity/` — 认证授权（session, verifycode, secrets）
- `internal/infra/` — 基础设施（jobs, notification, permission, ratelimit, river, scheduler）
- `internal/integration/` — 外部集成（newapi, datasource/feishu）
- `internal/pkg/` — 共享工具（budget, common, org, tree, newapiunits）
- `internal/store/` — 仓储接口
- `internal/store/postgres/` — PostgreSQL 实现（raw SQL，无 ORM）
- `internal/worker/` — 后台任务 worker
- `seed/` — 数据种子（demo bootstrap）
- `tests/` — 全部单元测试（mirrors internal/）

**后端域目录 (internal/domain/):** adminport, approval, audit, billing, budget, company, dashboard, gateway, grants, keys, memberanalytics, models, notification, org, port, types, usage

### sms
结构同 apps。测试 `sms/{frontend,backend}/tests/`，文档 `sms/docs/`。

### 通用
- 禁止在 src/internal/组件旁放测试文件
- 禁止在 frontend/backend/ 下新建 .md（README.md 除外）
- 文档统一放 `apps/docs/`

## 测试

```bash
pnpm test                 # apps 全量
pnpm test:sms             # sms 全量
# 单文件
cd apps/backend && go test -tags=testhook -run "TestXxx" ./tests/domain/xxx/ -v -count=1
cd sms/backend && go test -run "TestXxx" ./tests/domain/xxx/ -v
# 前端单文件
pnpm -F @tokenjoy/frontend exec vitest run tests/features/xxx/yyy.test.ts
```

### 后端（apps）
- 改 schema.sql/seed → bump `testTemplateVersion`（`tests/testutil/pg/template.go`）
- 时钟固定 `2026-06-19`，不用 `time.Now()` 断言
- 必须 `-tags=testhook`
- DB: `postgres://tokenjoy:tokenjoy@127.0.0.1:5510/tokenjoy`

### 后端（sms）
- DB: `postgres://sms:sms@127.0.0.1:5510/sms`，无需 build tag

### 前端
- Vitest + @testing-library/react，API 层 vi.mock
- E2E: Playwright（`apps/frontend/e2e/`）

## 环境变量

关键变量（见 `apps/backend/.env.example` 完整列表）：
- `DATABASE_URL` — PostgreSQL 连接
- `LOG_DATABASE_URL` — 日志库连接（可选）
- `SESSION_SECRET` — JWT session 签名
- `DATA_SOURCE_CREDENTIAL_KEY` — 数据源凭证加密 key
- `DEPLOY_ENV` — local / staging / production
- `BOOTSTRAP_MODE` — none / minimal / demo
- `NEW_API_ENABLED` — 启用 NewAPI 集成
- `NEW_API_BASE_URL` — NewAPI 服务地址
- `NEW_API_DATABASE_URL` — 读取 NewAPI 管理员 token（替代旧的 NEW_API_ADMIN_TOKEN）
- `SUPPORT_SAAS` — 多租户 SaaS 模式
- `REDIS_URL` — Redis（限流、网关预算缓存）
- `VITE_API_PROXY_TARGET` — 前端代理目标（默认 http://localhost:8010）
- `TOKENJOY_COMPANY_ID` / `LOCAL_COMPANY_ID` — SaaS 模式下的公司 ID
- `PLATFORM_BOOTSTRAP_EMAIL` / `PLATFORM_BOOTSTRAP_PASSWORD` — 平台初始化账号

## 脚本

- `scripts/dev.sh` — 开发环境编排主入口
- `scripts/dev-sms.sh` — SMS 产品开发脚本
- `scripts/reset-budget-data.sh` — 清空预算数据（保留总公司+admin）
- `scripts/dev/` — 子脚本（infra, reset, start, test, frontend-wait）
- `scripts/lib/` — 共享函数（common.sh, db-reset.sh）
- `scripts/postgres-init/` — Docker postgres 初始化（创建多数据库）

## Ponytail — lazy senior dev

写代码前按顺序停在第一个成立的：YAGNI → 复用已有 → 标准库 → 平台特性 → 已装依赖 → 一行搞定 → 最小实现。

- 先理解问题再动手，trace 完整流程
- Bug fix = 修根因，grep 所有 caller，修共享函数一次
- 不加未请求的抽象/依赖/样板代码
- 删除优于新增，无聊优于聪明，最少文件
- 最短 diff 胜出（前提：改对地方）
- 刻意简化标 `ponytail:` 注释说明天花板和升级路径
- 非 trivial 逻辑留一个最小可运行验证（assert/小测试）

不偷懒：理解问题、信任边界校验、防数据丢失的错误处理、安全、无障碍。
