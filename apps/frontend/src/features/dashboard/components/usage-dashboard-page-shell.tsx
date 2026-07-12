import { ErrorState } from '@/components/ui/error-state'
import { DataSection } from '@/components/layout/data-section'
import type { useUsageDashboardPage } from '../hooks/use-usage-dashboard-page'
import { DepartmentUsageTable } from './department-usage-table'
import { UsageModelChart } from './usage-model-chart'
import { UsageMemberRankingTable } from './usage-member-ranking-table'

interface UsageDashboardPageShellProps {
  pageData: ReturnType<typeof useUsageDashboardPage>
  onSelectDept?: (deptId: string) => void
}

export function UsageDashboardPageShell({ pageData, onSelectDept }: UsageDashboardPageShellProps) {
  const { departmentUsage, modelUsage, topConsumers, loading, error, refresh } = pageData

  if (error) {
    return <ErrorState message={error.message} onRetry={() => void refresh()} />
  }

  return (
    <div className="space-y-6">
      <DataSection
        title="团队用量与配额"
        loading={loading}
        skeletonColumns={6}
        className="border-border shadow-xs"
      >
        <DepartmentUsageTable departmentUsage={departmentUsage} onSelectDept={onSelectDept} />
      </DataSection>

      <div className="grid grid-cols-[5fr_3fr] gap-6">
        <DataSection
          title="模型费用分布"
          loading={loading}
          skeletonColumns={1}
          className="border-border shadow-xs"
        >
          <UsageModelChart modelUsage={modelUsage} />
        </DataSection>

        <UsageMemberRankingTable topConsumers={topConsumers} loading={loading} />
      </div>
    </div>
  )
}
