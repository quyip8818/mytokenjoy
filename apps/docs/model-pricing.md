# 模型定价架构

> **状态：已实现** — 代码在 `apps/backend/internal/domain/pricing/`
>
> TokenJoy 作为定价 SOT。一张 `model_pricing` 表统一全局价、合同价、历史记录。
> NewAPI 降级为 gateway 执行层，接收 TJ 推送的 ratio 缓存。

---

## 1. 核心原则

| 原则 | 说明 |
|------|------|
| TJ 是价格 SOT | 所有定价决策在 TJ 完成并持久化，不再依赖外部系统存储价格 |
| 时间线即历史 | 每次改价 = INSERT 新行，旧行自动成为历史，无需独立 changelog |
| 全局/合同同构 | 同一张表、同一查询逻辑，用 `company_id` 区分层级 |
| NewAPI 只做缓存同步 | TJ 改价后 best-effort 推送到 NewAPI，保持 gateway 预扣行为一致 |
| 入账自主 | TJ 用自己的 pricing 重算 quota，不信任 raw.Quota |
| 不绑定 ratio 体系 | 存/算均用 元/1M tokens，不使用 RatioFromPrice 做入账计算 |

---

## 2. 数据模型

### 2.1 model_pricing 表

```sql
CREATE TABLE model_pricing (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id      UUID NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    model_type      VARCHAR(128) NOT NULL,
    input_price     NUMERIC(12,6) NOT NULL,   -- 元/1M tokens
    output_price    NUMERIC(12,6) NOT NULL,   -- 元/1M tokens
    effective_from  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    note            TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (company_id, model_type, effective_from)
);

CREATE INDEX idx_model_pricing_current
    ON model_pricing (company_id, model_type, effective_from DESC);
```

### 2.2 字段语义

| 字段 | 说明 |
|------|------|
| `company_id` | `TokenJoyCompanyID` = 全局价，其他 = 合同价（per-company） |
| `model_type` | 模型标识，如 `deepseek-chat`、`gpt-4o` |
| `input_price` | 输入 token 单价（元/1M tokens） |
| `output_price` | 输出 token 单价（元/1M tokens） |
| `effective_from` | 生效时间。改价时设为当前时间 |
| `note` | 可选备注（如"2025Q3 合同续签折扣"） |

**不需要的字段**：
- ~~`source`~~：`company_id == TokenJoyCompanyID` → 全局价，其他 → 合同价，天然可推导
- ~~`updated_at`~~：append-only 行不可变
- ~~`effective_until`~~：YAGNI，未来扩展点

---

## 3. 价格解析

### 3.1 优先级

```
解析 (company_id=X, model_type=M, at=T) 的生效价格：

  1. model_pricing WHERE company_id=X AND model_type=M AND effective_from <= T
     ORDER BY effective_from DESC LIMIT 1
     → 命中 = 合同价

  2. model_pricing WHERE company_id=TokenJoyCompanyID AND model_type=M AND effective_from <= T
     ORDER BY effective_from DESC LIMIT 1
     → fallback = 全局价

  3. 都没有 → 保留 raw.Quota + slog.Warn 告警
```

### 3.2 代码位置

- Repository: `internal/store/model_pricing.go` + `internal/store/postgres/model_pricing_repo.go`
- 批量预加载用 `DISTINCT ON (model_type)` + `ORDER BY effective_from DESC`
- **不提供 Delete**。合同终止 = 停止设定新合同价，入账自动 fallback 全局价

---

## 4. 入账流程

### 4.1 EntryBuildSnapshot 加载

`internal/domain/usage/entry_load.go`

```go
type EntryBuildSnapshot struct {
    Catalog        []types.ModelInfo
    OrgTree        []types.OrgNode
    CompanyPricing []store.ModelPricingRow  // 合同价
    GlobalPricing  []store.ModelPricingRow  // 全局价
    QuotaPerUnit   int64                   // 公司 billing QPU
}
```

`LoadEntryBuildSnapshot` 接收 `tokenJoyCompanyID` 参数，一次性批量加载当前时刻的全局价和合同价。

### 4.2 Quota 计算

