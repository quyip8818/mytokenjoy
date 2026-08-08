# Plan: 渠道成本分析（Platform Admin）

> 状态：待做
> 前置：模型可见性架构完成后

## 问题

Platform Admin 无法在系统内看到各上游渠道的真实采购成本和毛利。当前只有对客售价（Model Ratio），没有渠道成本价。成本核算只能去各供应商后台手动对账。

## 目标

Platform Admin 能看到：
- 每个渠道（channel）消耗了多少 token
- 每个渠道花了多少钱（按成本价算）
- 每个模型的毛利 = 客户付的 − 渠道成本

## 数据来源

已有：
- NewAPI logs 表：每条请求记录了 `channel_id`、`model_name`、`input_tokens`、`output_tokens`

缺失：
- 各 channel 的采购成本单价

## 方案

### 新增：channel_cost 表

```sql
channel_cost (
  channel_id      INT,
  input_cost      DECIMAL(10,4),   -- 每百万 token 成本（输入）
  output_cost     DECIMAL(10,4),   -- 每百万 token 成本（输出）
  effective_from  TIMESTAMPTZ,     -- 生效时间（支持调价历史）
  note            TEXT
)
```

append-only，按 `(channel_id, effective_from DESC)` 取当前生效行。和 model_discount 表设计思路一致。

### Platform Admin API

```
GET  /api/platform/channels/:id/cost        → 查看渠道成本配置
PUT  /api/platform/channels/:id/cost        → 设置成本价
GET  /api/platform/analytics/channel-cost   → 渠道成本报表
GET  /api/platform/analytics/margin         → 毛利分析
```

### 报表计算

```
渠道成本 = Σ (logs.input_tokens × channel_cost.input_cost 
           + logs.output_tokens × channel_cost.output_cost)
           WHERE logs.channel_id = X AND logs.created_at in [start, end]

客户收入 = Σ (logs.input_tokens × model_ratio.input_price
           + logs.output_tokens × model_ratio.output_price)
           WHERE logs.model_name = M AND logs.created_at in [start, end]

毛利(model) = 客户收入(model) − Σ 渠道成本(该 model 的所有 channel)
```

### UI

Platform Admin → 新 Tab「成本分析」：
- 按渠道看：各 channel 的 token 消耗 + 成本
- 按模型看：售价收入 vs 渠道成本 = 毛利率
- 按时间段筛选

## 不做的事

- 不影响客户侧任何逻辑（纯 platform admin 运营工具）
- 不影响 Ingest 扣费（扣费按售价，不按成本）
- 不同步到 Local（这是 SaaS 平台内部数据）

## 依赖

- 模型可见性架构（channel 和 model 分层清晰后再做）
- NewAPI logs 表的 channel_id 字段可用（已有）

## 优先级

低。量小时直接看供应商账单即可。量大、渠道多、需要实时监控毛利时再做。
