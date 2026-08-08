# 模型可见性架构

## 概念模型

```
                    ┌─────────────────────────────┐
                    │     模型目录（models 表）     │
                    │                             │
                    │  一条记录 = 一个可调用模型    │
                    │  不绑定供应商，不存价格       │
                    └──────────────┬──────────────┘
                                   │
              ┌────────────────────┼────────────────────┐
              │                    │                    │
              ▼                    ▼                    ▼
    Platform Admin 视角      SaaS 客户视角        Local 客户视角
    /platform/models         /models/list         /models/list
    (真实provider,可编辑)    (provider="tokenjoy") (托管+自管)
                                   │
                                   │ 请求 →
                                   ▼
              ┌─────────────────────────────────────────┐
              │        NewAPI Channels（上游路由）        │
              │                                         │
              │  同一 model type → N 个 channel          │
              │  按 priority + weight 分流到不同供应商    │
              │  仅 Platform Admin 通过 NewAPI 后台管理   │
              │  客户完全不可见                           │
              └─────────────────────────────────────────┘
```

---

## 三个视角

### SaaS Platform Admin（`/platform/models`，权限 `platform:manage`）

| 看到 | 能做 |
|------|------|
| 全部模型，provider 显示真实值（openai/qwen/volcengine...） | CRUD 模型目录 |
| 售价（从 NewAPI ratio 实时读） | 设定/修改统一售价 |
| — | 发布目录（bump version → Local 同步） |
| — | 在 NewAPI 后台管理上游渠道（channel CRUD，multi-provider 分流） |

Platform Admin 管理的模型：`company_id = TokenJoyCompanyID`
- `source = "platform"`, `provider` = 真实主供应商标识（内部参考）
- 价格 push 到 NewAPI ratio（SOT）
- 上游渠道在 NewAPI 独立管理，可多个 channel 指向同一 model type

### SaaS 客户（`/models/list`，权限 `model:read`）

| 看到 | 能做 |
|------|------|
| 仅 platform/seed 模型 | 配置路由白名单（按部门限制可用模型） |
| provider 统一显示 "tokenjoy" | — |
| 统一售价（含 markup） | — |
| 看不到上游渠道细节 | **不能**自建模型、不能管 Provider Key |

**数据来源**：`GET /api/models` → `ListModelsWithPricing` → `MaskProviderForTenant`（provider → "tokenjoy"）

> SaaS 模式下 CreateModel 被拒绝，models 表中不会有 custom 模型，sanitize 作为防御性兜底。

**页面逻辑**：前端 `useModelListPage` 中 `!isSelfHosted` 分支无 tab 切换，无创建按钮。

### Local 客户（`/models/list`，companyType = selfhosted）

| 看到 | 能做 |
|------|------|
| Tab 1 - 托管模型：从 SaaS 同步，provider="tokenjoy"，只读 | 路由白名单 |
| Tab 2 - 自管模型：自行接入，provider="custom" | CRUD（endpoint + apiKey） |
| 两类模型的价格 | 自管模型价格可编辑 |

**数据来源**：
- 托管模型：CatalogSync 每 10min 从 SaaS 拉取 → `source = "platform"`, 价格从 pricing 通道同步
- 自管模型：本地 `CreateModel` → `source = "manual"`, `provider = "custom"`

**自管渠道管理**：
- Provider Key 页面（`/keys/provider`）：创建供应商密钥 → 自动 sync 为 NewAPI channel
- 或创建 Custom Model 时内联 endpoint + apiKey → 独立 channel

---

## 数据模型

### models 表（简化，完整见 schema.sql）

