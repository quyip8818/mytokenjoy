import { useCallback, useMemo } from 'react'
import type { AppApis } from '@/api/app-apis'
import { useApis } from '@/api/use-apis'
import type { BudgetNode } from '@/api/types'
import { useDemoCta } from '@/features/demo'
import { useAsyncResource } from '@/hooks/use-async-resource'
import { usePermissions } from '@/hooks/use-permissions'
import { useWorkflowRefresh } from '@/hooks/use-workflow-refresh'
import { computeUnallocated, formatBudgetPeriodLabel } from '@/lib/budget'
import { PERMISSION } from '@/lib/permissions'

export function useBudgetOverviewPage(injectedApis?: AppApis) {
  const ctxApis = useApis()
  const apis = injectedApis ?? ctxApis
  const budgetCta = useDemoCta('BUDGET')
  const { has } = usePermissions()
  const canAllocate = has(PERMISSION.BUDGET_ALLOCATE)
  const {
    data: tree = [],
    loading,
    error,
    refresh,
  } = useAsyncResource(() => apis.budgetApi.getTree(), [apis])
  const { openWithRefresh } = useWorkflowRefresh(refresh)

  const summary = useMemo(() => {
    const root = tree[0]
    if (!root) {
      return { budget: 0, consumed: 0, unallocated: 0 }
    }
    return {
      budget: root.budget,
      consumed: root.consumed,
      unallocated: computeUnallocated(root),
    }
  }, [tree])

  const periodLabel = useMemo(() => formatBudgetPeriodLabel(tree[0]?.period), [tree])

  const handleAllocate = useCallback(
    (node: BudgetNode, parent: BudgetNode | null) => {
      openWithRefresh('budget-node-edit', { node, parent })
    },
    [openWithRefresh],
  )

  const handleMemberQuota = useCallback(
    (node: BudgetNode) => {
      openWithRefresh('member-quota-config', {
        departmentId: node.id,
        departmentName: node.name,
      })
    },
    [openWithRefresh],
  )

  return {
    tree,
    loading,
    error,
    refresh,
    summary,
    periodLabel,
    canAllocate,
    budgetCta,
    handleAllocate,
    handleMemberQuota,
  }
}
