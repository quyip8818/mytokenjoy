import { useCallback, useMemo } from 'react'
import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'
import type { RoutingRule } from '@/api/types'
import { useAsyncResource } from '@/hooks/use-async-resource'
import { useWorkflowRefresh } from '@/hooks/use-workflow-refresh'
import { findParentDeptId } from '@/lib/org'

export function useModelRoutingPage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const { data, loading, error, refresh } = useAsyncResource(
    () =>
      Promise.all([apis.routingApi.getRules(), apis.departmentApi.getTree()]).then(
        ([rules, departments]) => ({ rules, departments }),
      ),
    [apis],
  )
  const rules = useMemo(() => data?.rules ?? [], [data?.rules])
  const departments = useMemo(() => data?.departments ?? [], [data?.departments])
  const { openWithRefresh } = useWorkflowRefresh(refresh)

  const getParentCount = useCallback(
    (rule: RoutingRule) => {
      const parentId = findParentDeptId(departments, rule.nodeId)
      const parent = parentId ? rules.find((r) => r.nodeId === parentId) : undefined
      return parent?.allowedModels.length ?? rule.allowedModels.length
    },
    [rules, departments],
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
