# Backend · 离线任务（as-built）

> **定位**：离线任务现状说明（consumed 在 Ingest 同事务写入；无独立 `budget_projection`）。  
> **入账**：[Backend-Ingest架构.md](./Backend-Ingest架构.md) · [Backend-预算.md](./Backend-预算.md)  
> Schema / Unique：`internal/infra/jobs/`

---

## 1. 运行时概览

进程内 **单一异步栈**：全部离线工作都是 `river_job`，由一个 `river.Client` 统一 claim/retry/Periodic：

```text
cmd/server
  ├─ HTTP / Gateway
  └─ infra/river.Client.Start()   ← 唯一异步入口：claim + Workers + Periodic
```

装配：`internal/app/compose_worker.go` → `buildBackgroundWorkers`。

```mermaid
flowchart TB
  subgraph river [river.Client]
    WH[webhook] -->|InsertIngest| ING[kind=ingest]
    ING --> ING_W[IngestWorker] --> LEDGER[ledger + budget_consumed + combined_key]
    LEDGER -->|post-commit HTTP| WS[set_quota override → NewAPI]
    LEDGER -.->|仅可能触顶| OV[overrun_可选]

    PER1[Periodic ingest_reconcile] --> IRC[kind=ingest_reconcile] --> IRC_W[IngestReconcileWorker] -->|逐条入队| ING

    WD[Periodic tenant_watchdog 1h] --> SCHED[infra/scheduler due + bulk enqueue]
    SCHED --> DP[dashboard_project] --> DPW[dashboard.Projector]
    OV --> OVW[OverrunWorker]
    RB[rebalance] --> RBW[RebalanceWorker]
    SCHED --> TENANT[per-company 子 job]

    PER2[Periodic catalog_sync] --> CS[kind=catalog_sync]（仅 SUPPORT_SAAS=false）
  end
```

**as-built：** 多数 Ingest 不入队任何 budget 投影 job；ingest 入账本身也是 River job，没有独立的日志库轮询进程。

---

## 2. 分层与依赖方向

| 层                | 路径                           | 职责                                        |
| ----------------- | ------------------------------ | ------------------------------------------- |
| **domain**        | `domain/*`                     | 业务逻辑；**不** import `river` / `pgx`     |
| **jobs**          | `internal/infra/jobs`          | `Enqueuer` 接口、Job Args、`Insert*` helper |
| **scheduler**     | `internal/infra/scheduler`     | L2 只读 due 查询 + 看门狗批量入队           |
| **river workers** | `internal/infra/river/workers` | 薄壳：`Work()` → 调 domain 一个方法         |
| **river client**  | `internal/infra/river`         | Client 装配、Periodic、队列权重             |

Domain 入队经各域 `JobEnqueuer` 端口（`domain/<域>/ports.go` + `adapter/enqueue/`）；底层统一 `jobs.Enqueuer`（`Insert` / `InsertInTx`）。事务内入队通过 `store.Tx`（`postgres.txStore` 实现），domain 不 import `pgx`。

### 2.1 Holder 与域端口

`jobs.Holder` **仅用于 bootstrap**：River Client 未就绪前占位 `NoopEnqueuer`，避免 nil。Client 启动后 `holder.Set(client.Enqueuer)`。

**业务入队走域端口**，不直接依赖 Holder：

| 域         | 端口                         | 适配器                                   |
| ---------- | ---------------------------- | ----------------------------------------- |
| budget     | `budget.JobEnqueuer`         | `adapter/enqueue/budget.go`               |
| usage      | `usage.IngestJobEnqueuer`    | `adapter/enqueue/usage_ingest.go`         |
| dashboard  | `dashboard.JobEnqueuer`      | `adapter/enqueue/dashboard.go`            |
| newapisync | `newapisync.SyncJobEnqueuer` | `adapter/enqueue/newapisync.go`           |
| org-remote | `remote.JobEnqueuer`         | `adapter/enqueue/org.go`                  |

billing 域没有独立入队端口——充值同步走 post-commit `QuotaSyncer.ManageUser` 实时 HTTP override，不经过 river_job。

`RIVER_ENABLED=false` 时 Holder 保持 `NoopEnqueuer`（`Insert` 返回 `nil`，**不入队**）。

