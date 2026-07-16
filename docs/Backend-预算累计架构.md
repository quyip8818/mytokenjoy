# Backend · 预算累计与 Gateway 预检架构

> 预算累计（`budget_consumed`）和 Gateway 预检余额（`combined_key_remain`）随 Ingest 同事务原子写入。
> 本文描述当前实现状态。

相关：[Backend-预算.md](./Backend-预算.md) · [Backend-离线任务.md](./Backend-离线任务.md) · [Notification.md](./Notification.md)

---

## 1. 架构总览

```mermaid
flowchart TB
  subgraph webhook [外部]
    NW[NewAPI Webhook]
  end

  subgraph ingest_tx [Ingest 事务 — companies FOR UPDATE]
    direction TB
    LOCK[LockCompanyForUpdate]
    IDEM[ExistsIdempotency?]
    LOT[ConsumeLotsLocked]
    LED[InsertSegments — usage_ledger]
    BC[IncrementConsumedBatch — budget_consumed]
    CK[DecrementBatch — combined_key_remain]
    JOB[EnqueueAfterIngest]
    LOCK --> IDEM
    IDEM -->|new| LOT --> LED --> BC --> CK --> JOB
    IDEM -->|dup| RET[return nil]
  end

  subgraph postcommit [提交后 best-effort]
    REDIS[RefreshCombinedKeySummaries]
    ALERT[CheckBudgetAlerts]
    NOTIFY[notification.DispatchAsync]
    ALERT --> NOTIFY
  end

  subgraph jobs [River Jobs]
    DASH[dashboard_project]
    WALL[wallet_sync]
    OVR[overrun — 条件入队]
    NOTI[notification_delivery]
    NOTIFY --> NOTI
  end

  subgraph cold [冷路径 ~24h]
    RECON[budget_reconcile]
    REBAL[rebalance — 修复后]
    RECON --> REBAL
  end

  NW --> ingest_tx
  ingest_tx -->|commit| postcommit
  JOB --> DASH
  JOB --> WALL
  JOB -->|remain ≤ 0| OVR
  RECON --> BC
  RECON --> CK
```

---

## 2. Ingest 事务

### 2.1 时序

```mermaid
sequenceDiagram
    participant I as IngestService
    participant PG as PostgreSQL
    participant R as River (in-tx)

    I->>PG: BEGIN
    I->>PG: SELECT ... FROM companies WHERE id=$1 FOR UPDATE
    I->>PG: SELECT EXISTS(idempotency_key)
    alt duplicate
        I->>PG: COMMIT (no-op)
    else new
        I->>PG: ConsumeLots (lot FIFO + wallet_remain)
        I->>PG: INSERT usage_ledger (segments)
        I->>PG: IncrementConsumedBatch (UNNEST batch UPSERT)
        I->>PG: DecrementBatch (combined_key_remain -= amount)
        alt key 未返回 (NULL remain)
            I->>PG: SELECT platform_keys FOR UPDATE
            I->>PG: ComputeGatewaySummaryUpdates
            I->>PG: UpdateBatch (absolute remain)
        end
        I->>R: INSERT wallet_sync job
        opt remain ≤ 0 或 Unknown
            I->>R: INSERT overrun job
        end
        I->>PG: COMMIT
    end
```

### 2.2 约束

| 约束 | 实现 |
|---|---|
| 公司级串行 | `companies FOR UPDATE` — Ingest 和 reconcile 共用 |
| 幂等在锁后 | 锁后检查 → 重复请求零副作用 |
| consumed 一次写 | `IncrementConsumedBatch` UNNEST 批量 UPSERT，最多 3 轴 |
| combined 原子扣减 | `GREATEST(remain - delta, 0)` |
| 绝对重算仅初始化 | NULL remain → 锁行 → 重算 → UpdateBatch |
| job 失败 = 账务回滚 | River job 在同一事务中插入 |
| 无 advisory lock | Ingest 热路径不拿 budget advisory lock |

