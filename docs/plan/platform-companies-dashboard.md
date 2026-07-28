# Platform 企业管理看板

> 在 platform admin 中新增企业列表页，展示所有公司的财务概览和钱包状态。

---

## 现状

现有 `GET /api/platform/companies` 返回 `[]store.Company`，只有基础字段（id, name, type, status, billingCurrency, walletRemainQuota, createdAt）。没有：
- 余额（需要换算 quota → 金额）
- 累计充值 / 累计消耗
- 赠送额度 / 透支额度
- 月度开销统计
- 成员数 / key 数

platform admin 目前只能通过 Recharge/Gift/Adjust 操作单个公司，无法一览所有公司的财务状况。

---

## 目标

新增 `/platform/companies` 前端页面 + 一个增强的后端 API，展示：

| 列 | 数据来源 | 说明 |
|----|---------|------|
| 公司名称 | companies.name | |
| 类型 | companies.type | standard / trial / demo |
| 状态 | companies.status | active / suspended |
| 币种 | companies.billing_currency | CNY |
| 余额 | WalletAggregate.balance | quota → 金额换算 |
| 赠送余额 | WalletAggregate.gift_quota | |
| 透支额度 | WalletAggregate.overdraft_quota | |
| 累计充值 | WalletAggregate.total_topup | |
| 累计消耗 | WalletAggregate.total_consumed | |
| 本月消耗 | usage logs 聚合 | 当月 token 花费 |
| 成员数 | COUNT(members) | |
| 创建时间 | companies.created_at | |

操作列：充值 / 赠送 / 停用 / 详情

---

## API 设计

### `GET /api/platform/companies` （增强现有）

替换现有简单的 `ListCompanies`，返回带财务概览的公司列表。

```json
[
  {
    "id": "uuid",
    "name": "Demo Company",
    "type": "standard",
    "status": "active",
    "billingCurrency": "CNY",
    "wallet": {
      "balance": 1500.00,
      "giftBalance": 200.00,
      "overdraft": 0,
      "totalTopup": 5000.00,
      "totalConsumed": 3500.00
    },
    "monthlySpend": 420.50,
    "memberCount": 15,
    "createdAt": "2026-06-01T00:00:00Z"
  }
]
```

### 实现方案

```go
// handler/platform/companies.go
func (h *Handler) ListCompanies(w http.ResponseWriter, r *http.Request) {
    companies := h.p.CompanySvc.ListCompanies(ctx)
    // 对每个 company：
    //   1. AggregateWallet → balance/topup/consumed/gift/overdraft
    //   2. 月度开销：从 usage logs 或 dashboard 表读取当月聚合
    //   3. 成员数：COUNT members WHERE company_id = X
}
```

**性能考虑：** 公司数量预期 <100（SaaS 早期），逐个查 wallet 聚合可接受。如果以后公司数增加，改为批量 SQL 或物化视图。

### 数据来源映射

| 字段 | SQL / 方法 |
|------|-----------|
| balance | `SumActiveLotsRemaining` → quota / quotaPerUnit，或直接用 `companies.wallet_remain_quota / quotaPerUnit` |
| giftBalance | `SELECT SUM(quota_remaining) FROM recharge_lots WHERE company_id=$1 AND lot_kind='gift' AND status='active'` |
| overdraft | `SELECT SUM(quota_remaining) FROM recharge_lots WHERE company_id=$1 AND lot_kind='overdraft' AND status='active'` |
| totalTopup | `SELECT SUM(paid_amount) FROM recharge_lots WHERE company_id=$1 AND lot_kind IN ('paid','adjust')` |
| totalConsumed | `SELECT SUM(quota_granted - quota_remaining) FROM recharge_lots WHERE company_id=$1` → 换算金额 |
| monthlySpend | `SELECT SUM(cost) FROM consume_logs WHERE company_id=$1 AND period_key=$currentMonth` 或从 dashboard projection 读 |
| memberCount | `SELECT COUNT(*) FROM members WHERE company_id=$1 AND status='active'` |

---

## 前端

### 路由

`/platform/companies`，navGroup "平台管理"，与"模型目录"同组。

### 页面功能

1. **表格**：上述所有列，默认按余额排序
2. **操作按钮**：
   - 充值（弹窗输入金额 → `POST /companies/:id/recharge`）
   - 赠送（弹窗 → `POST /companies/:id/gift`）
   - 停用/启用（`PATCH /companies/:id` status toggle）
3. **统计卡片**（表格上方）：
   - 总公司数
   - 活跃公司数
   - 总余额
   - 本月总消耗

### 代码组织

```
features/platform/companies/
  hooks/use-platform-companies-page.ts
  components/platform-companies-page-shell.tsx
  query-keys.ts  →  platformKeys.companies()
  index.ts
```

API 加在 `api/platform.ts` 中：
```ts
listCompanies: () => request<PlatformCompany[]>('/platform/companies')
```

---

## 不做的事

- 不做月度 limit 设置（目前没有 per-company spending cap 机制，只有 per-department budget）
- 不做 token 级别的 limit（NewAPI gateway 层面有 quota，不在此管理）
- 不做消费明细（已有 audit/调用日志页面）
- 不做分页（公司数 <100，全量返回）

---

## 文件变更预估

| 操作 | 路径 |
|------|------|
| 修改 | `handler/platform/handler.go` — ListCompanies 增强（或新建 companies.go） |
| 新建 | `features/platform/companies/` — 前端 feature module |
| 修改 | `api/platform.ts` — 加 listCompanies + PlatformCompany 类型 |
| 修改 | `config/routes.ts` — 加 /platform/companies 路由 |
| 修改 | `router/routes/platform.ts` — 注册 TanStack route |
| 修改 | `features/platform/index.ts` — barrel export |
| 可能 | `store/billing_repo.go` — 加批量聚合方法（如果逐个太慢） |
