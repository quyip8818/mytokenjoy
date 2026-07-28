# Platform 企业管理看板

> 在 platform admin 中新增企业列表页，展示所有公司的财务概览和钱包状态。

---

## 现状

现有 `GET /api/platform/companies` 返回 `[]store.Company`，只有基础字段（id, name, type, status, billingCurrency, walletRemainQuota, createdAt）。没有：
- 余额（需要换算 quota → 金额）
- 累计充值 / 累计消耗
- 赠送额度 / 透支额度
- 本月消耗
- 成员数

platform admin 目前只能通过 Recharge/Gift/Adjust 操作单个公司，无法一览所有公司的财务状况。

---

## 目标

新增 `/platform/companies` 前端页面 + 一个独立的后端 API，展示企业财务概览。

### 表格列

| 列 | 数据来源 | 说明 |
|----|---------|------|
| 公司名称 | companies.name | |
| 类型 | companies.type | standard / trial / demo |
| 状态 | companies.status | active / suspended，badge 区分 |
| 币种 | companies.billing_currency | CNY |
| 余额 | AggregateWallet → Balances[primary].Balance | 已是金额（paid_amount 维度） |
| 赠送余额 | AggregateWallet → GiftQuota / quotaPerUnit | quota_remaining 需换算，展示剩余可用赠送 |
| 透支额度 | AggregateWallet → OverdraftQuota / quotaPerUnit | quota_remaining 需换算 |
| 累计充值 | AggregateWallet → Balances[primary].TotalTopup | 已是金额 |
| 累计消耗 | AggregateWallet → Balances[primary].TotalConsumed | = topup - balance |
| 本月消耗 | usage_buckets 聚合 | 当月 SUM(cost)，小时级精度 |
| 成员数 | COUNT(members) | status='active' |
| 创建时间 | companies.created_at | |
| 操作 | — | 充值 / 更多（赠送、停用/启用） |

默认按余额升序排列（低余额排前面，方便发现需要充值的公司）。

不做统计卡片、不做搜索、不做分页（公司数 <100）。

---

## 权限控制

### 后端

新端点注册在 `handler/platform/handler.go` 的 `r.Group` 内，自动受已有中间件保护：

```go
r.Use(httpmiddleware.RequireSession(h.protected))
r.Use(httpmiddleware.RequirePlatformAdmin(h.p.Cfg.TokenJoyCompanyID))
```

`RequirePlatformAdmin` 双重校验：
1. session.CompanyID == TokenJoyCompanyID（必须是平台超级公司的成员）
2. session.Permissions 包含 `platform:manage`

无需新增权限 key，复用现有 `platform:manage`。

### 前端

路由定义中加 `requiredPermissions: [PERMISSION.PLATFORM_MANAGE]`，与现有 `/platform/models` 一致。非 platform admin 不可见、不可访问。

---

## API 设计

### `GET /api/platform/companies/overview`（新建独立端点）

不修改现有 `ListCompanies`，新建独立端点返回带财务聚合的列表。原因：现有 `ListCompanies` 被其他流程使用，保持轻量；聚合查询可独立优化。

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
// handler/platform/companies_overview.go
func (h *Handler) CompaniesOverview(w http.ResponseWriter, r *http.Request) {
    companies := h.p.CompanySvc.ListCompanies(ctx)
    // 批量查询（避免 N+1）：
    //   1. 逐个 AggregateWallet（已有方法，接受 companyID 参数，无需 company.WithContext）
    //   2. 批量 monthlySpend：usage_buckets 按时间范围聚合（利用分区裁剪 + idx_usage_buckets_time）
    //   3. 批量 memberCount：SELECT company_id, COUNT(*) FROM members WHERE status='active' GROUP BY company_id
    // 换算：giftQuota / quotaPerUnit, overdraftQuota / quotaPerUnit（需先 GetCurrency，全局只查一次）
    // ponytail: AggregateWallet 逐个调用，公司 <100 可接受；超 200 家需改为批量 SQL
}
```

### 数据来源映射

| 字段 | 来源 | 说明 |
|------|------|------|
| balance | `AggregateWallet(companyID).Balances[primary].Balance` | 已是金额（paid_amount 维度），直接用 |
| giftBalance | `AggregateWallet(companyID).GiftQuota` / quotaPerUnit | GiftQuota 是 quota_remaining 之和，换算为金额 |
| overdraft | `AggregateWallet(companyID).OverdraftQuota` / quotaPerUnit | 同上，展示剩余可用透支 |
| totalTopup | `AggregateWallet(companyID).Balances[primary].TotalTopup` | 已是金额 |
| totalConsumed | `AggregateWallet(companyID).Balances[primary].TotalConsumed` | = topup - balance |
| monthlySpend | **PlatformQueryRepo.SumMonthlyCost**（见下） | 基于 usage_buckets 聚合 |
| memberCount | **PlatformQueryRepo.CountActiveMembers**（见下） | 跨公司聚合 |

**注意事项：**
- `AggregateWallet` 直接接受 `companyID uuid.UUID` 参数，无需设置 company context
- `quotaPerUnit` 通过 `store.Billing().GetCurrency(ctx, "CNY")` 获取，全局只查一次（所有公司都是 CNY）

### 跨公司查询：新增 PlatformQueryRepository

现有 BillingRepository / OrgRepository 全部 company-scoped。跨公司聚合查询应放在独立 interface 中，
避免污染现有职责边界，同时明确只在 platform handler 中使用。

```go
// store/platform_query_repo.go
type PlatformQueryRepository interface {
    SumMonthlyCost(ctx context.Context, from, to time.Time) (map[uuid.UUID]float64, error)
    CountActiveMembers(ctx context.Context) (map[uuid.UUID]int, error)
}
```

实现：

```go
// store/postgres/platform_query_repo.go
type platformQueryRepo struct{ db Pool }

