# E2E 测试规范：data-testid 体系

> 状态：基础设施已就绪，spec 迁移进行中

---

## 1. 原则

1. **关键交互路径用 `data-testid`** — 稳定锚点，不受文案/布局变更影响
2. **ARIA 语义 locator 可保留** — `getByRole('dialog')`、`getByRole('tab')`、`getByRole('treeitem')` 是真正的语义标记，不 flaky
3. **禁止 CSS class selector** — `locator('[class*="cursor-pointer"]')` 极脆弱
4. **禁止裸 `.first()` / `.nth()`** — 除非配合 testid scope

---

## 2. testId 命名规范

格式：`{layer}-{identifier}`

### 三层体系

| 层 | 前缀 | 示例 | 用途 |
|---|---|---|---|
| **Page** | `page-` | `page-billing`, `page-budget-alerts` | 页面存在性断言 |
| **Section** | `{page}-` | `billing-stats`, `billing-discount-section` | 区域可见性 |
| **Action** | `{page}-` | `budget-alerts-create`, `platform-companies-discount-{id}` | 操作触发 |
| **Nav** | `nav-` | `nav-billing`, `nav-group-预算与财务` | 导航定位 |

### 命名规则

- 全小写 kebab-case（nav-group 值为中文时除外）
- Page testid 来源于 `config/routes.ts` 的 `key` 字段
- 动态 ID 用 suffix：`platform-companies-discount-${co.id}`
- 只标注 e2e 需要定位的元素

### 完整 testId 清单

| testId | 组件 | 描述 |
|--------|------|------|
| **Layout** | | |
| `nav-{path}` | `sidebar.tsx` NavItem Link | 侧栏导航项（path 去首 `/` 并用 `-` 连接） |
| `nav-group-{name}` | `sidebar.tsx` group button | 侧栏分组按钮（中文名） |
| **Pages（通过 PageHeader testId prop）** | | |
| `page-billing` | BillingPageShell | 钱包管理页 |
| `page-budget` | BudgetPageShell | 预算管理页 |
| `page-budget-alerts` | BudgetAlertsPageShell | 预警规则页 |
| `page-dashboard-cost` | CostDashboardLayoutPageShell | 成本看板页 |
| `page-dashboard-usage` | UsageDashboardLayoutPageShell | 用量分析页 |
| `page-keys-platform` | PlatformKeysPageShell | Key 管理页 |
| `page-keys-provider` | ProviderKeysPageShell | 供应商 Key 页 |
| `page-models-list` | ModelListPageShell | 模型列表页 |
| `page-models-routing` | ModelRoutingPageShell | 模型配置页 |
| `page-org-data-source` | DataSourcePageShell | 数据源页 |
| `page-org-structure` | StructurePageShell | 组织架构页 |
| `page-org-roles` | RolesPageShell | 角色管理页 |
| `page-audit-operations` | OperationsLogPageShell | 操作审计页 |
| `page-audit-calls` | CallLogsPageShell | 调用日志页 |
| `page-approval` | ApprovalPageShell | 审批中心页 |
| `page-me-keys` | MemberKeysPageShell | 我的 Key 页 |
| `page-me-usage` | MyCallLogsPageShell | 我的用量页 |
| `page-me-settings` | SettingsPageShell | 设置页 |
| `page-platform-models` | PlatformModelsPageShell | 平台模型目录页 |
| `page-platform-companies` | PlatformCompaniesPageShell | 平台企业管理页 |
| `page-platform-currencies` | PlatformCurrenciesPageShell | 平台汇率管理页 |
| **Sections** | | |
| `billing-stats` | BillingStats | 钱包统计区 |
| `billing-discount-section` | DiscountSection | 企业侧优惠列表 |
| `billing-recharge-panel` | RechargePanel | 充值面板 |
| `billing-records-table` | RechargeRecordsTable | 充值记录表 |
| `audit-export-btn` | AuditListToolbar | 导出 CSV 按钮 |
| `dashboard-cost-stats` | CostStatsCards | 成本统计卡片 |
| `dashboard-cost-chart` | DailyCostChart | 每日花费图表 |
| **Actions** | | |
| `budget-alerts-create` | Button in BudgetAlertsPageShell | 创建预警规则 |
| `me-keys-create` | Button in MemberKeysPageShell | 新建 Key |
| `me-keys-apply-budget` | Button in MemberKeysPageShell | 申请额度 |
| `platform-models-sync` | Button in PlatformModelsPageShell | 同步模型 |
| `platform-models-add` | Button in PlatformModelsPageShell | 添加模型 |
| `platform-models-publish` | Button in PlatformModelsPageShell | 发布 |
| `platform-companies-discount-{id}` | DropdownMenuItem | 优惠操作 |
| `platform-companies-gift-{id}` | DropdownMenuItem | 赠送操作 |
| `discount-sheet` | SheetContent in DiscountSheet | 优惠配置面板 |
| `discount-submit` | Button in DiscountSheet | 保存优惠 |

