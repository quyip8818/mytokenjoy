# Backend · 业务时钟与账期

> 现行说明。配置细节见 [Backend-配置架构.md](./Backend-配置架构.md)；预算业务见 [Backend-预算.md](./Backend-预算.md)。  
> outbox / lease / session TTL 属**墙钟**轨（与账期无关）；River job 调度（`scheduled_at`、retry、Unique `ByPeriod`）统一用 Postgres `NOW()`，不用 `cfg.Clock()`。  
> 详见 [实现-离线任务管理.md](./实现-离线任务管理.md)、[Backend-River实现.md](./Backend-River实现.md) §4（Unique `ByPeriod`）。
> 本篇不谈：NewAPI `remain_quota` 算法。

---

## 1. 「哪本预算」是什么意思

这里的「本」**不是**多套预算配置，而是 **按月切开的消耗账本**。

| 跨月保留（同一套） | 按月切开（多本） |
| --- | --- |
| 部门 `budget`、成员 `personal_budget`、Key `quota` 等 **额度上限** | `budget_snapshots` 里按 `period_key`（通常 `YYYY-MM`）累计的 **已消耗** |

每月一张消耗账：6 月的已用记在 `period_key=2026-06`，7 月从 `2026-07` **重新累计**，不用手工清零。预检问的「还能不能花」= 看 **当前这本** 的已用是否触顶。

```mermaid
flowchart TB
  Limit["额度上限（跨月同一套）"]
  subgraph books [消耗账本 · 按 period_key 分本]
    J["2026-06 已用"]
    L["2026-07 已用"]
  end
  Limit -.->|"limit 不变"| books
  Now["业务现在 = 7 月"] --> L
  Gate["预检 / 超支 / 预算树 used"] --> L
```

所以口语里的 **「现在开着哪本预算」** = **当前业务时钟落在哪个月的 `period_key`，门禁与快照读写的是哪一本消耗账**。  
答：只看 `cfg.Clock()`（生产即真实「今天」所在月；本地可用 `CLOCK_ANCHOR` 钉死）。

另一句 **「这笔调用发生在哪个月」** = 这条用量记在审计账本的哪个月，只看 `OccurredAt`，可以和「当前开着的那本」不同（跨月延迟入库时就会不同）。

```mermaid
flowchart LR
  Q1["现在开着哪本消耗账？"] --> CLK["cfg.Clock() → OpenBudgetPeriod"]
  Q2["这笔调用发生在哪月？"] --> EVT["OccurredAt → OccurrencePeriod"]
  CLK --> OPEN["budget_snapshots 当前 period"]
  EVT --> LEDGER["usage_ledger / buckets"]
```

只有 **Ingest** 会同时写这两条轨；其它路径要么只读开账，要么只记发生。

---

## 2. 为什么开账月和发生月要分开

调用可能跨月才入库：6/30 发生，7/1 才 Ingest。此时：

- **发生月**仍是 6 月 → 审计 / 财务要记在 6 月账本  
- **开着的消耗账**已是 7 月 → 门禁必须扣 / 读 7 月快照，否则「本月还能不能用」会看错本

| 若混用 | 后果 |
| --- | --- |
| 用发生月去扣「现在」的消耗账 | 7 月门禁读到 6 月已用，额度错位 |
| 用入库墙钟去写发生月 | 审计按发生月对不上 |

```mermaid
sequenceDiagram
  participant Call as 调用 OccurredAt=6/30
  participant Ingest as Ingest 业务日=7/1
  participant Ledger as usage_ledger
  participant Snap as budget_snapshots
  Call->>Ingest: 延迟入库
  Ingest->>Ledger: period_key = 2026-06（发生月）
  Ingest->>Snap: period_key = 2026-07（当前开着的本）
  Note over Snap: 预检 / 超支只看 2026-07
```

---

## 3. 三种时间（现状）

| 名称 | 来源 | 干什么 | 不干什么 |
| --- | --- | --- | --- |
| **墙钟** | 调度比较用 PG `NOW()`；ID 等可用 `time.Now()` | `river_job` 调度 / retry / Unique 窗口、session TTL、生成 ID | 不算开账 period、不写账本发生月；**不**读 `CLOCK_ANCHOR` |
| **业务时钟** | `cfg.Clock()` | 开账键、预检、超支、预算树 / Key used、看板「今天」、worker 月切触发、seed 开账快照 | 不驱动 lease |
| **事件时间** | `OccurredAt`（来自上游 `CreatedAt`） | ledger `period_key`、`usage_buckets`、审计归因 | 不写开账快照 |

```mermaid
flowchart TB
  subgraph wall [墙钟]
    W[PG NOW]
    W --> L[river_job 调度 / retry]
    ID[time.Now for IDs only]
  end
  subgraph biz [业务时钟]
    C[cfg.Clock]
    C --> O[OpenBudgetPeriod]
    O --> S[budget_snapshots]
    O --> G[precheck / overrun / Load*]
    O --> M[worker 月切]
  end
  subgraph evt [事件时间]
    E[OccurredAt]
    E --> P[OccurrencePeriod]
    P --> U[usage_ledger]
    E --> B[usage_buckets]
  end
```

