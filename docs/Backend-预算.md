# Backend 预算与消耗

企业钱包与组织预算双轴、入账、配额同步与超限的**当前实现**说明。

**相关：** [Backend.md](./Backend.md) · [Backend-架构.md](./Backend-架构.md) · [Backend-存储架构.md](./Backend-存储架构.md) · [Backend-计费模式.md](./Backend-计费模式.md)

---

## 阅读路径

| 章节 | 适合谁 | 内容 |
| --- | --- | --- |
| §1–2 | 产品 / 新同学 | 预算管什么、两条轴 |
| §3–4 | 前端 / 实施 | 分配规则、管理面 API |
| §5 | 全栈联调 | 一次调用全链路 |
| §6–7 | 后端开发 | 存储与入账 |
| §8–10 | 后端 / 运维 | Rebalance、超限、充值 |
| §11–12 | 查表 | 公式、负责域 |
| §13 | 排期 | 待优化与待修复 |

计量单位：内部统一 **point**；钱包展示币由 lot 成本价闭合；NewAPI `remain_quota` 为派生通道配额。详见 [Backend-计费模式.md](./Backend-计费模式.md)。

---

## 1. 产品视角

管理员配置额度，成员在额度内用 Platform Key 调模型：

```mermaid
flowchart LR
  subgraph admin [管理员]
    A1[充值 → 涨钱包]
    A2[组织树分配部门预算 / 预留池]
    A3[成员额度 / 预算组 / Key 配额]
  end

  subgraph runtime [运行时]
    R1[Gateway 预检]
    R2[入账记消耗]
    R3[Rebalance + Overrun]
  end

  admin --> R1 --> R2 --> R3
```

| 角色 | 关心什么 |
| --- | --- |
| 企业超管 | 钱包余额、根部门总预算、下级分配 |
| 部门 TL | 本部门预算、预留池、成员额度、预算组 |
| 普通成员 | 个人额度、Key 配额、能否继续调用 |
| 审计 / 财务 | 调用花费、归因部门 / 成员 |

**账期：** 分配配置（`budget`、`personal_budget`、Key `budget`）跨月保留；已消耗按开账 `period_key`（通常 `YYYY-MM`，来自业务时钟）写入 `budget_snapshots`，新月自动从新账期累计。账本发生月见 [Backend-业务时钟与账期.md](./Backend-业务时钟与账期.md)。

---

## 2. 两条轴

| 轴 | 权威数据 | 管什么 | 谁改 |
| --- | --- | --- | --- |
| **企业钱包** | 充值 lot、`balance_point` | 预付资金硬上限（point） | 充值确认 → 异步同步 NewAPI |
| **组织预算** | 组织树 limit + `budget_snapshots` | 部门 / 成员 / Key / 组的花费与上限 | 控制台写 limit；入账累加 consumed |

```mermaid
flowchart TB
  subgraph wallet [钱包轴]
    LOT[充值 lot] --> BAL[balance_point]
    BAL -.-> NA_W[NewAPI users.quota]
  end

  subgraph org [组织预算轴]
    TREE[组织树 budget / reserved_pool]
    TREE --> MEM[personal_budget]
    TREE --> BG[预算组 budget]
    MEM & BG --> PK[Key budget]
    SNAP[(budget_snapshots)]
    ING[入账] --> SNAP
  end

  BAL -->|Rebalance 封顶| PK
  SNAP -->|已用| PK
```

**约定：**

- 充值**只涨钱包**，不自动涨部门 `budget`。
- **limit** 在组织树、成员、Key、预算组；**consumed** 只在 `budget_snapshots`（四轴 × 账期）。
- API 返回的 `used` / `consumed` 为当前账期从快照合并的视图，不是 Key 表上的持久列。

---

## 2. SSOT 与投影

| 层       | 存储                                   | 写入                    | 读方                                  |
| -------- | -------------------------------------- | ----------------------- | ------------------------------------- |
| **事实** | `usage_ledger`                         | `usage.IngestService`   | 审计 `/audit/calls`、minute 看板      |
| **投影** | `budget_snapshots` / `usage_buckets`   | `usage.Apply`（同事务） | 超限/Rebalance、hour/day 看板、预算树 |

