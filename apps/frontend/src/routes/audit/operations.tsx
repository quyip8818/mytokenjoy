import { ScrollText } from 'lucide-react'
import { AuditFilteredPage } from '@/components/audit/audit-filtered-page'
import { AuditListToolbar } from '@/components/audit/audit-list-toolbar'
import { OptionsSelect } from '@/components/ui/options-select'
import { OPERATION_ACTION_LABELS } from '@/lib/labels'
import { OperationsLogTable } from '@/routes/audit/components/operations-log-table'
import { useAuditOperationsPage } from '@/routes/audit/hooks/use-audit-operations-page'

export default function OperationLogsPage() {
  const {
    logs,
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
    handleExport,
  } = useAuditOperationsPage()

  return (
    <AuditFilteredPage
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
          keyword={keyword}
          onKeywordChange={setKeyword}
          onExport={handleExport}
        >
          <OptionsSelect
            value={actionFilter}
            onValueChange={setActionFilter}
            options={OPERATION_ACTION_LABELS}
            allLabel="全部类型"
            className="w-40"
          />
        </AuditListToolbar>
      }
    >
      <OperationsLogTable logs={logs} />
    </AuditFilteredPage>
  )
}
