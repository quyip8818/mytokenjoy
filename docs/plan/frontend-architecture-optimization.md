# 前端架构优化方案（终态设计）

## 设计原则

1. **最小公共表面** — feature 的 barrel export 只暴露 route page 真正需要的最少符号
2. **lint 即规则** — 所有约束通过 `check-conventions.ts` 机器执行，不依赖人记住
3. **不动已验证的好设计** — DI、query infra、page shell pattern、workflow store 全部保留
4. **一次改到位** — 不做渐进式妥协，每步都朝终态走；但分阶段交付避免大爆炸

---

## 现状架构依赖图（实测）

通过 grep 统计的跨 feature 引用关系：

```
infrastructure (被广泛依赖的横切 feature)
├── query      — 被所有 domain feature 的 hooks 引用
├── session    — 被 models/keys/budget/billing/audit/approval/workflow 引用
└── workflow   — 被 models/keys/budget 引用

domain features (只被 route page 和 workflow 引用)
├── dashboard  — 引用 budget (getBudgetProgressClass, findBudgetNode)
├── keys       — 引用 session/workflow/org/models/budget
├── budget     — 引用 session/workflow/query
├── models     — 引用 session/workflow/query
├── audit      — 引用 session/query
├── org        — 引用 session/query
├── billing    — 引用 session/query
├── mydashboard — 引用 session/query
├── approval   — 引用 session/query
├── notifications — 引用 session/query
└── account    — 引用 session/query/auth/notifications
```

**关键事实**：
- Route pages 只从 feature barrel 引用 `useXxxPage` + `XxxPageShell`（lint 强制）
- Feature 内部的 components/ 通过自己 feature 的 barrel 引用 lib 和 types（lint 强制的 self-barrel 规则）
- Feature 的 hooks/ 通过相对路径引用同 feature 的 lib/（lint 禁止 hooks import self barrel）
- `lib/quota-display.ts` 被 10+ feature 引用 — 它确实是全局工具，放 lib/ 正确
- `lib/provider-labels.ts` 被 3 个 feature 引用（models/workflow/keys）— 留 lib/ 合理

---

## 终态 Barrel Export 规则

### 三级分类

| 级别 | 允许导出的内容 | 谁消费 |
|------|--------------|--------|
| **必须导出** | `use-{page}-page` hook + `{Page}PageShell` + `query-keys` | route page、query 聚合 |
| **按需导出** | 被其他 feature barrel 合法引用的符号 | 其他 domain feature |
| **禁止导出** | feature 内部子组件、内部 lib helper、内部常量 | 无人（内部 relative 或 self-barrel） |

### 判定标准

一个符号是否该出现在 index.ts：

1. 它是 page hook 或 page shell → **必须**
2. 它是 query-keys factory → **必须**（供 `features/query/query-keys.ts` 聚合）
3. 它被另一个 feature 合法引用（grep 可证）→ **允许**，并在 index.ts 中用注释标注消费者
4. 以上都不满足 → **不导出**

### 具体示例：budget/index.ts 终态

```typescript
// === 页面入口（route page 消费）===
export { budgetKeys } from './query-keys'
export { useBudgetPage } from './hooks/use-budget-page'
export { useBudgetAlertRulesPage } from './hooks/use-budget-alert-rules-page'
export { BudgetPageShell } from './components/budget-page-shell'
export { BudgetAlertsPageShell } from './components/budget-alerts-page-shell'

// === 跨 feature 共享（注明消费者）===
// consumed by: dashboard (budget-hero-card)
export { getBudgetProgressClass, getBudgetProgressTone } from './lib/mappers'
// consumed by: dashboard (use-cost-dashboard-page)
export { findBudgetNode } from './lib/mappers'
// consumed by: keys (use-platform-keys-page)
export { BUDGET_INSUFFICIENT_MESSAGE } from './lib/constants'
```

对比现状 35 项 export → 终态 ~8 项。

### dashboard/index.ts 终态

```typescript
export { dashboardKeys } from './query-keys'
export { useCostDashboardRoutePage } from './hooks/use-cost-dashboard-route-page'
export { useUsageDashboardRoutePage } from './hooks/use-usage-dashboard-route-page'
export { CostDashboardLayoutPageShell } from './components/cost-dashboard-layout-page-shell'
export { UsageDashboardLayoutPageShell } from './components/usage-dashboard-layout-page-shell'

// consumed by: workflow/approval-submit
export { MODEL_NOT_IN_DEPT_MESSAGE } from './lib/constants'
```

从 34 项 → ~6 项。

---

## Session Feature：保持不拆

审视后的结论：BillingExchange 的生命周期完全依赖 session 数据（`quotaPerUnit`、`billingCurrency`），没有独立数据源。拆分它只会引入一次间接跳转而不解耦任何东西。

**终态决定**：session 保持原样，包含：
- 认证会话管理
- 权限 hooks
- 路由守卫
- BillingExchange provider
- authz 同步广播

这些都是 "登录后全局可用的会话衍生数据"，内聚度没问题。

---

## Feature 归并

### `trial` → `components/layout/trial-banner.tsx`

