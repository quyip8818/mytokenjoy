import { CostDashboardPageShell } from './cost-dashboard-page-shell'
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
      <div className="mb-4">
        <p className="text-xs text-muted-foreground">{getBreadcrumb(selectedDeptId).join(' > ')}</p>
        <h1 className="text-lg font-semibold">成本看板</h1>
      </div>
      <CostDashboardPageShell pageData={pageData} onSelectDept={setSelectedDeptId} />
    </DashboardPageLayout>
  )
}
