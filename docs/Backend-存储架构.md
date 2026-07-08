# Backend 存储架构

Postgres 双库：**35** 张主库表 + **3** 张日志库表。`company_id` 租户隔离，管理面配置 + 运行面入账投影。

**本文定位：** 表结构、域关系、Store 映射与 ID 约定。请求链路见 [Backend-架构.md](./Backend-架构.md)；Ingest / Rebalance / Overrun 算法见 [Backend-预算.md](./Backend-预算.md)；计费单位见 [Backend-计费单位.md](./Backend-计费单位.md)。

| 库 | DDL | 连接 |
| --- | --- | --- |
| 主库 | `apps/backend/internal/store/postgres/schema.sql` | `DATABASE_URL` |
| 日志库 | `logs_schema.sql` | `LOG_DATABASE_URL`（可选） |

启动时 `go:embed` 全量 apply，再由 `schema_partitions.go` 创建月分区（2024-01 .. 2032-12）。改表后本地 `docker compose down -v` 重建。

---

## 1. 双库拓扑

```mermaid
flowchart LR
  subgraph main [主库 · 35 表]
    MGMT[管理面]
    RUN[运行面]
  end

  subgraph logs [日志库 · 3 表]
    NL[newapi.logs]
    IF[ingest_failures]
    RC[reconcile_cursors]
  end

  WH[Webhook] --> RUN & NL
  ING[Ingest] --> RUN
  WK[Worker] --> RC & IF & NL
  IF -.->|重试| ING
```

| 配置 | 行为 |
| --- | --- |
| `LOG_DATABASE_URL` + `NEW_API_WEBHOOK_SECRET` | Ingest 启用 |
| 未配置日志库 | `Logs()` 为 `NoopLogStore` |
| `LogSchemaIsolated=true` | 测试专用；生产禁止 |

---

## 2. 域级关系

```mermaid
flowchart TB
  CO[Company] --> ORG[组织<br/>org_nodes · members · roles]
  CO --> BUD[预算<br/>budget_groups · snapshots]
  CO --> KEY[密钥<br/>platform_keys]
  CO --> MOD[模型<br/>models · allowlist]
  ORG --> BUD & KEY
  MOD --> KEY
  KEY --> RLY[relay · async_jobs]
  RLY --> USG[ledger · buckets · snapshots]
```

| 概念 | 落点 |
| --- | --- |
| 部门 + 节点预算 + 路由 | `org_nodes`（HTTP 投影 `Department` / `BudgetNode` / `RoutingRule`） |
| 平台 Key 归因 | `platform_keys`；`relay_mappings` 仅同步状态 |
| 消耗 SSOT / 看板 / 预检缓存 | `usage_ledger` / `usage_buckets` / `budget_snapshots` |
| 企业钱包余额 | NewAPI `users.quota`；Postgres 存 `newapi_wallet_user_id` |
| SaaS 上游 Key | 全局 `provider_keys`（无 `company_id`） |

---

## 3. Store 与租户隔离

```mermaid
flowchart TB
  SVC[domain.Service] --> ST[store.Store]
  ING[Ingest] --> CW[ConsumptionWriter]
  CW --> LED & USG & SNAP & ORG & KEY & RLY
  ST --> MAIN[(主库)]
  ST --> LOG[(日志库)]
```

| 接口 | 主表 |
| --- | --- |
| `Org()` / `Nodes()` | `org_nodes`, `members`, `roles`, `permissions`, `org_integration`, … |
| `Budget()` | `budget_groups`, `budget_snapshots`, `alert_rules`, `budget_approvals`, … |
| `Keys()` | `provider_keys`, `platform_keys`, `key_approvals` |
| `Models()` / `Allowlist()` | `models`, `model_capabilities`, `model_allowlist` |
| `Ledger()` / `Usage()` / `BudgetSnapshots()` | `usage_ledger`, `usage_buckets`, `budget_snapshots` |
| `Relay()` | `relay_mappings`, `async_jobs` |
| `Audit()` | `audit_settings`, `operation_logs` |
| `Company()` / `Invite()` / `Billing()` / `Platform()` | 租户与充值 |
| `Notification()` / `SchedulerLock()` / `Logs()` | `notification_log`, `scheduler_locks`, 日志库三表 |

**租户：** 复合 PK `(company_id, …)`；全局表 `permissions` · `provider_keys` · `platform_operators` · `scheduler_locks`。

**分区：** `operation_logs` · `usage_ledger` · `usage_buckets` → 月子表 `{table}_YYYY_MM`。

---

## 4. 核心实体

### 组织与 RBAC

