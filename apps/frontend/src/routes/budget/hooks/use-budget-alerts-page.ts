import { useMemo } from 'react'
import type { AppApis } from '@/api/app-apis'
import { useApis } from '@/api/use-apis'
import { useDemoCta } from '@/features/demo'
import { useAsyncResource } from '@/hooks/use-async-resource'
import { useWorkflowRefresh } from '@/hooks/use-workflow-refresh'

export function useBudgetAlertsPage(injectedApis?: AppApis) {
  const ctxApis = useApis()
  const apis = injectedApis ?? ctxApis
  const overrunCta = useDemoCta('OVERRUN')
  const {
    data: policy = null,
    loading,
    error,
    refresh,
  } = useAsyncResource(() => apis.budgetApi.getOverrunPolicy(), [apis])
  const { openWithRefresh } = useWorkflowRefresh(refresh)

  const notifyLabels = useMemo(
    () =>
      [
        policy?.notifyEmail && '邮箱',
        policy?.notifyPhone && '手机',
        policy?.notifyIm && 'IM',
      ].filter(Boolean) as string[],
    [policy],
  )

  const openEditPolicy = () => openWithRefresh('overrun-policy')

  return {
    policy,
    loading,
    error,
    refresh,
    overrunCta,
    notifyLabels,
    openEditPolicy,
  }
}
