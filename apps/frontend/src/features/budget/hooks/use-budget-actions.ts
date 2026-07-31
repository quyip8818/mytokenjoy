import { useCallback } from 'react'
import { useMutation } from '@tanstack/react-query'
import type { AppApis } from '@/api/app-apis'
import type { PlatformKeyScope, ProjectView, UpdateMemberBudgetInput } from '@/api/types'
import { useInjectedApis } from '@/api/use-apis'
import { queryKeys } from '@/features/query'
import { useWorkflowRefresh } from '@/features/workflow'

interface UseBudgetActionsOptions {
  injectedApis?: AppApis
  refresh: () => Promise<void>
}

export function useBudgetActions({ injectedApis, refresh }: UseBudgetActionsOptions) {
  const apis = useInjectedApis(injectedApis)

  const { openWithRefresh } = useWorkflowRefresh({
    refresh,
    invalidateKeys: [queryKeys.keys.all],
  })

  const updateDepartmentMutation = useMutation({
    mutationFn: (params: {
      departmentId: string
      data: { budget: number; reservedPool?: number }
    }) => apis.budgetApi.updateDepartment(params.departmentId, params.data),
    onSuccess: () => void refresh(),
  })

  const createProjectMutation = useMutation({
    mutationFn: (data: {
      name: string
      budget: number
      memberIds: string[]
      ownerDepartmentId: string
    }) => apis.budgetApi.createProject(data),
    onSuccess: () => void refresh(),
  })

  const updateProjectMutation = useMutation({
    mutationFn: (params: {
      groupId: string
      data: {
        budget?: number
        memberIds?: string[]
        memberBudgets?: Record<string, number>
        ownerId?: string
      }
    }) => apis.budgetApi.updateProject(params.groupId, params.data),
    onSuccess: () => void refresh(),
  })

  const deleteProjectMutation = useMutation({
    mutationFn: (groupId: string) => apis.budgetApi.deleteProject(groupId),
    onSuccess: () => void refresh(),
  })

  const applyAverageBudgetMutation = useMutation({
    mutationFn: (params: {
      departmentId: string
      data: { personalBudget: number; recursive: boolean }
    }) => apis.budgetApi.applyAverageBudget(params.departmentId, params.data),
    onSuccess: () => void refresh(),
  })

  // ponytail: mutateAsync throws on error, 让调用方 try/catch 控制自己的 UI 状态（如 setDeleting）。
  // toast 在调用方的 catch 中显示，或由调用方决定是否显示。
  // .then(() => {}) 丢弃 API 返回值使签名为 Promise<void>。
  const updateDepartment = useCallback(
    (departmentId: string, data: { budget: number; reservedPool?: number }): Promise<void> =>
      updateDepartmentMutation.mutateAsync({ departmentId, data }).then(() => {}),
    [updateDepartmentMutation],
  )

  const createProject = useCallback(
    (data: {
      name: string
      budget: number
      memberIds: string[]
      ownerDepartmentId: string
    }): Promise<void> => createProjectMutation.mutateAsync(data).then(() => {}),
    [createProjectMutation],
  )

  const updateProject = useCallback(
    (
      groupId: string,
      data: {
        budget?: number
        memberIds?: string[]
        memberBudgets?: Record<string, number>
        ownerId?: string
      },
    ): Promise<void> => updateProjectMutation.mutateAsync({ groupId, data }).then(() => {}),
    [updateProjectMutation],
  )

  const deleteProject = useCallback(
    (groupId: string): Promise<void> => deleteProjectMutation.mutateAsync(groupId).then(() => {}),
    [deleteProjectMutation],
  )

  const applyAverageBudget = useCallback(
    (departmentId: string, data: { personalBudget: number; recursive: boolean }): Promise<void> =>
      applyAverageBudgetMutation.mutateAsync({ departmentId, data }).then(() => {}),
    [applyAverageBudgetMutation],
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

  const getDepartmentTree = useCallback(() => apis.orgApi.departments.getTree(), [apis])

  const getMembers = useCallback(
    async (departmentId: string) => {
      const result = await apis.orgApi.members.list({
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
      const result = await apis.orgApi.members.list({ departmentId, page: 1, pageSize: 200 })
      return result?.items ?? []
    },
    [apis],
  )

  const searchMembers = useCallback(
    async (keyword: string) => {
      const result = await apis.orgApi.members.list({ keyword, page: 1, pageSize: 50 })
      return result?.items ?? []
    },
    [apis],
  )

  return {
    updateDepartment,
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
    // isPending flags for UI disable
    mutating:
      updateDepartmentMutation.isPending ||
      createProjectMutation.isPending ||
      updateProjectMutation.isPending ||
      deleteProjectMutation.isPending ||
      applyAverageBudgetMutation.isPending,
  }
}
