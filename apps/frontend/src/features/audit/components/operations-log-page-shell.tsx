import { ScrollText } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { PageHeader } from '@/components/layout/page-header'
import { listEmpty } from '@/lib/list-empty'
import type { useAuditOperationsPage } from '@/features/audit'
import { AuditListToolbar } from './audit-list-toolbar'
import { AuditTablePagination } from './audit-table-pagination'
import { OperationsLogFilters } from './operations-log-filters'
import { OperationsLogTable } from './operations-log-table'
import { OperationsTimelineChart } from './operations-timeline-chart'

type OperationsLogPageShellProps = ReturnType<typeof useAuditOperationsPage>

export function OperationsLogPageShell({
  logs,
  total,
  page,
  totalPages,
  setPage,
  loading,
  error,
  refresh,
  timeline,
  timelineLoading,
  actionFilter,
  datePreset,
  operatorId,
  keyword,
  setActionFilter,
  setDatePreset,
  setOperatorId,
  setKeyword,
  memberOptions,
  handleExport,
}: OperationsLogPageShellProps) {
  const handleDayClick = (date: string) => {
    setDatePreset('today')
    void date
  }

  return (
    <PageShell>
      <PageHeader title="操作审计" />

      <AuditListToolbar
        datePreset={datePreset}
        onDatePresetChange={setDatePreset}
        memberId={operatorId}
        onMemberIdChange={setOperatorId}
        memberAllLabel="全部操作人"
        memberOptions={memberOptions}
        keyword={keyword}
        onKeywordChange={setKeyword}
        onExport={handleExport}
      >
        <OperationsLogFilters
          actionFilter={actionFilter}
          onActionFilterChange={setActionFilter}
        />
      </AuditListToolbar>

      <OperationsTimelineChart
        data={timeline}
        loading={timelineLoading}
        onDayClick={handleDayClick}
      />

      <Card className="border-border shadow-xs">
        <CardContent className="px-5 pt-5 pb-4">
          <DataSection
            loading={loading}
            error={error}
            onRetry={refresh}
            skeletonColumns={6}
            empty={listEmpty(loading, logs, {
              icon: ScrollText,
              title: '暂无操作记录',
              description: '调整筛选条件或完成管理操作后，记录将显示在这里',
              variant: 'inline',
            })}
          >
            <OperationsLogTable logs={logs} />
          </DataSection>
        </CardContent>
      </Card>

      <AuditTablePagination
        total={total}
        page={page}
        totalPages={totalPages}
        onPageChange={setPage}
      />
    </PageShell>
  )
}