理由：
- 只有一个纯 UI banner 组件，无 hooks、无 query、无 state
- 它是 app 壳层面的展示，不是独立业务领域
- 放 `components/layout/` 和 `notification-inbox.tsx` 同级

### `dev` → `features/billing/dev/`

理由：
- SimulateConsume 功能服务于 billing 测试
- 只有 2 个文件
- 归到 billing 下的 dev/ 子目录，prod build tree-shake 无影响

**API 层不动**：`api/dev.ts` 和 `AppApis.devApi` 保持原位不改名。只移动 feature 层文件（hooks + components）。API 层按领域拆分是独立关注点，不跟随 feature 归并。

### `mydashboard` — 不改名

理由：
- rename 代价（改 import path、query-keys、routes lazy path）高
- 收益仅是 "命名好看"
- 当前只有 `routes/me/usage.tsx` 一处引用，认知负担极低

---

## lib/ 治理

### 保留（满足 "无业务语义 + 被 ≥2 feature 引用"）

| 文件 | 引用数 | 理由 |
|------|--------|------|
| `quota-display.ts` | 10+ feature | 货币/额度格式化，全局通用 |
| `provider-labels.ts` | 3 feature | 供应商标签被 models/workflow/keys 引用 |
| `permissions.ts` | session + nav + route-access | 权限判断纯函数 |
| `permission-keys.ts` | 全局 | 权限枚举 |
| `date.ts` | 多处 | 日期格式化 |
| `currency-format.ts` | 多处 | 数字格式化 |
| `csv-export.ts` | audit + dashboard | 通用 CSV 生成 |
| `utils.ts` | 全局 | cn() 等 |
| `labels.ts` | 通用 StatusBadge 样式 | 无领域语义 |
| `api-error-toast.ts` | 多处 | API 错误展示 |
| `list-empty.ts` | layout/data-section | 判空 helper |

### 删除

| 文件 | 处理 |
|------|------|
| `route-access.ts` | 只是 `permissions.ts` 的 1 行 wrapper，inline 到 `session/use-route-access.ts` |

### 保留但明确不下沉

`provider-labels.ts` — 虽然看起来属于 models 领域，但被 3 个 feature 引用（models/workflow/keys），下沉到 models 会逼迫 workflow 和 keys 跨 feature 引用它。放 lib/ 是正确的。

`quota-display.ts` — 同理，10+ feature 引用，绝对属于 lib/。

---

## check-conventions.ts 规则更新

### 新增规则：barrel export 必须有合法消费者

```typescript
// 实现思路：
// 1. 解析每个 features/{domain}/index.ts 的 named export 列表
// 2. 对每个 export，在 src/ 中 grep 其 import 引用
// 3. 判定合法消费者（满足任一即可）：
//    a. 被 feature 外部文件引用（其他 feature、routes/、components/layout/）
//    b. 被自身 feature 的 components/ 引用（现有 lint 强制 components 走 self-barrel）
// 4. 两者都不满足 → 报错（说明这个 export 应该被移除或改为内部相对路径引用）
```

**注意**：自身 feature 的 components/ 通过 self-barrel import lib/types 是现有 lint 强制的合法模式（components 禁止 deep import `./lib/`，必须走 `@/features/{self}`）。因此 "只被自身 components/ 消费" 的 export 仍然合法，不应报错。

这条规则确保 barrel 不会重新膨胀。

### 现有规则确认（不改）

| 规则 | 状态 |
|------|------|
| route page 必须从 `@/features/` import hook | ✓ 保留 |
| route page 必须使用 PageShell + 展开 page hook | ✓ 保留 |
| 禁止跨 feature deep lib import | ✓ 保留 |
| 禁止跨 feature component import | ✓ 保留 |
| hooks/ 禁止 import self barrel | ✓ 保留 |
| components/ 禁止 import hooks/ | ✓ 保留（通过 self barrel 间接引用） |
| components/ 禁止 useApis/useInjectedApis | ✓ 保留 |
| 禁止 `../../` 深层相对导入 | ✓ 保留 |

---

## API Types 策略

**现状**：`api/types/` 定义 API 响应类型，feature 的 index.ts 有时 re-export types。

**终态规则**：

1. `api/types/` = **API 响应 DTO 类型**的唯一定义处
2. Feature 的 barrel **不 re-export** API types（route page 不需要 type annotation，TypeScript 推导足够）
3. Feature 内部的 view model types 放 `features/{domain}/lib/types.ts`，不通过 barrel 导出
4. 需要跨 feature 共享的 pure type（非 API DTO）放 `packages/contracts/`

**验证**：当前所有 route page 确实只做 `<Shell {...useHook()} />`，没有显式 import type。这条规则零成本落地。

---

## Loading/Error/Empty 策略统一

**现状**：`DataSection` 已经做了三态切换。问题是部分 page shell 绕过了 DataSection 自行处理 loading。

**终态规则**（约定，不新增组件）：

1. 所有 page shell 通过 `DataSection` 或 `FilteredPageShell` 托管三态
2. Loading 变体：
   - `loadingVariant="skeleton"` — 用于已知布局的表格/卡片
   - `loadingVariant="spinner"` — 用于动态内容
