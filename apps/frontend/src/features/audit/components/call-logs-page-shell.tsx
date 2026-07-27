import { Activity } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { PageHeader } from '@/components/layout/page-header'
import { listEmpty } from '@/lib/list-empty'
import type { useAuditCallsPage } from '@/features/audit'
import { AuditListToolbar } from './audit-list-toolbar'
import { AuditTablePagination } from './audit-table-pagination'
import { CallLogsFilters } from './call-logs-filters'
import { CallLogsTable } from './call-logs-table'
import { CallLogsSummaryCards } from './call-logs-summary-cards'

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
  summary,
  summaryLoading,
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
    <PageShell>
      <PageHeader title="调用日志" />

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

      <CallLogsSummaryCards summary={summary} loading={summaryLoading} />

      <Card className="border-border shadow-xs">
        <CardContent className="px-5 pt-5 pb-4">
          <DataSection
            loading={loading}
            error={error}
            onRetry={refresh}
            skeletonColumns={8}
            empty={listEmpty(loading, logs, {
              icon: Activity,
              title: '暂无调用记录',
              description: '模型 API 调用成功后，日志将显示在这里',
              variant: 'inline',
            })}
          >
            <CallLogsTable
              logs={logs}
              expandedId={expandedId}
              contentRetentionEnabled={contentRetentionEnabled}
              onToggleExpanded={toggleExpanded}
            />
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
