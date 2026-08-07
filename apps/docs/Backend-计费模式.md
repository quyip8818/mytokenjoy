# Backend 计费模式

内部统一 **quota (int64)** 计量；钱包 / CallLog / 看板以 **lot 冻结展示币** 为 SSOT；NewAPI `users.quota` 是 TokenJoy 的实时镜像，不是资金真相。

**相关：** [Backend-预算.md](./Backend-预算.md) · [Backend-存储架构.md](./Backend-存储架构.md) · [Backend-Ingest架构.md](./Backend-Ingest架构.md) · [Frontend.md](./Frontend.md)

---

## 1. 两套数

| 指标       | 用户看到                           | 后端字段                                                     | 用途                   |
| ---------- | ---------------------------------- | ------------------------------------------------------------ | ---------------------- |
| **展示币** | 钱包余额、CallLog 费用、看板 Spend | `balances[]`、`usage_ledger.cost`、`usage_buckets.cost` | 财务闭合；入账时冻结   |
| **Quota**  | 预算/Key 额度（UI 换算成「元」）   | `wallet_remain_quota`、`budget_*`、`key.budget`              | Gateway、预算、ingest  |

默认：`1 CNY = 10,000 quota`（`DefaultQuotaPerUnit`，与 `currencies` seed 对齐）。

易错：**量纲混用**（填 ¥ 当 quota 提交 → ×500000）、**二次换算**（对已是展示币再 ÷QPU）。

---

## 2. 三个世界

```
Usage tokens ──→ Quota (int64) ──→ Wallet 展示币（lot 冻结）
                       │
                       └──→ 预算 / soft remain
```

| 世界   | 含义                 | 典型字段                                          | 改公司币后现算？ |
| ------ | -------------------- | ------------------------------------------------- | ---------------- |
| Usage  | token / 次数         | `input_tokens` / `output_tokens`                  | —                |
| Quota  | 内部统一货币 (int64) | `wallet_remain_quota`、`usage_ledger.quota_amount`、`budget_*` | 不换币          |
| Wallet | lot 成本价 + 冻结    | `usage_ledger.cost`、`usage_buckets.cost`   | **否**           |

### 2.1 币种 / QPU

> 详见 [currency-management.md](./currency-management.md) — 币种 CRUD、SaaS→Local 同步、管理 UI。

| 位置                             | 作用                                              |
| -------------------------------- | ------------------------------------------------- |
| `currencies` 最新行 `.quota_per_unit` | QPU 表级 SSOT（append-only，取 `updated_at` 最大行） |
| `companies.billing_currency`     | 公司当前计费币（只影响**新**充值 / overdraft）     |
| Session `quotaPerUnit`           | FE 写边界注入（`ResolveCompanyChargeRate`）        |

### 2.2 冻结规则

1. 订单落 `currency` + `quota_per_unit` + `quota_granted`。
2. Lot：`display_amount = quota / lot.quota_per_unit`（paid/adjust）；gift/overdraft = 0。
3. 消耗段：`display_amount = take / lot.quota_per_unit`，币种 = lot.billing_currency。
4. 改 `billing_currency`：历史 lot / ledger 不回写。

---

## 3. 权威边界

```
事实面（强一致）         投影 / 派生（最终一致，可重建）
───────────────         ─────────────────────────────
company_recharge_lots   budget_consumed
usage_ledger            usage_buckets
wallet_remain_quota     gateway_soft_remain
                        NewAPI users.quota
```

| 能力           | SSOT                                   | 派生                              |
| -------------- | -------------------------------------- | --------------------------------- |
| 企业可用 quota | `wallet_remain_quota` / Σ lot remain   | —                                 |
| 展示币钱包     | lots（paid + adjust）                  | —                                 |
| 单笔消耗       | `usage_ledger`                         | —                                 |
| 组织 consumed  | —                                      | `budget_consumed`                 |
| 看板 Spend     | —                                      | `usage_buckets.display_cost`      |
| Gateway 挡单   | `wallet_remain_quota` + soft           | NewAPI 不参与预检                 |
| NewAPI wallet  | —                                      | 实时 override 镜像                |

