# Token 消费/入账架构评审与优化建议

> 评审范围：Ingest → Fanout → Projection 全链路  
> 日期：2026-07-16

---

## 1. 现有架构概览

### 1.1 数据流全貌

```
                     外部 API 平台 (NewAPI)
                             │
                    consume log 写入 (logs 表)
                             │
                    ┌────────▼────────┐
                    │  EnqueuePending  │  ← webhook / API 触发
                    │  + pg_notify()   │
                    └────────┬────────┘
                             │
               ┌─────────────▼─────────────┐
               │      Ingest Worker        │
               │  LISTEN/NOTIFY + Poll     │
               │  ClaimPendingJobs (SKIP   │
               │    LOCKED batch claim)    │
               │  GroupJobsByCompany →     │
               │    semaphore(8) 并行      │
               └─────────────┬─────────────┘
                             │
            ┌────────────────▼────────────────┐
            │      IngestService.IngestRaw     │
            │  (单公司事务，company row lock)   │
            │                                  │
            │  1. 幂等检查 (idempotency_key)   │
            │  2. FIFO lot 消耗 → ledger 分段  │
            │  3. budget_consumed 增量 UPSERT  │
            │  4. combined_key_remain 递减     │
            │  5. 事务内 enqueue River jobs    │
            └────────────────┬────────────────┘
                             │
         ┌───────────────────┼───────────────────┐
         ▼                   ▼                   ▼
  ┌─────────────┐   ┌──────────────┐   ┌────────────────┐
  │ WalletSync  │   │ Dashboard    │   │ Overrun (条件) │
  │ (default Q) │   │ Project      │   │ (default Q)    │
  │ 5s 去重     │   │ (low Q)      │   │ by args 去重   │
  └──────┬──────┘   │ 1h 去重      │   └────────┬───────┘
         │          └──────┬───────┘            │
         ▼                 ▼                    ▼
  NewAPI quota      Usage bucket 表       预算超限处理
  TopUp/Delta       (hourly 聚合投影)     + 通知
```

### 1.2 技术栈

