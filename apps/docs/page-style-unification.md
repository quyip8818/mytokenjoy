# TokenJoy 页面体系设计规范

> 目标：让用户在任何页面都感知到"同一个产品"，让开发者新建页面时无需做样式决策。

---

## 1. 设计原则

| 原则 | 释义 |
|------|------|
| 位置可预期 | 标题、操作、筛选、内容永远出现在固定位置，用户无需"找" |
| 状态可感知 | 加载、空数据、错误三种非正常态有一致的视觉语言 |
| 密度可控 | 信息密集型（审计日志）和操作型（设置）使用不同内容密度，但共享同一骨架 |
| 最少概念 | 开发侧只需 PageShell + 2 种内容容器（Card / SplitPanel），无选择困难 |

---

## 2. 页面骨架

```
┌─ AdminLayout ──────────────────────────────────────────────┐
│ Sidebar │ Header                                           │
│         │ ┌─ main (p-8, overflow-auto) ─────────────────┐  │
│         │ │                                             │  │
│         │ │  PageShell (space-y-6)                      │  │
│         │ │    ├─ PageHeader                            │  │
│         │ │    └─ Content (Card 或 SplitPanel)          │  │
│         │ │                                             │  │
│         │ └─────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

**PageShell 只有一种模式**：竖向 `space-y-6` 容器。不再有 layout prop。Split 页面通过 SplitPanel 自身的 `flex-1 min-h-0` 撑满高度。

---

## 3. 两种页面形态

### 3.1 Flow 页面（竖向流式）

```
┌─────────────────────────────────────────────┐
│ PageHeader: title · desc · actions          │
├─────────────────────────────────────────────┤
│ Stats Bar（可选）                            │
├─────────────────────────────────────────────┤
│ Toolbar（可选）: filters · search           │
├─────────────────────────────────────────────┤
│ Card                                        │
│   Content (table / list / form)             │
├─────────────────────────────────────────────┤
│ Pagination（可选）                           │
└─────────────────────────────────────────────┘
```

适用：/keys/provider, /approvals, /models/list, /budget/alerts, /billing, /org/data-source, /audit/*, /me/*, /notifications

### 3.2 Split 页面（主从分屏）

```
┌─────────────────────────────────────────────┐
│ PageHeader（可选）                           │
├────────────┬────────────────────────────────┤
│ Master     │ ContextHeader                  │
│ (Tree/List)│ ────────────────               │
│            │ Detail Content                 │
│  280px     │ (flex-1)                       │
└────────────┴────────────────────────────────┘
```

适用：/dashboard/cost, /dashboard/usage, /keys/platform, /models/routing, /budget, /org/structure, /org/roles

---

## 4. 组件清单

### 4.1 保留并简化的组件

| 组件 | 文件 | 改动 |
|------|------|------|
| **PageShell** | `components/layout/page-shell.tsx` | 重写为单一 `space-y-6` 容器，删除 layout/description/stats/sidebar/leading 全部 slot |
| **DataSection** | `components/layout/data-section.tsx` | 去掉内部 Card 包裹，去掉 title/headerAction，只做 loading→error→empty→children 状态切换 |
| **EmptyState** | `components/ui/empty-state.tsx` | 新增 `variant` prop（prominent/inline/minimal），现有样式作为 prominent |
| **ErrorState** | `components/ui/error-state.tsx` | 保持不变，已满足需求 |
| **PageLoading** | `components/ui/page-loading.tsx` | 保持不变，用于 spinner 场景 |
| **TableSkeleton** | `components/ui/table-skeleton.tsx` | 保持不变，用于 skeleton 场景 |
| **Card** | `components/ui/card.tsx` | 保持不变，Flow 页面内容区容器 |
| **StatCard** | `components/ui/stat-card.tsx` | 保持不变，用于 Stats Bar 区域 |
| **Pagination** | `components/ui/pagination.tsx` | 保持不变 |

### 4.2 新建的组件

| 组件 | 文件 | 职责 |
|------|------|------|
| **PageHeader** | `components/layout/page-header.tsx` | 统一渲染页面标题行 |
| **SplitPanel** | `components/layout/split-panel.tsx` | 统一主从分屏容器 |
| **ContextHeader** | `components/layout/context-header.tsx` | Split 右侧面板的上下文标题 |

### 4.3 删除的组件

| 组件 | 文件 | 替代 |
|------|------|------|
| FilteredPageShell | `components/layout/filtered-page-shell.tsx` | PageShell + Toolbar + Card + Pagination 组合 |
| DashboardPageLayout | `features/dashboard/components/dashboard-page-layout.tsx` | SplitPanel |

### 4.4 保留的工具函数

| 函数 | 文件 | 说明 |
|------|------|------|
| `listEmpty` | `lib/list-empty.ts` | 简化 empty prop 的条件逻辑，继续使用 |

---

## 5. 组件接口定义

### 5.1 PageShell

```tsx
interface PageShellProps {
  children: ReactNode
  className?: string
}

// 实现：
export function PageShell({ children, className }: PageShellProps) {
  return <div className={cn('space-y-6', className)}>{children}</div>
}
```

### 5.2 PageHeader

```tsx
interface PageHeaderProps {
  title: string
  description?: string
  actions?: ReactNode
}