**红线：** 禁止用 NewAPI quota 反算钱包；漂移以 Postgres 为准。

---

## 4. 核心流程

### 4.1 充值

```
用户充值 → ResolveCompanyChargeRate(currency, QPU)
        → quota = Round(amount × QPU)
        → TX: BuildLot(锁定 QPU/币种) + ApplyWalletDelta
        → TX commit
        → set_quota override → NewAPI
```

| 场景        | lot_kind   | 展示币              |
| ----------- | ---------- | ------------------- |
| 自助/平台   | `paid`     | `quota / QPU` 锁定 |
| 赠送        | `gift`     | 0                   |
| 调账        | `adjust`   | 显式写入            |
| ingest 透支 | `overdraft`| 0                   |

新企业 `wallet_remain_quota = 0`，无初始 lot。

### 4.2 消耗：FIFO + overdraft

```
consume log → TX: LockForUpdate(company)
           → 幂等检查
           → FIFO 扣 lot（跨 lot = 多段 ledger）
           → lot 不足 → 扩展 overdraft
           → SetWalletRemainQuota
           → TX commit
           → set_quota override → NewAPI
```

- 跨 lot 段：gift/overdraft 段 `display_amount = 0`。
- overdraft 保证 ingest 不因余额不足永久失败（应设告警）。

### 4.3 Gateway 预检

单位 = quota。不读 NewAPI。

| # | 检查                                        |
|---|---------------------------------------------|
| 1 | 企业 active                                 |
| 2 | `wallet_remain_quota ≥ minEstimate`         |
| 3 | `combined_key_remain > 0`（有配置时）       |
| 4 | Key active / 未过期                         |
| 5 | 模型白名单                                  |

### 4.4 Discount（模型折扣）

平台管理员可为指定公司配置 per-model 折扣系数，影响 ingest 时的 quota 扣减量。

**数据模型**：`model_discount` 表（append-only timeline）

```
(company_id, model_type, discount, effective_from, note)
```

**解析规则**（`domain/usage/discount.go` `resolveDiscount`）：
1. 精确匹配 `model_type` → 返回对应 discount
2. 通配 `*` → fallback
3. 无配置 → 1.0（无折扣）

**应用时机**：Ingest 入账时，在写 ledger 前对 `entry.QuotaAmount` 乘以 discount 系数：

```
entry.QuotaAmount = ceil(raw.Quota × discount)
```

例：`discount = 0.8` → 实际扣 quota = 原始的 80%（八折）。

**SaaS → Local 同步**：Catalog Sync 的 `discounts` 通道（per-company version，见 `sync_versions` 表）。平台管理员在 SaaS 设置折扣后 bump version → Local 下次轮询检测到变化 → 拉取 → 写入本地 `model_discount` 表。

**管理 UI**：
- 平台侧：`/platform/companies` 页面 → DropdownMenu →「优惠」→ Sheet 面板配置
- 企业侧：`/billing` 页面 → DiscountSection（只读，条件渲染：有数据才展示）

**API**：
| Method | Path | 权限 | 说明 |
|--------|------|------|------|
| GET | `/platform/companies/{id}/discounts` | `platform:manage` | 查看指定公司优惠 |
| PUT | `/platform/companies/{id}/discounts` | `platform:manage` | 设置优惠 |
| GET | `/api/billing/discounts` | `billing:read` | 企业查看自身优惠 |

---

### 4.5 Wallet 同步（NewAPI）

TokenJoy `wallet_remain_quota` 是 SOT。NewAPI wallet 是实时镜像。

