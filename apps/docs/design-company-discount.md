# 企业模型优惠（Discount）配置

> 状态：已实现  
> 作者：Kiro  
> 关联代码：`internal/store/model_discount.go`、`internal/domain/usage/discount.go`、`internal/http/handler/platform/pricing.go`、`internal/http/handler/billing/handler.go`

---

## 1. 背景

平台管理员需要针对特定公司配置模型调用折扣（discount coefficient），使该公司在消费时按优惠系数计算额度扣减。

**已有后端能力（无需新增）：**
- `model_discount` 表：append-only timeline，每行 `(company_id, model_type, discount, effective_from, note)`
- `ListCompanyDiscounts` / `SetCompanyDiscount` HTTP handler（`/platform/companies/{id}/discounts`）
- `ApplyDiscount(entry, discounts)` — exact match > wildcard `*` > 1.0

**缺失：**
- 前端 API 接入
- 平台管理员在「企业管理」页面配置优惠的 UI
- 企业侧查看自身优惠的只读页面

---

## 2. 目标

| # | 目标 | 说明 |
|---|------|------|
| G1 | 平台管理员可在企业管理页面为指定公司批量配置模型折扣 | SaaS-only，DropdownMenu 新增「优惠」项 |
| G2 | 企业侧可查看自身当前生效的优惠列表 | SaaS + Local 均可见（前提：有优惠数据） |
| G3 | 最小改动，复用已有 handler 和 store | 新增 `sync_versions` 表统一管理 catalog sync version |

---

## 3. 交互设计

### 3.1 平台侧：配置入口（SaaS-only）

**位置**：`/platform/companies` 页面 → 每行操作列 → DropdownMenu → 新增「优惠」菜单项

**交互流程**：
1. 点击「优惠」→ 打开 Sheet（右侧滑出面板）
2. Sheet 标题：`{公司名称} — 模型优惠配置`
3. Sheet 内容：
   - 当前已配置的优惠列表（Table）
   - 底部表单区：选择模型 + 输入折扣系数 + 备注 → 提交

**Sheet 布局**：

```
┌─────────────────────────────────────────┐
│  {公司名} — 模型优惠配置          [×]  │
├─────────────────────────────────────────┤
│  当前优惠                               │
│  ┌──────────┬──────────┬──────────────┐ │
│  │ 模型类型  │ 折扣系数  │ 生效时间     │ │
│  ├──────────┼──────────┼──────────────┤ │
│  │ gpt-4o   │ 0.80     │ 2025-06-01   │ │
│  │ *        │ 0.90     │ 2025-05-15   │ │
│  └──────────┴──────────┴──────────────┘ │
│                                         │
│  ── 新增/更新优惠 ──                     │
│  模型: [ Select / 输入 model_type     ] │
│  系数: [ 0.80 ] (0-1 打折, >1 加价)     │
│  备注: [ 合同优惠 ]                      │
│  [ 保存 ]                               │
└─────────────────────────────────────────┘
```

**模型选择**：
- 下拉列表数据来源 = platform catalog 中所有 model_type + 额外 `*`（通配符）选项
- 支持手动输入自定义 model_type（Combobox）

**折扣系数说明**：
- `0.8` = 八折（实际扣费 × 0.8）
- `1.0` = 无优惠
- `1.2` = 加价 20%（理论支持，UI 上不限制）

### 3.2 企业侧：查看优惠（SaaS + Local）

**UX 决策**：将优惠信息放在「钱包管理」（`/billing`）页面中。

**理由**：
1. 优惠直接影响扣费金额，属于财务/计费域
2. 企业用户自然会在「钱包」场景下关注"为什么我扣费比标准定价少"
3. 不需要独立路由——作为 billing 页面的一个 section 即可
4. 条件渲染：只有公司存在 discount 数据时才展示该 section（无 discount 时完全不可见）

**布局**：在 `BillingPageShell` 的 `BillingStats` 下方、`RechargePanel` 上方插入：

