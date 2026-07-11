import { CostDashboardPageShell, useCostDashboardPage } from '@/features/dashboard'

export default function CostDashboardPage() {
  const pageData = useCostDashboardPage({ deptId: null })
  return <CostDashboardPageShell pageData={pageData} />
}
