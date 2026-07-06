import { useCallback, useEffect, useMemo, useState } from 'react'
import type { AppApis } from '@/api/app-apis'
import type { BudgetProjectView } from '@/api/types'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { useInjectedApis } from '@/api/use-apis'
import {
  findBudgetNode,
  formatBudgetPeriodLabel,
  mapGroupsToProjectViews,
  groupsForDepartment,
  formatOverrunPolicyLabel,
} from '@/lib/budget'

function shiftPeriod(period: string, delta: number) {
  const [y, m] = period.split('-').map(Number)
  const d = new Date(y, m - 1 + delta, 1)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
}

export function useBudgetPage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const [period, setPeriod] = useState('2026-06')
  const [selectedTeamId, setSelectedTeamId] = useState<string | undefined>()
  const [activeProjectId, setActiveProjectId] = useState<string | null>(null)
  const [approvalOpen, setApprovalOpen] = useState(false)

  const {
    data: tree = [],
    loading: treeLoading,
    error: treeError,
    refresh: refreshTree,
  } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.budget.tree(period),
    queryFn: (api) => api.budgetApi.getTree(period),
  })

  const {
    data: groups = [],
    loading: groupsLoading,
    error: groupsError,
    refresh: refreshGroups,
  } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.budget.groups(),
    queryFn: (api) => api.budgetApi.getGroups(),
  })

  const { data: overrunPolicy } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.budget.overrunPolicy(),
    queryFn: (api) => api.budgetApi.getOverrunPolicy(),
  })

  const { data: approvals = [], refresh: refreshApprovals } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.budget.approvals(),
    queryFn: (api) => api.budgetApi.getApprovals(),
  })

  const pendingCount = useMemo(
    () => approvals.filter((item) => item.status === 'pending').length,
    [approvals],
  )

  useEffect(() => {
    if (tree[0]?.period && period === '2026-06') {
      setPeriod(tree[0].period)
    }
  }, [tree, period])

  useEffect(() => {
    if (!selectedTeamId && tree.length > 0) {
      setSelectedTeamId(tree[0].id)
    }
  }, [selectedTeamId, tree])

  const loading = treeLoading || groupsLoading
  const error = treeError ?? groupsError

  const refresh = useCallback(async () => {
    await Promise.all([refreshTree(), refreshGroups(), refreshApprovals()])
  }, [refreshTree, refreshGroups, refreshApprovals])

  const selectedNode = useMemo(
    () => (selectedTeamId ? findBudgetNode(tree, selectedTeamId) : null),
    [tree, selectedTeamId],
  )

  const projects = useMemo((): BudgetProjectView[] => {
    const deptName = selectedNode?.name ?? ''
    return mapGroupsToProjectViews(groups, deptName, period)
  }, [groups, selectedNode?.name, period])

  const activeProject = useMemo(
    () => (activeProjectId ? projects.find((project) => project.id === activeProjectId) : undefined),
    [projects, activeProjectId],
  )

  const periodLabel = useMemo(() => formatBudgetPeriodLabel(period), [period])

  const handleSelectTeam = useCallback((nodeId: string) => {
    setSelectedTeamId(nodeId)
    setActiveProjectId(null)
  }, [])

  const updateDepartment = useCallback(
    async (departmentId: string, data: { budget: number; reservedPool?: number }) => {
      await apis.budgetApi.updateDepartment(departmentId, data)
      await refresh()
    },
    [apis, refresh],
  )

  const resolveApproval = useCallback(
    async (id: string, data: { status: 'approved' | 'rejected'; rejectReason?: string }) => {
      await apis.budgetApi.resolveApproval(id, data)
      await refreshApprovals()
    },
    [apis, refreshApprovals],
  )

  const overrunPolicyLabel = formatOverrunPolicyLabel(
    activeProject?.overrunPolicy ?? projects[0]?.overrunPolicy ?? 'hard_reject',
  )

  return {
    tree,
    projects,
    period,
    periodLabel,
    selectedTeamId,
    selectedNode,
    activeProjectId,
    activeProject,
    approvalOpen,
    setApprovalOpen,
    pendingCount,
    approvals,
    refreshApprovals,
    resolveApproval,
    loading,
    error,
    refresh,
    overrunPolicy,
    shiftPeriod: (delta: number) => setPeriod((current) => shiftPeriod(current, delta)),
    handleSelectTeam,
    setActiveProjectId,
    updateDepartment,
    groupsForNode: (departmentId: string) => groupsForDepartment(groups, departmentId),
    overrunPolicyLabel,
  }
}
