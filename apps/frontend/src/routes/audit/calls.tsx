import { Activity } from 'lucide-react'
import { AuditFilteredPage } from '@/components/audit/audit-filtered-page'
import { AuditMemberSelect } from '@/components/audit/audit-member-select'
import { AuditToolbar } from '@/components/audit/audit-toolbar'
import { AuditDatePresetSelect } from '@/components/audit/audit-date-preset-select'
import { AuditKeywordInput } from '@/components/audit/audit-keyword-input'
import { OptionsSelect } from '@/components/ui/options-select'
import { CALL_LOG_STATUS_LABELS } from '@/lib/labels'
import { CallLogsTable } from '@/routes/audit/components/call-logs-table'
import { useAuditCallsPage } from '@/routes/audit/hooks/use-audit-calls-page'

export default function CallLogsPage() {
  const {
    logs,
    loading,
    error,
    refresh,
    statusFilter,
    callerId,
    datePreset,
    keyword,
    setStatusFilter,
    setCallerId,
    setDatePreset,
    setKeyword,
    expandedId,
    contentRetentionEnabled,
    handleExport,
    toggleExpanded,
  } = useAuditCallsPage()

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
        <div className="flex flex-wrap items-center gap-3">
          <AuditDatePresetSelect value={datePreset} onValueChange={setDatePreset} />
          <OptionsSelect
            value={statusFilter}
            onValueChange={setStatusFilter}
            options={CALL_LOG_STATUS_LABELS}
            allLabel="全部状态"
            className="w-32"
          />
          <AuditMemberSelect
            value={callerId}
            onValueChange={setCallerId}
            allLabel="全部调用人"
          />
          <AuditKeywordInput value={keyword} onChange={setKeyword} />
          <AuditToolbar onExport={handleExport} />
        </div>
      }
    >
      <CallLogsTable
        logs={logs}
        expandedId={expandedId}
        contentRetentionEnabled={contentRetentionEnabled}
        onToggleExpanded={toggleExpanded}
      />
    </AuditFilteredPage>
  )
}
