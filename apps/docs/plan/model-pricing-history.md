# 模型变更记录 & 定价快照

> 统一记录模型所有变更（价格 / provider / capabilities / 上下架），SaaS + Local 各自写、不同步。
> 每笔消费在 call_detail 中快照实际 ratio（来自 patch 后的 NewAPI logs），用于对账。

---

## 1. 问题

- 作为类 OpenRouter 产品，用户需要知道模型发生了什么变化（价格调整、能力变更、provider 切换）
- 对账需要知道某笔消费用的是哪个价格
- 价格 SOT 在 NewAPI `options` 表（ModelRatio / CompletionRatio JSON map），但 NewAPI logs 不记录计费时用的 ratio

---

## 2. 设计

### 2.1 两层方案

| 层 | 做什么 | 作用 |
|----|--------|------|
| `model_changelog` 表 | 记录所有模型变更事件 | 审计 + 用户 changelog + 价格历史反查 |
| NewAPI logs patch + call_detail 透传 | 每笔消费记录实际 ratio | 对账（100% 精确） |

两层独立。changelog 面向运营/用户展示，call_detail ratio 面向计费对账。

### 2.2 SaaS / Local 职责

| 系统 | model_changelog | call_detail ratio |
|------|-----------------|-------------------|
| SaaS | 写（admin 调价/改模型时） | 写（消费 log 入账时） |
| Local | 写（catalogsync 检测到变更时 + 客户自定义模型变更时） | 写（消费 log 入账时） |

两侧 changelog 不同步，各记各的。

---

## 3. model_changelog 表

### 3.1 Schema

```sql
CREATE TABLE IF NOT EXISTS model_changelog (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_type  VARCHAR(128) NOT NULL,
    event_type  VARCHAR(32) NOT NULL,
    payload     JSONB NOT NULL DEFAULT '{}',
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_model_changelog_lookup
    ON model_changelog (model_type, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_model_changelog_feed
    ON model_changelog (occurred_at DESC);
```

### 3.2 Event Types

| event_type | payload 示例 | 触发时机 |
|------------|-------------|----------|
| `pricing_changed` | `{"inputPrice":10,"outputPrice":20,"prevInput":5,"prevOutput":10,"modelRatio":5,"completionRatio":2}` | SetModelPricing / catalogsync 检测到价格变更 |
| `capabilities_changed` | `{"added":["vision"],"removed":[]}` | UpdateModel / catalogsync 检测到 caps 变更 |
| `provider_changed` | `{"from":"azure","to":"openai"}` | UpdateModel / catalogsync 检测到 provider 变更 |
| `model_added` | `{"name":"GPT-4o","provider":"openai","capabilities":["chat","vision"]}` | CreateModel / catalogsync 新增模型 |
| `model_removed` | `{"name":"GPT-4o","reason":"deprecated"}` | DeleteModel / catalogsync 模型下架 |

### 3.3 写入点

统一收口到一个 helper，所有变更路径调用同一个函数：

```go
// internal/domain/models/changelog.go
package models

type ChangelogWriter interface {
    RecordChange(ctx context.Context, modelType, eventType string, payload any) error
}

// 调用示例（在 service 层，不散落到各 handler）
func (s *Service) SetPricing(ctx context.Context, modelType string, input, output float64) error {
    // ... UpsertModelRatio ...
    s.changelog.RecordChange(ctx, modelType, "pricing_changed", map[string]any{
        "inputPrice": input, "outputPrice": output,
        "modelRatio": ratio, "completionRatio": compRatio,
    })
    return nil
}
```

写入路径汇总：
- `SetModelPricing` → pricing_changed
- `CreateModel` → model_added（+ pricing_changed 如果带价格）
- `UpdateModel` → 比较 diff，按需写 provider_changed / capabilities_changed
- `DeleteModel` → model_removed
- `catalogsync` → diff 检测，按变更类型写入

### 3.4 查询接口

```go
// 用户侧：最近的模型变更日志（changelog 页面）
func RecentChanges(ctx context.Context, limit, offset int) ([]ChangelogEntry, error)

// 对账侧：某模型在某时刻的生效价格
func PricingAt(ctx context.Context, modelType string, at time.Time) (*PricingSnapshot, error)
// → SELECT payload FROM model_changelog
//   WHERE model_type=$1 AND event_type='pricing_changed' AND occurred_at <= $2
//   ORDER BY occurred_at DESC LIMIT 1
```

---

## 4. NewAPI Patch：logs 表加 ratio 字段

### 4.1 DDL

```sql
ALTER TABLE logs ADD COLUMN model_ratio NUMERIC(12,6);
ALTER TABLE logs ADD COLUMN completion_ratio NUMERIC(12,6);
```

### 4.2 改动点

NewAPI 计费逻辑中（consume log 写入处），已经有 `modelRatio` 和 `completionRatio` 在内存中用于计算 quota：

