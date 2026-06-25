import { Activity } from 'lucide-react'
import { AuditFilteredPage } from '@/components/audit/audit-filtered-page'
import { AuditToolbar } from '@/components/audit/audit-toolbar'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { CallLogsTable } from '@/routes/audit/components/call-logs-table'
import { useAuditCallsPage } from '@/routes/audit/hooks/use-audit-calls-page'

export default function CallLogsPage() {
  const {
    logs,
    loading,
    error,
    refresh,
    statusFilter,
    setStatusFilter,
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
        <div className="flex items-center gap-3">
          <Select value={statusFilter} onValueChange={(v) => setStatusFilter(v ?? 'all')}>
            <SelectTrigger className="w-32 border-border/60 focus:ring-blue-500">
              <SelectValue placeholder="全部状态" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部状态</SelectItem>
              <SelectItem value="success">成功</SelectItem>
              <SelectItem value="error">错误</SelectItem>
              <SelectItem value="filtered">已过滤</SelectItem>
            </SelectContent>
          </Select>
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
