import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Input } from '@/components/ui/input'
import { PageShell } from '@/components/layout/page-shell'
import { ErrorState } from '@/components/ui/error-state'
import { COST_GRANULARITY_LABELS, COST_PERIOD, COST_PERIOD_LABELS } from '@/lib/dashboard-constants'
import type { CostGranularity } from '@/api/types'
import { useCostDashboardPage } from '@/routes/dashboard/hooks/use-cost-dashboard-page'
import { CostSummaryStats } from '@/routes/dashboard/components/cost-summary-stats'
import { CostTrendChart } from '@/routes/dashboard/components/cost-trend-chart'
import { CostDistributionChart } from '@/routes/dashboard/components/cost-distribution-chart'
import { CostDrillTable } from '@/routes/dashboard/components/cost-drill-table'
import { CostTopConsumersTable } from '@/routes/dashboard/components/cost-top-consumers-table'

export default function CostDashboardPage() {
  const {
    period,
    startDate,
    endDate,
    granularity,
    customDateInvalid,
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
    setStartDate,
    setEndDate,
    setGranularity,
    handleDrillDept,
    handleDrillBack,
  } = useCostDashboardPage()

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
        <div className="flex flex-wrap items-center gap-3">
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
          {period === COST_PERIOD.CUSTOM ? (
            <>
              <Input
                type="date"
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
                className="w-36 border-border/60"
              />
              <Input
                type="date"
                value={endDate}
                onChange={(e) => setEndDate(e.target.value)}
                className="w-36 border-border/60"
              />
            </>
          ) : null}
          <Select
            value={granularity}
            onValueChange={(v) => setGranularity((v ?? 'day') as CostGranularity)}
          >
            <SelectTrigger className="w-28 border-border/60">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {Object.entries(COST_GRANULARITY_LABELS).map(([value, label]) => (
                <SelectItem key={value} value={value}>
                  {label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      }
      stats={<CostSummaryStats stats={stats} loading={loading} />}
    >
      {customDateInvalid ? (
        <p className="mb-4 text-sm text-amber-700">自定义日期范围无效：开始日期不能晚于结束日期</p>
      ) : null}
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
