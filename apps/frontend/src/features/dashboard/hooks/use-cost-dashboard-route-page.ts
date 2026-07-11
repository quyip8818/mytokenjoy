import { useCostDashboardPage } from './use-cost-dashboard-page'
import { useDeptSelection } from './use-dept-selection'
import { useOrgTree } from './use-org-tree'

export function useCostDashboardRoutePage() {
  const { selectedDeptId, setSelectedDeptId } = useDeptSelection()
  const { departments, loading: treeLoading, getBreadcrumb } = useOrgTree()
  const pageData = useCostDashboardPage({ deptId: selectedDeptId })

  return {
    selectedDeptId,
    setSelectedDeptId,
    departments,
    treeLoading,
    getBreadcrumb,
    pageData,
  }
}
