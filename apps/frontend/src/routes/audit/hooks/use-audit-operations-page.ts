import { useCallback } from 'react'
import type { AppApis } from '@/api/app-apis'
import type { OperationLog } from '@/api/types'
import { AUDIT_DATE_PRESET } from '@/lib/audit-constants'
import {
  AUDIT_FILTER_ALL,
  buildOperationsQuery,
  type AuditOperationsFilter,
} from '@/lib/audit-query'
import { OPERATION_ACTION_LABELS } from '@/lib/labels'
import { downloadCsv } from '@/lib/csv-export'
import { useAuditListPage } from './use-audit-list-page'

const INITIAL_FILTER: AuditOperationsFilter = {
  action: AUDIT_FILTER_ALL,
  datePreset: AUDIT_DATE_PRESET.ALL,
  operatorId: AUDIT_FILTER_ALL,
  keyword: '',
}

export type { AuditOperationsFilter }

export function useAuditOperationsPage(injectedApis?: AppApis) {
  const {
    items: logs,
    filter,
    patchFilter,
    loading,
    error,
    refresh,
  } = useAuditListPage<
    AuditOperationsFilter,
    OperationLog,
    ReturnType<typeof buildOperationsQuery>
  >({
    initialFilter: INITIAL_FILTER,
    toQueryParams: buildOperationsQuery,
    fetchItems: (apis, query) => apis.auditApi.getOperations(query).then((res) => res.items),
    injectedApis,
  })

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
    actionFilter: filter.action,
    datePreset: filter.datePreset,
    operatorId: filter.operatorId,
    keyword: filter.keyword,
    setActionFilter: (action: string) => patchFilter({ action }),
    setDatePreset: (datePreset: string) => patchFilter({ datePreset }),
    setOperatorId: (operatorId: string) => patchFilter({ operatorId }),
    setKeyword: (keyword: string) => patchFilter({ keyword }),
    handleExport,
  }
}
