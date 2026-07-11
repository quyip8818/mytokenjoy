# 数据中心页面重设计 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Redesign the Cost Dashboard and Usage Analysis pages with an org-tree sidebar for department-scoped data viewing.

**Architecture:** Each page independently adopts a left-right master-detail layout. A shared `DashboardPageLayout` wraps the org tree sidebar (260px fixed) and content slot. The `?dept=` URL search param drives all API queries. Existing chart/table components are reused with adjusted data sources.

**Tech Stack:** React 19, react-router v7, Zustand, recharts, shadcn/ui, TailwindCSS v4, Vitest

## Global Constraints

- Pure frontend changes only — no backend API modifications
- react-router v7 imports from `'react-router'` (NOT `'react-router-dom'`)
- Path alias: `@/*` → `./src/*`, `@tests/*` → `./tests/*`
- Page hooks use `useInjectedApis(injectedApis?)` pattern for testability
- UI components from `components/ui/` (shadcn), icons from `lucide-react`
- `cn()` utility from `@/lib/utils`
- TailwindCSS v4 (CSS-first, no config file)

---

## File Map

### New Files

| File | Responsibility |
|------|---------------|
| `features/dashboard/hooks/use-dept-selection.ts` | Reads/writes `?dept=` URL param, provides `selectedDeptId` |
| `features/dashboard/hooks/use-org-tree.ts` | Fetches department tree via `departmentApi.getTree()` |
| `features/dashboard/components/org-tree-sidebar.tsx` | Read-only org tree with selection highlight |
| `features/dashboard/components/dashboard-page-layout.tsx` | Left-right split layout shell |
| `features/dashboard/components/dept-comparison-table.tsx` | Sub-department ranking table (replaces drill) |
| `tests/features/dashboard/use-dept-selection.test.ts` | Unit test for URL param hook |
| `tests/features/dashboard/use-org-tree.test.ts` | Unit test for tree fetching |
| `tests/features/dashboard/org-tree-sidebar.test.tsx` | Component test for tree selection |

### Modified Files

| File | Changes |
|------|---------|
| `features/dashboard/hooks/use-cost-dashboard-page.ts` | Remove `DrillState`, accept `deptId`, pass to all API calls |
| `features/dashboard/hooks/use-usage-dashboard-page.ts` | Accept `deptId`, pass to all API calls |
| `features/dashboard/components/cost-dashboard-page-shell.tsx` | Wrap in `DashboardPageLayout`, use new table |
| `features/dashboard/components/usage-dashboard-page-shell.tsx` | Wrap in `DashboardPageLayout` |
| `routes/dashboard/cost.tsx` | Wire `useDeptSelection` + `useOrgTree` into page |
| `routes/dashboard/usage.tsx` | Wire `useDeptSelection` + `useOrgTree` into page |
| `features/dashboard/lib/dashboard.ts` | Remove `DrillState` exports, add breadcrumb builder |
| `tests/features/dashboard/use-cost-dashboard-page.test.ts` | Update for new `deptId` param |
| `tests/features/dashboard/use-usage-dashboard-page.test.ts` | Update for new `deptId` param |

### Barrel Export

| File | Changes |
|------|---------|
| `features/dashboard/index.ts` | Export new hooks and components |

---

### Task 1: `use-dept-selection` Hook

**Files:**
- Create: `apps/frontend/src/features/dashboard/hooks/use-dept-selection.ts`
- Test: `apps/frontend/tests/features/dashboard/use-dept-selection.test.ts`

**Interfaces:**
- Consumes: react-router `useSearchParams`
- Produces: `useDeptSelection(): { selectedDeptId: string | null, setSelectedDeptId: (id: string | null) => void }`

- [ ] **Step 1: Write the failing test**

```ts
// tests/features/dashboard/use-dept-selection.test.ts
import { describe, expect, it } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { useDeptSelection } from '@/features/dashboard/hooks/use-dept-selection'

function wrapper({ children }: { children: React.ReactNode }) {
  return <MemoryRouter initialEntries={['/dashboard/cost']}>{children}</MemoryRouter>
}

function wrapperWithDept({ children }: { children: React.ReactNode }) {
  return <MemoryRouter initialEntries={['/dashboard/cost?dept=d1']}>{children}</MemoryRouter>
}

describe('useDeptSelection', () => {
  it('returns null when no dept param', () => {
    const { result } = renderHook(() => useDeptSelection(), { wrapper })
    expect(result.current.selectedDeptId).toBeNull()
  })

  it('reads dept from URL search param', () => {
    const { result } = renderHook(() => useDeptSelection(), { wrapper: wrapperWithDept })
    expect(result.current.selectedDeptId).toBe('d1')
  })

  it('updates URL when setSelectedDeptId is called', () => {
    const { result } = renderHook(() => useDeptSelection(), { wrapper })
    act(() => {
      result.current.setSelectedDeptId('d2')
    })
    expect(result.current.selectedDeptId).toBe('d2')
  })

  it('clears dept param when set to null', () => {
    const { result } = renderHook(() => useDeptSelection(), { wrapper: wrapperWithDept })
    act(() => {
      result.current.setSelectedDeptId(null)
    })
    expect(result.current.selectedDeptId).toBeNull()
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm -F @tokenjoy/frontend exec vitest run tests/features/dashboard/use-dept-selection.test.ts`
Expected: FAIL — module not found

- [ ] **Step 3: Write implementation**

