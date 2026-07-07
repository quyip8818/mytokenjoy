import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { ErrorState } from '@/components/ui/error-state'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Progress } from '@/components/ui/progress'
import { Badge } from '@/components/ui/badge'
import { UsageModelChart, useUsageDashboardPage } from '@/features/dashboard'

export default function UsageDashboardPage() {
  const { teamUsage, modelUsage, loading, error, refresh } = useUsageDashboardPage()

  if (error) {
    return (
      <PageShell>
        <ErrorState message={error.message} onRetry={() => void refresh()} />
      </PageShell>
    )
  }

  return (
    <PageShell>
      <DataSection title="团队用量" loading={loading} skeletonColumns={6}>
        <Table>
          <TableHeader>
            <TableRow className="border-border/50 hover:bg-transparent">
              <TableHead className="text-xs font-semibold text-muted-foreground">部门</TableHead>
              <TableHead className="text-xs font-semibold text-muted-foreground">
                额度 (¥)
              </TableHead>
              <TableHead className="text-xs font-semibold text-muted-foreground">
                已消耗 (¥)
              </TableHead>
              <TableHead className="text-xs font-semibold text-muted-foreground w-48">
                消耗进度
              </TableHead>
              <TableHead className="text-xs font-semibold text-muted-foreground text-right">
                成员数
              </TableHead>
              <TableHead className="text-xs font-semibold text-muted-foreground">
                主力模型
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {teamUsage.map((t) => {
              const pct = Math.round((t.consumed / t.quota) * 100)
              return (
                <TableRow key={t.departmentId} className="border-border-subtle hover:bg-muted/50">
                  <TableCell className="font-medium">{t.departmentName}</TableCell>
                  <TableCell className="text-muted-foreground">
                    {t.quota.toLocaleString()}
                  </TableCell>
                  <TableCell className="font-medium">{t.consumed.toLocaleString()}</TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2.5">
                      <Progress value={pct} className="flex-1 h-2" />
                      <span
                        className={`text-xs font-semibold ${pct >= 90 ? 'text-red-500' : pct >= 70 ? 'text-amber-500' : 'text-primary'}`}
                      >
                        {pct}%
                      </span>
                    </div>
                  </TableCell>
                  <TableCell className="text-right text-muted-foreground">
                    {t.memberCount}
                  </TableCell>
                  <TableCell>
                    <Badge variant="secondary" className="text-xs font-medium">
                      {t.topModel}
                    </Badge>
                  </TableCell>
                </TableRow>
              )
            })}
          </TableBody>
        </Table>
      </DataSection>

      <DataSection title="模型使用分布" loading={loading} skeletonColumns={1}>
        <UsageModelChart modelUsage={modelUsage} />
      </DataSection>
    </PageShell>
  )
}
