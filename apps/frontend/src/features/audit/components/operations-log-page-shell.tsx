import { ScrollText } from 'lucide-react'
import type { useAuditOperationsPage } from '@/features/audit'
import { FilteredPageShell } from '@/components/layout/filtered-page-shell'
import { AuditListToolbar } from './audit-list-toolbar'
import { AuditTablePagination } from './audit-table-pagination'
import { OperationsLogFilters } from './operations-log-filters'
import { OperationsLogTable } from './operations-log-table'

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
