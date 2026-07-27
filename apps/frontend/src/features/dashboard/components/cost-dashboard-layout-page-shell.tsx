import { PageShell } from '@/components/layout/page-shell'
import { SplitPanel } from '@/components/layout/split-panel'
import { ContextHeader } from '@/components/layout/context-header'
import { CostDashboardPageShell } from './cost-dashboard-page-shell'
import { DashboardDateRangePicker } from './dashboard-date-range-picker'
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
    <PageShell className="flex min-h-0 flex-1 flex-col">
      <SplitPanel
        master={
          <OrgTreeSidebar
            departments={departments}
            selectedDeptId={selectedDeptId}
            onSelect={setSelectedDeptId}
            loading={treeLoading}
          />
        }
        detail={
          <>
            <ContextHeader
              breadcrumb={getBreadcrumb(selectedDeptId)}
              actions={
                <DashboardDateRangePicker
                  value={pageData.period}
                  onChange={pageData.handlePeriodChange}
                />
              }
            />
            <div className="min-h-0 flex-1 overflow-y-auto p-6">
              <CostDashboardPageShell pageData={pageData} onSelectDept={setSelectedDeptId} />
            </div>
          </>
        }
      />
    </PageShell>
  )
}