`internal/domain/pricing/calc.go`

```go
func CalcQuota(inputTokens, outputTokens int64, inputPrice, outputPrice float64, quotaPerUnit int64) int64 {
    costYuan := (float64(inputTokens)*inputPrice + float64(outputTokens)*outputPrice) / 1_000_000
    return int64(math.Ceil(costYuan * float64(quotaPerUnit)))
}
```

### 4.3 入账覆盖

`internal/domain/usage/pricing_override.go`

```go
func ApplyTJPricing(entry, snap, tokenJoyCompanyID) → entry with:
  - QuotaAmount = CalcQuota(...)
  - CallDetail.InputPrice / OutputPrice / ContractPricing
```

调用位置：`IngestRaw` 中 `BuildCallSettledEntry` 之后、`ConsumeLots` 之前。

### 4.4 call_detail 记录实际价格

```go
type UsageCallDetail struct {
    // ...existing...
    InputPrice      float64 `json:"inputPrice,omitempty"`
    OutputPrice     float64 `json:"outputPrice,omitempty"`
    ContractPricing bool    `json:"contractPricing,omitempty"`
}
```

每笔消费记录入账时使用的实际价格，对账无需反查。

---

## 5. NewAPI 同步（降级为缓存推送）

### 5.1 写入方向

```
TJ model_pricing (SOT)
    │
    │  改价时 best-effort 推送
    ▼
NewAPI options (ModelRatio / CompletionRatio JSON map)
    │
    │  gateway 用于预扣/限速
    ▼
NewAPI logs (raw.Quota — TJ 不再信任此值)
```

### 5.2 触发推送的场景

| 场景 | 推送 NewAPI？ |
|------|:---:|
| SetGlobalPrice | ✅ |
| SetContractPrice | ❌ — gateway 无 per-company ratio 支持 |
| CreateModel (带价格) | ✅ |
| UpdateModel (改价格) | ✅ |
| CatalogSync (Local 拉取) | ✅ |
| FullSyncToNewAPI (定时) | ✅ |

### 5.3 失败容忍

推送失败不阻塞业务（`_ = client.UpsertModelRatio(...)` 或 `slog.Warn`）。
`FullSyncToNewAPI` 定时全量对齐兜底。

---

## 6. CatalogSync（Local 侧）

`internal/worker/catalogsync/execute.go`

```
SaaS 返回 CatalogModel[] 包含 inputPrice / outputPrice
Local catalogsync:
  1. 模型目录 → 写 models 表（SyncFromPlatform）
  2. 全局价 → 写 model_pricing (company_id=globalCompanyID, ON CONFLICT skip)
  3. 全局价 → 推送到本地 NewAPI（保持 gateway 一致）
```

---

## 7. API

### 7.1 Platform Admin（全局价）

```
GET    /api/platform/pricing                      → 所有模型当前全局价
PUT    /api/platform/pricing                      → 设定/更新全局价
GET    /api/platform/pricing/{modelType}/history   → 时间线
```

### 7.2 Platform Admin（合同价）

```
GET    /api/platform/companies/{id}/pricing                       → 某公司所有合同价
PUT    /api/platform/companies/{id}/pricing                       → 设定合同价
GET    /api/platform/companies/{id}/pricing/{modelType}/history   → 时间线
```

### 7.3 Platform Admin（per-model 快捷设价）

```
PUT    /api/platform/models/{id}/pricing          → 按模型 ID 设全局价
```

### 7.4 客户侧（只读）

`GET /api/models` → `ListModelsWithPricing` 从 model_pricing 读合同价 > 全局价填入响应。

---

## 8. 代码位置

| 模块 | 路径 |
|------|------|
| Store 定义 | `internal/store/model_pricing.go` |
| Postgres 实现 | `internal/store/postgres/model_pricing_repo.go` |
| Domain Service | `internal/domain/pricing/service.go` |
| CalcQuota | `internal/domain/pricing/calc.go` |
| 入账覆盖 | `internal/domain/usage/pricing_override.go` |
| Snapshot 加载 | `internal/domain/usage/entry_load.go` |
| HTTP Handler | `internal/http/handler/platform/pricing.go` |
| CatalogSync | `internal/worker/catalogsync/execute.go` |
| App wiring | `internal/app/compose_domain_wire.go` (`wirePricing`) |

