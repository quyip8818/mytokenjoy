# 全局模型 Catalog + Per-Company 定价覆盖

> 本文档描述 TokenJoy 模型 catalog 的现有架构，以及规划中的 per-company 定价覆盖方案。

---

## 1. 现有架构

### 1.1 数据模型

```sql
CREATE TABLE models (
    model_id     UUID PRIMARY KEY,
    company_id   UUID NOT NULL REFERENCES companies(id),
    provider     TEXT NOT NULL,          -- "deepseek" / "openai" / "custom"
    type         TEXT NOT NULL,          -- 模型标识符 (如 "gpt-4o")
    name         TEXT NOT NULL,          -- 展示名
    description  TEXT DEFAULT '',
    endpoint     TEXT,                   -- 自定义模型的 API 地址
    api_key      TEXT,                   -- 自定义模型的密钥
    endpoint_model_name TEXT,            -- 转发给上游时用的模型名
    max_context  INT DEFAULT 0,
    max_tokens   INT DEFAULT 0,
    active       BOOLEAN DEFAULT TRUE,   -- 是否可用
    capabilities TEXT[] DEFAULT '{}',
    source       TEXT DEFAULT 'manual',  -- 'sms' | 'manual' | 'seed' | 'test'
    sms_synced_at TIMESTAMPTZ,           -- SMS 同步时间戳（用于 diff-disable）
    UNIQUE (company_id, provider, type)
);
```

### 1.2 全局 Catalog 语义

单表多租户，通过 `company_id` 区分全局和租户模型：

| company_id | source | 含义 |
|---|---|---|
| globalCompanyID (= tokenJoyCompanyID) | `sms` | SMS 同步的全局模型目录 |
| 租户 ID | `manual` | 租户自行添加的第三方模型（provider='custom'） |
| 租户 ID | `manual` | 租户对全局模型的覆盖行（ToggleModel 创建） |

查询层自动合并全局 + 租户：

```go
// Models() — 查询全局 + 当前租户，DedupeEffective 以 provider+type 为 key 后出现的覆盖前面的
SELECT ... FROM models
WHERE company_id = $globalCompanyID OR company_id = $tenantCompanyID
ORDER BY CASE WHEN company_id = $globalCompanyID THEN 0 ELSE 1 END, model_id

→ DedupeEffective() 让租户行覆盖同 key 的全局行
```

### 1.3 SMS 同步（已完成全局写入）

SMS 同步直接写入 `company_id = globalCompanyID`，不循环 company：

```go
// SyncTarget 接口
type SyncTarget interface {
    ReplaceChannels(ctx context.Context, channels []sms.CatalogChannel) error
    SyncModels(ctx context.Context, models []sms.CatalogModel) error       // 全局写入
    ReplaceModelRatios(ctx context.Context, models []sms.CatalogModel) error // 全局定价
}

// Target 实现
type Target struct {
    port            adminport.Port
    store           store.Store
    globalCompanyID uuid.UUID
}

// SyncModels: upsert + diff-disable（用 sms_synced_at 时间戳判断过期）
func (t *Target) SyncModels(ctx, models) → store.Models().SyncFromSMS(ctx, globalCompanyID, infos)
```

`SyncFromSMS` 实现：
1. 获取 batch 时间戳：`SELECT NOW()` → `$batchTS`
2. 逐条 UPSERT：`ON CONFLICT (company_id, provider, type) DO UPDATE SET active=TRUE, sms_synced_at=$batchTS`
3. 标记过期：`UPDATE models SET active=FALSE WHERE source='sms' AND (sms_synced_at IS NULL OR sms_synced_at < $batchTS)`

> 注：使用 batch timestamp 做确定性判断，不依赖墙钟偏移。

新 company 注册即可看到全局模型，无需等同步。

### 1.4 租户覆盖（ToggleModel）

```go
// 全局模型 disable/enable：创建租户级副本
if model.CompanyID == cfg.TokenJoyCompanyID {
    override := *model
    override.Active = enabled
    store.Models().InsertModel(ctx, override) // 租户行，source='manual'
}
// DedupeEffective 自动用租户行覆盖全局行
```

