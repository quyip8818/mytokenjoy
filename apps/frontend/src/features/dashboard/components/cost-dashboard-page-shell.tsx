import { ErrorState } from '@/components/ui/error-state'
import type { useCostDashboardPage } from '@/features/dashboard'
import { CostSummaryStats } from './cost-summary-stats'
import { CostTrendChart } from './cost-trend-chart'
import { CostDistributionChart } from '@/features/dashboard'
import { DeptComparisonTable } from '@/features/dashboard'
import { CostTopConsumersTable } from '@/features/dashboard'
import { BudgetHeroCard } from './budget-hero-card'

interface CostDashboardPageShellProps {
  pageData: ReturnType<typeof useCostDashboardPage>
  onSelectDept?: (deptId: string) => void
}

export function CostDashboardPageShell({ pageData, onSelectDept }: CostDashboardPageShellProps) {
  const {
    loading,
    budgetLoading,
    budgetSummary,
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
      <BudgetHeroCard
        budget={budgetSummary.budget}
        consumed={budgetSummary.consumed}
        loading={budgetLoading}
      />
      <CostSummaryStats stats={stats} loading={loading} />
      <div className="grid grid-cols-[5fr_3fr] gap-6">
        <CostTrendChart dailyCosts={dailyCosts} loading={loading} granularity={granularity} />
        <CostDistributionChart data={deptCostsWithColors} loading={loading} />
      </div>
      <DeptComparisonTable deptCosts={deptCosts} loading={loading} onSelectDept={onSelectDept} />
      <CostTopConsumersTable topConsumers={topConsumers} loading={loading} />
    </div>
  )
}
