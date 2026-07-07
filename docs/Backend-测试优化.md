# Backend 测试：PostgreSQL + 每测独立 Schema

测试与生产共用 `postgres.New`；主库按测试隔离 schema，ingest 相关表库级共享并由串行 + TRUNCATE 保证干净。

**相关：** [Backend.md](./Backend.md) §5

---

## 1. 约定

| 项       | 约定                                                       |
| -------- | ---------------------------------------------------------- |
| Store    | `postgres.New`，与 prod 相同实现                           |
| 主库隔离 | `CREATE SCHEMA test_<hex>` + `search_path=<schema>,public` |
| 入口     | `testutil.NewTestStore` / `NewTestApp`                     |
| 依赖     | 本地与 CI 必须提供 `DATABASE_URL`（见 §3.4）               |
| build    | `go test -tags=testhook ./tests/...`                       |

本地须先 `pnpm start:postgres`。`pnpm verify` 不会自动起库；CI `verify` job 自带 postgres service。

---

## 2. 测试分层（含原 integration）

**不再单独跑 `make test-integration` / `//go:build integration`。** 凡依赖 PostgreSQL 的用例统一进 `go test -tags=testhook ./tests/...`，与 domain/handler 同一条命令。

| 目录 | 原角色 | 现在 |
| ---- | ------ | ---- |
| `tests/store/postgres/` | 独立 integration job（`DATABASE_URL` + 无 schema 隔离） | **Store 集成**：roundtrip、persist、credential、log_repo；走 `NewTestStore` per-schema |
| `tests/domain/**`、`tests/handler/**` | memory 单测 | **领域/HTTP 集成**：真实 SQL + seed |
| `tests/worker/**`、`tests/handler/gateway/` | memory + 假 log URL | **Ingest 集成**：`WithIngestEnabled` + 真 `newapi.*` 表 |
| `tests/integration/datasource/` | Feishu HTTP 客户端 | **外部 API 集成**（mock server，不依赖 PG store） |
| `tests/pkg/**` | 纯函数 | 无 store，最快 |

```bash
# 唯一入口（含上述全部）
make test-unit

# 只跑 store 集成
go test -tags=testhook ./tests/store/postgres/...

# 只跑 Feishu 集成
go test -tags=testhook ./tests/integration/...
```

**与旧方案差异：** 过去 `verify` 不启 PG，仅 `backend-integration` job 跑 `tests/store/postgres`；现在 **CI `verify` 必启 postgres**，所有 PG 用例一次跑完。

---

## 3. 测试基础设施

### 3.1 `tests/testutil/pgschema.go`

- `openTestSchema(t)` → `(baseURL, schemaURL)`
- admin 连接 `CREATE SCHEMA` / cleanup `DROP SCHEMA ... CASCADE`
- `sync.Map` 按 `t.Name()` 复用 schema（同 test 内多次 `NewTestStore` 不重复 DDL）
- `ingestTestMu`：ingest 用例串行化（全局 `newapi.*` / `backend.*` 表）

### 3.2 `tests/testutil/store.go` — `NewTestStore`

1. `TestConfig(opts...)` + per-schema `DATABASE_URL`
2. `WithIngestEnabled` 时 `LogDatabaseURL = schemaURL`，启动/结束 `TruncateLogTables`
3. `clearDemoRuntimeSeed`：TRUNCATE `usage_buckets`、`company_recharge_orders`（多数 domain 测要空桶；HTTP 测由 `NewTestApp` 补回 demo 数据）
4. cleanup `st.Close()`

### 3.3 `tests/testutil/app.go` — `NewTestApp`

- **双 profile**：`cfg.IsProdProfile()` 时 store 层强制 `ProfileDemo`（prod contract / dashboard 用例）
- `NewTestStore` 会清空 runtime seed；demo profile 下再 `ApplyUsageBuckets` / `ApplyRechargeOrders` 恢复看板类 fixture

### 3.4 `tests/testutil/config.go`

