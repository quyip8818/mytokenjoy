import { PageShell } from '@/components/layout/page-shell'
import { ErrorState } from '@/components/ui/error-state'
import {
  CostSummaryStats,
  CostTrendChart,
  CostDistributionChart,
  CostDrillTable,
  CostTopConsumersTable,
  useCostDashboardPage,
} from '@/features/dashboard'

export default function CostDashboardPage() {
  const {
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
  } = useCostDashboardPage()

  if (error) {
    return (
      <PageShell>
        <ErrorState message={error.message} onRetry={() => void refresh()} />
      </PageShell>
    )
  }

  return (
    <PageShell stats={<CostSummaryStats stats={stats} loading={loading} />}>
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
    </PageShell>
  )
}
