import { useCallback, useMemo } from 'react'
import { toast } from 'sonner'
import type { AppApis } from '@/api/app-apis'
import { useApis } from '@/api/use-apis'
import type { BudgetGroup } from '@/api/types'
import { useAsyncResource } from '@/hooks/use-async-resource'
import { usePermissions } from '@/hooks/use-permissions'
import { useRowHighlight } from '@/hooks/use-row-highlight'
import { useWorkflowRefresh } from '@/hooks/use-workflow-refresh'

export function useBudgetAllocationPage(injectedApis?: AppApis) {
  const ctxApis = useApis()
  const apis = injectedApis ?? ctxApis
  const { flashRow, rowClass } = useRowHighlight()
  const { canWrite } = usePermissions()
  const { data, loading, error, refresh } = useAsyncResource(async () => {
    const [groups, tree] = await Promise.all([apis.budgetApi.getGroups(), apis.budgetApi.getTree()])
    return { groups, tree }
  }, [apis])
  const groups = data?.groups ?? []
  const tree = useMemo(() => data?.tree ?? [], [data?.tree])
  const { openWithRefresh } = useWorkflowRefresh(refresh, flashRow)

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
