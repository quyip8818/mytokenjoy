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
    A3[成员额度 / 项目 / Key 配额]
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
| 部门 TL | 本部门预算、预留池、成员额度、项目 |
| 普通成员 | 个人额度、Key 配额、能否继续调用 |
| 审计 / 财务 | 调用花费、归因部门 / 成员 |

**账期：** 分配配置（`budget`、`personal_budget`、Key `budget`）跨月保留；已消耗按开账 `period_key`（通常 `YYYY-MM`，来自业务时钟）写入 `budget_consumed`，新月自动从新账期累计。账本发生月见 [Backend-业务时钟与账期.md](./Backend-业务时钟与账期.md)。

---

## 2. 两条轴

| 轴 | 权威数据 | 管什么 | 谁改 |
| --- | --- | --- | --- |
| **企业钱包** | 充值 lot、`wallet_remain` | 预付资金硬上限（point） | 充值确认 → 异步同步 NewAPI |
| **组织预算** | 组织树 limit + `budget_consumed` | 部门 / 成员 / Key / 组的花费与上限 | 控制台写 limit；**Ingest 同事务**累加 consumed |

```mermaid
flowchart TB
  subgraph wallet [钱包轴]
    LOT[充值 lot] --> BAL[wallet_remain]
    BAL -.-> NA_W[NewAPI users.quota]
  end

  subgraph org [组织预算轴]
    TREE[组织树 budget / reserved_pool]
    TREE --> MEM[personal_budget]
    TREE --> BG[项目 budget]
    MEM & BG --> PK[Key budget]
    SNAP[(budget_consumed)]
    ING[入账同事务] --> SNAP
    ING --> CK[combined_key_remain]
  end

  BAL -->|Rebalance 封顶| PK
  SNAP -->|已用| PK
  CK -->|Gateway 预检| GW[预检]
```

**约定：**

- 充值**只涨钱包**，不自动涨部门 `budget`。
- **limit** 在组织树、成员、Key、项目；**consumed** 只在 `budget_consumed`（三轴 × 账期；`project_member` sub 已用见 `pkg/budget/chain.go`）。
- API 返回的 `consumed` 为当前账期从快照合并的视图，不是 Key 表上的持久列。

---

## 2. SSOT 与读写路径（终态）

| 层 | 存储 | 写入 | 读方 |
| --- | --- | --- | --- |
| **事实** | `usage_ledger` | `usage.IngestService` | 审计、minute 看板 |
| **累计（热）** | `budget_consumed`、`combined_key_remain` | **Ingest 同事务** | 预算树、Overrun 判定、Gateway 预检 |
| **展示投影** | `usage_buckets` | `dashboard.Projector`（稀） | hour/day 看板 |
| **冷矫正** | 同上累计表 | `budget_reconcile` 窗口 `SetConsumed` | 修漂移 |

终态：**无** `budget_projection` / 游标 budget Projector。细节：[Backend-Projector.md](./Backend-Projector.md) · [Backend-budget_consumed迁回Ingest.md](./Backend-budget_consumed迁回Ingest.md)。

```mermaid
flowchart LR
  WH[Webhook] -->|EnqueuePending| Q[(ingest_jobs)]
  Q --> ING[IngestService]
  COMP[补偿轮询] --> ING
  ING --> UL[(usage_ledger)]
  ING --> BC[(budget_consumed)]
  ING --> CK[(combined_key_remain)]
  ING --> DP[dashboard_project]
  ING --> WS[wallet_sync]
  DP --> UB[(usage_buckets)]
  CK -->|仅可能触顶| OV[overrun_可选]
```

### 2.1 入账路径（终态）

1. NewAPI settle → 共享 `logs` → webhook / reconcile → `IngestByLogID`
2. 归因 + `BuildCallSettledEntry`（幂等键 `newapi:{log_id}`）
3. `store.WithTx`：ledger → FIFO lot → **`ApplyIncrement(budget_consumed)`** → **`DecrementBatch(combined_key_remain)`** → 入队 `dashboard_project` + `wallet_sync`
4. 可选：轻量预判后才 `InsertOverrun`；**不**入队 `budget_projection`
5. rebalance / 重 overrun：async；多数 Ingest 零 budget job

### 2.2 `budget_consumed` 三轴（Ingest 内）

消耗统一在 `budget_consumed`（`axis_kind`）。业务表无 `consumed` 列。

**三轴：** `platform_key` · `member` · `project`（**不写** `org_node`）。部门花费：`usage_ledger` 按 `department_id` 聚合。Gateway：`combined_key_remain`（`budgetcheck` 缓存）。

| 步骤 | 写入 | 说明 |
| ---- | ---- | ---- |
| 1 | `platform_key` += cost | Key 已用 |
| 2 | `project` += cost | `project` / `project_member` scope |
| 3 | `member` += cost | 仅 `member` scope |
| — | 无 org_node 轴 | 部门报表用 ledger |
| 同事务 | `combined_key_remain` | 预检热读 |
| 提交后 | 告警可直做；overrun/rebalance 按需 | [Backend-Projector.md](./Backend-Projector.md) §3 |