### 2.3 combined_key_remain 三态

```mermaid
stateDiagram-v2
    [*] --> DecrementBatch
    DecrementBatch --> Known : 返回了 key
    DecrementBatch --> MaybeNull : key 未返回
    MaybeNull --> absoluteRecompute : 锁行重算
    absoluteRecompute --> Known : 有预算约束
    absoluteRecompute --> Unconstrained : 无约束保持NULL
    absoluteRecompute --> Unknown : 计算失败

    Known --> OverrunGate : remain ≤ 0 → enqueue
    Known --> NoOp : remain > 0
    Unconstrained --> NoOp : 不发 overrun
    Unknown --> OverrunGate : 安全发一次 overrun
```

### 2.4 批量 consumed SQL

```sql
INSERT INTO budget_consumed (company_id, axis_kind, axis_id, period_key, consumed, updated_at)
SELECT $1, axis_kind, axis_id, period_key, amount, NOW()
FROM UNNEST($2::text[], $3::text[], $4::text[], $5::numeric[])
    AS input(axis_kind, axis_id, period_key, amount)
ON CONFLICT (company_id, axis_kind, axis_id, period_key)
DO UPDATE SET consumed = budget_consumed.consumed + EXCLUDED.consumed, updated_at = NOW();
```

---

## 3. Overrun Gate

```mermaid
flowchart TD
    START[DecrementBatch 后]
    START --> Q1{platformKeyID 为空?}
    Q1 -->|是| SKIP[跳过]
    Q1 -->|否| Q2{summaries == nil?}
    Q2 -->|是 Unconstrained| SKIP
    Q2 -->|否| Q3{key 在 summaries 中?}
    Q3 -->|否 Unknown| ENQUEUE[InsertOverrun]
    Q3 -->|是| Q4{remain ≤ 0?}
    Q4 -->|是| ENQUEUE
    Q4 -->|否| SKIP
```

- Ingest 只做 gate，不执行 Disable/NewAPI
- `OverrunService` worker 做多轴裁决（platform key → member → project → department）
- payload 含 `periodKey`，避免跨月误判

---

## 4. 告警：notification server 集成

```mermaid
flowchart LR
    subgraph ingest [Ingest post-commit]
        CA[CheckBudgetAlerts]
    end
    subgraph resolve [收件人解析]
        RULES[AlertRules by dept]
        ROLES[NotifyRoleIDs → role name]
        MEMBERS[active members by role name]
    end
    subgraph notify [notification server]
        DA[DispatchAsync per member]
        RIVER[River notification_delivery]
        CH[in_app / email / SMS / webhook]
    end

    CA --> RULES --> ROLES --> MEMBERS
    MEMBERS --> DA --> RIVER --> CH
```

| 要素 | 说明 |
|---|---|
| 触发 | Ingest commit 后，仅 touched department |
| 收件人 | `NotifyRoleIDs` → role ID→name → active members |
| 去重键 | `budget-alert:{companyID}:{ruleID}:{threshold}:{periodKey}:{memberID}` |
| 偏好 | notification server 处理（quiet hours / channel / rate limit） |
| 失败 | 只记日志，不影响已提交账务 |

---

## 5. Reconcile（冷路径）

```mermaid
sequenceDiagram
    participant W as ReconcileWorker
    participant PG as PostgreSQL

    W->>PG: BEGIN
    W->>PG: AcquireBudgetLock (advisory)
    W->>PG: LockCompanyForUpdate
    W->>PG: ListCallSettledSince (~2 月窗口)
    Note over W: 按 entry.OccurredAt 归属开账月
    W->>PG: ListConsumedByPeriods (actual)
    Note over W: diff expected vs actual
    alt drift / 缺行
        W->>PG: SetConsumed(expected)
    end
    alt 多余行
        W->>PG: SetConsumed(0)
    end
    opt 有修复
        W->>PG: LockPlatformKeysForUpdate (sorted)
        W->>PG: ComputeGatewaySummaryUpdates → UpdateBatch
    end
    W->>PG: COMMIT
    opt 修复后
        W->>PG: InsertRebalance(company)
    end
```

