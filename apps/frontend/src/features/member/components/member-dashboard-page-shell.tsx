import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import type { useMemberDashboardPage } from '@/features/member'
import { MemberConsumptionCharts } from './member-consumption-charts'
import { MemberDashboardStats } from './member-dashboard-stats'

type MemberDashboardPageShellProps = ReturnType<typeof useMemberDashboardPage>

export function MemberDashboardPageShell({
  loading,
  error,
  refresh,
  accountData,
  usageStats,
  resourceConsumption,
  performance,
  consumptionTrend,
  consumptionDistribution,
  callDistribution,
  callRanking,
  distributionTotal,
  trendTotal,
  callTotal,
}: MemberDashboardPageShellProps) {
  return (
    <PageShell>
      <DataSection
        loading={loading}
        error={error}
        onRetry={() => void refresh()}
        contentClassName="space-y-6"
      >
        <MemberDashboardStats
          loading={loading}
          accountData={accountData}
          usageStats={usageStats}
          resourceConsumption={resourceConsumption}
          performance={performance}
        />
        <MemberConsumptionCharts
          loading={loading}
          consumptionDistribution={consumptionDistribution}
          consumptionTrend={consumptionTrend}
          callDistribution={callDistribution}
          callRanking={callRanking}
          distributionTotal={distributionTotal}
          trendTotal={trendTotal}
          callTotal={callTotal}
        />
      </DataSection>
    </PageShell>
  )
}
