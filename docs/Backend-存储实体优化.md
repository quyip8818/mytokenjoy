# Backend 核心存储实体

**文档性质：** 架构参考  
**受众：** 后端  
**表数量：** **36** 张（以 `apps/backend/internal/store/postgres/schema.sql` 为准）

说明四张**合并型核心表**的 Schema、Repository 边界与读写约定。实体全景见 [Backend-存储架构.md](./Backend-存储架构.md)；入账契约见 [Backend-消耗数据SSOT对齐方案.md](./Backend-消耗数据SSOT对齐方案.md)；命名见 [Backend-命名规范.md](./Backend-命名规范.md)。

**部署约定：** 表结构只改 `schema.sql`，服务启动全量应用；本地改结构后 `docker compose down -v` 重建并由 seed 灌数。

---

## 1. 决策摘要

| 项             | 约定                                                                                   |
| -------------- | -------------------------------------------------------------------------------------- |
| **组织节点**   | `org_nodes` 一行 = 部门 + 预算 + 路由列                                                |
| **模型白名单** | `model_allowlist`（`owner_type` + `owner_id` + `model_name`）                          |
| **组织集成**   | `org_integration` 每企业一行（连接、同步策略、加密凭证）                               |
| **异步发件箱** | `outbox`（`channel` = `relay` \| `webhook`）                                           |
| **列名**       | 物理列 `department_id` 不改，语义为 `org_node_id`                                      |
| **路由 ID**    | `RoutingRule.id` = `nodeId` = `org_nodes.id`                                           |
| **Store 访问** | `Org().Nodes()`、`Org().Integration()`、`Models().Allowlist()`                         |
| **入账热路径** | `usage_ledger` + 同步 `projection.Apply`；`rebalance_queue` / `overrun_queue` 独立保留 |

---

## 2. `org_nodes`

组织树、预算树、节点路由配置合一；HTTP 仍投影为 `Department`、`BudgetNode`、`RoutingRule`。

```sql
CREATE TABLE org_nodes (
    id                TEXT NOT NULL,
    company_id        BIGINT NOT NULL DEFAULT 1 REFERENCES companies (id),
    name              TEXT NOT NULL,
    parent_id         TEXT,
    member_count      INT NOT NULL DEFAULT 0,
    external_id       TEXT,
    source            TEXT,
    manager_id        TEXT,
    sort_order        INT NOT NULL DEFAULT 0,
    budget            NUMERIC(18, 6) NOT NULL DEFAULT 0,
    consumed          NUMERIC(18, 6) NOT NULL DEFAULT 0,
    reserved_pool     NUMERIC(18, 6),
    period            TEXT NOT NULL,
    default_model     TEXT,
    fallback_model    TEXT,
    routing_inherited BOOLEAN NOT NULL DEFAULT FALSE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (company_id, id)
);
```

**引用同一节点 ID 的列（物理名 `department_id` / `node_id`）：**  
`members.department_id`、`usage_ledger.department_id`、`budget_group_departments.department_id`、`relay_mappings.department_id`、`usage_buckets.department_id`、`alert_rules.node_id`。

**组织树变更：** 单事务 `Org().Nodes().SetTree` + 受影响节点 `Models().Allowlist().Replace`（`domain/org/provision.go`）。

**路由语义：**

| 条件                                | 行为                                                                         |
| ----------------------------------- | ---------------------------------------------------------------------------- |
| 沿祖先链无 allowlist 且路由列为默认 | 返回全部 `enabled` 模型                                                      |
| 节点有 allowlist 或显式路由列       | 等同 `RoutingRule`；`routing_inherited = true` 时子集 intersect 父 allowlist |
| HTTP 投影                           | `id` / `nodeId` = `org_nodes.id`；`allowedModels` 来自 `model_allowlist`     |

**Repository：** `OrgNodeRepository` — `Tree`、`SetTree`、`RollupConsumed`、`GetNodeBudget`。

---

## 3. `model_allowlist`

平台密钥、组织节点路由、审批单共用一张白名单表。

```sql
CREATE TABLE model_allowlist (
    company_id   BIGINT NOT NULL DEFAULT 1,
    owner_type   TEXT NOT NULL,
    owner_id     TEXT NOT NULL,
    model_name   TEXT NOT NULL,
    PRIMARY KEY (company_id, owner_type, owner_id, model_name),
    CONSTRAINT chk_model_allowlist_owner_type
        CHECK (owner_type IN ('platform_key', 'org_node', 'key_approval'))
);
```

**生命周期：**

- `key_approval`：审批驳回或办结创建密钥后删除该 `owner_id` 下所有行。
- `platform_key` / `org_node`：随 owner 删除 prune。

**Repository：** `ModelAllowlistRepository` — `List`、`Replace`（按 `owner_type` + `owner_id`）。

