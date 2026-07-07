# Backend 测试优化（实现规格）

移除 `internal/store/memory/`，测试与生产统一走 `postgres.New`。本文按**当前代码实现**编写，可直接作为开发与 PR 拆分依据。

**相关：** [Backend.md](./Backend.md) §5 · [Backend-架构.md](./Backend-架构.md) · [Backend-存储.md](./Backend-存储.md)

---

## 1. 架构决策（ADR）

| 项 | 决策 | 依据 |
| -- | ---- | ---- |
| 存储实现 | 仅保留 `internal/store/postgres/` | `app.openStore` 已固定 `postgres.New`（`internal/app/app.go`） |
| 删除 | `internal/store/memory/`（27 文件，~2,650 行） | `internal/` 无 memory 引用，纯测试遗产 |
| 主库隔离 | 每测 `CREATE SCHEMA` + `search_path` | `schema.sql` 表名未限定 schema，跟随 `search_path` |
| 日志库隔离 | **全局** `newapi` / `backend` schema + 测后 TRUNCATE | `logs_schema.sql` 硬编码 `newapi.logs`、`backend.ingest_failures` |
| prod 改动 | 最小：`postgres` 增加 `testhook` 入口 | 不污染 `cmd/server` 与正常运行路径 |
| build tag | 保持 `-tags=testhook` 跑全量 `tests/` | 与 `Makefile`、`app.NewWithStore` 一致 |

---

## 2. 现状（代码锚点）

### 2.1 生产启动链

```
cmd/server/main.go
  → app.New(cfg, logger)
    → openStore() → postgres.New(ctx, cfg)
      → applySchema(schema.sql)           // internal/store/postgres/schema.go
      → ensureBootstrapCompany()
      → [IngestEnabled] applyLogsSchema() // internal/store/postgres/log_repo.go
      → loadOrSeedDomain() → seed.ApplyTables()  // 空库时
      → [IsDemoProfile] seed.ApplyUsageBuckets / ApplyRechargeOrders
```

`config.validate()` 已要求 `DATABASE_URL` 必填（`internal/config/config.go:84`）。文档中「无 PG 用 memory」描述已过时。

### 2.2 测试启动链（待替换）

```
tests/testutil/store.go
  NewMemoryStoreFromConfig(t)
    → memory.New(seed.Load(cfg))

tests/testutil/app.go
  NewTestApp(t) → NewMemoryStore(t, cfg) → app.NewWithStore(cfg, logger, st, WithoutWorker())
```

**引用 memory 的测试文件：52 个**（`rg 'NewMemoryStore|memory\.New|store/memory' tests/`）。

### 2.3 已有 PG 集成测试

| 文件 | 模式 |
| ---- | ---- |
| `tests/store/postgres/helpers_test.go` | `//go:build integration`，`testPostgresStore(t)` |
| `tests/store/postgres/*_test.go` | 7 个文件，共用 `helpers_test.go` |

问题：与 `test-unit` 分离，且**无 schema 隔离**（共用 `tokenjoy_test` 库）。

### 2.4 双库模型

| 连接 | 配置项 | 表 / schema | 测试注意 |
| ---- | ------ | ----------- | -------- |
| 主库 | `DATABASE_URL` | `schema.sql` 36 表（无 schema 前缀） | 可用 per-test `search_path` |
| 日志库 | `LOG_DATABASE_URL` | `newapi.logs`、`backend.reconcile_cursors`、`backend.ingest_failures` | **不受** 主库 `search_path` 影响 |

`WithIngestEnabled(true)` 当前写死 `LogDatabaseURL = "postgres://memory/logs"`（`tests/testutil/config.go:58`），仅 memory `log.go` 能工作；迁 PG 后**必须改为真实 URL**。

---

## 3. 目标架构

```mermaid
flowchart TB
    subgraph testutil [tests/testutil]
        NTS[NewTestStore]
        NTA[NewTestApp]
        SCH[pgschema.go]
        LOG[log.go SeedConsumeLog]
    end

    subgraph main_pg [DATABASE_URL]
        S1["schema test_<uuid>"]
        S2["schema test_<uuid>"]
    end

    subgraph log_pg [LOG_DATABASE_URL]
        NA[newapi.logs]
        BE[backend.ingest_failures]
    end

    NTS --> SCH
    SCH -->|CREATE + search_path| S1
    SCH -->|CREATE + search_path| S2
    NTA --> NTS
    LOG -->|INSERT| NA
    NTS -->|t.Cleanup DROP SCHEMA| main_pg
    LOG -->|t.Cleanup TRUNCATE| log_pg
```