父节点 **limit**：`org_nodes.budget`。看板桶：`dashboard.Projector`。

### 2.3 读路径分离

| 场景 | 读什么 | 为何 |
| --- | --- | --- |
| Gateway 预检 | `combined_key_remain` + limit | ≤0 403；经 budgetcheck 缓存 |
| 看板 cost 趋势 | `usage_buckets` SUM | Dashboard 投影 |
| 预算树 limit | `org_nodes.budget` 等 | 部门 spent 不读 org_node consumed 轴 |
| 部门本月花费 | `usage_ledger` 聚合 | 替代 org_node consumed |
| 审计调用列表 | `usage_ledger` | SSOT |
| minute 趋势 | `usage_ledger` 按分钟 | 窗口 ≤3h |

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
  DEPT --> BG[项目 + 组内 Key]
```

| 层级 | 配置 | 说明 |
| --- | --- | --- |
| 部门 | `budget`、`reserved_pool` | 子节点之和 + 预留池 ≤ 父节点 |
| 成员 | `personal_budget` | 部门内成员额度之和 ≤ capacity |
| Key | `budget`、模型白名单 | 从成员或项目剩余额度切分 |
| 项目 | `budget` | 挂项目 Key 走项目额度；Overrun 不走成员个人分支 |

**写入校验：**

| 操作 | 规则 |
| --- | --- |
| 改部门预算 | 子级：新 budget ≥ Σ子节点 + 预留池；对父级：新 budget + 兄弟 + 预留池 ≤ 父可用 |
| 改成员额度 | ≥ 已分配给 Key 的配额之和；部门内总和 ≤ capacity |
| 建 Key（成员） | budget ≤ 成员剩余可分配 |
| 建 Key（项目） | budget ≤ 组 budget − 组 consumed − 组内已分配 Key budget（含 `project_member`） |
| 建 Key（项目成员） | roster + `member_budget > 0`；budget ≤ sub 剩余；见 `pkg/budget/scope_validate.go` |
| 改项目成员子额度 | `PUT /api/budget/projects/{id}` · `memberBudgets`；须属于 roster |
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
| 项目 | CRUD | `/api/budget/projects/*`（含 `memberBudgets` patch） |
| 项目成员已用 | GET | `/api/budget/projects/{id}/member-consumed` |
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
  ING->>ING: 账本 + lot + 入队 budget_projection
  Note over ING: Projector 异步写 consumed
  ING->>W: rebalance / overrun（批末）
  W->>NA: UpdateToken
  W->>W: 超限则 Disable Key
```

**Gateway 预检（同步）** — 全部通过才代理（单位 point）；1× `LoadPrecheckContext` + 纯内存 `Evaluate`：

| scope | 公式（与 [预算分配与扣减.md](./预算分配与扣减.md) §14 一致） |
| --- | --- |
| `member` | `min(key, personal, wallet)` — **不含**未分配/预留池/部门报表 |
| `project` | `min(key, project, wallet)` |
| `project_member` | `min(key, sub_quota, project, wallet)`；sub 已用 = Σ Key 聚合 |

| 检查 | 数据 |
| --- | --- |
| 企业 active | `companies.status` |
| 钱包 ≥ 预估 | `wallet_remain` |
| Key / personal / 项目未超 | `gateway_soft_remain` + limit（`LoadPrecheckContext`） |
| 模型与 Key 状态 | allowlist、`platform_keys.status` |

NewAPI quota 与 `wallet_sync` **不参与**热路径预检；Gateway 读 Postgres `wallet_remain` 与 `gateway_soft_*`；漂移由异步 `wallet_sync` 与对账消化。

---

## 6. 数据层

```mermaid
flowchart TB
  UL[(usage_ledger 事实)]
  BS[(budget_consumed 三轴)]
  GS[gateway_soft_*]
  UB[(usage_buckets 看板)]
  CFG[配置表 limit]

  ING[入账] --> UL
  ING -->|budget_projection| BS
  BS --> GS
  PER[dashboard_project] --> UB
  CFG --> GW[预检]
  GS --> GW
  BS --> UI[预算树 / Key 列表]
  UL --> AUD[审计]
  UB --> DASH[看板]
```

| 存储 | 职责 |
| --- | --- |
| `usage_ledger` | 消耗 SSOT；幂等 `newapi:{log_id}` |
| `budget_consumed` | 三轴 `platform_key` · `member` · `project`；部门报表读 `usage_ledger` 聚合 |
| `platform_keys.gateway_soft_*` | Gateway 预检软剩余（Projector 批末刷新） |
| `usage_buckets` | 按小时聚合，供趋势图 |
| 组织树 / 成员 / Key / 组 | 仅存 limit |

| 读场景 | 数据源 |
| --- | --- |
| 预算树 limit、Key 已用 | `org_nodes.budget` 等配置 + `budget_consumed`（三轴） |
| Gateway 预检 | `gateway_soft_remain` + limit |
| 看板趋势 | `usage_buckets` |
| 调用审计 | `usage_ledger` |
| 分钟级短趋势 | `usage_ledger` 聚合 |

部门本月花费读 `usage_ledger` 按 `department_id` 聚合。表结构见 [Backend-存储架构.md](./Backend-存储架构.md) §5–§8。

---

## 7. 入账与累计（终态）

```mermaid
flowchart LR
  NA[NewAPI 结算] --> LOGS[(logs 库)]
  LOGS --> WH[Webhook / reconcile 补偿]
  WH --> ING[Ingest]
  ING --> TX[单事务]
  TX --> L[usage_ledger]
  TX --> F[FIFO 扣 lot]
  TX --> BC[budget_consumed]
  TX --> CK[combined_key_remain]
  TX --> Q[dashboard_plus_wallet_wake]
```

1. 结算日志 → Webhook 或补偿 → 按 mapping 归因  
2. 单事务：账本 → lot → **consumed + combined_key** → 稀唤醒 dashboard / wallet  
3. overrun / rebalance：**按需**异步（多数跳过）；见 [Backend-Projector.md](./Backend-Projector.md)  
4. 失败走 `ingest_jobs` 重试  

`usage_buckets` 由 `dashboard.Projector` 独立维护。账期见 [Backend-业务时钟与账期.md](./Backend-业务时钟与账期.md)。

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
| member | 入账带成员（`member` scope） |
| project | 入账命中项目 |
| platform_key | Key 创建 / 变更 / 入账 |
| company | 充值完成 |

（**已移除** `org_node` rebalance 触发；部门触顶仅 notify。）

去重：`dedupe_key = axis_kind:axis_id`。

**候选最小值（point）：** `GatewayChainRemain` 按 key `scope` 计算 remain → 换 NewAPI 单位（`ToNewAPIUnits` 饱和），并以 `wallet_remain` 作企业硬顶；多 key 累加用 `AddSat`/`SubFloor0`，避免 int64 绕回。不再按部门 org_node consumed 封顶。

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
| Platform Key | platform_key 轴 consumed ≥ key.budget | disable 该 Key |
| 成员 personal | member 轴 consumed ≥ personal_budget | 禁用该成员 **member** scope Key |
| 部门 | ledger 聚合 ≥ `org_nodes.budget` | **通知 only**；不封 Key |
| 项目 | project 轴 consumed ≥ budget | 禁用该项目 **project** + **project_member** Key |
| project_member sub | Σ Key consumed ≥ `member_budget` | 禁用该人该项目 **project_member** Key |

personal 用尽后的追加路径：**US-10 额度审批**（预留池 → `personal_budget`），不是运行时自动蹭未分配。见 [预算分配与扣减.md](./预算分配与扣减.md)。

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
  B->>DB: lot + wallet_remain
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
| 成员本账期已用 | `budget_consumed` member 轴 |
| 组可分给 Key | 组 budget − 组 consumed − Σ组内 Key budget |
| NewAPIKey 可用上限 | 上列候选取 min → 换 NewAPI 单位 |
| 企业硬顶 | Σ NewAPIKey remain ≤ wallet_remain 对应通道配额 |

---

## 12. 负责域

| 职责 | 域 |
| --- | --- |
| 预算树、组、成员额度、预警策略 | `domain/budget` |
| 入账与 ledger | `domain/usage` |
| 预算 / consumed 投影 | `domain/budget`（`budget_projector.go`） |
| 看板 buckets 投影 | `domain/dashboard` |
| Rebalance | `domain/budget/rebalance`（`adminport.Port` 更新 token） |
| NewAPI Admin 边界 | `domain/adminport` + `integration/newapi/admin_port_adapter.go` |
| Quota 换算 | `pkg/newapiunits` |
| Key 额度校验 | `domain/keys` + `pkg/budget` |
| Gateway 软缓存 | `domain/budget/gateway_summary.go` + `infra/budgetcheck` |
| consumed 加载 | `pkg/budget` + `store.BudgetConsumed()` |
| Gateway 预检 | `domain/gateway` |
| 充值 | `domain/billing` |
| 异步任务 | `river_job`（River；见 [Backend-离线任务.md](./Backend-离线任务.md)） |

---

## 13. 待优化与待修复

按优先级归纳；工程细节另见 [plan.md](./plan.md)、产品差距见 [Roadmap.md](./Roadmap.md)。

### 应修复（行为与配置不一致）

| 项 | 现状 | 建议 |
| --- | --- | --- |
| 百分比预警 | `alert_rules` 仅 CRUD，无运行时 Worker | 入账或定时任务按阈值发通知；与 PRD US-08 对齐 |
| 超限文案 | `overrun_policy.blockMessage` 已存库，Precheck 返回通用错误 | Gateway 拒绝时读取并返回配置文案 |
| 预留池扣减 | 额度审批只校验 `reserved_pool` 上限，字段不随审批减少 | 审批通过时扣减预留池或维护「已分配预留」子账，避免重复透支 |

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
| `budget_consumed` 三轴 | consumed SSOT；Gateway 读 `gateway_soft_*`；与 Overrun / UI 一致 |
| 自然月账期 | `period_key` 机制已满足按月清零 |
| 充值不涨部门 budget | 产品约定，非缺陷 |
