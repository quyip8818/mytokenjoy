import { useMemo } from 'react'
import { Activity } from 'lucide-react'
import { AuditFilteredPage } from '@/components/audit/audit-filtered-page'
import { AuditListToolbar } from '@/components/audit/audit-list-toolbar'
import { AuditTablePagination } from '@/components/audit/audit-table-pagination'
import { OptionsSelect } from '@/components/ui/options-select'
import { useAuditModelOptions } from '@/hooks/use-audit-model-options'
import { CALL_LOG_STATUS_LABELS } from '@/lib/labels'
import { CallLogsTable } from '@/routes/audit/components/call-logs-table'
import { useAuditCallsPage } from '@/routes/audit/hooks/use-audit-calls-page'

export default function CallLogsPage() {
  const {
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
    handleExport,
    toggleExpanded,
  } = useAuditCallsPage()
  const { models } = useAuditModelOptions()
  const modelOptions = useMemo(
    () => Object.fromEntries(models.map((model) => [model.name, model.displayName])),
    [models],
  )

  return (
    <AuditFilteredPage
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
          keyword={keyword}
          onKeywordChange={setKeyword}
          onExport={handleExport}
        >
          <OptionsSelect
            value={statusFilter}
            onValueChange={setStatusFilter}
            options={CALL_LOG_STATUS_LABELS}
            allLabel="全部状态"
            className="w-32"
          />
          <OptionsSelect
            value={modelFilter}
            onValueChange={setModelFilter}
            options={modelOptions}
            allLabel="全部模型"
            className="w-40"
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
    </AuditFilteredPage>
  )
}
