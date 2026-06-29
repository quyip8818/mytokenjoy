# TokenJoy Backend 设计

`apps/backend/` Go 服务，实现 [Frontend-API契约.md](./Frontend-API契约.md) 全部 **81** 个管理面端点。种子数据在 `internal/store/seed/`；运行时持久化于 Postgres；用量事实表 `usage_buckets` + webhook ingest。

**相关文档：** [Frontend-API契约.md](./Frontend-API契约.md) · [Backend-待实现.md](./Backend-待实现.md) · [Backend-test.md](./Backend-test.md)

---

## 1. 技术选型

| 类别 | 选型                                     |
| ---- | ---------------------------------------- |
| 语言 | Go 1.22+                                 |
| HTTP | chi v5 + 标准 `net/http`                 |
| 配置 | `caarlos0/env` 环境变量                  |
| 日志 | `log/slog`                               |
| JSON | `encoding/json`，字段 camelCase 对齐前端 |
| 测试 | `testing` + `httptest`，用例在 `tests/`  |
| DI   | 构造函数注入，组合根 `internal/app/`     |

---

## 2. 项目结构

```
apps/backend/
├── cmd/server/main.go
├── internal/
│   ├── app/                 # DI 组合根（app.go + wiring.go + registry.go）
│   ├── config/
│   ├── domain/              # session, org, budget, keys, models, dashboard, audit, usage, relay
│   ├── http/
│   │   ├── router.go        # 全局 middleware + /api 挂载
│   │   ├── handler/         # register.go 集中 API 路由；按域拆分（shared/、org/ 子包等）
│   │   ├── middleware/
│   │   └── response/
│   ├── infra/               # notification、worker、permission
│   ├── integration/         # newapi, datasource（飞书等）
│   ├── pkg/                 # budget/、org/、common/ 工具包
│   └── store/               # postgres/（运行时）；memory/（单测）；seed/；usagequery/
├── tests/                   # 镜像 internal；testutil 注入 store/memory
└── Makefile
```

域 DTO（含 org 的 `Member`、`Credential`、`SessionContext` 等）统一定义在 `internal/domain/types/`，各 domain 包仅保留 Service 接口与业务逻辑，不在包内 re-export 类型别名。

---

## 3. 分层

```
HTTP → middleware (CORS, Session, Authz, Recover)
     → handler（解析请求、写状态码）
     → domain.Service（业务规则）
     → store.Store（持久化）
```

**NewAPI（可选）：** `NEW_API_ENABLED=true` 时，`relay.TokenLifecycle` 同步 Token/Channel；`worker.Runner` 消费 outbox；`budget.IngestService` 处理 webhook 入账。

---

## 4. Store

```go
type Store interface {
    Org() OrgRepository
    Budget() BudgetRepository
    Keys() KeysRepository
    Models() ModelsRepository
    Audit() AuditRepository
    Relay() RelayRepository
    Usage() UsageRepository
    // ...
    WithTx(ctx context.Context, fn func(Store) error) error
}
```

| 模式     | 条件                          | 说明                                                                                     |
| -------- | ----------------------------- | ---------------------------------------------------------------------------------------- |
| Postgres | 运行时（必填 `DATABASE_URL`） | 域数据直读写 `domain_snapshot` JSONB + relay/usage/credential 关系表；demo 空库自动 seed |
| Memory   | 单元/Handler 测试             | `internal/store/memory` + `testutil`；`app.NewWithStore` 仅 `-tags=testhook` 构建        |

Schema SQL 唯一来源：`internal/store/postgres/schema.sql`（`go:embed`）。**不做增量迁移**；改表结构后清空 Postgres volume 重来。

**启动 bootstrap（Postgres）：** `postgres.New` → applySchema → 若 `domain_snapshot` 不完整且非 prod → `seedShards` 写入 5 个 shard；demo profile 下再 `ApplyUsageBuckets` 灌入看板用量。运行时 **不** 预热进程内 Memory；每次域读写 SELECT/UPSERT `domain_snapshot`（`WithTx` 内仅事务级 shard 缓存）。

---

## 5. 鉴权

| Profile | 环境变量                   | GET 读接口          | 写接口               |
| ------- | -------------------------- | ------------------- | -------------------- |
| Demo    | `APP_PROFILE=demo`（默认） | 多数 GET 免 Session | Session + permission |
| Prod    | `APP_PROFILE=prod`         | Session + 读权限    | Session + 写权限     |

- Session：`GET /api/session`；Cookie `tokenjoy_session_member` 或 `Authorization: Bearer`
- 权限 key 对齐 `apps/frontend/src/lib/permission-keys.ts`
- Webhook：`X-Webhook-Secret`（`/api/internal/webhooks/newapi-log`）

---

## 6. 环境变量

