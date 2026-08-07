import { PageShell } from '@/components/layout/page-shell'
import { SplitPanel } from '@/components/layout/split-panel'
import { ContextHeader } from '@/components/layout/context-header'
import { DashboardDateRangePicker } from './dashboard-date-range-picker'
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
    <PageShell testId="page-dashboard-usage" className="flex min-h-0 flex-1 flex-col">
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
              <UsageDashboardPageShell pageData={pageData} onSelectDept={setSelectedDeptId} />
            </div>
          </>
        }
      />
    </PageShell>
  )
}