> **术语：** `used` 与 `consumed` 同义，统一读法见 [Backend-存储架构.md](./Backend-存储架构.md) §8。组织轴 consumed SSOT 为 `budget_snapshots`（按 `axis_kind` 区分 `org_node`、`budget_group`、`platform_key`、`member`），各业务表上不存在 consumed/used 列。

```mermaid
flowchart LR
  WH[Webhook] -->|EnqueuePending| Q[(ingest_jobs)]
  Q --> ING[IngestService]
  COMP[补偿轮询] --> ING
  ING --> UL[(usage_ledger)]
  ING --> PROJ[projection.Apply]
  PROJ --> SNAP[budget_snapshots.consumed]
  PROJ --> UB[(usage_buckets)]
  ING --> JOB[river_job budget_project / wallet_sync]
```

### 2.1 入账路径（方案 B）

1. NewAPI settle → 写共享 `logs` 库 → `EnqueueNotify(log_id)` → `POST /api/internal/webhooks/newapi-log` → **入队 pending 并立即 ACK**
2. Worker：`ingest_jobs` 消费入账（含 webhook 快路径与失败重试）+ `reconcile_cursors` 全局水位补洞（均走 `IngestByLogID`）
3. `FindMappingByNewAPIKeyID` → `company_id`、部门/成员/组归因
4. `BuildCallSettledEntry` → `idempotency_key = newapi:{log_id}`
5. `store.WithTx`：ledger `INSERT ON CONFLICT` → projection → 副作用入队

### 2.2 `projection.Apply` 顺序

消耗追踪统一集中在 `budget_snapshots` 表，通过 `axis_kind` 区分不同维度（`org_node`、`budget_group`、`platform_key`、`member`）。各业务表（`platform_keys`、`budget_groups`、`org_nodes`）上**没有** `used` / `consumed` 列。

| 步骤 | 写入                                                        | 说明                        |
| ---- | ----------------------------------------------------------- | --------------------------- |
| 1    | `budget_snapshots (axis_kind=platform_key)` += cost         | Key 已用                    |
| 2    | `budget_snapshots (axis_kind=budget_group)` += cost         | 若挂组                      |
| 3    | `budget_snapshots (axis_kind=org_node)` 祖先 rollup += cost | 以 `department_id` 为叶子   |
| 4    | `usage_buckets` Upsert                                      | 小时桶 × 部门 × 成员 × 模型 |

父节点 consumed 含整棵子树花费（rollup），均从 `budget_snapshots` 读取。

### 2.3 读路径分离

| 场景                      | 读什么                                              | 为何                                          |
| ------------------------- | --------------------------------------------------- | --------------------------------------------- |
| 看板 cost / consumed 趋势 | `usage_buckets` SUM                                 | 与 Ingest 投影一致；不读 budget_snapshots     |
| 预算树展示 `consumed`     | `budget_snapshots (axis_kind=org_node)`             | 控制台实时投影                                |
| 审计调用列表              | `usage_ledger`                                      | SSOT；不查 NewAPI logs                        |
| minute 趋势               | `usage_ledger` 按分钟聚合                           | 窗口 ≤3h                                      |

---

## 3. 分配层级

```mermaid
flowchart TB
  ROOT[根节点<br/>未分配 = budget − reserved − Σ子部门]
  ROOT --> DEPT[子部门]
  DEPT --> CAP[成员可分配 capacity]
  CAP --> M[成员 personal_budget]
  M --> K[Key budget]
  DEPT --> POOL[预留池]
  DEPT --> BG[预算组 + 组内 Key]
```

| 层级 | 配置 | 说明 |
| --- | --- | --- |
| 部门 | `budget`、`reserved_pool` | 子节点之和 + 预留池 ≤ 父节点 |
| 成员 | `personal_budget` | 部门内成员额度之和 ≤ capacity |
| Key | `budget`、模型白名单 | 从成员或预算组剩余额度切分 |
| 预算组 | `budget` | 挂组 Key 走组额度；Overrun 不走成员个人分支 |

**写入校验：**