---

## 3. Job kind（as-built）

| kind                  | 队列     | Unique          | 触发层级                             | Worker                         | Domain 入口                             |
| --------------------- | -------- | --------------- | ------------------------------------ | ------------------------------- | ---------------------------------------- |
| `ingest`              | critical | `ByArgs`（按 log_id） | webhook / reconcile 逐条入队    | `workers/ingest.go`             | `usage.IngestService.IngestByLogID`      |
| `ingest_reconcile`    | default  | 无              | L2 Periodic（`INGEST_RECONCILE_INTERVAL_SEC`） | `workers/ingest_reconcile.go`   | 扫游标后逐条入队 `ingest`                |
| `newapi_sync`         | critical | 无              | L1 业务                              | `workers/newapi_sync.go`       | `newapisync.OutboxHandler`              |
| `rebalance`           | default  | per axis        | L1 按需 / 月切 / reconcile           | `workers/rebalance.go`         | `budget.Rebalancer.ProcessAxis`         |
| `overrun`             | default  | per payload     | L1 **仅可能触顶时**                  | `workers/overrun.go`           | `budget.OverrunProcessor`               |
| `org_sync`            | default  | per company     | L1 ScheduledAt；L2 看门狗            | `workers/org_sync.go`          | `org.RunScheduledSync`                  |
| `budget_reconcile`    | low      | args，~24h      | L1 手动；L2 看门狗                   | `workers/budget_reconcile.go`  | `budget.ReconcileService.RunCompany`    |
| `dashboard_project`   | low      | args，~1h       | L1 自续；L2 看门狗（每小时检测 lag） | `workers/dashboard_project.go` | `dashboard.Projector.RunBatch`          |
| `dashboard_reconcile` | low      | args，~24h      | L1 手动；L2 看门狗                   | `workers/dashboard_project.go` | `dashboard.ReconcileService.RunCompany` |
| `tenant_watchdog`     | low      | ByPeriod = 间隔 | L2 Periodic                          | `workers/watchdog.go`          | `scheduler.CollectDue` + `BulkEnqueue`  |
| `notification_delivery` | default | 无            | L1 各域通知触发点                    | `workers/notification_delivery.go` | `notification.Registry` 按渠道投递  |
| `catalog_sync`        | default  | 无              | L2 Periodic（`SUPPORT_SAAS=false` 时注册） | `workers/catalog_sync.go`      | `catalogsync` 从 SaaS 拉取 models/pricing/currencies/wallet_lots |

**没有** `wallet_sync` / `budget_projection` kind——wallet 同步改为 post-commit 实时 HTTP override（不入队）；budget consumed 在 Ingest 同事务写入（不经异步投影）。

Args：`internal/infra/jobs/kinds_*.go`；kind 常量 SSOT：`internal/infra/jobs/catalog.go`。入队：`enqueue.go`。Worker：`infra/river/client.go` 注册。

---

## 4. `tenant_background_state`（租户后台 SSOT）

表：`tenant_background_state`（`schema.sql`）。每 active company 一行，由 `CreateCompany` / seed `EnsureRow` 初始化。

| 字段                          | 写入时机                                   | 读取方                                       |
| ----------------------------- | ------------------------------------------ | --------------------------------------------- |
| `next_org_sync_at`            | `UpdateSyncConfig` / 同步成功后 reschedule | L1 org、`scheduler.orgDue`                   |
| `last_org_sync_at`            | 同步成功                                   | `ComputeNextOrgSync`                         |
| `last_rebalanced_period`      | **仅** company 轴 `rebalance` worker 成功  | `EnsureMonthRebalance`、`scheduler.monthDue` |
| `last_budget_reconcile_at`    | `budget_reconcile` worker 成功             | `scheduler.budgetReconcileDue`               |
| `last_dashboard_reconcile_at` | `dashboard_reconcile` worker 成功          | `scheduler.dashboardReconcileDue`            |

Store：`store/tenant_background_state.go` + `postgres/tenant_background_state_repo.go`。

---

## 5. 入队点（谁写入 `river_job`）

### 5.1 事务内（`InsertInTx`，与 ledger 同事务）

