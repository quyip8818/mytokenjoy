import { useCallback } from 'react'
import type { AppApis } from '@/api/app-apis'
import { useApis } from '@/api/use-apis'
import type { RoutingRule } from '@/api/types'
import { useAsyncResource } from '@/hooks/use-async-resource'
import { useWorkflowRefresh } from '@/hooks/use-workflow-refresh'

const PARENT_MAP: Record<string, string> = {
  'dept-2': 'dept-1',
  'dept-3': 'dept-2',
  'dept-4': 'dept-2',
  'dept-5': 'dept-2',
  'dept-6': 'dept-1',
  'dept-7': 'dept-1',
  'dept-8': 'dept-1',
}

export function useModelRoutingPage(injectedApis?: AppApis) {
  const ctxApis = useApis()
  const apis = injectedApis ?? ctxApis
  const {
    data: rules = [],
    loading,
    error,
    refresh,
  } = useAsyncResource(() => apis.routingApi.getRules(), [apis])
  const { openWithRefresh } = useWorkflowRefresh(refresh)

  const getParentCount = useCallback(
    (rule: RoutingRule) => {
      const parentId = PARENT_MAP[rule.nodeId]
      const parent = parentId ? rules.find((r) => r.nodeId === parentId) : undefined
      return parent?.allowedModels.length ?? rule.allowedModels.length
    },
    [rules],
  )

  const openWhitelistConfig = useCallback(
    (rule: RoutingRule) => {
      openWithRefresh('whitelist-config', { rule })
    },
    [openWithRefresh],
  )

  return {
    rules,
    loading,
    error,
    refresh,
    getParentCount,
    openWhitelistConfig,
  }
}
