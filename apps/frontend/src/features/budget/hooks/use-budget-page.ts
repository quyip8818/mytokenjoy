import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import type { AppApis } from '@/api/app-apis'
import type { PlatformKeyScope, ProjectView, UpdateMemberBudgetInput } from '@/api/types'
import { queryKeys, useInjectedQuery } from '@/features/query'
import { useInjectedApis } from '@/api/use-apis'
import { ApiError } from '@/api/client'
import { toast } from 'sonner'
import { useWorkflowRefresh } from '@/features/workflow'
import { getCurrentBudgetPeriod } from '@/lib/date'
import {
  findBudgetNode,
  formatBudgetPeriodLabel,
  mapProjectsToViews,
  projectsForDepartment,
  formatOverrunPolicyLabel,
  shiftBudgetPeriod,
  DEFAULT_OVERRUN_POLICY,
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
    data: projectsData = [],
    loading: projectsLoading,
    error: projectsError,
    refresh: refreshProjects,
  } = useInjectedQuery({
    injectedApis,
    queryKey: queryKeys.budget.projects(),
    queryFn: async (api) => (await api.budgetApi.getProjects()) ?? [],
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

  const loading = treeLoading || projectsLoading
  const error = treeError ?? projectsError

  const refresh = useCallback(async () => {
    await Promise.all([refreshTree(), refreshProjects(), refreshApprovals()])
  }, [refreshTree, refreshProjects, refreshApprovals])

  const { openWithRefresh } = useWorkflowRefresh({
    refresh,
    invalidateKeys: [queryKeys.keys.all],
  })

  const selectedNode = useMemo(
    () => (resolvedSelectedTeamId ? findBudgetNode(tree, resolvedSelectedTeamId) : null),
    [tree, resolvedSelectedTeamId],
  )

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
      try {
        await apis.budgetApi.updateDepartment(departmentId, data)
        await refresh()
      } catch (err) {
        toast.error(err instanceof ApiError ? err.message : '更新部门预算失败')
        throw err
      }
    },
    [apis, refresh],
  )

  const resolveApproval = useCallback(
    async (id: string, data: { status: 'approved' | 'rejected'; rejectReason?: string }) => {
      try {
        await apis.budgetApi.resolveApproval(id, data)
        await refreshApprovals()
      } catch (err) {
        toast.error(err instanceof ApiError ? err.message : '审批操作失败')
        throw err
      }
    },
    [apis, refreshApprovals],
  )

  const createProject = useCallback(
    async (data: {
      name: string
      budget: number
      memberIds: string[]
      ownerDepartmentId: string
    }) => {
      try {
        await apis.budgetApi.createProject(data)
        await refresh()
      } catch (err) {
        toast.error(err instanceof ApiError ? err.message : '创建项目失败')
        throw err
      }
    },
    [apis, refresh],
  )

  const updateProject = useCallback(
    async (
      groupId: string,
      data: { budget?: number; memberIds?: string[]; memberBudgets?: Record<string, number> },
    ) => {
      try {
        await apis.budgetApi.updateProject(groupId, data)
        await refresh()
      } catch (err) {
        toast.error(err instanceof ApiError ? err.message : '更新项目失败')
        throw err
      }
    },
    [apis, refresh],
  )

  const openCreateProjectKey = useCallback(
    (project: ProjectView, scope: PlatformKeyScope, memberId?: string, memberName?: string) => {
      openWithRefresh('key-create', {
        adminCreate: true,
        scope,
        projectId: project.id,
        projectName: project.name,
        targetMemberId: memberId,
        initialName: memberName ? `${memberName}-项目 Key` : `${project.name}-项目 Key`,
      })
    },
    [openWithRefresh],
  )

  const deleteProject = useCallback(
    async (groupId: string) => {
      try {
        await apis.budgetApi.deleteProject(groupId)
        await refresh()
      } catch (err) {
        toast.error(err instanceof ApiError ? err.message : '删除项目失败')
        throw err
      }
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

  const applyAverageBudget = useCallback(
    async (departmentId: string, data: { personalBudget: number; recursive: boolean }) => {
      await apis.budgetApi.applyAverageBudget(departmentId, data)
      await refresh()
    },
    [apis, refresh],
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
    activeProject?.overrunPolicy ?? DEFAULT_OVERRUN_POLICY,
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
    projectsForNode: (departmentId: string) => projectsForDepartment(projectsData, departmentId),
    overrunPolicyLabel,
    departmentMembers,
    departmentMembersLoading,
    projectMembers,
    createProject,
    updateProject,
    deleteProject,
    openCreateProjectKey,
    getMemberBudgets,
    updateMemberBudget,
    applyAverageBudget,
    getDepartmentTree,
    getMembers,
    getAllDeptMembers,
    searchMembers,
  }
}