**Ingest 成功路径**（as-built）：

1. `ledger` + lot + **`budget_consumed` + `combined_key_remain`**（同事务写入，无异步投影）
2. **可选** `InsertOverrun`（仅轻量预判认为可能触顶时）

**不**入队 `dashboard_project`（由看门狗小时级驱动）。rebalance 默认不在每笔 Ingest 入队（充值 / 月切 / reconcile / 按需）。wallet override 是 **commit 后的直接 HTTP 调用**，完全不经过 `river_job`。

### 5.2 L1 — 业务路径（`Insert`）

| 来源                              | kind                                        | 说明                               |
| --------------------------------- | -------------------------------------------- | ---------------------------------- |
| webhook / reconcile               | `ingest`                                     | 按 log_id 幂等                     |
| Ingest 预判触顶                   | `overrun`                                    | 可先判跳过                         |
| 按需 / Key 变更                   | `rebalance`                                  | 按轴 Unique；Key 创建/变更走同步 `RefreshPlatformKeyCombined`，不入队 |
| `schedule.EnsureMonthRebalance`   | `rebalance`（company）                       | 月切；reconcile 批首或看门狗       |
| 充值 / 消费 / 升级                | 直接 post-commit HTTP `ManageUser("set_quota")` | **不经过 river_job**            |
| `budget.ReconcileService` 修复后  | `rebalance`（company）                       |                                    |
| `newapisync/*`                    | `newapi_sync`                                |                                    |
| `org.UpdateSyncConfig` / 同步成功 | `org_sync`                                   |                                    |
| `dashboard.Projector` 批满        | `dashboard_project`                          | 自续（追赶积压）                   |
| 手动 reconcile API                | `budget_reconcile` / `dashboard_reconcile`   |                                    |
| 各域通知触发点                    | `notification_delivery`                      |                                    |

### 5.3 L2 — 看门狗（`tenant_watchdog`）

唯一「批量入队」类 Periodic：`infra/river/periodic/jobs.go`（`BuildPeriodicJobs`）→ 入队 `tenant_watchdog`。  
Worker：`workers/watchdog.go` → `scheduler.Service.CollectDue` + `BulkEnqueuer.EnqueueDue`（默认每批 200 tenant）。

另有两个独立 Periodic（同一注册函数）：`ingest_reconcile`（`INGEST_RECONCILE_INTERVAL_SEC`）与 `catalog_sync`（`CATALOG_SYNC_INTERVAL_SEC`，仅 `SUPPORT_SAAS=false` 注册）。

Due 判据（只读 store，见 `infra/scheduler/due.go`）：

- **org**：`next_org_sync_at <= now` 且无 active `org_sync` job；或 org 已启用但 `next_org_sync_at` 缺失（无 TBS 行或列为 NULL）且无 active job
- **月切 rebalance**：`last_rebalanced_period != 当前开账月`
- **dashboard 投影滞后**：ledger 游标之后仍有 settled 记录 → 触发 `dashboard_project`
- **budget reconcile**：**不依赖投影滞后**，纯按 `last_budget_reconcile_at` 超过 7 天触发
- **dashboard reconcile**：投影不滞后且 `last_dashboard_reconcile_at` 超过 7 天

---

## 6. Worker 行为摘要

### 6.1 `ingest` / `ingest_reconcile`

见 [Backend-Ingest架构.md](./Backend-Ingest架构.md) §7。wallet 同步已从异步 job 改为 Ingest / 充值路径 **post-commit 内联 HTTP** 调 `ManageUser("set_quota", mode="override")`，不再有 `wallet_sync` river job、debounce、delta 计算。

### 6.2 `rebalance` / `overrun`

- Args 带 `company_id` + axis / payload
- `RebalanceService.ProcessAxis` **只重算本地 `combined_key_remain`**，不调用 NewAPI Admin API——NewAPI token 为 `unlimited_quota`
- Gateway 独立检查 `wallet_remain_quota`（硬约束），与 `combined_key_remain` 解耦
- company 轴成功 → 写 `tenant_background_state.last_rebalanced_period`（`EnsureRow` 后 `SetLastRebalancedPeriod`）
- 充值不触发 rebalance（只触发 wallet override）；触发场景：月切、reconcile、approval、project 删除释放成员 Key