3. 统一使用已有的 `<TableSkeleton>` / `<Skeleton>` / `<EmptyState>` / `<ErrorState>`
4. **不新增组件**，只统一使用方式

---

## 终态目录结构

```
src/
├── api/
│   ├── client.ts               # fetch, auth refresh, error handling
│   ├── api-events.ts           # typed event bus
│   ├── app-apis.ts             # AppApis interface + defaultApis
│   ├── api-context.ts          # React context
│   ├── context.tsx             # ApiProvider
│   ├── use-apis.ts             # useApis / useInjectedApis
│   ├── types/                  # API 响应 DTO（唯一来源）
│   └── {domain}.ts             # budget.ts, keys.ts, ...
│
├── components/
│   ├── layout/
│   │   ├── admin-layout.tsx
│   │   ├── page-shell.tsx
│   │   ├── filtered-page-shell.tsx
│   │   ├── data-section.tsx
│   │   ├── trial-banner.tsx    # ← 从 features/trial/ 移入
│   │   └── ...
│   └── ui/                     # shadcn 原子（无业务语义）
│
├── config/
│   ├── routes.ts               # 路由定义 + 权限 + lazy + nav group
│   ├── nav.ts
│   ├── app.ts
│   ├── auth.ts
│   └── monitoring.ts
│
├── features/
│   ├── query/                  # TanStack Query infrastructure
│   ├── session/                # 认证 + 权限 + BillingExchange + authz sync
│   ├── workflow/               # workflow panel stack + form workflows
│   ├── notifications/          # 通知
│   ├── approval/
│   ├── audit/
│   ├── billing/
│   │   └── dev/               # ← 从 features/dev/ 移入
│   ├── budget/
│   ├── dashboard/
│   ├── mydashboard/            # 个人用量看板（不改名）
│   ├── keys/
│   ├── models/
│   ├── org/
│   ├── auth/
│   └── account/
│
├── hooks/                      # 全局 UI hooks（极少）
├── lib/                        # 纯工具函数（无业务语义，≥2 feature 引用）
└── routes/                     # 页面入口壳（3-5 行）
```

---

## 每个 Feature 标准结构

```
features/{domain}/
├── index.ts              # barrel：page hook + shell + query-keys + 少量跨 feature 共享符号
├── query-keys.ts         # TanStack Query key factory
├── hooks/
│   ├── use-{page}-page.ts        # 页面编排 hook
│   ├── use-{domain}-queries.ts   # query hooks（可选）
│   └── use-{domain}-actions.ts   # mutation hooks（可选）
├── components/
│   ├── {page}-page-shell.tsx     # 页面 shell
│   └── {sub-component}.tsx       # 内部子组件（不导出）
└── lib/
    ├── types.ts                  # view model types
    ├── constants.ts              # 领域常量
    └── mappers.ts                # 数据转换
```

---

## 实施计划

### Phase 1：Barrel 瘦身 + lint 规则新增

**前置**：运行一次 grep，列出每个 feature 的每个 export 的外部消费者清单。

**执行**：
1. 按上述三级分类瘦身每个 index.ts
2. 被移除的 export 如果有外部引用：
   - 如果引用者在同 feature 内 → 改为相对路径 import（已合规）
   - 如果引用者在其他 feature → 保留在 barrel，注释标注
3. 在 `check-conventions.ts` 增加 "barrel export 必须有外部消费者" 检查（可先为 warning）
4. 验证：`npm run lint` + `npm run build` 通过

**预计改动**：~15 个 index.ts + check-conventions.ts

### Phase 2：Feature 归并

1. `features/trial/` → `components/layout/trial-banner.tsx`（移文件 + 改 import）
2. `features/dev/` → `features/billing/dev/`（移文件 + 改 import + 改 app-apis.ts）
3. `lib/route-access.ts` → inline 到 `features/session/use-route-access.ts` 后删除

**预计改动**：5-6 个文件

### Phase 3：Loading 策略统一

逐页检查 page shell，确保都走 `DataSection` / `FilteredPageShell`。这是纯重构，不改运行时行为。

**预计改动**：~5 个 page shell

---

## 不做的事情

| 提议 | 不做原因 |
|------|---------|
| BillingExchange 从 session 拆出 | 无独立数据源，拆分增加间接性不解耦 |
| mydashboard 改名 me-analytics | rename 代价高，认知收益低 |
| 新增组件库 | 已有 shadcn + DataSection 足够 |
| API types 移到 packages/contracts | API DTO 只有前端消费，移到 monorepo package 增加编译链复杂度 |
| 引入状态管理新方案 | Zustand + TanStack Query 组合覆盖所有场景 |

---

## 验收标准

- [ ] `npm run lint` 通过（含 check-conventions 新规则）
- [ ] `npm run build` 通过
- [ ] `npm run test` 通过
- [ ] 每个 feature 的 index.ts export 数量 ≤ 10（合理例外需注释说明）
- [ ] features/trial/ 和 features/dev/ 目录不再存在
- [ ] lib/route-access.ts 不再存在
- [ ] 零运行时行为变更（UI/功能和优化前完全一致）
