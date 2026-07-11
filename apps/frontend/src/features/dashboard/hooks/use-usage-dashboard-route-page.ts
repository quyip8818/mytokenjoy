import { useUsageDashboardPage } from './use-usage-dashboard-page'
import { useDeptSelection } from './use-dept-selection'
import { useOrgTree } from './use-org-tree'

export function useUsageDashboardRoutePage() {
  const { selectedDeptId, setSelectedDeptId } = useDeptSelection()
  const { departments, loading: treeLoading, getBreadcrumb } = useOrgTree()
  const pageData = useUsageDashboardPage({ deptId: selectedDeptId })

  return {
    selectedDeptId,
    setSelectedDeptId,
    departments,
    treeLoading,
    getBreadcrumb,
    pageData,
  }
}
