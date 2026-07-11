import { CostDashboardPageShell } from './cost-dashboard-page-shell'
import { DashboardDateRangePicker } from './dashboard-date-range-picker'
import { DashboardPageLayout } from './dashboard-page-layout'
import { OrgTreeSidebar } from './org-tree-sidebar'
import type { useCostDashboardRoutePage } from '../hooks/use-cost-dashboard-route-page'

type CostDashboardLayoutPageShellProps = ReturnType<typeof useCostDashboardRoutePage>

export function CostDashboardLayoutPageShell({
  selectedDeptId,
  setSelectedDeptId,
  departments,
  treeLoading,
  getBreadcrumb,
  pageData,
}: CostDashboardLayoutPageShellProps) {
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
      <div className="mb-4 flex items-center justify-between">
        <div>
          <p className="text-xs text-muted-foreground">
            {getBreadcrumb(selectedDeptId).join(' > ')}
          </p>
          <h1 className="text-lg font-semibold">成本看板</h1>
        </div>
        <DashboardDateRangePicker value={pageData.period} onChange={pageData.handlePeriodChange} />
      </div>
      <CostDashboardPageShell pageData={pageData} onSelectDept={setSelectedDeptId} />
    </DashboardPageLayout>
  )
}
