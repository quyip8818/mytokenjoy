import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import type { AppApis } from '@/api/app-apis'
import type { BudgetProjectView, UpdateMemberBudgetInput } from '@/api/types'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { useInjectedApis } from '@/api/use-apis'
import { getCurrentBudgetPeriod } from '@/lib/date'
import {
  findBudgetNode,
  formatBudgetPeriodLabel,
  mapGroupsToProjectViews,
  groupsForDepartment,
  formatOverrunPolicyLabel,
  shiftBudgetPeriod,
} from '../lib/mappers'
import { filterProjectMembers, useBudgetDepartmentMembers } from './use-budget-department-members'

export function useBudgetPage(injectedApis?: AppApis) {
  const apis = useInjectedApis(injectedApis)
  const [period, setPeriod] = useState(getCurrentBudgetPeriod)
  const periodSyncedFromTree = useRef(false)
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
    queryFn: async (api) => (await api.budgetApi.getTree(period)) ?? [],
  })

  const {
    data: groups = [],
    loading: groupsLoading,
    error: groupsError,
    refresh: refreshGroups,
  } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.budget.groups(),
    queryFn: async (api) => (await api.budgetApi.getGroups()) ?? [],
  })

  const { data: overrunPolicy } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.budget.overrunPolicy(),
    queryFn: (api) => api.budgetApi.getOverrunPolicy(),
  })

  const { data: approvals = [], refresh: refreshApprovals } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.budget.approvals(),
    queryFn: async (api) => (await api.budgetApi.getApprovals()) ?? [],
  })

  const pendingCount = useMemo(
    () => (approvals ?? []).filter((item) => item.status === 'pending').length,
    [approvals],
  )

  useEffect(() => {
    if (!periodSyncedFromTree.current && tree[0]?.period) {
      setPeriod(tree[0].period)
      periodSyncedFromTree.current = true
    }
  }, [tree])

  const resolvedSelectedTeamId = selectedTeamId ?? tree[0]?.id

  const loading = treeLoading || groupsLoading
  const error = treeError ?? groupsError

  const refresh = useCallback(async () => {
    await Promise.all([refreshTree(), refreshGroups(), refreshApprovals()])
  }, [refreshTree, refreshGroups, refreshApprovals])

  const selectedNode = useMemo(
    () => (resolvedSelectedTeamId ? findBudgetNode(tree, resolvedSelectedTeamId) : null),
    [tree, resolvedSelectedTeamId],
  )

  const projects = useMemo((): BudgetProjectView[] => {
    const deptName = selectedNode?.name ?? ''
    return mapGroupsToProjectViews(groups, deptName, period)
  }, [groups, selectedNode?.name, period])

  const activeProject = useMemo(
    () =>
      activeProjectId ? projects.find((project) => project.id === activeProjectId) : undefined,
    [projects, activeProjectId],
  )

  const departmentIdForMembers = activeProject?.departmentId ?? selectedNode?.id
  const { departmentMembers, departmentMembersLoading } = useBudgetDepartmentMembers({
    injectedApis,
    departmentId: departmentIdForMembers,
  })

  const projectMembers = useMemo(
    () => (activeProject ? filterProjectMembers(departmentMembers, activeProject.memberIds) : []),
    [activeProject, departmentMembers],
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

  const createBudgetGroup = useCallback(
    async (data: {
      name: string
      budget: number
      memberIds: string[]
      departmentIds: string[]
    }) => {
      await apis.budgetApi.createGroup(data)
      await refresh()
    },
    [apis, refresh],
  )

  const updateBudgetGroup = useCallback(
    async (groupId: string, data: { budget?: number; memberIds?: string[] }) => {
      await apis.budgetApi.updateGroup(groupId, data)
      await refresh()
    },
    [apis, refresh],
  )

  const deleteBudgetGroup = useCallback(
    async (groupId: string) => {
      await apis.budgetApi.deleteGroup(groupId)
      await refresh()
    },
    [apis, refresh],
  )

  const getMemberBudgets = useCallback(
    (departmentId: string) => apis.budgetApi.getMemberBudgets(departmentId),
    [apis],
  )

  const updateMemberBudget = useCallback(
    async (memberId: string, data: UpdateMemberBudgetInput) => {
      const result = await apis.budgetApi.updateMemberBudget(memberId, data)
      return result
    },
    [apis],
  )

  const getDepartmentTree = useCallback(() => apis.departmentApi.getTree(), [apis])

  const getMembers = useCallback(
    async (departmentId: string) => {
      const result = await apis.memberApi.list({
        departmentId,
        directOnly: true,
        page: 1,
        pageSize: 200,
      })
      return result?.items ?? []
    },
    [apis],
  )

  const getAllDeptMembers = useCallback(
    async (departmentId: string) => {
      const result = await apis.memberApi.list({ departmentId, page: 1, pageSize: 200 })
      return result?.items ?? []
    },
    [apis],
  )

  const searchMembers = useCallback(
    async (keyword: string) => {
      const result = await apis.memberApi.list({ keyword, page: 1, pageSize: 50 })
      return result?.items ?? []
    },
    [apis],
  )

  const overrunPolicyLabel = formatOverrunPolicyLabel(
    activeProject?.overrunPolicy ?? projects[0]?.overrunPolicy ?? 'hard_reject',
  )

  return {
    tree,
    projects,
    period,
    periodLabel,
    selectedTeamId: resolvedSelectedTeamId,
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
    shiftPeriod: (delta: number) => setPeriod((current) => shiftBudgetPeriod(current, delta)),
    handleSelectTeam,
    setActiveProjectId,
    updateDepartment,
    groupsForNode: (departmentId: string) => groupsForDepartment(groups, departmentId),
    overrunPolicyLabel,
    departmentMembers,
    departmentMembersLoading,
    projectMembers,
    createBudgetGroup,
    updateBudgetGroup,
    deleteBudgetGroup,
    getMemberBudgets,
    updateMemberBudget,
    getDepartmentTree,
    getMembers,
    getAllDeptMembers,
    searchMembers,
  }
}
