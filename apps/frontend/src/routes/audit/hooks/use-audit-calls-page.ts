import { useCallback, useState } from 'react'
import type { AppApis } from '@/api/app-apis'
import { useApis } from '@/api/use-apis'
import { useFilteredResource } from '@/hooks/use-filtered-resource'
import { CALL_LOG_STATUS_LABELS } from '@/lib/labels'
import { downloadCsv } from '@/lib/csv-export'
import { useAuditSettings } from '@/hooks/use-audit-settings'

export function useAuditCallsPage(injectedApis?: AppApis) {
  const ctxApis = useApis()
  const apis = injectedApis ?? ctxApis
  const [expandedId, setExpandedId] = useState<string | null>(null)
  const { contentRetentionEnabled } = useAuditSettings(apis)
  const {
    data: logs = [],
    loading,
    error,
    refresh,
    filter: statusFilter,
    setFilter: setStatusFilter,
  } = useFilteredResource(async (filter) => {
    const params = filter !== 'all' ? { status: filter } : undefined
    const res = await apis.auditApi.getCalls(params)
    return res.items
  }, 'all')

  const handleExport = useCallback(() => {
    downloadCsv(
      'call-audit.csv',
      ['时间', '调用方', '类型', '模型', '输入Token', '输出Token', '延迟', '费用', '状态'],
      logs.map((log) => [
        log.createdAt,
        log.caller,
        log.callerType === 'member' ? '成员' : '应用',
        log.model,
        log.inputTokens,
        log.outputTokens,
        `${log.latencyMs}ms`,
        log.cost.toFixed(2),
        CALL_LOG_STATUS_LABELS[log.status] ?? log.status,
      ]),
    )
  }, [logs])

  const toggleExpanded = useCallback(
    (logId: string) => {
      if (!contentRetentionEnabled) return
      setExpandedId((current) => (current === logId ? null : logId))
    },
    [contentRetentionEnabled],
  )

  return {
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
  }
}