```ts
// src/features/dashboard/hooks/use-dept-selection.ts
import { useCallback } from 'react'
import { useSearchParams } from 'react-router'

const DEPT_PARAM = 'dept'

export function useDeptSelection() {
  const [searchParams, setSearchParams] = useSearchParams()

  const selectedDeptId = searchParams.get(DEPT_PARAM)

  const setSelectedDeptId = useCallback(
    (deptId: string | null) => {
      setSearchParams(
        (prev) => {
          const next = new URLSearchParams(prev)
          if (deptId) {
            next.set(DEPT_PARAM, deptId)
          } else {
            next.delete(DEPT_PARAM)
          }
          return next
        },
        { replace: true },
      )
    },
    [setSearchParams],
  )

  return { selectedDeptId, setSelectedDeptId }
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm -F @tokenjoy/frontend exec vitest run tests/features/dashboard/use-dept-selection.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add apps/frontend/src/features/dashboard/hooks/use-dept-selection.ts apps/frontend/tests/features/dashboard/use-dept-selection.test.ts
git commit -m "feat(dashboard): add useDeptSelection hook for URL-based dept selection"
```

---

### Task 2: `use-org-tree` Hook

**Files:**
- Create: `apps/frontend/src/features/dashboard/hooks/use-org-tree.ts`
- Test: `apps/frontend/tests/features/dashboard/use-org-tree.test.ts`

**Interfaces:**
- Consumes: `departmentApi.getTree()` via `useInjectedQuery`
- Produces: `useOrgTree(injectedApis?): { departments: Department[], loading: boolean, error: Error | null, getBreadcrumb: (deptId: string | null) => string[] }`

- [ ] **Step 1: Write the failing test**

```ts
// tests/features/dashboard/use-org-tree.test.ts
import { describe, expect, it, vi } from 'vitest'
import { useOrgTree } from '@/features/dashboard/hooks/use-org-tree'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

const mockTree = [
  {
    id: 'd1',
    name: '工程部',
    parentId: null,
    memberCount: 10,
    children: [
      { id: 'd1-1', name: '后端组', parentId: 'd1', memberCount: 5, children: [] },
      { id: 'd1-2', name: '前端组', parentId: 'd1', memberCount: 5, children: [] },
    ],
  },
  { id: 'd2', name: '产品部', parentId: null, memberCount: 8, children: [] },
]

describe('useOrgTree', () => {
  it('fetches department tree on mount', async () => {
    const apis = createMockApis({
      departmentApi: {
        getTree: vi.fn().mockResolvedValue(mockTree),
      },
    })
    const { result } = renderHookWithProviders(() => useOrgTree(apis), { apis })
    await waitForLoaded(result, 'loading')

    expect(result.current.departments).toHaveLength(2)
    expect(result.current.departments[0]?.name).toBe('工程部')
  })

  it('builds breadcrumb path for nested dept', async () => {
    const apis = createMockApis({
      departmentApi: {
        getTree: vi.fn().mockResolvedValue(mockTree),
      },
    })
    const { result } = renderHookWithProviders(() => useOrgTree(apis), { apis })
    await waitForLoaded(result, 'loading')

    expect(result.current.getBreadcrumb('d1-1')).toEqual(['工程部', '后端组'])
    expect(result.current.getBreadcrumb(null)).toEqual(['全公司'])
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm -F @tokenjoy/frontend exec vitest run tests/features/dashboard/use-org-tree.test.ts`
Expected: FAIL

- [ ] **Step 3: Write implementation**

```ts
// src/features/dashboard/hooks/use-org-tree.ts
import { useCallback, useMemo } from 'react'
import type { AppApis } from '@/api/app-apis'
import type { Department } from '@/api/types'
import { queryKeys, useInjectedQuery } from '@/features/query'

function buildPathMap(departments: Department[]): Map<string, string[]> {
  const map = new Map<string, string[]>()

  function walk(nodes: Department[], path: string[]) {
    for (const node of nodes) {
      const currentPath = [...path, node.name]
      map.set(node.id, currentPath)
      if (node.children && node.children.length > 0) {
        walk(node.children, currentPath)
      }
    }
  }

  walk(departments, [])
  return map
}

export function useOrgTree(injectedApis?: AppApis) {
  const {
    data: departments = [],
    loading,
    error,
  } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.org.tree(),
    queryFn: (apis) => apis.departmentApi.getTree(),
  })

  const pathMap = useMemo(() => buildPathMap(departments), [departments])

  const getBreadcrumb = useCallback(
    (deptId: string | null): string[] => {
      if (!deptId) return ['全公司']
      return pathMap.get(deptId) ?? ['全公司']
    },
    [pathMap],
  )

  return { departments, loading, error, getBreadcrumb }
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm -F @tokenjoy/frontend exec vitest run tests/features/dashboard/use-org-tree.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add apps/frontend/src/features/dashboard/hooks/use-org-tree.ts apps/frontend/tests/features/dashboard/use-org-tree.test.ts
git commit -m "feat(dashboard): add useOrgTree hook for department tree fetching"
```

---

### Task 3: `OrgTreeSidebar` and `DashboardPageLayout` Components

**Files:**
- Create: `apps/frontend/src/features/dashboard/components/org-tree-sidebar.tsx`
- Create: `apps/frontend/src/features/dashboard/components/dashboard-page-layout.tsx`
- Test: `apps/frontend/tests/features/dashboard/org-tree-sidebar.test.tsx`

**Interfaces:**
- Consumes: `Department[]` from `useOrgTree`, `selectedDeptId` / `setSelectedDeptId` from `useDeptSelection`
- Produces: `OrgTreeSidebar` (visual tree with click-to-select), `DashboardPageLayout` (left sidebar + right content slot)

- [ ] **Step 1: Write the component test**