```
quota = prompt_tokens * model_ratio + completion_tokens * model_ratio * completion_ratio
```

只需在 INSERT INTO logs 时多写这两个值：

```sql
INSERT INTO logs (..., model_ratio, completion_ratio)
VALUES (..., $modelRatio, $completionRatio)
```

### 4.3 侵入性评估

- 改 1 个 DDL + 1 个 INSERT 语句
- 不影响 NewAPI 任何读取路径和 API 接口
- NewAPI 升级时 merge conflict 概率极低

---

## 5. TokenJoy 侧：从 logs 读 ratio → 写入 call_detail

### 5.1 RawConsumeLog 扩展

```go
type RawConsumeLog struct {
    // ... existing fields
    ModelRatio      float64
    CompletionRatio float64
}
```

### 5.2 SELECT 改动

```sql
SELECT id, token_id, quota, model_name, created_at,
       prompt_tokens, completion_tokens, use_time, content,
       model_ratio, completion_ratio  -- 新增
FROM logs WHERE ...
```

### 5.3 UsageCallDetail 扩展

```go
type UsageCallDetail struct {
    // ... existing fields
    ModelRatio      float64 `json:"modelRatio,omitempty"`
    CompletionRatio float64 `json:"completionRatio,omitempty"`
}
```

### 5.4 buildCallDetail 填充

```go
detail.ModelRatio = input.Raw.ModelRatio
detail.CompletionRatio = input.Raw.CompletionRatio
```

零额外查询，数据同源（NewAPI 算 quota 时用的那个值）。

---

## 6. 对账流程

```
对账某笔 usage_ledger 记录：
  1. 看 call_detail.modelRatio / completionRatio（patch 后的新数据必有）
  2. 兜底（patch 前的老数据）→ 查 model_changelog WHERE event_type='pricing_changed' AND occurred_at <= occurred_at

验证：
  expected_quota = prompt_tokens * model_ratio
                 + completion_tokens * model_ratio * completion_ratio
  actual_quota = usage_ledger.quota_amount
  差异应为 0
```

---

## 7. catalogsync 与 changelog 的交互（Local 侧）

```go
func (e *Executor) syncModelsWithChangelog(ctx context.Context, remote []CatalogModel) error {
    existing := loadCurrentModels()  // 从 DB 读当前状态

    for _, m := range remote {
        old, found := existing[m.ModelID]
        if !found {
            // 新增
            changelog.RecordChange(ctx, m.ModelID, "model_added", ...)
        } else {
            // Diff 检测
            if old.InputPrice != m.InputPrice || old.OutputPrice != m.OutputPrice {
                changelog.RecordChange(ctx, m.ModelID, "pricing_changed", ...)
            }
            if old.Provider != m.Provider {
                changelog.RecordChange(ctx, m.ModelID, "provider_changed", ...)
            }
            if !slices.Equal(old.Capabilities, m.Capabilities) {
                changelog.RecordChange(ctx, m.ModelID, "capabilities_changed", ...)
            }
        }
    }
    // 检测下架
    for id, old := range existing {
        if _, stillExists := remoteMap[id]; !stillExists {
            changelog.RecordChange(ctx, id, "model_removed", ...)
        }
    }
}
```

---

## 8. 实施步骤

| # | 系统 | 改动 | 优先级 |
|---|------|------|--------|
| 1 | NewAPI | DDL + INSERT 加 model_ratio/completion_ratio | P0（对账基础） |
| 2 | TokenJoy | RawConsumeLog + SELECT 加两字段 | P0 |
| 3 | TokenJoy | UsageCallDetail 加 ratio + buildCallDetail 填充 | P0 |
| 4 | TokenJoy | model_changelog 表 schema | P1 |
| 5 | TokenJoy | ChangelogWriter + SetModelPricing/CreateModel/UpdateModel/DeleteModel 写 changelog | P1 |
| 6 | TokenJoy | catalogsync diff 检测 + 写 changelog（Local 侧） | P1 |
| 7 | 前端 | 模型 changelog 展示页面（可选） | P2 |

P0 = 对账能力，P1 = 审计/changelog，P2 = 用户可见。

---

## 9. 注意事项

- 价格不存在 models 表，只在 NewAPI `options` 的 JSON map 中（参见 model-pricing-path.md）
- model_changelog 是 append-only，不需要更新/删除
- 换算公式：`inputPrice = modelRatio * 2`，`outputPrice = modelRatio * completionRatio * 2`
- changelog payload 中同时存 ratio 和 price（冗余），方便查询无需换算
- SaaS / Local 各自写 changelog，不同步——Local 的 changelog 反映的是"从 SaaS 拉取时检测到的变更"
- call_detail ratio 以 NewAPI logs 实际记录为准（对账唯一真相源）