### 6.3 `org_sync`

- per-tenant job，`ScheduledAt` 由 L1 reschedule 设置
- `RunScheduledSync`：锁 `org_sync:{company_id}`；成功后更新 TBS 并 reschedule 下一条
- Worker 入口 `ensureScheduledOrgSync` 自愈：到期且无 pending job → reschedule

### 6.4 `dashboard_project`

- 看门狗每小时检测 `projection_lag`（游标后有未处理 ledger）→ 入队
- Worker 调 `dashboard.Projector.RunBatch`：按游标批量读 ledger，写 `usage_buckets`
- 批满自续（re-enqueue 追赶）

### 6.5 `budget_reconcile` / `dashboard_reconcile`

- per-company 对比 ledger 与投影表，修复 drift
- 成功 → 写 `last_budget_reconcile_at` / `last_dashboard_reconcile_at`

### 6.6 `tenant_watchdog`

- 扫描全 active company，批量入队 L2 补课 job（见 §5.3）
- **每小时一轮**，是 `dashboard_project` 的唯一触发源（Ingest 不再入队）
- reconcile 由各自 `staleWindow`（7 天）控制频率

### 6.7 `notification_delivery`

- 单条通知的异步渠道投递（Email/InApp/Webhook 等），由 `notification.Registry` 按 `Channel` 分发
- 渠道未注册或未配置 → `river.JobCancel`（不重试）

### 6.8 `catalog_sync`

- 仅 `SUPPORT_SAAS=false`（私有化本地部署）注册
- 定时从 SaaS 平台拉取 models / pricing / currencies / wallet_lots 增量数据

---

## 7. Periodic 与启动时看门狗

| 机制                           | 何时跑                                                                             | 是否挡启动                      |
| ------------------------------ | ---------------------------------------------------------------------------------- | -------------------------------- |
| **`tenant_watchdog` Periodic** | `RIVER_PERIODIC_ENABLED=true`（默认）时每 `WATCHDOG_INTERVAL_SEC`（1h）            | 否（River 后台 goroutine）      |
| **`ingest_reconcile` Periodic** | `IngestEnabled()` 时每 `INGEST_RECONCILE_INTERVAL_SEC`                            | 否                                |
| **`catalog_sync` Periodic**    | `SUPPORT_SAAS=false` 时每 `CATALOG_SYNC_INTERVAL_SEC`                              | 否                                |
| **Deferred 首次扫描**          | 进程启动 `WATCHDOG_STARTUP_DELAY_SEC`（默认 5s）后 `scheduler.RunOnce`             | **否**——只入队，Worker 后台消费 |
| **`/healthz`**                 | 立即可用                                                                           | —                                 |

注册 Periodic：`internal/infra/river/periodic/jobs.go`（`BuildPeriodicJobs`，需 `RIVER_ENABLED` + `RIVER_PERIODIC_ENABLED`）；执行体分别是 `workers/watchdog.go`、`workers/ingest_reconcile.go`、`workers/catalog_sync.go`。

Deferred 入队：`compose_watchdog.go` → `startDeferredWatchdog`（`app.go` 在 Worker 启动后调用）。

| env                          | 默认   | 含义                                           |
| ----------------------------- | ------ | ---------------------------------------------- |
| `WATCHDOG_INTERVAL_SEC`      | `3600` | Periodic 间隔（1h，驱动 dashboard projection） |
| `WATCHDOG_STARTUP_DELAY_SEC` | `5`    | 启动后首次 due 扫描延迟                        |
| `WATCHDOG_BULK_BATCH_SIZE`   | `200`  | 每批 tenant 数                                 |

---

## 8. 配置