### 3.1 两类测试隔离策略

| 类型 | 条件 | 隔离 | 并行 |
| ---- | ---- | ---- | ---- |
| **A. 域存储** | 默认；无 `WithIngestEnabled` | 主库 per-schema | `t.Parallel()` 安全 |
| **B. Ingest / Worker** | `WithIngestEnabled(true)` | 主库 per-schema + 日志库 TRUNCATE | `tests/worker`、`tests/store/memory/log_*` 包级 `t.Parallel()` **禁用**；或 ingest 子测串行 |

---

## 4. 生产代码改动（最小）

### 4.1 新增 `internal/store/postgres/open_testhook.go`

```go
//go:build testhook

package postgres

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/tokenjoy/backend/internal/config"
    "github.com/tokenjoy/backend/internal/store"
)

// NewWithSearchPath opens a store whose DDL/DML target the given PostgreSQL schema.
// searchPath must be a safe identifier (e.g. test_<uuid>); validated in testutil.
func NewWithSearchPath(ctx context.Context, cfg config.Config, searchPath string) (store.Store, error) {
    poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
    if err != nil {
        return nil, fmt.Errorf("parse database url: %w", err)
    }
    poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
        _, err := conn.Exec(ctx, "SET search_path TO "+quoteIdent(searchPath))
        return err
    }
    // ... pgxpool.NewWithConfig → 复用现有 New 内逻辑（applySchema、seed 等）
}
```

**实现要点：**

- 从 `postgres.New` 抽取 `openPools(ctx, cfg)` 与 `initStore(ctx, cfg, pool, logPool)`，避免重复。
- `applySchema` 无需改：`CREATE TABLE IF NOT EXISTS companies` 在 `search_path` 指向的 schema 内建表。
- `isDatabaseEmpty`（`SELECT COUNT(*) FROM members`）随 `search_path` 只查当前 schema。
- `quoteIdent` 仅允许 `[a-z0-9_]`，由 `testutil` 生成 schema 名。

**prod 运行时：** 不编译此文件（无 `testhook` tag），行为不变。

### 4.2 可选：日志测试种子 `open_testhook.go` 同文件或 `log_testhook.go`

```go
//go:build testhook

func InsertConsumeLog(ctx context.Context, pool *pgxpool.Pool, raw store.RawConsumeLog) error
func TruncateLogTables(ctx context.Context, pool *pgxpool.Pool) error
```

对应 SQL（与 `log_repo.go` 查询一致）：

```sql
INSERT INTO newapi.logs (id, token_id, quota, model_name, created_at, type, ...)
VALUES ($1, $2, ...);

TRUNCATE newapi.logs, backend.ingest_failures, backend.reconcile_cursors RESTART IDENTITY CASCADE;
-- 重新 INSERT reconcile_cursors 默认值
```

不在 `LogStore` 接口上增加 `PutConsumeLog`，避免 prod API 污染。

### 4.3 不改动

| 模块 | 原因 |
| ---- | ---- |
| `store.Store` 接口 | 不变 |
| `internal/store/seed/` | `postgres.loadOrSeedDomain` 已调用 `seed.ApplyTables` |
| `internal/store/clone_*.go` | postgres `keys_repo`、`org_repo_roles` 已使用 |
| `cmd/server` | 无 memory 分支 |

---

## 5. testutil 实现规格

### 5.1 环境变量与默认值

| 变量 | 测试默认 | 来源 |
| ---- | -------- | ---- |
| `DATABASE_URL` | `config.DefaultDatabaseURL` | `internal/config/database.go` |
| `LOG_DATABASE_URL` | 与 `DATABASE_URL` 同实例（见下） | `TestConfig` 注入 |
| `SESSION_SECRET` | `TestSessionSecret` | 已有 |

`TestConfig` 修改：

