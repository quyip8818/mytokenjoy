import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import type { useUsageDashboardPage } from '@/features/dashboard/hooks/use-usage-dashboard-page'
import { TeamUsageTable } from './team-usage-table'
import { UsageModelChart } from './usage-model-chart'

type UsageDashboardPageShellProps = ReturnType<typeof useUsageDashboardPage>

export function UsageDashboardPageShell({
  teamUsage,
  modelUsage,
  loading,
  error,
  refresh,
}: UsageDashboardPageShellProps) {
  return (
    <PageShell>
      <DataSection
        title="团队用量"
        loading={loading}
        error={error}
        onRetry={() => void refresh()}
        skeletonColumns={6}
      >
        <TeamUsageTable teamUsage={teamUsage} />
      </DataSection>

      <DataSection title="模型使用分布" loading={loading} skeletonColumns={1}>
        <UsageModelChart modelUsage={modelUsage} />
      </DataSection>
    </PageShell>
  )
}
