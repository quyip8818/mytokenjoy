# ADR: 预算体系从 Quota 单位迁移到货币单位

## 状态

已采纳（实施中）

## 背景

当前限额体系（budget、consumed、remain）全部以 quota 单位存储。用户设定限额（例如 `5000`）时，前端调用 `displayToQuota(5000)` 按当时的 `quota_per_unit`（QPU）换算成 quota 值存入后端。

**核心问题：QPU 会变动。** 变动后：
- 已存的 budget 对应的货币购买力变了——用户设了 `5000`，QPU 涨了后显示变成约 `4166`
- 消费端按新 QPU 生成 quota amount，同样的模型调用消耗的 quota 数量变了
- 用户的预期是「我设了 5000 就是 5000（公司所选币种）」，不是「我设了 25 亿 quota」

**行业标准做法：** OpenAI、Portkey、Solo AgentGateway 等主流 AI API 平台的 spending limit 均以货币计价，消费时按当时的模型定价实时计算花费金额。平台不固定币种，币种由客户在公司级配置中选择。

**当前系统已有的正确基础：** usage_ledger 已经同时记录 `amount`（quota）和 `display_amount`（货币），dashboard 的 TotalCost 已经通过 `display_cost` 正确反映货币花费。gap 只在限额执行层（budget_consumed / budget chain / overrun / gateway precheck）。

## 决策

**将预算限额体系的存储和比较单位从 quota 迁移到货币（公司 `billing_currency` 下的 display amount）。** 预算设多少就是多少，不受 QPU 变动影响。

平台**不指定**默认或强制币种；币种由客户选择，落在 `companies.billing_currency`。同一公司下 budget / consumed / remain / display_amount 一律按该公司所选币种解释与运算。

## 核心原则

- **Budget 存货币金额** — 用户设 `5000` 后端就存 `5000.00`（单位为该公司 `billing_currency`），QPU 怎么变都不影响
- **Consumed 存货币金额** — 每次消费用该笔结算时的 **lot segment `display_amount` 之和** 累加，而非 quota，也非未赋值的 `entry.DisplayAmount`
- **Gateway remain 存货币金额** — `remain = budget - consumed`，单位一致直接相减
- **Quota 仅用于钱包层**（lot 消费、wallet_remain_quota）——这是「有多少 quota 余额」的问题，与限额无关
- **币种公司级唯一** — 不在 key/budget 行上冗余 currency 列；展示与结算时查 `companies.billing_currency`
- **无约束 remain = NULL** — 未配置任何预算上限时，`combined_key_remain` 保持 NULL（放行）；禁止再用 `MaxInt64` / 巨大 float 作哨兵

## 不动的三块

| 组件 | 单位 | 为什么不动 |
|------|------|-----------|
| `wallet_remain_quota` | quota (int64) | 钱包余额是「还剩多少 quota 可消费」，lot FIFO 扣减是整数原子操作，改成货币金额会引入浮点精度问题 |
| `combined_key_remain` 列名 | 不改名 | 名字描述「综合剩余」，单位由 schema 约定，不编码在列名里；一个公司只有一个 `billing_currency`，不需要额外存 currency 列 |
| NewAPI `user.remain_quota` | quota (int64) | 外部系统，token 已设 unlimited，user quota 是松散天花板，tokenjoy 侧做精确控制 |

## 超额检查流程（改后）

```
请求进来
  │
  ├─ 检查 1: wallet_remain_quota > 0 ?      ← 单位：quota，不动
  │    "公司有没有可用余额"
  │
  └─ 检查 2: combined_key_remain > 0 ?       ← 单位：货币金额（改）
       "这把 key 的预算还剩多少"
       │
       ├─ remain 怎么算：
       │    remain = min(
       │      key.budget - key.consumed,
       │      member.personalBudget - member.consumed,
       │      project.budget - project.consumed,
       │    )
       │    若无任何预算约束 → remain 不写（NULL = 放行）
       │
       ├─ consumed 怎么累加：
       │    消费时 lot 扣减产生 segments[].display_amount（货币）
       │    spend = Σ segments.DisplayAmount
       │    → budget_consumed += spend
       │    （禁止用 entry.DisplayAmount：BuildCallSettledEntry 不填该字段，恒为 0）
       │
       └─ remain 怎么实时扣减：
            DecrementBatch 用 spend（货币）而非 quota
```

两个检查独立、不同单位、各管各的。全部通过才放行。

## 数据流对比

### 当前（quota 维度）

```
用户设限额 5000（公司 billing_currency）
  → displayToQuota(5000) = 2,500,000,000 quota
  → 存入 platform_keys.budget = 2500000000

消费一次（NewAPI 上报 quota=30000）
  → budget_consumed += 30000
  → remain = 2500000000 - 30000 = 2499970000

QPU 变了？budget 不变，但对应的货币显示值变了 ← BUG
```

### 目标（货币维度）