---

## 3. 实施方式

### PageHeader

```tsx
// page-header.tsx 已支持 testId prop
<PageHeader testId="page-billing" title="钱包管理" />
```

### Sidebar Nav

```tsx
// sidebar.tsx 已自动生成 testId
// 展开时: data-testid="nav-billing"（来自 /billing）
// 折叠时: data-testid="nav-billing"（同上）
// 分组按钮: data-testid="nav-group-预算与财务"
```

### Section / Action

```tsx
// 直接在组件上加
<div data-testid="billing-discount-section">...</div>
<Button data-testid="budget-alerts-create">创建规则</Button>
<DropdownMenuItem data-testid={`platform-companies-discount-${co.id}`}>优惠</DropdownMenuItem>
```

---

## 4. E2E spec 写法

```typescript
// ✅ 正确
await expect(page.getByTestId('page-billing')).toBeVisible()
await page.getByTestId('nav-group-组织与权限').click()
await page.getByTestId('nav-org-structure').click()
await expect(page.getByTestId('page-org-structure')).toBeVisible()

// ✅ ARIA 语义可保留
await expect(page.getByRole('dialog')).toBeVisible()
await page.getByRole('tab', { name: '待审批' }).click()
await page.getByRole('treeitem', { name: /总公司/ }).click()

// ❌ 禁止
await page.locator('[class*="cursor-pointer"]').first().click()
await page.getByRole('heading', { name: '钱包管理' }).toBeVisible()  // 多个同名 heading
await page.locator('button[class*="h-8 w-8"]').first().click()
```

---

## 5. 迁移进度

### 基础设施（✅ 已完成）

- [x] `PageHeader` 支持 `testId` prop
- [x] Sidebar `NavItem` Link 自动加 `data-testid="nav-{path}"`
- [x] Sidebar group button 自动加 `data-testid="nav-group-{name}"`

### 组件打标（进行中）

- [ ] billing feature 页面组件
- [ ] platform feature 页面组件
- [ ] org/budget/keys/models/dashboard/audit/approval/me 页面组件
- [ ] auth 组件

### Spec 重写（待做）

- [ ] smoke.spec.ts + navigation.spec.ts
- [ ] wallet.spec.ts + discount.spec.ts
- [ ] budget.spec.ts + budget-org-member-picker.spec.ts
- [ ] dashboard-cost.spec.ts + data-source.spec.ts
- [ ] keys-self-service.spec.ts
- [ ] org-structure.spec.ts + org-roles.spec.ts
- [ ] approval.spec.ts + audit-export.spec.ts
- [ ] auth.spec.ts + member-portal.spec.ts + member-delete-no-accumulate.spec.ts
- [ ] feishu-import.spec.ts + sms-sync-models.spec.ts

---

## 6. 不做的事

- ❌ 不给所有 UI 元素加 testid（只标注 e2e 需要的）
- ❌ 不在 ui/ 组件库里加 testId prop（让使用方按需加）
- ❌ 不用 `id` 属性代替 `data-testid`（id 有全局唯一性约束）
- ❌ 不删除合理的 ARIA locator（`getByRole('dialog')` 等）