```tsx
// tests/features/dashboard/org-tree-sidebar.test.tsx
import { describe, expect, it, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { OrgTreeSidebar } from '@/features/dashboard/components/org-tree-sidebar'
import type { Department } from '@/api/types'

const tree: Department[] = [
  {
    id: 'd1',
    name: '工程部',
    parentId: null,
    memberCount: 10,
    children: [
      { id: 'd1-1', name: '后端组', parentId: 'd1', memberCount: 5, children: [] },
    ],
  },
  { id: 'd2', name: '产品部', parentId: null, memberCount: 8, children: [] },
]

describe('OrgTreeSidebar', () => {
  it('renders root node and top-level departments', () => {
    render(
      <OrgTreeSidebar
        departments={tree}
        selectedDeptId={null}
        onSelect={() => {}}
        loading={false}
      />,
    )
    expect(screen.getByText('全公司')).toBeInTheDocument()
    expect(screen.getByText('工程部')).toBeInTheDocument()
    expect(screen.getByText('产品部')).toBeInTheDocument()
  })

  it('highlights the selected node', () => {
    render(
      <OrgTreeSidebar
        departments={tree}
        selectedDeptId="d1"
        onSelect={() => {}}
        loading={false}
      />,
    )
    const node = screen.getByText('工程部').closest('[role="treeitem"]')
    expect(node).toHaveAttribute('aria-selected', 'true')
  })

  it('calls onSelect when a node is clicked', () => {
    const onSelect = vi.fn()
    render(
      <OrgTreeSidebar
        departments={tree}
        selectedDeptId={null}
        onSelect={onSelect}
        loading={false}
      />,
    )
    fireEvent.click(screen.getByText('产品部'))
    expect(onSelect).toHaveBeenCalledWith('d2')
  })

  it('calls onSelect with null when root is clicked', () => {
    const onSelect = vi.fn()
    render(
      <OrgTreeSidebar
        departments={tree}
        selectedDeptId="d1"
        onSelect={onSelect}
        loading={false}
      />,
    )
    fireEvent.click(screen.getByText('全公司'))
    expect(onSelect).toHaveBeenCalledWith(null)
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm -F @tokenjoy/frontend exec vitest run tests/features/dashboard/org-tree-sidebar.test.tsx`
Expected: FAIL

- [ ] **Step 3: Implement OrgTreeSidebar**

```tsx
// src/features/dashboard/components/org-tree-sidebar.tsx
import { useMemo, useState } from 'react'
import type { Department } from '@/api/types'
import { cn } from '@/lib/utils'
import { ChevronRight, Building2, Users, FolderOpen, Folder } from 'lucide-react'
import { TableSkeleton } from '@/components/ui/table-skeleton'

interface OrgTreeSidebarProps {
  departments: Department[]
  selectedDeptId: string | null
  onSelect: (deptId: string | null) => void
  loading: boolean
}

function TreeNode({
  department,
  level,
  selectedId,
  expandedIds,
  onSelect,
  onToggle,
}: {
  department: Department
  level: number
  selectedId: string | null
  expandedIds: Set<string>
  onSelect: (id: string) => void
  onToggle: (id: string) => void
}) {
  const hasChildren = department.children && department.children.length > 0
  const isSelected = selectedId === department.id
  const isExpanded = expandedIds.has(department.id)

  return (
    <div>
      <div
        role="treeitem"
        tabIndex={0}
        aria-selected={isSelected}
        aria-expanded={hasChildren ? isExpanded : undefined}
        className={cn(
          'group flex items-center gap-2 rounded-md px-2 py-1.5 text-sm cursor-pointer',
          isSelected ? 'bg-primary/8 text-primary font-medium' : 'text-foreground hover:bg-muted',
        )}
        style={{ paddingLeft: `${level * 16 + 8}px` }}
        onClick={() => onSelect(department.id)}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault()
            onSelect(department.id)
          }
        }}
      >
        {hasChildren ? (
          <span
            role="button"
            tabIndex={-1}
            className="flex size-4 shrink-0 items-center justify-center"
            onClick={(e) => {
              e.stopPropagation()
              onToggle(department.id)
            }}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                e.stopPropagation()
                onToggle(department.id)
              }
            }}
          >
            <ChevronRight
              className={cn(
                'size-3.5 text-muted-foreground transition-transform duration-150',
                isExpanded && 'rotate-90',
              )}
            />
          </span>
        ) : (
          <span className="size-4" />
        )}

        {hasChildren ? (
          isExpanded ? (
            <FolderOpen className="size-4 shrink-0 text-muted-foreground" />
          ) : (
            <Folder className="size-4 shrink-0 text-muted-foreground" />
          )
        ) : (
          <Users className="size-4 shrink-0 text-muted-foreground" />
        )}

        <span className="flex-1 truncate">{department.name}</span>
      </div>

      {hasChildren && isExpanded && (
        <div>
          {department.children!.map((child) => (
            <TreeNode
              key={child.id}
              department={child}
              level={level + 1}
              selectedId={selectedId}
              expandedIds={expandedIds}
              onSelect={onSelect}
              onToggle={onToggle}
            />
          ))}
        </div>
      )}
    </div>
  )
}

export function OrgTreeSidebar({
  departments,
  selectedDeptId,
  onSelect,
  loading,
}: OrgTreeSidebarProps) {
  const defaultExpanded = useMemo(
    () => new Set(departments.map((d) => d.id)),
    [departments],
  )
  const [userExpanded, setUserExpanded] = useState<Set<string> | null>(null)
  const expanded = userExpanded ?? defaultExpanded

  const toggleExpand = (id: string) => {
    setUserExpanded((prev) => {
      const current = prev ?? defaultExpanded
      const next = new Set(current)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  return (
    <div className="flex w-[260px] shrink-0 flex-col border-r border-border bg-card">
      <div className="border-b border-border px-4 py-3">
        <span className="text-xs font-medium text-muted-foreground tracking-wide">
          选择查看范围
        </span>
      </div>

      <div
        role="treeitem"
        tabIndex={0}
        aria-selected={!selectedDeptId}
        className={cn(
          'flex cursor-pointer items-center gap-2 border-b border-border px-4 py-2.5 text-sm',
          !selectedDeptId ? 'bg-primary/8 text-primary font-medium' : 'text-foreground hover:bg-muted',
        )}
        onClick={() => onSelect(null)}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault()
            onSelect(null)
          }
        }}
      >
        <Building2 className="size-4 shrink-0 text-muted-foreground" />
        <span>全公司</span>
      </div>

      <div className="flex-1 overflow-y-auto p-2">
        {loading ? (
          <TableSkeleton rows={6} columns={1} />
        ) : (
          departments.map((dept) => (
            <TreeNode
              key={dept.id}
              department={dept}
              level={0}
              selectedId={selectedDeptId}
              expandedIds={expanded}
              onSelect={onSelect}
              onToggle={toggleExpand}
            />
          ))
        )}
      </div>
    </div>
  )
}
```

