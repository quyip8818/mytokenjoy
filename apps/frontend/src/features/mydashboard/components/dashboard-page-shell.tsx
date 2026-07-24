import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import type { useMyDashboardPage } from '@/features/mydashboard'
import { MyConsumptionCharts } from './consumption-charts'
import { MyDashboardStats } from './dashboard-stats'

type MyDashboardPageShellProps = ReturnType<typeof useMyDashboardPage>

export function MyDashboardPageShell({
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
}: MyDashboardPageShellProps) {
  return (
    <PageShell>
      <DataSection
        loading={loading}
        error={error}
        onRetry={() => void refresh()}
        contentClassName="space-y-6"
      >
        <MyDashboardStats
          loading={loading}
          accountData={accountData}
          usageStats={usageStats}
          resourceConsumption={resourceConsumption}
          performance={performance}
        />
        <MyConsumptionCharts
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
