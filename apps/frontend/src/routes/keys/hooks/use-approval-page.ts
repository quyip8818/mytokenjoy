import { useCallback, useMemo } from 'react'
import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'
import type { KeyApproval } from '@/api/types'
import { useDemoRole } from '@/features/demo'
import { useFilteredResource } from '@/hooks/use-filtered-resource'
import { usePermissions } from '@/hooks/use-permissions'
import { useRowHighlight } from '@/hooks/use-row-highlight'
import { useWorkflowRefresh } from '@/hooks/use-workflow-refresh'
import { PERMISSION } from '@/lib/permissions'

type ApprovalTab = 'pending' | 'mine' | 'all'

export function useApprovalPage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const { memberId } = useDemoRole()
  const { has } = usePermissions()
  const canApprove = has(PERMISSION.BUDGET_APPROVE)
  const canSubmit = has(PERMISSION.SELF_APPROVAL)
  const { flashRow, rowClass } = useRowHighlight()
  const {
    data: approvals = [],
    loading,
    error,
    refresh,
    filter: tab,
    setFilter: setTab,
  } = useFilteredResource(
    (filter) =>
      apis.approvalApi.list({
        tab: filter === 'all' ? undefined : filter,
        memberId: filter === 'mine' ? memberId : undefined,
      }),
    'pending' as ApprovalTab,
  )
  const { openWithRefresh } = useWorkflowRefresh(refresh, flashRow)

  const pendingCount = useMemo(
    () => approvals.filter((a) => a.status === 'pending').length,
    [approvals],
  )

  const hasKeyType = useMemo(() => approvals.some((a) => a.type === 'key'), [approvals])
  const hasQuotaType = useMemo(() => approvals.some((a) => a.type === 'quota'), [approvals])

  const handleRowClick = useCallback(
    (approval: KeyApproval) => {
      if (!canApprove && approval.status === 'pending') return
      openWithRefresh('approval-review', { approval })
    },
    [canApprove, openWithRefresh],
  )

  const openSubmit = useCallback(() => openWithRefresh('approval-submit'), [openWithRefresh])

  return {
    approvals,
    loading,
    error,
    refresh,
    tab,
    setTab,
    canApprove,
    canSubmit,
    pendingCount,
    hasKeyType,
    hasQuotaType,
    rowClass,
    handleRowClick,
    openSubmit,
  }
}