| 约束 | 说明 |
|---|---|
| 并发安全 | advisory 锁 + company 行锁，Ingest 只拿 company 锁 → 无死锁 |
| 账期归属 | `OpenDepartmentPeriodAt(entry.OccurredAt)` 而非当前 Clock |
| 多余行清零 | actual 中不在 expected 中的行 → SetConsumed(0) |
| 硬限 | 单公司 ledger 上限 50000 条，超过 job 失败重试 |
| 频率 | ~24h，scheduler 判定 due 后 enqueue |

---

## 6. 数据写入者

| 数据 | 写入者 | 时机 |
|---|---|---|
| `usage_ledger` | Ingest | 事务内 |
| `budget_consumed` | Ingest (IncrementConsumedBatch) | 事务内 |
| `budget_consumed` | Reconcile (SetConsumed) | 冷路径修复 |
| `combined_key_remain` | Ingest (DecrementBatch) | 事务内 |
| `combined_key_remain` | Ingest (absoluteRecompute) | 事务内 — 仅 NULL 初始化 |
| `combined_key_remain` | Reconcile (UpdateBatch) | 冷路径修复 |
| `combined_key_remain` | Rebalance | 充值/月切后 |
| `notification_log` | notification server | 异步投递 |
| `usage_buckets` | dashboard projector | 异步投影（看门狗每小时触发） |

---

## 7. 代码结构

```text
domain/usage/
├── ingest.go                        # IngestRaw 主路径
├── ports.go                         # IngestJobEnqueuer
└── ingest_overrun_gate_test.go      # ShouldEnqueueOverrun 单测

domain/budget/
├── alert_publisher.go               # AlertPublisher port + CheckBudgetAlerts
├── alert_publisher_test.go          # role 解析单测
├── consumed_attrib.go               # ConsumptionDeltas / ConsumedDrift
├── consumed_attrib_test.go          # delta 计算单测
├── budget_reconcile.go              # ReconcileService.RunCompany
├── budget_reconcile_test.go         # reconcile 工具函数单测
├── combined_key_summary.go          # ComputeGatewaySummaryUpdates
├── overrun.go                       # OverrunService (worker 多轴裁决)
├── rebalance.go                     # RebalanceService
└── ports.go                         # JobEnqueuer (overrun/rebalance/reconcile)

domain/billing/lot/
└── consume.go                       # ConsumeLots / ConsumeLotsLocked

app/
├── port_usage.go                    # EnqueueAfterIngest
├── port_budget.go                   # budget JobEnqueuer adapter
├── port_budget_alert.go             # AlertPublisher → notification.DispatchAsync
└── compose_domain_wire.go           # wireIngestService

infra/
├── jobs/kinds_budget.go             # Overrun / Rebalance / Reconcile args
├── river/client.go                  # Worker 注册（无 BudgetProjector）
├── river/workers/budget_reconcile.go
├── scheduler/due.go                 # NeedsBudgetReconcile
└── notification/                    # DispatchAsync → notification_delivery

store/
├── budget_consumed_repo.go          # ConsumedDelta + IncrementConsumedBatch
├── combined_key_summary.go          # LockPlatformKeysForUpdate
└── postgres/                        # SQL 实现
```

---

## 8. 不存在的组件

以下组件已删除，不在当前代码中：

- `budget_projector.go` / `budget_projector_alerts.go` / `async.go`
- `BudgetProjectionWorker` / `BudgetProjectionArgs` / `KindBudgetProjection`
- `InsertBudgetProjection`
- `budget_projection_progress` 表和 repository
- `NeedsBudgetProject` / projection lag 检查
- `ApplyIncrement` / `ConsumedIncrementWriter` / `ExpectedConsumed`（旧版）
