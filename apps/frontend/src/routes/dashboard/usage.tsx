import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { ErrorState } from '@/components/ui/error-state'
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
import { useUsageDashboardPage } from '@/routes/dashboard/hooks/use-usage-dashboard-page'

export default function UsageDashboardPage() {
  const { teamUsage, modelUsage, loading, error, refresh } = useUsageDashboardPage()

  if (error) {
    return (
      <PageShell>
        <ErrorState message={error.message} onRetry={refresh} />
      </PageShell>
    )
  }

  return (
    <PageShell>
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
