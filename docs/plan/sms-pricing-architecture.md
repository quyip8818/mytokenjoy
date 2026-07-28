# SMS 定价架构分析

## 计费链路

```
sms-newapi (ModelRatio/CompletionRatio)
    ↓ smssync worker 同步
tokenjoy-newapi (写入相同的 ratio)
    ↓ LLM 请求时计算 quota
    quota = (input_tokens × ModelRatio + output_tokens × ModelRatio × CompletionRatio) × GroupRatio / 1000 × 500000
    ↓ 写入 logs 表
tokenjoy ingest 读取 quota（直接透传）
    ↓ 扣减钱包
    cost_CNY = quota / 500000
```

## 换算公式

| 方向 | 公式 |
|------|------|
| Ratio → 显示价格 | `输入价 = ModelRatio × 2`，`输出价 = ModelRatio × CompletionRatio × 2` |
| 显示价格 → Ratio | `ModelRatio = 输入价 / 2`，`CompletionRatio = 输出价 / 输入价` |

示例：ModelRatio=30, CompletionRatio=2 → 输入 ¥60/M tokens, 输出 ¥120/M tokens

## 核心结论

**tokenjoy 没有独立的价格层**——它直接使用 NewAPI 的 quota 作为计费依据。NewAPI 的 ModelRatio + CompletionRatio + GroupRatio 完全决定了最终价格。

## 三层定价架构

| 层 | 数据 | 作用 | 位置 |
|---|---|---|---|
| 成本价 | sms-newapi 倍率 | 供应商给你的采购价格 | sms-newapi ModelRatio |
| 售价 | SMS admin 定价 | 面向客户的销售价格 | SMS models 表 input_price/output_price |
| 运行时计费 | tokenjoy-newapi 倍率 | 客户实际被扣费的价格 | tokenjoy-newapi ModelRatio |

**成本 ≠ 售价**——中间的差价就是利润，SMS admin 的售价层是这个利润控制的入口。

## 为什么 SMS admin 的售价层是必要的

### 差异化定价场景

```
sms-newapi: deepseek-v4-pro ModelRatio=0.2175 (成本: ¥0.435/M)

SMS admin 针对不同客户设置不同售价：
  客户 A (大客户): ¥0.5/M  → 利润率 15%
  客户 B (标准):   ¥0.65/M → 利润率 50%
  客户 C (小客户): ¥0.87/M → 利润率 100%
```

### 数据流（差异化定价）

```
sms-newapi (成本倍率)
    ↓ pull 同步到 SMS admin
SMS admin (展示成本价 + 管理售价)
    ↓ GET /sync/catalog?customer=A (按客户返回对应售价)
tokenjoy 实例 A (写入客户 A 的售价倍率到本地 newapi)
    ↓ Gateway 计费按客户 A 的售价执行
```

## 当前实现状态

| 能力 | 状态 | 备注 |
|---|---|---|
| 成本价从 sms-newapi 同步到 SMS admin | ✅ | `cost_input`/`cost_output` 字段 |
| 统一售价编辑 | ✅ | `input_price`/`output_price` 字段 |
| 售价同步到 tokenjoy-newapi | ✅ | smssync → `UpsertModelRatio` |
| 按客户差异化定价 | ❌ | 需要 `customer_pricing` 表 + catalog API 改造 |

## 后续演进方向

### Phase 1（当前）：统一售价

所有客户同价——SMS admin 的 `input_price`/`output_price` 就是唯一售价，同步给所有 tokenjoy 实例。

### Phase 2：按客户差异化

1. SMS admin 新增 `customer_pricing` 表：`(customer_id, model_id, input_price, output_price)`
2. Catalog API 改造：`GET /sync/catalog` 按请求方的 customer_id 返回对应售价
3. 未设置客户专属价格时，fallback 到默认售价
4. tokenjoy 侧无需改动——它只消费 catalog 返回的价格，不关心是默认还是专属

### 数据模型扩展

```sql
-- SMS admin 新增表
CREATE TABLE customer_pricing (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id  UUID NOT NULL REFERENCES customers(id),
    model_id     UUID NOT NULL REFERENCES models(id),
    input_price  NUMERIC(12,6) NOT NULL,
    output_price NUMERIC(12,6) NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (customer_id, model_id)
);
```

### Catalog API 查询逻辑

```sql
SELECT
    m.model_id,
    m.model_name AS display_name,
    COALESCE(cp.input_price, m.input_price) AS input_price,
    COALESCE(cp.output_price, m.output_price) AS output_price
FROM models m
LEFT JOIN customer_pricing cp ON cp.model_id = m.id AND cp.customer_id = $1
WHERE m.status = 'available'
```

## 总结

SMS admin 的售价层**不是多余的**，它是：
1. 当前统一定价的管理入口（成本价 + 利润 = 售价）
2. 未来差异化定价的基础（默认售价 + per-customer override）
3. 成本价和售价分离的关键——让管理员看到利润率，而不是直接操作抽象的 ratio 数字