- `defaultTestDatabaseURL()`：`DATABASE_URL` env → `config.DefaultDatabaseURL`
- `WithIngestEnabled(true)`：设置 `LogDatabaseURL` 与 webhook secret
- `WithConfig(cfg)`：注入完整配置（Feishu / authz 等场景）

### 3.5 `internal/store/postgres/log_testhook.go`（`//go:build testhook`）

| 函数                      | 池            | 用途                                    |
| ------------------------- | ------------- | --------------------------------------- |
| `MainPool` / `LogPool`    | 主库 / 日志库 | 测试直连                                |
| `InsertConsumeLog`        | LogPool       | 写 `newapi.logs`                        |
| `TruncateLogTables`       | LogPool       | 清 ingest 三张表 + 重置 cursor          |
| `GetIngestFailureByLogID` | LogPool       | 断言失败记录                            |
| `ListPendingRelayOutbox`  | MainPool      | 断言 outbox（无 `next_retry` 时钟过滤） |

`tests/testutil/pgquery.go`、`log.go` 封装上述能力供断言使用。

### 3.6 Makefile

```makefile
export DATABASE_URL ?= postgres://tokenjoy:tokenjoy@127.0.0.1:5432/tokenjoy?sslmode=disable

test-unit:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL required; run: pnpm start:postgres"; exit 1)
	go test -tags=testhook ./tests/...
```

### 3.7 CI

`verify` job：postgres:17（`tokenjoy/tokenjoy@tokenjoy`），`DATABASE_URL=postgres://tokenjoy:tokenjoy@localhost:5432/tokenjoy?sslmode=disable`，外加 `SESSION_SECRET`、`COMPANY_NAME`。

---

## 4. 隔离规则

### 4.1 两层存储

```mermaid
flowchart TB
    subgraph perTest ["per-test schema"]
        DDL["schema.sql 36 表"]
        NL[notification_log]
        RO[outbox]
    end
    subgraph global ["库级全局"]
        N[newapi.logs]
        B[backend.ingest_failures]
        C[backend.reconcile_cursors]
    end
    perTest -->|"search_path"| DDL
    global -->|"WithIngestEnabled"| ingestTests[ingest 测试]
```

| 范围                     | 策略                      | `t.Parallel()` |
| ------------------------ | ------------------------- | -------------- |
| 主库 per-schema          | 每测独立 schema           | 允许           |
| `newapi.*` / `backend.*` | `ingestTestMu` + TRUNCATE | **禁止**       |

### 4.2 Ingest 约束

凡 `WithIngestEnabled(true)`：`tests/worker/`、`handler/gateway/`、`domain/usage/`、`infra/ingestmetrics/` 等不得 `t.Parallel()`。脏数据残留可 `pnpm docker:reset`。

### 4.3 性能

- Schema 复用降低重复 DDL
- 基线（2026-07-07）：`make test-unit` cached ~2s，冷跑 ~60–70s
- `tests/pkg/*` 不依赖 store

---

## 5. 写测试

```go
// store
cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))

// ingest（须 WithIngestEnabled）
cfg, st := testutil.NewTestStore(t, testutil.WithIngestEnabled(true))
testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(2001, 99))

// HTTP
app := testutil.NewTestApp(t, func(c *config.Config) {
    testutil.WithSupportSaas(true)(c)
})

// 定制成员集（repo 裁剪，勿 snapshot）
members, _ := st.Org().Members(testutil.Ctx())
// filter → st.Org().SetMembers(...)
```

多租户 FK：先 `st.Company().Create(...)` 再写子表。

---

## 6. 回归锚点

- Store 集成：`tests/store/postgres/roundtrip_test.go`、`persist_test.go`、`log_repo_test.go`

- Prod contract：`tests/handler/core/contract_prod_test.go`
- Ingest：`tests/worker/ingest_worker_test.go`、`tests/handler/gateway/webhook_test.go`
- SaaS 隔离：`tests/handler/platform/company_isolation_test.go`
- Budget 裁剪 seed：`tests/domain/budget/service_test.go`（`TestUpdateMemberQuotaSuccess`）

---

## 7. 验收

```bash
pnpm start:postgres
cd apps/backend && make test-unit && make lint
pnpm verify
```
