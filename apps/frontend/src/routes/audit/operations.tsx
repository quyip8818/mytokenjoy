import { ScrollText } from 'lucide-react'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { AuditFilteredPage } from '@/components/audit/audit-filtered-page'
import { AuditToolbar } from '@/components/audit/audit-toolbar'
import { OperationsLogTable } from '@/routes/audit/components/operations-log-table'
import { useAuditOperationsPage } from '@/routes/audit/hooks/use-audit-operations-page'

export default function OperationLogsPage() {
  const { logs, loading, error, refresh, actionFilter, setActionFilter, handleExport } =
    useAuditOperationsPage()

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
        <div className="flex items-center gap-3">
          <Select value={actionFilter} onValueChange={(v) => setActionFilter(v ?? 'all')}>
            <SelectTrigger className="w-40 border-border/60 focus:ring-blue-500">
              <SelectValue placeholder="全部类型" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部类型</SelectItem>
              <SelectItem value="key_create">Key 创建</SelectItem>
              <SelectItem value="key_disable">Key 禁用</SelectItem>
              <SelectItem value="budget_change">预算变更</SelectItem>
              <SelectItem value="budget_approve">预算审批</SelectItem>
              <SelectItem value="permission_change">权限变更</SelectItem>
              <SelectItem value="model_whitelist_change">白名单变更</SelectItem>
              <SelectItem value="org_structure_change">组织结构变更</SelectItem>
            </SelectContent>
          </Select>
          <AuditToolbar onExport={handleExport} />
        </div>
      }
    >
      <OperationsLogTable logs={logs} />
    </AuditFilteredPage>
  )
}