- [ ] **Step 4: Implement DashboardPageLayout**

```tsx
// src/features/dashboard/components/dashboard-page-layout.tsx
import type { ReactNode } from 'react'

interface DashboardPageLayoutProps {
  sidebar: ReactNode
  children: ReactNode
}

export function DashboardPageLayout({ sidebar, children }: DashboardPageLayoutProps) {
  return (
    <div className="flex min-h-0 flex-1 overflow-hidden">
      {sidebar}
      <div className="flex min-w-0 flex-1 flex-col overflow-y-auto p-6">
        {children}
      </div>
    </div>
  )
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `pnpm -F @tokenjoy/frontend exec vitest run tests/features/dashboard/org-tree-sidebar.test.tsx`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add apps/frontend/src/features/dashboard/components/org-tree-sidebar.tsx apps/frontend/src/features/dashboard/components/dashboard-page-layout.tsx apps/frontend/tests/features/dashboard/org-tree-sidebar.test.tsx
git commit -m "feat(dashboard): add OrgTreeSidebar and DashboardPageLayout components"
```

---

### Task 4: Refactor `use-cost-dashboard-page` to Accept `deptId`

**Files:**
- Modify: `apps/frontend/src/features/dashboard/hooks/use-cost-dashboard-page.ts`
- Modify: `apps/frontend/src/features/dashboard/lib/dashboard.ts` (remove DrillState exports used only internally)
- Modify: `apps/frontend/tests/features/dashboard/use-cost-dashboard-page.test.ts`

**Interfaces:**
- Consumes: `deptId: string | null` from `useDeptSelection`
- Produces: same return shape minus `drill`/`drillTitle`/`canDrillBack`/`handleDrillDept`/`handleDrillBack`; adds `deptId`

- [ ] **Step 1: Update the test for new API**

```ts
// tests/features/dashboard/use-cost-dashboard-page.test.ts
import { describe, expect, it, vi } from 'vitest'
import { useCostDashboardPage } from '@/features/dashboard/hooks/use-cost-dashboard-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useCostDashboardPage', () => {
  it('loads cost summary and builds stats on mount', async () => {
    const summary = {
      totalCost: 1000,
      totalCostMom: 5,
      totalTokens: 2000000,
      totalRequests: 100,
      avgCostPerRequest: 10,
      avgCostPerRequestMom: 0,
      avgCostPerMember: 500,
      avgCostPerMemberMom: 0,
      totalRequestsMom: 0,
    }
    const apis = createMockApis({
      dashboardApi: {
        getCostSummary: vi.fn().mockResolvedValue(summary),
        getDailyCosts: vi.fn().mockResolvedValue([]),
        getDepartmentCosts: vi.fn().mockResolvedValue([]),
        getDepartmentMemberCosts: vi.fn().mockResolvedValue([]),
        getTopConsumers: vi.fn().mockResolvedValue([]),
      },
    })

    const { result } = renderHookWithProviders(
      () => useCostDashboardPage({ deptId: null, injectedApis: apis }),
      { apis },
    )

    await waitForLoaded(result, 'loading')

    expect(apis.dashboardApi.getCostSummary).toHaveBeenCalled()
    expect(result.current.summary).toEqual(summary)
    expect(result.current.stats).toHaveLength(5)
    expect(result.current.stats[0]?.label).toBe('总花费')
  })

  it('passes deptId as parentId to getDepartmentCosts', async () => {
    const apis = createMockApis({
      dashboardApi: {
        getCostSummary: vi.fn().mockResolvedValue({
          totalCost: 0, totalCostMom: 0, totalTokens: 0, totalRequests: 0,
          avgCostPerRequest: 0, avgCostPerRequestMom: 0, avgCostPerMember: 0,
          avgCostPerMemberMom: 0, totalRequestsMom: 0,
        }),
        getDailyCosts: vi.fn().mockResolvedValue([]),
        getDepartmentCosts: vi.fn().mockResolvedValue([]),
        getDepartmentMemberCosts: vi.fn().mockResolvedValue([]),
        getTopConsumers: vi.fn().mockResolvedValue([]),
      },
    })

    const { result } = renderHookWithProviders(
      () => useCostDashboardPage({ deptId: 'd1', injectedApis: apis }),
      { apis },
    )

    await waitForLoaded(result, 'loading')

    expect(apis.dashboardApi.getDepartmentCosts).toHaveBeenCalledWith(
      expect.objectContaining({ parentId: 'd1' }),
    )
  })
})
```

- [ ] **Step 2: Refactor the hook**

Replace `apps/frontend/src/features/dashboard/hooks/use-cost-dashboard-page.ts` with:

