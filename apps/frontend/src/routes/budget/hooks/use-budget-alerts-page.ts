import { useMemo } from 'react'
import type { AppApis } from '@/api/app-apis'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { useCtaHighlight } from '@/hooks/use-cta-highlight'
import { useWorkflowRefresh } from '@/hooks/use-workflow-refresh'

export function useBudgetAlertsPage(injectedApis?: AppApis) {
  const overrunCta = useCtaHighlight('OVERRUN')
  const {
    data: policy = null,
    loading,
    error,
    refresh,
  } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.budget.overrunPolicy(),
    queryFn: (apis) => apis.budgetApi.getOverrunPolicy(),
  })
  const { openWithRefresh } = useWorkflowRefresh({
    refresh,
    invalidateKeys: [queryKeys.budget.overrunPolicy()],
  })

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