| 操作 | 规则 |
| --- | --- |
| 改部门预算 | 子级：新 budget ≥ Σ子节点 + 预留池；对父级：新 budget + 兄弟 + 预留池 ≤ 父可用 |
| 改成员额度 | ≥ 已分配给 Key 的配额之和；部门内总和 ≤ capacity |
| 建 Key（成员） | budget ≤ 成员剩余可分配 |
| 建 Key（预算组） | budget ≤ 组 budget − 组 consumed − 组内已分配 Key budget |
| 额度追加审批 | 申请额 ≤ 部门 `reserved_pool`；通过后增加 `personal_budget` |

组织树结构变更与模型白名单同事务提交；预算数字仅经预算域服务修改。

---

## 4. 管理面 API

契约详情见 [Frontend.md](./Frontend.md) §5。

| 能力 | 方法 | 路径 |
| --- | --- | --- |
| 预算树 | GET | `/api/budget/tree` |
| 部门预算 | PUT | `/api/budget/departments/{departmentId}` |
| 成员额度 | GET / PUT | `/api/budget/members/{memberId}` |
| 预算组 | CRUD | `/api/budget/groups/*` |
| 预警规则 | CRUD | `/api/budget/alerts/*` |
| 超限策略 | GET / PUT | `/api/budget/overrun-policy` |
| 预算审批 | GET / PUT | `/api/budget/approvals`、`/api/budget/approvals/{id}` |
| Key / 额度审批 | — | `/api/keys/approvals/*`（密钥域，独立表） |
| 充值 | POST | `/api/billing/recharge`；平台代充见 [Backend.md](./Backend.md) §2 |

---

## 5. 一次调用全链路

生产须开 **Gateway**（`NEW_API_GATEWAY_ENABLED=true`）。

```mermaid
sequenceDiagram
  participant C as 调用方
  participant GW as Gateway
  participant NA as NewAPI
  participant Q as ingest_jobs
  participant W as Worker
  participant ING as 入账

  C->>GW: sk- + 请求
  GW->>GW: LoadPrecheckContext + Evaluate
  alt 失败
    GW-->>C: 4xx
  end
  GW->>NA: 透传 /v1/*
  NA-->>C: 响应
  NA->>Q: Webhook 入队 pending
  NA-->>NA: 200 accepted
  W->>Q: ClaimPending
  W->>ING: IngestByLogID
  ING->>ING: 账本 + 快照 + 扣 lot
  ING->>W: rebalance / overrun 入队
  W->>NA: UpdateToken
  W->>W: 超限则 Disable Key
```

**Gateway 预检（同步）** — 全部通过才代理（单位 point）；1× `LoadPrecheckContext` + 纯内存 `Evaluate`：

| 检查 | 数据 |
| --- | --- |
| 企业 active | `companies.status` |
| 钱包 ≥ 预估 | `balance_point` |
| 部门 / Key / 成员 / 组未超 | 四轴 snapshots + limit（SQL 一次带出） |
| 模型与 Key 状态 | allowlist、`platform_keys.status` |

NewAPI quota 与 `wallet_sync` **不参与**热路径预检（Phase 1）；执行面一致性见 [架构简化方案.md](./架构简化方案.md) Phase 3。

---

## 6. 数据层

```mermaid
flowchart TB
  UL[(usage_ledger 事实)]
  BS[(budget_snapshots 四轴 consumed)]
  UB[(usage_buckets 看板)]
  CFG[配置表 limit]

  ING[入账] --> UL & BS & UB
  CFG --> GW[预检]
  BS --> GW & UI[预算树 / Key 列表]
  UL --> AUD[审计]
  UB --> DASH[看板]
```

| 存储 | 职责 |
| --- | --- |
| `usage_ledger` | 消耗 SSOT；幂等 `newapi:{log_id}` |
| `budget_snapshots` | 四轴 `org_node` · `budget_group` · `platform_key` · `member` |
| `usage_buckets` | 按小时聚合，供趋势图 |
| 组织树 / 成员 / Key / 组 | 仅存 limit |