```ts
import { useCallback, useMemo, useState } from 'react'
import type { AppApis } from '@/api/app-apis'
import type { CostGranularity, CostPeriod, CostQueryParams } from '@/api/types'
import { COST_GRANULARITY, COST_PERIOD } from '../lib/constants'
import { getMonthStartLocal, getTodayLocal } from '@/lib/date'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { buildCostStats, buildDeptCostsWithColors, COST_CHART_COLORS } from '../lib/dashboard'
import type { CostStatItem } from '../lib/dashboard'

export type { CostStatItem }
export { COST_CHART_COLORS }

interface UseCostDashboardPageOptions {
  deptId: string | null
  injectedApis?: AppApis
}

function buildCostQuery(period: CostPeriod, startDate: string, endDate: string): CostQueryParams {
  if (period === COST_PERIOD.CUSTOM) {
    return { period, startDate, endDate }
  }
  return { period }
}

export function useCostDashboardPage({ deptId, injectedApis }: UseCostDashboardPageOptions) {
  const [period, setPeriod] = useState<CostPeriod>(COST_PERIOD.CURRENT_MONTH)
  const [startDate, setStartDate] = useState(getMonthStartLocal)
  const [endDate, setEndDate] = useState(getTodayLocal)
  const [granularity, setGranularity] = useState<CostGranularity>(COST_GRANULARITY.DAY)

  const costQuery = useMemo(
    () => buildCostQuery(period, startDate, endDate),
    [period, startDate, endDate],
  )

  const { data, loading, error, refresh } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.dashboard.cost(costQuery, deptId, granularity),
    queryFn: async (apis) => {
      const [summary, dailyCosts, deptCosts, topConsumers] = await Promise.all([
        apis.dashboardApi.getCostSummary(costQuery),
        apis.dashboardApi.getDailyCosts({ ...costQuery, granularity }),
        apis.dashboardApi.getDepartmentCosts({
          ...costQuery,
          parentId: deptId ?? undefined,
        }),
        apis.dashboardApi.getTopConsumers({ ...costQuery, limit: 5 }),
      ])
      return { summary, dailyCosts, deptCosts, topConsumers }
    },
  })

  const handlePeriodChange = useCallback((value: string | null) => {
    if (!value) return
    setPeriod(value as CostPeriod)
  }, [])

  const summary = data?.summary ?? null
  const dailyCosts = data?.dailyCosts ?? []
  const topConsumers = data?.topConsumers ?? []
  const deptCosts = data?.deptCosts ?? []

  const deptCostsWithColors = useMemo(
    () => buildDeptCostsWithColors('departments', deptCosts, []),
    [deptCosts],
  )

  const stats = useMemo(() => buildCostStats(summary), [summary])
  const customDateInvalid =
    period === COST_PERIOD.CUSTOM && Boolean(startDate && endDate && startDate > endDate)

  return {
    period,
    startDate,
    endDate,
    granularity,
    customDateInvalid,
    deptId,
    loading,
    error,
    refresh,
    summary,
    dailyCosts,
    topConsumers,
    deptCosts,
    deptCostsWithColors,
    stats,
    handlePeriodChange,
    setStartDate,
    setEndDate,
    setGranularity,
  }
}
```

- [ ] **Step 3: Run test to verify it passes**

Run: `pnpm -F @tokenjoy/frontend exec vitest run tests/features/dashboard/use-cost-dashboard-page.test.ts`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add apps/frontend/src/features/dashboard/hooks/use-cost-dashboard-page.ts apps/frontend/tests/features/dashboard/use-cost-dashboard-page.test.ts
git commit -m "refactor(dashboard): remove DrillState, accept deptId in cost hook"
```

---

### Task 5: Refactor `use-usage-dashboard-page` to Accept `deptId`

**Files:**
- Modify: `apps/frontend/src/features/dashboard/hooks/use-usage-dashboard-page.ts`
- Modify: `apps/frontend/tests/features/dashboard/use-usage-dashboard-page.test.ts`

**Interfaces:**
- Consumes: `deptId: string | null`
- Produces: same return shape plus `deptId`

- [ ] **Step 1: Update the test**

```ts
// tests/features/dashboard/use-usage-dashboard-page.test.ts
import { describe, expect, it, vi } from 'vitest'
import { waitFor } from '@testing-library/react'
import { useUsageDashboardPage } from '@/features/dashboard/hooks/use-usage-dashboard-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useUsageDashboardPage', () => {
  it('loads team and model usage on mount', async () => {
    const apis = createMockApis({
      dashboardApi: {
        getTeamUsage: vi
          .fn()
          .mockResolvedValue([{ departmentId: 'd1', departmentName: 'HQ', quota: 1000, consumed: 500, memberCount: 5, topModel: 'gpt-4' }]),
        getModelUsage: vi.fn().mockResolvedValue([
          { callType: 'gpt-4', modelName: 'GPT-4', tokens: 50, cost: 1, requests: 1, percentage: 100, provider: 'openai' },
        ]),
      },
    })

    const { result } = renderHookWithProviders(
      () => useUsageDashboardPage({ deptId: null, injectedApis: apis }),
      { apis },
    )

    await waitForLoaded(result, 'loading')
    await waitFor(() => {
      expect(result.current.teamUsage).toHaveLength(1)
    })

    expect(apis.dashboardApi.getTeamUsage).toHaveBeenCalled()
    expect(apis.dashboardApi.getModelUsage).toHaveBeenCalled()
  })

  it('passes period params when deptId changes', async () => {
    const apis = createMockApis({
      dashboardApi: {
        getTeamUsage: vi.fn().mockResolvedValue([]),
        getModelUsage: vi.fn().mockResolvedValue([]),
      },
    })

    const { result } = renderHookWithProviders(
      () => useUsageDashboardPage({ deptId: 'd1', injectedApis: apis }),
      { apis },
    )

    await waitForLoaded(result, 'loading')
    expect(apis.dashboardApi.getTeamUsage).toHaveBeenCalledWith(
      expect.objectContaining({ period: 'current_month' }),
    )
  })
})
```

- [ ] **Step 2: Refactor the hook**

```ts
// src/features/dashboard/hooks/use-usage-dashboard-page.ts
import { useCallback, useMemo, useState } from 'react'
import type { AppApis } from '@/api/app-apis'
import type { CostPeriod, CostQueryParams } from '@/api/types'
import { COST_PERIOD } from '../lib/constants'
import { getMonthStartLocal, getTodayLocal } from '@/lib/date'
import { queryKeys, useInjectedQuery } from '@/features/query'

