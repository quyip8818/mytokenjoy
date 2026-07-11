import { UsageDashboardPageShell, useUsageDashboardPage } from '@/features/dashboard'

export default function UsageDashboardPage() {
  const pageData = useUsageDashboardPage({ deptId: null })
  return <UsageDashboardPageShell pageData={pageData} />
}
