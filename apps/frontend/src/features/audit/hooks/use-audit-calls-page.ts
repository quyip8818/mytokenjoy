import { useCallback, useMemo, useState } from 'react'
import type { AppApis } from '@/api/app-apis'
import type { CallLog } from '@/api/types'
import { AUDIT_DATE_PRESET } from '../lib/constants'
import { AUDIT_FILTER_ALL, buildCallsQuery, type AuditCallsFilter } from '../lib/query'
import { CALL_AUDIT_CSV_HEADERS, buildCallAuditCsvRows } from '../lib/export'
import { useAuditListPage } from './use-audit-list-page'
import { useAuditSettings } from './use-audit-settings'
import { downloadCsv } from '@/lib/csv-export'
import { queryKeys } from '@/features/query'
import { useAuditModelOptions } from './use-audit-model-options'
import { useAuditMemberOptions } from './use-audit-member-options'

const INITIAL_FILTER: AuditCallsFilter = {
  status: AUDIT_FILTER_ALL,
  callerId: AUDIT_FILTER_ALL,
  model: AUDIT_FILTER_ALL,
  datePreset: AUDIT_DATE_PRESET.ALL,
  keyword: '',
}

export type { AuditCallsFilter }

export function useAuditCallsPage(injectedApis?: AppApis) {
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
  } = useAuditListPage<AuditCallsFilter, CallLog, ReturnType<typeof buildCallsQuery>>({
    initialFilter: INITIAL_FILTER,
    toQueryParams: buildCallsQuery,
    fetchPage: (apis, query) => apis.auditApi.getCalls(query),
    injectedApis,
    queryKeyFactory: ({ filter, page: currentPage }) =>
      queryKeys.audit.calls({ filter, page: currentPage }),
  })
  const { contentRetentionEnabled } = useAuditSettings(injectedApis)
  const { models } = useAuditModelOptions(injectedApis)
  const { members } = useAuditMemberOptions(injectedApis)
  const modelOptions = useMemo(
    () => Object.fromEntries(models.map((model) => [model.name, model.displayName])),
    [models],
  )
  const memberOptions = useMemo(
    () => Object.fromEntries(members.map((member) => [member.id, member.name])),
    [members],
  )
  const [expandedId, setExpandedId] = useState<string | null>(null)

  const handleExport = useCallback(() => {
    downloadCsv('call-audit.csv', [...CALL_AUDIT_CSV_HEADERS], buildCallAuditCsvRows(logs))
  }, [logs])

  const toggleExpanded = useCallback((id: string) => {
    setExpandedId((current) => (current === id ? null : id))
  }, [])

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
    statusFilter: filter.status,
    callerId: filter.callerId,
    modelFilter: filter.model,
    datePreset: filter.datePreset,
    keyword: filter.keyword,
    setStatusFilter: (status: string) => patchFilter({ status }),
    setCallerId: (callerId: string) => patchFilter({ callerId }),
    setModelFilter: (model: string) => patchFilter({ model }),
    setDatePreset: (datePreset: string) => patchFilter({ datePreset }),
    setKeyword: (keyword: string) => patchFilter({ keyword }),
    expandedId,
    contentRetentionEnabled,
    modelOptions,
    memberOptions,
    handleExport,
    toggleExpanded,
  }
}
