import { useCallback, useState } from 'react'
import type { AppApis } from '@/api/app-apis'
import type { CallLog } from '@/api/types'
import { useAuditSettings } from '@/hooks/use-audit-settings'
import { AUDIT_DATE_PRESET } from '@/lib/audit-constants'
import { AUDIT_FILTER_ALL, buildCallsQuery, type AuditCallsFilter } from '@/lib/audit-query'
import { CALL_AUDIT_CSV_HEADERS, buildCallAuditCsvRows } from '@/lib/audit-export'
import { downloadCsv } from '@/lib/csv-export'
import { queryKeys } from '@/features/query'
import { useAuditListPage } from './use-audit-list-page'

const INITIAL_FILTER: AuditCallsFilter = {
  status: AUDIT_FILTER_ALL,
  callerId: AUDIT_FILTER_ALL,
  datePreset: AUDIT_DATE_PRESET.ALL,
  keyword: '',
}

export type { AuditCallsFilter }

export function useAuditCallsPage(injectedApis?: AppApis) {
  const {
    items: logs,
    filter,
    patchFilter,
    loading,
    error,
    refresh,
  } = useAuditListPage<AuditCallsFilter, CallLog, ReturnType<typeof buildCallsQuery>>({
    initialFilter: INITIAL_FILTER,
    toQueryParams: buildCallsQuery,
    fetchItems: (apis, query) => apis.auditApi.getCalls(query).then((res) => res.items),
    injectedApis,
    queryKeyFactory: (filter) => queryKeys.audit.calls(filter),
  })
  const { contentRetentionEnabled } = useAuditSettings(injectedApis)
  const [expandedId, setExpandedId] = useState<string | null>(null)

  const handleExport = useCallback(() => {
    downloadCsv('call-audit.csv', [...CALL_AUDIT_CSV_HEADERS], buildCallAuditCsvRows(logs))
  }, [logs])

  const toggleExpanded = useCallback((id: string) => {
    setExpandedId((current) => (current === id ? null : id))
  }, [])

  return {
    logs,
    loading,
    error,
    refresh,
    statusFilter: filter.status,
    callerId: filter.callerId,
    datePreset: filter.datePreset,
    keyword: filter.keyword,
    setStatusFilter: (status: string) => patchFilter({ status }),
    setCallerId: (callerId: string) => patchFilter({ callerId }),
    setDatePreset: (datePreset: string) => patchFilter({ datePreset }),
    setKeyword: (keyword: string) => patchFilter({ keyword }),
    expandedId,
    contentRetentionEnabled,
    handleExport,
    toggleExpanded,
  }
}
