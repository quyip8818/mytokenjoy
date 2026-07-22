import { useCallback, useMemo } from 'react'
import type { AppApis } from '@/api/app-apis'
import type { OperationLog } from '@/api/types'
import { AUDIT_DATE_PRESET } from '../lib/constants'
import { AUDIT_FILTER_ALL, buildOperationsQuery, type AuditOperationsFilter } from '../lib/query'
import { OPERATION_AUDIT_CSV_HEADERS, buildOperationAuditCsvRows } from '../lib/export'
import { useAuditListPage } from './use-audit-list-page'
import { downloadCsv } from '@/lib/csv-export'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { useAuditMemberOptions } from './use-audit-member-options'

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
    total,
    page,
    pageSize,
    totalPages,
    setPage,
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
    fetchPage: (apis, query) => apis.auditApi.getOperations(query),
    injectedApis,
    queryKeyFactory: ({ filter, page: currentPage }) =>
      queryKeys.audit.operations({ filter, page: currentPage }),
  })

  const { members } = useAuditMemberOptions(injectedApis)
  const memberOptions = useMemo(
    () => Object.fromEntries(members.map((member) => [member.id, member.alias])),
    [members],
  )

  const timelineQuery = useMemo(() => buildOperationsQuery(filter), [filter])
  const { data: timeline = [], loading: timelineLoading } = useInjectedQuery({
    injectedApis,
    queryKey: [...queryKeys.audit.operations({ filter, page: 0 }), 'timeline'],
    queryFn: (apis) => apis.auditApi.getOperationsTimeline(timelineQuery),
  })

  const handleExport = useCallback(() => {
    downloadCsv(
      'operation-audit.csv',
      [...OPERATION_AUDIT_CSV_HEADERS],
      buildOperationAuditCsvRows(logs),
    )
  }, [logs])

  return {
    logs,
    total,
    page,
    pageSize,
    totalPages,
    setPage,
    loading,
    error,
    refresh,
    timeline,
    timelineLoading,
    actionFilter: filter.action,
    datePreset: filter.datePreset,
    operatorId: filter.operatorId,
    keyword: filter.keyword,
    setActionFilter: (action: string) => patchFilter({ action }),
    setDatePreset: (datePreset: string) => patchFilter({ datePreset }),
    setOperatorId: (operatorId: string) => patchFilter({ operatorId }),
    setKeyword: (keyword: string) => patchFilter({ keyword }),
    memberOptions,
    handleExport,
  }
}
