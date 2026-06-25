import { useCallback } from 'react'
import type { AppApis } from '@/api/app-apis'
import { useApis } from '@/api/use-apis'
import { useFilteredResource } from '@/hooks/use-filtered-resource'
import { OPERATION_ACTION_LABELS } from '@/lib/labels'
import { downloadCsv } from '@/lib/csv-export'

export function useAuditOperationsPage(injectedApis?: AppApis) {
  const ctxApis = useApis()
  const apis = injectedApis ?? ctxApis
  const {
    data: logs = [],
    loading,
    error,
    refresh,
    filter: actionFilter,
    setFilter: setActionFilter,
  } = useFilteredResource(async (filter) => {
    const params = filter !== 'all' ? { action: filter } : undefined
    const res = await apis.auditApi.getOperations(params)
    return res.items
  }, 'all')

  const handleExport = useCallback(() => {
    downloadCsv(
      'operation-audit.csv',
      ['时间', '操作类型', '操作人', '操作对象', '详情', 'IP'],
      logs.map((log) => [
        log.createdAt,
        OPERATION_ACTION_LABELS[log.action] ?? log.action,
        log.operator,
        log.target,
        log.detail,
        log.ip,
      ]),
    )
  }, [logs])

  return {
    logs,
    loading,
    error,
    refresh,
    actionFilter,
    setActionFilter,
    handleExport,
  }
}
