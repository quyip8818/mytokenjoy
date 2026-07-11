import { CostDashboardLayoutPageShell, useCostDashboardRoutePage } from '@/features/dashboard'

export default function CostDashboardPage() {
  return <CostDashboardLayoutPageShell {...useCostDashboardRoutePage()} />
}