```
┌─────────────────────────────────────────┐
│  💰 钱包管理                             │
│  ┌ BillingStats ──────────────────────┐ │
│  │ 余额 / 赠送余额 / ...              │ │
│  └────────────────────────────────────┘ │
│                                         │
│  ┌ DiscountSection (条件渲染) ────────┐ │
│  │ 当前优惠                           │ │
│  │ ┌──────────┬──────────┬─────────┐  │ │
│  │ │ 模型类型  │ 折扣系数  │ 说明    │  │ │
│  │ │ gpt-4o   │ 0.80 (8折)│ 合同   │  │ │
│  │ │ *        │ 0.90 (9折)│ 通用   │  │ │
│  │ └──────────┴──────────┴─────────┘  │ │
│  │ * = 未单独配置的模型统一适用         │ │
│  └────────────────────────────────────┘ │
│                                         │
│  ┌ RechargePanel ─────────────────────┐ │
│  └────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

**为什么不独立路由**：
- 不符合 "无数据则不可见" 的要求——路由注册必须静态；而 section 可以按数据条件渲染
- 避免增加导航噪音，优惠是低频查看信息
- 放在 billing 上下文中，用户看余额时自然看到优惠

---

## 4. 数据流

### 4.1 平台管理员配置

```
PlatformCompaniesPage
  └─ openDiscount(company) → Sheet
       └─ GET /platform/companies/{id}/discounts   → 列表
       └─ PUT /platform/companies/{id}/discounts   → 新增/更新
            body: { modelType, discount, note }
```

### 4.2 企业侧查看

```
BillingPage
  └─ GET /api/billing/discounts → 返回当前公司的 discount 列表
       (后端根据 session.company_id 查 model_discount 表)
```

**新增后端 endpoint**：

| Method | Path | 权限 | 说明 |
|--------|------|------|------|
| GET | `/api/billing/discounts` | `billing:read` | 返回当前公司的有效 discount 列表 |

> 注：平台侧的 `/platform/companies/{id}/discounts` 已存在，GET + POST 均可用。

---

## 5. 前端代码结构

### 5.1 platform feature 变更

```
features/platform/companies/
├── components/
│   ├── platform-companies-page-shell.tsx  (修改: DropdownMenu 加「优惠」)
│   └── discount-sheet.tsx                 (新增: Sheet 组件)
├── hooks/
│   ├── use-platform-companies-page.ts     (修改: 加 discount 状态)
│   └── use-company-discounts.ts           (新增: 拉取/提交 discount)
└── index.ts                               (修改: 导出新组件/hook)
```

### 5.2 billing feature 变更

```
features/billing/
├── components/
│   ├── billing-page-shell.tsx             (修改: 插入 DiscountSection)
│   └── discount-section.tsx               (新增: 企业侧只读优惠列表)
├── hooks/
│   └── use-billing-page.ts                (修改: 加载 discount 数据)
└── index.ts                               (修改: 导出)
```

### 5.3 API 层

```typescript
// apps/frontend/src/api/platform.ts — 新增
listCompanyDiscounts: (companyId: string) =>
  request<DiscountEntry[]>(`/platform/companies/${companyId}/discounts`),

setCompanyDiscount: (companyId: string, data: SetDiscountInput) =>
  request<void>(`/platform/companies/${companyId}/discounts`, {
    method: 'PUT',
    body: JSON.stringify(data),
  }),

// apps/frontend/src/api/app-apis.ts（或 billing API 文件）— 新增
myDiscounts: () => request<DiscountEntry[]>('/billing/discounts'),
```

### 5.4 类型

```typescript
interface DiscountEntry {
  modelType: string   // 具体模型或 "*"
  discount: number    // 0.8 = 八折
  note?: string
}

