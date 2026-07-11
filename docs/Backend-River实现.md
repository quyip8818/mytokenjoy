# Backend · River 实现

> **定位**：River **基础设施**实现说明（**已落地**）；业务现状见 [Backend-离线任务.md](./Backend-离线任务.md)；剩余项见 [实现-离线任务管理.md](./实现-离线任务管理.md)。  
> **替换**：`async_jobs` + `Runner` + `*_processor.go`。

---

## 1. 组件

| 组件 | 路径 | 职责 |
| --- | --- | --- |
| `river.Client` | `internal/infra/river/client.go` | Workers、PeriodicJobs、leader 选举、claim/retry |
| `jobs.Enqueuer` | `internal/infra/jobs/enqueuer.go` | 对 domain 暴露 `Insert` / `InsertInTx` |
| Workers | `internal/infra/river/workers/*.go` | 薄壳；**JobArgs 与 Worker 同文件** |
| Periodic | `internal/infra/river/periodic.go` | cron / interval 注册 |

依赖：`github.com/riverqueue/river`、`riverdriver/riverpgxv5`、`github.com/robfig/cron/v3`（cron 表达式时）。

---

## 2. Schema

### 2.1 删除

`async_jobs` 及索引、`AsyncJobsRepository`、相关 repo 接口。

### 2.2 新增（dump 合入 `schema.sql`）

```bash
# pin go.mod 中 River 版本后
river migrate-get --line main --all --up
# 输出合入 internal/store/postgres/schema.sql
```

| 表 / 类型 | 用途 |
| --- | --- |
| `river_job_state` ENUM | available / completed / discarded / retryable / … |
| `river_job` | 全部 job 行 |
| `river_leader` | Periodic、maintenance 单 leader |
| `river_queue` | 队列暂停、并发（v4+） |
| `river_migration` | River schema 版本 |

**不新增** `river_periodic_job`（Pro）。**保留** `scheduler_locks`（仅 Ingest reconcile）。

### 2.3 管理定案

- **单源**：`schema.sql` + wipe 重建
- **不** runtime `river migrate-up`
- 升 River 版本 → 重新 `migrate-get` 合入

---

## 3. 事务入队（不泄漏 pgx 到 domain）

```go
// jobs.Enqueuer
type Enqueuer interface {
    Insert(ctx context.Context, args JobArgs, opts *InsertOpts) error
    InsertInTx(ctx context.Context, st store.Tx, args JobArgs, opts *InsertOpts) error
}

// store.Tx：postgres 事务句柄窄接口，由 WithTx 传入
type Tx interface {
    // 供 enqueuer 内部取 pgx.Tx 做 river.InsertTx
}
```

```go
// domain 用法
return store.WithTx(ctx, func(st store.Store) error {
    // ... ledger 写入
    return enqueuer.InsertInTx(ctx, st, BudgetProjectArgs{CompanyID: id}, nil)
})
```

`WithTx` 回调签名保持 `func(st Store) error`；**不**在 domain 签名中出现 `pgx.Tx`。

---

## 4. UniqueOpts 映射

业务文档「去重策略」在此落地为 River `UniqueOpts`：

| 业务策略 | River 配置 |
| --- | --- |
| 事件短窗 1s（`budget_project`） | `ByArgs: true`, `ByPeriod: 1s`, **`ByState` 不含 `completed`** |
| 事件短窗 5s（`wallet_sync`） | `ByArgs: true`, `ByPeriod: 5s` |
| 周期 30min / 1h / 24h | `ByArgs: true`, `ByPeriod: 对应时长` |
| per axis / per company | `ByArgs: true` |
| 无 | 不设置 `UniqueOpts` |

**`budget_project` 续跑：** 排除 `completed` 后，批末 `Insert` 不会在 1s 窗口内因已完成 job 被 skip。

调度时间比较用 PG `NOW()`，不用 `cfg.Clock()`（见 [Backend-业务时钟与账期.md](./Backend-业务时钟与账期.md)）。

---

## 5. 队列与 Client 配置

### 5.1 队列（代码注册，非架构写死数字）

```go
// wire_river.go 示例 — 权重为实现细节，可随负载调整
Queues: map[string]river.QueueConfig{
    "critical": {MaxWorkers: N},  // 权重最高
    "default":  {MaxWorkers: N},
    "low":      {MaxWorkers: N},
}
```

业务准入见 [实现-离线任务管理.md](./实现-离线任务管理.md) §3.2。

### 5.2 环境变量（首版最小集）

| 变量 | 默认语义 |
| --- | --- |
| `RIVER_ENABLED` | 是否启动 Client |
| `RIVER_MAX_WORKERS` | 全局 worker 上限 |

Periodic interval、queue 权重 override **放代码常量 + 可选 env**；首版 **不** 暴露 `RIVER_QUEUE_*_WORKERS` 四个变量。归入 `config.RiverConfig`。

---

## 6. 失败恢复（River 机制）

| 机制 | 行为 |
| --- | --- |
| **Retry** | 失败 → `retryable` → 指数退避；`MaxAttempts` 可 per-kind 配置 |
| **Discard** | 超限 → `discarded`；`errors` JSONB 保留末条错误 |
| **Panic** | River 回收 running job，后续重试 |
| **Leader** | `river_leader` 选举；仅一实例跑 Periodic |
| **Pause** | `river_queue` 可暂停某队列（v4+） |

运维：查 `river_job` WHERE `state = 'discarded'`；admin API `Insert` 手动重跑单 tenant。

**幂等：** 业务 handler 必须容忍 at-least-once（见业务文档 §4）。

---

## 7. 可观测（首版 SQL）

| 指标 | 查询要点 |
| --- | --- |
| 队列深度 | `state = 'available'` AND `kind = $1` |
| 排队延迟 | `now() - scheduled_at` |
| 执行耗时 | `finalized_at - attempted_at`（completed） |
| 重试次数 | `attempts` |
| 失败 | `state = 'discarded'`；`errors` 末条 |
| 按 tenant | `args->>'company_id'` 或 `metadata` |

业务投影 lag：`projection_lag_seconds`（异步预算投影 §4.4）。

**后续：** Prometheus、River Pro UI。

---

## 8. Periodic 注册（现状）

| Periodic | 入队 kind | 间隔（默认 env） | 状态 |
| --- | --- | --- | --- |
| org sync | `org_sync` | `WORKER_ORG_SYNC_INTERVAL_SEC`（60s） | **已落地** |
| monthly rebalance | `monthly_rebalance` | `WORKER_POLL_INTERVAL_SEC`（5s） | **已落地** |
| budget / dashboard fanout | `budget_reconcile` 等 | — | **未落地**，见 [实现-离线任务管理.md](./实现-离线任务管理.md) |

fanout 规划：Periodic 入队 1 条 fanout job → domain `Fanout*Jobs` 扫全 tenant 再批量 `Insert` 子 job。

开源 Periodic **caveat**：leader 重启可能漏一次 tick；handler 幂等 + 周期 Unique 兜底。

---

## 9. Worker 文件约定

```text
internal/infra/river/workers/
  budget_project.go    # BudgetProjectArgs + BudgetProjectWorker
  wallet_sync.go
  rebalance.go
  ...
```

每个文件：`Args`（实现 `Kind()`、`InsertOpts()` 如需）+ `Worker.Work()` 调 domain。

---

## 10. 参考

- [River 文档](https://riverqueue.com/docs)
- [Transactional enqueueing](https://riverqueue.com/docs/transactional-enqueueing)
- [Unique jobs](https://riverqueue.com/docs/unique-jobs)
- [Periodic jobs](https://riverqueue.com/docs/periodic-jobs)