| 变量                                                                  | 默认                    | 说明                                                    |
| --------------------------------------------------------------------- | ----------------------- | ------------------------------------------------------- |
| `PORT`                                                                | `8080`                  | HTTP 端口                                               |
| `CORS_ORIGINS`                                                        | `http://localhost:5173` | 逗号分隔                                                |
| `APP_PROFILE`                                                         | `demo`                  | `demo` / `prod`                                         |
| `SIMULATE_DELAY`                                                      | `true`                  | 模拟数据源 test/import 延迟                             |
| `DEMO_TODAY`                                                          | `2026-06-19`            | Demo 看板锚定日期                                       |
| `DATABASE_URL`                                                        | **必填**                | Postgres 连接串；本地默认见 `config.DefaultDatabaseURL` |
| `NEW_API_ENABLED`                                                     | `false`                 | Relay + worker                                          |
| `NEW_API_BASE_URL` / `NEW_API_ADMIN_TOKEN` / `NEW_API_WEBHOOK_SECRET` | —                       | 启用 NewAPI 时必填                                      |
| `DATA_SOURCE_CREDENTIAL_KEY`                                          | —                       | 飞书等凭证 AES-GCM（32 字节 hex）                       |
| `WORKER_POLL_INTERVAL_SEC`                                            | `5`                     | Worker 轮询间隔（秒）                                   |
| `WORKER_ORG_SYNC_INTERVAL_SEC`                                        | `60`                    | 组织定时同步间隔（秒）                                  |

---

## 7. 运行与联调

```bash
# 推荐：根目录（自动起 Postgres + backend + frontend）
pnpm start

# 或分别启动
pnpm start:postgres
cd apps/backend && DATABASE_URL=postgres://tokenjoy:tokenjoy@127.0.0.1:5432/tokenjoy?sslmode=disable go run ./cmd/server
cd apps/frontend && pnpm start   # 同域 /api 反代到 :8080（需 backend 已运行）
```

Dev：访问 `/login` 选成员 → cookie → 所有 `/api/*` 经 Vite 代理到 Go。

```
Browser → /api/* → apps/backend:8080 → Postgres
                                      ├─ domain_snapshot（5 shard JSONB）
                                      ├─ usage_buckets / relay / credential ...
                                      └─ 空库首次启动由 store/seed 初始化
```

### 7.1 生产同域部署

与前端共用域名时，由边缘（nginx/Caddy 等）将 `/api/` 反代到 Go，**不得**把 `/api` 纳入 SPA `try_files` fallback。参考 [`deploy/nginx.conf.example`](../../deploy/nginx.conf.example)。

---

## 8. 错误与状态码

与契约 §2.4 一致：`{ "message": "..." }`。Service 返回 `domain.DomainError`，Handler 映射 400/401/403/404/422/500。

---

## 9. 测试

全部在 `apps/backend/tests/`。`make test-unit`；Postgres 集成 `make test-integration`。详见 [Backend-test.md](./Backend-test.md)。

---

## 10. 看板用量（US-13）

Dashboard 域**全部 GET、无副作用**；端点与类型见契约 §5.6。

```mermaid
flowchart TB
  subgraph write [写入]
    NA[NewAPI settle] --> WH[webhook] --> ING[ingest] --> UB[(usage_buckets)]
  end
  subgraph read [只读]
    API["GET /dashboard/*"]
    API -->|day,hour| UB
    API -->|minute| LOGS[ListLogs 聚合]
  end
```

| 决策     | 说明                                                                  |
| -------- | --------------------------------------------------------------------- |
| hour 桶  | 只持久化 hour；day/week/month 查询时 `date_trunc` 聚合                |
| minute   | 不落库；`log_aggregator.go` 代理 NewAPI，窗口 ≤3h                     |
| consumed | 看板读 **buckets 周期 SUM**，不读 snapshot `budget tree.Consumed`     |
| 时区     | UTC 存桶；展示默认 `Asia/Shanghai`                                    |
| 读写分离 | Handler 禁止注入 `IngestService`；与 worker `compensateLogs` 代码分离 |

**minute 语义：** `source: "logs"`、`approximate: true`、mapping 取查询时刻；禁止与 hour/day 混合环比。NewAPI 不可用 → 503 + `retryAfter`。

**一致性：** 月初 budget 重置只清 snapshot `Consumed`，buckets 保留全量历史；ingest 成本写入后不回溯。

关键代码：`internal/domain/usage/`（`log_aggregator.go`、`cost_from_log.go`）、`internal/domain/dashboard/`、`internal/store/postgres/schema.sql`（`usage_buckets` 表）。

---

## 11. 维护要点

| 项               | 说明                                                                         |
| ---------------- | ---------------------------------------------------------------------------- |
| HTTP 错误出口    | 收敛到 `httputil` / `writeDomainError`                                       |
| 读鉴权一致性     | prod profile 下各域 GET 挂 Session + 读权限                                  |
| Context 传递     | domain 内避免 `context.Background()` 滥用                                    |
| org.Service 体量 | 按需拆子 interface                                                           |
| 存储演进         | snapshot 写放大时按域分片；详见 [Backend-存储架构.md](./Backend-存储架构.md) |
| Worker 测试      | `app.WithoutWorker()`                                                        |

功能 backlog 见 [Backend-待实现.md](./Backend-待实现.md)。

---

## 12. 变更检查清单

- [ ] `apps/frontend/src/api/{domain}.ts` + `api/types/`
- [ ] [Frontend-API契约.md](./Frontend-API契约.md)
- [ ] `apps/backend/internal/domain/{domain}/`
- [ ] `apps/backend/internal/http/handler/`
- [ ] `apps/backend/internal/store/seed/`（若 demo 数据变）
- [ ] `tests/handler/contract_test.go`（新 GET）
- [ ] `features/query/query-keys.ts`（新读操作）
