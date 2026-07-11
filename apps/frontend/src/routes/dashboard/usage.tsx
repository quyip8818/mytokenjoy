import { UsageDashboardLayoutPageShell, useUsageDashboardRoutePage } from '@/features/dashboard'

export default function UsageDashboardPage() {
  return <UsageDashboardLayoutPageShell {...useUsageDashboardRoutePage()} />
}
