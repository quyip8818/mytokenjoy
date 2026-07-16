import { useCallback } from 'react'
import type { AppApis } from '@/api/app-apis'
import type { PlatformKeyScope, ProjectView, UpdateMemberBudgetInput } from '@/api/types'
import { useInjectedApis } from '@/api/use-apis'
import { withErrorToast } from '@/lib/api-error-toast'
import { queryKeys } from '@/features/query'
import { useWorkflowRefresh } from '@/features/workflow'

interface UseBudgetActionsOptions {
  injectedApis?: AppApis
  refresh: () => Promise<void>
  refreshApprovals: () => Promise<void>
}

export function useBudgetActions({
  injectedApis,
  refresh,
  refreshApprovals,
}: UseBudgetActionsOptions) {
  const apis = useInjectedApis(injectedApis)

  const { openWithRefresh } = useWorkflowRefresh({
    refresh,
    invalidateKeys: [queryKeys.keys.all],
  })

  const updateDepartment = useCallback(
    (departmentId: string, data: { budget: number; reservedPool?: number }) =>
      withErrorToast(async () => {
        await apis.budgetApi.updateDepartment(departmentId, data)
        await refresh()
      }, '更新部门预算失败'),
    [apis, refresh],
  )

  const resolveApproval = useCallback(
    (id: string, data: { status: 'approved' | 'rejected'; rejectReason?: string }) =>
      withErrorToast(async () => {
        await apis.budgetApi.resolveApproval(id, data)
        await refreshApprovals()
      }, '审批操作失败'),
    [apis, refreshApprovals],
  )

  const createProject = useCallback(
    (data: { name: string; budget: number; memberIds: string[]; ownerDepartmentId: string }) =>
      withErrorToast(async () => {
        await apis.budgetApi.createProject(data)
        await refresh()
      }, '创建项目失败'),
    [apis, refresh],
  )

  const updateProject = useCallback(
    (
      groupId: string,
      data: { budget?: number; memberIds?: string[]; memberBudgets?: Record<string, number> },
    ) =>
      withErrorToast(async () => {
        await apis.budgetApi.updateProject(groupId, data)
        await refresh()
      }, '更新项目失败'),
    [apis, refresh],
  )

  const deleteProject = useCallback(
    (groupId: string) =>
      withErrorToast(async () => {
        await apis.budgetApi.deleteProject(groupId)
        await refresh()
      }, '删除项目失败'),
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

  const getMemberBudgets = useCallback(
    (departmentId: string) => apis.budgetApi.getMemberBudgets(departmentId),
    [apis],
  )

  const updateMemberBudget = useCallback(
    (memberId: string, data: UpdateMemberBudgetInput) =>
      apis.budgetApi.updateMemberBudget(memberId, data),
    [apis],
  )

  const applyAverageBudget = useCallback(
    (departmentId: string, data: { personalBudget: number; recursive: boolean }) =>
      withErrorToast(async () => {
        await apis.budgetApi.applyAverageBudget(departmentId, data)
        await refresh()
      }, '应用均分额度失败'),
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

  return {
    updateDepartment,
    resolveApproval,
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