interface SetDiscountInput {
  modelType: string
  discount: number
  note?: string
}
```

---

## 6. 后端变更

| 文件 | 变更 | 状态 |
|------|------|------|
| `store/postgres/schema.sql` | 新增 `sync_versions` 表 | ✅ 已完成 |
| `store/sync_versions.go` | 新接口 `SyncVersionRepository` | ✅ 已完成 |
| `store/postgres/sync_versions_repo.go` | PG 实现 | ✅ 已完成 |
| `http/handler/platform/handler.go` | 注册 `/sync/catalog/discounts` 路由 | ✅ 已完成 |
| `http/handler/platform/catalog_discounts.go` | 新增 `CatalogDiscounts` sync endpoint | ✅ 已完成 |
| `http/handler/platform/pricing.go` | `SetCompanyDiscount` 加 per-company version bump | ✅ 已完成 |
| `http/handler/billing/handler.go` | 新增 `GetDiscounts` handler | ✅ 已完成 |
| `http/handler/billing/handler.go` route | 注册 `GET /api/billing/discounts` | ✅ 已完成 |

> 完整变更清单（含全局 version 迁移）见 `apps/docs/design-per-company-discount-version.md` 第 5 节。

**企业侧 handler 逻辑**：
```go
func (h *Handler) GetMyDiscounts(w http.ResponseWriter, r *http.Request) {
    companyID := session.CompanyID(r.Context())
    rows, err := h.modelDiscount.CurrentDiscounts(r.Context(), companyID)
    // ... map to DTO, respond
}
```

权限：复用 `billing:read` 中间件组。

---

## 7. 权限总结

| 操作 | 所需权限 | 环境 |
|------|---------|------|
| 配置公司优惠 | `platform:manage` | SaaS only |
| 查看自身优惠 | `billing:read` | SaaS + Local |

---

## 8. 边界情况

1. **公司无任何 discount** → 企业侧 billing 页面不展示 DiscountSection（条件渲染 `discounts.length > 0`）
2. **通配符 `*`** → UI 展示为"其他所有模型"
3. **重复提交同模型** → append-only，新行覆盖旧行（`CurrentDiscounts` 已处理取最新）
4. **discount = 1.0** → 语义上等于"取消优惠"，UI 可提示但不阻止
5. **Local 部署** → `/platform/*` 路由不可见（无 `platform:manage` 权限），但 billing 页面的 discount section 仍可渲染（数据通过 billing API 读取）

---

## 9. SaaS ↔ Local 同步机制

### 同步链路

```
SaaS 平台管理员
  └─ SetCompanyDiscount → model_discount.Insert
       └─ SyncVersions.Increment(companyID, "discounts")  ← per-company version bump
                                        ↓
Local 实例 (catalogsync worker, 定时轮询)
  └─ FetchVersions(带 sync token) → SaaS 查 sync_versions 返回该公司的 discounts version
       └─ version 不同 → FetchDiscounts() → syncDiscounts() → 写入本地 model_discount
```

### 关键设计

- Discount version 存在 `sync_versions` 表（`company_id + type = "discounts"`）
- 全局 version（models/pricing/currencies）也统一存在此表（`company_id = uuid.Nil`）
- 不再使用 `system_settings` 存储 catalog sync version
- 详细设计见 `apps/docs/design-per-company-discount-version.md`

### 同步保证

- **最终一致**：Local catalogsync worker 按配置间隔轮询，非实时
- **幂等写入**：SaaS 下发每条 discount 的 UUID id，Local 用 `INSERT ... ON CONFLICT (id) DO NOTHING` 实现幂等。重复 sync 不会产生冗余行
- **单向**：SaaS → Local，Local 不回写 discount（只读消费）
- **统一模式**：与 currencies / lots / orders sync 统一使用 id 幂等模式

---

## 10. 不做的事

- ❌ 不新增独立路由页面（如 `/billing/discounts`）
- ❌ 不支持 discount 删除（append-only 设计，设 1.0 即为取消）
- ❌ 不做 discount 历史变更记录 UI（后端已存储，但 v1 不暴露）
- ❌ 不做 effective_from 未来生效时间 UI（v1 立即生效）