func (r *platformQueryRepo) SumMonthlyCost(ctx context.Context, from, to time.Time) (map[uuid.UUID]float64, error) {
    rows, err := r.db.Query(ctx, `
        SELECT company_id, COALESCE(SUM(cost), 0)
        FROM usage_buckets
        WHERE bucket_start >= $1 AND bucket_start < $2
        GROUP BY company_id
    `, from, to)
    // ...scan into map
}

func (r *platformQueryRepo) CountActiveMembers(ctx context.Context) (map[uuid.UUID]int, error) {
    rows, err := r.db.Query(ctx, `
        SELECT company_id, COUNT(*)
        FROM members
        WHERE status = 'active'
        GROUP BY company_id
    `)
    // ...scan into map
}
```

**为什么用 usage_buckets 而不是 usage_ledger：**
- `usage_ledger` 是按 `occurred_at` 分区的大表，按 `period_key` 聚合需跨分区扫描
- `usage_buckets` 是预聚合表（小时粒度），有 `idx_usage_buckets_time` 索引，数据量小得多
- 月度汇总精度完全够用，规避了不同公司 period_key 多样性问题

时间范围构建：
```go
now := time.Now().UTC()
from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
to := from.AddDate(0, 1, 0)
```

这些方法只在 platform handler 中使用，SaaS 模式下才挂载（`router.go` 中 `if cfg.SupportSaas` 才 Mount platform handler），私有化版本不会访问到。

---

## 前端

### 路由

`/platform/companies`，navGroup "平台管理"，与"模型目录"同组。

```ts
{
  key: 'platformCompanies',
  path: '/platform/companies',
  label: '企业管理',
  icon: Building2,
  requiredPermissions: [PERMISSION.PLATFORM_MANAGE],
  lazy: () => import('@/routes/platform/companies'),
  navGroup: '平台管理',
}
```

### 页面功能

1. **表格**：上述列，默认按余额升序
2. **操作**：
   - 充值（主按钮，弹窗输入金额 → `POST /platform/companies/:id/recharge`）
   - 更多菜单：赠送（`POST /platform/companies/:id/gift`）、停用/启用（`PATCH /platform/companies/:id`）

### 代码组织

```
features/platform/companies/
  hooks/use-platform-companies-page.ts
  components/platform-companies-page-shell.tsx
  index.ts
```

API 加在 `api/platform.ts` 中：
```ts
companiesOverview: () => request<PlatformCompanyOverview[]>('/platform/companies/overview')
```

---

## 不做的事

- 不做统计卡片（公司少，表格一眼扫完）
- 不做搜索/过滤（公司 <100）
- 不做分页
- 不做消费趋势/sparkline
- 不做消费明细（已有 audit/调用日志页面）
- 不做月度 limit 设置
- 不做实时数据（聚合数据可有合理缓存）

---

## 文件变更预估

| 操作 | 路径 |
|------|------|
| 新建 | `internal/store/platform_query_repo.go` — PlatformQueryRepository interface |
| 新建 | `internal/store/postgres/platform_query_repo.go` — 实现 SumMonthlyCost / CountActiveMembers |
| 新建 | `internal/http/handler/platform/companies_overview.go` — CompaniesOverview 端点 |
| 修改 | `internal/http/handler/platform/handler.go` — 注册 `r.Get("/companies/overview", h.CompaniesOverview)` |
| 修改 | `internal/http/deps/platform.go` — 注入 PlatformQueryRepository |
| 新建 | `features/platform/companies/` — 前端 feature module |
| 修改 | `api/platform.ts` — 加 companiesOverview + PlatformCompanyOverview 类型 |
| 修改 | `config/routes.ts` — 加 /platform/companies 路由 |
| 修改 | `features/platform/index.ts` — barrel export |