保护机制：全局模型 `active=false`（SMS 下架）时，不允许租户 re-enable。

### 1.5 租户自定义第三方模型

```go
// CreateModel: provider 强制为 "custom"
// 参数：type, name, baseUrl, apiKey, endpointModelName, inputPrice, outputPrice, maxContext, maxTokens, capabilities
// 定价写入 NewAPI ModelRatio（全局 map）
```

### 1.6 定价链路（现状）

```
┌──────────────────────────────────────────────────────────────────────┐
│ SMS 运维设定售价 → UpsertModelRatio → NewAPI 全局 ModelRatio map      │
│                                                                      │
│ 用户请求 → NewAPI 网关 → 按全局 ModelRatio 计算 quota → 写入 logs 表  │
│                                                                      │
│ TokenJoy ingest → 读 logs.quota (已是最终值) → ConsumeLots 扣额度    │
│                    ↕                                                  │
│          QuotaAmount = Raw.Quota (直接 pass-through)                 │
│          spend = Σ segment.Cost (从 lot.QuotaPerUnit 转回货币金额)    │
│          budget_consumed += spend                                     │
└──────────────────────────────────────────────────────────────────────┘
```

关键：**定价发生在 NewAPI 层，全局唯一**。TokenJoy 不做二次定价。

---

## 2. 问题

当前所有 company 的模型价格完全一致（NewAPI 全局 ModelRatio）。

SaaS 场景需要为不同客户设定差异化售价（大客户折扣、战略客户优惠等），SMS 运维希望能为指定 company 覆盖模型价格。

---

## 3. 方案：Per-Company 定价覆盖

### 3.1 设计思路

- NewAPI 全局定价不变（控制网关侧 quota 额度限制）
- TokenJoy ingest 时，如果 company 有局部定价覆盖，用 `prompt_tokens × company_input_price + completion_tokens × company_output_price` **重算 quota**
- 无覆盖时直接用 `logs.quota`（向后兼容）
- SMS 侧新增"公司定价"管理功能，同步到 TokenJoy

### 3.2 新表

```sql
CREATE TABLE company_model_pricing (
    company_id   UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    model_type   TEXT NOT NULL,             -- 对应 models.type（也是 NewAPI logs.model_name）
    input_price  NUMERIC(12,6) NOT NULL,    -- 元/1M tokens
    output_price NUMERIC(12,6) NOT NULL,    -- 元/1M tokens
    source       TEXT NOT NULL DEFAULT 'sms',
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (company_id, model_type)
);

CREATE INDEX idx_company_model_pricing_company ON company_model_pricing (company_id);
```

> **约束说明**：PK 不含 `provider`。理由：在 NewAPI 日志中 `model_name` 全局唯一（同一 `type` 不会属于多个 provider），定价以 `model_type` 为 key 与 ingest 消费链路对齐。如果未来 provider 间出现 type 冲突，需要升级此表加 `provider` 列。

### 3.3 SMS 侧变化

SMS 新增同步分区 `company_pricing`：

```
GET /api/sync/versions → { channels, models, currencies, companyPricing }
GET /api/sync/catalog/company_pricing → {
    version: int,
    data: [{ companyId: UUID, modelType: string, inputPrice: float64, outputPrice: float64 }]
}
```

SMS `sync_versions` 表加一行 `('company_pricing', 0)`。
SMS 运维在"客户管理"或"模型管理"界面设定某客户的模型售价覆盖。

### 3.4 TokenJoy 同步变化

execute.go 加一个分区同步：

```go
const keyCompanyPricingVersion = "sms_sync.company_pricing_version"

func (e *SMSSyncExecutor) syncCompanyPricing(ctx context.Context) (int, error) {
    resp, err := e.client.FetchCompanyPricing(ctx)
    if err != nil { return 0, err }
    if err := e.target.ReplaceCompanyPricing(ctx, resp.Data); err != nil {
        return 0, fmt.Errorf("replace company pricing: %w", err)
    }
    return resp.Version, nil
}
```