```sql
models (
  model_id      UUID PRIMARY KEY,
  company_id    UUID NOT NULL,    -- TokenJoyCompanyID = platform 模型; 租户ID = custom 模型
  provider      TEXT NOT NULL,    -- "tokenjoy"(展示) | "custom"(自管) | 真实值(platform admin 内部)
  type          TEXT NOT NULL,    -- 模型标识（如 "qwen-plus"）
  name          TEXT NOT NULL,
  description   TEXT,
  source        TEXT NOT NULL DEFAULT 'manual',  -- "platform" | "manual" | "seed" | "test"
  endpoint      TEXT,             -- 仅 custom
  api_key       TEXT,             -- 仅 custom
  endpoint_model_name TEXT,       -- 仅 custom，上游实际模型名
  capabilities  TEXT[],
  max_context   INT,
  max_tokens    INT,
  deprecated    BOOL DEFAULT FALSE,
  catalog_synced_at TIMESTAMPTZ,  -- CatalogSync 最后同步时间
  updated_at    TIMESTAMPTZ,
  UNIQUE (company_id, provider, type)
  -- 价格不存 DB，从 NewAPI ratio 实时读取
)
```

> 唯一约束是 `(company_id, provider, type)`，非全局 type 唯一。
> 同一 type（如 "gpt-4o"）可以在同一 company 下存在 provider="tokenjoy" 和 provider="custom" 两条记录。

### 模型可见性过滤

```
Models(ctx)                      -- 按 company_id 查全部
  → FilterVisible()              -- 排除 source="test"
  → FilterNotDeprecated()        -- 排除 deprecated=true
= 用户可见的模型列表
```

CatalogSync 移除模型时标记 `deprecated=true`（软删除），不物理删除。

### 定价 SOT

```
NewAPI model_ratio 表 → 唯一定价真相
Platform Admin 设的是统一售价（对外价，含 markup）
背后多个渠道的各自成本 → 运营自行核算，系统不管
折扣 → model_discount 表（Ingest 时乘系数，不影响展示价）
```

### NewAPI Channel（上游路由，TokenJoy 不管理其 schema）

一个 model type 可匹配多个 channel → NewAPI 按 group + priority + weight 选择上游。
Platform Admin 直接在 NewAPI 后台管理 channel，TokenJoy 不封装。

---

## 关键场景：同一模型多上游

```
models 表:
  type="seedance-1.0", name="Seedance 1.0", price=¥3.0/Mtok

NewAPI channels:
  channel-101: 火山引擎  models="seedance-1.0" weight=70
  channel-102: 合作伙伴X models="seedance-1.0" weight=30

客户感知:
  一个模型，一个价格，不知道背后有几家供应商。
```

---

## Provider 脱敏机制

```go
// support/modelcatalog/mask_provider.go
const DisplayProvider = "tokenjoy"

func MaskProviderForTenant(models []types.ModelInfo) {
    for i := range models {
        if models[i].Source == "platform" || models[i].Source == "seed" {
            models[i].Provider = DisplayProvider
        }
    }
}
```

调用点：`handler/models/handler.go` 的 `List` 端点（企业用户 API）。
不调用：Platform Admin API、CatalogSync 内部、Gateway precheck。

---

## 各路径不受影响的理由

| 路径 | 用什么匹配 | provider 字段 |
|------|-----------|--------------|
| Gateway 模型白名单 | modelID (UUID) | 不涉及 |
| NewAPI channel 路由 | channel.models 字段 (modelType 字符串匹配) | 不涉及 |
| Ingest 计价 | modelType → NewAPI ratio | 不涉及 |
| Call log 展示 | modelType | 不涉及 |
| CatalogSync 同步 | modelType + provider（内部存储，不展示） | 保留真实值 |

---

## SaaS ↔ Local 模型同步

```
SaaS (Platform Admin 发布)
  │
  │  GET /api/platform/sync/catalog/models  (公开，无需鉴权)
  │  GET /api/platform/sync/catalog/pricing (需 sync token)
  │
  ▼
Local (CatalogSync worker, 每 10min)
  → upsert models 表 (source="platform", provider=真实值保留)
  → push pricing 到本地 NewAPI ratio
  → List API 响应时 MaskProviderForTenant → "tokenjoy"
```

---

## 约束

