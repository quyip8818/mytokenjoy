import { useCallback, useMemo } from 'react'
import type { AppApis } from '@/api/app-apis'
import type { RoutingRule } from '@/api/types'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { useWorkflowRefresh } from '@/hooks/use-workflow-refresh'
import { findParentDeptId } from '@/features/org/lib/departments'

export function useModelRoutingPage(injectedApis?: AppApis) {
  const { data, loading, error, refresh } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.models.routing(),
    queryFn: async (apis) => {
      const [rules, departments] = await Promise.all([
        apis.routingApi.getRules(),
        apis.departmentApi.getTree(),
      ])
      return { rules, departments }
    },
  })
  const rules = useMemo(() => data?.rules ?? [], [data?.rules])
  const departments = useMemo(() => data?.departments ?? [], [data?.departments])
  const { openWithRefresh } = useWorkflowRefresh({
    refresh,
    invalidateKeys: [queryKeys.models.all],
  })

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
