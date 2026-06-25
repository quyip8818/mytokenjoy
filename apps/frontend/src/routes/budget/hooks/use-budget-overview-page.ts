import { useCallback, useMemo } from 'react'
import type { AppApis } from '@/api/app-apis'
import type { BudgetNode } from '@/api/types'
import { queryKeys } from '@/features/query'
import { useCtaHighlight } from '@/hooks/use-cta-highlight'
import { usePermissions } from '@/hooks/use-permissions'
import { useWorkflowRefresh } from '@/hooks/use-workflow-refresh'
import { computeUnallocated, formatBudgetPeriodLabel } from '@/lib/budget'
import { PERMISSION } from '@/lib/permissions'
import { useBudgetTreeQuery } from './use-budget-tree-query'

export function useBudgetOverviewPage(injectedApis?: AppApis) {
  const budgetCta = useCtaHighlight('BUDGET')
  const { has } = usePermissions()
  const canAllocate = has(PERMISSION.BUDGET_ALLOCATE)
  const { data: tree = [], loading, error, refresh } = useBudgetTreeQuery(injectedApis)
  const { openWithRefresh } = useWorkflowRefresh({
    refresh,
    invalidateKeys: [queryKeys.budget.all],
  })

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