1. SaaS 客户不能创建模型 — `CreateModel` 在 `SupportSaas=true` 时返回 Forbidden ✅ 已实现
2. 模型不绑定供应商 — 供应商是 channel 层的事
3. 定价只有一个 — 不管几家上游，对外只有一个售价
4. Channel 管理不封装到 TokenJoy — Platform Admin 直接用 NewAPI 后台，不造轮子
5. Provider 脱敏在 API 层 — 一个函数，一行调用 ✅ 已实现
6. DB 保留真实 provider — 运营需要按供应商维度看模型，sanitize 只在 API 响应层

---

## NewAPI Group 策略

不做部门级渠道隔离。模型可见性由 TokenJoy Gateway precheck 的路由白名单控制，不靠 NewAPI group。

```
tokenjoy-upstream channel:    group = ""            ← 所有 token 都能访问 platform 模型
provider-key channels:        group = {companyId}   ← 仅本公司 token 能访问
所有员工 tokens:              group = {companyId}   ← 匹配 custom channels + 能访问 group="" 的 channel
```

SaaS 和 Local 模式策略完全相同，`CompanyChannelPolicy` 统一返回 `companyID.String()`。

**效果**：
- 员工 token (group={companyId}) → 能访问 group="" 的 channel（tokenjoy-upstream）+ group={companyId} 的 channel（自管渠道）
- 同一个 key 同时可调 platform 模型和 custom 模型
- NewAPI group 匹配规则：token 能访问同 group 或无 group（""）的 channel
- 为将来 SaaS per-company custom 模型预留了隔离能力（不同公司 companyId 不同 → channel 互不可见）
- SaaS 侧 register-local 创建的总 key 也使用 group={companyId}，统一规则

---

## 对比总结

同一个 seedance-1.0 模型在各处的呈现：

| 层 | provider | 价格 | 渠道信息 | 可编辑 |
|----|----------|------|---------|--------|
| SaaS NewAPI channels | volcengine + openai_compatible | — | weight 70/30 分流 | Platform Admin via NewAPI |
| SaaS NewAPI ratio | — | ratio=1.5 / completion=6.0 | — | Platform Admin |
| Platform Admin API | "volcengine" | ¥3.0 / ¥12.0 | — | ✅ CRUD |
| SaaS 客户 API | "tokenjoy" | ¥3.0 / ¥12.0 | 不可见 | ❌ 只读 |
| Local NewAPI channels | tokenjoy-upstream (透传到 SaaS) | — | — | — |
| Local NewAPI ratio | — | ratio=1.5 / completion=6.0 | — | — (sync 写入) |
| Local 客户 API | "tokenjoy" | ¥3.0 / ¥12.0 | 不可见 | ❌ 只读 |


---

## 附录：完整数据示例

以 Seedance 1.0 和 GPT-4o 两个模型为例，展示同一份数据在各端的完整形态。

### A. SaaS 侧 NewAPI 数据

#### Channels（SaaS NewAPI 后台）

```
┌────┬──────────────────────┬────────────────────┬──────────────────────────────┬──────────┬────────┬────────┐
│ ID │ Name                 │ Type (协议适配器)   │ Models                       │ Group    │ Pri    │ Weight │
├────┼──────────────────────┼────────────────────┼──────────────────────────────┼──────────┼────────┼────────┤
│ 1  │ openai-primary       │ openai             │ gpt-4o,gpt-4o-mini           │ (空)     │ 1      │ 100    │
│ 2  │ volcengine-seedance  │ volcengine         │ seedance-1.0                 │ (空)     │ 1      │ 70     │
│ 3  │ partner-x-seedance   │ openai_compatible  │ seedance-1.0                 │ (空)     │ 1      │ 30     │
└────┴──────────────────────┴────────────────────┴──────────────────────────────┴──────────┴────────┴────────┘
```

> Type 是 NewAPI 内部的协议适配器标识（openai/anthropic/qwen/volcengine 等），实际存储为 int。
> Group 为空 = 所有 token 都能访问。Platform channel 不限制 group。

#### Model Ratio（SaaS NewAPI，定价 SOT）

