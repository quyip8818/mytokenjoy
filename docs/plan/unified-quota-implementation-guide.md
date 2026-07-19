# 实现指导：统一 NewAPI Quota 记账

> 去掉 tokenjoy 自有 points 层，以 NewAPI 的 `quota (int64)` 为唯一内部记账单位。

---

## 1. 架构

```
NewAPI DB (只读)                       TokenJoy DB
────────────────                       ───────────
logs.quota (int64) ──── 直连读 ────► entry.Amount = raw.Quota
                                              │
                                              ▼
                                        FIFO lot 扣减 (int64)
                                              │
                                              ▼
                                        combined_key_remain (Gateway 限额)

NewAPI Admin API ◄───── 写 ──────── CreateToken(unlimited_quota=true)
  token (不限额)                     UpdateToken(status, model_limits)
```

| 交互 | 说明 |
|------|------|
| 读 | tokenjoy 直连 NewAPI DB，读 `logs.quota` |
| 写 | tokenjoy 调 NewAPI Admin API 管理 token（status/model_limits/group），**不设限额** |
| 限额 | tokenjoy Gateway precheck 的 `combined_key_remain` 是唯一限额控制面 |
| 同步 | 无 — wallet_sync 整套删除 |

---

## 2. 单位规则

| 名称 | 类型 | 用途 |
|------|------|------|
| quota | int64 | 一切存储、传输、计算 |
| display | float64 | 仅 UI 渲染 + 财务凭证 |

```
quota   = Round(display × quota_per_unit)
display = quota / quota_per_unit
```

`quota_per_unit` 在 `currencies` 表中按币种定义（`CNY = 500000`：1元 = 50万 quota）。

**为什么是 500000？** 该值等于 NewAPI 现有的 `QuotaPerUnit` 常量（定义了 500000 quota = 1 model price point）。对齐后，当 model_ratio = 1 时 `raw.Quota` 可直通记账，无需二次换算。若 NewAPI 未来调整此精度因子，需同步更新 currencies 表并处理历史 lot 快照差异。

---

## 3. 核心数据模型

### currencies

```sql
CREATE TABLE currencies (
    code           TEXT PRIMARY KEY,
    quota_per_unit BIGINT NOT NULL,     -- 1元 = 多少 quota
    enabled        BOOLEAN NOT NULL DEFAULT TRUE
);
-- seed: ('CNY', 500000)
```

### lot（充值批次）

```go
type RechargeLot struct {
    ID              uuid.UUID
    CompanyID       uuid.UUID
    BillingCurrency string   // "CNY"
    LotKind         string   // paid | gift | adjust | overdraft
    QuotaPerUnit    int64    // 充值时汇率快照
    QuotaGranted    int64
    QuotaRemaining  int64
    AmountDisplay   float64  // 付款金额（gift/overdraft 为 0）
    Status          string
}
```

### 其他表统一类型

| 表 | 列 | 类型 |
|----|----|------|
| companies | wallet_remain | bigint |
| members | personal_budget | bigint |
| platform_keys | budget, consumed | bigint |
| budget_nodes | budget (nullable), reserved_pool | bigint |
| budget_consumed | consumed | bigint |
| usage_ledger | amount | bigint |
| usage_ledger | display_amount | float64 (= amount / lot.QuotaPerUnit) |
| usage_buckets | cost | bigint |
| usage_buckets | display_cost | float64 |
| combined_key_summaries | remain | bigint |

---

## 4. 核心路径代码

### 充值

```go
func QuotaFromAmount(amount float64, quotaPerUnit int64) int64 {
    return int64(math.Round(amount * float64(quotaPerUnit)))
}

// ¥50 充值: QuotaFromAmount(50, 500000) = 25_000_000
```

### Ingest（消耗入账）

```go
entry.Amount = raw.Quota  // 直接用 NewAPI 的 quota，零转换
```

### FIFO Lot 消耗

```go
type Segment struct {
    LotID         uuid.UUID
    Quota         int64
    QuotaPerUnit  int64    // lot 的汇率快照
    DisplayAmount float64  // = float64(Quota) / float64(QuotaPerUnit)
    Currency      string
}

func consumeLots(co *Company, amount int64) ConsumeResult {
    remaining := amount
    for _, lot := range lots {
        if remaining <= 0 { break }
        take := min(lot.QuotaRemaining, remaining)
        segments = append(segments, Segment{
            LotID:         lot.ID,
            Quota:         take,
            QuotaPerUnit:  lot.QuotaPerUnit,
            DisplayAmount: float64(take) / float64(lot.QuotaPerUnit),
            Currency:      lot.BillingCurrency,
        })
        lot.QuotaRemaining -= take
        remaining -= take
    }
    // overdraft...
}
```