```go
func TestConfig(opts ...ConfigOption) config.Config {
    cfg := config.Config{
        DatabaseURL: os.Getenv("DATABASE_URL"), // fallback DefaultDatabaseURL
        DemoToday:   defaultDemoToday,
        // ...
    }
    if cfg.DatabaseURL == "" {
        cfg.DatabaseURL = config.DefaultDatabaseURL
    }
    // ...
}

func WithIngestEnabled(enabled bool) ConfigOption {
    return func(cfg *config.Config) {
        if enabled {
            cfg.LogDatabaseURL = cfg.DatabaseURL // 同实例；newapi/backend 为全局 schema
            cfg.NewAPIWebhookSecret = "test-webhook-secret"
        } else {
            cfg.LogDatabaseURL = ""
        }
    }
}
```

### 5.2 新文件 `tests/testutil/pgschema.go`

```go
func requireDatabaseURL(t *testing.T) string
func newTestSchemaName() string                    // test_<hex>
func createSchema(t *testing.T, adminURL, name string)
func dropSchema(t *testing.T, adminURL, name string)

func NewTestStore(t *testing.T, opts ...ConfigOption) (config.Config, store.Store) {
    t.Helper()
    cfg := TestConfig(opts...)
    schema := newTestSchemaName()
    createSchema(t, cfg.DatabaseURL, schema)
    st, err := postgres.NewWithSearchPath(context.Background(), cfg, schema)
    require.NoError(t, err)
    if cfg.IngestEnabled() {
        registerLogCleanup(t, cfg.LogDatabaseURL)
    }
    t.Cleanup(func() {
        st.Close()
        dropSchema(t, cfg.DatabaseURL, schema)
    })
    return cfg, st
}
```

**删除：** `NewMemoryStore`、`NewMemoryStoreFromConfig`（全库替换为 `NewTestStore`）。

### 5.3 修改 `tests/testutil/app.go`

```go
func NewTestApp(t *testing.T, mutate func(*config.Config)) *app.App {
    cfg := TestConfig()
    if mutate != nil {
        mutate(&cfg)
    }
    _, st := NewTestStore(t, func(c *config.Config) { *c = cfg })
    // postgres.New 已在 demo profile 下执行 ApplyUsageBuckets / ApplyRechargeOrders
    // 删除手动 seed.ApplyUsageBuckets 重复调用（与 postgres.New 对齐）
    ...
}
```

核对：`postgres.New` 在 `IsDemoProfile()` 时已调用 `seed.ApplyUsageBuckets`（`postgres.go:74-88`），`NewTestApp` 中重复调用可删。

### 5.4 内省 helper 改造（`memory.go` → 删除或改写）

| 现函数 | memory 实现 | PG 实现 |
| ------ | ----------- | ------- |
| `NewMemoryStoreFromSnapshot` | `memory.New(snapshot)` | `NewTestStore` + 用 `Org().Members()` 等 API 改数据；或 `seed.ApplyTables` 到当前 schema 的 testhook 辅助函数 |
| `UsageBucketRows(st)` | `mem.UsageBucketRows()` | `st.Usage().QueryFilteredBuckets(ctx, wideQuery)` |
| `NotificationLogs(st)` | `mem.NotificationLogs()` | testhook SQL：`SELECT * FROM notification_log WHERE company_id = $1` |
| `RelayOutboxEntry(st, id)` | `mem.RelayOutboxEntry(id)` | `ClaimPendingRelayOutbox` 过滤，或 testhook：`SELECT * FROM relay_outbox WHERE id = $1` |
| `SeedConsumeLog(t, st, raw)` | `mem.PutConsumeLog` | `postgres.InsertConsumeLog`（testhook） |
| `AssertIngestFailure(t, st, logID, source)` | `mem.IngestFailureByLogID` | `ClaimPendingFailures` / testhook 查 `backend.ingest_failures` |

**`budget/service_test.go` 快照定制用例：**

```go
// 旧：裁剪 snapshot.Members → NewMemoryStoreFromSnapshot
// 新：NewTestStore 后调用 Org repo 删除/更新成员，或 testhook seed.ApplyTables 写入裁剪数据
```

