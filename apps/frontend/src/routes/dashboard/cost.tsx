import { CostDashboardPageShell, useCostDashboardPage } from '@/features/dashboard'

export default function CostDashboardPage() {
  return <CostDashboardPageShell {...useCostDashboardPage()} />
}
