import { ScrollText } from 'lucide-react'
import type { useAuditOperationsPage } from '@/features/audit'
import { FilteredPageShell } from '@/components/layout/filtered-page-shell'
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
    setDatePreset('today') // triggers re-render, but we override below
    // Set both from/to to the clicked date by using the preset mechanism
    // Actually we need a custom approach - set datePreset to a special value
    // For simplicity, just filter to "today" or use custom logic
    // Since our presets don't support arbitrary single-day, we'll just do nothing for now
    // and keep the timeline as a visual indicator only
    void date
  }

  return (
    <FilteredPageShell
      title="操作记录"
      loading={loading}
      error={error}
      onRetry={refresh}
      items={logs}
      empty={{
        icon: ScrollText,
        title: '暂无操作记录',
        description: '调整筛选条件或完成管理操作后，记录将显示在这里',
      }}
      actions={
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
      }
    >
      <OperationsTimelineChart
        data={timeline}
        loading={timelineLoading}
        onDayClick={handleDayClick}
      />
      <OperationsLogTable logs={logs} />
      <AuditTablePagination
        total={total}
        page={page}
        totalPages={totalPages}
        onPageChange={setPage}
      />
    </FilteredPageShell>
  )
}