### 5.5 修改 `tests/testutil/org/service.go`、`relay/gateway.go`、`worker/runner.go`

统一将 `NewMemoryStoreFromConfig` 替换为 `NewTestStore`，签名保持不变。

---

## 6. 测试包与 build tag

### 6.1 合并 integration tag

| 阶段 | `tests/store/postgres/*` | 其余 `tests/**` |
| ---- | ------------------------ | --------------- |
| 迁移前 | `//go:build integration` | 无 tag |
| 迁移后 | **去掉** `integration` tag | 无 tag |

全部用 `go test -tags=testhook ./tests/...`，单一入口。

### 6.2 Makefile

```makefile
export DATABASE_URL ?= postgres://tokenjoy:tokenjoy@127.0.0.1:5432/tokenjoy?sslmode=disable

test-unit:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL required; run: pnpm start:postgres"; exit 1)
	go test -tags=testhook ./tests/...

test-integration:
	@echo "merged into test-unit"
	$(MAKE) test-unit
```

### 6.3 CI（`.github/workflows/ci.yml`）

`verify` job 增加 postgres service（复制 `backend-integration` 块），env：

```yaml
DATABASE_URL: postgres://postgres:postgres@localhost:5432/tokenjoy_test?sslmode=disable
SESSION_SECRET: test-session-secret
COMPANY_NAME: Demo Company
```

`backend-integration` job 可与 `verify` 合并或保留为冗余至迁移完成。

---

## 7. 删除清单

```
internal/store/memory/                    # 整目录
tests/store/memory/                       # 整目录（用例迁至 PG 或删除等价覆盖）
tests/testutil/memory.go                  # 内省迁 pg helper 后删除
tests/testutil/store.go                   # 改为 NewTestStore（或合并 pgschema.go）
```

**全库确认：**

```bash
rg 'store/memory|memory\.New|NewMemoryStore' apps/backend
# 期望：0 结果
```

---

## 8. PR 拆分（推荐顺序）

### PR-1：基础设施（可独立合并）

| 任务 | 文件 |
| ---- | ---- |
| `postgres.NewWithSearchPath` + 重构 `postgres.New` | `internal/store/postgres/open_testhook.go`，`postgres.go` |
| `InsertConsumeLog` / `TruncateLogTables` | `internal/store/postgres/log_testhook.go` |
| `pgschema.go`、`NewTestStore` | `tests/testutil/pgschema.go` |
| 更新 `WithIngestEnabled`、`TestConfig` | `tests/testutil/config.go` |
| 单测验证 schema 隔离 | `tests/testutil/pgschema_test.go` |

验收：`go test -tags=testhook ./tests/testutil/...`

### PR-2：testutil 入口与 app

| 任务 | 文件 |
| ---- | ---- |
| `NewTestApp` / `NewTestRouter` 改用 `NewTestStore` | `app.go` |
| `org/service.go`、`relay/gateway.go`、`worker/runner.go` | 替换入口 |
| 改写 `log.go` 内省 | `SeedConsumeLog`、`AssertIngestFailure` |

验收：跑通 `tests/handler/core/` + `tests/domain/audit/`

### PR-3：domain 批量迁移（可按子目录拆）

```
tests/domain/budget/   → tests/domain/org/   → tests/domain/usage/
tests/domain/keys/   → tests/domain/relay/ → tests/domain/company/
... 其余 domain 包
```

每包 PR 替换 `NewMemoryStoreFromConfig` → `NewTestStore`，改写 snapshot / 内省调用。

### PR-4：handler + identity + infra + worker

```
tests/handler/**     tests/identity/**
tests/infra/**       tests/worker/**
```

`worker` 包测试结束时 `TruncateLogTables`；包内禁止 `t.Parallel()`。

### PR-5：store 测试合并 + 删 memory

| 任务 | 说明 |
| ---- | ---- |
| `tests/store/postgres/helpers_test.go` | 改用 `NewTestStore` 或共享 `pgschema` |
| 去掉 `//go:build integration` | 8 个 postgres 测试文件 |
| 删除 `internal/store/memory/` | |
| 删除 `tests/store/memory/` | |
| 更新 Makefile、CI、Backend.md | |

验收：

```bash
pnpm start:postgres
cd apps/backend && make test-unit && make lint
pnpm verify
```