interface UseUsageDashboardPageOptions {
  deptId: string | null
  injectedApis?: AppApis
}

function buildCostQuery(period: CostPeriod, startDate: string, endDate: string): CostQueryParams {
  if (period === COST_PERIOD.CUSTOM) {
    return { period, startDate, endDate }
  }
  return { period }
}

export function useUsageDashboardPage({ deptId, injectedApis }: UseUsageDashboardPageOptions) {
  const [period, setPeriod] = useState<CostPeriod>(COST_PERIOD.CURRENT_MONTH)
  const [startDate, setStartDate] = useState(getMonthStartLocal)
  const [endDate, setEndDate] = useState(getTodayLocal)

  const costQuery = useMemo(
    () => buildCostQuery(period, startDate, endDate),
    [period, startDate, endDate],
  )

  const {
    data: teamUsage = [],
    loading: teamLoading,
    error: teamError,
    refresh: refreshTeam,
  } = useInjectedQuery({
    injectedApis,
    queryKey: [...queryKeys.dashboard.usage(), 'team', costQuery, deptId],
    queryFn: (a) => a.dashboardApi.getTeamUsage(costQuery),
  })

  const {
    data: modelUsage = [],
    loading: modelLoading,
    error: modelError,
    refresh: refreshModel,
  } = useInjectedQuery({
    injectedApis,
    queryKey: [...queryKeys.dashboard.usage(), 'model', costQuery, deptId],
    queryFn: (a) => a.dashboardApi.getModelUsage(costQuery),
  })

  const loading = teamLoading || modelLoading
  const error = teamError ?? modelError
  const refresh = async () => {
    await Promise.all([refreshTeam(), refreshModel()])
  }

  const handlePeriodChange = useCallback((value: string | null) => {
    if (!value) return
    setPeriod(value as CostPeriod)
  }, [])

  const customDateInvalid =
    period === COST_PERIOD.CUSTOM && Boolean(startDate && endDate && startDate > endDate)

  return {
    period,
    startDate,
    endDate,
    customDateInvalid,
    deptId,
    teamUsage,
    modelUsage,
    loading,
    error,
    refresh,
    handlePeriodChange,
    setStartDate,
    setEndDate,
  }
}
```

- [ ] **Step 3: Run tests**

Run: `pnpm -F @tokenjoy/frontend exec vitest run tests/features/dashboard/use-usage-dashboard-page.test.ts`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add apps/frontend/src/features/dashboard/hooks/use-usage-dashboard-page.ts apps/frontend/tests/features/dashboard/use-usage-dashboard-page.test.ts
git commit -m "refactor(dashboard): accept deptId in usage hook, add period selector"
```

---

### Task 6: Create `DeptComparisonTable` and Rewire Cost Dashboard Shell

**Files:**
- Create: `apps/frontend/src/features/dashboard/components/dept-comparison-table.tsx`
- Modify: `apps/frontend/src/features/dashboard/components/cost-dashboard-page-shell.tsx`

**Interfaces:**
- Consumes: `DepartmentCost[]`, `onSelectDept: (id: string) => void`
- Produces: `DeptComparisonTable` component (replaces `CostDrillTable`)

- [ ] **Step 1: Create DeptComparisonTable**

```tsx
// src/features/dashboard/components/dept-comparison-table.tsx
import { DataSection } from '@/components/layout/data-section'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import type { DepartmentCost } from '@/api/types'
import { COST_CHART_COLORS } from '../lib/dashboard'

interface DeptComparisonTableProps {
  deptCosts: DepartmentCost[]
  loading: boolean
  onSelectDept?: (deptId: string) => void
}

export function DeptComparisonTable({
  deptCosts,
  loading,
  onSelectDept,
}: DeptComparisonTableProps) {
  return (
    <DataSection
      title="子部门费用对比"
      loading={loading}
      skeletonColumns={5}
      className="border-border shadow-xs"
    >
      <Table>
        <TableHeader>
          <TableRow className="border-border/50 hover:bg-transparent">
            <TableHead className="w-12 text-xs font-semibold text-muted-foreground">
              排名
            </TableHead>
            <TableHead className="text-xs font-semibold text-muted-foreground">部门</TableHead>
            <TableHead className="text-right text-xs font-semibold text-muted-foreground">
              费用 (¥)
            </TableHead>
            <TableHead className="text-right text-xs font-semibold text-muted-foreground">
              占比
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {deptCosts.map((dept, i) => (
            <TableRow
              key={dept.departmentId}
              className="border-border-subtle hover:bg-muted/50 transition-colors cursor-pointer"
              onClick={() => onSelectDept?.(dept.departmentId)}
            >
              <TableCell>
                <div
                  className="flex h-6 w-6 items-center justify-center rounded-full text-[11px] font-bold text-white"
                  style={{ backgroundColor: COST_CHART_COLORS[i % COST_CHART_COLORS.length] }}
                >
                  {i + 1}
                </div>
              </TableCell>
              <TableCell className="font-medium">{dept.departmentName}</TableCell>
              <TableCell className="text-right font-semibold tabular-nums">
                {dept.cost.toFixed(2)}
              </TableCell>
              <TableCell className="text-right text-muted-foreground tabular-nums">
                {dept.percentage.toFixed(1)}%
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </DataSection>
  )
}
```