// 实现：
export function PageHeader({ title, description, actions }: PageHeaderProps) {
  return (
    <div className="flex items-start justify-between gap-4">
      <div>
        <h1 className="text-lg font-semibold">{title}</h1>
        {description && <p className="mt-1 text-sm text-muted-foreground">{description}</p>}
      </div>
      {actions && <div className="flex shrink-0 items-center gap-3">{actions}</div>}
    </div>
  )
}
```

### 5.3 SplitPanel

```tsx
interface SplitPanelProps {
  master: ReactNode
  detail: ReactNode
  masterWidth?: number  // default 280
  className?: string
}

// 实现：
export function SplitPanel({ master, detail, masterWidth = 280, className }: SplitPanelProps) {
  return (
    <div className={cn(
      'flex min-h-0 flex-1 overflow-hidden rounded-lg border border-border bg-card shadow-xs',
      className,
    )}>
      <div
        className="shrink-0 overflow-y-auto border-r border-border"
        style={{ width: masterWidth }}
      >
        {master}
      </div>
      <div className="flex min-w-0 flex-1 flex-col overflow-hidden">
        {detail}
      </div>
    </div>
  )
}
```

Split 页面只需 `PageShell` > `SplitPanel`，SplitPanel 自身通过 `flex-1 min-h-0` 吃满 AdminLayout main 区域的剩余高度。

### 5.4 ContextHeader

```tsx
interface ContextHeaderProps {
  breadcrumb?: string[]
  title?: string
  actions?: ReactNode
}

// 实现：
export function ContextHeader({ breadcrumb, title, actions }: ContextHeaderProps) {
  return (
    <div className="flex items-center justify-between border-b border-border px-5 py-3">
      <div className="flex items-center gap-1.5 text-xs">
        {breadcrumb?.map((segment, i) => (
          <Fragment key={i}>
            {i > 0 && <ChevronRight className="size-3 text-muted-foreground" />}
            <span className={i === breadcrumb.length - 1
              ? 'font-medium text-foreground'
              : 'text-muted-foreground'
            }>{segment}</span>
          </Fragment>
        ))}
      </div>
      {actions && <div className="flex items-center gap-3">{actions}</div>}
    </div>
  )
}
```

### 5.5 DataSection（重写）

```tsx
interface DataSectionProps {
  loading?: boolean
  loadingVariant?: 'skeleton' | 'spinner'
  skeletonRows?: number
  skeletonColumns?: number
  error?: Error | null
  onRetry?: () => void
  empty?: EmptyStateProps | null
  children: ReactNode
  className?: string
}