SyncTarget 新增：

```go
type SyncTarget interface {
    // ...existing...
    ReplaceCompanyPricing(ctx context.Context, entries []sms.CompanyPricingEntry) error
}
```

实现：全量替换 `company_model_pricing` 表。

> **原子性策略**：使用事务内 `DELETE WHERE source='sms' + batch INSERT`（与 SyncFromSMS 的 batch timestamp 模式一致）。不用 TRUNCATE 以避免影响非 SMS 来源的定价行（预留手动覆盖能力）。在事务提交前 ingest 读到的是旧数据，提交后读到新数据，不会出现空表窗口。

### 3.5 Ingest 层变化

`BuildCallSettledEntry` 中加入 per-company 定价覆盖逻辑：

```go
// entry_build.go — BuildCallSettledEntry

quotaAmount := input.Raw.Quota // 默认用 NewAPI 算的

// 查 company 级定价覆盖
if pricing := input.CompanyPricing; pricing != nil {
    modelName := ResolveConsumeModel(input.Raw)
    if p, ok := pricing[modelName]; ok {
        // 用 company 级单价重算 quota
        inputCost := float64(input.Raw.PromptTokens) / 1_000_000 * p.InputPrice
        outputCost := float64(input.Raw.CompletionTokens) / 1_000_000 * p.OutputPrice
        totalCost := inputCost + outputCost        // 元
        // quotaPerUnit 使用 common.DefaultQuotaPerUnit (500000)
        // 与 NewAPI 内部 QuotaPerUnit 对齐，确保 quota ↔ 金额换算一致
        quotaAmount = common.MoneyToQuota(totalCost, common.DefaultQuotaPerUnit)
    }
}

entry.QuotaAmount = quotaAmount
```

> **quotaPerUnit 来源**：固定使用 `common.DefaultQuotaPerUnit`（1 CNY = 500000 quota），这与 NewAPI 的 QuotaPerUnit 对齐。不从 lot 获取——因为 ingest 阶段尚未确定消费到哪个 lot，lot 的 QPU 仅在 ConsumeLots 时用于计算 segment.Cost。

`EntryBuildInput` 加字段：

```go
type EntryBuildInput struct {
    // ...existing...
    CompanyPricing map[string]CompanyModelPrice // model_type → price
}

type CompanyModelPrice struct {
    InputPrice  float64 // 元/1M tokens
    OutputPrice float64
}
```

#### 模型名称匹配规则

`pricing[modelName]` 的 `modelName` 来自 `ResolveConsumeModel(input.Raw)` 即 NewAPI 日志的 `model_name`。`company_model_pricing.model_type` 必须与之一致。

**SMS 侧保证**：SMS 同步定价时，`model_type` 字段使用与 `CatalogModel.ModelID` 相同的值（即 NewAPI 路由时使用的模型标识符）。这确保 ingest 时能命中覆盖。

**miss 处理**：如果 `pricing[modelName]` 未命中，静默 fallback 到 `Raw.Quota`（全局定价），不报错。运维可通过对比 usage 记录中的 model 字段和 pricing 表排查配置遗漏。

#### 加载策略

`LoadEntryBuildSnapshot` 中按 `store.CompanyID(ctx)` 查 `company_model_pricing` 表，结果放入 `EntryBuildSnapshot.CompanyPricing`。Snapshot 是 per-company 的（context 已绑定 companyID），每个 ingest batch 加载一次，batch 内复用。

### 3.6 Store 层

```go
type CompanyModelPricingRepository interface {
    // ListByCompany 返回该公司所有模型的定价覆盖
    ListByCompany(ctx context.Context, companyID uuid.UUID) ([]CompanyModelPrice, error)
    // ReplaceAll 全量替换某公司的定价（同步用）
    ReplaceAll(ctx context.Context, entries []CompanyModelPriceRow) error
    // ReplaceGlobal 全量替换所有公司的定价（SMS 整体同步用）
    ReplaceGlobal(ctx context.Context, entries []CompanyModelPriceRow) error
}
```

