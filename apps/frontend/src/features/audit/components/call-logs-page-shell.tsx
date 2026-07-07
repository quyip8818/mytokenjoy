import { Activity } from 'lucide-react'
import type { useAuditCallsPage } from '@/features/audit'
import { FilteredPageShell } from '@/components/layout/filtered-page-shell'
import { AuditListToolbar } from './audit-list-toolbar'
import { AuditTablePagination } from './audit-table-pagination'
import { CallLogsFilters } from './call-logs-filters'
import { CallLogsTable } from './call-logs-table'

type CallLogsPageShellProps = ReturnType<typeof useAuditCallsPage>

export function CallLogsPageShell({
  logs,
  total,
  page,
  totalPages,
  setPage,
  loading,
  error,
  refresh,
  statusFilter,
  callerId,
  modelFilter,
  datePreset,
  keyword,
  setStatusFilter,
  setCallerId,
  setModelFilter,
  setDatePreset,
  setKeyword,
  expandedId,
  contentRetentionEnabled,
  modelOptions,
  memberOptions,
  handleExport,
  toggleExpanded,
}: CallLogsPageShellProps) {
  return (
    <FilteredPageShell
      title="调用记录"
      loading={loading}
      error={error}
      onRetry={refresh}
      items={logs}
      empty={{
        icon: Activity,
        title: '暂无调用记录',
        description: '模型 API 调用成功后，日志将显示在这里',
      }}
      actions={
        <AuditListToolbar
          datePreset={datePreset}
          onDatePresetChange={setDatePreset}
          memberId={callerId}
          onMemberIdChange={setCallerId}
          memberAllLabel="全部调用人"
          memberOptions={memberOptions}
          keyword={keyword}
          onKeywordChange={setKeyword}
          onExport={handleExport}
        >
          <CallLogsFilters
            statusFilter={statusFilter}
            modelFilter={modelFilter}
            modelOptions={modelOptions}
            onStatusChange={setStatusFilter}
            onModelChange={setModelFilter}
          />
        </AuditListToolbar>
      }
    >
      <CallLogsTable
        logs={logs}
        expandedId={expandedId}
        contentRetentionEnabled={contentRetentionEnabled}
        onToggleExpanded={toggleExpanded}
      />
      <AuditTablePagination
        total={total}
        page={page}
        totalPages={totalPages}
        onPageChange={setPage}
      />
    </FilteredPageShell>
  )
}
