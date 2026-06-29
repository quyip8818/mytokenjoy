# TokenJoy Backend 设计

本文档描述 `apps/backend/` Go 后端服务的设计方案，用于对接前端 REST API。MSW Mock 已迁移完成，种子数据与业务逻辑位于 `internal/seed/` 与各 domain Service。

**相关文档：**

| 文档     | 路径                                                   | 职责                                                     |
| -------- | ------------------------------------------------------ | -------------------------------------------------------- |
| 系统架构 | [tokenjoy-architecture.md](./tokenjoy-architecture.md) | 双平面、预算消耗闭环（Relay 刷卡机）、目录、New API 集成 |
| API 契约 | [Frontend-API契约.md](./Frontend-API契约.md)           | 81 个端点、请求/响应体、错误码                           |
| 前端开发 | [Frontend-开发指南.md](./Frontend-开发指南.md)         | 前端架构、调用链、联调方式                               |
| 测试指南 | [Backend-测试指南.md](./Backend-测试指南.md)           | 测试目录、运行方式、编写规范                             |

**权威来源（按优先级）：**

1. 端点与类型 — [Frontend-API契约.md](./Frontend-API契约.md) §5–§6
2. 前端类型定义 — `apps/frontend/src/api/types/`
3. 后端类型与实现 — `apps/backend/internal/domain/types/`、`internal/seed/`

---

## 1. 目标与范围

### 1.1 目标

| 目标          | 说明                                                                          |
| ------------- | ----------------------------------------------------------------------------- |
| 真实 API 联调 | 开发环境通过 `VITE_API_PROXY_TARGET` 指向 Go 服务                             |
| 契约对齐      | 实现契约 §5 全部 **81** 个端点，JSON 结构与 `api/types/` 一致                 |
| 种子数据      | `internal/seed/` 加载演示组织/预算/Key；`usage_buckets` seed + webhook ingest |
| Monorepo 集成 | 置于 `apps/backend/`，根目录 `pnpm start` 一键启动                            |

### 1.2 首版不做

- 真实飞书 / 钉钉 / 企微对接（数据源 test/import 保持模拟延迟与固定结果）
- Redis 等缓存层（PostgreSQL 已可选启用，见 §5.3）
- 生产级 JWT / OAuth（Session 先复刻 Mock 双轨鉴权）
- OpenAPI 代码生成（类型以契约文档 + 前端 TS 为准，后端手写 struct）

### 1.3 成功标准

1. 根目录 `pnpm start` 下，前端 16 个业务页可正常浏览与操作
2. 前端 Vitest 使用 `createMockApis` 注入，不依赖 backend 进程
3. CI 可编译 backend、运行 `go test ./tests/...`

---

## 2. 迁移基线（已完成）

原 MSW fixtures/handlers/lib 已迁入：

- **种子数据** — [`internal/seed/`](../apps/backend/internal/seed/)（组织、预算、Key、审计日志等）
- **用量事实** — `usage_buckets` 表 + webhook ingest；启动时 `seed.ApplyUsageBuckets`
- **业务规则** — 各 domain Service（budget、keys、models、dashboard 等）

原 `internal/pkg/dashboardcalc` 假数据生成器已删除；看板读真实桶聚合。

---

## 3. 技术选型