```
用户设限额 5000（公司 billing_currency）
  → 存入 platform_keys.budget = 5000.00

消费一次
  → lot 消费得到 segments，spend = Σ display_amount（例 0.06）
  → budget_consumed += 0.06
  → remain = 5000.00 - 0.06 = 4999.94

QPU 变了？budget 不变（仍是 5000），消费金额由 lot 快照汇率决定 ← 正确
```

## Schema 变更

| 表 | 列 | 当前类型 | 目标类型 | 说明 |
|----|----|---------|---------|----|
| `platform_keys` | `budget` | BIGINT (quota) | NUMERIC(18,6) | 货币金额（公司 billing_currency） |
| `platform_keys` | `combined_key_remain` | BIGINT (quota) | NUMERIC(18,6) | 货币金额；NULL = 无约束放行 |
| `projects` | `budget` | BIGINT (quota) | NUMERIC(18,6) | 正规表列，非 JSONB |
| `project_members` | `member_budget` | BIGINT (quota) | NUMERIC(18,6) | 正规表列，非 JSONB |
| `members` | `personal_budget` | BIGINT (quota) | NUMERIC(18,6) | 货币金额 |
| `org_nodes` | `budget` | BIGINT (quota) | NUMERIC(18,6) | 货币金额 |
| `org_nodes` | `reserved_pool` | BIGINT (quota) | NUMERIC(18,6) | 与 budget 同单位 |
| `org_nodes` | `member_avg_budget` | BIGINT (quota) | NUMERIC(18,6) | 与 budget 同单位 |
| `budget_consumed` | `consumed` | BIGINT (quota) | NUMERIC(18,6) | 累计花费（货币） |

> 选 NUMERIC(18,6) 而非 float：精确到最小货币子单位以下，避免浮点累加误差。Go 侧统一用 float64（15-16 位有效数字，对金额远超需要）。
>
> `combined_key_remain` 落在 `platform_keys` 列上（无独立 `combined_key_summaries` 表）；store 接口名 `CombinedKeySummaries` 可保留。

## 改动清单

### 后端 — Domain/Types 层

| 文件 | 变更 |
|------|------|
| `domain/types/budget.go` | BudgetNode/Project/MemberBudget 相关字段 int64 → float64 |
| `domain/types/keys.go` | PlatformKey.Budget/Consumed int64 → float64 |
| `domain/budget/consumed_attrib.go` | `AxisDelta.Amount` int64 → float64；赋值改为 **spend（segment display 之和）** |
| `pkg/budget/chain.go` | `ChainInputs` 所有字段 int64 → float64；`GatewayChainRemain` 返回 `(remain float64, limiting string, uncapped bool)`（或等价：uncapped 时不产生 persist update） |
| `domain/budget/combined_key_cache.go` | `CombinedKeyEntry.Remain` int64 → float64 |
| Overrun 判断 | `BudgetExhausted(consumed, budget float64)` |

### 后端 — Store 接口层

| 文件 | 变更 |
|------|------|
| `store/budget_consumed_repo.go` | `ConsumedDelta.Amount` / `GetConsumed` / `ListConsumed*` → float64 |
| `store/combined_key_summary.go` | `Remain` / `DecrementBatch` map value → float64 |
| 审批相关 Amount/RequestedBudget | 与 budget 一并改为货币 float64 |

### 后端 — Store 实现层

| 文件 | 变更 |
|------|------|
| `store/postgres/budget_consumed_repo.go` | SQL scan/insert 对齐 NUMERIC(18,6)；Go 侧 float64 |
| `store/postgres/combined_key_summary_repo.go`（及 platform_keys remain 读写） | 同上 |
| org/member/project/key budget 读写 | BIGINT → NUMERIC scan |

### 后端 — Ingest 路径（关键）

| 文件 | 变更 |
|------|------|
| `domain/usage/ingest.go` | 在 lot consume 之后计算 `spend = Σ segments.DisplayAmount`；`ConsumptionDeltas` 与 `DecrementBatch` 均用 `spend`，**禁止** `entry.DisplayAmount` / `entry.Amount` |
| `domain/budget/consumed_attrib.go` | deltas 金额改为传入的 spend（float64） |
| `adapter/usage_budget.go` | Amount int64 → float64 |
| `domain/budget/alert_publisher.go` + `ledger_repo.SumAmountByDepartment` | 改为 `SUM(display_amount)`，与货币 budget 对齐 |

### 后端 — Rebalance/Precheck 路径

| 文件 | 变更 |
|------|------|
| `pkg/budget/mapping_remain.go` | PreloadedConsumed / getConsumed → float64 |
| `domain/budget/combined_key_summary.go` | 删除 `remain >= MaxInt64`；改为 `uncapped → skip update（NULL）` |
| `domain/gateway/precheck.go` | `budgetRemainCheck` 适配 float64（`remain <= 0` 不变；NULL 仍放行） |
| Redis combined_key cache | Remain float64；仍用 Set 覆盖，不用 DECRBY |

### 前端