- [ ] **Step 2: Rewrite CostDashboardPageShell**

```tsx
// src/features/dashboard/components/cost-dashboard-page-shell.tsx
import { ErrorState } from '@/components/ui/error-state'
import type { useCostDashboardPage } from '../hooks/use-cost-dashboard-page'
import { CostSummaryStats } from './cost-summary-stats'
import { CostTrendChart } from './cost-trend-chart'
import { CostDistributionChart } from './cost-distribution-chart'
import { DeptComparisonTable } from './dept-comparison-table'
import { CostTopConsumersTable } from './cost-top-consumers-table'

interface CostDashboardPageShellProps {
  pageData: ReturnType<typeof useCostDashboardPage>
  onSelectDept?: (deptId: string) => void
}

export function CostDashboardPageShell({ pageData, onSelectDept }: CostDashboardPageShellProps) {
  const {
    loading,
    error,
    refresh,
    stats,
    dailyCosts,
    topConsumers,
    deptCosts,
    deptCostsWithColors,
    granularity,
  } = pageData

  if (error) {
    return <ErrorState message={error.message} onRetry={() => void refresh()} />
  }

  return (
    <div className="space-y-6">
      <CostSummaryStats stats={stats} loading={loading} />
      <div className="grid grid-cols-[5fr_3fr] gap-6">
        <CostTrendChart dailyCosts={dailyCosts} loading={loading} granularity={granularity} />
        <CostDistributionChart data={deptCostsWithColors} loading={loading} />
      </div>
      <DeptComparisonTable
        deptCosts={deptCosts}
        loading={loading}
        onSelectDept={onSelectDept}
      />
      <CostTopConsumersTable topConsumers={topConsumers} loading={loading} />
    </div>
  )
}
```

- [ ] **Step 3: Run build to check types**

Run: `pnpm -F @tokenjoy/frontend build`
Expected: Build succeeds (may have unused-import warnings for old drill references — fix if needed)

- [ ] **Step 4: Commit**

```bash
git add apps/frontend/src/features/dashboard/components/dept-comparison-table.tsx apps/frontend/src/features/dashboard/components/cost-dashboard-page-shell.tsx
git commit -m "feat(dashboard): replace drill table with DeptComparisonTable in cost shell"
```

---

### Task 7: Rewire Usage Dashboard Shell

**Files:**
- Modify: `apps/frontend/src/features/dashboard/components/usage-dashboard-page-shell.tsx`

**Interfaces:**
- Consumes: `ReturnType<typeof useUsageDashboardPage>`, `onSelectDept`
- Produces: Updated shell with stats row + charts + tables layout

- [ ] **Step 1: Rewrite UsageDashboardPageShell**

```tsx
// src/features/dashboard/components/usage-dashboard-page-shell.tsx
import { ErrorState } from '@/components/ui/error-state'
import { DataSection } from '@/components/layout/data-section'
import type { useUsageDashboardPage } from '../hooks/use-usage-dashboard-page'
import { TeamUsageTable } from './team-usage-table'
import { UsageModelChart } from './usage-model-chart'

interface UsageDashboardPageShellProps {
  pageData: ReturnType<typeof useUsageDashboardPage>
  onSelectDept?: (deptId: string) => void
}

export function UsageDashboardPageShell({ pageData, onSelectDept }: UsageDashboardPageShellProps) {
  const { teamUsage, modelUsage, loading, error, refresh } = pageData

  if (error) {
    return <ErrorState message={error.message} onRetry={() => void refresh()} />
  }

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-[5fr_3fr] gap-6">
        <DataSection
          title="团队用量与配额"
          loading={loading}
          skeletonColumns={6}
          className="border-border shadow-xs"
        >
          <TeamUsageTable teamUsage={teamUsage} onSelectDept={onSelectDept} />
        </DataSection>

        <DataSection
          title="模型费用分布"
          loading={loading}
          skeletonColumns={1}
          className="border-border shadow-xs"
        >
          <UsageModelChart modelUsage={modelUsage} />
        </DataSection>
      </div>
    </div>
  )
}
```

Note: `TeamUsageTable` needs a small update to accept `onSelectDept` prop and make rows clickable. Add to `TeamUsageTable`:

```tsx
// Add to TeamUsageTableProps:
onSelectDept?: (deptId: string) => void

// Add to TableRow:
className="... cursor-pointer"
onClick={() => onSelectDept?.(t.departmentId)}
```

- [ ] **Step 2: Update TeamUsageTable to accept onSelectDept**

In `apps/frontend/src/features/dashboard/components/team-usage-table.tsx`, add the `onSelectDept` prop:

```tsx
interface TeamUsageTableProps {
  teamUsage: TeamUsage[]
  onSelectDept?: (deptId: string) => void
}

export function TeamUsageTable({ teamUsage, onSelectDept }: TeamUsageTableProps) {
  // ... existing code, add onClick to TableRow:
  // <TableRow ... onClick={() => onSelectDept?.(t.departmentId)} className="... cursor-pointer">
}
```

- [ ] **Step 3: Run build**