缺 `CreatedAt` / `OccurredAt`：**直接失败**，禁止回退墙钟。

---

## 4. 账期双轨

两套类型，禁止互换：

| 类型 | 时间源 | 写入 | 读取场景 |
| --- | --- | --- | --- |
| `OpenBudgetPeriod` | Clock | `budget_snapshots.period_key` | 预检、超支、预算树、Key 配额 |
| `OccurrencePeriod` | OccurredAt | `usage_ledger.period_key` | 审计、发生月统计 |

DB 列仍是 `string`；域边界用 `.String()` 进出。

### 4.1 键怎么算

**两层概念不要混：**

| 层 | 存什么 | 谁写 |
| --- | --- | --- |
| `org_nodes.period` | 账期**规格**，现行只允许 `monthly` | 建公司、部门 provision、seed、导入 |
| `budget_snapshots.period_key` / `usage_ledger.period_key` | 具体 **`YYYY-MM` 账本键** | 开账 / 发生工厂 + ingest |

`org_nodes.period` 有 `CHECK (period IN ('monthly'))`；业务路径（`BudgetPeriod()`、seed、开户）一律写 `monthly`。**不能把** `"2026-06"` 这类固定月串落库到 `org_nodes.period`。

开账时如何把 `monthly` 变成 `YYYY-MM`：

- `"monthly"`（或空）→ 用「那个瞬间」的 UTC 月，得到 `YYYY-MM`
- 经 `OpenDepartmentPeriod` / `OpenSnapshotKey` 等工厂，结合 `cfg.Clock()` 或 `OccurredAt`

```mermaid
flowchart TD
  DB["org_nodes.period = monthly"]
  DB --> Factory[Open* / Occurrence* 工厂]
  CLK[Clock 或 OccurredAt] --> Factory
  Factory --> Key["period_key = YYYY-MM"]
```

字符串原语 `SnapshotKey(orgPeriod, at)`（`period_key.go`）在 `orgPeriod` 为固定串（如 `"2026-06"`）时**原样返回**该串——仅供 store/seed 内部或单元测试构造 `period_key` 字符串；**生产域路径不依赖**在 DB 里存固定月。域内开账 / 发生路径 **不直调** `SnapshotKey`，走工厂（见下）。

### 4.2 工厂（唯一入口）

```go
// 开账 ← Clock
OpenDepartmentPeriod(ctx, nodes, departmentID, clk) (OpenBudgetPeriod, error)
OpenSnapshotKey(orgPeriod, clk) OpenBudgetPeriod
RootPeriodKey(nodes, at) string                    // seed：已有 SeedAt

// 发生 ← OccurredAt
OccurrenceDepartmentPeriod(ctx, nodes, departmentID, occurredAt) (OccurrencePeriod, error)
OccurrenceSnapshotKey(orgPeriod, occurredAt) OccurrencePeriod
```

```mermaid
flowchart LR
  subgraph open [开账]
    CLK[Clock] --> OD[OpenDepartmentPeriod]
    CLK --> OS[OpenSnapshotKey]
    SeedAt[SeedAt] --> RK[RootPeriodKey]
  end
  subgraph occ [发生]
    AT[OccurredAt] --> OCD[OccurrenceDepartmentPeriod]
    AT --> OCS[OccurrenceSnapshotKey]
  end
  OD --> Apply[Apply → snapshots]
  OS --> Read[Load* / SumMemberKeyUsed]
  RK --> SeedSnap[seed budget_snapshots]
  OCD --> Led[ledger.PeriodKey]
  OCS --> SeedLed[seed ledger]
```

看板日期区间在 `cost_range.go`（`Resolve` / `PreviousRange`），**不是** snapshot `period_key`。

---

## 5. Ingest：唯一双写点

```text
IngestRaw
  ├─ OccurrenceDepartmentPeriod(OccurredAt) → entry.PeriodKey → ledger
  ├─ OpenDepartmentPeriod(Clock)            → Apply → budget_snapshots
  └─ usage_buckets                          ← OccurredAt
```

```mermaid
flowchart TB
  Raw[RawConsumeLog] --> Build[BuildCallSettledEntry]
  Build -->|CreatedAt 无效| Fail[error]
  Build --> Entry[OccurredAt]
  Entry --> Occ[OccurrenceDepartmentPeriod]
  Clock[cfg.Clock] --> Open[OpenDepartmentPeriod]
  Occ --> Ledger[Insert ledger]
  Open --> Apply[Apply snapshots]
  Entry --> Buckets[usage_buckets]
```

读路径（预检、预算树、Key used、超支）只拿 `Open*` / `Load*(..., Clock)`，不读发生轨来做门禁。

---

## 6. 配置与本地锚点