| 读场景 | 数据源 |
| --- | --- |
| 预算树、Key 已用 | `budget_snapshots` 合并 |
| 看板趋势 | `usage_buckets` |
| 调用审计 | `usage_ledger` |
| 分钟级短趋势 | `usage_ledger` 聚合 |

部门 consumed 含祖先 rollup。`used` 与 `consumed` 同义。表结构见 [Backend-存储架构.md](./Backend-存储架构.md) §5–§8。

---

## 7. 入账与投影

```mermaid
flowchart LR
  NA[NewAPI 结算] --> LOGS[(logs 库)]
  LOGS --> WH[Webhook / reconcile 补偿]
  WH --> ING[Ingest]
  ING --> TX[单事务]
  TX --> L[usage_ledger]
  TX --> F[FIFO 扣 lot]
  TX --> P[投影 snapshots + buckets]
  TX --> Q[river_job InsertTx]
```

1. 结算日志 → Webhook 或 Worker 补洞 → 按 `newapi_key_id` 归因
2. 单事务：账本幂等插入 → 投影 → 扣 lot → 入队 rebalance / overrun
3. 失败走 `ingest_jobs` 重试（与 newapi_sync outbox 分离）

**投影顺序：**

快照轴用**开账** `OpenBudgetPeriod`（业务时钟）；`usage_buckets` 与 ledger 归因用 **OccurredAt**。详见 [Backend-业务时钟与账期.md](./Backend-业务时钟与账期.md)。

| 顺序 | 轴 | 说明 |
| --- | --- | --- |
| 1 | platform_key | Key 消耗（开账 period） |
| 2 | budget_group | 若挂组 |
| 3 | member | 若可归因成员 |
| 4 | org_node | 叶子部门向上 rollup |
| 5 | — | usage_buckets 小时桶（OccurredAt） |

---

## 8. Rebalance

在入账、充值、Key 变更后，将组织侧「还能花多少」换算为 NewAPI `remain_quota` 并 `UpdateToken`。

```mermaid
flowchart LR
  EVT[触发事件] --> RJ[river_job rebalance]
  AJ --> RB[RebalanceService]
  RB --> MIN[多候选取 min]
  MIN --> NA[UpdateToken]
```

| `axis_kind` | 触发 |
| --- | --- |
| member | 入账带成员 |
| department | 每次入账 |
| budget_group | 入账命中组 |
| company | 充值完成 |

去重：`dedupe_key = axis_kind:axis_id`。

**候选最小值（point）：** Key 剩余；成员剩余（非挂组）；组剩余（挂组）；部门 budget − consumed − reserved_pool。再换 NewAPI 单位，并以 `balance_point` 作企业硬顶。

---

## 9. 超限与预警

**双层封禁：**

| 时机 | 机制 |
| --- | --- |
| 请求前 | Precheck：consumed + 预估 > limit → 4xx |
| 入账后 | Overrun Worker：consumed ≥ limit → Disable Key |

```mermaid
flowchart LR
  PRE[Precheck] -->|通过| CALL[NewAPI]
  CALL --> ING[入账]
  ING --> OV[Overrun]
  OV -->|超限| OFF[Disable Key]
```

**Overrun 条件（当前账期快照，硬比较 ≥）：**

| 范围 | 条件 | 动作 |
| --- | --- | --- |
| 成员 | 未挂组 Key，member 轴 consumed ≥ personal_budget | 禁用该成员非组 Key |
| 部门 | org_node 轴 consumed ≥ budget | 禁用部门下全部 Key |
| 预算组 | group 轴 consumed ≥ budget | 禁用组内 Key |

**预警配置：** `alert_rules`、`overrun_policy` 可经 API 配置并持久化；超限通知经 `NOTIFY_WEBHOOK_URL` 出站（如 `overrun_blocked`）。

---

## 10. 充值

```mermaid
sequenceDiagram
  participant U as 操作者
  participant B as billing
  participant DB as Postgres
  participant NA as NewAPI

  U->>B: 创建并确认订单
  B->>DB: lot + balance_point
  B->>DB: wallet_sync 入队
  B->>NA: TopUp / SetUserQuota
  B->>DB: company 轴 rebalance 入队
```

