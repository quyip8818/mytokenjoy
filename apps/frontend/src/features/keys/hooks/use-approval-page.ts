import { useCallback, useMemo } from 'react'
import { toast } from 'sonner'
import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'
import type { KeyApproval } from '@/api/types'
import { queryKeys, useFilteredQuery } from '@/features/query'
import { usePermissions } from '@/features/session'
import { useRowHighlight } from '@/hooks/use-row-highlight'
import { useWorkflowRefresh } from '@/features/workflow'
import { PERMISSION } from '@/lib/permissions'
import type { ApprovalTab } from '../lib/types'

export type { ApprovalTab } from '../lib/types'

export function useApprovalPage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
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
  } = useFilteredQuery({
    injectedApis: apis,
    initialFilter: 'pending' as ApprovalTab,
    queryKeyFactory: (filter) => queryKeys.keys.approvals(filter),
    fetcher: (a, filter) =>
      a.approvalApi.list({
        tab: filter === 'all' ? 'all' : filter,
      }),
  })
  const { openWithRefresh } = useWorkflowRefresh({
    refresh,
    invalidateKeys: [queryKeys.keys.all],
    flashRow,
  })

  const pendingCount = useMemo(
    () => approvals.filter((approval) => approval.status === 'pending').length,
    [approvals],
  )

  const handleApprove = useCallback(
    async (id: string) => {
      await apis.approvalApi.approve(id)
      toast.success('已通过申请')
      flashRow(id)
      void refresh()
    },
    [apis, flashRow, refresh],
  )

  const handleReject = useCallback(
    async (id: string) => {
      await apis.approvalApi.reject(id, '已拒绝')
      toast.success('已拒绝申请')
      flashRow(id)
      void refresh()
    },
    [apis, flashRow, refresh],
  )

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
    rowClass,
    handleApprove,
    handleReject,
    handleRowClick,
    openSubmit,
  }
}
