import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { PageShell } from '@/components/layout/page-shell'
import { ErrorState } from '@/components/ui/error-state'
import { COST_PERIOD_LABELS } from '@/lib/dashboard-constants'
import { useCostDashboard } from '@/routes/dashboard/hooks/use-cost-dashboard'
import { CostSummaryStats } from '@/routes/dashboard/components/cost-summary-stats'
import { CostTrendChart } from '@/routes/dashboard/components/cost-trend-chart'
import { CostDistributionChart } from '@/routes/dashboard/components/cost-distribution-chart'
import { CostDrillTable } from '@/routes/dashboard/components/cost-drill-table'
import { CostTopConsumersTable } from '@/routes/dashboard/components/cost-top-consumers-table'

export default function CostDashboardPage() {
  const {
    period,
    drill,
    loading,
    error,
    refresh,
    dailyCosts,
    topConsumers,
    deptCosts,
    memberCosts,
    deptCostsWithColors,
    drillTitle,
    stats,
    canDrillBack,
    handlePeriodChange,
    handleDrillDept,
    handleDrillBack,
  } = useCostDashboard()

  if (error) {
    return (
      <PageShell>
        <ErrorState message={error.message} onRetry={refresh} />
      </PageShell>
    )
  }

  return (
    <PageShell
      actions={
        <Select value={period} onValueChange={handlePeriodChange}>
          <SelectTrigger className="w-32 border-border/60">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {Object.entries(COST_PERIOD_LABELS).map(([value, label]) => (
              <SelectItem key={value} value={value}>
                {label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      }
      stats={<CostSummaryStats stats={stats} loading={loading} />}
    >
      <div className="grid grid-cols-3 gap-6">
        <CostTrendChart dailyCosts={dailyCosts} loading={loading} />
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
