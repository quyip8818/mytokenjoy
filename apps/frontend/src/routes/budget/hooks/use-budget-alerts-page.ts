import { useMemo } from 'react'
import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'
import { useCtaHighlight } from '@/hooks/use-cta-highlight'
import { useAsyncResource } from '@/hooks/use-async-resource'
import { useWorkflowRefresh } from '@/hooks/use-workflow-refresh'

export function useBudgetAlertsPage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const overrunCta = useCtaHighlight('OVERRUN')
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
