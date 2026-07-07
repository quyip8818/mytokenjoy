import { UsageDashboardPageShell, useUsageDashboardPage } from '@/features/dashboard'

export default function UsageDashboardPage() {
  return <UsageDashboardPageShell {...useUsageDashboardPage()} />
}
