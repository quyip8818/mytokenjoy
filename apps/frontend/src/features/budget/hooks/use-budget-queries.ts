import { useCallback, useMemo, useState } from 'react'
import type { AppApis } from '@/api/app-apis'
import type { ProjectView } from '@/api/types'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { getCurrentBudgetPeriod } from '@/lib/date'
import {
  findBudgetNode,
  formatBudgetPeriodLabel,
  mapProjectsToViews,
  shiftBudgetPeriod,
} from '../lib/mappers'

export function useBudgetQueries(injectedApis?: AppApis) {
  const [period, setPeriod] = useState(getCurrentBudgetPeriod)

  const {
    data: rawTree,
    loading: treeLoading,
    error: treeError,
    refresh: refreshTree,
  } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.budget.tree(period),
    queryFn: (api) => api.budgetApi.getTree(period),
  })
  const tree = useMemo(() => rawTree ?? [], [rawTree])

  const {
    data: projectsData = [],
    loading: projectsLoading,
    error: projectsError,
    refresh: refreshProjects,
  } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.budget.projects(),
    queryFn: (api) => api.budgetApi.getProjects(),
  })

  const { data: overrunPolicy } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.budget.overrunPolicy(),
    queryFn: (api) => api.budgetApi.getOverrunPolicy(),
  })

  const loading = treeLoading || projectsLoading
  const error = treeError ?? projectsError

  const refresh = useCallback(async () => {
    await Promise.all([refreshTree(), refreshProjects()])
  }, [refreshTree, refreshProjects])

  const nodeNameMap = useMemo(() => {
    const map = new Map<string, string>()
    function walk(nodes: typeof tree) {
      for (const node of nodes) {
        map.set(node.id, node.name)
        if (node.children) walk(node.children)
      }
    }
    walk(tree)
    return map
  }, [tree])

  const projects = useMemo((): ProjectView[] => {
    return mapProjectsToViews(projectsData, nodeNameMap, period)
  }, [projectsData, nodeNameMap, period])

  const periodLabel = useMemo(() => formatBudgetPeriodLabel(period), [period])

  return {
    tree,
    projectsData,
    projects,
    period,
    periodLabel,
    overrunPolicy,
    loading,
    error,
    refresh,
    shiftPeriod: (delta: number) => setPeriod((current) => shiftBudgetPeriod(current, delta)),
    findNode: (nodeId: string) => findBudgetNode(tree, nodeId),
  }
}
