import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import type { useCostDashboardPage } from '@/features/dashboard/hooks/use-cost-dashboard-page'
import { CostSummaryStats } from './cost-summary-stats'
import { CostTrendChart } from './cost-trend-chart'
import { CostDistributionChart } from './cost-distribution-chart'
import { CostDrillTable } from './cost-drill-table'
import { CostTopConsumersTable } from './cost-top-consumers-table'

type CostDashboardPageShellProps = ReturnType<typeof useCostDashboardPage>

export function CostDashboardPageShell({
  loading,
  error,
  refresh,
  stats,
  dailyCosts,
  topConsumers,
  deptCosts,
  memberCosts,
  deptCostsWithColors,
  drill,
  drillTitle,
  granularity,
  canDrillBack,
  handleDrillDept,
  handleDrillBack,
}: CostDashboardPageShellProps) {
  return (
    <PageShell stats={<CostSummaryStats stats={stats} loading={loading} />}>
      <DataSection
        loading={loading}
        error={error}
        onRetry={() => void refresh()}
        contentClassName="space-y-6"
      >
        <div className="grid grid-cols-3 gap-6">
          <CostTrendChart dailyCosts={dailyCosts} loading={loading} granularity={granularity} />
          <CostDistributionChart data={deptCostsWithColors} loading={loading} />
        </div>
        <CostDrillTable
          drill={drill}
          drillTitle={drillTitle}
          deptCosts={deptCosts}
          memberCosts={memberCosts}
          loading={loading}
          canDrillBack={canDrillBack}
          onDrillBack={handleDrillBack}
          onDrillDept={handleDrillDept}
        />
        <CostTopConsumersTable topConsumers={topConsumers} loading={loading} />
      </DataSection>
    </PageShell>
  )
}
