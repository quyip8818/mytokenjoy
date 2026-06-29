import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { ErrorState } from '@/components/ui/error-state'
import { EmptyState } from '@/components/ui/empty-state'
import { StatusBadge } from '@/components/ui/status-badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { BudgetProgressCell } from '@/components/budget/budget-progress-cell'
import { UsageModelChart } from '@/routes/dashboard/components/usage-model-chart'
import { UsageSeriesChart } from '@/routes/dashboard/components/usage-series-chart'
import { useUsageDashboardPage } from '@/routes/dashboard/hooks/use-usage-dashboard-page'
import { useUsageSeriesPage } from '@/routes/dashboard/hooks/use-usage-series-page'
import { USAGE_GRANULARITY_LABELS } from '@/lib/dashboard-constants'
import type { UsageGranularity } from '@/api/types'

export default function UsageDashboardPage() {
  const { teamUsage, modelUsage, loading, error, refresh } = useUsageDashboardPage()
  const {
    granularity,
    series,
    chartData,
    loading: seriesLoading,
    error: seriesError,
    isServiceUnavailable,
    serviceUnavailableMessage,
    refresh: refreshSeries,
    handleGranularityChange,
  } = useUsageSeriesPage()

  if (error) {
    return (
      <PageShell>
        <ErrorState message={error.message} onRetry={refresh} />
      </PageShell>
    )
  }

  const seriesErrorMessage = isServiceUnavailable
    ? (serviceUnavailableMessage ?? seriesError?.message)
    : seriesError?.message

  return (
    <PageShell>
      <DataSection title="实时监控" loading={seriesLoading} skeletonColumns={1}>
        <div className="mb-4 flex flex-wrap items-center gap-3">
          <Select value={granularity} onValueChange={handleGranularityChange}>
            <SelectTrigger className="w-32">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {(Object.keys(USAGE_GRANULARITY_LABELS) as UsageGranularity[]).map((value) => (
                <SelectItem key={value} value={value}>
                  {USAGE_GRANULARITY_LABELS[value]}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          {series && (
            <span className="text-xs text-muted-foreground">
              数据源: {series.source} · 时区: {series.timezone}
            </span>
          )}
        </div>
        {series?.approximate && (
          <div className="mb-4 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-900">
            实时近似数据，与日报表可能存在偏差
          </div>
        )}
        {seriesErrorMessage ? (
          <ErrorState message={seriesErrorMessage} onRetry={refreshSeries} />
        ) : chartData.length === 0 && !seriesLoading ? (
          <EmptyState title="暂无用量数据" description="请调整时间粒度或稍后再试" />
        ) : (
          <UsageSeriesChart data={chartData} />
        )}
      </DataSection>

      <DataSection title="团队用量" loading={loading} skeletonColumns={6}>
        <Table>
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              <TableHead>部门</TableHead>
              <TableHead>额度 (¥)</TableHead>
              <TableHead>已消耗 (¥)</TableHead>
              <TableHead className="w-48">消耗进度</TableHead>
              <TableHead className="text-right">成员数</TableHead>
              <TableHead>主力模型</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {teamUsage.map((t) => (
              <TableRow key={t.departmentId}>
                <TableCell className="font-medium">{t.departmentName}</TableCell>
                <TableCell className="text-muted-foreground">{t.quota.toLocaleString()}</TableCell>
                <TableCell className="font-medium">{t.consumed.toLocaleString()}</TableCell>
                <TableCell>
                  <BudgetProgressCell
                    value={t.consumed}
                    total={t.quota}
                    className="gap-2.5"
                    accentLabel
                  />
                </TableCell>
                <TableCell className="text-right text-muted-foreground">{t.memberCount}</TableCell>
                <TableCell>
                  <StatusBadge variant="info">{t.topModel}</StatusBadge>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </DataSection>

      <DataSection title="模型使用分布" loading={loading} skeletonColumns={1}>
        <UsageModelChart modelUsage={modelUsage} />
      </DataSection>
    </PageShell>
  )
}