---

## 9. 读取路径汇总

| 场景 | 数据来源 |
|------|----------|
| 模型列表 + 价格展示 | `models` 表 + `model_pricing` 当前行 |
| 入账计算 quota | `model_pricing` 预加载到 snapshot + `billing_currency.quota_per_unit` |
| 价格历史查询 | `model_pricing` WHERE ... ORDER BY effective_from |
| 对账验证 | `usage_ledger.call_detail.inputPrice/outputPrice`（元） |
| NewAPI gateway 行为 | TJ 推送过去的 ratio 缓存（`adminport.UpsertModelRatio`） |

---

## 10. 架构图

```
┌────────────────────────────────────────────────────────────────────────┐
│                          TokenJoy (SOT)                                 │
│                                                                        │
│  ┌──────────────┐      ┌──────────────────┐      ┌────────────────┐   │
│  │   models     │      │  model_pricing   │      │billing_currency│   │
│  │  (目录实体)   │      │  (价格时间线)     │      │(quota_per_unit)│   │
│  └──────┬───────┘      └────────┬─────────┘      └───────┬────────┘   │
│         │                       │                        │            │
│         │  ListModels()         │  price                 │ qpu        │
│         ▼                       ▼                        ▼            │
│  ┌─────────────────────────────────────────────────────────────┐      │
│  │                      IngestRaw                               │      │
│  │                                                             │      │
│  │  1. LoadEntryBuildSnapshot (catalog + pricing + qpu)        │      │
│  │  2. BuildCallSettledEntry (tokens from raw)                 │      │
│  │  3. ApplyTJPricing → CalcQuota (price×tokens÷1M×qpu)       │      │
│  │  4. ConsumeLotsLocked (quota)                               │      │
│  │  5. call_detail 记录实际 inputPrice/outputPrice              │      │
│  └─────────────────────────────────────────────────────────────┘      │
│         │                                                              │
│         │ best-effort push (RatioFromPrice → NewAPI format)            │
│         ▼                                                              │
│  ┌─────────────────────┐                                              │
│  │  NewAPI (cache)     │                                              │
│  │  ModelRatio map     │  ← TJ 推送，不再是 SOT                        │
│  │  CompletionRatio    │  ← gateway 预扣/限速用                        │
│  └─────────────────────┘                                              │
└────────────────────────────────────────────────────────────────────────┘
```

---

## 11. 设计决策记录

| 决策 | 理由 |
|------|------|
| 不在 models 表存价格 | 模型目录和定价策略变更轴不同，分离关注点 |
| 全局价和合同价同表 | key 结构相同，统一查询逻辑 |
| append-only 时间线 | 自带历史，无需独立 changelog 表 |
| 入账不信任 raw.Quota | TJ 自主计算，消除对 NewAPI 计费逻辑的依赖 |
| call_detail 存实际价格 | 对账零 JOIN，入账瞬间快照 |
| NewAPI 同步 best-effort | 推送失败不阻塞入账，定时任务兜底对齐 |
| 合同价不推 NewAPI | gateway 无 per-company ratio 支持，差异在 TJ 入账时体现 |
| 存 input_price/output_price 而非 ratio | 业务事实 > 计算派生；不绑定 NewAPI ratio 体系 |
| 不提供 Delete API | append-only 设计下物理删除是反模式 |
| CalcQuota 用 `math.Ceil` | 对高频小调用有 ±1 quota 偏差，对账预留容差 |

---

## 12. 未来扩展点

| 扩展 | 做法 |
|------|------|
| 阶梯定价 | 加 `tier_threshold` 字段 |
| 合同到期 | 加 `effective_until` 字段 |
| 折扣策略 | 加 `discount_pct` 字段或子表 |
| 审批流 | 加 `status` 字段（pending/approved） |
| 批量调价 | 服务端批量 INSERT |
| 价格预告 | `effective_from` 设为未来时间 |