---

## 4. `org_integration`

每企业一行：飞书等 HR 数据源连接、定时同步策略、加密凭证。

```sql
CREATE TABLE org_integration (
    company_id                   BIGINT PRIMARY KEY DEFAULT 1 REFERENCES companies (id),
    platform                     TEXT,
    connected                    BOOLEAN NOT NULL DEFAULT FALSE,
    last_import                  TIMESTAMPTZ,
    last_import_ok               INT,
    last_import_fail             INT,
    enabled                      BOOLEAN NOT NULL DEFAULT FALSE,
    start_time                   TEXT NOT NULL DEFAULT '',
    frequency_hours              INT NOT NULL DEFAULT 24,
    delete_member_threshold      INT NOT NULL DEFAULT 0,
    delete_department_threshold  INT NOT NULL DEFAULT 0,
    notify_phone                 BOOLEAN NOT NULL DEFAULT FALSE,
    notify_email                 BOOLEAN NOT NULL DEFAULT FALSE,
    notify_im                    BOOLEAN NOT NULL DEFAULT FALSE,
    encrypted_credential         BYTEA,
    updated_at                   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

- `encrypted_credential IS NULL` = 未保存凭证。
- `connected`（探测/导入连通）与 `enabled`（定时任务开关）语义独立。
- `org_sync_logs`、`org_import_failures` 仍为追加日志表。

HTTP 投影为 `DataSourceStatus` + `SyncConfig`；凭证经 `GetIntegrationCredential` / `SaveIntegrationCredential` / `ClearIntegrationCredential`。

---

## 5. `outbox`

Relay 与 Webhook 重试共用一张发件箱。

```sql
CREATE TABLE outbox (
    id           TEXT PRIMARY KEY,
    channel      TEXT NOT NULL,
    kind         TEXT,
    payload      JSONB NOT NULL,
    status       TEXT NOT NULL DEFAULT 'pending',
    attempts     INT NOT NULL DEFAULT 0,
    next_retry   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_error   TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_outbox_channel CHECK (channel IN ('relay', 'webhook'))
);
```

- `kind`：仅 `channel = 'relay'` 时必填。
- Worker 按 `channel` 各自 claim：`relay_processor`、`webhook_processor`。
- Store 层保留 `RelayOutboxEntry` / `WebhookOutboxEntry` 方法名，SQL 统一查 `outbox`。

**不**与 `rebalance_queue` / `overrun_queue` 合并（语义与 SLA 不同）。

---

## 6. Repository 边界

| Repository                     | 职责                                                                                       |
| ------------------------------ | ------------------------------------------------------------------------------------------ |
| **`OrgRepository`**            | `Members`、`Roles`、`Permissions`、`Integration()`、`Nodes()`                              |
| **`OrgNodeRepository`**        | `Tree`、`SetTree`、`RollupConsumed`、`GetNodeBudget`                                       |
| **`BudgetRepository`**         | `Groups`、`SetGroups`、`AddGroupConsumed`、`GetGroupBudget`、`OverrunPolicy`、`AlertRules` |
| **`ModelsRepository`**         | 模型目录 CRUD                                                                              |
| **`ModelAllowlistRepository`** | `List`、`Replace`                                                                          |
| **`KeysRepository`**           | 密钥 CRUD；白名单写改调 `Allowlist().Replace`                                              |

`Store` 接口不在顶层新增 `OrgNode()` 或 `Credential()`。

**Snapshot：** `OrgNodes` + `OrgIntegration` + 按需加载的 `model_allowlist`；不再分 `Departments` / `BudgetTree` / `RoutingRules` 三份。

---

## 7. 明确不合并

| 项                              | 原因                            |
| ------------------------------- | ------------------------------- |
| 用量桶 / 账本                   | 看板扫全表或 Ingest 与审计耦合  |
| 平台密钥 / Relay 映射           | 管理 CRUD 与 Relay 热路径争锁   |
| 预算组并入组织树                | 产品能力：跨部门池 + Key 直挂池 |
| 异步投影 / 投影 Outbox          | SSOT 已决策同步投影             |
| 本阶段重命名 `department_id` 列 | 牵涉面过大；语义在命名规范约定  |

---

## 8. 姊妹文档

| 文档                                                                 | 职责                  |
| -------------------------------------------------------------------- | --------------------- |
| [Backend-存储架构.md](./Backend-存储架构.md)                         | 36 表实体关系与读图   |
| [Backend-消耗数据SSOT对齐方案.md](./Backend-消耗数据SSOT对齐方案.md) | 入账写入契约          |
| [Backend-命名规范.md](./Backend-命名规范.md)                         | 跨层术语与 HTTP 投影  |
| [Frontend-API契约.md](./Frontend-API契约.md)                         | REST 路径与 JSON 形状 |