Run: `pnpm -F @tokenjoy/frontend build`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add apps/frontend/src/features/dashboard/components/usage-dashboard-page-shell.tsx apps/frontend/src/features/dashboard/components/team-usage-table.tsx
git commit -m "refactor(dashboard): update usage shell with new layout and clickable rows"
```

---

### Task 8: Rewire Route Pages and Update Barrel Exports

**Files:**
- Modify: `apps/frontend/src/routes/dashboard/cost.tsx`
- Modify: `apps/frontend/src/routes/dashboard/usage.tsx`
- Modify: `apps/frontend/src/features/dashboard/index.ts`

**Interfaces:**
- Consumes: All hooks and components from previous tasks
- Produces: Working pages with org tree sidebar

- [ ] **Step 1: Rewrite cost route page**

```tsx
// src/routes/dashboard/cost.tsx
import {
  CostDashboardPageShell,
  useCostDashboardPage,
  useDeptSelection,
  useOrgTree,
  OrgTreeSidebar,
  DashboardPageLayout,
} from '@/features/dashboard'

export default function CostDashboardPage() {
  const { selectedDeptId, setSelectedDeptId } = useDeptSelection()
  const { departments, loading: treeLoading, getBreadcrumb } = useOrgTree()
  const pageData = useCostDashboardPage({ deptId: selectedDeptId })

  return (
    <DashboardPageLayout
      sidebar={
        <OrgTreeSidebar
          departments={departments}
          selectedDeptId={selectedDeptId}
          onSelect={setSelectedDeptId}
          loading={treeLoading}
        />
      }
    >
      <div className="mb-4">
        <p className="text-xs text-muted-foreground">
          {getBreadcrumb(selectedDeptId).join(' > ')}
        </p>
        <h1 className="text-lg font-semibold">成本看板</h1>
      </div>
      <CostDashboardPageShell pageData={pageData} onSelectDept={setSelectedDeptId} />
    </DashboardPageLayout>
  )
}
```

- [ ] **Step 2: Rewrite usage route page**

```tsx
// src/routes/dashboard/usage.tsx
import {
  UsageDashboardPageShell,
  useUsageDashboardPage,
  useDeptSelection,
  useOrgTree,
  OrgTreeSidebar,
  DashboardPageLayout,
} from '@/features/dashboard'

export default function UsageDashboardPage() {
  const { selectedDeptId, setSelectedDeptId } = useDeptSelection()
  const { departments, loading: treeLoading, getBreadcrumb } = useOrgTree()
  const pageData = useUsageDashboardPage({ deptId: selectedDeptId })

  return (
    <DashboardPageLayout
      sidebar={
        <OrgTreeSidebar
          departments={departments}
          selectedDeptId={selectedDeptId}
          onSelect={setSelectedDeptId}
          loading={treeLoading}
        />
      }
    >
      <div className="mb-4">
        <p className="text-xs text-muted-foreground">
          {getBreadcrumb(selectedDeptId).join(' > ')}
        </p>
        <h1 className="text-lg font-semibold">用量分析</h1>
      </div>
      <UsageDashboardPageShell pageData={pageData} onSelectDept={setSelectedDeptId} />
    </DashboardPageLayout>
  )
}
```

- [ ] **Step 3: Update barrel exports**

Add to `apps/frontend/src/features/dashboard/index.ts`:

```ts
export { useDeptSelection } from './hooks/use-dept-selection'
export { useOrgTree } from './hooks/use-org-tree'
export { OrgTreeSidebar } from './components/org-tree-sidebar'
export { DashboardPageLayout } from './components/dashboard-page-layout'
export { DeptComparisonTable } from './components/dept-comparison-table'
```

Remove old exports that are no longer used externally:
- Remove `ROOT_DRILL`, `DrillState`, `DrillLevel`, `canDrillBack`, `drillBack`, `drillIntoDepartment`, `getDrillTitle`
- Remove `CostDrillTable` (still keep the file for now but remove from barrel)

- [ ] **Step 4: Run full build**

Run: `pnpm -F @tokenjoy/frontend build`
Expected: PASS

- [ ] **Step 5: Run all dashboard tests**

Run: `pnpm -F @tokenjoy/frontend exec vitest run tests/features/dashboard/`
Expected: All tests PASS

- [ ] **Step 6: Commit**

```bash
git add apps/frontend/src/routes/dashboard/ apps/frontend/src/features/dashboard/index.ts
git commit -m "feat(dashboard): wire org tree sidebar into cost and usage pages"
```

---

### Task 9: Clean Up and Final Verification

**Files:**
- Possibly remove: `apps/frontend/src/features/dashboard/components/cost-drill-table.tsx` (if unused)
- Modify: `apps/frontend/src/features/dashboard/lib/dashboard.ts` (remove dead drill code)

- [ ] **Step 1: Check for unused imports and dead code**

Run: `pnpm -F @tokenjoy/frontend build` — if build passes with no errors, the dead code is isolated.

Grep for remaining references to removed exports:
```bash
grep -r "CostDrillTable\|DrillState\|ROOT_DRILL\|drillIntoDepartment\|drillBack\|canDrillBack\|getDrillTitle" apps/frontend/src/ --include="*.ts" --include="*.tsx"
```

If no references remain, delete `cost-drill-table.tsx` and remove drill-related code from `dashboard.ts`.

- [ ] **Step 2: Remove dead drill code from dashboard.ts**

Remove from `apps/frontend/src/features/dashboard/lib/dashboard.ts`:
- `DrillLevel` type
- `DrillState` interface
- `ROOT_DRILL` constant
- `drillIntoDepartment` function
- `drillBack` function
- `getDrillTitle` function
- `canDrillBack` function

Keep: `buildDeptCostsWithColors` (still used), `buildCostStats`, chart colors, format functions.

- [ ] **Step 3: Delete cost-drill-table.tsx if unused**

```bash
rm apps/frontend/src/features/dashboard/components/cost-drill-table.tsx
```

- [ ] **Step 4: Run full verify**

Run: `pnpm verify`
Expected: lint + test + build all PASS

- [ ] **Step 5: Final commit**

```bash
git add -A
git commit -m "chore(dashboard): remove dead drill code and unused components"
```