```mermaid
erDiagram
  Company ||--o{ OrgNode : org_nodes
  Company ||--o{ Member : members
  OrgNode ||--o{ OrgNode : parent
  OrgNode ||--o{ Member : department_id
  Member }o--o{ Role : member_roles
  Role }o--o{ Permission : grants
```

`members.department_id` = `org_nodes.id` = HTTP `departmentId`。组织同步：`org_integration` + `org_sync_logs` / `org_import_failures`。

### 组织节点 vs 预算组

| | `org_nodes` | `budget_groups` |
| --- | --- | --- |
| 结构 | 树形逐级分配 | 扁平共享池 |
| 与 Key | 经部门归因 | 挂 `budget_group_id` |
| consumed | `axis_kind=org_node` | `axis_kind=budget_group` |

### `org_nodes` 列组

| 列组 | 字段 | 读写入口 |
| --- | --- | --- |
| 树 | `id`, `parent_id`, `path`, `name`, `manager_id` | `Org().Nodes()` |
| 预算 | `budget`, `reserved_pool`, `period` | `budget.Service`；consumed → `budget_snapshots` |
| 路由 | `default_model_id`, `fallback_model_id`, `routing_inherited` | `models.Service`；白名单 → `model_allowlist` |

### 密钥与 Relay

```mermaid
flowchart LR
  PPK[provider_keys] -.-> PLK[platform_keys]
  M[members] --> PLK
  BG[budget_groups] --> PLK
  PLK --> RM[relay_mappings] --> NA[NewAPI]
```

`key_hash` 用于 Gateway 鉴权；明文 Key 不落库。`model_allowlist.owner_type`：`platform_key` · `org_node` · `key_approval`。

### `async_jobs`

| channel | 消费方 |
| --- | --- |
| `relay` | `relay_processor` |
| `rebalance` | `rebalance_processor`（`dedupe_key` = `axis_kind:axis_id`） |
| `overrun` | `overrun_processor` |

---

## 5. 用量与入账

```mermaid
flowchart LR
  WH[Webhook] --> ING[Ingest]
  RC[reconcile] --> NLG[(newapi.logs)] --> ING
  ING --> UL[(usage_ledger)]
  ING --> PROJ[projection] --> UB[(buckets)] & BS[(snapshots)]
  UL --> AUD[审计]
  UB --> DASH[看板]
  BS --> GW[预检]
```

| 表 | 角色 |
| --- | --- |
| `usage_ledger` | 消耗 SSOT |
| `usage_buckets` | 看板 hour/day 聚合 |
| `budget_snapshots` | 四轴 consumed：`org_node` · `budget_group` · `platform_key` · `member` |
| `ingest_failures` | 入账失败重试（日志库） |

入账算法见 [Backend-预算.md](./Backend-预算.md) §2。

---

## 6. 表清单

### 主库（35 张）

| 域 | 表 |
| --- | --- |
| 租户 | `companies`, `company_invites`, `company_recharge_orders`, `platform_operators` |
| 组织 | `org_nodes`, `members`, `roles`, `permissions`, `role_permission_grants`, `member_roles`, `org_integration`, `org_sync_logs`, `org_import_failures` |
| 预算 | `budget_groups`, `budget_snapshots`, `budget_group_members`, `budget_group_departments`, `overrun_policy`, `alert_rules`, `alert_rule_notify_roles`, `budget_approvals` |
| 密钥 | `provider_keys`, `platform_keys`, `key_approvals`, `relay_mappings` |
| 模型 | `models`, `model_capabilities`, `model_allowlist` |
| 审计 | `audit_settings`, `operation_logs`, `usage_ledger` |
| 运行面 | `usage_buckets`, `async_jobs`, `scheduler_locks`, `notification_log` |

### 日志库（3 张）

| 表 | 职责 |
| --- | --- |
| `newapi.logs` | consume 原始行（`type=2`） |
| `backend.ingest_failures` | 入账失败重试 |
| `backend.reconcile_cursors` | reconcile 水位 |

---

## 7. 关键 ID

| 易混项 | 说明 |
| --- | --- |
| `departmentId` | = `org_nodes.id` = `members.department_id` |
| `RoutingRule.id` | = `nodeId` |
| `sk-xxx` | → `platform_keys.key_hash` → `relay_mappings.newapi_token_id` |
| `newapi_wallet_user_id` | → NewAPI `users.quota` |
| 幂等键 | `newapi:{log_id}` |
| `members` | TokenJoy 成员，非 NewAPI user |
| `personalQuota` | 走 `MemberBudgetQuota` API，不在 Member JSON |

---

## 8. 消耗与额度术语

代码里 `consumed` / `used` / `quota` / `budget` 并存，**语义可统一为三个概念**；下列为文档与评审的**标准读法**（不要求改字段名）。

