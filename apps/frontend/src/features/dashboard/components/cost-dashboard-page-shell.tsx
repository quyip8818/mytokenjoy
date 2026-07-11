import { ErrorState } from '@/components/ui/error-state'
import type { useCostDashboardPage } from '../hooks/use-cost-dashboard-page'
import { CostSummaryStats } from './cost-summary-stats'
import { CostTrendChart } from './cost-trend-chart'
import { CostDistributionChart } from './cost-distribution-chart'
import { DeptComparisonTable } from './dept-comparison-table'
import { CostTopConsumersTable } from './cost-top-consumers-table'

interface CostDashboardPageShellProps {
  pageData: ReturnType<typeof useCostDashboardPage>
  onSelectDept?: (deptId: string) => void
}

export function CostDashboardPageShell({ pageData, onSelectDept }: CostDashboardPageShellProps) {
  const {
    loading,
    error,
    refresh,
    stats,
    dailyCosts,
    topConsumers,
    deptCosts,
    deptCostsWithColors,
    granularity,
  } = pageData

  if (error) {
    return <ErrorState message={error.message} onRetry={() => void refresh()} />
  }

  return (
    <div className="space-y-6">
      <CostSummaryStats stats={stats} loading={loading} />
      <div className="grid grid-cols-[5fr_3fr] gap-6">
        <CostTrendChart dailyCosts={dailyCosts} loading={loading} granularity={granularity} />
        <CostDistributionChart data={deptCostsWithColors} loading={loading} />
      </div>
      <DeptComparisonTable
        deptCosts={deptCosts}
        loading={loading}
        onSelectDept={onSelectDept}
      />
      <CostTopConsumersTable topConsumers={topConsumers} loading={loading} />
    </div>
  )
}