// 实现：纯状态切换，无容器装饰
export function DataSection({
  loading = false,
  loadingVariant = 'skeleton',
  skeletonRows = 5,
  skeletonColumns = 6,
  error = null,
  onRetry,
  empty = null,
  children,
  className,
}: DataSectionProps) {
  if (loading) {
    return loadingVariant === 'spinner'
      ? <PageLoading className={className} />
      : <TableSkeleton rows={skeletonRows} columns={skeletonColumns} />
  }
  if (error) return <ErrorState message={error.message} onRetry={onRetry} className={className} />
  if (empty) return <EmptyState {...empty} className={className} />
  return <div className={className}>{children}</div>
}
```

**不再包裹 Card。** 如果页面需要内容区有卡片边框，在外层用 `<Card><CardContent><DataSection>...</DataSection></CardContent></Card>` 显式包裹。

### 5.6 EmptyState（扩展 variant）

```tsx
interface EmptyStateProps {
  variant?: 'prominent' | 'inline' | 'minimal'
  icon?: LucideIcon
  title: string
  description?: string
  actionLabel?: string
  onAction?: () => void
  className?: string
}
```

| variant | 视觉 | 用途 |
|---------|------|------|
| `prominent`（默认） | 当前样式：圆形 icon + 标题 + 描述 + CTA，居中，py-12 | 首次使用引导 |
| `inline` | 无背景，小 icon 行内，单行文案 | 筛选/搜索无结果 |
| `minimal` | 仅 icon + 一行文字，垂直水平居中，无边框 | Split 右侧未选中 |

---

## 6. 状态展示规范

### 6.1 Loading

| 场景 | loadingVariant | 组件 |
|------|---------------|------|
| Flow 页面表格 | `skeleton` | TableSkeleton |
| Split 左侧 | `spinner` | PageLoading |
| Split 右侧 | `skeleton` | TableSkeleton |
| 有缓存数据刷新 | 不切 loading | TanStack Query stale 数据 |

### 6.2 Error

- 统一用 `ErrorState`：icon + 描述 + 重试按钮
- Split 两侧独立——左侧出错不影响右侧

### 6.3 Empty

- 首次使用 → `prominent`
- 筛选无结果 → `inline`（"调整筛选条件"引导文案）
- Split 未选中 → `minimal`（"选择左侧 XXX 查看详情"）

---

## 7. 间距系统

```
AdminLayout main padding    p-8 (32px)
PageShell 子元素间距        space-y-6 (24px)
Card 内 section 间距        space-y-5 (20px)
Card 内 padding             px-5 py-4
Toolbar 控件间距            gap-3 (12px)
按钮组间距                  gap-3 (12px)
表格行高                    py-3 (12px)
SplitPanel master 宽度      280px
```

---

## 8. 按钮规范

| 场景 | variant | size |
|------|---------|------|
| 页面主操作（每页最多 1 个） | `default` | `sm` |
| 页面次操作 | `outline` | `sm` |
| 三级操作 | `ghost` | `sm` |
| Dialog 底部 | `default` / `outline` | `default` |
| 表格行内 | `ghost` | `icon` 或 `sm` |

创建类按钮带 icon：`<Plus className="size-3.5" /> 创建XXX`

---

## 9. 响应式

| 断点 | 行为 |
|------|------|
| ≥ 1440px | 正常布局 |
| 1280–1440px | SplitPanel master 缩至 240px |
| < 1280px | master 收折为 drawer |
| < 1024px | AdminLayout sidebar 收折 |

---

## 10. 页面归属映射

| 路由 | 形态 | PageHeader | 内容容器 | ContextHeader |
|------|------|-----------|---------|---------------|
| /dashboard/cost | Split | 省略 | SplitPanel | ✓ breadcrumb |
| /dashboard/usage | Split | 省略 | SplitPanel | ✓ |
| /keys/platform | Split | 省略 | SplitPanel | ✓ tab+search |
| /keys/provider | Flow | ✓ | Card > Table | — |
| /approvals | Flow | ✓ | Tabs > Card > Table | — |
| /models/list | Flow | ✓ | Card > Tabs > Table | — |
| /models/routing | Split | 省略 | SplitPanel | ✓ |
| /budget | Split | 省略 | SplitPanel | ✓ breadcrumb |
| /budget/alerts | Flow | ✓ | Stats + Card > Table | — |
| /billing | Flow | ✓ (带 desc) | Stats + Card | — |
| /org/data-source | Flow | ✓ | Card (向导) | — |
| /org/structure | Split | 省略 | SplitPanel | ✓ |
| /org/roles | Split | 省略 | SplitPanel | ✓ |
| /audit/operations | Flow | ✓ | Toolbar + Card > Table + Pagination | — |
| /audit/calls | Flow | ✓ | Toolbar + Card > Table + Pagination | — |
| /me/keys | Flow | ✓ | Card > List | — |
| /me/usage | Flow | ✓ | Toolbar + Card > Table + Pagination | — |
| /me/settings | Flow | ✓ | Card > Form | — |
| /notifications | Flow | ✓ | Card > Matrix | — |

---

## 11. Flow 页面典型代码结构

```tsx
export function BudgetAlertsPageShell({ ... }: Props) {
  return (
    <PageShell>
      <PageHeader title="预警规则" description="设置阈值，超支时通知负责人" actions={<CreateBtn />} />

      <StatsBar stats={stats} />

      <Card>
        <CardContent>
          <BudgetAlertsToolbar ... />
          <DataSection loading={loading} error={error} onRetry={refresh} empty={emptyConfig}>
            <BudgetAlertsTable ... />
          </DataSection>
        </CardContent>
      </Card>
    </PageShell>
  )
}
```

## 12. Split 页面典型代码结构

```tsx
export function BudgetPageShell({ ... }: Props) {
  return (
    <PageShell>
      <SplitPanel
        master={<BudgetTreePanel ... />}
        detail={
          <>
            <ContextHeader breadcrumb={['研发部', '项目A']} actions={<EditBtn />} />
            <div className="min-h-0 flex-1 overflow-y-auto p-5">
              <DataSection loading={loading} error={error} onRetry={refresh}>
                <BudgetDetail ... />
              </DataSection>
            </div>
          </>
        }
      />
    </PageShell>
  )
}
```

---

## 13. 实施任务

| # | 任务 | 涉及文件 |
|---|------|---------|
| 1 | 重写 PageShell → 单一 `space-y-6` 容器 | `components/layout/page-shell.tsx` |
| 2 | 新建 PageHeader | `components/layout/page-header.tsx` |
| 3 | 新建 SplitPanel | `components/layout/split-panel.tsx` |
| 4 | 新建 ContextHeader | `components/layout/context-header.tsx` |
| 5 | 重写 DataSection → 纯状态切换，去掉 Card 包裹 | `components/layout/data-section.tsx` |
| 6 | EmptyState 新增 variant prop | `components/ui/empty-state.tsx` |
| 7 | 删除 FilteredPageShell | 删除文件 |
| 8 | 删除 DashboardPageLayout | 删除文件 |
| 9 | 迁移 7 个 Split 页面 shell | features/{dashboard,keys,models,budget,org} |
| 10 | 迁移 12 个 Flow 页面 shell | features/{approval,audit,billing,keys,models,budget,account,mydashboard,notifications,org} |
| 11 | 更新 `listEmpty` 类型引用（DataSection 接口变了） | `lib/list-empty.ts` |
| 12 | 清理全局 dead imports | 全局 |