**Gift/Overdraft lot**：`QuotaPerUnit` 设为创建时公司当前的 `currencies.quota_per_unit`，`AmountDisplay = 0`。消耗时 `DisplayAmount` 产出"等价金额"。

### 预算校验（combined_key_remain）

```go
// combined_key_remain = min(key 维度余额, 成员/项目维度余额)
// 注意：wallet_remain 不参与此计算，由 Gateway precheck 独立检查。
remain := GatewayChainRemain(key.Scope, ChainInputs{
    KeyBudget:   key.Budget,
    KeyConsumed: key.Consumed,
    // + member / project 维度（按 scope 决定）
})
```

Gateway precheck 分两层独立检查：
1. **wallet_remain > 0**：公司级余额（`Evaluate()` 中 `pc.Wallet.WalletRemain < minEstimate` 判断）
2. **combined_key_remain > 0**：key 级预算链（key、member、project 取 min）

两者互不包含，任一为零即拒绝。

### NewAPI Token

```go
// create: 不限额
req := adminport.CreateTokenInput{
    UnlimitedQuota: true,
    // ...
}

// update: 只管 status / model_limits / group，不传 RemainQuota
```

### 前端

```typescript
// lib/quota-display.ts
export const quotaToAmount = (q: number, qpu: number) => qpu > 0 ? q / qpu : 0
export const amountToQuota = (a: number, qpu: number) => Math.round(a * qpu)
export const formatQuota = (q: number, qpu: number, cur = 'CNY') =>
  new Intl.NumberFormat('zh-CN', { style: 'currency', currency: cur, minimumFractionDigits: 2 })
    .format(quotaToAmount(q, qpu))
```

`quotaPerUnit` 由 Session API 下发（`identity/authz/service.go` → `SessionContext.PointsPerUnit`，改名为 `QuotaPerUnit`）。

### 预算设置（UI → 存储）

```typescript
// 前端：用户输入 ¥100 → 转成 quota 发给后端
const budget = amountToQuota(100, quotaPerUnit) // 100 × 500000 = 50_000_000
await api.updateDepartment(deptId, { budget })
```

```go
// 后端 handler：接收 int64 quota，直接传给 domain service
func (h *Handler) UpdateNode(w http.ResponseWriter, r *http.Request) {
    var body struct {
        Budget       int64  `json:"budget"`
        ReservedPool *int64 `json:"reservedPool"`
    }
    // ...
    h.service.UpdateNode(ctx, id, body.Budget, body.ReservedPool)
}
```

---

## 5. 变更清单

### 删除

| 目标 |
|------|
| `internal/pkg/exchange/` 整包 |
| `internal/domain/billing/wallet_sync.go` |
| `internal/domain/usage/cost_from_log.go` |
| `internal/pkg/newapiunits/` 中 `ToNewAPIUnits` / `FromNewAPIUnits` / `CostFromQuota` |
| `internal/pkg/common/constants.go` 中 `DefaultPointsPerUnit` / `WalletSyncDriftEpsilon` / `WalletSyncDebounceSecs` |
| `internal/domain/billing/lot.go` 中 `PointsGrantedFromAmount` / `PaidLotDisplayAmount` / `UnitPriceDisplay` |
| `seed/points/` 整目录 + `seed/snapshot/points.go` |
| `frontend/src/lib/points.ts` |
| wallet_sync River job + worker + `ReconcileWalletDrift` 定时任务 |
| `internal/domain/company/wallet.go` 中 `FreshNewAPIUnits` (无其他用途时) |

### 修改