| 组件 | 技术选型 |
|------|----------|
| 消息队列 | PostgreSQL 表 (ingest_jobs) + LISTEN/NOTIFY |
| 后台任务 | [River](https://github.com/riverqueue/river) (PG-backed Go job queue) |
| 锁 | PG 行锁 (`FOR UPDATE`) + advisory lock |
| 事件驱动 | pg_notify → Go channel → Worker |
| 投影 | Dashboard Projector (cursor-batch) / Budget Consumed (inline UPSERT) |
| 缓存 | Redis (combined_key_remain 副本) |

### 1.3 关键设计决策

1. **Company 行锁序列化**：同一公司的所有 ingest 串行执行，避免并发 lot 消耗/余额竞争
2. **双写 + 补偿**：budget_consumed 在 ingest 事务内增量写入，定期 BudgetReconcile 全量修正漂移
3. **PG-only 队列**：没有引入 Redis/Kafka，所有队列语义用 PG 表 + SKIP LOCKED 实现
4. **幂等设计**：idempotency_key 在锁后检查，保证重复消费安全

---

## 2. 架构优势

| 方面 | 优势 |
|------|------|
| 一致性 | 所有写操作在单一 PG 事务内完成，无分布式事务，强一致 |
| 运维简单 | 无外部中间件依赖（无 Kafka/RabbitMQ/Redis 必要依赖），部署链路短 |
| 幂等保证 | 基于 idempotency_key + 行锁，重放安全 |
| 错误处理 | 分类重试 (IngestOutcome) + 指数退避 + 20次上限 + dead letter |
| 可观测 | metrics refresh、结构化日志、reconcile 漂移告警 |
| 自愈 | Reconcile 层（Budget + Dashboard）可修复 incremental 写入的漂移 |

---

## 3. 瓶颈与风险点

### 3.1 Company 行锁热点

**问题**：高流量公司所有 ingest 串行等锁。`LockForUpdate` 在事务内包含 lot 消耗（可能 N 个 lot 逐一 UPDATE）、budget_consumed batch UPSERT、combined_key 递减、River job insert——锁持有时间长。

**影响**：
- 大客户峰值延迟飙升
- 其他公司不受影响（按 company 隔离），但 worker goroutine 被占用

### 3.2 Ingest 事务过重

单个 ingest 事务执行步骤：
1. `Company.LockForUpdate` — 行锁
2. `Ledger.ExistsIdempotency` — 查询
3. `Billing.ListActiveLotsFIFO` + N × `UpdateLotRemaining` — 多次写
4. `Ledger.InsertSegments` — 写
5. `BudgetConsumed.IncrementConsumedBatch` — UPSERT
6. `CombinedKeySummaries.DecrementBatch` — 条件 UPDATE
7. River `InsertInTx` × 2~3 — 写

最坏情况下单事务包含 **10+ SQL 语句**，latency 高且锁持有时间长。

### 3.3 Dashboard Projection 延迟

- DashboardProject 去重窗口 1 小时，意味着新消费最多 1h 后才投影到 usage bucket
- 实时性依赖直接查 ledger 表 (`MinuteSeriesFromLedger`)，bucket 表只用于聚合查询
- Projector 是 per-company cursor，重启/漂移时需全量重放

### 3.4 Reconcile 可扩展性

- `BudgetReconcile.RunCompany` 加载最近 2 个月全部 ledger 条目（硬限 50000 条）
- `DashboardReconcile` 加载 90 天全部条目
- 随业务增长，reconcile 时间 / 内存将成为问题

### 3.5 Wallet Sync 外部依赖

- 每次 ingest 都会 enqueue WalletSync → 调用 NewAPI TopUp
- 5s 去重窗口在高频消费下仍可能产生大量 TopUp 调用
- `ReconcileWalletDrift` 全量扫描所有公司，O(N) 复杂度

### 3.6 自建 Job Queue 维护成本

- `ingest_jobs` 表是独立于 River 的自建轻量 queue
- 需要自行维护 claim lease、retry backoff、dead letter 语义
- 与 River 功能部分重叠

---

## 4. 优化建议

### 4.1 减轻 Ingest 事务权重（高优先级）

> **注意**：`budget_consumed` 和 `combined_key_remain` 不适合移到 post-commit。  
> 虽然网关侧读取本身有延迟，但 ingest worker 同公司 job 串行执行时，后一笔  
> 依赖前一笔的 budget 写入来判断"已消耗总量"。移出事务会导致同一批次内  
> 多笔 job 之间看不到彼此的消耗，叠加造成超额放行。

#### 方案 A：减少事务内 Lot 消耗 SQL 次数

当前逐 lot `UpdateLotRemaining`（N 个活跃 lot = N 条 UPDATE），改为：
```sql
-- 单条批量 UPDATE，一次扫完 FIFO 队列
WITH consumption AS (
    SELECT id, LEAST(quota_remaining, $remaining) AS take
    FROM recharge_lots
    WHERE company_id = $1 AND status = 'active'
    ORDER BY created_at
)
UPDATE recharge_lots SET quota_remaining = quota_remaining - take ...
```

**收益**：N 条 SQL → 1~2 条，锁持有时间减半  
**代价**：SQL 稍复杂，需处理跨 lot 分段逻辑

#### 方案 B：仅将 Fanout Job Insert 移到 post-commit

安全可移出的：
- `InsertDashboardProject` — 纯投影聚合，延迟无业务影响
- `InsertOverrun` — 通知类，延迟秒级无影响

必须留在事务内的：
- `InsertWalletSync` — 需保证 lot 消耗与 wallet 同步原子性
- `budget_consumed` — 同公司串行一致性需要
- `combined_key_remain` — 同上

**收益**：事务内减少 1~2 条 River INSERT  
**风险**：极低，post-commit 失败时 periodic reconcile 兜底

#### 方案 C：Lot 消耗预扣 + 异步确认

将 wallet_remain 视为 "预扣余额"，先 `DecrementWalletRemain`（原子），然后异步写 lot 明细。

**适用场景**：如果 lot FIFO 逻辑不需要实时精确到每笔 lot 分段。  
**风险**：lot 分段账目短暂不准，需强力 reconcile 补偿。

### 4.2 合并自建 Queue 到 River（中优先级）

**现状**：ingest_jobs + River 并行存在，维护成本双倍。

**建议**：将 ingest pending 迁移为 River job，利用 River 的：
- 内置 SKIP LOCKED claim
- 指数退避 + max attempts
- Dead letter / discard
- 统一监控与 UI (river UI)

**具体做法**：
```go
type IngestArgs struct {
    LogID  int64  `json:"log_id" river:"unique"`
    Source string `json:"source"`
}
func (IngestArgs) Kind() string { return "ingest" }
func (IngestArgs) InsertOpts() river.InsertOpts {
    return river.InsertOpts{
        Queue: config.RiverQueueCritical,
        UniqueOpts: river.UniqueOpts{ByArgs: true},
    }
}
```

**保留 LISTEN/NOTIFY**：River 本身监听 `river_notify` channel，延迟等价。

**开源参考**：
- [River](https://github.com/riverqueue/river) — 已在使用，只需扩展
- 如需 River UI 监控：[riverui](https://github.com/riverqueue/riverui)

### 4.3 Company 级并行优化（中优先级）

#### 方案：细粒度锁 (Platform Key 级)

当前同一公司所有 key 串行。若 lot 消耗可以 **按 platform_key 隔离**（每个 key 有独立余额池），则可以将锁粒度从 company 降到 platform_key。

**评估**：当前 lot 是公司级共享池，因此短期内 company 锁无法避免。但可以：
1. 将 budget_consumed + combined_key_remain 移到事务外（post-commit，已有 reconcile 兜底）
2. 减少事务内 SQL 到绝对最小集（lock + idempotency + lot + ledger）

### 4.4 Dashboard Projection 改为 CDC 或实时增量（低优先级）

#### 方案 A：Ingest 事务内直接写 bucket

在 `IngestRaw` 事务内调用 `upsertDashboardBucket`（类似 budget_consumed 的做法），去掉独立 Projector。

**优点**：消除 1h 投影延迟；去掉 cursor 管理  
**代价**：Ingest 事务再加 1 SQL（但 bucket UPSERT 是小操作）  
**Reconcile 保留**：DashboardReconcile 继续作为安全网

#### 方案 B：PG Logical Replication → 独立消费者

用 PG logical decoding（如 [pglogrepl](https://github.com/jackc/pglogrepl)）监听 ledger 表 INSERT，独立进程实时投影到 bucket。完全解耦 ingest 事务。

### 4.5 Reconcile 增量化（中优先级）

**问题**：全量加载 50000/90天条目。

**建议**：
1. **增量 reconcile**：记录 `last_reconciled_at` 或 version，只加载自上次 reconcile 以来的条目
2. **分片**：按 department / platform_key 分片 reconcile，每次只处理一个分区
3. **抽样检查**：90% 场景无漂移，可先做 checksum 对比，发现差异再全量修

### 4.6 Wallet Sync 合并与限流（低优先级）

**现状**：每次 ingest 都 enqueue WalletSync，5s 去重窗口在高频下仍可能频繁调用外部 API。

**建议**：
1. 将去重窗口从 5s 提高到 **30s~60s**（NewAPI quota 不需要实时同步）
2. 使用 **delta 累积**模式：在本地累积 delta，定期 flush 一次 TopUp
3. `ReconcileWalletDrift` 改为分页 + 只处理 `wallet_remain` 变化的公司

### 4.7 可考虑的开源库/工具

| 需求 | 推荐 | 说明 |
|------|------|------|
| 统一 Job Queue | [River](https://github.com/riverqueue/river) | 已在用，建议统一 ingest queue |
| Job Queue 监控 | [River UI](https://github.com/riverqueue/riverui) | Web UI 查看 job 状态 |
| PG CDC | [pglogrepl](https://github.com/jackc/pglogrepl) | 替代 LISTEN/NOTIFY，支持 WAL-level 变更捕获 |
| 分布式锁 | [pglock](https://github.com/cirello-io/pglock) | 如需更细粒度的 advisory lock 管理 |
| 限流 | [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate) | 外部 API 调用限流（WalletSync） |
| 指标 | [river-prometheus](https://github.com/riverqueue/river/tree/main/rivertype) | River 原生指标导出 |
| Outbox Pattern | 内置 River InsertInTx | 已在使用的模式，无需额外库 |

---

## 5. 优先级矩阵

| 优化项 | 收益 | 实施难度 | 建议优先级 |
|--------|------|----------|------------|
| 减轻 Ingest 事务 (budget/combined 移到 post-commit) | 高 | 低 | P0 |
| 合并 ingest queue 到 River | 中 | 中 | P1 |
| Wallet Sync 去重窗口扩大 | 中 | 低 | P1 |
| Dashboard inline 写入 | 中 | 低 | P1 |
| Reconcile 增量化 | 中 | 中 | P2 |
| Company 锁优化 (评估 per-key) | 高 | 高 | P2 (需设计) |
| PG CDC 替代 Projector | 低 | 高 | P3 |

---

## 6. P0 优化详细方案：Lot 消耗批量化

### 当前事务边界

```go
s.store.WithTx(ctx, func(st store.Store) error {
    // 1. LockForUpdate              ← 必须，1 SQL
    // 2. ExistsIdempotency          ← 必须，1 SQL
    // 3. ConsumeLotsLocked          ← 必须，但当前 N+2 SQL (ListFIFO + N×Update + SetWalletRemain)
    // 4. InsertSegments             ← 必须，1 SQL
    // 5. IncrementConsumedBatch     ← 必须（同公司串行一致性），1 SQL
    // 6. DecrementBatch             ← 必须（同上），1 SQL
    // 7. EnqueueAfterIngest         ← 部分可移出，2~3 SQL
})
```

### 优化目标：将步骤 3 从 N+2 SQL 压缩到 2~3 SQL

```sql
-- 1. 单条 CTE 完成 FIFO 消耗 + lot 状态更新
WITH ordered_lots AS (
    SELECT id, quota_remaining,
           SUM(quota_remaining) OVER (ORDER BY created_at) AS cumulative
    FROM recharge_lots
    WHERE company_id = $1 AND status = 'active'
      AND (fifo_head IS NULL OR id >= fifo_head)
),
consumption AS (
    SELECT id,
           LEAST(quota_remaining, GREATEST($2 - (cumulative - quota_remaining), 0)) AS take
    FROM ordered_lots
    WHERE cumulative - quota_remaining < $2
)
UPDATE recharge_lots r
SET quota_remaining = r.quota_remaining - c.take,
    status = CASE WHEN r.quota_remaining - c.take <= 0 THEN 'exhausted' ELSE r.status END,
    updated_at = NOW()
FROM consumption c
WHERE r.id = c.id
RETURNING r.id, c.take, r.billing_currency, r.unit_price_display;

-- 2. 如有 remaining > 0 → ExpandOverdraftLot (1 SQL)
-- 3. SetWalletRemain (1 SQL)
```

**收益**：N 个活跃 lot 的情况下，从 N+2 SQL 降至 2~3 SQL，显著减少锁持有时间  
**代价**：SQL 复杂度增加，需充分测试边界情况（单 lot 不足、全部用完进 overdraft）

### 同时：将 DashboardProject / Overrun 移到 post-commit

```go
s.store.WithTx(ctx, func(st store.Store) error {
    // 1. LockForUpdate
    // 2. ExistsIdempotency
    // 3. ConsumeLotsLocked (批量化 CTE)
    // 4. InsertSegments
    // 5. IncrementConsumedBatch    ← 保留事务内
    // 6. DecrementBatch            ← 保留事务内
    // 7. InsertWalletSync          ← 保留事务内
})

// Post-commit（安全，有 reconcile 兜底）
enqueuer.InsertDashboardProject(ctx, companyID)  // 纯投影
enqueuer.InsertOverrun(ctx, companyID, payload)  // 通知类
```

**总收益**：事务内 SQL 从 ~10 条降至 ~7 条，其中最重的 lot 消耗从 O(N) 降至 O(1)

---

## 7. 总结

当前架构 **设计合理、一致性强、运维简单**，是 PG-native 事件处理的优秀实践。主要优化方向不是"换技术栈"，而是：

1. **缩短事务持有锁的时间**（将非关键写操作移到 post-commit）
2. **统一队列基础设施**（消除 ingest_jobs 表的维护负担）
3. **提升投影实时性**（Dashboard inline 写入）
4. **降低 Reconcile 成本**（增量化）

这些优化在保持现有 PG-native 架构优势的前提下，可显著改善大客户高频消费场景的 latency 和系统可扩展性。
