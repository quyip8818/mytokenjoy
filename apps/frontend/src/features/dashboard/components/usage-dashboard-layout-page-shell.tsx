import { DashboardDateRangePicker } from './dashboard-date-range-picker'
import { DashboardPageLayout } from './dashboard-page-layout'
import { OrgTreeSidebar } from './org-tree-sidebar'
import { UsageDashboardPageShell } from './usage-dashboard-page-shell'
import type { useUsageDashboardRoutePage } from '../hooks/use-usage-dashboard-route-page'

type UsageDashboardLayoutPageShellProps = ReturnType<typeof useUsageDashboardRoutePage>

export function UsageDashboardLayoutPageShell({
  selectedDeptId,
  setSelectedDeptId,
  departments,
  treeLoading,
  getBreadcrumb,
  pageData,
}: UsageDashboardLayoutPageShellProps) {
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
          <p className="text-xs text-muted-foreground">{getBreadcrumb(selectedDeptId).join(' > ')}</p>
          <h1 className="text-lg font-semibold">用量分析</h1>
        </div>
        <DashboardDateRangePicker
          value={pageData.period}
          onChange={pageData.handlePeriodChange}
        />
      </div>
      <UsageDashboardPageShell pageData={pageData} onSelectDept={setSelectedDeptId} />
    </DashboardPageLayout>
  )
}