```
┌────────────────┬─────────────┬──────────────────┬─────────────┐
│ ModelName      │ ModelRatio  │ CompletionRatio  │ CacheRatio  │
├────────────────┼─────────────┼──────────────────┼─────────────┤
│ gpt-4o         │ 1.25        │ 5.0              │ 0.625       │
│ seedance-1.0   │ 1.5         │ 6.0              │ 0           │
└────────────────┴─────────────┴──────────────────┴─────────────┘
```

> ratio 换算公式：`displayPrice = ratio × 2`（基准 ¥2/Mtok），即 gpt-4o input=¥2.5, output=¥10.0

#### Tokens（SaaS NewAPI，per-company）

```
┌──────┬─────────────────────────┬───────────────────────────────────────────┬───────────────┐
│ ID   │ Name                    │ Group                                     │ RemainQuota   │
├──────┼─────────────────────────┼───────────────────────────────────────────┼───────────────┤
│ 101  │ company-A-总key          │ 00000000-0000-7000-8000-00000000000a      │ unlimited     │
│ 102  │ company-B-总key          │ 00000000-0000-7000-8000-00000000000b      │ unlimited     │
└──────┴─────────────────────────┴───────────────────────────────────────────┴───────────────┘
```

> 每个 Local 注册获得的总 key。group={companyId} 能访问 group="" 的 channel。

---

### B. Local 侧 NewAPI 数据

#### Channels（Local NewAPI）

```
┌────┬──────────────────────┬────────────────────┬──────────────────────────────┬───────────────────────────────────────────┬────────┬────────┐
│ ID │ Name                 │ Type (协议适配器)   │ Models                       │ Group                                     │ Pri    │ Weight │
├────┼──────────────────────┼────────────────────┼──────────────────────────────┼───────────────────────────────────────────┼────────┼────────┤
│ 1  │ tokenjoy-upstream    │ openai             │ (空=全部)                     │ (空)                                      │ 1      │ 100    │
│ 2  │ siliconflow-custom   │ openai_compatible  │ siliconflow/qwen-72b         │ 00000000-0000-7000-8000-000000000002      │ 1      │ 100    │
└────┴──────────────────────┴────────────────────┴──────────────────────────────┴───────────────────────────────────────────┴────────┴────────┘
```

- `tokenjoy-upstream`：group="" → 所有 token 都能访问，指向 SaaS Gateway，承载所有 platform 模型
- `siliconflow-custom`：group={companyId} → 本公司员工 token 能访问，客户自建 Provider Key 自动 sync 的 channel

#### Model Ratio（Local NewAPI，从 SaaS pricing sync 写入）

```
┌────────────────────────┬─────────────┬──────────────────┬─────────────┐
│ ModelName              │ ModelRatio  │ CompletionRatio  │ CacheRatio  │
├────────────────────────┼─────────────┼──────────────────┼─────────────┤
│ gpt-4o                 │ 1.25        │ 5.0              │ 0.625       │
│ seedance-1.0           │ 1.5         │ 6.0              │ 0           │
│ siliconflow/qwen-72b   │ 0.25        │ 1.0              │ 0           │
└────────────────────────┴─────────────┴──────────────────┴─────────────┘
```

- 前两行：CatalogSync pricing 通道同步写入
- 第三行：客户 CreateModel 时 push 的本地价格

#### Tokens（Local NewAPI，per-employee）

```
┌──────┬─────────────────────────┬───────────────────────────────────────────┬───────────────┐
│ ID   │ Name                    │ Group                                     │ RemainQuota   │
├──────┼─────────────────────────┼───────────────────────────────────────────┼───────────────┤
│ 201  │ tokenjoy:plk-员工A       │ 00000000-0000-7000-8000-000000000002      │ unlimited     │
│ 202  │ tokenjoy:plk-员工B       │ 00000000-0000-7000-8000-000000000002      │ unlimited     │
└──────┴─────────────────────────┴───────────────────────────────────────────┴───────────────┘
```