| 项 | 现状 |
| --- | --- |
| `CLOCK_ANCHOR` | 可选 `YYYY-MM-DD`；空 = 系统时钟；**生产禁止** |
| `cfg.Clock()` | 空锚点 → `System()`；有锚点 → `Fixed(UTC 零点)` |
| 域代码 | 只调 `cfg.Clock()` / 注入的 `clock.Clock`，不读 env |
| demo | 建议 `BOOTSTRAP_MODE=demo` + `CLOCK_ANCHOR`，让种子与门禁同月 |
| `Snapshot.SeedAt` | `clock.NowUTC(cfg.Clock())`；缺则 seed 开账快照 fail-fast |
| seed 开账快照 | `RootPeriodKey(nodes, SeedAt)` |
| seed ledger | `OccurrenceSnapshotKey(PeriodMonthly, OccurredAt)`（可与开账月不同） |

```mermaid
flowchart TD
  Env[CLOCK_ANCHOR]
  Env -->|空| Sys[System 墙钟感业务现在]
  Env -->|YYYY-MM-DD| Fix[Fixed 锚定日 UTC 0:00]
  Env -->|production 非空| Boom[启动失败]
  Sys --> Clock[cfg.Clock]
  Fix --> Clock
  Clock --> Gates[预检 / Load* / 看板 / worker 月切]
  Clock --> SeedAt[Snapshot.SeedAt]
  SeedAt --> Snap[seed snapshots]
```

瞬时取值统一：`clock.NowUTC(clk)`。

---

## 7. 代码放哪

```text
internal/pkg/clock/          Clock 接口、System / Fixed / NowUTC
internal/config/             Clock()、SeedReferenceDate、生产禁锚点

internal/pkg/budget/
  period_key.go              PeriodMonthly、SnapshotKey
  period.go                  Open* / Occurrence*、RootPeriodKey
  cost_range.go              看板 Resolve / PreviousRange
  snapshotload.go            Load* 读消耗

domain/usage/ingest.go       双轨写入
domain/usage/projection.go   Apply(..., OpenBudgetPeriod)
domain/gateway/precheck.go     LoadPrecheckContext + Evaluate（开账 period 在 SQL 内）
domain/budget/overrun.go     开账超支
infra/worker/runner.go       → 删除；月切改 Periodic monthly_rebalance_fanout

seed/snapshot/*.go           SeedAt、ledger OccurrenceSnapshotKey
seed/apply/tables.go         RootPeriodKey → snapshots
scripts/lint-clock.sh        符号护栏
```

```mermaid
flowchart TB
  subgraph pkg [pkg]
    ClockPkg[clock]
    Key[period_key]
    Per[period 工厂]
    Load[snapshotload]
    Cost[cost_range]
  end
  subgraph domain [domain]
    Ingest[usage ingest]
    Pre[gateway precheck]
    Bud[budget overrun / tree]
    Dash[dashboard]
  end
  ClockPkg --> Per
  Key --> Per
  Per --> Ingest
  Per --> Pre
  Per --> Bud
  Load --> Bud
  ClockPkg --> Dash
  Cost --> Dash
```

---

## 8. 护栏（怎么防写漂）

| 手段 | 作用 |
| --- | --- |
| `OpenBudgetPeriod` / `OccurrencePeriod` | 类型上区分开账 vs 发生 |
| Ingest 单写入口 | 快照只经 `Apply(..., OpenBudgetPeriod)` |
| `OccurredAtFromPayload` | 缺事件时间 fail |
| `make lint-clock` | 禁 `SnapshotKey(...time.Now)`；`domain/{budget,gateway,newapisync,usage}` 禁直调 `SnapshotKey` |

钉行为的测试（改时钟语义时先跑这些）：

| 场景 | 测试 |
| --- | --- |
| 账本跟 OccurredAt | `TestIngestStoresLedgerPeriodKey` |
| 跨月：快照跟 Clock | `TestIngestSnapshotUsesNowPeriodForMonthlyOrg` |
| 缺 OccurredAt | `TestOccurredAtFromPayloadRejectsMissing` |
| 锚点预检 | `TestPrecheckUsesClockAnchorForPeriodKey` |
| 树与工厂同月 | `TestOpenBudgetPeriodAlignsTreeAndDepartmentFactory` |
| seed 快照跟 Clock、ledger 跟 OccurredAt | `TestSeedBudgetSnapshotsAlignWithClockAnchor` |
| Load* 开账月跟 Clock | `TestLoadPlatformKeysWithUsedResolvesDepartmentPeriod`、`TestLoadBudgetGroupsWithConsumedUsesOpenPeriod`（用 `clock.Fixed`，勿改 `org_nodes.period`） |
| 生产禁锚点 | `TestProductionRejectsClockAnchor` |

---

## 9. 改代码时三问

1. 这段要的是墙钟、**当前开着哪本消耗账**，还是 **事件发生在哪月**？  
2. 开账 `period_key` 是否只来自 `Open*` / `RootPeriodKey`（seed）？  
3. 有 `CLOCK_ANCHOR` 时：开账路径（seed 快照 / 看板 / 预检 / ingest Apply）是否落在同一本？ledger 是否仍跟 `OccurredAt`？