### 8.1 三个统一概念

| 统一词 | 含义 | 典型计算 |
| --- | --- | --- |
| **limit**（上限） | 管理员配置的本周期可花额度 | 控制台写入 |
| **consumed**（已消耗） | 本周期已累计花费（CNY） | Ingest 累加 |
| **remaining**（剩余） | 还能花多少 | `limit - consumed`（多为 API 计算，少落库） |

```mermaid
flowchart LR
  subgraph assign [配置 limit]
    OB[org_nodes.budget]
    PQ[members.personal_quota]
    BG[budget_groups.budget]
    PKQ[platform_keys.quota]
    WAL[NewAPI users.quota]
  end

  subgraph spend [累计 consumed]
    BS[(budget_snapshots.consumed)]
    UL[(usage_ledger.amount_cny)]
  end

  subgraph api [JSON 展示]
    C[consumed]
    U[used]
    R[remaining / remain]
  end

  ING[Ingest] --> BS & UL
  BS --> C & U
  assign --> R
  BS --> R
```

### 8.2 两条轴（limit 归属不同）

| 轴 | limit 权威源 | consumed 权威源 | 交汇点 |
| --- | --- | --- | --- |
| **企业钱包** | NewAPI `users.quota` | NewAPI 侧扣减（非 Postgres 主账） | rebalance 给 Token 分 `remain_quota` |
| **组织预算** | `org_nodes.budget` · `personal_quota` · `budget_groups.budget` · `platform_keys.quota` | **`budget_snapshots.consumed`**（四轴） | Gateway 预检、预算树、Overrun |

组织轴 **consumed 不以列形式存在**于 `org_nodes` / `platform_keys`；`budget_snapshots` 是 consumed 的存储 SSOT。单笔事实在 `usage_ledger.amount_cny`。

### 8.3 字段对照（代码名 → 统一词）

| 统一词 | 代码 / 表字段 | 实体 | 说明 |
| --- | --- | --- | --- |
| **limit** | `org_nodes.budget` | 部门节点 | 组织树分配额 |
| **limit** | `members.personal_quota` | 成员 | 个人可分配上限 |
| **limit** | `budget_groups.budget` | 预算组 | 池额度 |
| **limit** | `platform_keys.quota` | 平台 Key | Key 分配额 |
| **limit** | NewAPI `users.quota` | 企业钱包 | 预付资金硬顶 |
| **limit** | NewAPI `remain_quota` / `relay_mappings.newapi_token_remain_quota` | Token | Relay 侧剩余额度（分配视图，非组织 consumed） |
| **consumed** | `budget_snapshots.consumed` | 四轴 | **组织轴 consumed SSOT** |
| **consumed** | `usage_ledger.amount_cny` | 单笔调用 | 事实账本 |
| **consumed** | `usage_buckets.cost_cny` | 看板聚合 | 展示投影 |
| **consumed** | JSON `consumed` | `BudgetNode` · `Department` · 看板 | 读自 snapshot |
| **consumed** | JSON `used` | `PlatformKey` · `MemberBudgetQuota` | **与 consumed 同义**，仅 JSON 名不同 |
| **remaining** | JSON `remaining` / `remain` | `MemberQuotaSummary` 等 | 计算字段 |
| **remaining** | NewAPI `remain_quota` | Token | Relay 配额，受钱包 rebalance 封顶 |

### 8.4 读代码速查

| 看到 | 统一读作 | 备注 |
| --- | --- | --- |
| `used` | **consumed** | 仅 Platform Key / 成员额度 API 用 `used` 字段名 |
| `quota`（`platform_keys`） | **limit** | Key 分配上限 |
| `quota`（NewAPI user） | **limit**（钱包轴） | 与企业组织 budget 独立 |
| `remain_quota` | **remaining**（Token） | 不是 Postgres consumed |
| `budget` | **limit** | 部门/预算组语境 |
| `personalQuota` | **limit**（成员） | 不在 `Member` JSON，走专用 API |
| `amount_cny` / `cost_cny` | **consumed**（增量/聚合） | 账本与看板 |

### 8.5 不在此表内的词

| 字段 | 原因 |
| --- | --- |
| `provider_keys.balance` | 上游供应商余额，与组织预算轴无关 |
| `usage_ledger` 的 `input_tokens` / `output_tokens` | 用量计数，不是金额额度 |
| `reserved_pool` | 预留池配置，不是 consumed |

算法与入账顺序见 [Backend-预算.md](./Backend-预算.md) §1–§2。CNY 与 NewAPI quota、token 数的区别见 [Backend-计费单位.md](./Backend-计费单位.md)。
