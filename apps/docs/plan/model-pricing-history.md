# 模型定价历史记录

> SMS 侧记录完整定价变更历史（审计），TokenJoy 侧在 usage_ledger.call_detail 中快照当笔消费时的价格（对账）。

**前置依赖**：`sms-model-sync-v2.md` 完成（SMS models 表由 sync 维护后，trigger 才能捕获定价变更）。

---

## 1. 问题

- 模型定价会随 SMS sync 周期性变更
- TokenJoy usage_ledger 只记录了 quota_amount / tokens，没有记录当时生效的 model_ratio
- 对账时无法回推某笔消费用的是哪个价格

---

## 2. 方案

| 系统 | 策略 | 作用 |
|------|------|------|
| SMS | `model_pricing_history` 表 | 完整审计轨迹，可查任意时刻任意模型的生效价格 |
| NewAPI | logs 表加 `model_ratio` + `completion_ratio` | 记录计费时实际使用的倍率 |
| TokenJoy | `call_detail` JSONB 加 modelRatio/completionRatio | 单笔对账，无需 join 外部表 |

三者独立。SMS history 是运营审计视角，NewAPI log 是计算事实，TokenJoy call_detail 是消费快照。

---

## 3. SMS 侧：model_pricing_history 表

### 3.1 Schema

```sql
-- sms/backend/schema.sql
CREATE TABLE IF NOT EXISTS model_pricing_history (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id     VARCHAR(128) NOT NULL,
    input_price  NUMERIC(12,6) NOT NULL,
    output_price NUMERIC(12,6) NOT NULL,
    effective_at TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_model_pricing_history_lookup
    ON model_pricing_history (model_id, effective_at DESC);
```

### 3.2 写入时机

模型定价变更时（models 表 input_price / output_price 被 UPDATE 或 INSERT）INSERT 一条记录。

```sql
CREATE OR REPLACE FUNCTION fn_model_pricing_history()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF NEW.input_price IS NOT NULL OR NEW.output_price IS NOT NULL THEN
            INSERT INTO model_pricing_history (model_id, input_price, output_price, effective_at)
            VALUES (COALESCE(NEW.model_id, NEW.model_name),
                    COALESCE(NEW.input_price, 0), COALESCE(NEW.output_price, 0), NOW());
        END IF;
    ELSIF TG_OP = 'UPDATE' THEN
        IF OLD.input_price IS DISTINCT FROM NEW.input_price
           OR OLD.output_price IS DISTINCT FROM NEW.output_price THEN
            INSERT INTO model_pricing_history (model_id, input_price, output_price, effective_at)
            VALUES (COALESCE(NEW.model_id, NEW.model_name), NEW.input_price, NEW.output_price, NOW());
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_model_pricing_history
    AFTER INSERT OR UPDATE ON models
    FOR EACH ROW EXECUTE FUNCTION fn_model_pricing_history();
```

### 3.3 查询示例

```sql
-- 查某模型在某时刻的生效价格
SELECT input_price, output_price
FROM model_pricing_history
WHERE model_id = 'deepseek-chat'
  AND effective_at <= '2026-07-28T15:00:00Z'
ORDER BY effective_at DESC
LIMIT 1;
```

---

## 4. NewAPI 侧：logs 表加 ratio 字段

### 4.1 改动

NewAPI logs 表新增两列：

```sql
ALTER TABLE logs ADD COLUMN model_ratio NUMERIC(12,6);
ALTER TABLE logs ADD COLUMN completion_ratio NUMERIC(12,6);
```

写 consume log 时，将当时用于计算 quota 的 `model_ratio` 和 `completion_ratio` 一并写入。

### 4.2 quota 计算公式

```
quota = prompt_tokens * model_ratio + completion_tokens * model_ratio * completion_ratio
```

这两个值在计算 quota 时已经在内存中，顺手写入即可。

---

## 5. TokenJoy 侧：从 consume log 读 ratio，写入 call_detail

### 5.1 RawConsumeLog 扩展

```go
type RawConsumeLog struct {
    // ... existing
    ModelRatio      float64  // 新增：input token 倍率
    CompletionRatio float64  // 新增：output token 倍率
}
```

### 5.2 UsageCallDetail 扩展

```go
type UsageCallDetail struct {
    // ... existing
    ModelRatio      float64 `json:"modelRatio,omitempty"`      // 新增
    CompletionRatio float64 `json:"completionRatio,omitempty"` // 新增
}
```

### 5.3 buildCallDetail 填充

```go
func buildCallDetail(input EntryBuildInput, ...) types.UsageCallDetail {
    detail := types.UsageCallDetail{...}
    detail.ModelRatio = input.Raw.ModelRatio
    detail.CompletionRatio = input.Raw.CompletionRatio
    return detail
}
```

零额外查询。值就是 NewAPI 算 quota 时实际用的那个——最准确。

### 5.4 对已有数据的影响

- JSONB schema-free，新增字段不影响旧数据
- 旧 ledger 记录 call_detail 里没有 ratio（omitempty）
- 对账旧数据需要回查 SMS model_pricing_history

---

## 6. 对账流程

```
对账某笔 usage_ledger 记录：
  1. 优先看 call_detail.modelRatio / completionRatio（新数据有）
  2. 没有 → 查 SMS model_pricing_history WHERE effective_at <= occurred_at

验证公式：
  expected_quota = prompt_tokens * model_ratio
                 + completion_tokens * model_ratio * completion_ratio
  actual_quota = usage_ledger.quota_amount
  差异应为 0（同源计算）
```

---

## 7. 实施步骤

| # | 系统 | 改动位置 | 内容 |
|---|------|----------|------|
| 1 | SMS | `schema.sql` | 加 model_pricing_history 表 + trigger |
| 2 | NewAPI | logs 表 | 加 model_ratio / completion_ratio 列，写 log 时填充 |
| 3 | TokenJoy | `store/log_repo.go` | RawConsumeLog 加 ModelRatio / CompletionRatio |
| 4 | TokenJoy | `domain/types/usage_ledger.go` | UsageCallDetail 加 ModelRatio / CompletionRatio |
| 5 | TokenJoy | `domain/usage/entry_build.go` | buildCallDetail 从 Raw 读 ratio 填充 |

---

## 8. 注意事项

- SMS trigger 只在价格真正变更时写入（IS DISTINCT FROM），不会因为其他字段 UPDATE 产生噪音
- NewAPI log 记录的是实际计算用的 ratio，TokenJoy 只是透传，不做转换
- SMS history 记录的是 input_price/output_price（运营配置），NewAPI 的 model_ratio 是从 price 转换来的——两者可互推但精度可能有微小差异
- 对账时以 NewAPI log 的 ratio 为准（因为 quota 是用它算的）
