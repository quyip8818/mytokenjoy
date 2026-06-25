import { useCallback } from 'react'
import { toast } from 'sonner'
import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'
import type { BudgetGroup } from '@/api/types'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { usePermissions } from '@/hooks/use-permissions'
import { useRowHighlight } from '@/hooks/use-row-highlight'
import { useWorkflowRefresh } from '@/hooks/use-workflow-refresh'
import { useBudgetTreeQuery } from './use-budget-tree-query'

export function useBudgetAllocationPage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const { flashRow, rowClass } = useRowHighlight()
  const { canWrite } = usePermissions()
  const {
    data: groups = [],
    loading: groupsLoading,
    error: groupsError,
    refresh: refreshGroups,
  } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.budget.groups(),
    queryFn: (apis) => apis.budgetApi.getGroups(),
  })
  const {
    data: tree = [],
    loading: treeLoading,
    error: treeError,
    refresh: refreshTree,
  } = useBudgetTreeQuery(injectedApis)

  const loading = groupsLoading || treeLoading
  const error = groupsError ?? treeError
  const refresh = useCallback(async () => {
    await Promise.all([refreshGroups(), refreshTree()])
  }, [refreshGroups, refreshTree])

  const { openWithRefresh } = useWorkflowRefresh({
    refresh,
    invalidateKeys: [queryKeys.budget.all],
    flashRow,
  })

  const handleDelete = useCallback(
    async (id: string) => {
      await apis.budgetApi.deleteGroup(id)
      toast.success('预算组已删除')
      void refresh()
    },
    [apis, refresh],
  )

  const openForm = useCallback(
    (group?: BudgetGroup) => {
      openWithRefresh('budget-group-form', { group, tree })
    },
    [openWithRefresh, tree],
  )

  const openGroupKeys = useCallback(
    (group: BudgetGroup) => {
      openWithRefresh('key-create', {
        adminCreate: true,
        budgetGroupId: group.id,
        budgetGroupName: group.name,
      })
    },
    [openWithRefresh],
  )

  return {
    groups,
    loading,
    error,
    refresh,
    canWrite,
    rowClass,
    handleDelete,
    openForm,
    openGroupKeys,
  }
}