| 路径 | 摘要 |
|------|------|
| `pkg/common/constants.go` | `QuotaPerUnit` → `DefaultQuotaPerUnit` |
| `store/billing_repo.go` + postgres 实现 | struct 字段 rename + int64；SQL 列名对齐 |
| `store/postgres/billing_repo_wallet.go` | AggregateWallet SQL：`quota_remaining::numeric / quota_per_unit` |
| `store/postgres/bootstrap.go` | `points_per_unit` → `quota_per_unit` |
| `domain/billing/lot.go` | 新 `BuildPaidLot`/`BuildGiftLot` 使用 QuotaPerUnit |
| `domain/billing/lot/consume.go` | float64→int64；Segment 加 QuotaPerUnit |
| `domain/billing/lot/ledger.go` | `display_amount = quota / quotaPerUnit` |
| `domain/billing/lot_confirm.go` | 用 `QuotaFromAmount()` |
| `domain/billing/currency.go` | `lookupCurrencyPPU` → `lookupQuotaPerUnit` |
| `domain/usage/ingest.go` + `entry_build.go` | `entry.Amount = raw.Quota` |
| `domain/budget/` + `pkg/budget/` 全部 | float64→int64；删 `exchange.Format` |
| `domain/newapisync/platformkey/` | 删 `ToNewAPIUnits`；`UnlimitedQuota: true` |
| `domain/budget/rebalance.go` | 删 `ToNewAPIUnits` 调用（token unlimited 后不再同步 remain） |
| `http/handler/budget/` | API 请求/响应改 int64；下发 quotaPerUnit |
| `domain/dashboard/bucket_from_ledger.go` | `Cost` 字段 = quota(int64)，`DisplayCost` = displayAmount(float64) |
| `store/postgres/usage_aggregate.go` | dashboard series/summary API 返回 `display_cost` 而非 `cost` 给前端 |
| `seed/filler/members.go` | `x * DefaultQuotaPerUnit` |
| `seed/snapshot/models.go` | 模型价格存人民币元 |
| 前端 `features/budget/` ~12 文件 | 改 import → `@/lib/quota-display` |
| 前端 `features/dashboard/` ~4 文件 | 同上 |
| 前端 `features/keys/` ~3 文件 | 同上 |
| 前端 `features/session/` | 提供 quotaPerUnit context |

### 新建

| 路径 |
|------|
| `frontend/src/lib/quota-display.ts` |

---

## 6. 验证

| 场景 | 预期 |
|------|------|
| `QuotaFromAmount(50, 500000)` | 25,000,000 |
| Ingest quota=1000 → ledger.amount | 1000 |
| FIFO: lot1(QPU=500000, rem=5M) + lot2(QPU=600000), 消耗 6M → lot1 扣 5M(display ¥10), lot2 扣 1M(display ¥1.67) | 正确分段 |
| Gift lot QPU=500000, 消耗 1M | display = ¥2.00 |
| key.budget=10000, consumed=9500, 新 quota=600 | 拒绝 |
| Gateway precheck: combined_key_remain=0 | 拒绝 |
| AggregateWallet: paid lot QPU=500000, remaining=3M | balance=¥6.00 |
| 前端 formatQuota(25000000, 500000) | ¥50.00 |
| NewAPI token create | unlimited_quota=true |

---

## 7. 决策与风险

| 决策/风险 | 说明 |
|-----------|------|
| **lot 存 QuotaPerUnit 而非 UnitPriceDisplay** | int64 精确；display 可衍生；避免浮点 |
| **Gift/Overdraft lot QuotaPerUnit = 公司当前值** | 保证除法有意义；AmountDisplay=0 表明无实际支付 |
| **NewAPI token unlimited_quota=true** | 限额由 combined_key_remain 控制；避免 NewAPI 自动扣减 vs tokenjoy 异步 ingest 的 race |
| **模型价格仅 UI** | NewAPI model_ratio 已决定 quota 消耗量；tokenjoy 不二次定价 |
| **wallet_sync 完全删除** | 同单位无 drift；token unlimited 后无需同步 |
| **全部 int64** | 消灭浮点精度问题；quota 天然整数 |
| **AggregateWallet SQL numeric cast** | bigint/bigint 在 PG 中是整数除法，需 `::numeric` 保留小数 |
| **float64 精度边界** | amount × quotaPerUnit < 9×10^15 安全。1亿元 × 500000 = 5×10^13 ✓ |
| **Gateway precheck 是唯一限额** | 实时性 = 当前方案（PG 查询）；ingest 延迟的短暂超发是既有行为，未引入新风险 |
| **Gift lot display 语义** | 展示为"等价金额"而非"花费金额"；UI 可用标签区分 |

---

## 8. 公式速查

```
充值:       quota = Round(amount × quota_per_unit)
消耗入账:   entry.amount = consume_log.quota
lot 扣减:   lot.quota_remaining -= take
lot cost:   display = take / lot.quota_per_unit
余额:       wallet_remain -= entry.amount
预算:       combined_key_remain = min(key余额, 成员/项目余额)  ← 不含 wallet
Gateway:    拒绝 if wallet_remain ≤ 0 OR combined_key_remain ≤ 0
UI 展示:    display = quota / quotaPerUnit
UI 输入:    quota = Round(display × quotaPerUnit)
```