```
┌────────────────────────────────────────────────────────────┐
│                        TokenJoy                             │
│                                                            │
│  ┌─────────────┐           ┌────────────────────┐          │
│  │ 充值路径     │           │ 消费路径 (Ingest)   │          │
│  │ CreditFromLot│           │ ConsumeLotsLocked  │          │
│  └──────┬──────┘           └─────────┬──────────┘          │
│         │ TX: ApplyWalletDelta       │ TX: SetWalletRemain │
│         ▼                            ▼                     │
│  ┌─────────────────────────────────────────┐               │
│  │     companies.wallet_remain_quota       │  ← SOT        │
│  └─────────────────────────────────────────┘               │
│         │         Post-commit              │               │
│         ▼         (best-effort)            ▼               │
│  ┌─────────────────────────────────────────┐               │
│  │  ManageUser("set_quota", value)         │               │
│  │  mode: "override"                       │               │
│  └────────────────────┬────────────────────┘               │
└───────────────────────┼────────────────────────────────────┘
                        │ HTTP POST /api/user/manage
                        ▼
            ┌─────────────────────┐
            │      NewAPI         │
            │  user.quota = value │  ← 纯镜像
            └─────────────────────┘
```

**写入路径：**

| 路径 | 触发时机 | 代码 |
|------|----------|------|
| 充值 | PlatformRecharge / Gift / Adjust / ConfirmPayment | `billing/lot_confirm.go` → `syncWalletBestEffort` |
| 消费 | 每条 consume log 入账 | `usage/ingest.go` → post-commit block |
| 升级 | trial/demo → standard | `company/service.go` → `UpgradeToStandard` |

**失败处理：**

| 场景 | 行为 |
|------|------|
| NewAPI 不可达 | warn log，不阻塞 |
| 重复 override 同一个值 | 幂等 |
| 并发 ingest（同 company） | company row lock 保证串行 |
| NewAPI 自身扣了 quota | 下一次 ingest override 覆盖回来 |

偏差方向安全：NewAPI wallet 暂时偏高（用户能多用一点），不会"有钱被拒"。

**NewAPI API 格式：**

```json
POST /api/user/manage
{
  "id": <walletUserID>,
  "action": "add_quota",
  "mode": "override",
  "value": <wallet_remain_quota 绝对值>
}
```

`action` 固定 `"add_quota"`，`mode` 区分语义：`"add"` 增量 / `"override"` 覆盖绝对值。

---

## 5. 数据模型

### 5.1 表关系

```
currencies ──1:N──→ companies (billing_currency)
companies  ──1:N──→ company_recharge_orders
company_recharge_orders ──1:1──→ company_recharge_lots
company_recharge_lots ──1:N──→ usage_ledger (debit segments)
company_recharge_lots ──1:N──→ lot_transactions (credit/refund 变动记录)
```

### 5.2 展示币闭合

```
单条:  display = quota / lot.quota_per_unit
余额:  balance(c) = Σ (quota_remaining × amount_display / quota_granted)
                     WHERE currency=c AND kind∈{paid, adjust}
```

### 5.3 Quota 守恒

```
Σ quota_granted − Σ ledger.amount = Σ quota_remaining = wallet_remain_quota
```

### 5.4 lot_kind

| kind      | 可花 | 计 totalTopup | 消耗 display              |
| --------- | ---- | ------------- | ------------------------- |
| paid      | ✅   | ✅            | `take / lot.quota_per_unit` |
| adjust    | ✅   | ✅            | 同上                       |
| gift      | ✅   | ❌            | 同公式                     |
| overdraft | ✅   | ❌            | 同公式                     |

---

## 6. 公式

```
entry.Amount = raw.Quota                    (NewAPI 日志直通，零转换)
display_amount = take / lot.quota_per_unit  (FIFO 冻结)
quota_granted = Round(amount × QPU)
DefaultQuotaPerUnit = 10000
```

**闭环验证：**

| # | 不变量 |
|---|--------|
| 1 | quota 守恒：授予 − ledger = remaining |
| 2 | 展示币闭合：`wallet_closure_test` |
| 3 | FIFO + ledger + wallet 同事务 |
| 4 | 幂等：`(company_id, idempotency_key)` |
| 5 | NewAPI token `unlimited_quota = true`（无需 per-token 同步） |
| 6 | 投影终态：`Σ ledger.amount ≈ budget_consumed`（reconcile 修复） |