| 变量                         | 默认   | 含义                                                     |
| ---------------------------- | ------ | -------------------------------------------------------- |
| `RIVER_ENABLED`              | `true` | 是否启动 River Client                                    |
| `RIVER_PERIODIC_ENABLED`     | `true` | 是否注册 Periodic（tenant_watchdog / ingest_reconcile / catalog_sync） |
| `RIVER_MAX_WORKERS`          | `20`   | 全局 worker 上限；按 2:2:1 分到 critical / default / low |
| `WATCHDOG_INTERVAL_SEC`      | `3600` | 看门狗 Periodic 间隔（1h，驱动 dashboard projection）    |
| `WATCHDOG_STARTUP_DELAY_SEC` | `5`    | 启动后 deferred due 扫描延迟（不挡 health）              |
| `WATCHDOG_BULK_BATCH_SIZE`   | `200`  | 看门狗每批处理 tenant 数                                 |
| `INGEST_RECONCILE_INTERVAL_SEC` | `300` | ingest_reconcile Periodic 间隔                          |
| `INGEST_RECONCILE_BATCH_SIZE`   | `500` | 单批扫描 log 数                                         |
| `INGEST_RECONCILE_MAX_ROUNDS`   | `10`  | 单次 reconcile 最多批次数                               |
| `CATALOG_SYNC_INTERVAL_SEC`  | `600`  | catalog_sync Periodic 间隔（仅 `SUPPORT_SAAS=false`）    |

---

## 9. 与 Ingest / 预算的关系（as-built）

- Ingest **同事务**写 ledger + lot + `budget_consumed` + `combined_key_remain`
- wallet sync 为 post-commit 内联 HTTP（不入队 River job）；**无** `budget_projection`、`dashboard_project`（由看门狗驱动）
- overrun：轻量预判后**按需**入队；百分比告警在 Ingest post-commit 直接调用，不经异步 job
- rebalance：月切 / reconcile / approval / project 删除；**充值不触发**；只重算本地 `combined_key_remain`，不调 NewAPI
- Gateway 读 `combined_key_remain`；看板读 `usage_buckets`

---

## 10. 测试

| 区域                      | 路径                                                                             |
| ------------------------- | ---------------------------------------------------------------------------------- |
| Worker 集成               | `tests/worker/`                                                                    |
| 看门狗 due                | `tests/infra/scheduler/due_test.go`                                                |
| TBS 生命周期              | `tests/domain/company/create_company_test.go`                                      |
| 钱包 override             | `tests/domain/billing/wallet_sync_test.go`                                         |
| 入队 / Unique             | `tests/store/postgres/enqueue_tx_test.go`                                          |
| Ingest 入队               | `tests/domain/usage/ingest_enqueue_test.go`                                         |
| Org ScheduledAt           | `tests/integration/worker/org_sync_test.go`                                        |
| 月切 EnsureMonthRebalance | `tests/domain/budget/schedule/monthly_test.go`                                     |
| 测试辅助                  | `tests/testutil/river/`（`NewRuntime`、`DisablePeriodic`）                         |

单测需 PostgreSQL：`make test-unit`（`-tags=testhook`）。

---

## 11. 代码索引

```text
internal/
  app/compose_worker.go
  adapter/enqueue/*.go
  config/watchdog.go
  domain/usage/ingest.go
  domain/budget/rebalance.go             # 纯本地重算 combined_key_remain
  domain/budget/schedule/monthly.go      # EnsureMonthRebalance
  domain/org/remote/sync.go              # reschedule / cancel / 自愈
  domain/org/remote/schedule.go          # ComputeNextOrgSync
  infra/jobs/                            # catalog.go（kind SSOT）, kinds_*.go, enqueue.go
  infra/scheduler/due.go                 # L2 due 查询
  infra/scheduler/bulk_enqueue.go
  infra/river/client.go
  infra/river/periodic/jobs.go           # BuildPeriodicJobs（watchdog + ingest_reconcile + catalog_sync）
  infra/river/workers/*.go
  store/tenant_background_state.go
  store/postgres/tenant_background_state_repo.go
  store/river_job_repo.go                # HasActiveOrgSync, cancel 查询
```

---

## 12. 关联文档

| 文档                                             | 内容                         |
| ------------------------------------------------ | ---------------------------- |
| [Backend-架构.md](./Backend-架构.md) §7          | 后台运行时                   |
| [Backend-Ingest架构.md](./Backend-Ingest架构.md) | webhook → river job → ingest |
| [Backend-预算.md](./Backend-预算.md)             | 预算域、同事务写入           |