| 变更 |
|------|
| 设 budget 时直接发货币金额（删除 `displayToQuota` 调用） |
| 展示 budget/consumed 时不再做 `quotaToDisplay`（后端直接返回货币）；改用 `formatMoney` / 公司 `billing_currency` |
| 保留钱包余额的 quota→display 转换（钱包仍是 quota 维度） |
| 审批提交金额改为直发货币，去掉 `displayToQuota` |

### 不改的部分

| 组件 | 原因 |
|------|------|
| `wallet_remain_quota` | 钱包层，quota 维度，lot FIFO 整数扣减 |
| Lot 消费逻辑 | FIFO 扣 quota、记录 segment display_amount——完全正确 |
| usage_ledger.amount | 保留 quota 记录，用于钱包扣减和对账 |
| Dashboard display_cost | 已经是货币，不动 |
| `common.QuotaFromAmount` / `QuotaToDisplay` | 保留给钱包/充值场景 |
| NewAPI user quota 同步 | 仍然用 quota 维度（PreCreditFunc 中 `add_quota(lot.QuotaGranted)`） |
| `companies.billing_currency` | 客户选择币种的唯一来源；预算层只存金额，不另存 currency |

## 边界 case

**Ingest 金额来源？**
必须用 `Σ consumeResult.Segments[].DisplayAmount`。`entry.DisplayAmount` 在 `BuildCallSettledEntry` 中未赋值，恒为 0；若误用会导致 consumed/remain 永不更新。

**无预算约束时 remain？**
`GatewayChainRemain` 判定 uncapped 后，rebalance **不写** `combined_key_remain`（保持/恢复为 NULL）。Gateway 对 NULL 放行。禁止 `MaxInt64` / `+Inf` 哨兵。

**QPU 变动后，in-flight 的 combined_key_remain 缓存会不会有问题？**
不会。remain 单位是货币金额，QPU 变动不影响已缓存的 remain 值。新消费按新 lot 的 QPU 计算出 display_amount 后正常扣减。

**多 lot 不同 QPU 场景（multirate）？**
正确处理。一次消费可能跨两个 lot（QPU 不同），每个 segment 的 display_amount 按各自 lot 的快照 QPU 计算。累加到 budget_consumed 的是 display_amount 之和。

**Overdraft segment 的 DisplayAmount 为 0？**
正确。`lot/consume.go` 中 overdraft segment 的 `DisplayAmount = 0`。透支不计入 budget_consumed——符合预期：透支是钱包层行为。若「预算没超但公司欠费」，看 `wallet_remain_quota`。

**Gift lot 的 DisplayAmount？**
Gift **会计入** budget_consumed（代码对 gift 调用 `QuotaToDisplay`，非 0）。与今日按 quota 计入预算一致；仅 overdraft 为 0。

**Gateway precheck float 精度问题？**
不会。累加在 PG SQL 层完成（`consumed = consumed + $amount`，NUMERIC 精确运算），Go float64 仅用于传参和展示读取。Redis cache 用 `Set`（覆盖写整个值），不用 `DECRBY`。判断 `remain <= 0` 在 float64 精度下没有风险。

**预算不存 currency 列？**
不需要。一个公司只有一个由客户选择的 `billing_currency`。

**客户更换 billing_currency？**
本 ADR 不覆盖换币种迁移。若未来支持更换，需单独定义历史 budget/consumed/display_amount 的换算与冻结策略；当前假设公司币种选定后稳定使用。

## 实施

项目未上线，破坏性更新，直接改 schema + 重新 seed，不需要迁移脚本。改 schema / seed 后必须 bump `testTemplateVersion`。

### 步骤

1. Schema（BIGINT → NUMERIC(18,6)）+ Go types（int64 → float64）+ store interface + **uncapped→NULL**
2. Ingest 路径（`spend = Σ segments.DisplayAmount` → deltas + DecrementBatch）；部门告警 `SUM(display_amount)`
3. Rebalance/Chain（`ChainInputs` float64，uncapped 不写 remain）
4. 前端（去掉 budget 的 `displayToQuota`/`quotaToDisplay`；展示跟 `billing_currency`）
5. Seed（budget 直接写货币金额，去掉 `seedQuota` 乘 QPU）+ 测试对齐货币语义

## 验证标准

- 用户设限额 `5000` → DB 存 `5000.0` → QPU 变动后 → API 返回仍是 `5000`（币种取自公司 `billing_currency`）
- 消费一次后 consumed 累加的是 Σ segment display_amount（货币），不是 quota；误用 `entry.DisplayAmount` 的路径不存在
- `budget - consumed = remain`（货币），gateway precheck 用此值拦截；无约束时 remain 为 NULL 且放行
- Dashboard TotalCost 与限额 consumed 一致（数据源相同：display_amount）
- QPU 变动不触发任何 budget 数据修正
- wallet_remain_quota 不受影响，仍为 quota 维度整数运算
- Overdraft 消费不累加到 budget_consumed（DisplayAmount = 0）；Gift 会计入
- API/UI 不硬编码 CNY/USD 等币种符号；展示跟随客户所选 `billing_currency`
- 部门 overrun/alert 使用 `SUM(display_amount)`，与货币 budget 同轴