| 类别      | 选型                                                                              | 理由                                  |
| --------- | --------------------------------------------------------------------------------- | ------------------------------------- |
| 语言      | Go 1.22+                                                                          | 单二进制部署、与前端 monorepo 解耦    |
| HTTP 路由 | [chi](https://github.com/go-chi/chi) v5                                           | 轻量、标准 `net/http`、中间件生态成熟 |
| 配置      | [caarlos0/env](https://github.com/caarlos0/env) + 环境变量                        | 与 Vite proxy 配置风格一致            |
| 日志      | `log/slog`                                                                        | 标准库结构化日志                      |
| JSON      | `encoding/json` + struct tag                                                      | 与前端字段名对齐（camelCase）         |
| 校验      | 手写 + 可选 [go-playground/validator](https://github.com/go-playground/validator) | 422 业务规则以 Mock 行为为准          |
| 测试      | `testing` + `httptest`                                                            | Handler 级契约测试                    |
| 依赖注入  | 构造函数注入（无全局单例）                                                        | 与前端 `AppApis` DI 理念一致          |

不引入 ORM、不引入重型框架（如 Gin 全栈方案），保持首版可维护性。

---

## 4. 项目结构

```
apps/backend/
├── cmd/
│   └── server/
│       └── main.go              # 组装 Config、Store、Services、Router、启动 HTTP
├── internal/
│   ├── config/
│   │   └── config.go            # PORT、CORS、Demo 模式开关
│   ├── domain/                  # 按契约域划分；每域含 service（大域按职责拆文件）
│   │   ├── session/
│   │   ├── org/                 # datasource / department / member / role
│   │   ├── budget/              # service + ingest + rebalance
│   │   ├── keys/                # service + provider_key + platform_key + approval
│   │   ├── relay/               # TokenLifecycle + outbox payload（NewAPI 同步）
│   │   ├── models/
│   │   ├── dashboard/
│   │   └── audit/
│   ├── app/                     # 组合根（DI wiring）
│   ├── integration/newapi/      # NewAPI AdminClient + webhook DTO
│   ├── worker/                  # outbox / rebalance / log 补偿轮询
│   ├── http/
│   │   ├── router.go            # chi 路由注册，挂载 /api 前缀
│   │   ├── middleware/          # CORS、RequestID、Recover、Auth
│   │   ├── response/            # JSON 成功/错误响应、Paginated 辅助
│   │   └── handler/             # 薄 Handler：解析请求 → 调 Service → 写响应
│   │       ├── session.go
│   │       ├── org.go
│   │       └── ...
│   ├── permission/              # 权限 key 常量、resolveMemberPermissions
│   ├── seed/                    # 启动时加载种子数据
│   │   ├── data/                # JSON 种子文件（由 TS fixtures 导出）
│   │   └── loader.go
│   └── store/                   # Store 接口 + Memory 实现
│       ├── store.go             # Store 聚合接口（含 WithTx）
│       ├── memory.go            # 读写锁保护的 in-memory 实现
│       └── postgres/            # 可选 PG：JSON 快照 + relay 关系表
│           └── migrations/      # embed 迁移（唯一来源）
├── tests/                       # 外部测试树，镜像 internal 结构
├── go.mod
├── go.sum
└── Makefile                     # dev、test、lint（可选 golangci-lint）
```

根目录 `package.json` 后续可增加：

```json
"dev:backend": "cd apps/backend && go run ./cmd/server",
"dev:all": "concurrently \"pnpm dev:backend\" \"pnpm start\""
```

（`concurrently` 为可选 devDependency，首版文档不强制。）

---

## 5. 分层架构

```
HTTP Request
  └─ middleware (CORS, Auth, Recover)
       └─ handler.{Domain}Handler   // 解析 path/query/body，HTTP 状态码
            └─ domain.{Domain}Service  // 业务规则、校验、编排
                 └─ store.Store         // 持久化抽象（Memory 默认；可选 Postgres）
                      └─ seed 初始状态
```

**NewAPI 集成（可选）：** 当 `NEW_API_ENABLED=true` 时，`domain/relay.TokenLifecycle` 负责 Token/Channel 同步与 outbox 入队；`worker.Runner` 后台消费 outbox、webhook 重试与 rebalance；`budget.IngestService` 处理 webhook 日志入账。

### 5.1 Handler 层

- 职责：HTTP 适配，不含业务规则
- 统一错误：`response.Error(w, status, message)` → `{ "message": "..." }`
- `void` 端点返回 `{}` 或 `null`（避免前端 `res.json()` 解析失败，见契约 §2.3）
- 分页响应使用统一 `Paginated[T]` 结构

### 5.2 Service 层

- 每个域一个 `Service` 接口 + `service` 实现，构造函数注入 `store.Store` 及跨域依赖（如 Keys 依赖 Budget、Org）
- 业务校验失败返回 typed error，Handler 映射为 400 / 404 / 422
- 模拟延迟（如 data-source test 1s、import 2s）通过 `config` 可关闭，默认与 Mock 一致

### 5.3 Store 层

```go
type Store interface {
    Org() OrgRepository
    Budget() BudgetRepository
    Keys() KeysRepository
    Models() ModelsRepository
    Dashboard() DashboardRepository
    Audit() AuditRepository
    Relay() RelayRepository          // mapping / outbox / ingest cursor / rebalance queue
    WithTx(ctx context.Context, fn func(Store) error) error
}
```

**Memory（默认）：** `DATABASE_URL` 为空时使用。启动时从 `seed/` 深拷贝可变状态；`Set*` 写操作直接修改内存并返回 `error`（Memory 恒为 nil）。

**Postgres（可选）：** 设置 `DATABASE_URL` 后启用。域数据以单 JSONB 行（`domain_snapshot`）持久化，Relay 基础设施走独立 SQL 表；`Set*` 经 `persist*Repo` 装饰器写回 DB，失败会向上传播。跨 repo 写入（如 ingest、审批通过）通过 `WithTx` 保证 domain 快照与 relay 表在同一事务内提交。

迁移 SQL 唯一来源：`internal/store/postgres/migrations/`（`go:embed`）。

### 5.4 依赖注入（app.go 组装）

```go
cfg := config.Load()
st := openStore(ctx, cfg)   // Memory 或 Postgres
lifecycle := relay.NewTokenLifecycle(cfg, st, adminClient)
ingest := budget.NewIngestService(cfg, st, lifecycle, logger)
runner := worker.NewRunner(cfg, st, adminClient, lifecycle, ingest, rebalance, logger)
keysSvc := keys.NewService(cfg, st, lifecycle)
// ...
r := httpapi.NewRouter(deps)
```

禁止在 handler / service 内使用 `init()` 或包级可变全局状态。

---

## 6. API 路由与契约对齐

所有路由挂载在 **`/api`** 前缀下，与前端 `API_BASE_PATH` 一致。

### 6.1 路由注册示例

```go
r.Route("/api", func(r chi.Router) {
    r.Get("/session", sessionHandler.Get)

    r.Route("/org", func(r chi.Router) {
        r.Get("/data-source/status", orgHandler.DataSourceStatus)
        // ...
    })

    r.Route("/budget", func(r chi.Router) { /* ... */ })
    r.Route("/keys", func(r chi.Router) { /* ... */ })
    r.Route("/models", func(r chi.Router) { /* ... */ })
    r.Route("/dashboard", func(r chi.Router) { /* ... */ })
    r.Route("/audit", func(r chi.Router) { /* ... */ })
})
```

完整端点清单见 [Frontend-API契约.md](./Frontend-API契约.md) §5，实现时按域分批交付（§12）。

### 6.2 JSON 字段命名

前端 TypeScript 使用 **camelCase**。Go struct 示例：

```go
type Member struct {
    ID             string   `json:"id"`
    Name           string   `json:"name"`
    DepartmentID   string   `json:"departmentId"`
    DepartmentName string   `json:"departmentName"`
    Status         string   `json:"status"`
    Roles          []string `json:"roles"`
    Source         string   `json:"source"`
}
```

可选字段使用指针或 `omitempty`；与 Mock 返回 `{}` / `null` 的行为保持一致。

### 6.3 查询参数

| 规则                             | 实现                                                            |
| -------------------------------- | --------------------------------------------------------------- |
| 跳过 `undefined` / `null` / `""` | 与前端 `buildQuery()` 对称：空 query 用默认值                   |
| 布尔                             | 接受 `"true"` / `"false"`                                       |
| 分页                             | `page` 默认 1，`pageSize` 默认 20（与 Mock `paginate.ts` 一致） |

---

## 7. 鉴权与会话

### 7.1 Session 解析

- `GET /api/session`（无 query）：从请求解析成员身份
- 解析顺序：Cookie `tokenjoy_session_member` → `Authorization: Bearer {token}`
- 无法解析 → 401 `{ "message": "Unauthorized" }`
- Dev：前端 `/login` 写入 cookie；成员列表与 seed 对齐

### 7.2 权限

- Mock 对其他端点不做鉴权；首版后端同样**不在中间件拦截写操作**
- `permissions[]` 由 `permission` 包根据成员角色计算，规则对齐 `apps/frontend/src/lib/permissions.ts`
- 权限 key 常量对齐 `apps/frontend/src/lib/permission-keys.ts`
- 后续 Phase 2 可在 middleware 按 path + method 校验 permission key

### 7.3 CORS

开发环境允许前端 origin（`http://localhost:5173`），`credentials: true`，暴露 cookie 所需 header。

---

## 8. 种子数据迁移

### 8.1 迁移策略

| 步骤         | 说明                                                                                                         |
| ------------ | ------------------------------------------------------------------------------------------------------------ |
| 1. 导出 JSON | 将 `fixtures/*.ts` 静态数据导出为 `seed/data/*.json`（一次性脚本或手工）                                     |
| 2. 生成逻辑  | `member-factory` 的 `buildMockMembers()` 改为 Go `seed.BuildMembers()`，输入 `LEAF_DEPT_QUOTAS` + 角色分配表 |
| 3. 加载      | `seed.Load()` 读取 JSON + 执行生成逻辑，返回 `store.Snapshot`                                                |
| 4. 校验      | 启动时用 smoke test 对比关键 ID 数量（如成员数、部门数）与前端 fixture 一致                                  |

### 8.2 种子文件规划

```
seed/data/
├── org_departments.json
├── org_roles.json
├── org_permissions.json
├── org_sync_config.json
├── org_sync_logs.json
├── budget_tree.json
├── budget_groups.json
├── budget_overrun_policy.json
├── budget_alert_rules.json
├── keys_provider.json
├── keys_platform.json
├── keys_approvals.json
├── models.json
├── models_routing.json
├── dashboard_cost.json          # summary / departments / daily / top
├── dashboard_usage.json
└── audit_logs.json
```

`member-quota` 等运行时池可在 `seed.Load()` 中由成员与 Key 数据派生初始化。

### 8.3 Demo 时钟

`DEMO_TODAY` 位于 [`internal/config/config.go`](../apps/backend/internal/config/config.go)（默认 `2026-06-19`）。Demo profile 下 Dashboard period 解析与 seed 用量桶均锚定该日期。

---

## 9. 错误与状态码

与契约 §2.4 一致：

| 状态码 | 场景                         | 响应体                          |
| ------ | ---------------------------- | ------------------------------- |
| 400    | 缺必填参数、不可删预设角色等 | `{ "message": "..." }`          |
| 401    | 生产 Session 未鉴权          | `{ "message": "Unauthorized" }` |
| 404    | 资源不存在                   | `{ "message": "..." }`          |
| 422    | 额度不足、白名单校验失败等   | `{ "message": "..." }`          |
| 500    | 内部错误（首版应极少）       | `{ "message": "..." }`          |

Service 层定义：

```go
type DomainError struct {
    Status  int
    Message string
}
```

Handler 统一 `errors.As` 映射。

---

## 10. 配置与运行

### 10.1 环境变量

| 变量                     | 默认                    | 说明                               |
| ------------------------ | ----------------------- | ---------------------------------- |
| `PORT`                   | `8080`                  | HTTP 监听端口                      |
| `CORS_ORIGINS`           | `http://localhost:5173` | 逗号分隔                           |
| `SIMULATE_DELAY`         | `true`                  | 是否模拟 Mock 延迟                 |
| `DEMO_TODAY`             | `2026-06-19`            | Dashboard 演示日期                 |
| `DATABASE_URL`           | （空）                  | 非空时启用 Postgres 持久化         |
| `NEW_API_ENABLED`        | `false`                 | 启用 NewAPI Relay 集成与 worker    |
| `NEW_API_BASE_URL`       |                         | NewAPI 管理面地址（启用时必填）    |
| `NEW_API_ADMIN_TOKEN`    |                         | NewAPI Admin Token（启用时必填）   |
| `NEW_API_WEBHOOK_SECRET` |                         | Webhook 签名校验密钥（启用时必填） |

### 10.2 本地联调

```bash
# 推荐：根目录一键启动
pnpm start

# 或分别启动
cd apps/backend && go run ./cmd/server
cd apps/frontend && pnpm start   # .env.development 已配置 proxy
```

本地在 `/login` 选择成员后，所有 `/api/*` 请求经 Vite proxy 转发至 Go 后端；看板用量来自后端 `usage_buckets`（启动时 seed + webhook ingest）。

### 10.3 部署拓扑（简图）

```
Browser
  └─ GET /api/*  →  Vite dev proxy / 生产反向代理
       └─ apps/backend:8080
            ├─ memory store（默认，seed on boot + usage_buckets seed）
            └─ 可选 postgres（DATABASE_URL）+ worker（NEW_API_ENABLED）
```

本地开发与预发环境均通过真实后端 API；前端 Vitest 使用 `createMockApis` 注入，不依赖 backend 进程。

---

## 11. 测试策略

**目录：** 全部测试在 `apps/backend/tests/`（镜像 `internal/` 结构），详见 [Backend-测试指南.md](./Backend-测试指南.md)。

| 层级             | 范围                             | 工具                                 |
| ---------------- | -------------------------------- | ------------------------------------ |
| Service 单元测试 | 额度校验、路由继承、分页过滤     | `go test ./tests/...` + table-driven |
| Handler 契约测试 | 每个端点 status + JSON 形状      | `httptest` + 黄金 JSON 快照（可选）  |
| 跨域集成测试     | Session → 创建 Key → quota-check | 内存 Store 全链路                    |
| 前端测试         | `createMockApis` 注入            | 不依赖 backend 进程                  |

建议新增 `apps/backend/testdata/` 存放期望响应片段，与 `apps/frontend/tests/fixtures/` 概念对齐但不强制共享文件格式。

---

## 12. 实施阶段

### Phase 0 — 脚手架（1–2 天）

- [x] 初始化 `apps/backend/go.mod`、`cmd/server`、`config`、`chi` 路由、`/healthz`
- [x] `response` 包、`middleware`（CORS、Recover）
- [x] 根目录 / CI 增加 `go build`、`go test`

### Phase 1 — Session + Org（3–5 天）

- [x] `seed` 加载部门、角色、成员、权限
- [x] `GET /session` 双轨鉴权
- [x] Org 全部 31 端点（含分页成员列表、批量导入、角色成员绑定）
- [x] 前端：`VITE_API_PROXY_TARGET` 验证组织管理页

### Phase 2 — Budget + Keys（4–6 天）

- [x] Budget 14 端点（树、成员额度、组、策略、告警）
- [x] Keys 18 端点（provider / platform / approval / validation）
- [x] 跨域：审批通过扣预留池、Group Key 额度

### Phase 3 — Models + Dashboard + Audit（3–4 天）

- [x] Models 6 端点（路由继承、resolve）
- [x] Dashboard 7 端点（`CostQueryParams` 过滤）
- [x] Audit 4 端点（分页 + 过滤）

### Phase 4 — 收尾

- [x] 契约文档 §8.3 迁移检查清单勾选
- [x] 更新 [Frontend-API契约.md](./Frontend-API契约.md)「当前状态」为已有 backend
- [x] 可选：`pnpm dev:all` 脚本、README 联调说明（用户未要求则不写 README）

### Phase 5 — 前后端联调

- [x] 移除 MSW；前端直连 Go backend
- [x] `pnpm start` 并发 backend + frontend；`VITE_API_PROXY_TARGET` 代理
- [x] Dev `/login` 成员选择 + `tokenjoy_session_member` cookie
- [x] GET 契约测试（`tests/handler/contract_test.go`）
- [x] Session 401 跳登录
- [ ] 16 个业务页人工验收

---

## 13. CI 集成

已在 [`.github/workflows/ci.yml`](../.github/workflows/ci.yml) 增加 `backend` job，与 `pnpm lint` / `pnpm test` / `pnpm build` 并行：

```yaml
backend:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v5
    - uses: actions/setup-go@v5
      with:
        go-version: '1.24'
    - run: go build -o /dev/null ./...
      working-directory: apps/backend
    - run: go test ./tests/...
      working-directory: apps/backend
```

---

## 14. 后续演进

| 阶段     | 内容                                                              |
| -------- | ----------------------------------------------------------------- |
| 持久化   | 域数据从 JSON blob 规范化为按域拆表；Dashboard/Audit 运行时写路径 |
| 真实鉴权 | JWT / Session cookie 签名（dev cookie 已支持）                    |
| 权限网关 | 写操作 middleware 校验 `permission-keys`                          |
| 外部集成 | 飞书 / 钉钉 API、模型供应商 Key 校验                              |
| 可观测性 | 请求日志、metrics、audit 写路径落库                               |
| 契约同步 | 可选从 TS 类型生成 OpenAPI，或共享 JSON Schema                    |

---

## 15. 风险与对策

| 风险                   | 对策                                                   |
| ---------------------- | ------------------------------------------------------ |
| Mock 行为文档不全      | 以 handler 源码为行为真相；契约测试锁定响应            |
| 前后端字段漂移         | 改 API 时同步更新契约 §6 与 Go struct                  |
| 成员工厂逻辑复杂       | Phase 1 优先 port `member-factory`，对比成员 ID 列表   |
| Dashboard 只读数据量大 | 首版整包加载内存；后续按 period 索引                   |
| 双端并行维护           | Mock 冻结策略不变；新 API 只改 `api/` + backend + 契约 |

---

## 16. 变更检查清单

新增或修改 API 时，同步更新：

- [ ] `apps/frontend/src/api/{domain}.ts`
- [ ] `apps/frontend/src/api/types/{domain}.ts`
- [ ] [Frontend-API契约.md](./Frontend-API契约.md)
- [ ] `apps/backend/internal/domain/{domain}/`
- [ ] `apps/backend/internal/http/handler/{domain}.go`
- [ ] `apps/backend/internal/seed/data/`（若种子数据变化）
- [ ] `apps/backend` 契约测试
- [ ] 前端 `features/query/query-keys.ts`（若读操作变更）