---

## 7. 代码地图

| 模块 | 文件 | 职责 |
|------|------|------|
| 常量 | `support/quota/quota.go` | DefaultBillingCurrency / DefaultQuotaPerUnit |
| 币种解析 | `domain/billing/currency.go` | ResolveCompanyChargeRate |
| Lot 写入 | `domain/billing/lot/consume.go` | CreditFromLot / ConsumeLots |
| 充值确认 | `domain/billing/lot_confirm.go` | confirmPaidRecharge / syncWalletBestEffort |
| 退费 | `domain/billing/refund.go` | PlatformRefund（事务内锁 → 减 lot → 减 wallet → 记 lot_transactions） |
| Lot 审计 | `domain/billing/lot_audit.go` | ListLots / PlatformListLots / recordCreditTransaction |
| 钱包聚合 | `domain/billing/wallet_view.go` | AggregateWallet |
| Discount | `domain/usage/discount.go` | ApplyDiscount / resolveDiscount |
| Discount store | `store/model_discount.go` | ModelDiscountRepository (append-only) |
| 入账 | `domain/usage/ingest.go` | IngestRaw（FIFO 消费 + post-commit sync） |
| 预检 | `domain/gateway/evaluate.go` | wallet_remain_quota + combined_key_remain |
| NewAPI 客户端 | `integration/newapi/user.go` | ManageUser (set_quota → override) |
| Store | `store/postgres/company_repo.go` | ApplyWalletDelta / SetWalletRemainQuota |
| Store | `store/postgres/billing_repo_lots.go` | ListAllLots / lot_transactions CRUD |
| Sync Versions | `store/sync_versions.go` | SyncVersionRepository (per-company + global) |
| FE 换算 | `frontend/src/lib/quota-display.ts` | formatMoney / formatDisplayCurrency |

---

## 8. API 契约

### 8.1 Session

```json
{ "billingCurrency": "CNY", "quotaPerUnit": 500000 }
```

FE 用 `quotaToDisplay` / `displayToQuota` 做表单换算。

### 8.2 钱包

```json
{
  "billingCurrency": "CNY",
  "balances": [{ "currency": "CNY", "balance": 37.5, "totalTopup": 100, "totalConsumed": 62.5 }],
  "walletRemainQuota": 18750000,
  "giftQuota": 0,
  "overdraftQuota": 0
}
```

### 8.3 展示规则

| 数据类型              | 用什么展示           |
| --------------------- | -------------------- |
| 钱包余额 / CallLog    | `formatMoney`（禁止再 ÷QPU） |
| 预算额度 / Key remain | `formatDisplayCurrency` |

预算/Key API 只收发 **quota (int64)**。

---

## 9. 边界行为

| 场景     | 行为 |
| -------- | ---- |
| 预检不足 | 拒绝，不 proxy |
| lot 不足 | 扩展 overdraft |
| 改币种   | 旧 lot/CallLog 不变；新充值用新币 |
| 退费     | platform 管理员选定 lot → 扣减 `quota_remaining`（不删除 lot） → 减 `wallet_remain_quota` → 插 `lot_transactions` 记录 → bump version + sync |

---

## 10. 风险与演进

### 接受中的风险

| 风险 | 缓解 |
|------|------|
| soft lag | 投影加速 + budget_reconcile |
| wallet sync 失败 | best-effort + 下次变更自动覆盖；偏差方向安全 |
| 固定 minEstimate | 故意保持粗闸门 |
| float64 精度 | NUMERIC 落库 |

### 待做

| 优先级 | 项 |
|--------|-----|
| P1 | overdraft 告警/打点 |
| P2 | gift/adjust 运营 UI |
| P3 | 多币种 / 改币种流程 / lot 归档 |

### 红线

- 不用 NewAPI quota 反算钱包
- 不 UPDATE 历史 display_amount
- 不旁路直连 NewAPI 消费
- 不做 Gateway 动态 estimate（除非产品单独需求）
