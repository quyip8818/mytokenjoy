import { useCallback, useMemo } from 'react'
import { toast } from '@/lib/toast'
import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'
import { withErrorToast } from '@/lib/api-error-toast'
import { useFilteredQuery } from '@/features/query'
import { usePermissions } from '@/features/session'
import { PERMISSION } from '@/lib/permission-keys'
import { approvalKeys } from '../lib/query-keys'
import type { ApprovalTab } from '../lib/types'

export function useApprovalPage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const { has } = usePermissions()
  const canSubmit = has(PERMISSION.SELF_APPROVAL)

  const {
    data,
    loading,
    error,
    refresh,
    filter: tab,
    setFilter: setTab,
  } = useFilteredQuery({
    injectedApis: apis,
    initialFilter: 'pending' as ApprovalTab,
    queryKeyFactory: (filter) => approvalKeys.list(filter),
    fetcher: (a, filter) =>
      a.approvalApi.list({
        status: filter === 'all' ? undefined : filter,
      }),
  })

  const approvals = useMemo(() => data?.items ?? [], [data])
  const total = data?.total ?? 0

  // When tab='pending', all items in the list are pending so count === total.
  // No need for a separate client-side filter.
  const pendingCount = tab === 'pending' ? total : 0

  const handleApprove = useCallback(
    async (id: string) => {
      await withErrorToast(async () => {
        await apis.approvalApi.approve(id)
        toast.success('已通过申请')
        void refresh()
      }, '审批失败')
    },
    [apis, refresh],
  )

  const handleReject = useCallback(
    async (id: string, reason: string) => {
      await withErrorToast(async () => {
        await apis.approvalApi.reject(id, reason)
        toast.success('已拒绝申请')
        void refresh()
      }, '拒绝失败')
    },
    [apis, refresh],
  )

  const handleCancel = useCallback(
    async (id: string) => {
      await withErrorToast(async () => {
        await apis.approvalApi.cancel(id)
        toast.success('已撤回申请')
        void refresh()
      }, '撤回失败')
    },
    [apis, refresh],
  )

  const handleRetry = useCallback(
    async (id: string) => {
      await withErrorToast(async () => {
        await apis.approvalApi.retry(id)
        toast.success('重试成功')
        void refresh()
      }, '重试失败')
    },
    [apis, refresh],
  )

  return {
    approvals,
    total,
    loading,
    error,
    refresh,
    tab,
    setTab,
    canSubmit,
    pendingCount,
    handleApprove,
    handleReject,
    handleCancel,
    handleRetry,
  }
}
