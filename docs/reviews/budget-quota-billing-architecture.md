# 预算 · 配额 · 账单架构

> **日期**：2026-07-20（v2 重写）  
> **范围**：企业钱包、组织预算、Gateway 预检、token 消耗入账、账单、充值、超限、预警  
> **代码验证**：所有描述均已与 `apps/backend/internal/` 源码核对

---

## 目录

1. [概念总览](#1-概念总览)
2. [核心数据模型](#2-核心数据模型)
3. [流程一：充值](#3-流程一充值)
4. [流程二：一次 API 调用（预检 → 消费 → 入账）](#4-流程二一次-api-调用预检--消费--入账)
5. [流程三：调整预算 → 配额刷新](#5-流程三调整预算--配额刷新)
6. [流程四：超限处理（Overrun）](#6-流程四超限处理overrun)
7. [流程五：预警通知（Alert）](#7-流程五预警通知alert)
8. [流程六：预算审批（Approval）](#8-流程六预算审批approval)
9. [关键文件索引](#9-关键文件索引)
10. [架构优化建议](#10-架构优化建议)

---

## 1. 概念总览

TokenJoy 的计费体系由两根**独立的轴**控成本：

```text
┌─────────────────────────────────────────────────────────┐
│  轴 1：企业钱包（wallet）— 硬上限，预付模型              │
│    充值 → lot（FIFO） → wallet_remain                   │
│    用完即止，允许 overdraft 扩展                         │
├─────────────────────────────────────────────────────────┤
│  轴 2：组织预算（budget）— 软限额，月度分配              │
│    部门 budget → 成员 personal_budget → Key budget       │
│    consumed 按月累加，新月归零                           │
│    combined_key_remain = min(各级剩余) → Gateway 预检    │
└─────────────────────────────────────────────────────────┘
```

两轴**同时满足**才放行请求；任一不足则 403。

### 1.1 量纲对齐

系统内部只有一个量纲：**point（quota）**。与 NewAPI 的 `logs.quota` 完全对齐，不做二次换算。

| 方向 | 公式 | 代码 |
|------|------|------|
| 展示币 → point | `round(amount × quota_per_unit)` | `common.QuotaFromAmount` |
| point → 展示币 | `quota / quota_per_unit` | `common.QuotaToDisplay` |

默认：**1 CNY = 500,000 point**（`DefaultQuotaPerUnit`）。

### 1.2 与 NewAPI 的关系

| 组件 | 当前状态 | 说明 |
|------|---------|------|
| NewAPI Token | `UnlimitedQuota = true` | 不限额，Gateway 不靠它做拦截 |
| NewAPI User | 充值时 TopUp(+delta) | 镜像钱包金额，仅防直连穿透 |
| NewAPI logs.quota | 真实消费金额 | Ingest 透传入账，不做换算 |

---

## 2. 核心数据模型

### 2.1 整体架构图

```mermaid
flowchart TB
  subgraph admin ["管理面（CRUD）"]
    BA[Budget API]
    KA[Keys API]
    BI[Billing API]
  end

  subgraph pg ["Postgres（SSOT）"]
    WAL[(wallet_remain<br/>+ lots FIFO)]
    ONB[(org_node_budget)]
    MEM[(members.personal_budget)]
    PK[(platform_keys.budget)]
    BC[(budget_consumed)]
    CKR[(combined_key_remain)]
    UL[(usage_ledger)]
  end

  subgraph cache ["Redis（可选热读）"]
    RC[(combined_key_remain<br/>cache)]
  end

  subgraph runtime ["运行时"]
    GW[Gateway /v1]
    NA[NewAPI]
    ING[IngestService]
  end

  BA -->|写 limit| ONB & MEM
  KA -->|写 Key budget| PK
  BA & KA -->|enqueue Rebalance| CKR
  BI -->|CreditFromLot| WAL
  BI -->|TopUp delta| NA

  GW -->|读| WAL & CKR & RC
  GW -->|proxy| NA
  NA -->|settle| ING
  ING -->|同事务| UL & BC & CKR & WAL
  ING -->|post-commit| RC
```

### 2.2 预算层级树

```text
企业
 └─ 根部门 (org_node_budget.budget)
      ├─ reserved_pool（预留给审批）
      ├─ 子部门 budget
      ├─ 成员 personal_budget
      │    └─ Key budget (scope=member)
      └─ 项目 budget
           ├─ Key budget (scope=project)
           └─ 项目成员子配额 + Key (scope=project_member)
```

### 2.3 consumed 三轴

`budget_consumed` 只记录三种 axis：

| axis_kind | axis_id | 用途 |
|-----------|---------|------|
| `platform_key` | Key ID | Key 级消耗 |
| `member` | 成员 ID | 成员总消耗 |
| `project` | 项目 ID | 项目总消耗 |

**部门花费**不在 consumed 里——通过 `usage_ledger.department_id` 聚合。

### 2.4 combined_key_remain 公式

`pkg/budget.GatewayChainRemain(scope, inputs)` 按 scope 取各级剩余的 **min**：

| scope | remain = min of |
|-------|-----------------|
| `member` | key_remain, personal_remain |
| `project` | key_remain, project_remain |
| `project_member` | sub_quota_remain, project_remain |

注意：`key_remain` 只在 `KeyBudget > 0` 时参与计算（无预算 = 无限制）。  
钱包不进入此链——在 Evaluate 中独立检查。

---

## 3. 流程一：充值

```mermaid
sequenceDiagram
  participant Admin as 管理员
  participant API as Billing API
  participant PG as Postgres
  participant NA as NewAPI

  Admin->>API: 充值确认（gift/adjust/paid）
  API->>API: QuotaFromAmount(amount, QPU)
  API->>PG: CreditFromLot<br/>(insert lot + wallet_remain += delta)
  API->>NA: TopUp(walletUserID, +delta)
  Note over NA: best-effort，失败仅 Warn
  API-->>Admin: 成功
```

**关键逻辑（`domain/billing/lot_confirm.go`）：**

1. 计算 `quotaGranted = QuotaFromAmount(amount, ppu)`
2. 构建 `RechargeOrder` + `RechargeLot`
3. `CreditFromLot`：事务内 insert lot + `ApplyWalletDelta(+quotaGranted)`
4. `topUpNewAPIQuota`：best-effort 调 NewAPI Admin TopUp

**FIFO lot 队列（`domain/billing/lot/consume.go`）：**
- 充值产生 lot（paid/gift/adjust/overdraft 四种）
- 消费时从 `FIFOHeadLotID` 开始逐个扣减
- lot 耗尽标记 `exhausted`，指针前移
- 余额不足时自动 `ExpandOverdraftLot`（透支扩展）

---

## 4. 流程二：一次 API 调用（预检 → 消费 → 入账）

这是系统最核心的流程。

### 4.1 Gateway 预检

```mermaid
flowchart TD
  Start[请求到达 Gateway] --> C1{企业状态<br/>gateway-blocked?}
  C1 -->|blocked| R1[403]
  C1 -->|正常| C2{wallet_remain >= 1?}
  C2 -->|不足| R2[403 insufficient wallet]
  C2 -->|足够| C3{Key active & 未过期?}
  C3 -->|否| R3[403 key inactive/expired]
  C3 -->|是| C4{模型在白名单?}
  C4 -->|否| R4[403 model not allowed]
  C4 -->|是| C5{combined_key_remain}
  C5 -->|nil 无限制| PASS[放行 → 反代 NewAPI]
  C5 -->|> 0| PASS
  C5 -->|<= 0| R5[403 budget exhausted]
```

**重要设计决策：**
- 预检**不做按请求估价**，只判断「还有没有额度」
- Redis 缓存 fail-open：miss 或 Redis 挂了 → 回退到 PG 查询
- `combined_key_remain` 的 version 字段用于防止 stale cache 误拦

### 4.2 完整入账流程

```mermaid
sequenceDiagram
  participant C as Client
  participant GW as Gateway
  participant PG as Postgres
  participant NA as NewAPI
  participant WH as Webhook
  participant ING as IngestService

  C->>GW: POST /v1/chat/completions
  GW->>PG: LoadPrecheckContext
  GW->>GW: Evaluate（见 4.1）
  GW->>NA: ReverseProxy
  NA-->>C: 模型响应（streaming）
  NA->>WH: POST /webhooks/newapi-log (settle log_id)
  WH->>ING: EnqueueIngestJob

  Note over ING: === 事务开始 ===
  ING->>PG: 1. LockForUpdate(company)
  ING->>PG: 2. ExistsIdempotency(newapi:{log_id})
  alt 重复
    ING-->>ING: skip
  else 首次
    ING->>PG: 3. ConsumeLotsLocked (FIFO 扣 lot)
    ING->>PG: 4. InsertSegments (usage_ledger)
    ING->>PG: 5. IncrementConsumedBatch (budget_consumed)
    ING->>PG: 6. DecrementBatch (combined_key_remain)
    ING->>PG: 7. EnqueueAfterIngest (overrun/dashboard jobs)
  end
  Note over ING: === 事务提交 ===

  ING->>ING: Post-commit (best-effort):
  ING->>PG: RefreshCombinedKeySummaries → Redis
  ING->>ING: CheckBudgetAlerts
  ING->>ING: NotifyOverdraft (if overdraft used)
```

### 4.3 入账金额来源

```text
NewAPI logs.quota ──直接透传──► entry.Amount
                                   │
                    ┌───────────────┼───────────────┐
                    ▼               ▼               ▼
             wallet_remain    budget_consumed   combined_key_remain
              (lot FIFO)        (+= amount)       (-= amount)
```

**没有任何换算**——`entry.Amount = input.Raw.Quota`（`entry_build.go` L80）。

### 4.4 Lot FIFO 消费细节

```mermaid
flowchart LR
  subgraph lots ["Lot 队列 (FIFO)"]
    L1[Lot A<br/>remain=2000]
    L2[Lot B<br/>remain=8000]
    L3[Lot C<br/>remain=5000]
  end

  consume["消费 3000 quota"]
  L1 -->|扣 2000| L1E[Lot A exhausted]
  L2 -->|扣 1000| L2R[Lot B remain=7000]

  consume --> L1
  L1E --> L2
```

若所有 lot 不足：`ExpandOverdraftLot` 创建透支 lot → 通知 overdraft expansion。

---

## 5. 流程三：调整预算 → 配额刷新

```mermaid
sequenceDiagram
  participant Admin as 管理员
  participant API as Budget API / Keys API
  participant PG as Postgres
  participant RB as RebalanceService
  participant Redis as Redis Cache

  Admin->>API: PUT 修改部门/成员/Key 预算
  API->>PG: 写 limit 到配置表
  API->>PG: InsertRebalance(axis, axisID)

  Note over RB: Worker 拉取 rebalance job
  RB->>PG: LoadBudgetContext (一次加载全量)
  loop 每个受影响的 PlatformKey
    RB->>RB: ComputeRemainForMapping
    RB->>PG: UpdateBatch(combined_key_remain)
    RB->>Redis: RefreshCombinedKeySummaries
  end
  Note over RB: Token unlimited → 不调 NewAPI
```

**关键点：**
- Rebalance 只刷 PG 的 `combined_key_remain`，**不同步 NewAPI Token**
- 代码注释：`"Token is unlimited on NewAPI — no remote quota to sync"`
- 下一个请求到达 Gateway 时即按新限额拦截

| 操作 | PG 影响 | NewAPI Token | NewAPI User |
|------|---------|-------------|-------------|
| 改部门/成员预算 | 写 limit + Rebalance | 不调用 | 不变 |
| 改 Key budget | 写 + RefreshPlatformKeyCombined | UpdateToken 只同步 status/models/group | 不变 |
| 创建 Key | 写 + Refresh | CreateToken(Unlimited=true) | 不变 |
| 充值 | Credit lot + wallet | 不变 | TopUp(+delta) |
| 消费入账 | wallet/consumed/remain 递减 | user quota 由 NewAPI 自扣 | — |

---

## 6. 流程四：超限处理（Overrun）

Overrun 在入账**事务内**构建 payload、事务后 enqueue job，由 `OverrunService` 异步处理。

```mermaid
flowchart TD
  Start[Ingest 入账完成<br/>enqueue overrun job] --> KEY{Key budget > 0<br/>且 consumed >= budget?}
  KEY -->|是| DK[Disable 该 Key<br/>+ 通知]
  KEY -->|否| MEM{scope=member 且<br/>personal consumed >= budget?}
  MEM -->|是| DM[Disable 该成员所有 Key<br/>+ 通知]
  MEM -->|否| PM{scope=project_member 且<br/>sub-consumed >= member_budget?}
  PM -->|是| DPM[Disable 项目中该成员的 Key<br/>+ 通知]
  PM -->|否| PROJ{project consumed >= budget?}
  PROJ -->|是| DP[Disable 项目所有 Key<br/>+ 通知]
  PROJ -->|否| DEPT{department ledger sum >= budget?}
  DEPT -->|是| NOTI[仅通知，不 Disable]
  DEPT -->|否| SKIP[无操作]
```

**评估顺序**（由细到粗）：Key → Member → ProjectMember → Project → Department

- Key/Member/Project 级别：**自动 Disable Key**（通过 `newapisync.DisablePlatformKey`）
- Department 级别：**仅发通知**（不自动禁用）
- 所有 Disable 动作通过 `OverrunKeyControl` 接口，既 Disable PG 记录也更新 NewAPI Token status

---

## 7. 流程五：预警通知（Alert）

```mermaid
sequenceDiagram
  participant ING as IngestService
  participant AP as AlertPublisher
  participant PG as Postgres
  participant NF as Notifier

  Note over ING: post-commit best-effort
  ING->>AP: CheckBudgetAlerts(touchedDepts)
  AP->>PG: 加载 alert_rules (enabled)
  AP->>PG: 加载 roles + members
  loop 每个 touched department
    AP->>PG: GetNodeBudget(dept)
    AP->>PG: SumAmountByDepartment(dept, periodKey)
    AP->>AP: pct = consumed * 100 / budget
    alt pct >= 最高阈值
      AP->>NF: PublishBudgetAlerts (按角色展开收件人)
    end
  end
```

**机制：**
- 触发时机：Ingest post-commit，对本次入账涉及的 department 检查
- 规则存储：`alert_rules` 表，每个规则绑定一个 `NodeID`（部门）
- 阈值：`rule.Thresholds` 数组，取已突破的最高阈值
- 收件人：通过 `NotifyRoleIDs` → 角色名 → 成员 ID 展开
- 去重：`DedupeKey = budget-alert:{companyID}:{ruleID}:{threshold}:{periodKey}:{memberID}`

---

## 8. 流程六：预算审批（Approval）

```mermaid
sequenceDiagram
  participant M as 成员
  participant API as Budget API
  participant PG as Postgres
  participant RB as RebalanceService

  M->>API: 申请额度（从 reserved_pool）
  API->>PG: Insert budget_approvals (status=pending)

  Note over API: 管理员审批
  API->>PG: AcquireBudgetLock
  API->>PG: 读 org_node_budget.reserved_pool
  alt reserved_pool >= amount
    API->>PG: reserved_pool -= amount
    API->>PG: member.personal_budget += amount
    API->>PG: UpdateBudgetApproval(approved)
    API->>RB: InsertRebalance(member axis)
  else 不足
    API-->>API: 错误：预留池余额不足
  end
```

**已实现的完整闭环：**
- 审批通过时在同一事务内扣减 `reserved_pool` 并增加成员 `personal_budget`
- 之后触发 member 轴的 Rebalance 刷新 `combined_key_remain`

---

## 9. 关键文件索引

| 职责 | 路径（相对 `apps/backend/internal/`） |
|------|------|
| Gateway 预检 | `domain/gateway/evaluate.go` |
| 预检上下文 | `domain/gateway/precheck_context.go` |
| Remain 链公式 | `pkg/budget/chain.go` |
| 刷 combined remain | `domain/budget/combined_key_summary.go` |
| Rebalance | `domain/budget/rebalance.go` |
| Redis 缓存 | `domain/budget/combined_key_cache.go` |
| 入账 | `domain/usage/ingest.go` |
| 金额构建 | `domain/usage/entry_build.go` |
| Lot FIFO 消费 | `domain/billing/lot/consume.go` |
| 充值确认 | `domain/billing/lot_confirm.go` |
| NewAPI TopUp | `domain/billing/wallet_topup.go` |
| 超限处理 | `domain/budget/overrun.go` |
| 预警检查 | `domain/budget/alert_publisher.go` |
| 预算审批 | `domain/budget/approvals.go` |
| Token 同步 | `integration/newapi/admin_port_adapter.go` |
| 预算树变更 | `domain/budget/tree_mutate.go` |
| QPU 常量 | `pkg/common/constants.go` |

---

## 10. 架构优化建议

### 10.1 可简化项

| 现状 | 问题 | 建议 |
|------|------|------|
| **NewAPI User TopUp** 镜像钱包 | 仅防直连穿透，但增加了充值路径复杂度和 TopUp 失败告警噪音 | 如果所有流量必须经过 Gateway，可评估移除 User quota 依赖，改为网络层封堵直连 |
| **Overrun job 异步 Disable** | 入账与 Disable 之间有时间窗口，期间可能多放行 1-2 个请求 | 若 combined_key_remain 已 ≤ 0，Gateway 下一请求自然拦截；Overrun Disable 是补充保险，当前设计可接受 |
| **tree_mutate.go 过时注释** | L218/L227 仍写 "NewAPI token remain_quota" | 改为 "refresh combined_key_remain"（纯文本修改，零风险） |

### 10.2 可提升效率项

| 现状 | 瓶颈 | 优化方向 |
|------|------|---------|
| **Ingest 按 company LockForUpdate** | 高并发同公司串行化 | 可拆细锁粒度到 platform_key 级别（lot 消费仍需公司锁，但 consumed/remain 可并行）；或接受当前简单方案 |
| **Rebalance 全量加载 BudgetContext** | 大租户 keys 多时 LoadBudgetContext 查询重 | 按 axis 范围只加载受影响的 keys/consumed；但当前 preload-once 策略在大多数租户下性能足够 |
| **CheckBudgetAlerts 每次 post-commit 全查 rules** | 每次入账后都 SELECT alert_rules | 可缓存 rules（TTL 几分钟）；或改为 Ingest 只 enqueue，由独立 worker 批量检查 |
| **部门花费报表靠 ledger 实时聚合** | 数据量增长后 SUM 变慢 | 增加 `budget_consumed` 的 `org_node` 轴作为预聚合；或用 `usage_buckets` 物化视图 |

### 10.3 可收紧的风险窗口

```mermaid
flowchart LR
  subgraph window ["超卖窗口"]
    A[Gateway 预检通过] --> B[NewAPI 执行]
    B --> C[Webhook 到达]
    C --> D[Ingest 入账]
  end

  style window fill:#fff3cd
```

| 风险 | 影响 | 收紧方案 | 成本 |
|------|------|---------|------|
| **Precheck → Ingest 超卖** | 最后一滴额度可能多放 N 个并发请求 | Redis 原子 DECRBY 预扣（粗估费用），Ingest 后 reconcile | Gateway 增加 Redis 写；需处理估价不准时的 reconcile |
| **TopUp best-effort 失败** | PG 已充值，NewAPI user 偏低，直连时被拒 | 已有 bootstrap 补齐；可加「TopUp 失败」告警 + 重试队列 | 低成本 |
| **absoluteRecompute 锁竞争** | 新 Key 首次消费走 LockPlatformKeysForUpdate | 创建 Key 时同步 INSERT combined_key_summaries 初始行 | 极低成本 |

### 10.4 架构层面的简化机会

**当前复杂度根源**：双轴模型 + NewAPI 镜像 + 异步入账。

如果要做更大的架构简化，有两个方向值得评估：

1. **去掉 NewAPI User quota 依赖**  
   前提：所有流量 100% 经 Gateway（网络策略保证）。  
   收益：移除 TopUp 逻辑、bootstrap 补齐、充值路径的 NewAPI 调用；lot_confirm 只写 PG。  
   风险：失去「最后一道」物理止损。

2. **预检引入 Redis 预扣（收紧超卖）**  
   在 Gateway 对 `combined_key_remain` 做 `DECRBY estimated_cost`（用模型价目表粗估）。  
   Ingest 入账后做真实 reconcile（INCRBY 差额或重置）。  
   收益：超卖窗口从「整个请求延迟 + ingest 延迟」缩短到「估价误差」。  
   成本：Gateway 增加 Redis 写 + 估价逻辑；Redis 不可用时回退当前 fail-open。

这两个方向都不是必须的——当前架构在中等规模下运行正常。建议在出现实际超卖投诉或 TopUp 噪音影响运维时再推进。

### 10.5 已无需修改的正确设计

| 设计 | 理由 |
|------|------|
| 双轴模型（钱包 + 组织预算） | 职责清晰，互不干扰 |
| Token Unlimited + User TopUp | 与 ADR 一致 |
| point ≡ NewAPI quota（无换算） | 正确简化 |
| consumed 三轴 + ledger 审计 | SSOT 清晰 |
| 充值不涨部门 budget | 产品约定 |
| Lot FIFO + overdraft 扩展 | 灵活处理余额不足 |
| Overrun 逐级评估（细→粗） | 精确定位超限范围 |

---

## 附录：与旧文档的差异

| 旧文档描述 | 实际代码 | 本文档修正 |
|-----------|---------|-----------|
| "审批通过后不扣减 reserved_pool" | `approvals.go` 同事务扣减 | §8 已正确描述 |
| tree_mutate 注释写"同步 NewAPI token remain_quota" | Rebalance 只刷 PG | §5 明确说明 |
| gateway_soft_* 命名 | 已改名 `combined_key_remain*` | 全文统一使用新名称 |
| Ingest 入队 wallet_sync | 充值后直接 TopUp | §3 描述当前逻辑 |
| Gateway 钱包≥预估 | 仅 `wallet_remain >= 1` | §4.1 流程图明确 |
| pkg/newapiunits 做 point↔quota 换算 | 换算函数已删 | §1.1 说明无换算 |