充值不改 `org_nodes.budget`；钱包闭合见 [Backend-计费模式.md](./Backend-计费模式.md)。

---

## 11. 公式速查

| 名称 | 计算 |
| --- | --- |
| 部门可分给成员 | budget − reserved_pool − Σ子部门 budget |
| 成员可分给 Key | personal_budget − Σ已分配 Key budget |
| 成员本账期已用 | snapshots member 轴 |
| 组可分给 Key | 组 budget − 组 consumed − Σ组内 Key budget |
| NewAPIKey 可用上限 | 上列候选取 min → 换 NewAPI 单位 |
| 企业硬顶 | Σ NewAPIKey remain ≤ balance_point 对应通道配额 |

---

## 12. 负责域

| 职责 | 域 |
| --- | --- |
| 预算树、组、成员额度、预警策略 | `domain/budget` |
| 入账与投影 | `domain/usage` |
| Rebalance | `domain/budget/rebalance`（`adminport.Port` 更新 token） |
| NewAPI Admin 边界 | `domain/adminport` + `integration/newapi/admin_port_adapter.go` |
| Quota 换算 | `pkg/newapiunits` |
| Key 额度校验 | `domain/keys` + `pkg/budget` |
| 快照加载 | `pkg/budget` |
| Gateway 预检 | `domain/gateway` |
| 充值 | `domain/billing` |
| 异步任务 | `river_job`（River；见 [实现-离线任务管理.md](./实现-离线任务管理.md)） |

---

## 13. 待优化与待修复

按优先级归纳；工程细节另见 [plan.md](./plan.md)、产品差距见 [Roadmap.md](./Roadmap.md)。

### 应修复（行为与配置不一致）

| 项 | 现状 | 建议 |
| --- | --- | --- |
| 百分比预警 | `alert_rules` 仅 CRUD，无运行时 Worker | 入账或定时任务按阈值发通知；与 PRD US-08 对齐 |
| 超限文案 | `overrun_policy.blockMessage` 已存库，Precheck 返回通用错误 | Gateway 拒绝时读取并返回配置文案 |
| 预留池扣减 | 额度审批只校验 `reserved_pool` 上限，字段不随审批减少 | 审批通过时扣减预留池或维护「已分配预留」子账，避免重复透支 |
| 挂组 Key 预检 | Precheck 在存在 `member_id` 时仍检成员轴 | 与 Overrun / Rebalance 一致：挂组 Key 跳过成员 personal_budget 检查 |

### 应优化（可靠性 / 可观测）

| 项 | 现状 | 建议 |
| --- | --- | --- |
| NewAPI 关闭时 Worker | Rebalance / Overrun 可能空转或静默跳过 | NewAPI 未启用时 job 标记失败或明确 503，避免「以为已同步」 |
| 通知失败 | Webhook 失败常无感知 | 写 `notification_log` 失败态；关键事件告警 |
| 双层封禁窗口 | Precheck 通过后、入账前仍可能短暂超卖 | 评估是否收紧预估或缩短入账延迟；文档化可接受窗口 |
| 入账联调 | 依赖 logs 库、webhook secret、Worker 同时就绪 | 用 `pnpm verify:integration` 断言 Gateway + ledger（plan §1） |

### 可优化（体验 / 结构，非阻断）

| 项 | 说明 |
| --- | --- |
| 两套审批 | `budget_approvals` 与 `key_approvals` 并存，前端需分清入口 |
| 前端账期 | 演示环境仍有固定账期硬编码，应跟随后端 `period_key` |
| 列表规模 | Key 全量加载 + 内存 enrich；超 500 行时需 SQL 筛选与分页（plan §7） |
| 部门管理员 RBAC | 非管理员默认应只能看本部门 Key 与预算（plan §7 #4） |

### 暂不需要改

| 项 | 说明 |
| --- | --- |
| 双轴模型 | 钱包与组织预算分离是当前设计，运行正常 |
| snapshots 四轴 | consumed SSOT，与预检 / Overrun / UI 一致 |
| 自然月账期 | `period_key` 机制已满足按月清零 |
| 充值不涨部门 budget | 产品约定，非缺陷 |