---

## 4. 计费精度分析

### NewAPI quota 和 company 定价的关系

| 环节 | 用谁的价格 | 影响 |
|------|----------|------|
| NewAPI quota limit (限流) | 全局 ModelRatio | 控制"能不能调"——粗粒度 |
| TokenJoy ingest (扣款) | company 定价（有覆盖时） | 控制"扣多少钱"——精确 |
| 预算消耗 (budget_consumed) | spend = Σ segment.Cost | 跟随 ingest 后的 quota |

偏差场景：
- company 价格 < 全局 → NewAPI 认为用了更多 quota，可能提前触发 quota limit → 实际钱还没花完。影响小（quota limit 是"松上限"，overdraft 兜底）
- company 价格 > 全局 → NewAPI 允许了但实际扣款更多 → 可能超预算。已有 overdraft lot 机制吸收
- 部分覆盖 → 同一 company 内，有覆盖的模型走 per-company 价格，无覆盖的走 NewAPI raw quota。如果全局价格调整但 company 覆盖未同步更新，可能出现"覆盖的模型反而比新全局价格贵"。这是运营配置问题，非系统 bug——SMS 界面应提示运维"该客户覆盖价格高于当前全局价格"。

结论：偏差可控，不需要改 NewAPI 侧。

---

## 5. 实现步骤

| # | 事项 | 影响面 |
|---|------|--------|
| 1 | 建 `company_model_pricing` 表 | schema.sql |
| 2 | 实现 `CompanyModelPricingRepository` | store 层 |
| 3 | SMS 侧加 `company_pricing` 分区 + 管理界面 | sms/backend |
| 4 | TokenJoy `sms.Client` 加 `FetchCompanyPricing` | integration/sms |
| 5 | `SyncTarget` 加 `ReplaceCompanyPricing` | worker/smssync |
| 6 | `execute.go` 加 `syncCompanyPricing` 分区 | worker/smssync |
| 7 | `EntryBuildInput` 加 `CompanyPricing` 字段 | domain/usage |
| 8 | `LoadEntryBuildSnapshot` 加载公司定价 | domain/usage |
| 9 | `BuildCallSettledEntry` 中加覆盖逻辑 | domain/usage |
| 10 | 前端模型列表显示公司级价格（如果有覆盖） | frontend |

**步骤 10 说明**：现有 `ListModelsWithPricing` API 从 NewAPI 获取全局价格。改为：后端查 `company_model_pricing` 表，有覆盖的模型用 company 价格，无覆盖的仍返回全局价格。接口响应加 `priceSource: "company" | "global"` 字段让前端区分显示。不新增 API。

---

## 6. 对现有系统的影响

| 模块 | 影响 |
|------|------|
| 计费 (lot/consume) | 无直接改动——ingest 传入的 quotaAmount 变了，下游自动跟随 |
| 预算 (budget) | 无直接改动——spend 跟随 segment.Cost |
| NewAPI 路由 | 无——全局 ModelRatio 不变 |
| 前端模型选择器 | 可选：显示公司级价格代替全局价格 |
| model_allowlist | 无——定价独立于路由控制 |
| 网关 precheck | 无——precheck 不涉及定价 |
| ToggleModel | 无——定价覆盖和模型 active 状态正交 |
| SMS 模型同步 | 无——全局同步路径不变 |

---

## 7. 不做什么

- 不改 NewAPI 的 ModelRatio 机制——全局定价仍由 SMS 统一管控
- 不做 per-company 的 NewAPI quota limit 联动——偏差由 overdraft 吸收
- 不做实时定价变更生效——同步周期内用旧价格，下次同步后生效
- 不做定价版本历史——YAGNI
- 不做折扣率字段——直接存绝对价格更简单，折扣是 SMS 界面计算的展示逻辑
- 不做模型名称自动规范化——SMS 保证 `model_type` 和 NewAPI 路由标识一致，miss 时 fallback 全局价格