---

## 9. 文件级迁移清单（52 文件）

### 9.1 testutil（必先改）

- `tests/testutil/store.go`
- `tests/testutil/app.go`
- `tests/testutil/config.go`
- `tests/testutil/memory.go` → 删除
- `tests/testutil/log.go`
- `tests/testutil/org/feishu_env.go`
- `tests/testutil/org/service.go`
- `tests/testutil/relay/gateway.go`
- `tests/testutil/worker/runner.go`

### 9.2 直接 `memory.New`（需单独处理）

- `tests/domain/billing/service_test.go`
- `tests/domain/company/accept_invite_test.go`
- `tests/domain/company/create_company_test.go`
- `tests/worker/runner_test.go`（`*memory.Store` cast）
- `tests/worker/ingest_worker_test.go`（`IngestFailureByLogID`）

### 9.3 `tests/store/memory/` → 删除或并入 `tests/store/postgres/`

- `ledger_repo_test.go`
- `log_repo_test.go` → 迁 PG + `WithIngestEnabled`
- `org_repo_test.go`
- `usage_test.go`

### 9.4 其余

凡 `rg -l NewMemoryStoreFromConfig tests/` 列出的文件，机械替换为 `NewTestStore`（**52 文件**）。

---

## 10. 性能与后续优化

### 10.1 预期耗时

| 阶段 | 全量 `go test -tags=testhook ./tests/...` |
| ---- | ---------------------------------------- |
| 初版（每测 schema + migrate） | ~45–90s |
| 优化后（template schema） | ~25–45s |

### 10.2 Phase 2：template schema（可选）

```sql
-- 一次性（TestMain）
CREATE SCHEMA IF NOT EXISTS test_template;
SET search_path TO test_template;
-- applySchema + seed

-- 每测
CREATE SCHEMA test_<uuid> ;
CREATE TABLE test_<uuid>.companies (LIKE test_template.companies INCLUDING ALL);
-- 或 pg_dump --schema-only 克隆
```

在 `tests/testutil/pgschema_test.go` 基准达标后再做，不阻塞 PR-1–5。

### 10.3 并行策略

| 包 | `t.Parallel()` |
| -- | -------------- |
| `tests/pkg/**` | 允许 |
| `tests/domain/**`（无 ingest） | 允许 |
| `tests/handler/**` | 允许 |
| `tests/worker/**`、`log_repo` 相关 | **禁止**（共享 `newapi`/`backend`） |

---

## 11. 风险与约束

| 风险 | 缓解 |
| ---- | ---- |
| 日志表全局共享，并行 ingest 测试互相污染 | ingest 包串行 + `TruncateLogTables` cleanup |
| `CREATE SCHEMA` + 全量 migrate 慢 | template schema（§10.2）；CI 缓存 PG |
| `NewMemoryStoreFromSnapshot` 用例 | 改为 repo API 准备数据 |
| 本地未起 PG | `make test-unit` 明确报错；与 `pnpm start` 一致 |
| `platform_operators` 等全局表在 test schema 内 | 每 schema 独立一份，与多租户模型一致 |

---

## 12. 断言与 HTTP DSL（并行轨道）

Store 迁移与测试 DX 优化可并行，不阻塞 PR-1：

| 项 | 说明 |
| -- | ---- |
| `testify/require` + `go-cmp` | 见原断言规范 |
| `testhttp.Client` | 减少 handler 样板 |
| table-driven | 合并 session/auth 等重复用例 |

Store 迁移完成后，handler 改造统一基于 `NewTestApp`（已是 PG）。

---

## 13. 验收检查清单

- [ ] `rg 'store/memory|NewMemoryStore|memory\.New' apps/backend` 无结果
- [ ] `go test -tags=testhook ./tests/...` 全绿（`DATABASE_URL` 已设）
- [ ] `go test`（无 tag）不包含测试包失败
- [ ] `pnpm verify` 通过（含 PG service）
- [ ] `internal/` 无 `*_test.go`
- [ ] ingest 测试跑后 `newapi.logs` 无残留（cleanup 验证）
- [ ] [Backend.md](./Backend.md) 环境变量表与 §5 测试说明已更新