> 所有员工 token group={companyId} → 能访问 group="" 的 tokenjoy-upstream + group={companyId} 的 custom channels。
> 同一个 key 可以同时调 platform 模型（走 tokenjoy-upstream）和 custom 模型（走 siliconflow-custom）。

---

### C. Platform Admin API 返回

`GET /api/platform/models`（SaaS，需 `platform:manage` 权限）

```json
[
  {
    "modelId": "550e8400-e29b-41d4-a716-446655440001",
    "provider": "openai",
    "type": "gpt-4o",
    "name": "GPT-4o",
    "source": "platform",
    "deprecated": false,
    "capabilities": ["chat", "vision"],
    "maxContext": 128000,
    "inputPrice": 2.5,
    "outputPrice": 10.0,
    "cacheInputPrice": 1.25
  },
  {
    "modelId": "550e8400-e29b-41d4-a716-446655440002",
    "provider": "volcengine",
    "type": "seedance-1.0",
    "name": "Seedance 1.0",
    "source": "platform",
    "deprecated": false,
    "capabilities": ["video"],
    "maxContext": 32000,
    "inputPrice": 3.0,
    "outputPrice": 12.0,
    "cacheInputPrice": 0
  }
]
```

> provider = 真实供应商（内部参考），价格 = 对外售价（从 NewAPI ratio 实时 merge）

---

### D. SaaS 客户 API 返回

`GET /api/models`（SaaS，需 `model:read` 权限，标准/试用公司）

```json
[
  {
    "modelId": "550e8400-e29b-41d4-a716-446655440001",
    "provider": "tokenjoy",
    "type": "gpt-4o",
    "name": "GPT-4o",
    "deprecated": false,
    "capabilities": ["chat", "vision"],
    "maxContext": 128000,
    "inputPrice": 2.5,
    "outputPrice": 10.0,
    "cacheInputPrice": 1.25
  },
  {
    "modelId": "550e8400-e29b-41d4-a716-446655440002",
    "provider": "tokenjoy",
    "type": "seedance-1.0",
    "name": "Seedance 1.0",
    "deprecated": false,
    "capabilities": ["video"],
    "maxContext": 32000,
    "inputPrice": 3.0,
    "outputPrice": 12.0,
    "cacheInputPrice": 0
  }
]
```

> 与 Platform Admin 对比：provider 全部变为 "tokenjoy"，其余完全相同。
> 客户不知道 gpt-4o 走 OpenAI、seedance 走火山引擎+合作伙伴分流。

---

### E. Local 客户 API 返回

`GET /api/models`（Local，需 `model:read` 权限，selfhosted 公司）

```json
[
  {
    "modelId": "660e8400-e29b-41d4-a716-446655440001",
    "provider": "tokenjoy",
    "type": "gpt-4o",
    "name": "GPT-4o",
    "source": "platform",
    "deprecated": false,
    "capabilities": ["chat", "vision"],
    "maxContext": 128000,
    "inputPrice": 2.5,
    "outputPrice": 10.0,
    "cacheInputPrice": 1.25
  },
  {
    "modelId": "660e8400-e29b-41d4-a716-446655440002",
    "provider": "tokenjoy",
    "type": "seedance-1.0",
    "name": "Seedance 1.0",
    "source": "platform",
    "deprecated": false,
    "capabilities": ["video"],
    "maxContext": 32000,
    "inputPrice": 3.0,
    "outputPrice": 12.0,
    "cacheInputPrice": 0
  },
  {
    "modelId": "770e8400-e29b-41d4-a716-446655440003",
    "provider": "custom",
    "type": "siliconflow/qwen-72b",
    "name": "硅基流动 Qwen 72B",
    "source": "manual",
    "deprecated": false,
    "capabilities": ["chat"],
    "maxContext": 32000,
    "endpoint": "https://api.siliconflow.cn/v1",
    "inputPrice": 0.5,
    "outputPrice": 2.0,
    "cacheInputPrice": 0
  }
]
```

> 前两条：CatalogSync 同步来的托管模型，provider="tokenjoy"，只读。
> 第三条：客户自建的自管模型，provider="custom"，可编辑/删除。
